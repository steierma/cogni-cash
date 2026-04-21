package service_test

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockForecastingRepo struct {
	txns []entity.Transaction
}

func (m *mockForecastingRepo) Save(_ context.Context, _ entity.BankStatement) error { return nil }
func (m *mockForecastingRepo) FindByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) (entity.BankStatement, error) {
	return entity.BankStatement{}, nil
}
func (m *mockForecastingRepo) FindAll(_ context.Context, _ uuid.UUID) ([]entity.BankStatement, error) {
	return nil, nil
}
func (m *mockForecastingRepo) FindSummaries(_ context.Context, _ uuid.UUID) ([]entity.BankStatementSummary, error) {
	return nil, nil
}
func (m *mockForecastingRepo) FindTransactions(_ context.Context, _ entity.TransactionFilter) ([]entity.Transaction, error) {
	return m.txns, nil
}
func (m *mockForecastingRepo) SearchTransactions(_ context.Context, _ entity.TransactionFilter) ([]entity.Transaction, error) {
	return m.txns, nil
}
func (m *mockForecastingRepo) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error { return nil }
func (m *mockForecastingRepo) UpdateTransactionCategory(_ context.Context, _ string, _ *uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingRepo) UpdateTransactionSubscription(_ context.Context, _ string, _ *uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingRepo) MarkTransactionReviewed(_ context.Context, _ string, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingRepo) UpdateTransactionSkipForecasting(_ context.Context, _ string, _ bool, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingRepo) UpdateTransactionBaseAmount(_ context.Context, _ string, _ float64, _ string, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingRepo) GetCategorizationExamples(_ context.Context, _ uuid.UUID, _ int) ([]entity.CategorizationExample, error) {
	return nil, nil
}
func (m *mockForecastingRepo) FindMatchingCategory(_ context.Context, _ uuid.UUID, _ port.TransactionToCategorize) (*uuid.UUID, error) {
	return nil, nil
}
func (m *mockForecastingRepo) MarkTransactionReconciled(_ context.Context, _ string, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingRepo) LinkTransactionToStatement(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingRepo) CreateTransactions(_ context.Context, _ []entity.Transaction) error {
	return nil
}

type mockPayslipRepo struct {
	payslips []entity.Payslip
}

func (m *mockPayslipRepo) Save(_ context.Context, _ *entity.Payslip) error { return nil }
func (m *mockPayslipRepo) ExistsByHash(_ context.Context, _ string, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockPayslipRepo) ExistsByOriginalFileName(_ context.Context, _ string, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockPayslipRepo) FindAll(_ context.Context, _ entity.PayslipFilter) ([]entity.Payslip, error) {
	return m.payslips, nil
}
func (m *mockPayslipRepo) FindByID(_ context.Context, _ string, _ uuid.UUID) (entity.Payslip, error) {
	return entity.Payslip{}, nil
}
func (m *mockPayslipRepo) Update(_ context.Context, _ *entity.Payslip) error     { return nil }
func (m *mockPayslipRepo) UpdateBaseAmount(_ context.Context, _ string, _, _, _ float64, _ string, _ uuid.UUID) error {
	return nil
}
func (m *mockPayslipRepo) Delete(_ context.Context, _ string, _ uuid.UUID) error { return nil }
func (m *mockPayslipRepo) GetOriginalFile(_ context.Context, _ string, _ uuid.UUID) ([]byte, string, string, error) {
	return nil, "", "", nil
}
func (m *mockPayslipRepo) GetSummary(_ context.Context, _ uuid.UUID) (entity.PayslipSummary, error) {
	return entity.PayslipSummary{}, nil
}

type mockForecastingBankRepo struct {
	accounts []entity.BankAccount
}

func (m *mockForecastingBankRepo) CreateConnection(_ context.Context, _ *entity.BankConnection) error {
	return nil
}
func (m *mockForecastingBankRepo) GetConnection(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*entity.BankConnection, error) {
	return nil, nil
}
func (m *mockForecastingBankRepo) GetConnectionByRequisition(_ context.Context, _ string, _ uuid.UUID) (*entity.BankConnection, error) {
	return nil, nil
}
func (m *mockForecastingBankRepo) GetConnectionsByUserID(_ context.Context, _ uuid.UUID) ([]entity.BankConnection, error) {
	return nil, nil
}
func (m *mockForecastingBankRepo) UpdateConnectionStatus(_ context.Context, _ uuid.UUID, _ entity.ConnectionStatus, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingBankRepo) UpdateRequisitionID(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingBankRepo) DeleteConnection(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingBankRepo) UpsertAccounts(_ context.Context, _ []entity.BankAccount, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingBankRepo) GetAccountsByUserID(_ context.Context, _ uuid.UUID) ([]entity.BankAccount, error) {
	return m.accounts, nil
}
func (m *mockForecastingBankRepo) GetAccountByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*entity.BankAccount, error) {
	return nil, nil
}
func (m *mockForecastingBankRepo) GetAccountsByConnectionID(_ context.Context, _ uuid.UUID, _ uuid.UUID) ([]entity.BankAccount, error) {
	return nil, nil
}
func (m *mockForecastingBankRepo) GetAccountByProviderID(_ context.Context, _ string, _ uuid.UUID) (*entity.BankAccount, error) {
	return nil, nil
}
func (m *mockForecastingBankRepo) UpdateAccountBalance(_ context.Context, _ uuid.UUID, _ float64, _ interface{}, _ *string, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingBankRepo) UpdateAccountType(_ context.Context, _ uuid.UUID, _ entity.StatementType, _ uuid.UUID) error {
	return nil
}

func TestForecastingService_VariableSpending(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	catID := uuid.New()

	// 1. Setup transactions for a variable category (e.g. Groceries)
	// Month 1: -300
	// Month 2: -500
	// Average: -400
	txns := []entity.Transaction{
		{
			Description: "Groceries 1",
			Amount:      -300.0,
			BaseAmount:  -300.0,
			BookingDate: now.AddDate(0, -2, -15), // ~1.5 months ago
			CategoryID:  &catID,
		},
		{
			Description: "Groceries 2",
			Amount:      -500.0,
			BaseAmount:  -500.0,
			BookingDate: now.AddDate(0, -1, -5), // ~1 month ago
			CategoryID:  &catID,
		},
	}

	repo := &mockForecastingRepo{txns: txns}
	bankRepo := &mockForecastingBankRepo{
		accounts: []entity.BankAccount{
			{Balance: 1000.0},
		},
	}
	catRepo := &mockCategoryRepo{saved: []entity.Category{
		{
			ID:                 catID,
			UserID:             userID,
			Name:               "Groceries",
			IsVariableSpending: true,
		},
	}}

	svc := service.NewForecastingService(repo, bankRepo, catRepo, nil, nil, &mockExclusionRepo{}, nil, nil, nil)

	from := now
	to := now.AddDate(0, 1, 0) // Forecast 1 month ahead

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Average is -400.
	// Since we are in the current month, and no Groceries are yet recorded in the current month (in our mock data),
	// it should forecast -400 for the current month.
	// And if the 'to' date spans into the next month, it might forecast there too.

	foundBudget := false
	for _, p := range forecast.Predictions {
		if strings.Contains(p.Description, "Variable Budget: Groceries") {
			foundBudget = true
			if math.Abs(p.Amount-(-400.0)) > 0.001 {
				t.Errorf("expected budget amount -400.0, got %f", p.Amount)
			}
		}
	}

	if !foundBudget {
		t.Error("expected to find a variable budget prediction for Groceries")
	}
}

type mockExclusionRepo struct {
	exclusions []entity.ExcludedForecast
}

func (m *mockExclusionRepo) SaveExclusion(_ context.Context, e entity.ExcludedForecast) error {
	m.exclusions = append(m.exclusions, e)
	return nil
}

func (m *mockExclusionRepo) FindExclusions(_ context.Context, _ uuid.UUID) ([]entity.ExcludedForecast, error) {
	return m.exclusions, nil
}

func (m *mockExclusionRepo) DeleteExclusion(_ context.Context, _ uuid.UUID, forecastID uuid.UUID) error {
	for i, e := range m.exclusions {
		if e.ForecastID == forecastID {
			m.exclusions = append(m.exclusions[:i], m.exclusions[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockExclusionRepo) SavePatternExclusion(_ context.Context, _ entity.PatternExclusion) error {
	return nil
}

func (m *mockExclusionRepo) FindPatternExclusions(_ context.Context, _ uuid.UUID) ([]entity.PatternExclusion, error) {
	return nil, nil
}

func (m *mockExclusionRepo) DeletePatternExclusion(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func TestForecastingService_GetCashFlowForecast(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	txns := []entity.Transaction{
		// Rent: Recurring
		{
			Description: "Rent Payment",
			Amount:      -1200.0,
			BaseAmount:  -1200.0,
			BookingDate: now.AddDate(0, -3, 0),
		},
		{
			Description: "Rent Payment",
			Amount:      -1200.0,
			BaseAmount:  -1200.0,
			BookingDate: now.AddDate(0, -2, 0),
		},
		{
			Description: "Rent Payment",
			Amount:      -1200.0,
			BaseAmount:  -1200.0,
			BookingDate: now.AddDate(0, -1, 0),
		},
		// Salary: Recurring
		{
			Description: "Salary",
			Amount:      3000.0,
			BaseAmount:  3000.0,
			BookingDate: now.AddDate(0, -3, 0),
		},
		{
			Description: "Salary",
			Amount:      3000.0,
			BaseAmount:  3000.0,
			BookingDate: now.AddDate(0, -2, 0),
		},
		{
			Description: "Salary",
			Amount:      3000.0,
			BaseAmount:  3000.0,
			BookingDate: now.AddDate(0, -1, 0),
		},
	}

	repo := &mockForecastingRepo{txns: txns}
	bankRepo := &mockForecastingBankRepo{
		accounts: []entity.BankAccount{
			{Balance: 5000.0},
		},
	}
	catRepo := &mockCategoryRepo{saved: []entity.Category{}}

	svc := service.NewForecastingService(repo, bankRepo, catRepo, nil, nil, &mockExclusionRepo{}, nil, nil, nil)

	from := now
	to := now.AddDate(0, 2, 0) // Forecast 2 months ahead

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if forecast.CurrentBalance != 5000.0 {
		t.Errorf("expected current balance 5000.0, got %f", forecast.CurrentBalance)
	}

	// We expect 2 recurring patterns (Rent and Salary)
	// Each should appear 3 times (April, May, June) given the 2-month window from 'now'
	if len(forecast.Predictions) != 6 {
		t.Errorf("expected 6 predictions, got %d", len(forecast.Predictions))
	}

	// Verify balance trend
	lastPoint := forecast.TimeSeries[len(forecast.TimeSeries)-1]
	// Starting 5000.0 + 3*3000.0 (Income) - 3*1200.0 (Rent) = 10400.0
	expectedFinalBalance := 5000.0 + 3*3000.0 - 3*1200.0
	if math.Abs(lastPoint.ExpectedBalance-expectedFinalBalance) > 0.001 {
		t.Errorf("expected final balance %f, got %f", expectedFinalBalance, lastPoint.ExpectedBalance)
	}
}

func TestForecastingService_BonusForecasting(t *testing.T) {
	now := time.Now()
	userID := uuid.New()

	// Bonus in month 11 (November) of last two years
	// This should trigger a yearly pattern (2 occurrences + verified)
	payslips := []entity.Payslip{
		{
			PeriodMonthNum: 11,
			PeriodYear:     now.Year() - 2,
			GrossPay:       5000.0,
			BaseGrossPay:   5000.0,
			NetPay:         3000.0,
			BaseNetPay:     3000.0,
			PayoutAmount:   3000.0,
			BasePayoutAmount: 3000.0,
			Bonuses: []entity.Bonus{
				{Description: "Christmas Bonus", Amount: 2000.0, BaseAmount: 2000.0},
			},
		},
		{
			PeriodMonthNum: 11,
			PeriodYear:     now.Year() - 1,
			GrossPay:       5000.0,
			BaseGrossPay:   5000.0,
			NetPay:         3000.0,
			BaseNetPay:     3000.0,
			PayoutAmount:   3000.0,
			BasePayoutAmount: 3000.0,
			Bonuses: []entity.Bonus{
				{Description: "Christmas Bonus", Amount: 2000.0, BaseAmount: 2000.0},
			},
		},
	}

	// Matching bank transactions
	txns := []entity.Transaction{
		{
			ID:          uuid.New(),
			UserID:      userID,
			BookingDate: time.Date(now.Year()-2, 11, 28, 0, 0, 0, 0, time.UTC),
			Description: "Employer Salary Payout",
			Amount:      3000.0,
			BaseAmount:  3000.0,
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			BookingDate: time.Date(now.Year()-1, 11, 28, 0, 0, 0, 0, time.UTC),
			Description: "Employer Salary Payout",
			Amount:      3000.0,
			BaseAmount:  3000.0,
		},
	}

	repo := &mockForecastingRepo{txns: txns}
	bankRepo := &mockForecastingBankRepo{
		accounts: []entity.BankAccount{
			{Balance: 1000.0},
		},
	}
	catRepo := &mockCategoryRepo{saved: []entity.Category{}}
	payslipRepo := &mockPayslipRepo{payslips: payslips}

	svc := service.NewForecastingService(repo, bankRepo, catRepo, payslipRepo, nil, &mockExclusionRepo{}, nil, nil, nil)

	// Range covering November of current year
	from := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(now.Year(), 12, 31, 0, 0, 0, 0, time.UTC)

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundBonus := false
	for _, p := range forecast.Predictions {
		// New logic uses "Bonus: Description"
		if strings.Contains(p.Description, "Bonus: Christmas Bonus") {
			foundBonus = true
			expectedNet := 2000.0 * (3000.0 / 5000.0)
			if p.Amount != expectedNet {
				t.Errorf("expected net bonus amount %f, got %f", expectedNet, p.Amount)
			}
			if p.BookingDate.Month() != time.November {
				t.Errorf("expected bonus in November, got %v", p.BookingDate.Month())
			}
		}
	}

	if !foundBonus {
		t.Error("expected to find a bonus prediction for Christmas Bonus")
	}
}

// --- Mock PlannedTransactionRepository ---

type mockPlannedTransactionRepo struct {
	planned []entity.PlannedTransaction
}

func (m *mockPlannedTransactionRepo) Create(ctx context.Context, pt *entity.PlannedTransaction) error {
	return nil
}
func (m *mockPlannedTransactionRepo) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.PlannedTransaction, error) {
	return nil, nil
}
func (m *mockPlannedTransactionRepo) Update(ctx context.Context, pt *entity.PlannedTransaction) error {
	return nil
}
func (m *mockPlannedTransactionRepo) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return nil
}
func (m *mockPlannedTransactionRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error) {
	return m.planned, nil
}
func (m *mockPlannedTransactionRepo) FindPendingByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error) {
	var pending []entity.PlannedTransaction
	for _, p := range m.planned {
		if p.Status == entity.PlannedTransactionStatusPending {
			pending = append(pending, p)
		}
	}
	return pending, nil
}
func TestForecastingService_RecurringPlannedTransactions(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	repo := &mockForecastingRepo{}
	bankRepo := &mockForecastingBankRepo{}
	catRepo := &mockCategoryRepo{saved: []entity.Category{}}

	// 1. Recurring Monthly: Rent
	pt := entity.PlannedTransaction{
		ID:             uuid.New(),
		UserID:         userID,
		Amount:         -1200.0,
		BaseAmount:     -1200.0,
		Date:           time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		Description:    "Monthly Rent",
		Status:         entity.PlannedTransactionStatusPending,
		IntervalMonths: 1,
	}

	ptRepo := &mockPlannedTransactionRepo{planned: []entity.PlannedTransaction{pt}}
	svc := service.NewForecastingService(repo, bankRepo, catRepo, nil, ptRepo, &mockExclusionRepo{}, nil, nil, nil)

	// Forecast covering 3 months
	from := pt.Date
	to := pt.Date.AddDate(0, 3, 0) // Should include 4 instances (current + 3 future)

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, from, to)
	assert.NoError(t, err)

	count := 0
	for _, p := range forecast.Predictions {
		if strings.Contains(p.Description, "Planned: Monthly Rent") {
			count++
			assert.Equal(t, -1200.0, p.Amount)
		}
	}
	// Depending on the exact window, it should be 4 (Month 0, 1, 2, 3)
	assert.Equal(t, 4, count)
}

func TestForecastingService_SoftSuppression(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	catID := uuid.New()

	// 1. Auto-detected pattern for "Salary" (3000.0)
	txns := []entity.Transaction{
		{Description: "Salary", Amount: 3000.0, BaseAmount: 3000.0, BookingDate: now.AddDate(0, -3, 0), CategoryID: &catID},
		{Description: "Salary", Amount: 3000.0, BaseAmount: 3000.0, BookingDate: now.AddDate(0, -2, 0), CategoryID: &catID},
		{Description: "Salary", Amount: 3000.0, BaseAmount: 3000.0, BookingDate: now.AddDate(0, -1, 0), CategoryID: &catID},
	}

	// 2. Manual planned transaction for "Salary" (3000.0) in the same category
	// One-off manual (IntervalMonths = 0) must be within 7 days of auto-forecast
	pt := entity.PlannedTransaction{
		ID:             uuid.New(),
		UserID:         userID,
		Amount:         3000.0,
		BaseAmount:     3000.0,
		Date:           now.AddDate(0, 0, 1), // Tomorrow, very close to "now"
		Description:    "Manual Salary Forecast",
		Status:         entity.PlannedTransactionStatusPending,
		CategoryID:     &catID,
		IntervalMonths: 0,
	}

	repo := &mockForecastingRepo{txns: txns}
	bankRepo := &mockForecastingBankRepo{}
	catRepo := &mockCategoryRepo{saved: []entity.Category{
		{ID: catID, Name: "Income"},
	}}
	ptRepo := &mockPlannedTransactionRepo{planned: []entity.PlannedTransaction{pt}}

	svc := service.NewForecastingService(repo, bankRepo, catRepo, nil, ptRepo, &mockExclusionRepo{}, nil, nil, nil)

	from := now
	to := now.AddDate(0, 1, 0)

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, from, to)
	assert.NoError(t, err)

	foundManual := false
	foundAuto := false
	for _, p := range forecast.Predictions {
		if strings.Contains(p.Description, "Planned: Manual Salary Forecast") {
			foundManual = true
			assert.True(t, p.SkipForecasting, "Manual forecast should be suppressed")
			assert.Contains(t, p.Description, "(Superseded by auto-forecast)")
		}
		if p.Description == "Salary" {
			foundAuto = true
			assert.False(t, p.SkipForecasting, "Auto-forecast should NOT be suppressed")
		}
	}

	assert.True(t, foundManual, "Should find manual forecast")
	assert.True(t, foundAuto, "Should find auto-forecast")
}

func TestForecastingService_WithPlannedTransactions(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	repo := &mockForecastingRepo{}
	bankRepo := &mockForecastingBankRepo{}
	catRepo := &mockCategoryRepo{saved: []entity.Category{}}

	pt := entity.PlannedTransaction{
		ID:          uuid.New(),
		UserID:      userID,
		Amount:      -150.0,
		BaseAmount:  -150.0,
		Date:        now.AddDate(0, 0, 5), // 5 days from now
		Description: "Planned Bill",
		Status:      entity.PlannedTransactionStatusPending,
	}

	ptRepo := &mockPlannedTransactionRepo{planned: []entity.PlannedTransaction{pt}}

	svc := service.NewForecastingService(repo, bankRepo, catRepo, nil, ptRepo, &mockExclusionRepo{}, nil, nil, nil)

	from := now
	to := now.AddDate(0, 1, 0)

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundPlanned := false
	for _, p := range forecast.Predictions {
		if strings.Contains(p.Description, "Planned: Planned Bill") {
			foundPlanned = true
			if p.Amount != -150.0 {
				t.Errorf("expected planned amount -150.0, got %f", p.Amount)
			}
			if p.Probability != 1.0 {
				t.Errorf("expected probability 1.0, got %f", p.Probability)
			}
		}
	}

	if !foundPlanned {
		t.Error("expected to find a planned prediction for 'Planned Bill'")
	}
}

func TestForecastingService_ForecastStrategies(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	catID3m := uuid.New()
	catIDAll := uuid.New()
	catID3y := uuid.New()

	txns := []entity.Transaction{
		// For catID3m (Strategy: 3m)
		{
			Description: "Recent 1",
			Amount:      -100.0,
			BaseAmount:  -100.0,
			BookingDate: now.AddDate(0, -1, 0), // 1 month ago
			CategoryID:  &catID3m,
		},
		{
			Description: "Recent 2",
			Amount:      -100.0,
			BaseAmount:  -100.0,
			BookingDate: now.AddDate(0, -2, 0), // 2 months ago
			CategoryID:  &catID3m,
		},
		{
			Description: "Old 1",
			Amount:      -500.0,
			BaseAmount:  -500.0,
			BookingDate: now.AddDate(0, -4, 0), // 4 months ago -> Should be IGNORED for 3m
			CategoryID:  &catID3m,
		},

		// For catIDAll (Strategy: all)
		{
			Description: "Recent 1",
			Amount:      -100.0,
			BaseAmount:  -100.0,
			BookingDate: now.AddDate(0, -1, 0),
			CategoryID:  &catIDAll,
		},
		{
			Description: "Old 1",
			Amount:      -500.0,
			BaseAmount:  -500.0,
			BookingDate: now.AddDate(0, -4, 0), // Should be INCLUDED for 'all'
			CategoryID:  &catIDAll,
		},

		// For catID3y (Strategy: 3y)
		{
			Description: "Recent 1",
			Amount:      -100.0,
			BaseAmount:  -100.0,
			BookingDate: now.AddDate(0, -1, 0),
			CategoryID:  &catID3y,
		},
		{
			Description: "Ancient",
			Amount:      -1000.0,
			BaseAmount:  -1000.0,
			BookingDate: now.AddDate(-4, 0, 0), // 4 years ago -> Should be IGNORED for 3y
			CategoryID:  &catID3y,
		},
	}

	repo := &mockForecastingRepo{txns: txns}
	bankRepo := &mockForecastingBankRepo{
		accounts: []entity.BankAccount{
			{Balance: 1000.0},
		},
	}
	catRepo := &mockCategoryRepo{saved: []entity.Category{
		{
			ID:                 catID3m,
			UserID:             userID,
			Name:               "3m-Cat",
			IsVariableSpending: true,
			ForecastStrategy:   "3m",
		},
		{
			ID:                 catIDAll,
			UserID:             userID,
			Name:               "All-Cat",
			IsVariableSpending: true,
			ForecastStrategy:   "all",
		},
		{
			ID:                 catID3y,
			UserID:             userID,
			Name:               "3y-Cat",
			IsVariableSpending: true,
			ForecastStrategy:   "3y",
		},
	}}

	svc := service.NewForecastingService(repo, bankRepo, catRepo, nil, nil, &mockExclusionRepo{}, nil, nil, nil)

	// Forecast for the next month
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, 1, 0)
	to := from.AddDate(0, 0, 27) // ~ end of month

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found3m := false
	foundAll := false
	found3y := false

	for _, p := range forecast.Predictions {
		if strings.Contains(p.Description, "Variable Budget: 3m-Cat") {
			found3m = true
			// Expected: (-100 + -100) / 2 = -100
			if math.Abs(p.Amount-(-100.0)) > 0.001 {
				t.Errorf("3m-Cat: expected -100.0, got %f", p.Amount)
			}
		}
		if strings.Contains(p.Description, "Variable Budget: All-Cat") {
			foundAll = true
			// Expected: (-100 + -500) / 2 = -300
			if math.Abs(p.Amount-(-300.0)) > 0.001 {
				t.Errorf("All-Cat: expected -300.0, got %f", p.Amount)
			}
		}
		if strings.Contains(p.Description, "Variable Budget: 3y-Cat") {
			found3y = true
			// Expected: -100 / 1 = -100
			if math.Abs(p.Amount-(-100.0)) > 0.001 {
				t.Errorf("3y-Cat: expected -100.0, got %f", p.Amount)
			}
		}
	}

	if !found3m {
		t.Error("expected to find a variable budget prediction for 3m-Cat")
	}
	if !foundAll {
		t.Error("expected to find a variable budget prediction for All-Cat")
	}
	if !found3y {
		t.Error("expected to find a variable budget prediction for 3y-Cat")
	}
}

func TestForecastingService_CalculateCategoryAverage(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	catID := uuid.New()

	// 1. Setup transactions for testing averages
	txns := []entity.Transaction{
		// Transaction in current month
		{
			Amount:      -100.0,
			BaseAmount:  -100.0,
			BookingDate: now.AddDate(0, 0, -2),
			CategoryID:  &catID,
		},
		// Transaction in Month -1
		{
			Amount:      -200.0,
			BaseAmount:  -200.0,
			BookingDate: now.AddDate(0, -1, -2),
			CategoryID:  &catID,
		},
		// Transaction in Month -3 (safely outside 3m strategy window which is exactly 3 months ago from now)
		{
			Amount:      -400.0,
			BaseAmount:  -400.0,
			BookingDate: now.AddDate(0, -3, -5),
			CategoryID:  &catID,
		},
		// Transaction in Month -13 (outside 1y)
		{
			Amount:      -1200.0,
			BaseAmount:  -1200.0,
			BookingDate: now.AddDate(-1, -1, 0),
			CategoryID:  &catID,
		},
	}

	repo := &mockForecastingRepo{txns: txns}
	svc := service.NewForecastingService(repo, &mockForecastingBankRepo{}, &mockCategoryRepo{}, nil, nil, &mockExclusionRepo{}, nil, nil, nil)

	tests := []struct {
		name     string
		strategy string
		expected float64
	}{
		{
			name:     "Strategy 3m: Includes current month and Month -1",
			strategy: "3m",
			expected: (-100.0 + -200.0) / 2.0, // = -150.0
		},
		{
			name:     "Strategy 6m: Includes everything up to Month -3",
			strategy: "6m",
			expected: (-100.0 + -200.0 + -400.0) / 3.0, // = -233.333333
		},
		{
			name:     "Strategy 1y: Includes everything up to Month -3",
			strategy: "1y",
			expected: (-100.0 + -200.0 + -400.0) / 3.0, // = -233.333333
		},
		{
			name:     "Strategy all: Includes all historical transactions",
			strategy: "all",
			expected: (-100.0 + -200.0 + -400.0 + -1200.0) / 4.0, // = -475.0
		},
		{
			name:     "Strategy 1m: Includes only current month",
			strategy: "1m",
			expected: -100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.CalculateCategoryAverage(context.Background(), userID, catID, tt.strategy)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Abs(got-tt.expected) > 1e-4 {
				t.Errorf("got %f, expected %f", got, tt.expected)
			}
		})
	}

	t.Run("Empty history returns 0", func(t *testing.T) {
		emptyRepo := &mockForecastingRepo{txns: []entity.Transaction{}}
		svcEmpty := service.NewForecastingService(emptyRepo, &mockForecastingBankRepo{}, &mockCategoryRepo{}, nil, nil, &mockExclusionRepo{}, nil, nil, nil)
		got, err := svcEmpty.CalculateCategoryAverage(context.Background(), userID, catID, "3m")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 0 {
			t.Errorf("got %f, expected 0", got)
		}
	})
}

func TestForecastingService_CalculateCategoryAverage_Grouping(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	catID := uuid.New()

	txns := []entity.Transaction{
		{Amount: -100.0, BaseAmount: -100.0, BookingDate: now.AddDate(0, 0, -2), CategoryID: &catID},
		{Amount: -50.0, BaseAmount: -50.0, BookingDate: now.AddDate(0, 0, -3), CategoryID: &catID},
		{Amount: -200.0, BaseAmount: -200.0, BookingDate: now.AddDate(0, -1, -2), CategoryID: &catID},
	}

	repo := &mockForecastingRepo{txns: txns}
	svc := service.NewForecastingService(repo, &mockForecastingBankRepo{}, &mockCategoryRepo{}, nil, nil, &mockExclusionRepo{}, nil, nil, nil)

	got, err := svc.CalculateCategoryAverage(context.Background(), userID, catID, "3m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := (-150.0 + -200.0) / 2.0 // = -175.0
	if math.Abs(got-expected) > 1e-6 {
		t.Errorf("got %f, expected %f", got, expected)
	}
}
