package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SharingRepository implements port.SharingRepository using pgx.
type SharingRepository struct {
	pool   *pgxpool.Pool
	Logger *slog.Logger
}

// NewSharingRepository creates a new SharingRepository.
func NewSharingRepository(pool *pgxpool.Pool, logger *slog.Logger) *SharingRepository {
	return &SharingRepository{pool: pool, Logger: logger}
}

func (r *SharingRepository) ShareCategory(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID, permission string) error {
	r.Logger.Info("Sharing category", "category_id", categoryID, "owner_id", ownerID, "shared_with", sharedWithID)

	_, err := r.pool.Exec(ctx, `
		INSERT INTO shared_categories (category_id, owner_user_id, shared_with_user_id, permission_level)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (owner_user_id, category_id, shared_with_user_id) DO UPDATE SET permission_level = EXCLUDED.permission_level`,
		categoryID, ownerID, sharedWithID, permission)
	if err != nil {
		return fmt.Errorf("sharing repo: share category: %w", err)
	}
	return nil
}

func (r *SharingRepository) RevokeShare(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID) error {
	r.Logger.Info("Revoking category share", "category_id", categoryID, "owner_id", ownerID, "shared_with", sharedWithID)

	_, err := r.pool.Exec(ctx, `
		DELETE FROM shared_categories 
		WHERE category_id = $1 AND owner_user_id = $2 AND shared_with_user_id = $3`,
		categoryID, ownerID, sharedWithID)
	if err != nil {
		return fmt.Errorf("sharing repo: revoke share: %w", err)
	}
	// Note: We intentionally do not check RowsAffected() == 0.
	// Deletions should be idempotent. If it's already gone, we consider it a success.
	return nil
}

func (r *SharingRepository) ListShares(ctx context.Context, categoryID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	r.Logger.Info("Listing shares for category", "category_id", categoryID, "owner_id", ownerID)

	rows, err := r.pool.Query(ctx, `
		SELECT shared_with_user_id FROM shared_categories 
		WHERE category_id = $1 AND owner_user_id = $2`,
		categoryID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("sharing repo: list shares: %w", err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("sharing repo: scan: %w", err)
		}
		userIDs = append(userIDs, id)
	}
	return userIDs, rows.Err()
}

func (r *SharingRepository) ShareInvoice(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID, permission string) error {
	r.Logger.Info("Sharing invoice", "invoice_id", invoiceID, "owner_id", ownerID, "shared_with", sharedWithID)

	_, err := r.pool.Exec(ctx, `
		INSERT INTO shared_invoices (invoice_id, owner_user_id, shared_with_user_id, permission_level)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (owner_user_id, invoice_id, shared_with_user_id) DO UPDATE SET permission_level = EXCLUDED.permission_level`,
		invoiceID, ownerID, sharedWithID, permission)
	if err != nil {
		return fmt.Errorf("sharing repo: share invoice: %w", err)
	}
	return nil
}

func (r *SharingRepository) RevokeInvoiceShare(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID) error {
	r.Logger.Info("Revoking invoice share", "invoice_id", invoiceID, "owner_id", ownerID, "shared_with", sharedWithID)

	_, err := r.pool.Exec(ctx, `
		DELETE FROM shared_invoices 
		WHERE invoice_id = $1 AND owner_user_id = $2 AND shared_with_user_id = $3`,
		invoiceID, ownerID, sharedWithID)
	if err != nil {
		return fmt.Errorf("sharing repo: revoke invoice share: %w", err)
	}
	// Idempotent return
	return nil
}

func (r *SharingRepository) ListInvoiceShares(ctx context.Context, invoiceID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	r.Logger.Info("Listing shares for invoice", "invoice_id", invoiceID, "owner_id", ownerID)

	rows, err := r.pool.Query(ctx, `
		SELECT shared_with_user_id FROM shared_invoices 
		WHERE invoice_id = $1 AND owner_user_id = $2`,
		invoiceID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("sharing repo: list invoice shares: %w", err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("sharing repo: scan: %w", err)
		}
		userIDs = append(userIDs, id)
	}
	return userIDs, rows.Err()
}
