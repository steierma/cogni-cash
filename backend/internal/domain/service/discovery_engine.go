package service

import (
	"context"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

func (s *DiscoveryService) GetSuggestedSubscriptions(ctx context.Context, userID uuid.UUID) ([]entity.SuggestedSubscription, error) {
	// 0. Fetch preferences
	lookbackYears := 3
	val, err := s.settingsRepo.Get(ctx, "subscription_lookback_years", userID)
	if err == nil && val != "" {
		if v, err := strconv.Atoi(val); err == nil && v > 0 {
			lookbackYears = v
		}
	}

	amountTolerance := 0.10 // 10% default
	valTolerance, err := s.settingsRepo.Get(ctx, "subscription_discovery_amount_tolerance", userID)
	if err == nil && valTolerance != "" {
		if v, err := strconv.ParseFloat(valTolerance, 64); err == nil && v > 0 {
			amountTolerance = v
		}
	}

	minTxGeneric := 3 // Default minimum for generic merchants
	valMinTx, err := s.settingsRepo.Get(ctx, "subscription_discovery_min_transactions_generic", userID)
	if err == nil && valMinTx != "" {
		if v, err := strconv.Atoi(valMinTx); err == nil && v > 0 {
			minTxGeneric = v
		}
	}

	dateTolerance := 3.0 // Default 3 days
	valDate, err := s.settingsRepo.Get(ctx, "subscription_discovery_date_tolerance", userID)
	if err == nil && valDate != "" {
		if v, err := strconv.ParseFloat(valDate, 64); err == nil && v > 0 {
			dateTolerance = v
		}
	}

	// 1. Fetch historical transactions
	histStart := time.Now().AddDate(-lookbackYears, 0, 0)
	history, err := s.txRepo.FindTransactions(ctx, entity.TransactionFilter{
		UserID:        userID,
		FromDate:      &histStart,
		IncludeShared: true,
	})
	if err != nil {
		return nil, err
	}

	// 2. Fetch existing subscriptions and discovery feedback to filter suggestions
	existingSubs, err := s.subRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.Logger.Warn("Could not fetch existing subscriptions", "error", err)
	}

	feedback, err := s.subRepo.GetDiscoveryFeedback(ctx, userID)
	if err != nil {
		s.Logger.Warn("Could not fetch discovery feedback", "error", err)
	}

	feedbackMap := make(map[string]entity.DiscoveryFeedback)
	for _, f := range feedback {
		feedbackMap[normalizeDescription(f.MerchantName)] = f
	}

	// Build a set of all hashes, mandates, and IBANs already covered by existing subscriptions
	coveredHashes := make(map[string]bool)
	coveredMandates := make(map[string]bool)
	coveredIbans := make(map[string]bool)
	for _, sub := range existingSubs {
		for _, h := range sub.MatchingHashes {
			coveredHashes[h] = true
		}
		for _, m := range sub.LinkedMandates {
			if m != "" {
				coveredMandates[m] = true
			}
		}
		for _, i := range sub.LinkedIbans {
			if i != "" {
				coveredIbans[i] = true
			}
		}
	}

	// 3. Detect recurring patterns
	recurring := s.detectRecurring(history, amountTolerance, minTxGeneric, dateTolerance)

	var suggestions []entity.SuggestedSubscription
	for _, rt := range recurring {
		normDesc := normalizeDescription(rt.template.Description)
		normCounterparty := ""
		if rt.template.CounterpartyName != "" {
			normCounterparty = normalizeDescription(rt.template.CounterpartyName)
		}

		// Filter out if hashes are already mostly covered by an existing subscription
		matchedHashes := 0
		for _, h := range rt.matchingHashes {
			if coveredHashes[h] {
				matchedHashes++
			}
		}
		// If more than 50% of the sequence is already covered, skip it.
		// This handles the case where discovery finds the same pattern as an existing sub.
		if len(rt.matchingHashes) > 0 && float64(matchedHashes)/float64(len(rt.matchingHashes)) >= 0.5 {
			continue
		}

		// Priority Filter: If the pattern uses a Mandate or IBAN that is ALREADY covered, skip it.
		// This is the most reliable check for "I already have this subscription" regardless of renaming.
		if rt.template.MandateReference != "" && coveredMandates[rt.template.MandateReference] {
			continue
		}
		if rt.template.CounterpartyIban != "" && coveredIbans[rt.template.CounterpartyIban] {
			continue
		}

		// Filter out if already a subscription with similar amount and description
		isExisting := false
		for _, sub := range existingSubs {
			normSub := normalizeDescription(sub.MerchantName)
			// Match if: same merchant name (either Desc or CP) AND similar amount
			isSameMerchant := (normSub == normDesc || (normCounterparty != "" && normSub == normCounterparty))
			if isSameMerchant && isAmountClose(rt.template.Amount, sub.Amount, amountTolerance) {
				// We ALSO require the description to match to allow multiple distinct subscriptions
				// for the same merchant (e.g. different Barmenia contracts).
				if normalizeDescription(rt.template.Description) == normalizeDescription(sub.MerchantName) {
					isExisting = true
					break
				}
			}
		}
		if isExisting {
			continue
		}

		// Filter based on feedback (Hybrid Fuzzy Matching)
		isRejected := false
		for fbName, fb := range feedbackMap {
			if fb.Status != entity.DiscoveryStatusDeclined && fb.Status != entity.DiscoveryStatusAIRejected {
				continue
			}

			// Hybrid Threshold: 90% for <= 8 chars, 80% for > 8 chars
			getThreshold := func(s1, s2 string) float64 {
				maxLen := len(s1)
				if len(s2) > maxLen {
					maxLen = len(s2)
				}
				if maxLen <= 8 {
					return 0.90
				}
				return 0.80
			}

			if calculateSimilarity(fbName, normDesc) >= getThreshold(fbName, normDesc) {
				isRejected = true
				break
			}
			if normCounterparty != "" && calculateSimilarity(fbName, normCounterparty) >= getThreshold(fbName, normCounterparty) {
				isRejected = true
				break
			}
		}
		if isRejected {
			continue
		}

		cycle, interval := getBillingCycle(rt.interval)

		merchantName := rt.template.Description
		if rt.template.CounterpartyName != "" {
			merchantName = rt.template.CounterpartyName
		}

		// 4. AI Verification (Stricter Discovery)
		// Check if it's explicitly allowed (Whitelisted)
		isAllowed := false
		if fb, ok := feedbackMap[normalizeDescription(merchantName)]; ok {
			isAllowed = fb.Status == entity.DiscoveryStatusAllowed
		}

		if !isAllowed {
			// We use AI to verify if this is actually a subscription service.
			isSub, err := s.llm.VerifySubscriptionSuggestion(ctx, userID, merchantName, rt.template.Amount, "EUR", cycle)
			if err != nil {
				s.Logger.Warn("AI verification failed for suggestion, allowing as fallback", "merchant", merchantName, "error", err)
				// Fallback: show suggestion but don't cache in whitelist
			} else if !isSub {
				s.Logger.Info("AI rejected suggestion as non-subscription, permanently declining", "merchant", merchantName)
				_ = s.subRepo.SetDiscoveryFeedback(ctx, userID, merchantName, entity.DiscoveryStatusAIRejected, "AI")
				continue
			} else {
				// Cache the positive result to prevent future LLM calls
				s.Logger.Info("AI verified suggestion, adding to whitelist", "merchant", merchantName)
				_ = s.subRepo.SetDiscoveryFeedback(ctx, userID, merchantName, entity.DiscoveryStatusAllowed, "AI")
			}
		} else {
			s.Logger.Info("Merchant is in whitelist, bypassing AI verification", "merchant", merchantName)
		}

		suggestions = append(suggestions, entity.SuggestedSubscription{
			MerchantName:     merchantName,
			EstimatedAmount:  rt.template.Amount,
			Currency:         "EUR", // Defaulting to EUR for now
			BillingCycle:     cycle,
			BillingInterval:  interval,
			LastOccurrence:   rt.template.BookingDate,
			NextOccurrence:   rt.template.BookingDate.AddDate(0, rt.interval, 0),
			MatchingHashes:   rt.matchingHashes,
			BaseTransactions: rt.baseTransactions,
			CategoryID:       rt.template.CategoryID,
			BankAccountID:    rt.template.BankAccountID,
		})
	}

	// 5. Stable Sort (by Merchant Name)
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].MerchantName != suggestions[j].MerchantName {
			return suggestions[i].MerchantName < suggestions[j].MerchantName
		}
		return math.Abs(suggestions[i].EstimatedAmount) > math.Abs(suggestions[j].EstimatedAmount)
	})

	return suggestions, nil
}

// Internal helper for pattern detection
type discoveryPattern struct {
	template         entity.Transaction
	interval         int // months
	matchingHashes   []string
	baseTransactions []entity.BaseTransaction
}

func (s *DiscoveryService) detectRecurring(history []entity.Transaction, amountTolerance float64, minTxGeneric int, dateTolerance float64) []discoveryPattern {
	// 1. Initial Grouping: Mandate-first, then IBAN, then Description-fallback
	// This ensures that all transactions sharing a deterministic identifier are grouped together.

	groups := make(map[string][]entity.Transaction)
	identifierToGroup := make(map[string]string)

	for _, tx := range history {
		if tx.Amount >= 0 {
			continue
		}

		// If a transaction is already linked to a subscription, ignore it for pattern discovery.
		// This is the most robust way to prevent approved subscriptions from reappearing.
		if tx.SubscriptionID != nil {
			continue
		}

		normDesc := normalizeDescription(tx.Description)
		groupKey := normDesc

		// Deterministic Identifier Linkage (Mandate > IBAN)
		if tx.MandateReference != "" {
			if root, ok := identifierToGroup[tx.MandateReference]; ok {
				groupKey = root
			} else {
				identifierToGroup[tx.MandateReference] = normDesc
				groupKey = normDesc
			}
		} else if tx.CounterpartyIban != "" {
			if root, ok := identifierToGroup[tx.CounterpartyIban]; ok {
				groupKey = root
			} else {
				identifierToGroup[tx.CounterpartyIban] = normDesc
				groupKey = normDesc
			}
		}

		groups[groupKey] = append(groups[groupKey], tx)
	}

	var patterns []discoveryPattern
	for _, group := range groups {
		if len(group) < 2 {
			continue
		}

		sort.Slice(group, func(i, j int) bool {
			return group[i].BookingDate.Before(group[j].BookingDate)
		})

		// 2. Multi-Sequence Detection within each group
		// Note: For deterministic groups (same Mandate), we allow more amount flexibility
		// because a mandate is an unambiguous contract link.
		type seqState struct {
			txs      []entity.Transaction
			interval int
		}
		var activeSequences []*seqState

		for _, tx := range group {
			matched := false
			for _, seq := range activeSequences {
				last := seq.txs[len(seq.txs)-1]
				diff := tx.BookingDate.Sub(last.BookingDate)
				days := diff.Hours() / 24
				interval := getIntervalFromDays(days, dateTolerance)

				if interval > 0 && (seq.interval == 0 || interval == seq.interval) {
					// Check amount similarity
					amtDiff := math.Abs(tx.Amount - last.Amount)
					
					// If they share the SAME MandateReference, we allow much higher tolerance
					// because the mandate is the authoritative link.
					isSameMandate := tx.MandateReference != "" && tx.MandateReference == last.MandateReference
					effectiveTolerance := amountTolerance
					if isSameMandate {
						effectiveTolerance = 0.50 // Allow up to 50% change for mandate-linked payments (e.g. mobile data overage)
					}
					
					maxAllowed := math.Max(math.Abs(last.Amount)*effectiveTolerance, 10.0)

					if amtDiff <= maxAllowed || isSameMandate {
						seq.txs = append(seq.txs, tx)
						seq.interval = interval
						matched = true
						break
					}
				}
			}

			if !matched {
				activeSequences = append(activeSequences, &seqState{
					txs: []entity.Transaction{tx},
				})
			}
		}

		for _, seq := range activeSequences {
			if len(seq.txs) < 2 || seq.interval == 0 {
				continue
			}

			// Stricter threshold for generic payment processors
			merchant := seq.txs[len(seq.txs)-1].CounterpartyName
			if merchant == "" {
				merchant = seq.txs[len(seq.txs)-1].Description
			}
			if isGenericMerchant(merchant) && len(seq.txs) < minTxGeneric {
				continue
			}

			hashes := make([]string, len(seq.txs))
			baseTxns := make([]entity.BaseTransaction, len(seq.txs))
			for i, tx := range seq.txs {
				hashes[i] = tx.ContentHash
				baseTxns[i] = entity.BaseTransaction{
					Date:   tx.BookingDate,
					Amount: tx.Amount,
				}
			}
			patterns = append(patterns, discoveryPattern{
				template:         seq.txs[len(seq.txs)-1],
				interval:         seq.interval,
				matchingHashes:   hashes,
				baseTransactions: baseTxns,
			})
		}
	}

	// 3. Strict Name Deduplication
	return deduplicatePatterns(patterns, amountTolerance)
}

func deduplicatePatterns(patterns []discoveryPattern, amountTolerance float64) []discoveryPattern {
	var result []discoveryPattern

	for _, p := range patterns {
		normName := normalizeDescription(p.template.Description)
		if p.template.CounterpartyName != "" {
			normName = normalizeDescription(p.template.CounterpartyName)
		}

		merged := false
		for i, existing := range result {
			existingNorm := normalizeDescription(existing.template.Description)
			if existing.template.CounterpartyName != "" {
				existingNorm = normalizeDescription(existing.template.CounterpartyName)
			}

			// Merging criteria:
			// 1. Same MandateReference -> Unambiguous merge, always do it.
			// 2. Same normalized name AND similar amount AND similar description.
			
			isSameMandate := p.template.MandateReference != "" && p.template.MandateReference == existing.template.MandateReference
			
			pDesc := normalizeDescription(p.template.Description)
			eDesc := normalizeDescription(existing.template.Description)
			
			isSameMerchant := normName == existingNorm &&
				isAmountClose(p.template.Amount, existing.template.Amount, amountTolerance) &&
				pDesc == eDesc

			if isSameMandate || isSameMerchant {
				// Merge: Keep the one with more transactions
				if len(p.matchingHashes) > len(existing.matchingHashes) ||
					(len(p.matchingHashes) == len(existing.matchingHashes) && p.template.BookingDate.After(existing.template.BookingDate)) {
					
					// Transfer matching hashes to the winner to keep full history
					allHashes := append(existing.matchingHashes, p.matchingHashes...)
					// Simple deduplication (just in case)
					uniqueHashes := make(map[string]bool)
					finalHashes := []string{}
					for _, h := range allHashes {
						if !uniqueHashes[h] {
							uniqueHashes[h] = true
							finalHashes = append(finalHashes, h)
						}
					}
					
					p.matchingHashes = finalHashes
					result[i] = p
				} else {
					// Transfer p's hashes to existing
					allHashes := append(existing.matchingHashes, p.matchingHashes...)
					uniqueHashes := make(map[string]bool)
					finalHashes := []string{}
					for _, h := range allHashes {
						if !uniqueHashes[h] {
							uniqueHashes[h] = true
							finalHashes = append(finalHashes, h)
						}
					}
					existing.matchingHashes = finalHashes
					result[i] = existing
				}
				merged = true
				break
			}
		}

		if !merged {
			result = append(result, p)
		}
	}

	return result
}

func isGenericMerchant(name string) bool {
	lower := strings.ToLower(name)
	generic := []string{"first data", "stripe", "paypal", "adyen", "sumup", "worldpay", "klarna", "giropay", "apple.com/bill"}
	for _, g := range generic {
		if strings.Contains(lower, g) {
			return true
		}
	}
	return false
}

func getBillingCycle(months int) (string, int) {
	if months == 12 {
		return "yearly", 1
	}
	return "monthly", months
}
