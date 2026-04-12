package port

import (
	"context"

	"github.com/google/uuid"

	"cogni-cash/internal/domain/entity"
)

// PlannedTransactionRepository handles the persistence of planned transactions.
type PlannedTransactionRepository interface {
	Create(ctx context.Context, pt *entity.PlannedTransaction) error
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.PlannedTransaction, error)
	Update(ctx context.Context, pt *entity.PlannedTransaction) error
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error)
	FindPendingByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error)
}
