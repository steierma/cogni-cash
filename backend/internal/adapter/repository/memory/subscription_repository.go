package memory

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

type SubscriptionRepository struct {
	mu            sync.RWMutex
	subscriptions map[uuid.UUID]entity.Subscription
	events        map[uuid.UUID][]entity.SubscriptionEvent
	feedback      map[uuid.UUID][]entity.DiscoveryFeedback
}

func NewSubscriptionRepository() *SubscriptionRepository {
	return &SubscriptionRepository{
		subscriptions: make(map[uuid.UUID]entity.Subscription),
		events:        make(map[uuid.UUID][]entity.SubscriptionEvent),
		feedback:      make(map[uuid.UUID][]entity.DiscoveryFeedback),
	}
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sub, ok := r.subscriptions[id]
	if !ok || sub.UserID != userID {
		return entity.Subscription{}, errors.New("subscription not found")
	}
	return sub, nil
}

func (r *SubscriptionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var subs []entity.Subscription
	for _, sub := range r.subscriptions {
		if sub.UserID == userID {
			subs = append(subs, sub)
		}
	}

	sort.Slice(subs, func(i, j int) bool {
		return subs[i].CreatedAt.After(subs[j].CreatedAt)
	})

	return subs, nil
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub entity.Subscription) (entity.Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()

	r.subscriptions[sub.ID] = sub
	return sub, nil
}

func (r *SubscriptionRepository) CreateWithBackfill(ctx context.Context, sub entity.Subscription, matchingHashes []string) (entity.Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()

	r.subscriptions[sub.ID] = sub

	// For memory adapter, backfilling transactions requires cross-repository interaction 
	// which is typically complex in an in-memory setup without a shared state or transaction boundary.
	// We'll return the subscription for now. In a real memory implementation for testing, 
	// it might need access to the BankStatementRepository.
	return sub, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub entity.Subscription) (entity.Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.subscriptions[sub.ID]; !ok {
		return entity.Subscription{}, errors.New("subscription not found")
	}

	sub.UpdatedAt = time.Now()
	r.subscriptions[sub.ID] = sub
	return sub, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sub, ok := r.subscriptions[id]
	if !ok {
		return errors.New("subscription not found")
	}

	if sub.UserID != userID {
		return errors.New("unauthorized deletion")
	}

	delete(r.subscriptions, id)
	delete(r.events, id)
	return nil
}

func (r *SubscriptionRepository) LogEvent(ctx context.Context, event entity.SubscriptionEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	r.events[event.SubscriptionID] = append(r.events[event.SubscriptionID], event)
	return nil
}

func (r *SubscriptionRepository) GetEvents(ctx context.Context, subID uuid.UUID, userID uuid.UUID) ([]entity.SubscriptionEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	events := r.events[subID]
	var userEvents []entity.SubscriptionEvent
	for _, e := range events {
		if e.UserID == userID {
			userEvents = append(userEvents, e)
		}
	}

	sort.Slice(userEvents, func(i, j int) bool {
		return userEvents[i].CreatedAt.After(userEvents[j].CreatedAt)
	})

	return userEvents, nil
}

func (r *SubscriptionRepository) SetDiscoveryFeedback(ctx context.Context, userID uuid.UUID, merchantName string, status entity.DiscoveryFeedbackStatus, source string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, f := range r.feedback[userID] {
		if f.MerchantName == merchantName {
			r.feedback[userID][i].Status = status
			r.feedback[userID][i].Source = source
			r.feedback[userID][i].UpdatedAt = time.Now()
			return nil
		}
	}

	r.feedback[userID] = append(r.feedback[userID], entity.DiscoveryFeedback{
		UserID:       userID,
		MerchantName: merchantName,
		Status:       status,
		Source:       source,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})
	return nil
}

func (r *SubscriptionRepository) GetDiscoveryFeedback(ctx context.Context, userID uuid.UUID) ([]entity.DiscoveryFeedback, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.feedback[userID], nil
}

func (r *SubscriptionRepository) DeleteDiscoveryFeedback(ctx context.Context, userID uuid.UUID, merchantName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var updated []entity.DiscoveryFeedback
	for _, f := range r.feedback[userID] {
		if f.MerchantName != merchantName {
			updated = append(updated, f)
		}
	}
	r.feedback[userID] = updated
	return nil
}

var _ port.SubscriptionRepository = (*SubscriptionRepository)(nil)
