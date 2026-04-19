package postgres

import (
	"context"
	"errors"
	_ "io"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/service"
	_ "log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBankStatementRepository_SaveAndFindByID(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t) // Instant cleanup!

	repo := NewBankStatementRepository(globalPool, setupLogger())

	userID := uuid.New()
	_, err := globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'testuser_save', 'hash', 'test_save@example.com')", userID)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	var catID uuid.UUID
	_, err = globalPool.Exec(ctx, "INSERT INTO categories (id, user_id, name, color) VALUES ($1, $2, 'Sonstige Ausgaben', '#000000')", catID, userID)
	if err != nil {
		t.Fatalf("failed to insert category: %v", err)
	}

	stmtDate := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)
	stmt := entity.BankStatement{
		UserID:        userID,
		AccountHolder: "Jane Doe",
		StatementDate: stmtDate,
		Currency:      "EUR",
		ContentHash:   "stmt_hash_1",
		StatementType: entity.StatementType("giro"),
		Transactions: []entity.Transaction{
			{
				UserID:      userID,
				Description: "Rewe",
				Amount:      -45.50,
				Currency:    "EUR",
				BookingDate: stmtDate,
				ValutaDate:  stmtDate,
				Type:        "debit",
				CategoryID:  &catID,
				ContentHash: "tx_hash_1",
			},
		},
	}

	err = repo.Save(ctx, stmt)
	if err != nil {
		t.Fatalf("expected no error on save, got: %v", err)
	}

	// Safely lookup ID without assuming length of FindAll matches 1 globally
	var savedID uuid.UUID
	err = globalPool.QueryRow(ctx, "SELECT id FROM bank_statements WHERE content_hash = 'stmt_hash_1' AND user_id = $1", userID).Scan(&savedID)
	if err != nil {
		t.Fatalf("failed to fetch saved statement id: %v", err)
	}

	found, err := repo.FindByID(ctx, savedID, userID)
	if err != nil {
		t.Fatalf("expected to find statement, got error: %v", err)
	}
	if found.AccountHolder != "Jane Doe" {
		t.Errorf("expected AccountHolder Jane Doe, got %s", found.AccountHolder)
	}
	if len(found.Transactions) != 1 {
		t.Fatalf("expected 1 transaction loaded, got %d", len(found.Transactions))
	}
	if found.Transactions[0].CategoryID == nil || *found.Transactions[0].CategoryID != catID {
		t.Errorf("expected transaction CategoryID to be %v, got %v", catID, found.Transactions[0].CategoryID)
	}

	err = repo.Save(ctx, entity.BankStatement{UserID: userID, ContentHash: "stmt_hash_1", Currency: "EUR"})
	if !errors.Is(err, service.ErrDuplicate) {
		t.Errorf("expected service.ErrDuplicate, got: %v", err)
	}
}

func TestBankStatementRepository_FindSummaries(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t) // Instant cleanup!

	repo := NewBankStatementRepository(globalPool, setupLogger())

	userID := uuid.New()
	_, err := globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'testuser_summary', 'hash', 'test_summary@example.com')", userID)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	// Use far-future date to safely isolate test
	repo.Save(ctx, entity.BankStatement{
		UserID:        userID,
		ContentHash:   "summary_stmt_1",
		Currency:      "EUR",
		StatementType: entity.StatementType("giro"),
		StatementDate: time.Date(2098, 2, 28, 0, 0, 0, 0, time.UTC),
		NewBalance:    1500.00,
		Transactions: []entity.Transaction{
			{UserID: userID, BookingDate: time.Date(2098, 2, 1, 0, 0, 0, 0, time.UTC), ValutaDate: time.Date(2098, 2, 1, 0, 0, 0, 0, time.UTC), Amount: -500.00, Currency: "EUR", Type: "debit", ContentHash: "tx_feb_1_2098"},
			{UserID: userID, BookingDate: time.Date(2098, 2, 15, 0, 0, 0, 0, time.UTC), ValutaDate: time.Date(2098, 2, 15, 0, 0, 0, 0, time.UTC), Amount: -100.00, Currency: "EUR", Type: "debit", ContentHash: "tx_feb_2_2098"},
		},
	})

	summaries, err := repo.FindSummaries(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var target *entity.BankStatementSummary
	for _, s := range summaries {
		if s.PeriodLabel == "Feb 2098" && s.NewBalance == 1500.00 {
			target = &s
			break
		}
	}

	if target == nil {
		t.Fatalf("expected to find isolated summary for Feb 2098")
	}
	if target.TransactionCount != 2 {
		t.Errorf("expected 2 transactions, got %d", target.TransactionCount)
	}
}

func TestBankStatementRepository_FindTransactionsAndReconciliationFilter(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t) // Instant cleanup!

	repo := NewBankStatementRepository(globalPool, setupLogger())

	userID := uuid.New()
	_, err := globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'testuser_filter', 'hash', 'test_filter@example.com')", userID)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	// Use 2099 to isolate from other tests that use 2026!
	baseDate := time.Date(2099, 3, 1, 0, 0, 0, 0, time.UTC)

	_ = repo.Save(ctx, entity.BankStatement{
		UserID:        userID,
		ContentHash:   "giro_stmt_2099",
		Currency:      "EUR",
		StatementType: entity.StatementType("giro"),
		Transactions: []entity.Transaction{
			{UserID: userID, Description: "Salary", Amount: 4000.0, Currency: "EUR", ContentHash: "tx_salary_2099", BookingDate: baseDate, ValutaDate: baseDate, Type: "credit"},
			{UserID: userID, Description: "Rent", Amount: -1000.0, Currency: "EUR", ContentHash: "tx_rent_2099", BookingDate: baseDate.Add(24 * time.Hour), ValutaDate: baseDate.Add(24 * time.Hour), Type: "debit"},
		},
	})

	_ = repo.Save(ctx, entity.BankStatement{
		UserID:        userID,
		ContentHash:   "cc_stmt_2099",
		Currency:      "EUR",
		StatementType: entity.StatementType("credit_card"),
		Transactions: []entity.Transaction{
			{UserID: userID, Description: "CC Payment Income", Amount: 500.0, Currency: "EUR", ContentHash: "tx_cc_in_2099", BookingDate: baseDate.Add(48 * time.Hour), ValutaDate: baseDate.Add(48 * time.Hour), Type: "credit"},
			{UserID: userID, Description: "Amazon", Amount: -50.0, Currency: "EUR", ContentHash: "tx_cc_out_2099", BookingDate: baseDate.Add(72 * time.Hour), ValutaDate: baseDate.Add(72 * time.Hour), Type: "debit"},
		},
	})

	falseVal := false
	trueVal := true
	fromDate := baseDate.Add(-1 * time.Hour)
	toDate := baseDate.Add(100 * time.Hour)

	tests := []struct {
		name          string
		filter        entity.TransactionFilter
		expectedCount int
	}{
		{"All Unreconciled", entity.TransactionFilter{UserID: userID, FromDate: &fromDate, ToDate: &toDate, IsReconciled: &falseVal}, 4},
		{"All Reconciled", entity.TransactionFilter{UserID: userID, FromDate: &fromDate, ToDate: &toDate, IsReconciled: &trueVal}, 0},
		{"Filter by Type Debit", entity.TransactionFilter{UserID: userID, FromDate: &fromDate, ToDate: &toDate, Type: "debit"}, 2},
		{"Filter by Type Credit", entity.TransactionFilter{UserID: userID, FromDate: &fromDate, ToDate: &toDate, Type: "credit"}, 2},
		{"Search Description ILIKE", entity.TransactionFilter{UserID: userID, FromDate: &fromDate, ToDate: &toDate, Search: "amazon"}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txns, err := repo.FindTransactions(ctx, tt.filter)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(txns) != tt.expectedCount {
				t.Errorf("expected %d transactions, got %d", tt.expectedCount, len(txns))
			}
		})
	}
}

func TestBankStatementRepository_Mutations(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t) // Instant cleanup!

	repo := NewBankStatementRepository(globalPool, setupLogger())

	userID := uuid.New()
	_, err := globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'testuser_mutate', 'hash', 'test_mutate@example.com')", userID)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	catID := uuid.New()
	_, err = globalPool.Exec(ctx, "INSERT INTO categories (id, user_id, name, color) VALUES ($1, $2, 'Einkommen', '#000000')", catID, userID)
	if err != nil {
		t.Fatalf("failed to insert category: %v", err)
	}

	targetHash := "tx_mutate_1"
	stmtDate := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)
	stmt := entity.BankStatement{
		UserID:        userID,
		ContentHash:   "stmt_mutate_1",
		Currency:      "EUR",
		StatementType: entity.StatementType("giro"),
		Transactions: []entity.Transaction{
			{UserID: userID, Description: "Hardware Store", Amount: -100.0, Currency: "EUR", ContentHash: targetHash, IsReconciled: false, BookingDate: stmtDate, ValutaDate: stmtDate, Type: "debit"},
		},
	}
	_ = repo.Save(ctx, stmt)

	t.Run("UpdateTransactionCategory", func(t *testing.T) {
		err := repo.UpdateTransactionCategory(ctx, targetHash, &catID, userID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		var fetchedCatID *uuid.UUID
		err = globalPool.QueryRow(ctx, "SELECT category_id FROM transactions WHERE content_hash = $1 AND user_id = $2", targetHash, userID).Scan(&fetchedCatID)
		if err != nil {
			t.Fatalf("failed to query transaction: %v", err)
		}
		if fetchedCatID == nil || *fetchedCatID != catID {
			t.Errorf("expected category_id to be %v, got %v", catID, fetchedCatID)
		}
	})

	t.Run("ClearTransactionCategory", func(t *testing.T) {
		err := repo.UpdateTransactionCategory(ctx, targetHash, nil, userID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		var categoryID *uuid.UUID
		err = globalPool.QueryRow(ctx, "SELECT category_id FROM transactions WHERE content_hash = $1 AND user_id = $2", targetHash, userID).Scan(&categoryID)
		if err != nil {
			t.Fatalf("failed to query transaction: %v", err)
		}
		if categoryID != nil {
			t.Errorf("expected category_id to be NULL, got %v", categoryID)
		}
	})

	t.Run("MarkTransactionReconciled", func(t *testing.T) {
		recID := uuid.New()

		// 1. Nur EINEN sauberen Insert für das neue 1:1 Mapping ausführen.
		// Wir nutzen einen fiktiven Settlement-Hash und den echten targetHash
		// der "Hardware Store"-Transaktion aus dem äußeren Test-Setup.
		_, err = globalPool.Exec(ctx, `
		INSERT INTO reconciliations (id, user_id, settlement_transaction_hash, target_transaction_hash, amount) 
		VALUES ($1, $2, 'dummy_settlement_hash_for_test', $3, 100.00)`,
			recID, userID, targetHash,
		)
		if err != nil {
			t.Fatalf("failed to insert dummy reconciliation: %v", err)
		}

		// 2. Methode aufrufen
		err = repo.MarkTransactionReconciled(ctx, targetHash, recID, userID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// 3. Validieren
		trueVal := true
		txns, err := repo.FindTransactions(ctx, entity.TransactionFilter{
			UserID:       userID,
			IsReconciled: &trueVal,
			Search:       "Hardware Store",
		})
		if err != nil || len(txns) != 1 {
			t.Fatalf("failed to fetch reconciled transactions, got %d", len(txns))
		}

		if !txns[0].IsReconciled {
			t.Error("expected transaction to be marked as reconciled")
		}
		if txns[0].ReconciliationID == nil || *txns[0].ReconciliationID != recID {
			t.Error("expected reconciliation ID to be assigned")
		}
	})
}

func TestBankStatementRepository_Delete(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t) // Instant cleanup!

	repo := NewBankStatementRepository(globalPool, setupLogger())

	userID := uuid.New()
	_, err := globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'testuser_delete', 'hash', 'test_delete@example.com')", userID)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	stmtDate := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)
	stmt := entity.BankStatement{
		UserID:        userID,
		AccountHolder: "To Be Deleted",
		StatementDate: stmtDate,
		Currency:      "EUR",
		ContentHash:   "stmt_hash_delete_test",
		StatementType: entity.StatementType("giro"),
		Transactions: []entity.Transaction{
			{
				UserID:      userID,
				Description: "Delete Me",
				Amount:      -10.00,
				Currency:    "EUR",
				BookingDate: stmtDate,
				ValutaDate:  stmtDate,
				Type:        "debit",
				ContentHash: "tx_hash_delete_test",
			},
		},
	}

	err = repo.Save(ctx, stmt)
	if err != nil {
		t.Fatalf("expected no error on save, got: %v", err)
	}

	var savedID uuid.UUID
	err = globalPool.QueryRow(ctx, "SELECT id FROM bank_statements WHERE content_hash = 'stmt_hash_delete_test' AND user_id = $1", userID).Scan(&savedID)
	if err != nil {
		t.Fatalf("failed to fetch statement id: %v", err)
	}

	err = repo.Delete(ctx, savedID, userID)
	if err != nil {
		t.Errorf("expected no error on delete, got: %v", err)
	}

	_, err = repo.FindByID(ctx, savedID, userID)
	if err == nil {
		t.Error("expected error when finding deleted statement, got nil")
	}

	var count int
	err = globalPool.QueryRow(ctx, "SELECT count(*) FROM transactions WHERE content_hash = 'tx_hash_delete_test' AND user_id = $1", userID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query transactions count: %v", err)
	}
	if count != 0 {
		t.Error("expected transactions to be cascading deleted, but found them in db")
	}

	err = repo.Delete(ctx, uuid.New(), userID)
	if err == nil {
		t.Error("expected error when deleting non-existent statement, got nil")
	}
}
