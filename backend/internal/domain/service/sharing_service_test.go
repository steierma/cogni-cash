package service_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"cogni-cash/internal/domain/service"
)

// -- Isolated Mocks for Transaction Service Tests --
// Prefixing with mockTxnSvc to prevent namespace collisions with other test files in service_test

type mockTxnSvcBankStatementRepo struct {
	port.BankStatementRepository
	FindTransactionsFunc          func(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error)
	SearchTransactionsFunc        func(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error)
	UpdateTransactionCategoryFunc func(ctx context.Context, hash string, categoryID *uuid.UUID, userID uuid.UUID) error
	MarkTransactionReviewedFunc   func(ctx context.Context, hash string, userID uuid.UUID) error
	FindMatchingCategoryFunc      func(ctx context.Context, userID uuid.UUID, tx port.TransactionToCategorize) (*uuid.UUID, error)
	GetCategorizationExamplesFunc func(ctx context.Context, userID uuid.UUID, limit int) ([]entity.CategorizationExample, error)
}

func (m *mockTxnSvcBankStatementRepo) FindTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	return m.FindTransactionsFunc(ctx, filter)
}
func (m *mockTxnSvcBankStatementRepo) SearchTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	return m.SearchTransactionsFunc(ctx, filter)
}
func (m *mockTxnSvcBankStatementRepo) UpdateTransactionCategory(ctx context.Context, hash string, categoryID *uuid.UUID, userID uuid.UUID) error {
	return m.UpdateTransactionCategoryFunc(ctx, hash, categoryID, userID)
}
func (m *mockTxnSvcBankStatementRepo) MarkTransactionReviewed(ctx context.Context, hash string, userID uuid.UUID) error {
	return m.MarkTransactionReviewedFunc(ctx, hash, userID)
}
func (m *mockTxnSvcBankStatementRepo) FindMatchingCategory(ctx context.Context, userID uuid.UUID, tx port.TransactionToCategorize) (*uuid.UUID, error) {
	return m.FindMatchingCategoryFunc(ctx, userID, tx)
}
func (m *mockTxnSvcBankStatementRepo) GetCategorizationExamples(ctx context.Context, userID uuid.UUID, limit int) ([]entity.CategorizationExample, error) {
	if m.GetCategorizationExamplesFunc != nil {
		return m.GetCategorizationExamplesFunc(ctx, userID, limit)
	}
	return nil, nil
}

type mockTxnSvcCategoryRepo struct {
	port.CategoryRepository
	FindAllFunc func(ctx context.Context, userID uuid.UUID) ([]entity.Category, error)
}

func (m *mockTxnSvcCategoryRepo) FindAll(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
	return m.FindAllFunc(ctx, userID)
}

type mockTxnSvcCategorizer struct {
	port.TransactionCategorizer
	CategorizeBatchFunc func(ctx context.Context, userID uuid.UUID, txns []port.TransactionToCategorize, categories []string, examples []entity.CategorizationExample) ([]port.CategorizedTransaction, error)
}

func (m *mockTxnSvcCategorizer) CategorizeTransactionsBatch(ctx context.Context, userID uuid.UUID, txns []port.TransactionToCategorize, categories []string, examples []entity.CategorizationExample) ([]port.CategorizedTransaction, error) {
	return m.CategorizeBatchFunc(ctx, userID, txns, categories, examples)
}

// -- Tests --

func TestTransactionService_SimpleOperations(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	userID := uuid.New()
	hash := "test-hash-123"

	t.Run("ListTransactions", func(t *testing.T) {
		repo := &mockTxnSvcBankStatementRepo{
			FindTransactionsFunc: func(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
				return []entity.Transaction{{ContentHash: "h1"}, {ContentHash: "h2"}}, nil
			},
		}
		svc := service.NewTransactionService(repo, nil, nil, nil, logger)

		txns, err := svc.ListTransactions(ctx, entity.TransactionFilter{UserID: userID})
		require.NoError(t, err)
		assert.Len(t, txns, 2)
	})

	t.Run("UpdateCategory", func(t *testing.T) {
		catID := uuid.New()
		called := false
		repo := &mockTxnSvcBankStatementRepo{
			UpdateTransactionCategoryFunc: func(ctx context.Context, h string, cID *uuid.UUID, uID uuid.UUID) error {
				assert.Equal(t, hash, h)
				assert.Equal(t, catID, *cID)
				assert.Equal(t, userID, uID)
				called = true
				return nil
			},
		}
		svc := service.NewTransactionService(repo, nil, nil, nil, logger)

		err := svc.UpdateCategory(ctx, hash, &catID, userID)
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("MarkAsReviewed", func(t *testing.T) {
		called := false
		repo := &mockTxnSvcBankStatementRepo{
			MarkTransactionReviewedFunc: func(ctx context.Context, h string, uID uuid.UUID) error {
				assert.Equal(t, hash, h)
				assert.Equal(t, userID, uID)
				called = true
				return nil
			},
		}
		svc := service.NewTransactionService(repo, nil, nil, nil, logger)

		err := svc.MarkAsReviewed(ctx, hash, userID)
		require.NoError(t, err)
		assert.True(t, called)
	})
}

func TestTransactionService_GetTransactionAnalytics(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	userID := uuid.New()
	catGroceries := uuid.New()
	catSalary := uuid.New()

	repo := &mockTxnSvcBankStatementRepo{
		FindTransactionsFunc: func(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
			t1 := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
			t2 := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)

			return []entity.Transaction{
				{ContentHash: "h1", UserID: userID, Amount: -50.0, CategoryID: &catGroceries, Description: "REWE", BookingDate: t1},
				{ContentHash: "h2", UserID: userID, Amount: -150.0, CategoryID: &catGroceries, Description: "Aldi", BookingDate: t2},
				{ContentHash: "h3", UserID: userID, Amount: 3000.0, CategoryID: &catSalary, Description: "Employer", BookingDate: t1},
				{ContentHash: "h4", UserID: userID, Amount: -20.0, CategoryID: nil, Description: "Unknown Vendor", BookingDate: t2},
			}, nil
		},
	}

	catRepo := &mockTxnSvcCategoryRepo{
		FindAllFunc: func(ctx context.Context, uID uuid.UUID) ([]entity.Category, error) {
			return []entity.Category{
				{ID: catGroceries, Name: "Groceries", Color: "#ff0000"},
				{ID: catSalary, Name: "Salary", Color: "#00ff00"},
			}, nil
		},
	}

	svc := service.NewTransactionService(repo, catRepo, nil, nil, logger)

	filter := entity.TransactionFilter{UserID: userID}
	analytics, err := svc.GetTransactionAnalytics(ctx, filter)

	require.NoError(t, err)

	// Verify Net Savings (3000 income - 220 expense)
	assert.Equal(t, 3000.0, analytics.TotalIncome)
	assert.Equal(t, 220.0, analytics.TotalExpense)
	assert.Equal(t, 2780.0, analytics.NetSavings)

	// Verify Category Totals (sorted by amount)
	require.Len(t, analytics.CategoryTotals, 3) // Groceries, Uncategorized, Salary

	assert.Equal(t, 3000.0, analytics.CategoryTotals[0].Amount)
	assert.Equal(t, catSalary.String(), analytics.CategoryTotals[0].CategoryID)

	assert.Equal(t, 200.0, analytics.CategoryTotals[1].Amount)
	assert.Equal(t, catGroceries.String(), analytics.CategoryTotals[1].CategoryID)

	// Verify Top Merchants
	require.Len(t, analytics.TopMerchants, 3)
	assert.Equal(t, "Aldi", analytics.TopMerchants[0].Merchant)
	assert.Equal(t, 150.0, analytics.TopMerchants[0].Amount)
}

func TestTransactionService_StartAutoCategorizeAsync_HybridMatching(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	userID := uuid.New()
	catGroceries := uuid.New()

	updatedCategories := make(map[string]*uuid.UUID)

	repo := &mockTxnSvcBankStatementRepo{
		SearchTransactionsFunc: func(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
			// Service maps entity.Transaction to port.TransactionToCategorize internally
			return []entity.Transaction{
				{ContentHash: "h1", Description: "REWE", Amount: -10.0},
				{ContentHash: "h2", Description: "New Vendor", Amount: -20.0},
			}, nil
		},
		FindMatchingCategoryFunc: func(ctx context.Context, uID uuid.UUID, tx port.TransactionToCategorize) (*uuid.UUID, error) {
			if tx.Hash == "h1" {
				return &catGroceries, nil
			}
			return nil, errors.New("no high confidence match")
		},
		UpdateTransactionCategoryFunc: func(ctx context.Context, hash string, cID *uuid.UUID, uID uuid.UUID) error {
			updatedCategories[hash] = cID
			return nil
		},
	}

	catRepo := &mockTxnSvcCategoryRepo{
		FindAllFunc: func(ctx context.Context, uID uuid.UUID) ([]entity.Category, error) {
			return []entity.Category{{ID: catGroceries, Name: "Groceries"}}, nil
		},
	}

	llmCalls := 0
	llm := &mockTxnSvcCategorizer{
		CategorizeBatchFunc: func(ctx context.Context, uID uuid.UUID, txns []port.TransactionToCategorize, categories []string, examples []entity.CategorizationExample) ([]port.CategorizedTransaction, error) {
			llmCalls++
			// Should only receive h2, because h1 was matched in DB
			require.Len(t, txns, 1)
			assert.Equal(t, "h2", txns[0].Hash)

			return []port.CategorizedTransaction{
				{Hash: "h2", Category: "Groceries"},
			}, nil
		},
	}

	svc := service.NewTransactionService(repo, catRepo, nil, llm, logger)

	err := svc.StartAutoCategorizeAsync(ctx, userID, 10)
	require.NoError(t, err)

	// Wait for job completion (simple polling for test purposes)
	require.Eventually(t, func() bool {
		return svc.GetJobStatus().Status == "completed"
	}, 2*time.Second, 10*time.Millisecond, "Categorization job did not complete in time")

	// Verify LLM was only called once
	assert.Equal(t, 1, llmCalls, "LLM should have been called exactly once for the unmatched transaction")

	// Verify updates
	require.Contains(t, updatedCategories, "h1", "h1 should have been updated via DB match")
	assert.Equal(t, catGroceries, *updatedCategories["h1"])

	require.Contains(t, updatedCategories, "h2", "h2 should have been updated via LLM match")
	assert.Equal(t, catGroceries, *updatedCategories["h2"])
}
