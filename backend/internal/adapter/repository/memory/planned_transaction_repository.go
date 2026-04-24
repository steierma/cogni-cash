package memory

import (
	"context"
	"sync"
	"time"

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
	r := &PlannedTransactionRepository{
		data:  make(map[uuid.UUID]entity.PlannedTransaction),
		order: make([]uuid.UUID, 0, maxPlannedTransactions),
	}
	r.seedData()
	return r
}

func (r *PlannedTransactionRepository) seedData() {
	userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
	now := time.Now()

	// 1. Annual Car Insurance
	carInsID := uuid.New()
	carIns := entity.PlannedTransaction{
		ID:             carInsID,
		UserID:         userID,
		Amount:         850.00,
		Currency:       "EUR",
		BaseAmount:     850.00,
		BaseCurrency:   "EUR",
		Date:           time.Date(now.Year(), time.November, 15, 0, 0, 0, 0, time.UTC),
		Description:    "Annual Car Insurance (Allianz)",
		Status:         entity.PlannedTransactionStatusPending,
		IntervalMonths: 12,
		CreatedAt:      now.Add(-30 * 24 * time.Hour),
	}

	// 2. Upcoming Vacation Booking
	vacationID := uuid.New()
	vacation := entity.PlannedTransaction{
		ID:             vacationID,
		UserID:         userID,
		Amount:         1500.00,
		Currency:       "EUR",
		BaseAmount:     1500.00,
		BaseCurrency:   "EUR",
		Date:           now.Add(45 * 24 * time.Hour), // 45 days in the future
		Description:    "Summer Vacation Flights & Hotel",
		Status:         entity.PlannedTransactionStatusPending,
		IntervalMonths: 0, // One-off
		CreatedAt:      now.Add(-5 * 24 * time.Hour),
	}

	r.data[carInsID] = carIns
	r.order = append(r.order, carInsID)

	r.data[vacationID] = vacation
	r.order = append(r.order, vacationID)
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
	
	pt.IsShared = pt.UserID != userID
	pt.OwnerID = pt.UserID

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
		return entity.ErrPlannedTransactionNotFound
	}

	delete(r.data, id)

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
		pt.IsShared = pt.UserID != userID
		pt.OwnerID = pt.UserID
		pts = append(pts, pt)
	}
	return pts, nil
}

func (r *PlannedTransactionRepository) FindPendingByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var pts []entity.PlannedTransaction
	for _, id := range r.order {
		pt := r.data[id]
		if pt.Status == entity.PlannedTransactionStatusPending {
			pt.IsShared = pt.UserID != userID
			pt.OwnerID = pt.UserID
			pts = append(pts, pt)
		}
	}
	return pts, nil
}
