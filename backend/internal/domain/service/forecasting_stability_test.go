package service_test

import (
	"context"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
)

func TestForecastingService_ExclusionStability(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	// 1. Initial history: Jan, Feb, Mar
	marID := uuid.New()
	txns := []entity.Transaction{
		{UserID: userID, Description: "Rent", Amount: -1000.0, BookingDate: now.AddDate(0, -3, 0), ContentHash: "rent-jan"},
		{UserID: userID, Description: "Rent", Amount: -1000.0, BookingDate: now.AddDate(0, -2, 0), ContentHash: "rent-feb"},
		{ID: marID, UserID: userID, Description: "Rent", Amount: -1000.0, BookingDate: now.AddDate(0, -1, 0), ContentHash: "rent-mar"},
	}

	repo := &mockForecastingRepo{txns: txns}
	bankRepo := &mockBankRepo{accounts: []entity.BankAccount{{Balance: 5000.0}}}
	catRepo := &mockCategoryRepo{}
	exRepo := &mockExclusionRepo{}

	svc := service.NewForecastingService(repo, bankRepo, catRepo, nil, nil, exRepo, nil)

	// 2. Forecast May
	from := now.AddDate(0, 2, 0) // May
	to := now.AddDate(0, 2, 0)
	forecast, _ := svc.GetCashFlowForecast(context.Background(), userID, from, to)
	
	if len(forecast.Predictions) != 1 {
		t.Fatalf("expected 1 prediction for May, got %d", len(forecast.Predictions))
	}
	mayRentID := forecast.Predictions[0].ID

	// 3. Exclude May's Rent
	svc.ExcludeForecast(context.Background(), userID, mayRentID)

	// Verify it's marked as excluded
	forecast, _ = svc.GetCashFlowForecast(context.Background(), userID, from, to)
	if len(forecast.Predictions) != 1 || !forecast.Predictions[0].SkipForecasting {
		t.Fatal("May rent should be marked as excluded")
	}

	// 4. Import April's Rent (real transaction)
	aprID := uuid.New()
	txns = append(txns, entity.Transaction{
		ID: aprID, UserID: userID, Description: "Rent", Amount: -1000.0, BookingDate: now, ContentHash: "rent-apr",
	})
	repo.txns = txns // update repo

	// 5. Forecast May AGAIN
	forecast, _ = svc.GetCashFlowForecast(context.Background(), userID, from, to)

	// CRITICAL CHECK: Did it stay marked as excluded?
	foundActive := false
	for _, p := range forecast.Predictions {
		if !p.SkipForecasting {
			foundActive = true
			break
		}
	}
	if foundActive {
		t.Errorf("May rent RE-APPEARED as active because the template ID changed from %s to %s", marID, aprID)
	}
}
