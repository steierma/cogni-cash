package memory

import (
	"context"
	"sync"

	"cogni-cash/internal/domain/entity"
	"github.com/google/uuid"
)

const maxPlannedTransactions = 1000

type PlannedTransactionRepository struct {
	mu    sync.RWMutex
	data  map[uuid.UUID]entity.PlannedTransaction
	order []uuid.UUID
}

func NewPlannedTransactionRepository() *PlannedTransactionRepository {
	return &PlannedTransactionRepository{
		data:  make(map[uuid.UUID]entity.PlannedTransaction),
		order: make([]uuid.UUID, 0, maxPlannedTransactions),
	}
}

func (r *PlannedTransactionRepository) Create(ctx context.Context, pt *entity.PlannedTransaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if pt.ID == uuid.Nil {
		pt.ID = uuid.New()
	}

	// FIFO Eviction
	if len(r.data) >= maxPlannedTransactions {
		oldestID := r.order[0]
		r.order = r.order[1:]
		delete(r.data, oldestID)
	}

	r.data[pt.ID] = *pt
	r.order = append(r.order, pt.ID)

	return nil
}

func (r *PlannedTransactionRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.PlannedTransaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pt, ok := r.data[id]
	if !ok {
		return nil, entity.ErrPlannedTransactionNotFound
	}
	if pt.UserID != userID {
		return nil, entity.ErrPlannedTransactionNotFound
	}
	return &pt, nil
}

func (r *PlannedTransactionRepository) Update(ctx context.Context, pt *entity.PlannedTransaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[pt.ID]; !ok {
		return entity.ErrPlannedTransactionNotFound
	}
	r.data[pt.ID] = *pt
	return nil
}

func (r *PlannedTransactionRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	pt, ok := r.data[id]
	if !ok {
		return entity.ErrPlannedTransactionNotFound
	}
	if pt.UserID != userID {
		return entity.ErrPlannedTransactionNotFound // simulating Not Found or Forbidden
	}

	delete(r.data, id)

	// Remove from order tracking
	for i, k := range r.order {
		if k == id {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}

	return nil
}

func (r *PlannedTransactionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var pts []entity.PlannedTransaction
	for _, id := range r.order {
		pt := r.data[id]
		if pt.UserID == userID {
			pts = append(pts, pt)
		}
	}
	return pts, nil
}

func (r *PlannedTransactionRepository) FindPendingByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var pts []entity.PlannedTransaction
	for _, id := range r.order {
		pt := r.data[id]
		if pt.UserID == userID && pt.Status == entity.PlannedTransactionStatusPending {
			pts = append(pts, pt)
		}
	}
	return pts, nil
}
