package memory

import (
	"context"
	"sync"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

const maxReconciliations = 500

type ReconciliationRepository struct {
	mu              sync.RWMutex
	reconciliations map[uuid.UUID]entity.Reconciliation
	order           []uuid.UUID
}

func NewReconciliationRepository() *ReconciliationRepository {
	return &ReconciliationRepository{
		reconciliations: make(map[uuid.UUID]entity.Reconciliation),
		order:           make([]uuid.UUID, 0, maxReconciliations),
	}
}

func (r *ReconciliationRepository) Save(ctx context.Context, rec entity.Reconciliation) (entity.Reconciliation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if rec.ID == uuid.Nil {
		rec.ID = uuid.New()
	}

	if _, exists := r.reconciliations[rec.ID]; !exists {
		if len(r.order) >= maxReconciliations {
			// Evict oldest
			oldestID := r.order[0]
			delete(r.reconciliations, oldestID)
			r.order = r.order[1:]
		}
		r.order = append(r.order, rec.ID)
	}

	r.reconciliations[rec.ID] = rec
	return rec, nil
}

func (r *ReconciliationRepository) FindBySettlementTx(ctx context.Context, hash string) (entity.Reconciliation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rec := range r.reconciliations {
		if rec.SettlementTransactionHash == hash {
			return rec, nil
		}
	}
	return entity.Reconciliation{}, entity.ErrReconciliationNotFound
}

func (r *ReconciliationRepository) FindByTargetTx(ctx context.Context, hash string) (entity.Reconciliation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rec := range r.reconciliations {
		if rec.TargetTransactionHash == hash {
			return rec, nil
		}
	}
	return entity.Reconciliation{}, entity.ErrReconciliationNotFound
}

func (r *ReconciliationRepository) FindAll(ctx context.Context) ([]entity.Reconciliation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var recs []entity.Reconciliation
	for _, rec := range r.reconciliations {
		recs = append(recs, rec)
	}
	return recs, nil
}

func (r *ReconciliationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.reconciliations[id]; !ok {
		return entity.ErrReconciliationNotFound
	}
	delete(r.reconciliations, id)
	return nil
}

var _ port.ReconciliationRepository = (*ReconciliationRepository)(nil)
