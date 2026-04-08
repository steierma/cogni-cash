package port

import (
	"cogni-cash/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

// ReconciliationRepository is the output port for reconciliation persistence.
type ReconciliationRepository interface {
	Save(ctx context.Context, rec entity.Reconciliation) (entity.Reconciliation, error)
	FindBySettlementTx(ctx context.Context, hash string, userID uuid.UUID) (entity.Reconciliation, error)
	FindByTargetTx(ctx context.Context, hash string, userID uuid.UUID) (entity.Reconciliation, error)
	FindAll(ctx context.Context, userID uuid.UUID) ([]entity.Reconciliation, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}
