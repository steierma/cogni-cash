// File: internal/adapter/repository/postgres/reconciliation_repository_test.go
package postgres

import (
	"context"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

func TestReconciliationRepository(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t)

	repo := NewReconciliationRepository(globalPool, nil, setupLogger())

	stmtID := uuid.New()
	settlementHash := "giro_debit_tx_hash"
	targetHash := "cc_credit_tx_hash"

	_, _ = globalPool.Exec(ctx, `INSERT INTO bank_statements (id, content_hash) VALUES ($1, 'cc_statement_hash')`, stmtID)

	_, _ = globalPool.Exec(ctx, `INSERT INTO transactions (id, bank_statement_id, booking_date, valuta_date, amount, transaction_type, content_hash) VALUES ($1, $2, $3, $3, -500.00, 'debit', $4)`, uuid.New(), stmtID, time.Now().UTC(), settlementHash)
	_, _ = globalPool.Exec(ctx, `INSERT INTO transactions (id, bank_statement_id, booking_date, valuta_date, amount, transaction_type, content_hash) VALUES ($1, $2, $3, $3, 500.00, 'credit', $4)`, uuid.New(), stmtID, time.Now().UTC(), targetHash)

	t.Run("Save, Find, and Delete Reconciliations", func(t *testing.T) {
		rec := entity.Reconciliation{
			SettlementTransactionHash: settlementHash,
			TargetTransactionHash:     targetHash,
			Amount:                    500.00,
		}

		// 1. Save and check for valid UUID
		saved, err := repo.Save(ctx, rec)
		if err != nil {
			t.Fatalf("expected no error saving reconciliation, got: %v", err)
		}
		if saved.ID == uuid.Nil {
			t.Error("expected valid UUID from Save")
		}

		// 2. Verify both transactions were marked as reconciled
		var isReconciled1, isReconciled2 bool
		_ = globalPool.QueryRow(ctx, "SELECT is_reconciled FROM transactions WHERE content_hash = $1", settlementHash).Scan(&isReconciled1)
		_ = globalPool.QueryRow(ctx, "SELECT is_reconciled FROM transactions WHERE content_hash = $1", targetHash).Scan(&isReconciled2)

		if !isReconciled1 || !isReconciled2 {
			t.Errorf("expected both transactions to be marked as reconciled")
		}

		// 3. Search via Settlement Hash
		found, err := repo.FindBySettlementTx(ctx, settlementHash)
		if err != nil || found.Amount != 500.00 {
			t.Errorf("expected amount 500.00, got %f", found.Amount)
		}

		// 4. Test FindAll (with the new LEFT JOIN logic)
		all, err := repo.FindAll(ctx)
		if err != nil {
			t.Fatalf("expected no error finding all, got: %v", err)
		}
		if len(all) != 1 {
			t.Fatalf("expected 1 reconciliation total, got %d", len(all))
		}

		// 5. Test Delete (Unlinking transactions)
		err = repo.Delete(ctx, saved.ID)
		if err != nil {
			t.Fatalf("expected no error deleting reconciliation, got: %v", err)
		}

		// Verify transactions are unlinked
		_ = globalPool.QueryRow(ctx, "SELECT is_reconciled FROM transactions WHERE content_hash = $1", settlementHash).Scan(&isReconciled1)
		_ = globalPool.QueryRow(ctx, "SELECT is_reconciled FROM transactions WHERE content_hash = $1", targetHash).Scan(&isReconciled2)

		if isReconciled1 || isReconciled2 {
			t.Errorf("expected both transactions to be unlinked (is_reconciled = false) after deletion")
		}

		// Verify reconciliation record is gone
		allAfterDelete, _ := repo.FindAll(ctx)
		if len(allAfterDelete) != 0 {
			t.Errorf("expected 0 reconciliations after deletion, got %d", len(allAfterDelete))
		}
	})
}
