package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type minimalMockPtRepo struct {
	planned []entity.PlannedTransaction
}
func (m *minimalMockPtRepo) Create(_ context.Context, _ *entity.PlannedTransaction) error { return nil }
func (m *minimalMockPtRepo) GetByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*entity.PlannedTransaction, error) { return nil, nil }
func (m *minimalMockPtRepo) Update(_ context.Context, _ *entity.PlannedTransaction) error { return nil }
func (m *minimalMockPtRepo) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error { return nil }
func (m *minimalMockPtRepo) FindByUserID(_ context.Context, _ uuid.UUID) ([]entity.PlannedTransaction, error) { return m.planned, nil }
func (m *minimalMockPtRepo) FindPendingByUserID(_ context.Context, _ uuid.UUID) ([]entity.PlannedTransaction, error) { return m.planned, nil }

type minimalMockCatRepo struct{}
func (m *minimalMockCatRepo) Save(_ context.Context, cat entity.Category) (entity.Category, error) { return cat, nil }
func (m *minimalMockCatRepo) Update(_ context.Context, cat entity.Category) (entity.Category, error) { return cat, nil }
func (m *minimalMockCatRepo) FindByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) (entity.Category, error) { return entity.Category{}, nil }
func (m *minimalMockCatRepo) FindAll(_ context.Context, _ uuid.UUID) ([]entity.Category, error) { return []entity.Category{}, nil }
func (m *minimalMockCatRepo) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error { return nil }

type minimalMockBankStmtRepo struct{}
func (m *minimalMockBankStmtRepo) Save(_ context.Context, _ entity.BankStatement) error { return nil }
func (m *minimalMockBankStmtRepo) FindByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) (entity.BankStatement, error) { return entity.BankStatement{}, nil }
func (m *minimalMockBankStmtRepo) FindAll(_ context.Context, _ uuid.UUID) ([]entity.BankStatement, error) { return []entity.BankStatement{}, nil }
func (m *minimalMockBankStmtRepo) FindSummaries(_ context.Context, _ uuid.UUID) ([]entity.BankStatementSummary, error) { return nil, nil }
func (m *minimalMockBankStmtRepo) FindTransactions(_ context.Context, _ entity.TransactionFilter) ([]entity.Transaction, error) { return []entity.Transaction{}, nil }
func (m *minimalMockBankStmtRepo) SearchTransactions(_ context.Context, _ entity.TransactionFilter) ([]entity.Transaction, error) { return nil, nil }
func (m *minimalMockBankStmtRepo) GetCategorizationExamples(_ context.Context, _ uuid.UUID, _ int) ([]entity.CategorizationExample, error) { return nil, nil }
func (m *minimalMockBankStmtRepo) FindMatchingCategory(_ context.Context, _ uuid.UUID, _ port.TransactionToCategorize) (*uuid.UUID, error) { return nil, nil }
func (m *minimalMockBankStmtRepo) UpdateTransactionCategory(_ context.Context, _ string, _ *uuid.UUID, _ uuid.UUID) error { return nil }
func (m *minimalMockBankStmtRepo) UpdateTransactionSubscription(_ context.Context, _ string, _ *uuid.UUID, _ uuid.UUID) error { return nil }
func (m *minimalMockBankStmtRepo) MarkTransactionReconciled(_ context.Context, _ string, _ uuid.UUID, _ uuid.UUID) error { return nil }
func (m *minimalMockBankStmtRepo) MarkTransactionReviewed(_ context.Context, _ string, _ uuid.UUID) error { return nil }
func (m *minimalMockBankStmtRepo) MarkTransactionsReviewedBulk(_ context.Context, _ []string, _ uuid.UUID) error { return nil }
func (m *minimalMockBankStmtRepo) UpdateTransactionBaseAmount(_ context.Context, _ string, _ float64, _ string, _ uuid.UUID) error { return nil }
func (m *minimalMockBankStmtRepo) UpdateStatementAccount(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ uuid.UUID) error { return nil }
func (m *minimalMockBankStmtRepo) GetTransactionsByAccountID(_ context.Context, _ uuid.UUID, _ uuid.UUID) ([]entity.Transaction, error) { return nil, nil }
func (m *minimalMockBankStmtRepo) LinkTransactionToStatement(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ uuid.UUID) error { return nil }
func (m *minimalMockBankStmtRepo) CreateTransactions(_ context.Context, _ []entity.Transaction) error { return nil }
func (m *minimalMockBankStmtRepo) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error { return nil }

func TestGetLastBusinessDay(t *testing.T) {
	loc := time.UTC

	tests := []struct {
		year     int
		month    time.Month
		expected time.Time
	}{
		{2024, time.January, time.Date(2024, time.January, 31, 0, 0, 0, 0, loc)},   // Wed
		{2024, time.February, time.Date(2024, time.February, 29, 0, 0, 0, 0, loc)}, // Thu (Leap year)
		{2024, time.March, time.Date(2024, time.March, 29, 0, 0, 0, 0, loc)},       // Fri (31st is Sun)
		{2024, time.June, time.Date(2024, time.June, 28, 0, 0, 0, 0, loc)},         // Fri (30th is Sun)
		{2027, time.March, time.Date(2027, time.March, 31, 0, 0, 0, 0, loc)},       // Wed
		{2027, time.February, time.Date(2027, time.February, 26, 0, 0, 0, 0, loc)}, // Fri (28th is Sun)
	}

	for _, tt := range tests {
		actual := GetLastBusinessDay(tt.year, tt.month, loc)
		assert.True(t, tt.expected.Equal(actual), "Expected %v, got %v for %d-%02d", tt.expected, actual, tt.year, tt.month)
	}
}

func TestForecastingService_LastBankDayProjection(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ptRepo := &minimalMockPtRepo{}
	catRepo := &minimalMockCatRepo{}
	bankRepo := &minimalMockBankStmtRepo{}
	
	svc := NewForecastingService(bankRepo, nil, catRepo, nil, ptRepo, nil, nil, nil, logger)

	userID := uuid.New()
	fromDate := time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2027, time.April, 1, 0, 0, 0, 0, time.UTC)

	// Setup Planned Transaction with LastBankDay strategy
	pt := entity.PlannedTransaction{
		ID:                 uuid.New(),
		UserID:             userID,
		Description:        "Salary",
		Amount:             4500,
		Currency:           "EUR",
		Date:               time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC),
		IntervalMonths:     1,
		SchedulingStrategy: entity.SchedulingStrategyLastBankDay,
		Status:             entity.PlannedTransactionStatusPending,
	}

	ptRepo.planned = []entity.PlannedTransaction{pt}

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, fromDate, toDate)
	assert.NoError(t, err)

	foundJan := false
	foundFeb := false
	foundMar := false

	for _, p := range forecast.Predictions {
		if p.BookingDate.Year() == 2027 && p.BookingDate.Month() == time.January {
			assert.Equal(t, 29, p.BookingDate.Day(), "Jan 2027 should be 29th")
			foundJan = true
		}
		if p.BookingDate.Year() == 2027 && p.BookingDate.Month() == time.February {
			assert.Equal(t, 26, p.BookingDate.Day(), "Feb 2027 should be 26th")
			foundFeb = true
		}
		if p.BookingDate.Year() == 2027 && p.BookingDate.Month() == time.March {
			assert.Equal(t, 31, p.BookingDate.Day(), "Mar 2027 should be 31st")
			foundMar = true
		}
	}

	assert.True(t, foundJan, "Should have prediction for Jan 2027")
	assert.True(t, foundFeb, "Should have prediction for Feb 2027")
	assert.True(t, foundMar, "Should have prediction for Mar 2027")
}
