package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogni-cash/internal/domain/entity"
)

func TestSubscriptionRepository_Integration(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t)

	repo := NewSubscriptionRepository(globalPool, setupLogger())

	userID := uuid.New()
	_, _ = globalPool.Exec(ctx, "INSERT INTO users (id, username, password_hash, email) VALUES ($1, 'sub_user', 'hash', 'sub@example.com')", userID)

	t.Run("Create_and_List", func(t *testing.T) {
		next := time.Now().AddDate(0, 1, 0)
		sub := entity.Subscription{
			ID:              uuid.New(),
			UserID:          userID,
			MerchantName:    "Netflix",
			Amount:          15.99,
			Currency:        "EUR",
			Status:          entity.SubscriptionStatusActive,
			BillingCycle:    "monthly",
			BillingInterval: 1,
			NextOccurrence:  &next,
		}

		savedSub, err := repo.Create(ctx, sub)
		require.NoError(t, err)
		assert.Equal(t, sub.ID, savedSub.ID)

		subs, err := repo.FindByUserID(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, subs, 1)
		assert.Equal(t, "Netflix", subs[0].MerchantName)
	})

	t.Run("Events", func(t *testing.T) {
		subID := uuid.New()
		// Create sub first (FK)
		_, _ = repo.Create(ctx, entity.Subscription{
			ID: subID, UserID: userID, MerchantName: "EventSub", Amount: 1, 
			Currency: "EUR", Status: "active", BillingCycle: "monthly", BillingInterval: 1,
		})

		err := repo.LogEvent(ctx, entity.SubscriptionEvent{
			SubscriptionID: subID,
			UserID:         userID,
			EventType:      "created",
			Title:          "Created",
			Content:        "initial creation",
		})
		assert.NoError(t, err)

		events, err := repo.GetEvents(ctx, subID, userID)
		assert.NoError(t, err)
		assert.Len(t, events, 1)
		assert.Equal(t, "created", events[0].EventType)
	})
}
