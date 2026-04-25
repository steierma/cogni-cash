package postgres

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cogni-cash/internal/domain/entity"
)

// InvoiceRepository implements port.InvoiceRepository using pgx.
type InvoiceRepository struct {
	pool     *pgxpool.Pool
	vaultKey string
	Logger   *slog.Logger
}

// NewInvoiceRepository creates a new InvoiceRepository.
func NewInvoiceRepository(pool *pgxpool.Pool, vaultKey string, logger *slog.Logger) *InvoiceRepository {
	return &InvoiceRepository{pool: pool, vaultKey: vaultKey, Logger: logger}
}

// Save inserts or upserts an Invoice record (keyed on id).
func (r *InvoiceRepository) Save(ctx context.Context, inv entity.Invoice) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var issuedAt *time.Time
	if !inv.IssuedAt.IsZero() {
		issuedAt = &inv.IssuedAt
	}
	var contentHash *string
	if inv.ContentHash != "" {
		contentHash = &inv.ContentHash
	}
	var origName *string
	if inv.OriginalFileName != "" {
		origName = &inv.OriginalFileName
	}
	var origContent []byte
	if len(inv.OriginalFileContent) > 0 {
		origContent = inv.OriginalFileContent
	}

	r.Logger.Info("Saving invoice", "id", inv.ID, "user_id", inv.UserID)
	_, err = tx.Exec(ctx, `
		INSERT INTO invoices (
			id, user_id, vendor, amount, currency, base_amount, base_currency, invoice_date,
			description, category_id,
			content_hash, original_file_name, original_file_content
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,pgp_sym_encrypt_bytea($13, $14))
		ON CONFLICT (id) DO UPDATE SET
			user_id              = EXCLUDED.user_id,
			vendor               = EXCLUDED.vendor,
			amount               = EXCLUDED.amount,
			currency             = EXCLUDED.currency,
			base_amount          = EXCLUDED.base_amount,
			base_currency        = EXCLUDED.base_currency,
			invoice_date         = EXCLUDED.invoice_date,
			description          = EXCLUDED.description,
			category_id          = EXCLUDED.category_id,
			content_hash         = COALESCE(EXCLUDED.content_hash, invoices.content_hash),
			original_file_name   = COALESCE(EXCLUDED.original_file_name, invoices.original_file_name),
			original_file_content= COALESCE(EXCLUDED.original_file_content, invoices.original_file_content)`,
		inv.ID, inv.UserID, inv.Vendor.Name, inv.Amount, inv.Currency, inv.BaseAmount, inv.BaseCurrency, issuedAt,
		inv.Description, inv.CategoryID,
		contentHash, origName, origContent, r.vaultKey,
	)
	if err != nil {
		return fmt.Errorf("save invoice: %w", err)
	}

	// Save splits
	if len(inv.Splits) > 0 {
		for _, split := range inv.Splits {
			_, err = tx.Exec(ctx, `
				INSERT INTO invoice_line_items (id, user_id, invoice_id, category_id, amount, base_amount, description)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				ON CONFLICT (id) DO UPDATE SET
					category_id = EXCLUDED.category_id,
					amount      = EXCLUDED.amount,
					base_amount = EXCLUDED.base_amount,
					description = EXCLUDED.description`,
				split.ID, inv.UserID, inv.ID, split.CategoryID, split.Amount, split.BaseAmount, split.Description)
			if err != nil {
				return fmt.Errorf("save split: %w", err)
			}
		}
	}

	return tx.Commit(ctx)
}

// Update patches the mutable fields of an existing invoice (does not touch
// content_hash or original file data, which are immutable after import).
func (r *InvoiceRepository) Update(ctx context.Context, inv entity.Invoice) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var issuedAt *time.Time
	if !inv.IssuedAt.IsZero() {
		issuedAt = &inv.IssuedAt
	}
	r.Logger.Info("Updating invoice (checking permissions)", "id", inv.ID, "user_id", inv.UserID)
	// Permission check: either the owner or someone with 'edit' permission can update.
	tag, err := tx.Exec(ctx, `
		UPDATE invoices SET
			vendor        = $2,
			amount        = $3,
			currency      = $4,
			base_amount   = $5,
			base_currency = $6,
			invoice_date  = $7,
			description   = $8,
			category_id   = $9
		WHERE id = $1 AND (
			user_id = $10 
			OR id IN (SELECT invoice_id FROM shared_invoices WHERE shared_with_user_id = $10 AND permission_level = 'edit')
			OR category_id IN (SELECT category_id FROM shared_categories WHERE shared_with_user_id = $10 AND permission_level = 'edit')
		)`,
		inv.ID, inv.Vendor.Name, inv.Amount, inv.Currency, inv.BaseAmount, inv.BaseCurrency, issuedAt, inv.Description, inv.CategoryID, inv.UserID,
	)

	if err != nil {
		return fmt.Errorf("update invoice: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return entity.ErrInvoiceNotFound
	}

	// Update splits: Replace existing splits with new ones (scoped to user for safety)
	if _, err = tx.Exec(ctx, `DELETE FROM invoice_line_items WHERE invoice_id = $1 AND user_id = $2`, inv.ID, inv.UserID); err != nil {
		return fmt.Errorf("delete splits: %w", err)
	}

	for _, split := range inv.Splits {
		if split.ID == uuid.Nil {
			split.ID = uuid.New()
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO invoice_line_items (id, user_id, invoice_id, category_id, amount, base_amount, description)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			split.ID, inv.UserID, inv.ID, split.CategoryID, split.Amount, split.BaseAmount, split.Description)
		if err != nil {
			return fmt.Errorf("insert split: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *InvoiceRepository) UpdateBaseAmount(ctx context.Context, id uuid.UUID, baseAmount float64, baseCurrency string, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE invoices SET base_amount = $1, base_currency = $2
		WHERE id = $3 AND user_id = $4`, baseAmount, baseCurrency, id, userID)
	return err
}

// FindByID retrieves a single Invoice by UUID, including shared ones.
func (r *InvoiceRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Invoice, error) {
	r.Logger.Info("Finding invoice by ID (including shared)", "id", id, "user_id", userID)
	row := r.pool.QueryRow(ctx, `
		SELECT 
			i.id, i.user_id, i.category_id, i.vendor, i.amount, i.currency, i.base_amount, i.base_currency, i.invoice_date,
			i.content_hash, i.original_file_name,
			i.description,
			(i.user_id != $2) as is_shared,
			COALESCE((SELECT array_agg(shared_with_user_id) FROM shared_invoices WHERE invoice_id = i.id), '{}') as shared_with,
			i.user_id as owner_id
		FROM invoices i 
		WHERE i.id = $1 AND (
			i.user_id = $2 
			OR i.id IN (SELECT invoice_id FROM shared_invoices WHERE shared_with_user_id = $2)
			OR i.category_id IN (SELECT category_id FROM shared_categories WHERE shared_with_user_id = $2)
		)`, id, userID)
	inv, err := scanInvoiceWithSharing(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Invoice{}, entity.ErrInvoiceNotFound
		}
		return entity.Invoice{}, fmt.Errorf("invoice repo: find by id: %w", err)
	}

	// Fetch splits
	splits, err := r.fetchSplits(ctx, inv.ID, userID)
	if err != nil {
		return entity.Invoice{}, err
	}
	inv.Splits = splits

	return inv, nil
}

// FindAll returns all Invoices (owned and shared) ordered by creation time descending.
func (r *InvoiceRepository) FindAll(ctx context.Context, filter entity.InvoiceFilter) ([]entity.Invoice, error) {
	r.Logger.Info("Finding all invoices (including shared)", "user_id", filter.UserID, "include_shared", filter.IncludeShared, "source", filter.Source)

	query := `
		SELECT 
			i.id, i.user_id, i.category_id, i.vendor, i.amount, i.currency, i.base_amount, i.base_currency, i.invoice_date,
			i.content_hash, i.original_file_name,
			i.description,
			(i.user_id != $1) as is_shared,
			COALESCE((SELECT array_agg(shared_with_user_id) FROM shared_invoices WHERE invoice_id = i.id), '{}') as shared_with,
			i.user_id as owner_id
		FROM invoices i`

	args := []any{filter.UserID}
	where := ""

	if filter.IncludeShared || filter.Source == "all" {
		where = " WHERE (i.user_id = $1 OR i.id IN (SELECT invoice_id FROM shared_invoices WHERE shared_with_user_id = $1) OR i.category_id IN (SELECT category_id FROM shared_categories WHERE shared_with_user_id = $1))"
	} else if filter.Source == "shared" {
		where = " WHERE (i.id IN (SELECT invoice_id FROM shared_invoices WHERE shared_with_user_id = $1) OR i.category_id IN (SELECT category_id FROM shared_categories WHERE shared_with_user_id = $1)) AND i.user_id != $1"
	} else {
		// Default: only mine
		where = " WHERE i.user_id = $1"
	}

	if filter.Year > 0 {
		where += fmt.Sprintf(" AND EXTRACT(YEAR FROM i.invoice_date) = $%d", len(args)+1)
		args = append(args, filter.Year)
	}

	query += where + " ORDER BY i.created_at DESC"

	if filter.Limit > 0 {
		args = append(args, filter.Limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	if filter.Offset > 0 {
		args = append(args, filter.Offset)
		query += fmt.Sprintf(" OFFSET $%d", len(args))
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("invoice repo: find all: %w", err)
	}
	defer rows.Close()

	var invoices []entity.Invoice
	var invoiceIDs []uuid.UUID
	for rows.Next() {
		inv, err := scanInvoiceWithSharing(rows)
		if err != nil {
			return nil, fmt.Errorf("invoice repo: scan: %w", err)
		}
		invoices = append(invoices, inv)
		invoiceIDs = append(invoiceIDs, inv.ID)
	}

	if len(invoices) > 0 {
		// Fetch all splits for these invoices in one query
		allSplits, err := r.fetchSplitsForInvoices(ctx, invoiceIDs, filter.UserID)
		if err != nil {
			return nil, err
		}

		// Map splits to invoices
		for i := range invoices {
			invoices[i].Splits = allSplits[invoices[i].ID]
		}
	}

	return invoices, rows.Err()
}

// Delete removes an Invoice by UUID.
func (r *InvoiceRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	r.Logger.Info("Deleting invoice", "id", id, "user_id", userID)
	tag, err := r.pool.Exec(ctx, `DELETE FROM invoices WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("invoice repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("invoice repo: delete: %w", entity.ErrInvoiceNotFound)
	}
	return nil
}

// DeleteSplits removes all line items for an invoice.
func (r *InvoiceRepository) DeleteSplits(ctx context.Context, invoiceID, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM invoice_line_items WHERE invoice_id = $1 AND user_id = $2`, invoiceID, userID)
	return err
}

// ExistsByContentHash returns true when an invoice with the given SHA-256 hash
// is already stored (used for deduplication on file import).
func (r *InvoiceRepository) ExistsByContentHash(ctx context.Context, hash string, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM invoices WHERE content_hash = $1 AND user_id = $2)`, hash, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("invoice repo: exists by hash: %w", err)
	}
	return exists, nil
}

// GetOriginalFile returns the stored binary content, MIME type and file name for
// an invoice, or an error when no file was attached.
func (r *InvoiceRepository) GetOriginalFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) ([]byte, string, string, error) {
	var rawContent []byte
	var name *string
	err := r.pool.QueryRow(ctx, `
		SELECT original_file_content, original_file_name 
		FROM invoices 
		WHERE id = $1 AND (
			user_id = $2 
			OR id IN (SELECT invoice_id FROM shared_invoices WHERE shared_with_user_id = $2)
			OR category_id IN (SELECT category_id FROM shared_categories WHERE shared_with_user_id = $2)
		)`, id, userID).
		Scan(&rawContent, &name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", "", entity.ErrInvoiceNotFound
		}
		return nil, "", "", fmt.Errorf("invoice repo: get original file: %w", err)
	}
	if len(rawContent) == 0 {
		return nil, "", "", fmt.Errorf("invoice repo: no file stored for invoice %s", id)
	}

	// Try decryption
	var content []byte
	decryptQuery := "SELECT pgp_sym_decrypt_bytea($1, $2)"
	err = r.pool.QueryRow(ctx, decryptQuery, rawContent, r.vaultKey).Scan(&content)
	if err != nil {
		// Fallback: If decryption fails, it might be legacy plain text or corrupt.
		r.Logger.Warn("Invoice decryption failed, returning raw content", "id", id, "error", err)
		content = rawContent
	}

	nameStr := ""
	if name != nil {
		nameStr = *name
	}

	mimeType := http.DetectContentType(content)
	if idx := strings.IndexByte(mimeType, ';'); idx >= 0 {
		mimeType = mimeType[:idx]
	}

	return content, mimeType, nameStr, nil
}

// ── scanner helper ──────────────────────────────────────────────────────────

func (r *InvoiceRepository) fetchSplits(ctx context.Context, invoiceID, userID uuid.UUID) ([]entity.InvoiceSplit, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, invoice_id, category_id, amount, base_amount, description 
		FROM invoice_line_items 
		WHERE invoice_id = $1 AND user_id = $2`, invoiceID, userID)
	if err != nil {
		return nil, fmt.Errorf("fetch splits: %w", err)
	}
	defer rows.Close()

	var splits []entity.InvoiceSplit
	for rows.Next() {
		var s entity.InvoiceSplit
		if err := rows.Scan(&s.ID, &s.UserID, &s.InvoiceID, &s.CategoryID, &s.Amount, &s.BaseAmount, &s.Description); err != nil {
			return nil, fmt.Errorf("scan split: %w", err)
		}
		splits = append(splits, s)
	}
	return splits, nil
}

func (r *InvoiceRepository) fetchSplitsForInvoices(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) (map[uuid.UUID][]entity.InvoiceSplit, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, invoice_id, category_id, amount, base_amount, description 
		FROM invoice_line_items 
		WHERE invoice_id = ANY($1) AND user_id = $2`, ids, userID)
	if err != nil {
		return nil, fmt.Errorf("fetch all splits: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]entity.InvoiceSplit)
	for rows.Next() {
		var s entity.InvoiceSplit
		if err := rows.Scan(&s.ID, &s.UserID, &s.InvoiceID, &s.CategoryID, &s.Amount, &s.BaseAmount, &s.Description); err != nil {
			return nil, fmt.Errorf("scan split: %w", err)
		}
		result[s.InvoiceID] = append(result[s.InvoiceID], s)
	}
	return result, nil
}

func scanInvoiceWithSharing(row scanner) (entity.Invoice, error) {
	var (
		inv         entity.Invoice
		categoryID  *uuid.UUID
		vendorName  string
		issuedAt    *time.Time
		contentHash *string
		origName    *string
		description string
	)
	err := row.Scan(
		&inv.ID,
		&inv.UserID,
		&categoryID,
		&vendorName,
		&inv.Amount,
		&inv.Currency,
		&inv.BaseAmount,
		&inv.BaseCurrency,
		&issuedAt,
		&contentHash,
		&origName,
		&description,
		&inv.IsShared,
		&inv.SharedWith,
		&inv.OwnerID,
	)
	if err != nil {
		return entity.Invoice{}, err
	}

	inv.CategoryID = categoryID
	inv.Vendor = entity.Vendor{Name: vendorName}
	inv.Description = description
	if issuedAt != nil {
		inv.IssuedAt = *issuedAt
	}
	if contentHash != nil {
		inv.ContentHash = *contentHash
	}
	if origName != nil {
		inv.OriginalFileName = *origName
	}
	return inv, nil
}
