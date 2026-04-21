package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strconv"
	"strings"
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
	forecastRepo port.ForecastingRepository
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
	forecastRepo port.ForecastingRepository,
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
		forecastRepo: forecastRepo,
		settingsRepo: settingsRepo,
		ratePort:     ratePort,
		Logger:       logger,
	}
}

type recurringPattern struct {
	template entity.Transaction
	interval int // in months
}

// normalizeHistory decomposes bundled bank transactions (e.g. Salary + Bonus) based on payslip data.
func (s *ForecastingService) normalizeHistory(ctx context.Context, userID uuid.UUID, history []entity.Transaction) []entity.Transaction {
	if s.payslipRepo == nil {
		return history
	}

	// 1. Fetch all payslips for the user
	payslips, err := s.payslipRepo.FindAll(ctx, entity.PayslipFilter{UserID: userID})
	if err != nil {
		s.Logger.Warn("Could not fetch payslips for history normalization", "error", err)
		return history
	}

	// 2. Map payslips by Period (Year-Month)
	payslipMap := make(map[string]entity.Payslip)
	for _, ps := range payslips {
		key := fmt.Sprintf("%d-%02d", ps.PeriodYear, ps.PeriodMonthNum)
		payslipMap[key] = ps
	}

	var normalized []entity.Transaction
	for _, tx := range history {
		// Only look at income transactions that could be salary
		if tx.BaseAmount <= 0 {
			normalized = append(normalized, tx)
			continue
		}

		monthKey := tx.BookingDate.Format("2006-01")
		ps, ok := payslipMap[monthKey]
		if !ok {
			normalized = append(normalized, tx)
			continue
		}

		// 3. Matching logic: Is this bank transaction the payout for this payslip?
		// We use BasePayoutAmount and BaseAmount because that is what actually hits the bank (after all deductions)
		// and allows for cross-currency matching.
		// We allow a small tolerance (e.g. ±2.00) for rounding or minor fee differences.
		if math.Abs(tx.BaseAmount-ps.BasePayoutAmount) > 2.0 {
			normalized = append(normalized, tx)
			continue
		}

		// 4. Decomposition: If matched and contains bonuses, split it
		if len(ps.Bonuses) == 0 {
			// Tag it as verified even if no bonus
			tx.IsPayslipVerified = true
			normalized = append(normalized, tx)
			continue
		}

		// Calculate net factor for this specific payslip to split gross bonuses accurately.
		// We use BaseNetPay / BaseGrossPay as the "tax ratio" to find the approximate net component of the bonus.
		netFactor := ps.BaseNetPay / ps.BaseGrossPay
		bonusNetTotal := 0.0

		for _, b := range ps.Bonuses {
			netBonus := b.BaseAmount * netFactor
			bonusNetTotal += netBonus

			// Create a virtual transaction for the bonus
			bonusTx := tx
			bonusTx.ID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("%s-vbonus-%s-%s", userID, b.Description, tx.ID)))
			bonusTx.Description = fmt.Sprintf("Bonus: %s", b.Description)
			bonusTx.Amount = b.Amount * netFactor // Keep original amount for completeness
			bonusTx.BaseAmount = netBonus
			bonusTx.IsPayslipVerified = true
			normalized = append(normalized, bonusTx)
		}

		// The remaining amount is the "Base Salary"
		baseSalaryTx := tx
		baseSalaryTx.BaseAmount = tx.BaseAmount - bonusNetTotal
		if tx.BaseAmount != 0 {
			baseSalaryTx.Amount = tx.Amount - (bonusNetTotal * (tx.Amount / tx.BaseAmount))
		} else {
			baseSalaryTx.Amount = tx.Amount
		}
		baseSalaryTx.IsPayslipVerified = true
		normalized = append(normalized, baseSalaryTx)
	}

	return normalized
}

func (s *ForecastingService) GetCashFlowForecast(ctx context.Context, userID uuid.UUID, fromDate, toDate time.Time) (entity.CashFlowForecast, error) {
	// 0. Fetch exclusions
	var exclusions []entity.ExcludedForecast
	var patternExclusions []entity.PatternExclusion
	if s.forecastRepo != nil {
		ex, err := s.forecastRepo.FindExclusions(ctx, userID)
		if err != nil {
			s.Logger.Warn("Could not fetch forecast exclusions", "error", err, "user_id", userID)
		} else {
			exclusions = ex
		}

		pe, err := s.forecastRepo.FindPatternExclusions(ctx, userID)
		if err != nil {
			s.Logger.Warn("Could not fetch pattern exclusions", "error", err, "user_id", userID)
		} else {
			patternExclusions = pe
		}
	}
	excludedMap := make(map[uuid.UUID]bool)
	for _, ex := range exclusions {
		excludedMap[ex.ForecastID] = true
	}

	patternExclusionMap := make(map[string]bool)
	for _, pe := range patternExclusions {
		patternExclusionMap[pe.MatchTerm] = true
	}

	// 1. Fetch historical transactions (last 3 years for pattern detection)
	histStart := time.Now().AddDate(-3, 0, 0)
	history, err := s.repo.FindTransactions(ctx, entity.TransactionFilter{
		UserID:   userID,
		FromDate: &histStart,
	})
	if err != nil {
		return entity.CashFlowForecast{}, err
	}

	// 2. Fetch categories to identify variable ones
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

	// 3. Detect recurring patterns (excluding variable categories and pattern exclusions)
	normalizedHistory := s.normalizeHistory(ctx, userID, history)
	recurring := s.detectRecurring(normalizedHistory, varCats, patternExclusionMap)

	// 4. Project discrete events into future
	predictions := make([]entity.PredictedTransaction, 0)
	predictions = append(predictions, s.projectFuture(userID, recurring, fromDate, toDate)...)

	// 5. Add Variable Spending Budgets
	budgetPredictions := s.calculateVariableBudgets(userID, history, varCats, fromDate, toDate)
	predictions = append(predictions, budgetPredictions...)

	// 5b. Determine base currency for multi-currency handling (moved up for planned transactions)
	baseCurrency := "EUR"
	if s.settingsRepo != nil {
		if val, _ := s.settingsRepo.Get(ctx, "BASE_DISPLAY_CURRENCY", userID); val != "" {
			baseCurrency = val
		}
	}

	// 5c. Add Planned Transactions
	if s.ptRepo != nil {
		plannedTransactions, err := s.ptRepo.FindByUserID(ctx, userID)
		if err != nil {
			s.Logger.Warn("Could not fetch planned transactions for forecasting", "error", err)
		} else {
			for _, pt := range plannedTransactions {
				// Only process Pending or Matched
				if pt.Status != entity.PlannedTransactionStatusPending {
					continue
				}

				// Calculate Base Amount dynamically for accurate superseding and forecasting
				baseAmt := pt.BaseAmount
				if baseAmt == 0 {
					if pt.Currency == baseCurrency {
						baseAmt = pt.Amount
					} else if s.ratePort != nil {
						rate, err := s.ratePort.GetRate(ctx, pt.Currency, baseCurrency, time.Now())
						if err == nil {
							baseAmt = pt.Amount * rate
						} else {
							baseAmt = pt.Amount // Fallback
						}
					} else {
						baseAmt = pt.Amount // Fallback
					}
				}

				// Soft Suppression: Check if this manual planned transaction overlaps with an auto-detected pattern
				isSuperseded := false
				for _, pred := range predictions {
					// We only compare against auto-detected predictions
					if !pred.IsPrediction || strings.HasPrefix(pred.Description, "Planned:") {
						continue
					}

					if pt.CategoryID != nil && pred.CategoryID != nil && *pt.CategoryID == *pred.CategoryID {
						// Amount similarity (±20%, min 50.0) based on base amounts
						amtDiff := math.Abs(baseAmt - pred.BaseAmount)
						maxAllowed := math.Max(math.Abs(pred.BaseAmount)*0.20, 50.0)

						if amtDiff <= maxAllowed {
							// Date proximity: Must be within the same month and roughly same day (±7 days)
							dayDiff := math.Abs(pt.Date.Sub(pred.BookingDate).Hours() / 24)
							if dayDiff <= 7 || (pt.Date.Month() == pred.BookingDate.Month() && pt.Date.Year() == pred.BookingDate.Year()) {
								isSuperseded = true
								break
							}
						}
					}
				}

				// Project occurrences
				current := pt.Date
				for {
					if !current.After(toDate) && !current.Before(fromDate) {
						idSeed := fmt.Sprintf("%s-pt-%s-%s", userID, pt.ID, current.Format("2006-01-02"))
						id := uuid.NewSHA1(uuid.NameSpaceOID, []byte(idSeed))

						desc := fmt.Sprintf("Planned: %s", pt.Description)
						if isSuperseded {
							desc += " (Superseded by auto-forecast)"
						}

						predictions = append(predictions, entity.PredictedTransaction{
							Transaction: entity.Transaction{
								ID:              id,
								Description:     desc,
								Amount:          pt.Amount,
								Currency:        pt.Currency,
								BaseAmount:      baseAmt,
								BaseCurrency:    baseCurrency,
								BookingDate:     current,
								ValutaDate:      current,
								CategoryID:      pt.CategoryID,
								IsPrediction:    true,
								Type:            templateType(pt.Amount),
								SkipForecasting: isSuperseded,
							},
							Probability: 1.0,
						})
					}

					if pt.IntervalMonths <= 0 {
						break
					}
					current = current.AddDate(0, pt.IntervalMonths, 0)
					if current.After(toDate) || (pt.EndDate != nil && current.After(*pt.EndDate)) {
						break
					}
				}
			}
		}
	}

	// Sort all predictions by date
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].BookingDate.Before(predictions[j].BookingDate)
	})

	// 5d. Mark excluded forecasts so they are not counted but still returned for the UI
	for i := range predictions {
		if excludedMap[predictions[i].ID] {
			predictions[i].SkipForecasting = true
		}
	}

	// 6. Get current balance (convert to base currency)
	accounts := []entity.BankAccount{}
	if s.bankRepo != nil {
		var err error
		accounts, err = s.bankRepo.GetAccountsByUserID(ctx, userID)
		if err != nil {
			s.Logger.Warn("Could not fetch bank accounts for balance forecast", "error", err)
		}
	}
	currentBalance := 0.0
	for _, acc := range accounts {
		if acc.Currency != baseCurrency && s.ratePort != nil {
			rate, err := s.ratePort.GetRate(ctx, acc.Currency, baseCurrency, time.Now())
			if err != nil {
				s.Logger.Error("Could not fetch rate for account balance conversion", "acc", acc.ID, "error", err)
				currentBalance += acc.Balance // Fallback: sum raw (incorrect but better than 0)
			} else {
				currentBalance += acc.Balance * rate
			}
		} else {
			currentBalance += acc.Balance
		}
	}

	// 7. Build time series (passing all predictions; buildTimeSeries will handle the SkipForecasting flag)
	timeSeries := s.buildTimeSeries(currentBalance, predictions, fromDate, toDate)

	return entity.CashFlowForecast{
		CurrentBalance: currentBalance,
		TimeSeries:     timeSeries,
		Predictions:    predictions,
	}, nil
}

func (s *ForecastingService) ExcludeForecast(ctx context.Context, userID uuid.UUID, forecastID uuid.UUID) error {
	if s.forecastRepo == nil {
		return fmt.Errorf("forecasting repo not configured")
	}
	return s.forecastRepo.SaveExclusion(ctx, entity.ExcludedForecast{
		UserID:     userID,
		ForecastID: forecastID,
		CreatedAt:  time.Now(),
	})
}

func (s *ForecastingService) IncludeForecast(ctx context.Context, userID uuid.UUID, forecastID uuid.UUID) error {
	if s.forecastRepo == nil {
		return fmt.Errorf("forecasting repo not configured")
	}
	return s.forecastRepo.DeleteExclusion(ctx, userID, forecastID)
}

func (s *ForecastingService) ExcludePattern(ctx context.Context, userID uuid.UUID, matchTerm string) error {
	if s.forecastRepo == nil {
		return fmt.Errorf("forecasting repo not configured")
	}
	return s.forecastRepo.SavePatternExclusion(ctx, entity.PatternExclusion{
		UserID:    userID,
		MatchTerm: matchTerm,
		CreatedAt: time.Now(),
	})
}

func (s *ForecastingService) IncludePattern(ctx context.Context, userID uuid.UUID, matchTerm string) error {
	if s.forecastRepo == nil {
		return fmt.Errorf("forecasting repo not configured")
	}
	return s.forecastRepo.DeletePatternExclusion(ctx, userID, matchTerm)
}

func (s *ForecastingService) ListPatternExclusions(ctx context.Context, userID uuid.UUID) ([]entity.PatternExclusion, error) {
	if s.forecastRepo == nil {
		return nil, fmt.Errorf("forecasting repo not configured")
	}
	return s.forecastRepo.FindPatternExclusions(ctx, userID)
}

func (s *ForecastingService) detectRecurring(history []entity.Transaction, varCats map[uuid.UUID]entity.Category, patternExclusions map[string]bool) []recurringPattern {
	// Group by normalized description
	groups := make(map[string][]entity.Transaction)
	for _, tx := range history {
		if tx.SkipForecasting {
			continue
		}

		desc := normalizeDescription(tx.Description)

		// SKIP pattern-level exclusions
		if patternExclusions[desc] {
			continue
		}

		// SKIP transactions in variable categories to avoid duplicates
		if tx.CategoryID != nil {
			if _, ok := varCats[*tx.CategoryID]; ok {
				continue
			}
		}

		groups[desc] = append(groups[desc], tx)
	}

	var recurring []recurringPattern
	for _, group := range groups {
		// Minimum 2 for verified, 3 for unverified
		if len(group) < 2 {
			continue
		}

		sort.Slice(group, func(i, j int) bool {
			return group[i].BookingDate.Before(group[j].BookingDate)
		})

		var sequence []entity.Transaction
		var detectedInterval int
		for i := 0; i < len(group); i++ {
			if len(sequence) == 0 {
				sequence = append(sequence, group[i])
				continue
			}

			last := sequence[len(sequence)-1]
			diff := group[i].BookingDate.Sub(last.BookingDate)
			days := diff.Hours() / 24

			interval := getIntervalFromDays(days, 3.0)

			// Match existing interval or any new valid interval
			if interval > 0 && (detectedInterval == 0 || interval == detectedInterval) {
				amtDiff := math.Abs(group[i].Amount - last.Amount)
				maxAllowed := math.Max(math.Abs(last.Amount)*0.20, 50.0)

				if amtDiff <= maxAllowed {
					sequence = append(sequence, group[i])
					detectedInterval = interval
					continue
				}
			}

			// No match: Break if we already have enough, otherwise reset
			minRequired := 3
			if isSequenceVerified(sequence) {
				minRequired = 2
			}

			if len(sequence) >= minRequired {
				break
			}
			sequence = []entity.Transaction{group[i]}
			detectedInterval = 0
		}

		minRequired := 3
		if isSequenceVerified(sequence) {
			minRequired = 2
		}

		if len(sequence) >= minRequired && detectedInterval > 0 {
			recurring = append(recurring, recurringPattern{
				template: sequence[len(sequence)-1],
				interval: detectedInterval,
			})
		}
	}

	return recurring
}

func isSequenceVerified(seq []entity.Transaction) bool {
	for _, tx := range seq {
		if tx.IsPayslipVerified {
			return true
		}
	}
	return false
}

func (s *ForecastingService) CalculateCategoryAverage(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID, strategy string) (float64, error) {
	// 1. Fetch historical transactions (last 3 years for pattern detection)
	histStart := time.Now().AddDate(-3, 0, 0)
	history, err := s.repo.FindTransactions(ctx, entity.TransactionFilter{
		UserID:     userID,
		CategoryID: &categoryID,
		FromDate:   &histStart,
	})
	if err != nil {
		return 0, err
	}

	startDate := getStartDateForStrategy(strategy, time.Now())

	// Group history by month
	monthlyTotals := make(map[string]float64)
	for _, tx := range history {
		if !startDate.IsZero() && tx.BookingDate.Before(startDate) {
			continue
		}
		monthKey := tx.BookingDate.Format("2006-01")
		monthlyTotals[monthKey] += tx.BaseAmount
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
				return now.AddDate(0, -val, 0)
			} else if unit == 'y' {
				return now.AddDate(-val, 0, 0)
			}
		}
	}

	// Default fallback (3y)
	return now.AddDate(-3, 0, 0)
}

func (s *ForecastingService) calculateVariableBudgets(userID uuid.UUID, history []entity.Transaction, varCats map[uuid.UUID]entity.Category, from, to time.Time) []entity.PredictedTransaction {
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
	for catID, months := range catMonthlyTotals {
		sum := 0.0
		for _, amt := range months {
			sum += amt
		}
		if len(months) > 0 {
			catAverages[catID] = sum / float64(len(months))
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

func templateType(amt float64) entity.TransactionType {
	if amt >= 0 {
		return entity.TransactionTypeCredit
	}
	return entity.TransactionTypeDebit
}

func (s *ForecastingService) projectFuture(userID uuid.UUID, recurring []recurringPattern, from, to time.Time) []entity.PredictedTransaction {
	var predictions []entity.PredictedTransaction

	for _, rt := range recurring {
		if rt.interval <= 0 {
			continue // Safety: Only project patterns with a valid interval
		}
		template := rt.template
		current := template.BookingDate
		for {
			current = current.AddDate(0, rt.interval, 0)
			if current.After(to) {
				break
			}

			if current.After(from) || current.Equal(from) {
				// Create a deterministic ID for each future instance
				// STABILITY: Use UserID (Salt), normalized description + target month + interval to survive new imports and prevent cross-user inference
				descHash := normalizeDescription(template.Description)
				idSeed := fmt.Sprintf("%s-recurring-%s-%s-%d", userID, descHash, current.Format("2006-01"), rt.interval)
				id := uuid.NewSHA1(uuid.NameSpaceOID, []byte(idSeed))

				pred := entity.PredictedTransaction{
					Transaction: template,
					Probability: 0.9,
				}
				pred.ID = id
				pred.BookingDate = current
				pred.ValutaDate = current
				pred.IsPrediction = true
				// Note: template already has BaseAmount and BaseCurrency populated from history
				predictions = append(predictions, pred)
			}
		}
	}

	return predictions
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
				// SKIP marked predictions for balance calculation
				if tx.SkipForecasting {
					continue
				}

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
