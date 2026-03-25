package postgres

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"cogni-cash/internal/domain/entity"
)

// InvoiceRepository implements port.InvoiceRepository using pgx.
type InvoiceRepository struct {
	pool   *pgxpool.Pool
	Logger *slog.Logger // Structured logger
}

// NewInvoiceRepository creates a new InvoiceRepository.
func NewInvoiceRepository(pool *pgxpool.Pool, logger *slog.Logger) *InvoiceRepository {
	return &InvoiceRepository{pool: pool, Logger: logger}
}

// Save inserts or updates an Invoice record.
func (r *InvoiceRepository) Save(ctx context.Context, inv entity.Invoice) error {
	var issuedAt *time.Time
	if !inv.IssuedAt.IsZero() {
		issuedAt = &inv.IssuedAt
	}
	r.Logger.Info("Saving invoice", "id", inv.ID)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO invoices (id, raw_text, category_id, vendor, amount, currency, invoice_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			raw_text     = EXCLUDED.raw_text,
			category_id  = EXCLUDED.category_id,
			vendor       = EXCLUDED.vendor,
			amount       = EXCLUDED.amount,
			currency     = EXCLUDED.currency,
			invoice_date = EXCLUDED.invoice_date`,
		inv.ID,
		inv.RawText,
		inv.CategoryID,
		inv.Vendor.Name,
		inv.Amount,
		inv.Currency,
		issuedAt,
	)
	if err != nil {
		r.Logger.Error("Failed to save invoice", "id", inv.ID, "error", err)
		return fmt.Errorf("invoice repo: save: %w", err)
	}
	r.Logger.Info("Invoice saved successfully", "id", inv.ID)
	return nil
}

// FindByID retrieves a single Invoice by UUID.
func (r *InvoiceRepository) FindByID(ctx context.Context, id uuid.UUID) (entity.Invoice, error) {
	r.Logger.Info("Finding invoice by ID", "id", id)
	row := r.pool.QueryRow(ctx, `
		SELECT id, raw_text, category_id, vendor, amount, currency, invoice_date
		FROM invoices WHERE id = $1`, id)

	inv, err := scanInvoice(row)
	if err != nil {
		r.Logger.Error("Failed to find invoice by ID", "id", id, "error", err)
		return entity.Invoice{}, fmt.Errorf("invoice repo: find by id: %w", err)
	}
	return inv, nil
}

// FindAll returns all Invoices ordered by creation time descending.
func (r *InvoiceRepository) FindAll(ctx context.Context) ([]entity.Invoice, error) {
	r.Logger.Info("Finding all invoices")
	rows, err := r.pool.Query(ctx, `
		SELECT id, raw_text, category_id, vendor, amount, currency, invoice_date
		FROM invoices
		ORDER BY created_at DESC`)
	if err != nil {
		r.Logger.Error("Failed to query invoices", "error", err)
		return nil, fmt.Errorf("invoice repo: find all: %w", err)
	}
	defer rows.Close()

	var invoices []entity.Invoice
	for rows.Next() {
		inv, err := scanInvoice(rows)
		if err != nil {
			r.Logger.Error("Failed to scan invoice row", "error", err)
			return nil, fmt.Errorf("invoice repo: scan: %w", err)
		}
		invoices = append(invoices, inv)
	}
	return invoices, rows.Err()
}

// Delete removes an Invoice by UUID.
func (r *InvoiceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.Logger.Info("Deleting invoice", "id", id)
	tag, err := r.pool.Exec(ctx, `DELETE FROM invoices WHERE id = $1`, id)
	if err != nil {
		r.Logger.Error("Failed to delete invoice", "id", id, "error", err)
		return fmt.Errorf("invoice repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		r.Logger.Warn("Invoice not found for delete", "id", id)
		return fmt.Errorf("invoice repo: not found: %s", id)
	}
	return nil
}

// ---- helpers ---------------------------------------------------------------

func scanInvoice(row scanner) (entity.Invoice, error) {
	var (
		inv        entity.Invoice
		categoryID *uuid.UUID
		vendorName string
		issuedAt   *time.Time
	)
	err := row.Scan(
		&inv.ID,
		&inv.RawText,
		&categoryID,
		&vendorName,
		&inv.Amount,
		&inv.Currency,
		&issuedAt,
	)
	if err != nil {
		return entity.Invoice{}, err
	}

	inv.CategoryID = categoryID
	inv.Vendor = entity.Vendor{Name: vendorName}
	if issuedAt != nil {
		inv.IssuedAt = *issuedAt
	}
	return inv, nil
}
