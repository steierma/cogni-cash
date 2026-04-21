package service_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"cogni-cash/internal/domain/service"
)

type mockBankSvcRepo struct {
	mock.Mock
}

func (m *mockBankSvcRepo) CreateConnection(ctx context.Context, conn *entity.BankConnection) error {
	args := m.Called(ctx, conn)
	return args.Error(0)
}
func (m *mockBankSvcRepo) GetConnectionsByUserID(ctx context.Context, userID uuid.UUID) ([]entity.BankConnection, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.BankConnection), args.Error(1)
}
func (m *mockBankSvcRepo) GetConnection(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.BankConnection, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.BankConnection), args.Error(1)
}
func (m *mockBankSvcRepo) GetConnectionByRequisition(ctx context.Context, reqID string, userID uuid.UUID) (*entity.BankConnection, error) {
	args := m.Called(ctx, reqID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.BankConnection), args.Error(1)
}
func (m *mockBankSvcRepo) UpdateConnectionStatus(ctx context.Context, id uuid.UUID, status entity.ConnectionStatus, userID uuid.UUID) error {
	args := m.Called(ctx, id, status, userID)
	return args.Error(0)
}
func (m *mockBankSvcRepo) UpdateRequisitionID(ctx context.Context, id uuid.UUID, newReqID string, userID uuid.UUID) error {
	args := m.Called(ctx, id, newReqID, userID)
	return args.Error(0)
}
func (m *mockBankSvcRepo) UpsertAccounts(ctx context.Context, accs []entity.BankAccount, userID uuid.UUID) error {
	args := m.Called(ctx, accs, userID)
	return args.Error(0)
}
func (m *mockBankSvcRepo) GetAccountsByConnectionID(ctx context.Context, connID uuid.UUID, userID uuid.UUID) ([]entity.BankAccount, error) {
	args := m.Called(ctx, connID, userID)
	return args.Get(0).([]entity.BankAccount), args.Error(1)
}
func (m *mockBankSvcRepo) GetAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]entity.BankAccount, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.BankAccount), args.Error(1)
}
func (m *mockBankSvcRepo) GetAccountByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.BankAccount, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.BankAccount), args.Error(1)
}
func (m *mockBankSvcRepo) GetAccountByProviderID(ctx context.Context, providerID string, userID uuid.UUID) (*entity.BankAccount, error) {
	args := m.Called(ctx, providerID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.BankAccount), args.Error(1)
}
func (m *mockBankSvcRepo) UpdateAccountBalance(ctx context.Context, id uuid.UUID, balance float64, lastSync interface{}, errorStr *string, userID uuid.UUID) error {
	args := m.Called(ctx, id, balance, lastSync, errorStr, userID)
	return args.Error(0)
}
func (m *mockBankSvcRepo) UpdateAccountType(ctx context.Context, id uuid.UUID, accType entity.StatementType, userID uuid.UUID) error {
	args := m.Called(ctx, id, accType, userID)
	return args.Error(0)
}
func (m *mockBankSvcRepo) DeleteConnection(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

type mockBankSvcProvider struct {
	mock.Mock
}

func (m *mockBankSvcProvider) GetInstitutions(ctx context.Context, userID uuid.UUID, country string, isSandbox bool) ([]entity.BankInstitution, error) {
	args := m.Called(ctx, userID, country, isSandbox)
	return args.Get(0).([]entity.BankInstitution), args.Error(1)
}
func (m *mockBankSvcProvider) CreateRequisition(ctx context.Context, userID uuid.UUID, instID, instName, country, redirectURL, refID string, isSandbox bool, ip, ua string) (*entity.BankConnection, error) {
	args := m.Called(ctx, userID, instID, instName, country, redirectURL, refID, isSandbox, ip, ua)
	return args.Get(0).(*entity.BankConnection), args.Error(1)
}
func (m *mockBankSvcProvider) ExchangeCodeForSession(ctx context.Context, userID uuid.UUID, code string) (string, error) {
	args := m.Called(ctx, userID, code)
	return args.String(0), args.Error(1)
}
func (m *mockBankSvcProvider) FetchAccounts(ctx context.Context, userID uuid.UUID, sessionID string) ([]entity.BankAccount, error) {
	args := m.Called(ctx, userID, sessionID)
	return args.Get(0).([]entity.BankAccount), args.Error(1)
}
func (m *mockBankSvcProvider) FetchTransactions(ctx context.Context, userID uuid.UUID, accountID string, from, to *time.Time) ([]entity.Transaction, float64, error) {
	args := m.Called(ctx, userID, accountID, from, to)
	return args.Get(0).([]entity.Transaction), args.Get(1).(float64), args.Error(2)
}
func (m *mockBankSvcProvider) GetRequisitionStatus(ctx context.Context, userID uuid.UUID, requisitionID string) (entity.ConnectionStatus, error) {
	args := m.Called(ctx, userID, requisitionID)
	return args.Get(0).(entity.ConnectionStatus), args.Error(1)
}

type mockBankSvcStmtRepo struct {
	mock.Mock
}

func (m *mockBankSvcStmtRepo) CreateTransactions(ctx context.Context, txns []entity.Transaction) error {
	args := m.Called(ctx, txns)
	return args.Error(0)
}
func (m *mockBankSvcStmtRepo) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}
func (m *mockBankSvcStmtRepo) FindAll(ctx context.Context, userID uuid.UUID) ([]entity.BankStatement, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.BankStatement), args.Error(1)
}
func (m *mockBankSvcStmtRepo) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.BankStatement, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(entity.BankStatement), args.Error(1)
}
func (m *mockBankSvcStmtRepo) FindSummaries(ctx context.Context, userID uuid.UUID) ([]entity.BankStatementSummary, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.BankStatementSummary), args.Error(1)
}
func (m *mockBankSvcStmtRepo) FindTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Transaction), args.Error(1)
}
func (m *mockBankSvcStmtRepo) SearchTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Transaction), args.Error(1)
}
func (m *mockBankSvcStmtRepo) Save(ctx context.Context, stmt entity.BankStatement) error {
	args := m.Called(ctx, stmt)
	return args.Error(0)
}
func (m *mockBankSvcStmtRepo) UpdateTransactionCategory(ctx context.Context, hash string, categoryID *uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, hash, categoryID, userID)
	return args.Error(0)
}
func (m *mockBankSvcStmtRepo) UpdateTransactionSubscription(ctx context.Context, hash string, subID *uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, hash, subID, userID)
	return args.Error(0)
}
func (m *mockBankSvcStmtRepo) MarkTransactionReviewed(ctx context.Context, hash string, userID uuid.UUID) error {
	args := m.Called(ctx, hash, userID)
	return args.Error(0)
}
func (m *mockBankSvcStmtRepo) MarkTransactionReconciled(ctx context.Context, hash string, reconID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, hash, reconID, userID)
	return args.Error(0)
}
func (m *mockBankSvcStmtRepo) UpdateTransactionSkipForecasting(ctx context.Context, hash string, skip bool, userID uuid.UUID) error {
	args := m.Called(ctx, hash, skip, userID)
	return args.Error(0)
}
func (m *mockBankSvcStmtRepo) UpdateTransactionBaseAmount(ctx context.Context, hash string, baseAmount float64, baseCurrency string, userID uuid.UUID) error {
	args := m.Called(ctx, hash, baseAmount, baseCurrency, userID)
	return args.Error(0)
}
func (m *mockBankSvcStmtRepo) LinkTransactionToStatement(ctx context.Context, id uuid.UUID, stmtID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, id, stmtID, userID)
	return args.Error(0)
}
func (m *mockBankSvcStmtRepo) GetCategorizationExamples(ctx context.Context, userID uuid.UUID, count int) ([]entity.CategorizationExample, error) {
	args := m.Called(ctx, userID, count)
	return args.Get(0).([]entity.CategorizationExample), args.Error(1)
}
func (m *mockBankSvcStmtRepo) FindMatchingCategory(ctx context.Context, userID uuid.UUID, txn port.TransactionToCategorize) (*uuid.UUID, error) {
	args := m.Called(ctx, userID, txn)
	return args.Get(0).(*uuid.UUID), args.Error(1)
}

func TestBankService(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	userID := uuid.New()

	t.Run("CreateConnection", func(t *testing.T) {
		repo := new(mockBankSvcRepo)
		provider := new(mockBankSvcProvider)
		settings := new(mockSettingsRepo)
		svc := service.NewBankService(repo, nil, settings, nil, provider, logger)

		conn := &entity.BankConnection{ID: uuid.New(), RequisitionID: "req-123", AuthLink: "https://auth.me"}
		provider.On("CreateRequisition", ctx, userID, "inst-1", "Bank 1", "DE", "https://redirect", mock.Anything, false, "1.1.1.1", "ua").Return(conn, nil)
		settings.On("Get", ctx, "bank_provider", userID).Return("enablebanking", nil)
		repo.On("CreateConnection", ctx, mock.Anything).Return(nil)

		result, err := svc.CreateConnection(ctx, userID, "inst-1", "Bank 1", "DE", "https://redirect", false, "1.1.1.1", "ua")
		require.NoError(t, err)
		assert.Equal(t, "req-123", result.RequisitionID)
		repo.AssertExpectations(t)
	})

	t.Run("SyncAllAccounts", func(t *testing.T) {
		repo := new(mockBankSvcRepo)
		provider := new(mockBankSvcProvider)
		settings := new(mockSettingsRepo)
		stmtRepo := new(mockBankSvcStmtRepo)
		svc := service.NewBankService(repo, stmtRepo, settings, nil, provider, logger)

		connID := uuid.New()
		accID := uuid.New()
		repo.On("GetConnectionsByUserID", ctx, userID).Return([]entity.BankConnection{
			{ID: connID, Status: entity.StatusLinked},
		}, nil)
		settings.On("Get", ctx, "bank_sync_history_days", userID).Return("7", nil)
		repo.On("GetAccountsByConnectionID", ctx, connID, userID).Return([]entity.BankAccount{
			{ID: accID, ProviderAccountID: "p-acc-1", IBAN: "DE123"},
		}, nil)

		txns := []entity.Transaction{{Description: "Test", Amount: -10.0}}
		provider.On("FetchTransactions", ctx, userID, "p-acc-1", mock.Anything, mock.Anything).Return(txns, 100.0, nil)
		stmtRepo.On("CreateTransactions", ctx, mock.MatchedBy(func(txs []entity.Transaction) bool {
			return len(txs) == 1 && txs[0].BankAccountID != nil && *txs[0].BankAccountID == accID
		})).Return(nil)
		repo.On("UpdateAccountBalance", ctx, accID, 100.0, mock.Anything, mock.MatchedBy(func(s *string) bool { return s == nil }), userID).Return(nil)

		err := svc.SyncAllAccounts(ctx, userID)
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})
}
