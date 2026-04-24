package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"cogni-cash/internal/domain/service"
)

type mockSharingRepoForService struct {
	mock.Mock
}

func (m *mockSharingRepoForService) ShareCategory(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID, permission string) error {
	return m.Called(ctx, categoryID, ownerID, sharedWithID, permission).Error(0)
}
func (m *mockSharingRepoForService) RevokeShare(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID) error {
	return m.Called(ctx, categoryID, ownerID, sharedWithID).Error(0)
}
func (m *mockSharingRepoForService) ListShares(ctx context.Context, categoryID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, categoryID, ownerID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}
func (m *mockSharingRepoForService) ShareInvoice(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID, permission string) error {
	return m.Called(ctx, invoiceID, ownerID, sharedWithID, permission).Error(0)
}
func (m *mockSharingRepoForService) RevokeInvoiceShare(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID) error {
	return m.Called(ctx, invoiceID, ownerID, sharedWithID).Error(0)
}
func (m *mockSharingRepoForService) ListInvoiceShares(ctx context.Context, invoiceID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, invoiceID, ownerID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}
func (m *mockSharingRepoForService) ShareBankAccount(ctx context.Context, bankAccountID, ownerID, sharedWithID uuid.UUID, permission string) error {
	return m.Called(ctx, bankAccountID, ownerID, sharedWithID, permission).Error(0)
}
func (m *mockSharingRepoForService) RevokeBankAccountShare(ctx context.Context, bankAccountID, ownerID, sharedWithID uuid.UUID) error {
	return m.Called(ctx, bankAccountID, ownerID, sharedWithID).Error(0)
}
func (m *mockSharingRepoForService) ListBankAccountShares(ctx context.Context, bankAccountID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, bankAccountID, ownerID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

type mockUserRepoForSharing struct {
	mock.Mock
}

func (m *mockUserRepoForSharing) FindByUsername(ctx context.Context, username string) (entity.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(entity.User), args.Error(1)
}
func (m *mockUserRepoForSharing) FindByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entity.User), args.Error(1)
}
func (m *mockUserRepoForSharing) GetAdminID(ctx context.Context) (uuid.UUID, error) {
	args := m.Called(ctx)
	return args.Get(0).(uuid.UUID), args.Error(1)
}
func (m *mockUserRepoForSharing) FindAll(ctx context.Context, search string) ([]entity.User, error) {
	args := m.Called(ctx, search)
	return args.Get(0).([]entity.User), args.Error(1)
}
func (m *mockUserRepoForSharing) Create(ctx context.Context, user entity.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepoForSharing) Update(ctx context.Context, user entity.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepoForSharing) Upsert(ctx context.Context, user entity.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepoForSharing) UpdatePassword(ctx context.Context, id uuid.UUID, password string) error {
	return m.Called(ctx, id, password).Error(0)
}
func (m *mockUserRepoForSharing) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// Additional mocks for SharingService
type mockCatRepoForSharing struct{ mock.Mock }
func (m *mockCatRepoForSharing) FindAll(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.Category), args.Error(1)
}
func (m *mockCatRepoForSharing) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Category, error) { return entity.Category{}, nil }
func (m *mockCatRepoForSharing) Create(ctx context.Context, cat *entity.Category) error { return nil }
func (m *mockCatRepoForSharing) Update(ctx context.Context, cat entity.Category) (entity.Category, error) { 
	args := m.Called(ctx, cat)
	return args.Get(0).(entity.Category), args.Error(1)
}
func (m *mockCatRepoForSharing) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error { return nil }
func (m *mockCatRepoForSharing) Save(ctx context.Context, cat entity.Category) (entity.Category, error) { 
	args := m.Called(ctx, cat)
	return args.Get(0).(entity.Category), args.Error(1)
}

type mockInvoiceRepoForSharing struct{ mock.Mock }
func (m *mockInvoiceRepoForSharing) FindAll(ctx context.Context, filter entity.InvoiceFilter) ([]entity.Invoice, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Invoice), args.Error(1)
}
func (m *mockInvoiceRepoForSharing) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Invoice, error) { 
	args := m.Called(ctx, id, userID)
	return args.Get(0).(entity.Invoice), args.Error(1)
}
func (m *mockInvoiceRepoForSharing) Create(ctx context.Context, inv entity.Invoice) (entity.Invoice, error) { 
	args := m.Called(ctx, inv)
	return args.Get(0).(entity.Invoice), args.Error(1)
}
func (m *mockInvoiceRepoForSharing) Update(ctx context.Context, inv entity.Invoice) error { 
	return m.Called(ctx, inv).Error(0)
}
func (m *mockInvoiceRepoForSharing) Save(ctx context.Context, inv entity.Invoice) error {
	return m.Called(ctx, inv).Error(0)
}
func (m *mockInvoiceRepoForSharing) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error { return nil }
func (m *mockInvoiceRepoForSharing) GetOriginalFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) ([]byte, string, string, error) { 
	return nil, "", "", nil 
}
func (m *mockInvoiceRepoForSharing) DeleteSplits(ctx context.Context, invoiceID uuid.UUID, userID uuid.UUID) error { return nil }
func (m *mockInvoiceRepoForSharing) ExistsByContentHash(ctx context.Context, hash string, userID uuid.UUID) (bool, error) { return false, nil }
func (m *mockInvoiceRepoForSharing) UpdateBaseAmount(ctx context.Context, id uuid.UUID, amount float64, currency string, userID uuid.UUID) error { return nil }

type mockBankStmtRepoForSharing struct{ mock.Mock }
func (m *mockBankStmtRepoForSharing) FindTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Transaction), args.Error(1)
}
func (m *mockBankStmtRepoForSharing) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.BankStatement, error) { 
	args := m.Called(ctx, id, userID)
	return args.Get(0).(entity.BankStatement), args.Error(1)
}
func (m *mockBankStmtRepoForSharing) FindAll(ctx context.Context, userID uuid.UUID) ([]entity.BankStatement, error) { 
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.BankStatement), args.Error(1)
}
func (m *mockBankStmtRepoForSharing) FindSummaries(ctx context.Context, userID uuid.UUID) ([]entity.BankStatementSummary, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.BankStatementSummary), args.Error(1)
}
func (m *mockBankStmtRepoForSharing) Create(ctx context.Context, stmt entity.BankStatement) (entity.BankStatement, error) { 
	args := m.Called(ctx, stmt)
	return args.Get(0).(entity.BankStatement), args.Error(1)
}
func (m *mockBankStmtRepoForSharing) Save(ctx context.Context, stmt entity.BankStatement) error {
	return nil
}
func (m *mockBankStmtRepoForSharing) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error { return nil }
func (m *mockBankStmtRepoForSharing) UpdateStatementAccount(ctx context.Context, id uuid.UUID, accountID *uuid.UUID, userID uuid.UUID) error { return nil }
func (m *mockBankStmtRepoForSharing) GetTransactionsByAccountID(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) ([]entity.Transaction, error) { return nil, nil }
func (m *mockBankStmtRepoForSharing) CreateTransactions(ctx context.Context, txns []entity.Transaction) error { return nil }
func (m *mockBankStmtRepoForSharing) FindMatchingCategory(ctx context.Context, userID uuid.UUID, t port.TransactionToCategorize) (*uuid.UUID, error) {
	return nil, nil
}
func (m *mockBankStmtRepoForSharing) GetCategorizationExamples(ctx context.Context, userID uuid.UUID, limit int) ([]entity.CategorizationExample, error) {
	return nil, nil
}
func (m *mockBankStmtRepoForSharing) LinkTransactionToStatement(ctx context.Context, txnID, stmtID, userID uuid.UUID) error {
	return nil
}
func (m *mockBankStmtRepoForSharing) MarkTransactionReconciled(ctx context.Context, txnHash string, reconciliationID uuid.UUID, userID uuid.UUID) error {
	return nil
}
func (m *mockBankStmtRepoForSharing) MarkTransactionReviewed(ctx context.Context, txnHash string, userID uuid.UUID) error {
	return nil
}
func (m *mockBankStmtRepoForSharing) MarkTransactionsReviewedBulk(ctx context.Context, txnHashes []string, userID uuid.UUID) error {
	return nil
}
func (m *mockBankStmtRepoForSharing) SearchTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	return nil, nil
}
func (m *mockBankStmtRepoForSharing) UpdateTransactionBaseAmount(ctx context.Context, txnHash string, baseAmount float64, baseCurrency string, userID uuid.UUID) error {
	return nil
}
func (m *mockBankStmtRepoForSharing) UpdateTransactionCategory(ctx context.Context, txnHash string, categoryID *uuid.UUID, userID uuid.UUID) error {
	return nil
}
func (m *mockBankStmtRepoForSharing) UpdateTransactionSubscription(ctx context.Context, txnHash string, subscriptionID *uuid.UUID, userID uuid.UUID) error {
	return nil
}

func TestSharingService_GetDashboard(t *testing.T) {
	ctx := context.Background()
	catRepo := new(mockCatRepoForSharing)
	invRepo := new(mockInvoiceRepoForSharing)
	sharingRepo := new(mockSharingRepoForService)
	stmtRepo := new(mockBankStmtRepoForSharing)
	userRepo := new(mockUserRepoForSharing)
	
	svc := service.NewSharingService(catRepo, invRepo, sharingRepo, stmtRepo, userRepo, setupLogger())

	userID := uuid.New()
	targetID := uuid.New()

	t.Run("GetDashboard", func(t *testing.T) {
		userRepo.On("FindByID", ctx, targetID).Return(entity.User{ID: targetID, Username: "target"}, nil)
		
		// Setup various shares
		sharingRepo.On("ListShares", ctx, mock.Anything, userID).Return([]uuid.UUID{targetID}, nil)
		sharingRepo.On("ListInvoiceShares", ctx, mock.Anything, userID).Return([]uuid.UUID{targetID}, nil)
		sharingRepo.On("ListBankAccountShares", ctx, mock.Anything, userID).Return([]uuid.UUID{targetID}, nil)

		// Repositories called during dashboard generation
		catRepo.On("FindAll", ctx, userID).Return([]entity.Category{}, nil)
		invRepo.On("FindAll", ctx, mock.Anything).Return([]entity.Invoice{}, nil)
		stmtRepo.On("FindSummaries", ctx, userID).Return([]entity.BankStatementSummary{}, nil)
		stmtRepo.On("FindTransactions", ctx, mock.Anything).Return([]entity.Transaction{}, nil)

		dash, err := svc.GetDashboard(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, dash)
	})
}
