package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

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
