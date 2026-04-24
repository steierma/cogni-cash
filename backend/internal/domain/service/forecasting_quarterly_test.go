package service_test

import (
	"context"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestForecastingService_QuarterlySubscription(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	// Quarterly subscription (every 3 months)
	nextOcc := now.AddDate(0, 1, 0) // next occurrence in 1 month
	subs := []entity.Subscription{
		{
			ID:              uuid.New(),
			UserID:          userID,
			MerchantName:    "Quarterly Insurance",
			Amount:          -300.0,
			Currency:        "EUR",
			BillingCycle:    "monthly",
			BillingInterval: 3, // every 3 months
			Status:          entity.SubscriptionStatusActive,
			NextOccurrence:  &nextOcc,
		},
	}

	repo := &mockForecastingRepo{}
	bankRepo := &mockForecastingBankRepo{accounts: []entity.BankAccount{{Balance: 5000.0}}}
	catRepo := &mockCategoryRepo{}
	subRepo := &mockForecastSubRepo{subs: subs}

	svc := service.NewForecastingService(repo, bankRepo, catRepo, nil, nil, subRepo, nil, nil, nil)

	from := now
	to := now.AddDate(0, 7, 0) // 7 months — should find ~2 occurrences (at month+1 and month+4)

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, from, to)
	assert.NoError(t, err)

	count := 0
	for _, p := range forecast.Predictions {
		if p.Description == "Quarterly Insurance" {
			count++
			assert.Equal(t, -300.0, p.Amount)
		}
	}

	assert.GreaterOrEqual(t, count, 2, "Expected at least 2 quarterly predictions in 7-month window")
}
