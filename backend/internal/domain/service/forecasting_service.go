package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

var _ port.ForecastingUseCase = (*ForecastingService)(nil)

type ForecastingService struct {
	repo         port.BankStatementRepository
	bankRepo     port.BankRepository
	catRepo      port.CategoryRepository
	payslipRepo  port.PayslipRepository
	ptRepo       port.PlannedTransactionRepository
	subRepo      port.SubscriptionRepository
	settingsRepo port.SettingsRepository
	ratePort     port.CurrencyExchangeRatePort
	Logger       *slog.Logger
}

func NewForecastingService(
	repo port.BankStatementRepository,
	bankRepo port.BankRepository,
	catRepo port.CategoryRepository,
	payslipRepo port.PayslipRepository,
	ptRepo port.PlannedTransactionRepository,
	subRepo port.SubscriptionRepository,
	settingsRepo port.SettingsRepository,
	ratePort port.CurrencyExchangeRatePort,
	logger *slog.Logger,
) *ForecastingService {
	if logger == nil {
		logger = slog.Default()
	}
	return &ForecastingService{
		repo:         repo,
		bankRepo:     bankRepo,
		catRepo:      catRepo,
		payslipRepo:  payslipRepo,
		ptRepo:       ptRepo,
		subRepo:      subRepo,
		settingsRepo: settingsRepo,
		ratePort:     ratePort,
		Logger:       logger,
	}
}

func (s *ForecastingService) GetCashFlowForecast(ctx context.Context, userID uuid.UUID, fromDate, toDate time.Time) (entity.CashFlowForecast, error) {
	// 1. Fetch categories to identify variable ones
	allCats, err := s.catRepo.FindAll(ctx, userID)
	if err != nil {
		s.Logger.Warn("Could not fetch categories for forecasting", "error", err)
	}
	varCats := make(map[uuid.UUID]entity.Category)
	for _, c := range allCats {
		if c.IsVariableSpending {
			varCats[c.ID] = c
		}
	}

	// 2. Determine base currency
	baseCurrency := "EUR"
	if s.settingsRepo != nil {
		if val, _ := s.settingsRepo.Get(ctx, "BASE_DISPLAY_CURRENCY", userID); val != "" {
			baseCurrency = val
		}
	}

	// 3. Fetch bank accounts for balance and filtering
	accounts := []entity.BankAccount{}
	if s.bankRepo != nil {
		var err error
		accounts, err = s.bankRepo.GetAccountsByUserID(ctx, userID)
		if err != nil {
			s.Logger.Warn("Could not fetch bank accounts for forecasting", "error", err)
		}
	}

	// 4. Project subscriptions (replaces pattern detection)
	predictions := make([]entity.PredictedTransaction, 0)
	subMonthlyByCat := make(map[uuid.UUID]float64) // monthly base-currency equivalent per category
	if s.subRepo != nil {
		subPredictions, monthlyByCat := s.projectSubscriptions(ctx, userID, fromDate, toDate, baseCurrency, varCats, accounts)
		predictions = append(predictions, subPredictions...)
		subMonthlyByCat = monthlyByCat
	}

	// 4. Add Variable Spending Budgets for ALL variable categories.
	// For categories that also have active subscriptions the budget is reduced by the subscription
	// monthly equivalent so the total (sub + variable) always matches the historical burn rate.
	// The result is capped so it can never flip sign (i.e. never becomes income for an expense category).
	// Reconciled transactions must not skew variable budget calculations.
	histStart := time.Now().AddDate(-3, 0, 0)
	excludeReconciled := false
	history, err := s.repo.FindTransactions(ctx, entity.TransactionFilter{
		UserID:        userID,
		FromDate:      &histStart,
		IsReconciled:  &excludeReconciled,
		IncludeShared: true,
	})
	if err != nil {
		return entity.CashFlowForecast{}, err
	}

	budgetPredictions := s.calculateVariableBudgets(userID, history, varCats, subMonthlyByCat, fromDate, toDate)
	predictions = append(predictions, budgetPredictions...)

	// 5. Add Planned Transactions (no supersession logic — clean pass-through)
	if s.ptRepo != nil {
		plannedTransactions, err := s.ptRepo.FindByUserID(ctx, userID)
		if err != nil {
			s.Logger.Warn("Could not fetch planned transactions for forecasting", "error", err)
		} else {
			// If we have accounts, we should only show planned transactions linked to THESE accounts
			// OR those that are GLOBAL (bank_account_id is nil).
			// If no accounts are provided, show all.
			validAccIDs := make(map[uuid.UUID]bool)
			for _, acc := range accounts {
				validAccIDs[acc.ID] = true
			}

			for _, pt := range plannedTransactions {
				if pt.Status != entity.PlannedTransactionStatusPending {
					continue
				}

				// Tenancy/Account Filter:
				// Only include if:
				// 1. PT is global (no bank account)
				// 2. OR PT is linked to one of the accounts we are currently forecasting
				if pt.BankAccountID != nil && len(accounts) > 0 && !validAccIDs[*pt.BankAccountID] {
					continue
				}

				// Calculate Base Amount dynamically
				baseAmt := pt.BaseAmount
				if baseAmt == 0 {
					if pt.Currency == baseCurrency {
						baseAmt = pt.Amount
					} else if s.ratePort != nil {
						rate, err := s.ratePort.GetRate(ctx, pt.Currency, baseCurrency, time.Now())
						if err == nil {
							baseAmt = pt.Amount * rate
						} else {
							baseAmt = pt.Amount
						}
					} else {
						baseAmt = pt.Amount
					}
				}

				// Project occurrences
				current := pt.Date
				for {
					projectionDate := current
					if pt.SchedulingStrategy == entity.SchedulingStrategyLastBankDay {
						projectionDate = GetLastBusinessDay(current.Year(), current.Month(), current.Location())
					}

					if !projectionDate.After(toDate) && (projectionDate.After(fromDate) || projectionDate.Equal(fromDate)) {
						idSeed := fmt.Sprintf("%s-pt-%s-%s", userID, pt.ID, projectionDate.Format("2006-01-02"))
						id := uuid.NewSHA1(uuid.NameSpaceOID, []byte(idSeed))

						desc := fmt.Sprintf("Planned: %s", pt.Description)

						predictions = append(predictions, entity.PredictedTransaction{
							Transaction: entity.Transaction{
								ID:           id,
								Description:  desc,
								Amount:       pt.Amount,
								Currency:     pt.Currency,
								BaseAmount:   baseAmt,
								BaseCurrency: baseCurrency,
								BookingDate:  projectionDate,
								ValutaDate:   projectionDate,
								CategoryID:   pt.CategoryID,
								BankAccountID: pt.BankAccountID,
								IsPrediction: true,
								Type:         templateType(pt.Amount),
							},
							Probability: 1.0,
						})
					}

					if pt.IntervalMonths <= 0 {
						break
					}
					// Always increment based on the calendar month to avoid day-drifting
					// We use a safe way to add months: 
					// 1. Move to the first of the month
					// 2. Add the interval
					// 3. Try to restore the original day, or use the last day of that month
					year, month, _ := current.Date()
					originalDay := pt.Date.Day()
					
					// Move to next target month
					nextMonth := time.Date(year, month, 1, 0, 0, 0, 0, current.Location()).AddDate(0, pt.IntervalMonths, 0)
					
					// Try to restore original day
					lastDayOfNext := time.Date(nextMonth.Year(), nextMonth.Month()+1, 0, 0, 0, 0, 0, current.Location()).Day()
					targetDay := originalDay
					if targetDay > lastDayOfNext {
						targetDay = lastDayOfNext
					}
					
					current = time.Date(nextMonth.Year(), nextMonth.Month(), targetDay, 0, 0, 0, 0, current.Location())

					if current.After(toDate) || (pt.EndDate != nil && current.After(*pt.EndDate)) {
						break
					}
				}
			}
		}
	}

	// Sort all predictions by date, then by ID as a stable tiebreaker so that
	// rows with identical dates always appear in the same order across refreshes.
	// (sort.Slice is not stable and Go map iteration is random, both cause jumping.)
	sort.SliceStable(predictions, func(i, j int) bool {
		di := predictions[i].BookingDate
		dj := predictions[j].BookingDate
		if !di.Equal(dj) {
			return di.Before(dj)
		}
		return predictions[i].ID.String() < predictions[j].ID.String()
	})

	// 6. Get current balance (convert to base currency)
	currentBalance := 0.0
	for _, acc := range accounts {
		if acc.Currency != baseCurrency && s.ratePort != nil {
			rate, err := s.ratePort.GetRate(ctx, acc.Currency, baseCurrency, time.Now())
			if err != nil {
				s.Logger.Error("Could not fetch rate for account balance conversion", "acc", acc.ID, "error", err)
				currentBalance += acc.Balance
			} else {
				currentBalance += acc.Balance * rate
			}
		} else {
			currentBalance += acc.Balance
		}
	}

	// 7. Build time series
	timeSeries := s.buildTimeSeries(currentBalance, predictions, fromDate, toDate)

	return entity.CashFlowForecast{
		CurrentBalance: currentBalance,
		TimeSeries:     timeSeries,
		Predictions:    predictions,
	}, nil
}
