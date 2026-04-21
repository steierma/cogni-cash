package service_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
)

func TestForecastingService_Exclusions(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	// 1. Setup recurring transaction template
	templateID := uuid.New()
	txns := []entity.Transaction{
		{
			ID:          templateID,
			UserID:      userID,
			Description: "Recurring Rent",
			Amount: -1000.0, BaseAmount: -1000.0,
			BookingDate: now.AddDate(0, -3, 0),
			Type:        entity.TransactionTypeDebit,
			ContentHash: "rent-1",
		},
		{
			UserID:      userID,
			Description: "Recurring Rent",
			Amount: -1000.0, BaseAmount: -1000.0,
			BookingDate: now.AddDate(0, -2, 0),
			Type:        entity.TransactionTypeDebit,
			ContentHash: "rent-2",
		},
		{
			UserID:      userID,
			Description: "Recurring Rent",
			Amount: -1000.0, BaseAmount: -1000.0,
			BookingDate: now.AddDate(0, -1, 0),
			Type:        entity.TransactionTypeDebit,
			ContentHash: "rent-3",
		},
	}

	repo := &mockForecastingRepo{txns: txns}
	bankRepo := &mockForecastingBankRepo{accounts: []entity.BankAccount{{Balance: 5000.0}}}
	catRepo := &mockCategoryRepo{}
	exRepo := &mockExclusionRepo{}

	svc := service.NewForecastingService(repo, bankRepo, catRepo, nil, nil, exRepo, nil, nil, nil)

	// 2. Generate forecast WITHOUT exclusions
	from := now
	to := now.AddDate(0, 2, 0)
	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("failed to get forecast: %v", err)
	}

	initialCount := len(forecast.Predictions)
	if initialCount < 1 {
		t.Fatal("expected at least one prediction")
	}

	// Capture the ID of the first prediction
	targetForecastID := forecast.Predictions[0].ID

	// 3. Exclude the first prediction
	err = svc.ExcludeForecast(context.Background(), userID, targetForecastID)
	if err != nil {
		t.Fatalf("failed to exclude forecast: %v", err)
	}

	// 4. Generate forecast WITH exclusion
	forecastWithEx, err := svc.GetCashFlowForecast(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("failed to get forecast with exclusion: %v", err)
	}

	if len(forecastWithEx.Predictions) != initialCount {
		t.Errorf("expected %d predictions after exclusion (all returned), got %d", initialCount, len(forecastWithEx.Predictions))
	}

	foundExcluded := false
	for _, p := range forecastWithEx.Predictions {
		if p.ID == targetForecastID && p.SkipForecasting {
			foundExcluded = true
			break
		}
	}
	if !foundExcluded {
		t.Error("expected excluded forecast to be marked with SkipForecasting=true")
	}

	// 5. Re-include the forecast
	err = svc.IncludeForecast(context.Background(), userID, targetForecastID)
	if err != nil {
		t.Fatalf("failed to include forecast: %v", err)
	}

	forecastReIncluded, err := svc.GetCashFlowForecast(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("failed to get forecast after re-inclusion: %v", err)
	}

	if len(forecastReIncluded.Predictions) != initialCount {
		t.Errorf("expected %d predictions after re-inclusion, got %d", initialCount, len(forecastReIncluded.Predictions))
	}
}

func TestForecastingService_SkipHistorical(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	// Setup transactions that WOULD be recurring, but one is marked as SkipForecasting
	txns := []entity.Transaction{
		{
			UserID:      userID,
			Description: "Gym",
			Amount: -50.0, BaseAmount: -50.0,
			BookingDate: now.AddDate(0, -3, 0),
			ContentHash: "gym-1",
		},
		{
			UserID:          userID,
			Description:     "Gym",
			Amount: -50.0, BaseAmount: -50.0,
			BookingDate:     now.AddDate(0, -2, 0),
			ContentHash:     "gym-2",
			SkipForecasting: true, // This should break the pattern detection
		},
		{
			UserID:      userID,
			Description: "Gym",
			Amount: -50.0, BaseAmount: -50.0,
			BookingDate: now.AddDate(0, -1, 0),
			ContentHash: "gym-3",
		},
	}

	repo := &mockForecastingRepo{txns: txns}
	bankRepo := &mockForecastingBankRepo{accounts: []entity.BankAccount{{Balance: 5000.0}}}
	catRepo := &mockCategoryRepo{}
	exRepo := &mockExclusionRepo{}

	svc := service.NewForecastingService(repo, bankRepo, catRepo, nil, nil, exRepo, nil, nil, nil)

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, now, now.AddDate(0, 1, 0))
	if err != nil {
		t.Fatalf("failed to get forecast: %v", err)
	}

	for _, p := range forecast.Predictions {
		if strings.Contains(p.Description, "Gym") {
			t.Error("did not expect Gym prediction because historical pattern was interrupted by SkipForecasting")
		}
	}
}
