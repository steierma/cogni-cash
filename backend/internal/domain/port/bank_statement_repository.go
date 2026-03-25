package port

import (
	"cogni-cash/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

// BankStatementRepository defines the storage operations for Bank Statements.
type BankStatementRepository interface {
	Save(ctx context.Context, stmt entity.BankStatement) error
	FindByID(ctx context.Context, id uuid.UUID) (entity.BankStatement, error)
	FindAll(ctx context.Context) ([]entity.BankStatement, error)

	FindSummaries(ctx context.Context) ([]entity.BankStatementSummary, error)
	FindTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error)

	SearchTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error)
	UpdateTransactionCategory(ctx context.Context, hash string, categoryID *uuid.UUID) error
	MarkTransactionReconciled(ctx context.Context, contentHash string, reconciliationID uuid.UUID) error

	Delete(ctx context.Context, id uuid.UUID) error
}
