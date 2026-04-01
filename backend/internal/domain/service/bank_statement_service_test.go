package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
)

func setupLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type mockParser struct {
	result entity.BankStatement
	err    error
}

func (m *mockParser) Parse(_ context.Context, _ uuid.UUID, _ string) (entity.BankStatement, error) {
	return m.result, m.err
}

type mockRepo struct {
	saveErr       error
	saved         []entity.BankStatement
	existingTxns  []entity.Transaction
	linkedTxIDs   []uuid.UUID
	linkedStmtIDs []uuid.UUID
}

func (m *mockRepo) Save(_ context.Context, stmt entity.BankStatement) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, stmt)
	return nil
}

func (m *mockRepo) FindByID(_ context.Context, id uuid.UUID, _ uuid.UUID) (entity.BankStatement, error) {
	if len(m.saved) == 0 {
		return entity.BankStatement{ID: id}, nil
	}
	for _, s := range m.saved {
		if s.ID == id {
			return s, nil
		}
	}
	return entity.BankStatement{}, errors.New("not found")
}

func (m *mockRepo) FindAll(_ context.Context, _ uuid.UUID) ([]entity.BankStatement, error) {
	return m.saved, nil
}

func (m *mockRepo) FindSummaries(_ context.Context, _ uuid.UUID) ([]entity.BankStatementSummary, error) {
	return nil, nil
}

func (m *mockRepo) FindTransactions(_ context.Context, f entity.TransactionFilter) ([]entity.Transaction, error) {
	var result []entity.Transaction
	for _, tx := range m.existingTxns {
		if tx.IsReconciled && (f.IsReconciled == nil || !*f.IsReconciled) {
			continue
		}
		result = append(result, tx)
	}
	return result, nil
}

func (m *mockRepo) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockRepo) SearchTransactions(_ context.Context, _ entity.TransactionFilter) ([]entity.Transaction, error) {
	return m.existingTxns, nil
}

func (m *mockRepo) GetCategorizationExamples(_ context.Context, _ uuid.UUID, _ int) ([]entity.CategorizationExample, error) {
	return nil, nil
}

func (m *mockRepo) UpdateTransactionCategory(_ context.Context, hash string, categoryID *uuid.UUID, _ uuid.UUID) error {
	for i, tx := range m.existingTxns {
		if tx.ContentHash == hash {
			m.existingTxns[i].CategoryID = categoryID
			return nil
		}
	}
	return errors.New("transaction not found")
}

func (m *mockRepo) MarkTransactionReconciled(_ context.Context, _ string, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockRepo) MarkTransactionReviewed(_ context.Context, _ string, _ uuid.UUID) error {
	return nil
}

func (m *mockRepo) LinkTransactionToStatement(_ context.Context, id uuid.UUID, statementID uuid.UUID, _ uuid.UUID) error {
	m.linkedTxIDs = append(m.linkedTxIDs, id)
	m.linkedStmtIDs = append(m.linkedStmtIDs, statementID)
	return nil
}

func (m *mockRepo) CreateTransactions(_ context.Context, txns []entity.Transaction) error {
	m.existingTxns = append(m.existingTxns, txns...)
	return nil
}

type mockCategoryRepo struct {
	saved []entity.Category
	err   error
}

func (m *mockCategoryRepo) Save(_ context.Context, cat entity.Category) (entity.Category, error) {
	if m.err != nil {
		return entity.Category{}, m.err
	}
	for _, existing := range m.saved {
		if existing.Name == cat.Name {
			return existing, nil
		}
	}
	if cat.ID == uuid.Nil {
		cat.ID = uuid.New()
	}
	m.saved = append(m.saved, cat)
	return cat, nil
}

func (m *mockCategoryRepo) FindAll(_ context.Context, _ uuid.UUID) ([]entity.Category, error) {
	return m.saved, nil
}

func (m *mockCategoryRepo) FindByID(_ context.Context, id uuid.UUID, _ uuid.UUID) (entity.Category, error) {
	for _, c := range m.saved {
		if c.ID == id {
			return c, nil
		}
	}
	return entity.Category{}, errors.New("not found")
}

func (m *mockCategoryRepo) Update(_ context.Context, cat entity.Category) (entity.Category, error) {
	return cat, nil
}

func (m *mockCategoryRepo) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

type mockCategorizer struct {
	results []port.CategorizedTransaction
	err     error
	calls   int
}

func (m *mockCategorizer) CategorizeBatch(ctx context.Context, _ uuid.UUID, txns []port.TransactionToCategorize, categories []string, examples []entity.CategorizationExample) ([]port.CategorizedTransaction, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}

	var batchResults []port.CategorizedTransaction
	for _, reqTx := range txns {
		for _, mockRes := range m.results {
			if reqTx.Hash == mockRes.Hash {
				batchResults = append(batchResults, mockRes)
			}
		}
	}

	return batchResults, nil
}

func setupTestBankStatementService(p port.BankStatementParser, repo port.BankStatementRepository) *service.BankStatementService {
	svc := service.NewBankStatementService(repo, setupLogger())
	if p != nil {
		svc.RegisterParser(".pdf", p)
		svc.RegisterParser(".csv", p)
	}
	return svc
}

func tmpFile(t *testing.T, ext string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "test*"+ext)
	if err != nil {
		t.Fatalf("tmpFile: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestBankStatementService_ImportFromDirectory(t *testing.T) {
	tempDir := t.TempDir()

	_ = os.WriteFile(filepath.Join(tempDir, "statement1.pdf"), []byte("dummy"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "statement2.csv"), []byte("dummy"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "ignore_me.txt"), []byte("dummy"), 0644)

	repo := &mockRepo{}

	parser := &mockParser{result: entity.BankStatement{
		AccountHolder: "Test",
		IBAN:          "DE12345678901234567890",
		StatementDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Transactions:  []entity.Transaction{{Amount: 100.0}},
	}}
	svc := setupTestBankStatementService(parser, repo)

	dummyUserID := uuid.New()
	count, errs := svc.ImportFromDirectory(context.Background(), dummyUserID, tempDir)

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if count != 2 {
		t.Errorf("expected 2 files imported, got %d", count)
	}
}

func TestBankStatementService_ImportFromFile_HappyPath(t *testing.T) {
	expected := entity.BankStatement{
		AccountHolder: "Max Mustermann",
		IBAN:          "DE2354586224568642550",
		NewBalance:    3503.82,
		Currency:      "EUR",
		StatementDate: time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
		Transactions: []entity.Transaction{
			{BookingDate: time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC), Amount: 1300.00, Description: "Gutschrift Max Mustermann"},
		},
	}

	dummyUserID := uuid.New()
	svc := setupTestBankStatementService(&mockParser{result: expected}, &mockRepo{})
	stmt, err := svc.ImportFromFile(context.Background(), dummyUserID, tmpFile(t, ".pdf"), false, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stmt.AccountHolder != expected.AccountHolder {
		t.Errorf("expected account holder %q, got %q", expected.AccountHolder, stmt.AccountHolder)
	}
	if stmt.NewBalance != expected.NewBalance {
		t.Errorf("expected new balance %.2f, got %.2f", expected.NewBalance, stmt.NewBalance)
	}
	if len(stmt.Transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(stmt.Transactions))
	}
}

func TestBankStatementService_ImportFromFile_SkipsDuplicates(t *testing.T) {
	existingTx := entity.Transaction{
		BookingDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		Amount:      -50.0,
		Description: "Grocery Store",
	}

	repo := &mockRepo{
		existingTxns: []entity.Transaction{existingTx},
	}

	parsedStmt := entity.BankStatement{
		AccountHolder: "John Doe",
		IBAN:          "DE12345678901234567890",
		StatementDate: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
		Transactions: []entity.Transaction{
			{
				BookingDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				Amount:      -50.0,
				Description: "Grocery Store",
			},
			{
				BookingDate: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
				Amount:      -20.0,
				Description: "Bakery",
			},
		},
	}

	dummyUserID := uuid.New()
	svc := setupTestBankStatementService(&mockParser{result: parsedStmt}, repo)

	stmt, err := svc.ImportFromFile(context.Background(), dummyUserID, tmpFile(t, ".pdf"), false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stmt.Transactions) != 1 {
		t.Fatalf("expected 1 unique transaction, got %d", len(stmt.Transactions))
	}
	if stmt.Transactions[0].Description != "Bakery" {
		t.Errorf("expected Bakery transaction to be kept, got %s", stmt.Transactions[0].Description)
	}
}

func TestBankStatementService_ImportFromFile_DoesNotSkipSameDateAmountDifferentText(t *testing.T) {
	repo := &mockRepo{
		existingTxns: []entity.Transaction{
			{
				BookingDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				Amount:      -50.0,
				Description: "Fuel Station",
				Reference:   "REF-111",
			},
		},
	}

	parsedStmt := entity.BankStatement{
		AccountHolder: "John Doe",
		IBAN:          "DE12345678901234567890",
		StatementDate: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
		Transactions: []entity.Transaction{
			{
				BookingDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				Amount:      -50.0,
				Description: "Grocery Store",
				Reference:   "REF-222",
			},
		},
	}

	dummyUserID := uuid.New()
	svc := setupTestBankStatementService(&mockParser{result: parsedStmt}, repo)

	stmt, err := svc.ImportFromFile(context.Background(), dummyUserID, tmpFile(t, ".pdf"), false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stmt.Transactions) != 1 {
		t.Fatalf("expected transaction to be kept, got %d", len(stmt.Transactions))
	}
	if len(stmt.SkippedTransactions) != 0 {
		t.Fatalf("expected 0 skipped transactions, got %d", len(stmt.SkippedTransactions))
	}
}

func TestBankStatementService_ImportFromFile_SkipsDuplicatesWithNormalizedText(t *testing.T) {
	repo := &mockRepo{
		existingTxns: []entity.Transaction{
			{
				BookingDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				Amount:      -50.0,
				Description: "AMAZON   EU S.A.R.L.",
			},
		},
	}

	parsedStmt := entity.BankStatement{
		AccountHolder: "John Doe",
		IBAN:          "DE12345678901234567890",
		StatementDate: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
		Transactions: []entity.Transaction{
			{
				BookingDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				Amount:      -50.0,
				Description: "amazon eu sarl",
			},
			{
				BookingDate: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
				Amount:      -12.0,
				Description: "Bakery",
			},
		},
	}

	dummyUserID := uuid.New()
	svc := setupTestBankStatementService(&mockParser{result: parsedStmt}, repo)

	stmt, err := svc.ImportFromFile(context.Background(), dummyUserID, tmpFile(t, ".pdf"), false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stmt.Transactions) != 1 {
		t.Fatalf("expected 1 unique transaction, got %d", len(stmt.Transactions))
	}
	if stmt.Transactions[0].Description != "Bakery" {
		t.Errorf("expected Bakery transaction to be kept, got %s", stmt.Transactions[0].Description)
	}
	if len(stmt.SkippedTransactions) != 1 {
		t.Fatalf("expected 1 skipped transaction, got %d", len(stmt.SkippedTransactions))
	}
}

func TestBankStatementService_DeleteStatement(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewBankStatementService(repo, setupLogger())

	err := svc.DeleteStatement(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error on delete: %v", err)
	}
}

func TestBankStatementService_ImportFromFile_LinksExistingTransactions(t *testing.T) {
	existingTxID := uuid.New()
	repo := &mockRepo{
		existingTxns: []entity.Transaction{
			{
				ID:          existingTxID,
				BookingDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				Amount:      -50.0,
				Description: "AMAZON   EU S.A.R.L.",
			},
		},
	}

	parsedStmt := entity.BankStatement{
		AccountHolder: "John Doe",
		IBAN:          "DE12345678901234567890",
		StatementDate: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
		Transactions: []entity.Transaction{
			{
				BookingDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				Amount:      -50.0,
				Description: "amazon eu sarl", // Fuzzy match
			},
		},
	}

	dummyUserID := uuid.New()
	svc := setupTestBankStatementService(&mockParser{result: parsedStmt}, repo)

	stmt, err := svc.ImportFromFile(context.Background(), dummyUserID, tmpFile(t, ".pdf"), false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stmt.Transactions) != 0 {
		t.Errorf("expected 0 new transactions, got %d", len(stmt.Transactions))
	}
	if len(stmt.SkippedTransactions) != 1 {
		t.Errorf("expected 1 skipped transaction, got %d", len(stmt.SkippedTransactions))
	}

	if len(repo.linkedTxIDs) != 1 {
		t.Fatalf("expected 1 linking call, got %d", len(repo.linkedTxIDs))
	}
	if repo.linkedTxIDs[0] != existingTxID {
		t.Errorf("expected linked tx ID %s, got %s", existingTxID, repo.linkedTxIDs[0])
	}
	if repo.linkedStmtIDs[0] != stmt.ID {
		t.Errorf("expected linked stmt ID %s, got %s", stmt.ID, repo.linkedStmtIDs[0])
	}
}

type mockReconciliationRepo struct {
	saved []entity.Reconciliation
	err   error
}

func (m *mockReconciliationRepo) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return m.err
}

func (m *mockReconciliationRepo) Save(_ context.Context, r entity.Reconciliation) (entity.Reconciliation, error) {
	if m.err != nil {
		return entity.Reconciliation{}, m.err
	}
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	m.saved = append(m.saved, r)
	return r, nil
}

func (m *mockReconciliationRepo) FindBySettlementTx(_ context.Context, hash string, _ uuid.UUID) (entity.Reconciliation, error) {
	for _, r := range m.saved {
		if r.SettlementTransactionHash == hash {
			return r, nil
		}
	}
	return entity.Reconciliation{}, errors.New("not found")
}

func (m *mockReconciliationRepo) FindByTargetTx(_ context.Context, hash string, _ uuid.UUID) (entity.Reconciliation, error) {
	for _, r := range m.saved {
		if r.TargetTransactionHash == hash {
			return r, nil
		}
	}
	return entity.Reconciliation{}, errors.New("not found")
}

func (m *mockReconciliationRepo) FindAll(_ context.Context, _ uuid.UUID) ([]entity.Reconciliation, error) {
	return m.saved, nil
}

func TestReconcileStatements_Success(t *testing.T) {
	settlementTx := entity.Transaction{
		BookingDate:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		Amount:        -1994.14,
		ContentHash:   "settlementhash123",
		StatementType: entity.StatementTypeGiro,
	}
	targetTx := entity.Transaction{
		BookingDate:   time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
		Amount:        1994.14,
		ContentHash:   "targethash456",
		StatementType: entity.StatementTypeCreditCard,
	}

	repo := &mockRepo{
		existingTxns: []entity.Transaction{settlementTx, targetTx},
	}
	reconcRepo := &mockReconciliationRepo{}

	svc := service.NewReconciliationService(repo, reconcRepo, setupLogger())

	dummyUserID := uuid.New()
	rec, err := svc.ReconcileStatements(context.Background(), dummyUserID, "settlementhash123", "targethash456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.SettlementTransactionHash != "settlementhash123" {
		t.Errorf("expected settlement hash, got %q", rec.SettlementTransactionHash)
	}
	if rec.TargetTransactionHash != "targethash456" {
		t.Errorf("expected target hash, got %q", rec.TargetTransactionHash)
	}
	if rec.Amount != 1994.14 {
		t.Errorf("expected amount 1994.14, got %f", rec.Amount)
	}
}

func TestReconcileStatements_TransactionNotFound(t *testing.T) {
	repo := &mockRepo{existingTxns: []entity.Transaction{}}
	reconcRepo := &mockReconciliationRepo{}

	svc := service.NewReconciliationService(repo, reconcRepo, setupLogger())

	dummyUserID := uuid.New()
	_, err := svc.ReconcileStatements(context.Background(), dummyUserID, "nonexistenthash", "nonexistenttarget")

	if !errors.Is(err, entity.ErrTransactionNotFound) {
		t.Errorf("expected ErrTransactionNotFound, got %v", err)
	}
}

func TestAnalytics_ExcludesReconciledTx(t *testing.T) {
	catID1 := uuid.New()
	catID2 := uuid.New()

	repo := &mockRepo{
		existingTxns: []entity.Transaction{
			{BookingDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), Amount: 3000.0, CategoryID: &catID1, Description: "Employer"},
			{BookingDate: time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC), Amount: -100.0, CategoryID: &catID2, Description: "Supermarket"},
			{BookingDate: time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC), Amount: -500.0, CategoryID: nil, Description: "Internal Transfer Leg 1", IsReconciled: true},
			{BookingDate: time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC), Amount: 500.0, CategoryID: nil, Description: "Internal Transfer Leg 2", IsReconciled: true},
		},
	}

	catRepo := &mockCategoryRepo{
		saved: []entity.Category{
			{ID: catID1, Name: "Salary", Color: "#00ff00"},
			{ID: catID2, Name: "Groceries", Color: "#ff0000"},
		},
	}

	svc := service.NewTransactionService(repo, catRepo, nil, nil, setupLogger())

	analytics, err := svc.GetTransactionAnalytics(context.Background(), entity.TransactionFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if analytics.TotalExpense != 100.0 {
		t.Errorf("expected TotalExpense 100 (excluding reconciled -500), got %f", analytics.TotalExpense)
	}
	if analytics.TotalIncome != 3000.0 {
		t.Errorf("expected TotalIncome 3000 (excluding reconciled +500), got %f", analytics.TotalIncome)
	}
	if analytics.NetSavings != 2900.0 {
		t.Errorf("expected NetSavings 2900, got %f", analytics.NetSavings)
	}
}

func TestReconciliationService_SuggestReconciliations_FindsMatches(t *testing.T) {
	repo := &mockRepo{
		existingTxns: []entity.Transaction{
			{BookingDate: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC), Amount: -200, ContentHash: "giro1", StatementType: entity.StatementTypeGiro},
			{BookingDate: time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC), Amount: 200, ContentHash: "cc1", StatementType: entity.StatementTypeCreditCard},
		},
	}
	svc := service.NewReconciliationService(repo, nil, setupLogger())

	dummyUserID := uuid.New()
	suggestions, err := svc.SuggestReconciliations(context.Background(), dummyUserID, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}
	if suggestions[0].SourceTransaction.ContentHash != "giro1" || suggestions[0].TargetTransaction.ContentHash != "cc1" {
		t.Errorf("suggestion mismatched hashes")
	}
}

func TestReconciliationService_SuggestReconciliations_OutsideWindow(t *testing.T) {
	repo := &mockRepo{
		existingTxns: []entity.Transaction{
			{BookingDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), Amount: -500, ContentHash: "giro1", StatementType: entity.StatementTypeGiro},
			{BookingDate: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC), Amount: 500, ContentHash: "cc1", StatementType: entity.StatementTypeCreditCard},
		},
	}
	svc := service.NewReconciliationService(repo, nil, setupLogger())

	dummyUserID := uuid.New()
	suggestions, err := svc.SuggestReconciliations(context.Background(), dummyUserID, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(suggestions) != 0 {
		t.Errorf("expected 0 matches outside the 7-day window, got %d", len(suggestions))
	}
}

func TestStartAutoCategorizeAsync_Success(t *testing.T) {
	catGroceries := uuid.New()
	catTransport := uuid.New()

	var pendingTxns []entity.Transaction
	for i := 0; i < 25; i++ {
		pendingTxns = append(pendingTxns, entity.Transaction{
			ContentHash: uuid.NewString(),
			Description: "Test Vendor",
			Amount:      -10.0,
			CategoryID:  nil,
		})
	}

	repo := &mockRepo{existingTxns: pendingTxns}
	catRepo := &mockCategoryRepo{
		saved: []entity.Category{
			{ID: catGroceries, Name: "Groceries"},
			{ID: catTransport, Name: "Transport"},
		},
	}

	llm := &mockCategorizer{
		results: []port.CategorizedTransaction{
			{Hash: pendingTxns[0].ContentHash, Category: "Groceries"},
			{Hash: pendingTxns[1].ContentHash, Category: "Transport"},
			{Hash: pendingTxns[2].ContentHash, Category: "MadeUpCategory"},
		},
	}

	svc := service.NewTransactionService(repo, catRepo, nil, llm, setupLogger())

	dummyUserID := uuid.New()
	err := svc.StartAutoCategorizeAsync(context.Background(), dummyUserID, 10)

	if err != nil {
		t.Fatalf("expected no error starting async job, got: %v", err)
	}

	timeout := time.After(2 * time.Second)
	for {
		select {
		case <-timeout:
			t.Fatal("timed out waiting for background categorization to complete")
		default:
			status := svc.GetJobStatus()
			if status.Status == "completed" || status.Status == "error" || status.Status == "cancelled" {
				goto Validate
			}
			time.Sleep(10 * time.Millisecond)
		}
	}

Validate:
	status := svc.GetJobStatus()
	if status.Status != "completed" {
		t.Errorf("expected job status 'completed', got '%s'", status.Status)
	}

	if len(status.Results) != 2 {
		t.Errorf("expected 2 successfully categorized items, got %d", len(status.Results))
	}

	if status.Processed != 25 {
		t.Errorf("expected 25 processed items, got %d", status.Processed)
	}
}
