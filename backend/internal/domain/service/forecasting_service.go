package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strconv"
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

// billingCycleToDays returns the number of days per interval for weekly subscriptions.
// For monthly/yearly, returns 0 (use AddDate with months instead).
func billingCycleToDays(cycle string, interval int) int {
	if cycle == "weekly" {
		return 7 * interval
	}
	return 0
}

// billingCycleToMonths returns the number of months per interval for monthly/yearly subscriptions.
func billingCycleToMonths(cycle string, interval int) int {
	switch cycle {
	case "monthly":
		return interval
	case "yearly":
		return interval * 12
	default:
		return interval
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

// projectSubscriptions projects future occurrences for active/cancellation_pending subscriptions.
// Returns predictions and a map of category ID → monthly base-currency equivalent of subscriptions
// (used by calculateVariableBudgets to reduce the variable budget for covered categories).
func (s *ForecastingService) projectSubscriptions(ctx context.Context, userID uuid.UUID, from, to time.Time, baseCurrency string, _ map[uuid.UUID]entity.Category, accounts []entity.BankAccount) ([]entity.PredictedTransaction, map[uuid.UUID]float64) {
	var predictions []entity.PredictedTransaction
	subMonthlyByCat := make(map[uuid.UUID]float64)

	subs, err := s.subRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.Logger.Warn("Could not fetch subscriptions for forecasting", "error", err, "user_id", userID)
		return predictions, subMonthlyByCat
	}

	// For filtering
	validAccIDs := make(map[uuid.UUID]bool)
	for _, acc := range accounts {
		validAccIDs[acc.ID] = true
	}

	now := time.Now()

	for _, sub := range subs {
		// Only project active or cancellation_pending
		if sub.Status != entity.SubscriptionStatusActive && sub.Status != entity.SubscriptionStatusCancellationPending {
			continue
		}

		// Tenancy/Account Filter:
		// Only include if:
		// 1. Subscription is global (no bank account)
		// 2. OR Subscription is linked to one of the accounts we are currently forecasting
		if sub.BankAccountID != nil && len(accounts) > 0 && !validAccIDs[*sub.BankAccountID] {
			continue
		}

		// Skip zero-amount subscriptions
		if sub.Amount == 0 {
			s.Logger.Warn("Skipping zero-amount subscription", "sub_id", sub.ID, "merchant", sub.MerchantName)
			continue
		}

		// Determine the starting point for projection
		startDate := s.resolveSubscriptionStart(sub, now)
		if startDate.IsZero() {
			s.Logger.Warn("Skipping subscription with no start date", "sub_id", sub.ID, "merchant", sub.MerchantName)
			continue
		}

		// Staleness check: if start date is more than 2 intervals in the past, skip
		days := billingCycleToDays(sub.BillingCycle, sub.BillingInterval)
		months := billingCycleToMonths(sub.BillingCycle, sub.BillingInterval)
		var stalenessLimit time.Time
		if days > 0 {
			stalenessLimit = startDate.AddDate(0, 0, days*2)
		} else {
			stalenessLimit = startDate.AddDate(0, months*2, 0)
		}
		if now.After(stalenessLimit) {
			s.Logger.Warn("Skipping stale subscription", "sub_id", sub.ID, "merchant", sub.MerchantName,
				"next_occurrence", sub.NextOccurrence, "last_occurrence", sub.LastOccurrence)
			continue
		}

		// ContractEndDate guard for cancellation_pending
		endDate := to
		if sub.Status == entity.SubscriptionStatusCancellationPending && sub.ContractEndDate != nil {
			if sub.ContractEndDate.Before(now) {
				continue // Contract already ended
			}
			if sub.ContractEndDate.Before(endDate) {
				endDate = *sub.ContractEndDate
			}
		}

		// Convert amount to base currency
		baseAmt := sub.Amount
		if sub.Currency != baseCurrency && s.ratePort != nil {
			rate, err := s.ratePort.GetRate(ctx, sub.Currency, baseCurrency, now)
			if err == nil {
				baseAmt = sub.Amount * rate
			}
		}

		// Accumulate monthly base-currency equivalent for variable budget adjustment
		if sub.CategoryID != nil {
			subMonthlyByCat[*sub.CategoryID] += subBaseMonthly(baseAmt, sub.BillingCycle, sub.BillingInterval)
		}

		// Determine probability
		probability := 0.95
		if sub.IsTrial {
			probability = 0.7
		}

		// Project occurrences
		current := startDate
		for {
			if current.After(endDate) {
				break
			}

			if !current.After(to) && (current.After(from) || current.Equal(from)) {
				idSeed := fmt.Sprintf("%s-sub-%s-%s", userID, sub.ID, current.Format("2006-01"))
				id := uuid.NewSHA1(uuid.NameSpaceOID, []byte(idSeed))

				predictions = append(predictions, entity.PredictedTransaction{
					Transaction: entity.Transaction{
						ID:              id,
						Description:     sub.MerchantName,
						Amount:          sub.Amount,
						Currency:        sub.Currency,
						BaseAmount:      baseAmt,
						BaseCurrency:    baseCurrency,
						BookingDate:     current,
						ValutaDate:      current,
						CategoryID:      sub.CategoryID,
						BankAccountID:   sub.BankAccountID,
						IsPrediction:    true,
						Type:            templateType(sub.Amount),
						SubscriptionID:  &sub.ID,
					},
					Probability: probability,
				})
			}

			// Advance to next occurrence
			if days > 0 {
				current = current.AddDate(0, 0, days)
			} else {
				current = current.AddDate(0, months, 0)
			}
		}
	}

	return predictions, subMonthlyByCat
}

// subBaseMonthly converts a subscription's base-currency amount to a monthly equivalent.
func subBaseMonthly(baseAmt float64, cycle string, interval int) float64 {
	months := billingCycleToMonths(cycle, interval)
	if months > 0 {
		return baseAmt / float64(months)
	}
	days := billingCycleToDays(cycle, interval)
	if days > 0 {
		return baseAmt * 30.0 / float64(days)
	}
	return baseAmt
}

// resolveSubscriptionStart determines the first projection date for a subscription.
func (s *ForecastingService) resolveSubscriptionStart(sub entity.Subscription, now time.Time) time.Time {
	if sub.NextOccurrence != nil {
		return *sub.NextOccurrence
	}
	// Fallback: compute from LastOccurrence + 1 interval
	if sub.LastOccurrence != nil {
		days := billingCycleToDays(sub.BillingCycle, sub.BillingInterval)
		if days > 0 {
			return sub.LastOccurrence.AddDate(0, 0, days)
		}
		months := billingCycleToMonths(sub.BillingCycle, sub.BillingInterval)
		return sub.LastOccurrence.AddDate(0, months, 0)
	}
	return time.Time{} // Zero — caller must skip
}

func (s *ForecastingService) CalculateCategoryAverage(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID, strategy string) (float64, error) {
	// 1. Fetch historical transactions (last 3 years for pattern detection)
	// Reconciled transactions (e.g. Kartenabrechnung credit-card settlements) are internal
	// cashflow entries and must not skew the burn-rate / spending average.
	histStart := time.Now().AddDate(-3, 0, 0)
	excludeReconciled := false
	history, err := s.repo.FindTransactions(ctx, entity.TransactionFilter{
		UserID:        userID,
		CategoryID:    &categoryID,
		FromDate:      &histStart,
		IsReconciled:  &excludeReconciled,
		IncludeShared: true,
	})
	if err != nil {
		return 0, err
	}

	now := time.Now()
	startDate := getStartDateForStrategy(strategy, now)
	currentMonthKey := now.Format("2006-01")

	// Group history by month
	monthlyTotals := make(map[string]float64)
	for _, tx := range history {
		if !startDate.IsZero() && tx.BookingDate.Before(startDate) {
			continue
		}
		monthKey := tx.BookingDate.Format("2006-01")
		monthlyTotals[monthKey] += tx.BaseAmount
	}

	// Exclude current month if we have other historical months to avoid skewing the average
	if len(monthlyTotals) > 1 {
		delete(monthlyTotals, currentMonthKey)
	}

	if len(monthlyTotals) == 0 {
		return 0, nil
	}

	sum := 0.0
	for _, amt := range monthlyTotals {
		sum += amt
	}
	return sum / float64(len(monthlyTotals)), nil
}

func getStartDateForStrategy(strategy string, now time.Time) time.Time {
	if strategy == "all" {
		return time.Time{} // No limit
	}

	// Try parsing Nm or Ny (e.g., 6m, 2y)
	if len(strategy) > 1 {
		unit := strategy[len(strategy)-1]
		valStr := strategy[:len(strategy)-1]
		val, err := strconv.Atoi(valStr)
		if err == nil && val > 0 {
			if unit == 'm' {
				// Align to the start of the month val months ago
				return time.Date(now.Year(), now.Month()-time.Month(val), 1, 0, 0, 0, 0, now.Location())
			} else if unit == 'y' {
				// Align to the start of the year val years ago
				return time.Date(now.Year()-val, 1, 1, 0, 0, 0, 0, now.Location())
			}
		}
	}

	// Default fallback (3y)
	return time.Date(now.Year()-3, 1, 1, 0, 0, 0, 0, now.Location())
}

func (s *ForecastingService) calculateVariableBudgets(userID uuid.UUID, history []entity.Transaction, varCats map[uuid.UUID]entity.Category, subMonthlyByCat map[uuid.UUID]float64, from, to time.Time) []entity.PredictedTransaction {
	var predictions []entity.PredictedTransaction
	if len(varCats) == 0 {
		return predictions
	}

	// Calculate start dates for each category based on its strategy
	now := time.Now()
	catStartDates := make(map[uuid.UUID]time.Time)
	for id, cat := range varCats {
		catStartDates[id] = getStartDateForStrategy(cat.ForecastStrategy, now)
	}

	// 1. Calculate historical monthly averages for each variable category
	// Group history by category and month
	catMonthlyTotals := make(map[uuid.UUID]map[string]float64)
	for _, tx := range history {
		if tx.CategoryID == nil {
			continue
		}
		if _, ok := varCats[*tx.CategoryID]; !ok {
			continue
		}

		catID := *tx.CategoryID
		startDate := catStartDates[catID]
		if !startDate.IsZero() && tx.BookingDate.Before(startDate) {
			continue
		}

		monthKey := tx.BookingDate.Format("2006-01")
		if _, ok := catMonthlyTotals[catID]; !ok {
			catMonthlyTotals[catID] = make(map[string]float64)
		}
		catMonthlyTotals[catID][monthKey] += tx.BaseAmount
	}

	catAverages := make(map[uuid.UUID]float64)
	currentMonthKey := now.Format("2006-01")
	for catID, months := range catMonthlyTotals {
		// Exclude current month if we have other historical months
		if len(months) > 1 {
			delete(months, currentMonthKey)
		}

		sum := 0.0
		for _, amt := range months {
			sum += amt
		}
		if len(months) > 0 {
			catAverages[catID] = sum / float64(len(months))
		}
	}

	// Subtract the monthly subscription equivalent from the historical average so that the
	// variable budget only covers spend *beyond* what subscriptions already project.
	// The result is capped: it can never flip sign relative to the raw average
	// (an expense category can never become a net income source here).
	for catID, avg := range catAverages {
		if subMonthly, ok := subMonthlyByCat[catID]; ok && subMonthly != 0 {
			adjusted := avg - subMonthly
			// Cap: prevent sign flip
			if avg < 0 && adjusted > 0 {
				adjusted = 0
			} else if avg > 0 && adjusted < 0 {
				adjusted = 0
			}
			catAverages[catID] = adjusted
		}
	}

	// 2. Project for each month in the range [from, to]
	curr := from
	for !curr.After(to) {
		startOfMonth := time.Date(curr.Year(), curr.Month(), 1, 0, 0, 0, 0, curr.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, -1)

		// For each variable category
		for catID, cat := range varCats {
			avg := catAverages[catID]
			if avg == 0 {
				continue
			}

			remainingBudget := avg

			// If we are in the current month, subtract what we already spent
			if curr.Year() == now.Year() && curr.Month() == now.Month() {
				monthSpent := 0.0
				monthKey := now.Format("2006-01")
				if months, ok := catMonthlyTotals[catID]; ok {
					monthSpent = months[monthKey]
				}
				remainingBudget = avg - monthSpent

				// If we already overspent, don't forecast more for this month
				if (avg < 0 && remainingBudget > 0) || (avg > 0 && remainingBudget < 0) {
					remainingBudget = 0
				}
			}

			if remainingBudget == 0 {
				continue
			}

			// Place at the end of the month
			targetDate := endOfMonth
			if targetDate.After(to) {
				targetDate = to
			}
			if targetDate.Before(from) {
				targetDate = from
			}

			idSeed := fmt.Sprintf("%s-var-budget-%s-%s", userID, cat.ID, targetDate.Format("2006-01"))
			id := uuid.NewSHA1(uuid.NameSpaceOID, []byte(idSeed))

			predictions = append(predictions, entity.PredictedTransaction{
				Transaction: entity.Transaction{
					ID:           id,
					Description:  fmt.Sprintf("Variable Budget: %s", cat.Name),
					Amount:       remainingBudget,
					BaseAmount:   remainingBudget, // Multi-currency fix added here
					BookingDate:  targetDate,
					ValutaDate:   targetDate,
					CategoryID:   &catID,
					IsPrediction: true,
					Type:         templateType(remainingBudget),
				},
				Probability: 0.8,
			})
		}

		curr = startOfMonth.AddDate(0, 1, 0)
	}

	return predictions
}

// GetLastBusinessDay returns the last weekday (Mon-Fri) of the given month and year.
// It does not account for public holidays as they are region-specific.
func GetLastBusinessDay(year int, month time.Month, loc *time.Location) time.Time {
	// Start at the last day of the month
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, loc)

	// Rewind if it's a weekend
	for lastDay.Weekday() == time.Saturday || lastDay.Weekday() == time.Sunday {
		lastDay = lastDay.AddDate(0, 0, -1)
	}

	return lastDay
}

func templateType(amt float64) entity.TransactionType {
	if amt >= 0 {
		return entity.TransactionTypeCredit
	}
	return entity.TransactionTypeDebit
}

func (s *ForecastingService) buildTimeSeries(startBalance float64, predictions []entity.PredictedTransaction, from, to time.Time) []entity.ForecastPoint {
	var points []entity.ForecastPoint
	balance := startBalance

	daily := make(map[string][]entity.PredictedTransaction)
	for _, p := range predictions {
		day := p.BookingDate.Format("2006-01-02")
		daily[day] = append(daily[day], p)
	}

	curr := from
	for !curr.After(to) {
		dayStr := curr.Format("2006-01-02")
		dayIncome := 0.0
		dayExpense := 0.0
		catAmounts := make(map[string]float64)

		if txns, ok := daily[dayStr]; ok {
			for _, tx := range txns {

				if tx.BaseAmount > 0 {
					dayIncome += tx.BaseAmount
				} else {
					dayExpense += math.Abs(tx.BaseAmount)
				}
				balance += tx.BaseAmount
				if tx.CategoryID != nil {
					catAmounts[tx.CategoryID.String()] += tx.BaseAmount
				}
			}
		}

		points = append(points, entity.ForecastPoint{
			Date:            curr,
			ExpectedBalance: balance,
			Income:          dayIncome,
			Expense:         dayExpense,
			CategoryAmounts: catAmounts,
		})

		curr = curr.AddDate(0, 0, 1)
	}

	return points
}
