package port

import (
	"cogni-cash/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

type SubscriptionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Subscription, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, error)
	Create(ctx context.Context, sub entity.Subscription) (entity.Subscription, error)
	CreateWithBackfill(ctx context.Context, sub entity.Subscription, matchingHashes []string) (entity.Subscription, error)
	Update(ctx context.Context, sub entity.Subscription) (entity.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// Audit Trail
	LogEvent(ctx context.Context, event entity.SubscriptionEvent) error
	GetEvents(ctx context.Context, subID uuid.UUID, userID uuid.UUID) ([]entity.SubscriptionEvent, error)

	// Discovery Feedback
	SetDiscoveryFeedback(ctx context.Context, userID uuid.UUID, merchantName string, status entity.DiscoveryFeedbackStatus, source string) error
	GetDiscoveryFeedback(ctx context.Context, userID uuid.UUID) ([]entity.DiscoveryFeedback, error)
	DeleteDiscoveryFeedback(ctx context.Context, userID uuid.UUID, merchantName string) error
}
