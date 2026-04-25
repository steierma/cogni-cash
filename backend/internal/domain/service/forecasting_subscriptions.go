package service

import (
	"context"
	"fmt"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

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
