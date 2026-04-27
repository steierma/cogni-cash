package mock

import (
	"context"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockAuthUseCase is a mock for port.AuthUseCase
type MockAuthUseCase struct {
	mock.Mock
}

func (m *MockAuthUseCase) Login(ctx context.Context, username, password string) (entity.AuthResponse, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(entity.AuthResponse), args.Error(1)
}

func (m *MockAuthUseCase) Refresh(ctx context.Context, refreshToken string) (entity.AuthResponse, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(entity.AuthResponse), args.Error(1)
}

func (m *MockAuthUseCase) Logout(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthUseCase) ValidateToken(tokenString string) (string, error) {
	args := m.Called(tokenString)
	return args.String(0), args.Error(1)
}

func (m *MockAuthUseCase) ChangePassword(ctx context.Context, userIDStr string, oldPassword, newPassword string) error {
	args := m.Called(ctx, userIDStr, oldPassword, newPassword)
	return args.Error(0)
}

func (m *MockAuthUseCase) EnsureAdminUser(ctx context.Context, username, plainPassword string) error {
	args := m.Called(ctx, username, plainPassword)
	return args.Error(0)
}

func (m *MockAuthUseCase) RequestPasswordReset(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockAuthUseCase) ValidateResetToken(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthUseCase) ConfirmPasswordReset(ctx context.Context, token string, newPassword string) error {
	args := m.Called(ctx, token, newPassword)
	return args.Error(0)
}

// MockReconciliationUseCase is a mock for port.ReconciliationUseCase
type MockReconciliationUseCase struct {
	mock.Mock
}

func (m *MockReconciliationUseCase) ReconcileStatements(ctx context.Context, userID uuid.UUID, settlementTxHash, targetTxHash string) (entity.Reconciliation, error) {
	args := m.Called(ctx, userID, settlementTxHash, targetTxHash)
	return args.Get(0).(entity.Reconciliation), args.Error(1)
}

func (m *MockReconciliationUseCase) SuggestReconciliations(ctx context.Context, userID uuid.UUID, matchWindowDays int) ([]entity.ReconciliationPairSuggestion, error) {
	args := m.Called(ctx, userID, matchWindowDays)
	return args.Get(0).([]entity.ReconciliationPairSuggestion), args.Error(1)
}

func (m *MockReconciliationUseCase) DeleteReconciliation(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

// MockForecastingUseCase is a mock for port.ForecastingUseCase
type MockForecastingUseCase struct {
	mock.Mock
}

func (m *MockForecastingUseCase) GetCashFlowForecast(ctx context.Context, userID uuid.UUID, fromDate, toDate time.Time) (entity.CashFlowForecast, error) {
	args := m.Called(ctx, userID, fromDate, toDate)
	return args.Get(0).(entity.CashFlowForecast), args.Error(1)
}

func (m *MockForecastingUseCase) CalculateCategoryAverage(ctx context.Context, userID, categoryID uuid.UUID, strategy string) (float64, error) {
	args := m.Called(ctx, userID, categoryID, strategy)
	return args.Get(0).(float64), args.Error(1)
}

// MockNotificationUseCase is a mock for port.NotificationUseCase
type MockNotificationUseCase struct {
	mock.Mock
}

func (m *MockNotificationUseCase) SendWelcomeEmail(ctx context.Context, user entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockNotificationUseCase) SendPasswordResetEmail(ctx context.Context, user entity.User, resetURL string) error {
	args := m.Called(ctx, user, resetURL)
	return args.Error(0)
}

func (m *MockNotificationUseCase) SendTestEmail(ctx context.Context, to string, userID uuid.UUID) error {
	args := m.Called(ctx, to, userID)
	return args.Error(0)
}

func (m *MockNotificationUseCase) SendAdminAlert(ctx context.Context, subject, message string) error {
	args := m.Called(ctx, subject, message)
	return args.Error(0)
}

func (m *MockNotificationUseCase) SendBankExpiryWarning(ctx context.Context, user entity.User, connection entity.BankConnection, daysRemaining int) error {
	args := m.Called(ctx, user, connection, daysRemaining)
	return args.Error(0)
}

// MockEmailProvider is a mock for infrastructure providers, although used like a port
// for Mocking NotificationService dependencies
type MockEmailProvider struct {
	mock.Mock
}

func (m *MockEmailProvider) Send(ctx context.Context, userID uuid.UUID, to, subject, body string) error {
	args := m.Called(ctx, userID, to, subject, body)
	return args.Error(0)
}

// MockBankUseCase is a mock for port.BankUseCase
type MockBankUseCase struct {
	mock.Mock
}

func (m *MockBankUseCase) GetInstitutions(ctx context.Context, userID uuid.UUID, countryCode string, isSandbox bool) ([]entity.BankInstitution, error) {
	args := m.Called(ctx, userID, countryCode, isSandbox)
	return args.Get(0).([]entity.BankInstitution), args.Error(1)
}

func (m *MockBankUseCase) CreateConnection(ctx context.Context, userID uuid.UUID, institutionID string, institutionName string, country string, redirectURL string, isSandbox bool, ip string, userAgent string) (*entity.BankConnection, error) {
	args := m.Called(ctx, userID, institutionID, institutionName, country, redirectURL, isSandbox, ip, userAgent)
	return args.Get(0).(*entity.BankConnection), args.Error(1)
}

func (m *MockBankUseCase) RefreshConnection(ctx context.Context, id uuid.UUID, userID uuid.UUID, redirectURL string, isSandbox bool, ip string, userAgent string) (*entity.BankConnection, error) {
	args := m.Called(ctx, id, userID, redirectURL, isSandbox, ip, userAgent)
	return args.Get(0).(*entity.BankConnection), args.Error(1)
}

func (m *MockBankUseCase) FinishConnection(ctx context.Context, userID uuid.UUID, requisitionID string, code string) error {
	args := m.Called(ctx, userID, requisitionID, code)
	return args.Error(0)
}

func (m *MockBankUseCase) GetConnections(ctx context.Context, userID uuid.UUID) ([]entity.BankConnection, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.BankConnection), args.Error(1)
}

func (m *MockBankUseCase) DeleteConnection(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockBankUseCase) CreateVirtualAccount(ctx context.Context, account *entity.BankAccount) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *MockBankUseCase) SyncAccount(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, accountID, userID)
	return args.Error(0)
}

func (m *MockBankUseCase) SyncAllAccounts(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockBankUseCase) UpdateAccountType(ctx context.Context, accountID uuid.UUID, accType entity.StatementType, userID uuid.UUID) error {
	args := m.Called(ctx, accountID, accType, userID)
	return args.Error(0)
}

func (m *MockBankUseCase) ShareAccount(ctx context.Context, accountID, ownerID, sharedWithID uuid.UUID, permission string) error {
	args := m.Called(ctx, accountID, ownerID, sharedWithID, permission)
	return args.Error(0)
}

func (m *MockBankUseCase) RevokeShare(ctx context.Context, accountID, ownerID, sharedWithID uuid.UUID) error {
	args := m.Called(ctx, accountID, ownerID, sharedWithID)
	return args.Error(0)
}

func (m *MockBankUseCase) ListShares(ctx context.Context, accountID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, accountID, ownerID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

// MockDocumentUseCase is a mock for port.DocumentUseCase
type MockDocumentUseCase struct {
	mock.Mock
}

func (m *MockDocumentUseCase) Upload(ctx context.Context, req entity.DocumentUploadRequest) (entity.Document, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(entity.Document), args.Error(1)
}

func (m *MockDocumentUseCase) List(ctx context.Context, filter entity.DocumentFilter) ([]entity.Document, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Document), args.Error(1)
}

func (m *MockDocumentUseCase) GetDetail(ctx context.Context, id, userID uuid.UUID) (entity.Document, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(entity.Document), args.Error(1)
}

func (m *MockDocumentUseCase) Update(ctx context.Context, id, userID uuid.UUID, req entity.DocumentUpdateRequest) (entity.Document, error) {
	args := m.Called(ctx, id, userID, req)
	return args.Get(0).(entity.Document), args.Error(1)
}

func (m *MockDocumentUseCase) Delete(ctx context.Context, id, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockDocumentUseCase) Download(ctx context.Context, id, userID uuid.UUID) ([]byte, string, string, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).([]byte), args.String(1), args.String(2), args.Error(3)
}

func (m *MockDocumentUseCase) GetTaxYearSummary(ctx context.Context, userID uuid.UUID, year int) (entity.TaxYearSummary, error) {
	args := m.Called(ctx, userID, year)
	return args.Get(0).(entity.TaxYearSummary), args.Error(1)
}

// MockUserUseCase is a mock for port.UserUseCase
type MockUserUseCase struct {
	mock.Mock
}

func (m *MockUserUseCase) ListUsers(ctx context.Context, search string) ([]entity.User, error) {
	args := m.Called(ctx, search)
	return args.Get(0).([]entity.User), args.Error(1)
}

func (m *MockUserUseCase) GetUser(ctx context.Context, idStr string) (entity.User, error) {
	args := m.Called(ctx, idStr)
	return args.Get(0).(entity.User), args.Error(1)
}

func (m *MockUserUseCase) CreateUser(ctx context.Context, req entity.User, plainPassword string) (entity.User, error) {
	args := m.Called(ctx, req, plainPassword)
	return args.Get(0).(entity.User), args.Error(1)
}

func (m *MockUserUseCase) UpdateUser(ctx context.Context, idStr string, updates entity.User) (entity.User, error) {
	args := m.Called(ctx, idStr, updates)
	return args.Get(0).(entity.User), args.Error(1)
}

func (m *MockUserUseCase) DeleteUser(ctx context.Context, idStr string) error {
	args := m.Called(ctx, idStr)
	return args.Error(0)
}

func (m *MockUserUseCase) GetAdminID(ctx context.Context) (uuid.UUID, error) {
	args := m.Called(ctx)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

// MockDiscoveryUseCase is a mock for port.DiscoveryUseCase
type MockDiscoveryUseCase struct {
	mock.Mock
}

func (m *MockDiscoveryUseCase) ListSubscriptions(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.Subscription), args.Error(1)
}

func (m *MockDiscoveryUseCase) GetSubscription(ctx context.Context, subID, userID uuid.UUID) (entity.Subscription, error) {
	args := m.Called(ctx, subID, userID)
	return args.Get(0).(entity.Subscription), args.Error(1)
}

func (m *MockDiscoveryUseCase) UpdateSubscription(ctx context.Context, sub entity.Subscription) (entity.Subscription, error) {
	args := m.Called(ctx, sub)
	return args.Get(0).(entity.Subscription), args.Error(1)
}

func (m *MockDiscoveryUseCase) GetSuggestedSubscriptions(ctx context.Context, userID uuid.UUID) ([]entity.SuggestedSubscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.SuggestedSubscription), args.Error(1)
}

func (m *MockDiscoveryUseCase) ApproveSubscription(ctx context.Context, userID uuid.UUID, suggestion entity.SuggestedSubscription) (entity.Subscription, error) {
	args := m.Called(ctx, userID, suggestion)
	return args.Get(0).(entity.Subscription), args.Error(1)
}

func (m *MockDiscoveryUseCase) DeclineSuggestion(ctx context.Context, userID uuid.UUID, merchantName string) error {
	args := m.Called(ctx, userID, merchantName)
	return args.Error(0)
}

func (m *MockDiscoveryUseCase) GetDiscoveryFeedback(ctx context.Context, userID uuid.UUID) ([]entity.DiscoveryFeedback, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.DiscoveryFeedback), args.Error(1)
}

func (m *MockDiscoveryUseCase) RemoveDiscoveryFeedback(ctx context.Context, userID uuid.UUID, merchantName string) error {
	args := m.Called(ctx, userID, merchantName)
	return args.Error(0)
}

func (m *MockDiscoveryUseCase) AllowSuggestion(ctx context.Context, userID uuid.UUID, merchantName string) error {
	args := m.Called(ctx, userID, merchantName)
	return args.Error(0)
}

func (m *MockDiscoveryUseCase) EnrichSubscription(ctx context.Context, userID, subID uuid.UUID) (entity.Subscription, error) {
	args := m.Called(ctx, userID, subID)
	return args.Get(0).(entity.Subscription), args.Error(1)
}

func (m *MockDiscoveryUseCase) CreateSubscriptionFromTransaction(ctx context.Context, userID uuid.UUID, txnHash string, merchantName, billingCycle string, billingInterval int) (entity.Subscription, error) {
	args := m.Called(ctx, userID, txnHash, merchantName, billingCycle, billingInterval)
	return args.Get(0).(entity.Subscription), args.Error(1)
}

func (m *MockDiscoveryUseCase) PreviewCancellation(ctx context.Context, userID, subID uuid.UUID, language string) (port.CancellationLetterResult, error) {
	args := m.Called(ctx, userID, subID, language)
	return args.Get(0).(port.CancellationLetterResult), args.Error(1)
}

func (m *MockDiscoveryUseCase) CancelSubscription(ctx context.Context, userID, subID uuid.UUID, subject, body string) error {
	args := m.Called(ctx, userID, subID, subject, body)
	return args.Error(0)
}

func (m *MockDiscoveryUseCase) DeleteSubscription(ctx context.Context, userID, subID uuid.UUID) error {
	args := m.Called(ctx, userID, subID)
	return args.Error(0)
}

func (m *MockDiscoveryUseCase) GetSubscriptionEvents(ctx context.Context, userID, subID uuid.UUID) ([]entity.SubscriptionEvent, error) {
	args := m.Called(ctx, userID, subID)
	return args.Get(0).([]entity.SubscriptionEvent), args.Error(1)
}

func (m *MockDiscoveryUseCase) MatchTransactions(ctx context.Context, userID uuid.UUID, txns []entity.Transaction) error {
	args := m.Called(ctx, userID, txns)
	return args.Error(0)
}

func (m *MockDiscoveryUseCase) LinkTransaction(ctx context.Context, userID, subID uuid.UUID, txnHash string) error {
	args := m.Called(ctx, userID, subID, txnHash)
	return args.Error(0)
}

func (m *MockDiscoveryUseCase) LinkTransactions(ctx context.Context, userID, subID uuid.UUID, txnHashes []string) error {
	args := m.Called(ctx, userID, subID, txnHashes)
	return args.Error(0)
}

func (m *MockDiscoveryUseCase) UnlinkTransaction(ctx context.Context, userID, subID uuid.UUID, txnHash string) error {
	args := m.Called(ctx, userID, subID, txnHash)
	return args.Error(0)
}

// MockTransactionUseCase is a mock for port.TransactionUseCase
type MockTransactionUseCase struct {
	mock.Mock
}

func (m *MockTransactionUseCase) ListTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Transaction), args.Error(1)
}

func (m *MockTransactionUseCase) GetTransactionAnalytics(ctx context.Context, filter entity.TransactionFilter) (entity.TransactionAnalytics, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(entity.TransactionAnalytics), args.Error(1)
}

func (m *MockTransactionUseCase) UpdateCategory(ctx context.Context, hash string, categoryID *uuid.UUID, userID uuid.UUID) error {
	return m.Called(ctx, hash, categoryID, userID).Error(0)
}

func (m *MockTransactionUseCase) UpdateCategoriesBulk(ctx context.Context, hashes []string, categoryID *uuid.UUID, userID uuid.UUID) error {
	return m.Called(ctx, hashes, categoryID, userID).Error(0)
}

func (m *MockTransactionUseCase) MarkAsReviewed(ctx context.Context, hash string, userID uuid.UUID) error {
	return m.Called(ctx, hash, userID).Error(0)
}

func (m *MockTransactionUseCase) MarkAsReviewedBulk(ctx context.Context, hashes []string, userID uuid.UUID) error {
	return m.Called(ctx, hashes, userID).Error(0)
}

func (m *MockTransactionUseCase) StartAutoCategorizeAsync(ctx context.Context, userID uuid.UUID, batchSize int) error {
	return m.Called(ctx, userID, batchSize).Error(0)
}

func (m *MockTransactionUseCase) GetJobStatus() port.JobState {
	return m.Called().Get(0).(port.JobState)
}

func (m *MockTransactionUseCase) CancelJob() {
	m.Called()
}

// MockBankStatementUseCase is a mock for port.BankStatementUseCase
type MockBankStatementUseCase struct {
	mock.Mock
}

func (m *MockBankStatementUseCase) ImportFromFile(ctx context.Context, userID uuid.UUID, fileName string, fileBytes []byte, useAI bool, statementType entity.StatementType) (entity.BankStatement, error) {
	args := m.Called(ctx, userID, fileName, fileBytes, useAI, statementType)
	return args.Get(0).(entity.BankStatement), args.Error(1)
}

func (m *MockBankStatementUseCase) DeleteStatement(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}

func (m *MockBankStatementUseCase) UpdateStatementAccount(ctx context.Context, id uuid.UUID, accountID *uuid.UUID, userID uuid.UUID) error {
	return m.Called(ctx, id, accountID, userID).Error(0)
}

func (m *MockBankStatementUseCase) GetTransactionsByAccountID(ctx context.Context, accountID, userID uuid.UUID) ([]entity.Transaction, error) {
	args := m.Called(ctx, accountID, userID)
	return args.Get(0).([]entity.Transaction), args.Error(1)
}

// MockPayslipUseCase is a mock for port.PayslipUseCase
type MockPayslipUseCase struct {
	mock.Mock
}

func (m *MockPayslipUseCase) Import(ctx context.Context, userID uuid.UUID, fileName, mimeType string, fileBytes []byte, overrides *entity.Payslip, useAI bool) (*entity.Payslip, error) {
	args := m.Called(ctx, userID, fileName, mimeType, fileBytes, overrides, useAI)
	return args.Get(0).(*entity.Payslip), args.Error(1)
}

func (m *MockPayslipUseCase) GetAll(ctx context.Context, filter entity.PayslipFilter) ([]entity.Payslip, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Payslip), args.Error(1)
}

func (m *MockPayslipUseCase) GetByID(ctx context.Context, id string, userID uuid.UUID) (entity.Payslip, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(entity.Payslip), args.Error(1)
}

func (m *MockPayslipUseCase) Update(ctx context.Context, payslip *entity.Payslip) error {
	return m.Called(ctx, payslip).Error(0)
}

func (m *MockPayslipUseCase) Delete(ctx context.Context, id string, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}

func (m *MockPayslipUseCase) GetOriginalFile(ctx context.Context, id string, userID uuid.UUID) ([]byte, string, string, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).([]byte), args.String(1), args.String(2), args.Error(3)
}

func (m *MockPayslipUseCase) GetSummary(ctx context.Context, userID uuid.UUID) (entity.PayslipSummary, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(entity.PayslipSummary), args.Error(1)
}

// MockSettingsUseCase is a mock for port.SettingsUseCase
type MockSettingsUseCase struct {
	mock.Mock
}

func (m *MockSettingsUseCase) GetAll(ctx context.Context, userID uuid.UUID) (map[string]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockSettingsUseCase) GetAllMasked(ctx context.Context, userID uuid.UUID, isAdmin bool) (map[string]string, error) {
	args := m.Called(ctx, userID, isAdmin)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockSettingsUseCase) Get(ctx context.Context, key string, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, key, userID)
	return args.String(0), args.Error(1)
}

func (m *MockSettingsUseCase) UpdateMultiple(ctx context.Context, settings map[string]string, userID uuid.UUID, isAdmin bool) error {
	args := m.Called(ctx, settings, userID, isAdmin)
	return args.Error(0)
}

// MockInvoiceUseCase is a mock for port.InvoiceUseCase
type MockInvoiceUseCase struct {
	mock.Mock
}

func (m *MockInvoiceUseCase) ImportFromFile(ctx context.Context, userID uuid.UUID, fileName, mimeType string, fileBytes []byte, overrides port.ImportOverrides) (entity.Invoice, error) {
	args := m.Called(ctx, userID, fileName, mimeType, fileBytes, overrides)
	return args.Get(0).(entity.Invoice), args.Error(1)
}

func (m *MockInvoiceUseCase) ImportManual(ctx context.Context, userID uuid.UUID, invoice entity.Invoice) (entity.Invoice, error) {
	args := m.Called(ctx, userID, invoice)
	return args.Get(0).(entity.Invoice), args.Error(1)
}

func (m *MockInvoiceUseCase) CategorizeDocument(ctx context.Context, userID uuid.UUID, rawText string) (entity.Invoice, error) {
	args := m.Called(ctx, userID, rawText)
	return args.Get(0).(entity.Invoice), args.Error(1)
}

func (m *MockInvoiceUseCase) GetAll(ctx context.Context, filter entity.InvoiceFilter) ([]entity.Invoice, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Invoice), args.Error(1)
}

func (m *MockInvoiceUseCase) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Invoice, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(entity.Invoice), args.Error(1)
}

func (m *MockInvoiceUseCase) Update(ctx context.Context, invoice entity.Invoice) (entity.Invoice, error) {
	args := m.Called(ctx, invoice)
	return args.Get(0).(entity.Invoice), args.Error(1)
}

func (m *MockInvoiceUseCase) UpdateCategoriesBulk(ctx context.Context, ids []uuid.UUID, categoryID *uuid.UUID, userID uuid.UUID) error {
	return m.Called(ctx, ids, categoryID, userID).Error(0)
}

func (m *MockInvoiceUseCase) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}

func (m *MockInvoiceUseCase) GetOriginalFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) ([]byte, string, string, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).([]byte), args.String(1), args.String(2), args.Error(3)
}

func (m *MockInvoiceUseCase) ShareInvoice(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID, permission string) error {
	return m.Called(ctx, invoiceID, ownerID, sharedWithID, permission).Error(0)
}

func (m *MockInvoiceUseCase) RevokeInvoiceShare(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID) error {
	return m.Called(ctx, invoiceID, ownerID, sharedWithID).Error(0)
}

func (m *MockInvoiceUseCase) ListInvoiceShares(ctx context.Context, invoiceID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, invoiceID, ownerID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

// --- REPOSITORY MOCKS ---

// MockSubscriptionRepository is a mock for port.SubscriptionRepository
type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (entity.Subscription, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(entity.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Create(ctx context.Context, sub entity.Subscription) (entity.Subscription, error) {
	args := m.Called(ctx, sub)
	return args.Get(0).(entity.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) CreateWithBackfill(ctx context.Context, sub entity.Subscription, matchingHashes []string) (entity.Subscription, error) {
	args := m.Called(ctx, sub, matchingHashes)
	return args.Get(0).(entity.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Update(ctx context.Context, sub entity.Subscription) (entity.Subscription, error) {
	args := m.Called(ctx, sub)
	return args.Get(0).(entity.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) LogEvent(ctx context.Context, event entity.SubscriptionEvent) error {
	return m.Called(ctx, event).Error(0)
}

func (m *MockSubscriptionRepository) GetEvents(ctx context.Context, subID, userID uuid.UUID) ([]entity.SubscriptionEvent, error) {
	args := m.Called(ctx, subID, userID)
	return args.Get(0).([]entity.SubscriptionEvent), args.Error(1)
}

func (m *MockSubscriptionRepository) SetDiscoveryFeedback(ctx context.Context, userID uuid.UUID, merchantName string, status entity.DiscoveryFeedbackStatus, source string) error {
	return m.Called(ctx, userID, merchantName, status, source).Error(0)
}

func (m *MockSubscriptionRepository) GetDiscoveryFeedback(ctx context.Context, userID uuid.UUID) ([]entity.DiscoveryFeedback, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.DiscoveryFeedback), args.Error(1)
}

func (m *MockSubscriptionRepository) DeleteDiscoveryFeedback(ctx context.Context, userID uuid.UUID, merchantName string) error {
	return m.Called(ctx, userID, merchantName).Error(0)
}

// MockUserRepository is a mock for port.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (entity.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entity.User), args.Error(1)
}

func (m *MockUserRepository) GetAdminID(ctx context.Context) (uuid.UUID, error) {
	args := m.Called(ctx)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockUserRepository) FindAll(ctx context.Context, search string) ([]entity.User, error) {
	args := m.Called(ctx, search)
	return args.Get(0).([]entity.User), args.Error(1)
}

func (m *MockUserRepository) Upsert(ctx context.Context, user entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Create(ctx context.Context, user entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, newHash string) error {
	args := m.Called(ctx, userID, newHash)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockBankStatementRepository is a mock for port.BankStatementRepository
type MockBankStatementRepository struct {
	mock.Mock
}

func (m *MockBankStatementRepository) Save(ctx context.Context, stmt entity.BankStatement) error {
	args := m.Called(ctx, stmt)
	return args.Error(0)
}

func (m *MockBankStatementRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (entity.BankStatement, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(entity.BankStatement), args.Error(1)
}

func (m *MockBankStatementRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]entity.BankStatement, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.BankStatement), args.Error(1)
}

func (m *MockBankStatementRepository) FindSummaries(ctx context.Context, userID uuid.UUID) ([]entity.BankStatementSummary, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.BankStatementSummary), args.Error(1)
}

func (m *MockBankStatementRepository) FindTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Transaction), args.Error(1)
}

func (m *MockBankStatementRepository) SearchTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Transaction), args.Error(1)
}

func (m *MockBankStatementRepository) GetCategorizationExamples(ctx context.Context, userID uuid.UUID, limit int) ([]entity.CategorizationExample, error) {
	args := m.Called(ctx, userID, limit)
	return args.Get(0).([]entity.CategorizationExample), args.Error(1)
}

func (m *MockBankStatementRepository) FindMatchingCategory(ctx context.Context, userID uuid.UUID, txn port.TransactionToCategorize) (*uuid.UUID, error) {
	args := m.Called(ctx, userID, txn)
	return args.Get(0).(*uuid.UUID), args.Error(1)
}

func (m *MockBankStatementRepository) UpdateTransactionCategory(ctx context.Context, hash string, categoryID *uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, hash, categoryID, userID)
	return args.Error(0)
}

func (m *MockBankStatementRepository) UpdateTransactionCategoriesBulk(ctx context.Context, hashes []string, categoryID *uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, hashes, categoryID, userID)
	return args.Error(0)
}

func (m *MockBankStatementRepository) UpdateTransactionSubscription(ctx context.Context, hash string, subscriptionID *uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, hash, subscriptionID, userID)
	return args.Error(0)
}

func (m *MockBankStatementRepository) MarkTransactionReconciled(ctx context.Context, hash string, reconciliationID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, hash, reconciliationID, userID)
	return args.Error(0)
}

func (m *MockBankStatementRepository) MarkTransactionReviewed(ctx context.Context, hash string, userID uuid.UUID) error {
	args := m.Called(ctx, hash, userID)
	return args.Error(0)
}

func (m *MockBankStatementRepository) MarkTransactionsReviewedBulk(ctx context.Context, hashes []string, userID uuid.UUID) error {
	args := m.Called(ctx, hashes, userID)
	return args.Error(0)
}

func (m *MockBankStatementRepository) UpdateTransactionBaseAmount(ctx context.Context, hash string, baseAmount float64, baseCurrency string, userID uuid.UUID) error {
	args := m.Called(ctx, hash, baseAmount, baseCurrency, userID)
	return args.Error(0)
}

func (m *MockBankStatementRepository) LinkTransactionToStatement(ctx context.Context, transactionHash, statementID, userID uuid.UUID) error {
	args := m.Called(ctx, transactionHash, statementID, userID)
	return args.Error(0)
}

func (m *MockBankStatementRepository) UpdateStatementAccount(ctx context.Context, statementID uuid.UUID, accountID *uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, statementID, accountID, userID)
	return args.Error(0)
}

func (m *MockBankStatementRepository) GetTransactionsByAccountID(ctx context.Context, accountID, userID uuid.UUID) ([]entity.Transaction, error) {
	args := m.Called(ctx, accountID, userID)
	return args.Get(0).([]entity.Transaction), args.Error(1)
}

func (m *MockBankStatementRepository) CreateTransactions(ctx context.Context, txns []entity.Transaction) error {
	args := m.Called(ctx, txns)
	return args.Error(0)
}

func (m *MockBankStatementRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

// MockSettingsRepository is a mock for port.SettingsRepository
type MockSettingsRepository struct {
	mock.Mock
}

func (m *MockSettingsRepository) Get(ctx context.Context, key string, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, key, userID)
	return args.String(0), args.Error(1)
}

func (m *MockSettingsRepository) GetGlobal(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockSettingsRepository) GetAll(ctx context.Context, userID uuid.UUID) (map[string]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockSettingsRepository) Set(ctx context.Context, key, value string, userID uuid.UUID, isAdmin bool) error {
	return m.Called(ctx, key, value, userID, isAdmin).Error(0)
}

func (m *MockSettingsRepository) UpdateMultiple(ctx context.Context, settings map[string]string, userID uuid.UUID) error {
	args := m.Called(ctx, settings, userID)
	return args.Error(0)
}

// MockSubscriptionEnricher is a mock for port.SubscriptionEnricher
type MockSubscriptionEnricher struct {
	mock.Mock
}

func (m *MockSubscriptionEnricher) Enrich(ctx context.Context, userID uuid.UUID, sub entity.Subscription) (entity.Subscription, error) {
	args := m.Called(ctx, userID, sub)
	return args.Get(0).(entity.Subscription), args.Error(1)
}

func (m *MockSubscriptionEnricher) EnrichSubscription(ctx context.Context, userID uuid.UUID, merchantName string, txnHashes []string, language string) (port.SubscriptionEnrichmentResult, error) {
	args := m.Called(ctx, userID, merchantName, txnHashes, language)
	return args.Get(0).(port.SubscriptionEnrichmentResult), args.Error(1)
}

func (m *MockSubscriptionEnricher) VerifySubscriptionSuggestion(ctx context.Context, userID uuid.UUID, merchantName string, amount float64, currency string, billingCycle string) (bool, error) {
	args := m.Called(ctx, userID, merchantName, amount, currency, billingCycle)
	return args.Bool(0), args.Error(1)
}

func (m *MockSubscriptionEnricher) GuessBillingInterval(ctx context.Context, userID uuid.UUID, txns []entity.Transaction) (string, int, error) {
	args := m.Called(ctx, userID, txns)
	return args.String(0), args.Int(1), args.Error(2)
}

// MockCancellationLetterGenerator is a mock for port.CancellationLetterGenerator
type MockCancellationLetterGenerator struct {
	mock.Mock
}

func (m *MockCancellationLetterGenerator) Generate(ctx context.Context, userID uuid.UUID, sub entity.Subscription, language string) (port.CancellationLetterResult, error) {
	args := m.Called(ctx, userID, sub, language)
	return args.Get(0).(port.CancellationLetterResult), args.Error(1)
}

func (m *MockCancellationLetterGenerator) GenerateCancellationLetter(ctx context.Context, userID uuid.UUID, req port.CancellationLetterRequest) (port.CancellationLetterResult, error) {
	args := m.Called(ctx, userID, req)
	return args.Get(0).(port.CancellationLetterResult), args.Error(1)
}

// MockInvoiceRepository is a mock for port.InvoiceRepository
type MockInvoiceRepository struct {
	mock.Mock
}

func (m *MockInvoiceRepository) Save(ctx context.Context, invoice entity.Invoice) error {
	return m.Called(ctx, invoice).Error(0)
}

func (m *MockInvoiceRepository) Update(ctx context.Context, invoice entity.Invoice) error {
	return m.Called(ctx, invoice).Error(0)
}

func (m *MockInvoiceRepository) UpdateBaseAmount(ctx context.Context, id uuid.UUID, baseAmount float64, baseCurrency string, userID uuid.UUID) error {
	return m.Called(ctx, id, baseAmount, baseCurrency, userID).Error(0)
}

func (m *MockInvoiceRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Invoice, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(entity.Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) FindAll(ctx context.Context, filter entity.InvoiceFilter) ([]entity.Invoice, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) UpdateCategoriesBulk(ctx context.Context, ids []uuid.UUID, categoryID *uuid.UUID, userID uuid.UUID) error {
	return m.Called(ctx, ids, categoryID, userID).Error(0)
}

func (m *MockInvoiceRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}

func (m *MockInvoiceRepository) DeleteSplits(ctx context.Context, invoiceID, userID uuid.UUID) error {
	return m.Called(ctx, invoiceID, userID).Error(0)
}

func (m *MockInvoiceRepository) ExistsByContentHash(ctx context.Context, hash string, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, hash, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockInvoiceRepository) GetOriginalFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) ([]byte, string, string, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).([]byte), args.String(1), args.String(2), args.Error(3)
}
