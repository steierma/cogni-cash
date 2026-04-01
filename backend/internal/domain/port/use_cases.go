package port

import (
	"cogni-cash/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

// --- Driving-side (use-case) ports ---
// These interfaces define what the HTTP adapter (or any other driving adapter)
// may call on the domain services. They ensure the adapter depends on
// abstractions, not concrete service structs — the hallmark of hexagonal
// architecture.

// AuthUseCase covers authentication and credential management.
type AuthUseCase interface {
	Login(ctx context.Context, username, password string) (string, error)
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
}

// InvoiceUseCase is the full driving-side port for invoice management.
// It covers file import (with duplicate detection), raw-text categorization,
// manual CRUD, and original-file download.
type InvoiceUseCase interface {
	ImportFromFile(ctx context.Context, userID uuid.UUID, filePath, fileName, mimeType string, fileBytes []byte, categoryID *uuid.UUID) (entity.Invoice, error)
	CategorizeDocument(ctx context.Context, userID uuid.UUID, rawText string) (entity.Invoice, error)
	GetAll(ctx context.Context, userID uuid.UUID) ([]entity.Invoice, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Invoice, error)
	Update(ctx context.Context, invoice entity.Invoice) (entity.Invoice, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	GetOriginalFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) ([]byte, string, string, error)
}

// BankStatementUseCase covers file import and deletion.
type BankStatementUseCase interface {
	ImportFromFile(ctx context.Context, userID uuid.UUID, filePath string, useAI bool, userStmtType entity.StatementType) (entity.BankStatement, error)
	DeleteStatement(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

// TransactionUseCase covers analytics and batch categorization.
type TransactionUseCase interface {
	ListTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error)
	GetTransactionAnalytics(ctx context.Context, filter entity.TransactionFilter) (entity.TransactionAnalytics, error)
	UpdateCategory(ctx context.Context, hash string, categoryID *uuid.UUID, userID uuid.UUID) error
	MarkAsReviewed(ctx context.Context, hash string, userID uuid.UUID) error

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
	Get(ctx context.Context, key string, userID uuid.UUID) (string, error)
	UpdateMultiple(ctx context.Context, settings map[string]string, userID uuid.UUID) error
}

// PayslipUseCase covers payslip import, update, and deletion.
type PayslipUseCase interface {
	Import(ctx context.Context, userID uuid.UUID, filePath, fileName, mimeType string, fileBytes []byte, overrides *entity.Payslip, useAI bool) (*entity.Payslip, error)
	Update(ctx context.Context, payslip *entity.Payslip) error
	Delete(ctx context.Context, id string, userID uuid.UUID) error
}

type BankUseCase interface {
	// Connections
	GetInstitutions(ctx context.Context, userID uuid.UUID, countryCode string, isSandbox bool) ([]entity.BankInstitution, error)
	CreateConnection(ctx context.Context, userID uuid.UUID, institutionID string, institutionName string, country string, redirectURL string, isSandbox bool) (*entity.BankConnection, error)
	FinishConnection(ctx context.Context, userID uuid.UUID, requisitionID string, code string) error
	GetConnections(ctx context.Context, userID uuid.UUID) ([]entity.BankConnection, error)
	DeleteConnection(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// Sync
	SyncAccount(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) error
	SyncAllAccounts(ctx context.Context, userID uuid.UUID) error
	UpdateAccountType(ctx context.Context, accountID uuid.UUID, accType entity.StatementType, userID uuid.UUID) error
}

// NotificationUseCase covers email and system notifications.
type NotificationUseCase interface {
	SendWelcomeEmail(ctx context.Context, user entity.User) error
	SendPasswordResetEmail(ctx context.Context, user entity.User, resetURL string) error
	SendTestEmail(ctx context.Context, to string, userID uuid.UUID) error
}
