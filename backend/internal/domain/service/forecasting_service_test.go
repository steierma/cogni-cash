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
func (m *mockForecastingRepo) UpdateTransactionCategoriesBulk(_ context.Context, _ []string, _ *uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingRepo) UpdateTransactionSubscription(_ context.Context, _ string, _ *uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingRepo) MarkTransactionReviewed(_ context.Context, _ string, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingRepo) MarkTransactionsReviewedBulk(_ context.Context, _ []string, _ uuid.UUID) error {
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
func (m *mockForecastingRepo) UpdateStatementAccount(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingRepo) GetTransactionsByAccountID(_ context.Context, _ uuid.UUID, _ uuid.UUID) ([]entity.Transaction, error) {
	return nil, nil
}

// --- Simple stub SubscriptionRepository for forecasting tests ---

type mockForecastSubRepo struct {
	subs []entity.Subscription
}

func (m *mockForecastSubRepo) GetByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) (entity.Subscription, error) {
	return entity.Subscription{}, nil
}
func (m *mockForecastSubRepo) FindByUserID(_ context.Context, _ uuid.UUID) ([]entity.Subscription, error) {
	return m.subs, nil
}
func (m *mockForecastSubRepo) Create(_ context.Context, sub entity.Subscription) (entity.Subscription, error) {
	return sub, nil
}
func (m *mockForecastSubRepo) CreateWithBackfill(_ context.Context, sub entity.Subscription, _ []string) (entity.Subscription, error) {
	return sub, nil
}
func (m *mockForecastSubRepo) Update(_ context.Context, sub entity.Subscription) (entity.Subscription, error) {
	return sub, nil
}
func (m *mockForecastSubRepo) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error { return nil }
func (m *mockForecastSubRepo) LogEvent(_ context.Context, _ entity.SubscriptionEvent) error {
	return nil
}
func (m *mockForecastSubRepo) GetEvents(_ context.Context, _ uuid.UUID, _ uuid.UUID) ([]entity.SubscriptionEvent, error) {
	return nil, nil
}
func (m *mockForecastSubRepo) SetDiscoveryFeedback(_ context.Context, _ uuid.UUID, _ string, _ entity.DiscoveryFeedbackStatus, _ string) error {
	return nil
}
func (m *mockForecastSubRepo) GetDiscoveryFeedback(_ context.Context, _ uuid.UUID) ([]entity.DiscoveryFeedback, error) {
	return nil, nil
}
func (m *mockForecastSubRepo) DeleteDiscoveryFeedback(_ context.Context, _ uuid.UUID, _ string) error {
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
func (m *mockForecastingBankRepo) SaveAccount(_ context.Context, _ *entity.BankAccount) error {
	return nil
}
func (m *mockForecastingBankRepo) FindAccountByIBAN(_ context.Context, _ string, _ uuid.UUID) (*entity.BankAccount, error) {
	return nil, nil
}
func (m *mockForecastingBankRepo) DeleteAccount(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}
func (m *mockForecastingBankRepo) UpdateExpiryNotifiedAt(_ context.Context, _ uuid.UUID, _ *time.Time) error {
	return nil
}
func (m *mockForecastingBankRepo) GetExpiringConnections(_ context.Context, _ int) ([]entity.BankConnection, error) {
	return nil, nil
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

	svc := service.NewForecastingService(repo, bankRepo, catRepo, nil, nil, &mockForecastSubRepo{}, nil, nil, nil)

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

func TestForecastingService_SubscriptionProjection(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	catID := uuid.New()

	nextOcc := time.Date(now.Year(), now.Month(), 15, 0, 0, 0, 0, time.UTC)
	if nextOcc.Before(now) {
		nextOcc = nextOcc.AddDate(0, 1, 0)
	}

	subID := uuid.New()
	subs := []entity.Subscription{
		{
			ID:              subID,
			UserID:          userID,
			MerchantName:    "Netflix",
			Amount:          -12.99,
			Currency:        "EUR",
			BillingCycle:    "monthly",
			BillingInterval: 1,
			CategoryID:      &catID,
			Status:          entity.SubscriptionStatusActive,
			NextOccurrence:  &nextOcc,
		},
	}

	repo := &mockForecastingRepo{}
	bankRepo := &mockForecastingBankRepo{accounts: []entity.BankAccount{{Balance: 5000.0}}}
	catRepo := &mockCategoryRepo{saved: []entity.Category{}}
	subRepo := &mockForecastSubRepo{subs: subs}

	svc := service.NewForecastingService(repo, bankRepo, catRepo, nil, nil, subRepo, nil, nil, nil)

	from := now
	to := now.AddDate(0, 3, 0)

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, from, to)
	assert.NoError(t, err)

	count := 0
	for _, p := range forecast.Predictions {
		if p.Description == "Netflix" {
			count++
			assert.Equal(t, -12.99, p.Amount)
			assert.NotNil(t, p.SubscriptionID)
			assert.Equal(t, subID, *p.SubscriptionID)
		}
	}
	assert.GreaterOrEqual(t, count, 2, "Expected at least 2 Netflix predictions over 3 months")
}

func TestForecastingService_CancelledSubscriptionExcluded(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	nextOcc := now.AddDate(0, 0, 15)

	subs := []entity.Subscription{
		{
			ID:              uuid.New(),
			UserID:          userID,
			MerchantName:    "Cancelled Service",
			Amount:          -9.99,
			Currency:        "EUR",
			BillingCycle:    "monthly",
			BillingInterval: 1,
			Status:          entity.SubscriptionStatusCancelled,
			NextOccurrence:  &nextOcc,
		},
	}

	subRepo := &mockForecastSubRepo{subs: subs}
	svc := service.NewForecastingService(&mockForecastingRepo{}, &mockForecastingBankRepo{}, &mockCategoryRepo{}, nil, nil, subRepo, nil, nil, nil)

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, now, now.AddDate(0, 2, 0))
	assert.NoError(t, err)
	assert.Empty(t, forecast.Predictions, "Cancelled subscriptions should not appear in forecast")
}

func TestForecastingService_CancellationPendingWithEndDate(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	nextOcc := now.AddDate(0, 0, 5)
	endDate := now.AddDate(0, 1, 15) // ends in ~1.5 months

	subs := []entity.Subscription{
		{
			ID:              uuid.New(),
			UserID:          userID,
			MerchantName:    "Expiring Service",
			Amount:          -19.99,
			Currency:        "EUR",
			BillingCycle:    "monthly",
			BillingInterval: 1,
			Status:          entity.SubscriptionStatusCancellationPending,
			NextOccurrence:  &nextOcc,
			ContractEndDate: &endDate,
		},
	}

	subRepo := &mockForecastSubRepo{subs: subs}
	svc := service.NewForecastingService(&mockForecastingRepo{}, &mockForecastingBankRepo{}, &mockCategoryRepo{}, nil, nil, subRepo, nil, nil, nil)

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, now, now.AddDate(0, 6, 0))
	assert.NoError(t, err)

	count := 0
	for _, p := range forecast.Predictions {
		if p.Description == "Expiring Service" {
			count++
			assert.True(t, !p.BookingDate.After(endDate), "Should not project past ContractEndDate")
		}
	}
	assert.GreaterOrEqual(t, count, 1, "Should have at least 1 prediction before end date")
	assert.LessOrEqual(t, count, 2, "Should have at most 2 predictions before end date")
}

func TestForecastingService_YearlySubscription(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	nextOcc := now.AddDate(0, 2, 0) // 2 months from now

	subs := []entity.Subscription{
		{
			ID:              uuid.New(),
			UserID:          userID,
			MerchantName:    "Annual Insurance",
			Amount:          -600.0,
			Currency:        "EUR",
			BillingCycle:    "yearly",
			BillingInterval: 1,
			Status:          entity.SubscriptionStatusActive,
			NextOccurrence:  &nextOcc,
		},
	}

	subRepo := &mockForecastSubRepo{subs: subs}
	svc := service.NewForecastingService(&mockForecastingRepo{}, &mockForecastingBankRepo{}, &mockCategoryRepo{}, nil, nil, subRepo, nil, nil, nil)

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, now, now.AddDate(0, 6, 0))
	assert.NoError(t, err)

	count := 0
	for _, p := range forecast.Predictions {
		if p.Description == "Annual Insurance" {
			count++
		}
	}
	assert.Equal(t, 1, count, "Yearly subscription should appear once in 6-month window")
}

func TestForecastingService_TrialSubscriptionProbability(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	nextOcc := now.AddDate(0, 0, 10)

	subs := []entity.Subscription{
		{
			ID:              uuid.New(),
			UserID:          userID,
			MerchantName:    "Trial App",
			Amount:          -4.99,
			Currency:        "EUR",
			BillingCycle:    "monthly",
			BillingInterval: 1,
			Status:          entity.SubscriptionStatusActive,
			IsTrial:         true,
			NextOccurrence:  &nextOcc,
		},
	}

	subRepo := &mockForecastSubRepo{subs: subs}
	svc := service.NewForecastingService(&mockForecastingRepo{}, &mockForecastingBankRepo{}, &mockCategoryRepo{}, nil, nil, subRepo, nil, nil, nil)

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, now, now.AddDate(0, 1, 0))
	assert.NoError(t, err)

	for _, p := range forecast.Predictions {
		if p.Description == "Trial App" {
			assert.Equal(t, 0.7, p.Probability, "Trial subscriptions should have 0.7 probability")
		}
	}
}

func TestForecastingService_NoSubscriptions(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	subRepo := &mockForecastSubRepo{}
	svc := service.NewForecastingService(&mockForecastingRepo{}, &mockForecastingBankRepo{accounts: []entity.BankAccount{{Balance: 1000.0}}}, &mockCategoryRepo{}, nil, nil, subRepo, nil, nil, nil)

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, now, now.AddDate(0, 1, 0))
	assert.NoError(t, err)
	assert.Empty(t, forecast.Predictions)
	assert.Equal(t, 1000.0, forecast.CurrentBalance)
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
	svc := service.NewForecastingService(repo, bankRepo, catRepo, nil, ptRepo, &mockForecastSubRepo{}, nil, nil, nil)

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

func TestForecastingService_PlannedTransactionsNoSupersession(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	catID := uuid.New()
	nextOcc := now.AddDate(0, 0, 1)

	// Active subscription for "Netflix"
	subs := []entity.Subscription{
		{
			ID:              uuid.New(),
			UserID:          userID,
			MerchantName:    "Netflix",
			Amount:          -12.99,
			Currency:        "EUR",
			BillingCycle:    "monthly",
			BillingInterval: 1,
			CategoryID:      &catID,
			Status:          entity.SubscriptionStatusActive,
			NextOccurrence:  &nextOcc,
		},
	}

	// Manual planned transaction for same category
	pt := entity.PlannedTransaction{
		ID:          uuid.New(),
		UserID:      userID,
		Amount:      -12.99,
		BaseAmount:  -12.99,
		Date:        now.AddDate(0, 0, 2),
		Description: "Manual Netflix Override",
		Status:      entity.PlannedTransactionStatusPending,
		CategoryID:  &catID,
	}

	svc := service.NewForecastingService(
		&mockForecastingRepo{}, &mockForecastingBankRepo{},
		&mockCategoryRepo{saved: []entity.Category{{ID: catID, Name: "Entertainment"}}},
		nil,
		&mockPlannedTransactionRepo{planned: []entity.PlannedTransaction{pt}},
		&mockForecastSubRepo{subs: subs},
		nil, nil, nil,
	)

	forecast, err := svc.GetCashFlowForecast(context.Background(), userID, now, now.AddDate(0, 1, 0))
	assert.NoError(t, err)

	// Both should appear — no supersession
	foundSub := false
	foundPlanned := false
	for _, p := range forecast.Predictions {
		if p.Description == "Netflix" {
			foundSub = true
		}
		if strings.Contains(p.Description, "Planned: Manual Netflix Override") {
			foundPlanned = true
			assert.NotContains(t, p.Description, "Superseded")
		}
	}

	assert.True(t, foundSub, "Should find subscription prediction")
	assert.True(t, foundPlanned, "Should find planned transaction prediction")
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
		Date:        now.AddDate(0, 0, 5),
		Description: "Planned Bill",
		Status:      entity.PlannedTransactionStatusPending,
	}

	ptRepo := &mockPlannedTransactionRepo{planned: []entity.PlannedTransaction{pt}}

	svc := service.NewForecastingService(repo, bankRepo, catRepo, nil, ptRepo, &mockForecastSubRepo{}, nil, nil, nil)

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

	svc := service.NewForecastingService(repo, bankRepo, catRepo, nil, nil, &mockForecastSubRepo{}, nil, nil, nil)

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
	svc := service.NewForecastingService(repo, &mockForecastingBankRepo{}, &mockCategoryRepo{}, nil, nil, &mockForecastSubRepo{}, nil, nil, nil)

	tests := []struct {
		name     string
		strategy string
		expected float64
	}{
		{
			name:     "Strategy 3m: Excludes current month, includes Month -1 and Month -3",
			strategy: "3m",
			expected: (-200.0 + -400.0) / 2.0, // = -300.0 (Feb is empty in this test data)
		},
		{
			name:     "Strategy 6m: Excludes current month, includes Month -1 and Month -3",
			strategy: "6m",
			expected: (-200.0 + -400.0) / 2.0, // = -300.0
		},
		{
			name:     "Strategy 1y: Excludes current month, includes Month -1, -3 and -13",
			strategy: "1y",
			expected: (-200.0 + -400.0 + -1200.0) / 3.0, // = -600.0
		},
		{
			name:     "Strategy all: Excludes current month, includes all historical transactions",
			strategy: "all",
			expected: (-200.0 + -400.0 + -1200.0) / 3.0, // = -600.0
		},
		{
			name:     "Strategy 1m: Excludes current month, includes only Month -1",
			strategy: "1m",
			expected: -200.0,
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
		svcEmpty := service.NewForecastingService(emptyRepo, &mockForecastingBankRepo{}, &mockCategoryRepo{}, nil, nil, &mockForecastSubRepo{}, nil, nil, nil)
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
	svc := service.NewForecastingService(repo, &mockForecastingBankRepo{}, &mockCategoryRepo{}, nil, nil, &mockForecastSubRepo{}, nil, nil, nil)

	got, err := svc.CalculateCategoryAverage(context.Background(), userID, catID, "3m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := -200.0 // Current month is excluded, only Month -1 (-200) remains
	if math.Abs(got-expected) > 1e-6 {
		t.Errorf("got %f, expected %f", got, expected)
	}
}
