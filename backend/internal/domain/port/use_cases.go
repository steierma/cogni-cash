package port

import (
	"cogni-cash/internal/domain/entity"
	"context"
	"time"

	"github.com/google/uuid"
)

// --- Driving-side (use-case) ports ---
// These interfaces define what the HTTP adapter (or any other driving adapter)
// may call on the domain services. They ensure the adapter depends on
// abstractions, not concrete service structs — the hallmark of hexagonal
// architecture.

// AuthUseCase covers authentication and credential management.
type AuthUseCase interface {
	Login(ctx context.Context, username, password string) (entity.AuthResponse, error)
	Refresh(ctx context.Context, refreshToken string) (entity.AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	ValidateToken(tokenString string) (string, error)
	ChangePassword(ctx context.Context, userIDStr string, oldPassword, newPassword string) error
	EnsureAdminUser(ctx context.Context, username, plainPassword string) error

	RequestPasswordReset(ctx context.Context, email string) error
	ValidateResetToken(ctx context.Context, token string) (bool, error)
	ConfirmPasswordReset(ctx context.Context, token string, newPassword string) error
}

// UserUseCase covers CRUD operations on user accounts.
type UserUseCase interface {
	ListUsers(ctx context.Context, search string) ([]entity.User, error)
	GetUser(ctx context.Context, idStr string) (entity.User, error)
	CreateUser(ctx context.Context, req entity.User, plainPassword string) (entity.User, error)
	UpdateUser(ctx context.Context, idStr string, updates entity.User) (entity.User, error)
	DeleteUser(ctx context.Context, idStr string) error
	GetAdminID(ctx context.Context) (uuid.UUID, error)
}

type ImportOverrides struct {
	VendorName *string
	Amount     *float64
	Currency   *string
	IssuedAt   *time.Time
	CategoryID *uuid.UUID
	Splits     []entity.InvoiceSplit
}

// InvoiceUseCase is the full driving-side port for invoice management.
// It covers file import (with duplicate detection), raw-text categorization,
// manual CRUD, and original-file download.
type InvoiceUseCase interface {
	ImportFromFile(ctx context.Context, userID uuid.UUID, fileName, mimeType string, fileBytes []byte, overrides ImportOverrides) (entity.Invoice, error)
	ImportManual(ctx context.Context, userID uuid.UUID, invoice entity.Invoice) (entity.Invoice, error)
	CategorizeDocument(ctx context.Context, userID uuid.UUID, rawText string) (entity.Invoice, error)
	GetAll(ctx context.Context, filter entity.InvoiceFilter) ([]entity.Invoice, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Invoice, error)
	Update(ctx context.Context, invoice entity.Invoice) (entity.Invoice, error)
	UpdateCategoriesBulk(ctx context.Context, ids []uuid.UUID, categoryID *uuid.UUID, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	GetOriginalFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) ([]byte, string, string, error)

	// Sharing
	ShareInvoice(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID, permission string) error
	RevokeInvoiceShare(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID) error
	ListInvoiceShares(ctx context.Context, invoiceID, ownerID uuid.UUID) ([]uuid.UUID, error)
}

// BankStatementUseCase covers file import and deletion.
type BankStatementUseCase interface {
	ImportFromFile(ctx context.Context, userID uuid.UUID, fileName string, fileBytes []byte, useAI bool, userStmtType entity.StatementType) (entity.BankStatement, error)
	DeleteStatement(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	UpdateStatementAccount(ctx context.Context, statementID uuid.UUID, bankAccountID *uuid.UUID, userID uuid.UUID) error
	GetTransactionsByAccountID(ctx context.Context, bankAccountID uuid.UUID, userID uuid.UUID) ([]entity.Transaction, error)
}

// TransactionUseCase covers analytics and batch categorization.
type TransactionUseCase interface {
	ListTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error)
	GetTransactionAnalytics(ctx context.Context, filter entity.TransactionFilter) (entity.TransactionAnalytics, error)
	UpdateCategory(ctx context.Context, hash string, categoryID *uuid.UUID, userID uuid.UUID) error
	UpdateCategoriesBulk(ctx context.Context, hashes []string, categoryID *uuid.UUID, userID uuid.UUID) error
	MarkAsReviewed(ctx context.Context, hash string, userID uuid.UUID) error
	MarkAsReviewedBulk(ctx context.Context, hashes []string, userID uuid.UUID) error

	StartAutoCategorizeAsync(ctx context.Context, userID uuid.UUID, batchSize int) error
	GetJobStatus() JobState
	CancelJob()
}

// ReconciliationUseCase covers reconciliation operations.
type ReconciliationUseCase interface {
	ReconcileStatements(ctx context.Context, userID uuid.UUID, settlementTxHash, targetTxHash string) (entity.Reconciliation, error)
	SuggestReconciliations(ctx context.Context, userID uuid.UUID, matchWindowDays int) ([]entity.ReconciliationPairSuggestion, error)
	DeleteReconciliation(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

// SettingsUseCase covers application settings.
type SettingsUseCase interface {
	GetAll(ctx context.Context, userID uuid.UUID) (map[string]string, error)
	GetAllMasked(ctx context.Context, userID uuid.UUID, isAdmin bool) (map[string]string, error)
	Get(ctx context.Context, key string, userID uuid.UUID) (string, error)
	UpdateMultiple(ctx context.Context, settings map[string]string, userID uuid.UUID, isAdmin bool) error
}

// PayslipUseCase covers payslip import, update, and deletion.
type PayslipUseCase interface {
	Import(ctx context.Context, userID uuid.UUID, fileName, mimeType string, fileBytes []byte, overrides *entity.Payslip, useAI bool) (*entity.Payslip, error)
	GetAll(ctx context.Context, filter entity.PayslipFilter) ([]entity.Payslip, error)
	GetByID(ctx context.Context, id string, userID uuid.UUID) (entity.Payslip, error)
	Update(ctx context.Context, payslip *entity.Payslip) error
	Delete(ctx context.Context, id string, userID uuid.UUID) error
	GetOriginalFile(ctx context.Context, id string, userID uuid.UUID) ([]byte, string, string, error)
	GetSummary(ctx context.Context, userID uuid.UUID) (entity.PayslipSummary, error)
}

type BankUseCase interface {
	// Connections
	GetInstitutions(ctx context.Context, userID uuid.UUID, countryCode string, isSandbox bool) ([]entity.BankInstitution, error)
	CreateConnection(ctx context.Context, userID uuid.UUID, institutionID string, institutionName string, country string, redirectURL string, isSandbox bool, ip string, userAgent string) (*entity.BankConnection, error)
	RefreshConnection(ctx context.Context, id uuid.UUID, userID uuid.UUID, redirectURL string, isSandbox bool, ip string, userAgent string) (*entity.BankConnection, error)
	FinishConnection(ctx context.Context, userID uuid.UUID, requisitionID string, code string) error
	GetConnections(ctx context.Context, userID uuid.UUID) ([]entity.BankConnection, error)
	DeleteConnection(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// Accounts
	CreateVirtualAccount(ctx context.Context, account *entity.BankAccount) error

	// Sync
	SyncAccount(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) error
	SyncAllAccounts(ctx context.Context, userID uuid.UUID) error
	UpdateAccountType(ctx context.Context, accountID uuid.UUID, accType entity.StatementType, userID uuid.UUID) error

	// Sharing
	ShareAccount(ctx context.Context, accountID, ownerID, sharedWithID uuid.UUID, permission string) error
	RevokeShare(ctx context.Context, accountID, ownerID, sharedWithID uuid.UUID) error
	ListShares(ctx context.Context, accountID, ownerID uuid.UUID) ([]uuid.UUID, error)
}

// ForecastingUseCase covers transaction forecasting and cash flow prediction.
type ForecastingUseCase interface {
	GetCashFlowForecast(ctx context.Context, userID uuid.UUID, fromDate, toDate time.Time) (entity.CashFlowForecast, error)

	CalculateCategoryAverage(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID, strategy string) (float64, error)
}

// NotificationUseCase covers email and system notifications.
type NotificationUseCase interface {
	SendWelcomeEmail(ctx context.Context, user entity.User) error
	SendPasswordResetEmail(ctx context.Context, user entity.User, resetURL string) error
	SendTestEmail(ctx context.Context, to string, userID uuid.UUID) error
	SendAdminAlert(ctx context.Context, subject, message string) error
	SendBankExpiryWarning(ctx context.Context, user entity.User, connection entity.BankConnection, daysRemaining int) error
}

// PlannedTransactionUseCase covers the management of user-defined manual transactions.
type PlannedTransactionUseCase interface {
	Create(ctx context.Context, pt *entity.PlannedTransaction) error
	Update(ctx context.Context, pt *entity.PlannedTransaction) error
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error)
	MatchTransactions(ctx context.Context, userID uuid.UUID, txns []entity.Transaction) error
}

// CategoryUseCase covers CRUD operations on categories and their sharing.
type CategoryUseCase interface {
	GetAll(ctx context.Context, userID uuid.UUID) ([]entity.Category, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Category, error)
	Create(ctx context.Context, cat entity.Category) (entity.Category, error)
	Update(ctx context.Context, cat entity.Category) (entity.Category, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// Sharing
	ShareCategory(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID, permission string) error
	RevokeShare(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID) error
	ListShares(ctx context.Context, categoryID, ownerID uuid.UUID) ([]uuid.UUID, error)
}

// SharingUseCase covers the centralized sharing dashboard.
type SharingUseCase interface {
	GetDashboard(ctx context.Context, userID uuid.UUID) (entity.SharingDashboard, error)
}

// DocumentUseCase covers the generic Document Vault management.
type DocumentUseCase interface {
	Upload(ctx context.Context, req entity.DocumentUploadRequest) (entity.Document, error)
	List(ctx context.Context, filter entity.DocumentFilter) ([]entity.Document, error)
	GetDetail(ctx context.Context, id, userID uuid.UUID) (entity.Document, error)
	Update(ctx context.Context, id, userID uuid.UUID, req entity.DocumentUpdateRequest) (entity.Document, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	Download(ctx context.Context, id, userID uuid.UUID) (content []byte, mimeType string, fileName string, err error)
	GetTaxYearSummary(ctx context.Context, userID uuid.UUID, year int) (entity.TaxYearSummary, error)
}

// DiscoveryUseCase covers the identification and approval of recurring subscriptions.
type DiscoveryUseCase interface {
	ListSubscriptions(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, error)
	GetSubscription(ctx context.Context, subID, userID uuid.UUID) (entity.Subscription, error)
	UpdateSubscription(ctx context.Context, sub entity.Subscription) (entity.Subscription, error)
	GetSuggestedSubscriptions(ctx context.Context, userID uuid.UUID) ([]entity.SuggestedSubscription, error)
	ApproveSubscription(ctx context.Context, userID uuid.UUID, suggestion entity.SuggestedSubscription) (entity.Subscription, error)
	DeclineSuggestion(ctx context.Context, userID uuid.UUID, merchantName string) error
	GetDiscoveryFeedback(ctx context.Context, userID uuid.UUID) ([]entity.DiscoveryFeedback, error)
	RemoveDiscoveryFeedback(ctx context.Context, userID uuid.UUID, merchantName string) error
	AllowSuggestion(ctx context.Context, userID uuid.UUID, merchantName string) error
	EnrichSubscription(ctx context.Context, userID, subID uuid.UUID) (entity.Subscription, error)
	CreateSubscriptionFromTransaction(ctx context.Context, userID uuid.UUID, txnHash string, merchantName, billingCycle string, billingInterval int) (entity.Subscription, error)

	// Cancellation
	PreviewCancellation(ctx context.Context, userID, subID uuid.UUID, language string) (CancellationLetterResult, error)
	CancelSubscription(ctx context.Context, userID, subID uuid.UUID, subject, body string) error
	DeleteSubscription(ctx context.Context, userID, subID uuid.UUID) error
	GetSubscriptionEvents(ctx context.Context, userID, subID uuid.UUID) ([]entity.SubscriptionEvent, error)
	MatchTransactions(ctx context.Context, userID uuid.UUID, txns []entity.Transaction) error

	// Manual Linking
	LinkTransaction(ctx context.Context, userID, subID uuid.UUID, txnHash string) error
	LinkTransactions(ctx context.Context, userID, subID uuid.UUID, txnHashes []string) error
	UnlinkTransaction(ctx context.Context, userID, subID uuid.UUID, txnHash string) error
}
