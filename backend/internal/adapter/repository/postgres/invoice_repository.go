package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cogni-cash/internal/domain/entity"
)

// InvoiceRepository implements port.InvoiceRepository using pgx.
type InvoiceRepository struct {
	pool   *pgxpool.Pool
	Logger *slog.Logger
}

// NewInvoiceRepository creates a new InvoiceRepository.
func NewInvoiceRepository(pool *pgxpool.Pool, logger *slog.Logger) *InvoiceRepository {
	return &InvoiceRepository{pool: pool, Logger: logger}
}

// Save inserts or upserts an Invoice record (keyed on id).
func (r *InvoiceRepository) Save(ctx context.Context, inv entity.Invoice) error {
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
	_, err := r.pool.Exec(ctx, `
		INSERT INTO invoices (
			id, user_id, category_id, vendor, amount, currency, invoice_date,
			description,
			content_hash, original_file_name, original_file_content
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			user_id              = EXCLUDED.user_id,
			category_id          = EXCLUDED.category_id,
			vendor               = EXCLUDED.vendor,
			amount               = EXCLUDED.amount,
			currency             = EXCLUDED.currency,
			invoice_date         = EXCLUDED.invoice_date,
			description          = EXCLUDED.description,
			content_hash         = COALESCE(EXCLUDED.content_hash, invoices.content_hash),
			original_file_name   = COALESCE(EXCLUDED.original_file_name, invoices.original_file_name),
			original_file_content= COALESCE(EXCLUDED.original_file_content, invoices.original_file_content)`,
		inv.ID, inv.UserID, inv.CategoryID, inv.Vendor.Name, inv.Amount, inv.Currency, issuedAt,
		inv.Description,
		contentHash, origName, origContent,
	)
	if err != nil {
		r.Logger.Error("Failed to save invoice", "id", inv.ID, "user_id", inv.UserID, "error", err)
		return fmt.Errorf("invoice repo: save: %w", err)
	}
	r.Logger.Info("Invoice saved successfully", "id", inv.ID)
	return nil
}

// Update patches the mutable fields of an existing invoice (does not touch
// content_hash or original file data, which are immutable after import).
func (r *InvoiceRepository) Update(ctx context.Context, inv entity.Invoice) error {
	var issuedAt *time.Time
	if !inv.IssuedAt.IsZero() {
		issuedAt = &inv.IssuedAt
	}
	r.Logger.Info("Updating invoice", "id", inv.ID, "user_id", inv.UserID)
	tag, err := r.pool.Exec(ctx, `
		UPDATE invoices SET
			category_id  = $2,
			vendor       = $3,
			amount       = $4,
			currency     = $5,
			invoice_date = $6,
			description  = $7
		WHERE id = $1 AND user_id = $8`,
		inv.ID, inv.CategoryID, inv.Vendor.Name, inv.Amount, inv.Currency, issuedAt, inv.Description, inv.UserID,
	)
	if err != nil {
		return fmt.Errorf("invoice repo: update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("invoice repo: update: %w", entity.ErrInvoiceNotFound)
	}
	return nil
}

// FindByID retrieves a single Invoice by UUID.
func (r *InvoiceRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Invoice, error) {
	r.Logger.Info("Finding invoice by ID", "id", id, "user_id", userID)
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, category_id, vendor, amount, currency, invoice_date,
		       content_hash, original_file_name,
		       description
		FROM invoices WHERE id = $1 AND user_id = $2`, id, userID)
	inv, err := scanInvoice(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Invoice{}, entity.ErrInvoiceNotFound
		}
		return entity.Invoice{}, fmt.Errorf("invoice repo: find by id: %w", err)
	}
	return inv, nil
}

// FindAll returns all Invoices ordered by creation time descending.
func (r *InvoiceRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]entity.Invoice, error) {
	r.Logger.Info("Finding all invoices", "user_id", userID)
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, category_id, vendor, amount, currency, invoice_date,
		       content_hash, original_file_name,
		       description
		FROM invoices
		WHERE user_id = $1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("invoice repo: find all: %w", err)
	}
	defer rows.Close()

	var invoices []entity.Invoice
	for rows.Next() {
		inv, err := scanInvoice(rows)
		if err != nil {
			return nil, fmt.Errorf("invoice repo: scan: %w", err)
		}
		invoices = append(invoices, inv)
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
	var content []byte
	var name *string
	err := r.pool.QueryRow(ctx,
		`SELECT original_file_content, original_file_name FROM invoices WHERE id = $1 AND user_id = $2`, id, userID).
		Scan(&content, &name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", "", entity.ErrInvoiceNotFound
		}
		return nil, "", "", fmt.Errorf("invoice repo: get original file: %w", err)
	}
	if len(content) == 0 {
		return nil, "", "", fmt.Errorf("invoice repo: no file stored for invoice %s", id)
	}
	nameStr := ""
	if name != nil {
		nameStr = *name
	}
	return content, "application/pdf", nameStr, nil
}

// ── scanner helper ──────────────────────────────────────────────────────────

func scanInvoice(row scanner) (entity.Invoice, error) {
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
		&issuedAt,
		&contentHash,
		&origName,
		&description,
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
