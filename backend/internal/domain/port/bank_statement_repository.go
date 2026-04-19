package port

import (
	"cogni-cash/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

// BankStatementRepository defines the storage operations for Bank Statements.
type BankStatementRepository interface {
	Save(ctx context.Context, stmt entity.BankStatement) error
	FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.BankStatement, error)
	FindAll(ctx context.Context, userID uuid.UUID) ([]entity.BankStatement, error)

	FindSummaries(ctx context.Context, userID uuid.UUID) ([]entity.BankStatementSummary, error)
	FindTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error)

	SearchTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error)
	GetCategorizationExamples(ctx context.Context, userID uuid.UUID, examplesCount int) ([]entity.CategorizationExample, error)
	FindMatchingCategory(ctx context.Context, userID uuid.UUID, txn TransactionToCategorize) (*uuid.UUID, error)
	UpdateTransactionCategory(ctx context.Context, hash string, categoryID *uuid.UUID, userID uuid.UUID) error
	UpdateTransactionSubscription(ctx context.Context, contentHash string, subscriptionID *uuid.UUID, userID uuid.UUID) error
	MarkTransactionReconciled(ctx context.Context, contentHash string, reconciliationID uuid.UUID, userID uuid.UUID) error
	MarkTransactionReviewed(ctx context.Context, contentHash string, userID uuid.UUID) error
	UpdateTransactionSkipForecasting(ctx context.Context, contentHash string, skip bool, userID uuid.UUID) error

	LinkTransactionToStatement(ctx context.Context, id uuid.UUID, statementID uuid.UUID, userID uuid.UUID) error

	CreateTransactions(ctx context.Context, txns []entity.Transaction) error

	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}
