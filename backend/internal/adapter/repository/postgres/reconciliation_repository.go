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

type ReconciliationRepository struct {
	pool     *pgxpool.Pool
	stmtRepo *BankStatementRepository
	Logger   *slog.Logger
}

func NewReconciliationRepository(pool *pgxpool.Pool, stmtRepo *BankStatementRepository, logger *slog.Logger) *ReconciliationRepository {
	return &ReconciliationRepository{pool: pool, stmtRepo: stmtRepo, Logger: logger}
}

func (r *ReconciliationRepository) Save(ctx context.Context, rec entity.Reconciliation) (entity.Reconciliation, error) {
	r.Logger.Info("Saving 1:1 reconciliation", "settlement_tx_hash", rec.SettlementTransactionHash, "target_tx_hash", rec.TargetTransactionHash)

	if rec.ID == uuid.Nil {
		rec.ID = uuid.New()
	}
	if rec.ReconciledAt.IsZero() {
		rec.ReconciledAt = time.Now().UTC()
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return entity.Reconciliation{}, fmt.Errorf("reconciliation repo: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO reconciliations
			(id, settlement_transaction_hash, target_transaction_hash, amount, reconciled_at)
		VALUES ($1, $2, $3, $4, $5)`,
		rec.ID,
		rec.SettlementTransactionHash,
		rec.TargetTransactionHash,
		rec.Amount,
		rec.ReconciledAt,
	)
	if err != nil {
		return entity.Reconciliation{}, fmt.Errorf("reconciliation repo: insert: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE transactions
		SET is_reconciled = true, reconciliation_id = $1
		WHERE content_hash IN ($2, $3)`,
		rec.ID, rec.SettlementTransactionHash, rec.TargetTransactionHash,
	)
	if err != nil {
		return entity.Reconciliation{}, fmt.Errorf("reconciliation repo: mark transactions reconciled: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return entity.Reconciliation{}, fmt.Errorf("reconciliation repo: commit: %w", err)
	}

	return rec, nil
}

func (r *ReconciliationRepository) FindBySettlementTx(ctx context.Context, hash string) (entity.Reconciliation, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, settlement_transaction_hash, target_transaction_hash, amount, reconciled_at
		FROM reconciliations
		WHERE settlement_transaction_hash = $1`, hash)

	return scanReconciliation(row)
}

func (r *ReconciliationRepository) FindByTargetTx(ctx context.Context, hash string) (entity.Reconciliation, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, settlement_transaction_hash, target_transaction_hash, amount, reconciled_at
		FROM reconciliations
		WHERE target_transaction_hash = $1`, hash)

	return scanReconciliation(row)
}

func (r *ReconciliationRepository) FindAll(ctx context.Context) ([]entity.Reconciliation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT 
			r.id, 
			r.settlement_transaction_hash, 
			r.target_transaction_hash, 
			r.amount, 
			r.reconciled_at,
			s.description AS settlement_description,
			t.description AS target_description,
			s.booking_date AS settlement_booking_date,
			t.booking_date AS target_booking_date,
			bs_s.statement_type AS settlement_statement_type,
			bs_t.statement_type AS target_statement_type
		FROM reconciliations r
		LEFT JOIN transactions s ON r.settlement_transaction_hash = s.content_hash
		LEFT JOIN transactions t ON r.target_transaction_hash = t.content_hash
		LEFT JOIN bank_statements bs_s ON s.bank_statement_id = bs_s.id
		LEFT JOIN bank_statements bs_t ON t.bank_statement_id = bs_t.id
		ORDER BY r.reconciled_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("reconciliation repo: find all: %w", err)
	}
	defer rows.Close()

	var results []entity.Reconciliation
	for rows.Next() {
		var rec entity.Reconciliation
		var settleDesc, targetDesc, settleStmtType, targetStmtType *string
		var settleDate, targetDate *time.Time

		err := rows.Scan(
			&rec.ID,
			&rec.SettlementTransactionHash,
			&rec.TargetTransactionHash,
			&rec.Amount,
			&rec.ReconciledAt,
			&settleDesc,
			&targetDesc,
			&settleDate,
			&targetDate,
			&settleStmtType,
			&targetStmtType,
		)
		if err != nil {
			return nil, fmt.Errorf("reconciliation repo: scan: %w", err)
		}

		if settleDesc != nil {
			rec.SettlementTransactionDescription = *settleDesc
		}
		if targetDesc != nil {
			rec.TargetTransactionDescription = *targetDesc
		}
		if settleDate != nil {
			rec.SettlementBookingDate = *settleDate
		}
		if targetDate != nil {
			rec.TargetBookingDate = *targetDate
		}
		if settleStmtType != nil {
			rec.SettlementStatementType = *settleStmtType
		}
		if targetStmtType != nil {
			rec.TargetStatementType = *targetStmtType
		}

		results = append(results, rec)
	}
	return results, rows.Err()
}

type reconciliationScanner interface {
	Scan(dest ...any) error
}

func scanReconciliation(row reconciliationScanner) (entity.Reconciliation, error) {
	var rec entity.Reconciliation
	err := row.Scan(
		&rec.ID,
		&rec.SettlementTransactionHash,
		&rec.TargetTransactionHash,
		&rec.Amount,
		&rec.ReconciledAt,
	)
	return rec, err
}

func (r *ReconciliationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("reconciliation repo: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Unlink the transactions by resetting the flags
	_, err = tx.Exec(ctx, `
		UPDATE transactions
		SET is_reconciled = false, reconciliation_id = NULL
		WHERE reconciliation_id = $1`, id)
	if err != nil {
		return fmt.Errorf("reconciliation repo: unlink transactions: %w", err)
	}

	// Delete the actual reconciliation record
	_, err = tx.Exec(ctx, `DELETE FROM reconciliations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("reconciliation repo: delete record: %w", err)
	}

	return tx.Commit(ctx)
}
