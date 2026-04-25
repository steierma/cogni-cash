package service

import (
	"math"
	"time"

	"cogni-cash/internal/domain/entity"
)

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
