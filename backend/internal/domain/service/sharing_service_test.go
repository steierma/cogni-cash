package service_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"cogni-cash/internal/domain/service"
)

type mockSharingSvcCategoryRepo struct {
	port.CategoryRepository
	FindAllFunc func(ctx context.Context, userID uuid.UUID) ([]entity.Category, error)
}

func (m *mockSharingSvcCategoryRepo) FindAll(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
	return m.FindAllFunc(ctx, userID)
}

type mockSharingSvcInvoiceRepo struct {
	port.InvoiceRepository
	FindAllFunc func(ctx context.Context, filter entity.InvoiceFilter) ([]entity.Invoice, error)
}

func (m *mockSharingSvcInvoiceRepo) FindAll(ctx context.Context, filter entity.InvoiceFilter) ([]entity.Invoice, error) {
	return m.FindAllFunc(ctx, filter)
}

func (m *mockSharingSvcInvoiceRepo) UpdateBaseAmount(ctx context.Context, id uuid.UUID, baseAmount float64, baseCurrency string, userID uuid.UUID) error {
	return nil
}

type mockSharingSvcBankStmtRepo struct {
	port.BankStatementRepository
	FindTransactionsFunc func(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error)
}

func (m *mockSharingSvcBankStmtRepo) FindTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	return m.FindTransactionsFunc(ctx, filter)
}

func (m *mockSharingSvcBankStmtRepo) UpdateTransactionBaseAmount(ctx context.Context, hash string, baseAmount float64, baseCurrency string, userID uuid.UUID) error {
	return nil
}

type mockSharingSvcUserRepo struct {
	port.UserRepository
	FindByIDFunc func(ctx context.Context, id uuid.UUID) (entity.User, error)
}

func (m *mockSharingSvcUserRepo) FindByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	return m.FindByIDFunc(ctx, id)
}

func TestSharingService_GetDashboard(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	ownerID := uuid.New()
	sharedWithID := uuid.New()

	catID := uuid.New()

	t.Run("Generate dashboard with shared content", func(t *testing.T) {
		catRepo := &mockSharingSvcCategoryRepo{
			FindAllFunc: func(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
				return []entity.Category{
					{ID: catID, UserID: ownerID, Name: "Shared Cat", IsShared: true},
				}, nil
			},
		}

		invRepo := &mockSharingSvcInvoiceRepo{
			FindAllFunc: func(ctx context.Context, filter entity.InvoiceFilter) ([]entity.Invoice, error) {
				return []entity.Invoice{
					{ID: uuid.New(), UserID: ownerID, CategoryID: &catID, Vendor: entity.Vendor{Name: "Shared Invoice"}},
				}, nil
			},
		}

		stmtRepo := &mockSharingSvcBankStmtRepo{
			FindTransactionsFunc: func(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
				return []entity.Transaction{
					{ContentHash: "h1", UserID: ownerID, Amount: -100.0, CategoryID: &catID},
					{ContentHash: "h2", UserID: sharedWithID, Amount: -50.0, CategoryID: &catID},
				}, nil
			},
		}

		userRepo := &mockSharingSvcUserRepo{
			FindByIDFunc: func(ctx context.Context, id uuid.UUID) (entity.User, error) {
				if id == ownerID {
					return entity.User{ID: ownerID, Username: "owner"}, nil
				}
				return entity.User{ID: sharedWithID, Username: "shared"}, nil
			},
		}

		svc := service.NewSharingService(catRepo, invRepo, nil, stmtRepo, userRepo, logger)

		dashboard, err := svc.GetDashboard(ctx, ownerID)
		require.NoError(t, err)

		assert.Len(t, dashboard.SharedCategories, 1)
		assert.Equal(t, "owner", dashboard.SharedCategories[0].Permissions)

		assert.Len(t, dashboard.SharedInvoices, 1)

		require.Len(t, dashboard.Balances, 1)
		assert.Equal(t, -150.0, dashboard.Balances[0].TotalSpent)
		assert.Len(t, dashboard.Balances[0].UserBreakdown, 2)
	})
}
