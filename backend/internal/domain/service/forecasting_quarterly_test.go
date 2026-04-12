package service_test

import (
	"context"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
)

func TestForecastingService_QuarterlyPattern(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	// 1. Setup quarterly transaction template (every 3 months)
	txns := []entity.Transaction{
		{
			UserID:      userID,
			Description: "Quarterly Insurance",
			Amount:      -300.0,
			BookingDate: now.AddDate(0, -9, 0),
			ContentHash: "ins-1",
		},
		{
			UserID:      userID,
			Description: "Quarterly Insurance",
			Amount:      -300.0,
			BookingDate: now.AddDate(0, -6, 0),
			ContentHash: "ins-2",
		},
		{
			UserID:      userID,
			Description: "Quarterly Insurance",
			Amount:      -300.0,
			BookingDate: now.AddDate(0, -3, 0),
			ContentHash: "ins-3",
		},
	}

	repo := &mockForecastingRepo{txns: txns}
	bankRepo := &mockBankRepo{accounts: []entity.BankAccount{{Balance: 5000.0}}}
	catRepo := &mockCategoryRepo{}
	exRepo := &mockExclusionRepo{}

	svc := service.NewForecastingService(repo, bankRepo, catRepo, nil, nil, exRepo, nil)

	// 2. Generate forecast
	from := now
	to := now.AddDate(0, 6, 0) // Should find one in 0 months (now) and one in 3 months
	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("failed to get forecast: %v", err)
	}

	found := false
	for _, p := range forecast.Predictions {
		if p.Description == "Quarterly Insurance" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected to find Quarterly Insurance prediction, but it was not detected")
	}
}
