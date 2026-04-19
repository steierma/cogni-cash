package service

import (
	"context"
	"errors"
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

var _ port.DiscoveryUseCase = (*DiscoveryService)(nil)

// DiscoveryService identifies recurring patterns and suggests them as subscriptions.
type DiscoveryService struct {
	txRepo       port.BankStatementRepository
	subRepo      port.SubscriptionRepository
	userRepo     port.UserRepository
	settingsRepo port.SettingsRepository
	llm          port.SubscriptionEnricher
	letterGen    port.CancellationLetterGenerator
	email        port.EmailProvider
	Logger       *slog.Logger
}

// NewDiscoveryService creates a new DiscoveryService.
func NewDiscoveryService(
	txRepo port.BankStatementRepository,
	subRepo port.SubscriptionRepository,
	userRepo port.UserRepository,
	settingsRepo port.SettingsRepository,
	llm port.SubscriptionEnricher,
	letterGen port.CancellationLetterGenerator,
	email port.EmailProvider,
	logger *slog.Logger,
) *DiscoveryService {
	if logger == nil {
		logger = slog.Default()
	}
	return &DiscoveryService{
		txRepo:       txRepo,
		subRepo:      subRepo,
		userRepo:     userRepo,
		settingsRepo: settingsRepo,
		llm:          llm,
		letterGen:    letterGen,
		email:        email,
		Logger:       logger,
	}
}

func (s *DiscoveryService) ListSubscriptions(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, error) {
	return s.subRepo.FindByUserID(ctx, userID)
}

func (s *DiscoveryService) GetSubscription(ctx context.Context, subID, userID uuid.UUID) (entity.Subscription, error) {
	return s.subRepo.GetByID(ctx, subID, userID)
}

func (s *DiscoveryService) UpdateSubscription(ctx context.Context, sub entity.Subscription) (entity.Subscription, error) {
	// 1. Fetch current version to ensure ownership and get existing data
	current, err := s.subRepo.GetByID(ctx, sub.ID, sub.UserID)
	if err != nil {
		return entity.Subscription{}, err
	}

	// 2. Map ONLY editable fields
	if current.Status != sub.Status && sub.Status != "" {
		_ = s.subRepo.LogEvent(ctx, entity.SubscriptionEvent{
			ID:             uuid.New(),
			SubscriptionID: current.ID,
			UserID:         current.UserID,
			EventType:      "status_changed",
			Title:          "Status Updated Manually",
			Content:        fmt.Sprintf("Status changed from %s to %s via manual edit.", current.Status, sub.Status),
			CreatedAt:      time.Now(),
		})
		current.Status = sub.Status
	}

	current.MerchantName = sub.MerchantName
	current.Amount = sub.Amount
	current.BillingCycle = sub.BillingCycle
	current.CustomerNumber = sub.CustomerNumber
	current.ContactEmail = sub.ContactEmail
	current.ContactPhone = sub.ContactPhone
	current.ContactWebsite = sub.ContactWebsite
	current.SupportURL = sub.SupportURL
	current.CancellationURL = sub.CancellationURL
	current.IsTrial = sub.IsTrial
	current.Notes = sub.Notes
	current.UpdatedAt = time.Now()

	// 3. Persist
	return s.subRepo.Update(ctx, current)
}

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
		UserID:   userID,
		FromDate: &histStart,
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

	// 3. Detect recurring patterns
	recurring := s.detectRecurring(history, amountTolerance, minTxGeneric, dateTolerance)

	var suggestions []entity.SuggestedSubscription
	for _, rt := range recurring {
		normName := normalizeDescription(rt.template.Description)

		// Filter out if already a subscription
		isExisting := false
		for _, sub := range existingSubs {
			if normalizeDescription(sub.MerchantName) == normName {
				isExisting = true
				break
			}
		}
		if isExisting {
			continue
		}

		// Filter based on feedback
		if fb, ok := feedbackMap[normName]; ok {
			if fb.Status == entity.DiscoveryStatusDeclined || fb.Status == entity.DiscoveryStatusAIRejected {
				continue
			}
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

func (s *DiscoveryService) ApproveSubscription(ctx context.Context, userID uuid.UUID, suggestion entity.SuggestedSubscription) (entity.Subscription, error) {
	sub := entity.Subscription{
		ID:              uuid.New(),
		UserID:          userID,
		MerchantName:    suggestion.MerchantName,
		Amount:          suggestion.EstimatedAmount,
		Currency:        suggestion.Currency,
		BillingCycle:    suggestion.BillingCycle,
		BillingInterval: suggestion.BillingInterval,
		CategoryID:      suggestion.CategoryID,
		Status:          entity.SubscriptionStatusActive,
		LastOccurrence:  &suggestion.LastOccurrence,
		NextOccurrence:  &suggestion.NextOccurrence,
	}

	// 1. Broad Historical Backfill
	// The discovery engine is strict to avoid false positives, but once a user confirms
	// "Yes, this is a subscription", we want to link ALL historical transactions that match.
	txns, err := s.txRepo.FindTransactions(ctx, entity.TransactionFilter{
		UserID: userID,
	})
	if err == nil {
		normSub := normalizeDescription(suggestion.MerchantName)
		hashSet := make(map[string]bool)
		for _, h := range suggestion.MatchingHashes {
			hashSet[h] = true
		}

		for _, tx := range txns {
			// Link if: no sub yet, it's a debit, and name matches normalized
			if tx.SubscriptionID == nil && tx.Amount < 0 && normalizeDescription(tx.Description) == normSub {
				if !hashSet[tx.ContentHash] {
					hashSet[tx.ContentHash] = true
					suggestion.MatchingHashes = append(suggestion.MatchingHashes, tx.ContentHash)
				}
			}
		}
	}

	sub, err = s.subRepo.CreateWithBackfill(ctx, sub, suggestion.MatchingHashes)
	if err != nil {
		return entity.Subscription{}, err
	}

	// Log creation event
	_ = s.subRepo.LogEvent(ctx, entity.SubscriptionEvent{
		ID:             uuid.New(),
		SubscriptionID: sub.ID,
		UserID:         userID,
		EventType:      "subscription_created",
		Title:          "Subscription Tracked",
		Content:        "Manually approved and tracked from discovery suggestions.",
		CreatedAt:      time.Now(),
	})

	return sub, nil
}

func (s *DiscoveryService) DeclineSuggestion(ctx context.Context, userID uuid.UUID, merchantName string) error {
	return s.subRepo.SetDiscoveryFeedback(ctx, userID, merchantName, entity.DiscoveryStatusDeclined, "USER")
}

func (s *DiscoveryService) GetDiscoveryFeedback(ctx context.Context, userID uuid.UUID) ([]entity.DiscoveryFeedback, error) {
	return s.subRepo.GetDiscoveryFeedback(ctx, userID)
}

func (s *DiscoveryService) RemoveDiscoveryFeedback(ctx context.Context, userID uuid.UUID, merchantName string) error {
	return s.subRepo.DeleteDiscoveryFeedback(ctx, userID, merchantName)
}

func (s *DiscoveryService) AllowSuggestion(ctx context.Context, userID uuid.UUID, merchantName string) error {
	return s.subRepo.SetDiscoveryFeedback(ctx, userID, merchantName, entity.DiscoveryStatusAllowed, "USER")
}

func (s *DiscoveryService) EnrichSubscription(ctx context.Context, userID, subID uuid.UUID) (entity.Subscription, error) {
	// 1. Fetch the existing subscription
	sub, err := s.subRepo.GetByID(ctx, subID, userID)
	if err != nil {
		return entity.Subscription{}, err
	}

	// 2. Fetch linked transactions to give AI more context (e.g., references)
	txns, err := s.txRepo.FindTransactions(ctx, entity.TransactionFilter{
		UserID:         userID,
		SubscriptionID: &subID,
	})
	if err != nil {
		s.Logger.Warn("Could not fetch linked transactions for enrichment", "error", err)
	}

	var descriptions []string
	for _, tx := range txns {
		if tx.Reference != "" {
			descriptions = append(descriptions, tx.Description+" (Ref: "+tx.Reference+")")
		} else {
			descriptions = append(descriptions, tx.Description)
		}
	}

	// 3. Call AI to enrich
	enrichment, err := s.llm.EnrichSubscription(ctx, userID, sub.MerchantName, descriptions)
	if err != nil {
		return entity.Subscription{}, err
	}

	// 4. Update the subscription entity with new data
	if enrichment.MerchantName != "" {
		sub.MerchantName = enrichment.MerchantName
	}
	if enrichment.CustomerNumber != "" {
		sub.CustomerNumber = &enrichment.CustomerNumber
	}
	if enrichment.ContactEmail != "" {
		sub.ContactEmail = &enrichment.ContactEmail
	}
	if enrichment.ContactPhone != "" {
		sub.ContactPhone = &enrichment.ContactPhone
	}
	if enrichment.ContactWebsite != "" {
		sub.ContactWebsite = &enrichment.ContactWebsite
	}
	if enrichment.SupportURL != "" {
		sub.SupportURL = &enrichment.SupportURL
	}
	if enrichment.CancellationURL != "" {
		sub.CancellationURL = &enrichment.CancellationURL
	}
	sub.IsTrial = enrichment.IsTrial
	if enrichment.Notes != "" {
		sub.Notes = &enrichment.Notes
	}
	sub.UpdatedAt = time.Now()

	// 5. Save the updated subscription
	updated, err := s.subRepo.Update(ctx, sub)
	if err != nil {
		return entity.Subscription{}, err
	}

	// 6. Log enrichment event
	_ = s.subRepo.LogEvent(ctx, entity.SubscriptionEvent{
		ID:             uuid.New(),
		SubscriptionID: subID,
		UserID:         userID,
		EventType:      "subscription_enriched",
		Title:          "AI Data Enrichment",
		Content:        "Subscription details (contact info, billing URLs, etc.) were successfully enriched using AI.",
		CreatedAt:      time.Now(),
	})

	return updated, nil
}

func (s *DiscoveryService) CreateSubscriptionFromTransaction(ctx context.Context, userID uuid.UUID, txnHash string, billingCycle string) (entity.Subscription, error) {
	// 1. Fetch the source transaction
	txns, err := s.txRepo.FindTransactions(ctx, entity.TransactionFilter{
		UserID: userID,
		Search: txnHash, // Use search as a proxy for finding by specific hash if dedicated method doesn't exist
	})
	if err != nil {
		return entity.Subscription{}, err
	}

	var sourceTx *entity.Transaction
	for _, t := range txns {
		if t.ContentHash == txnHash {
			sourceTx = &t
			break
		}
	}

	if sourceTx == nil {
		return entity.Subscription{}, fmt.Errorf("transaction not found: %s", txnHash)
	}

	// 2. Validate
	if sourceTx.SubscriptionID != nil {
		return entity.Subscription{}, fmt.Errorf("transaction is already part of a subscription")
	}

	// 3. Prepare Subscription details
	merchant := sourceTx.CounterpartyName
	if merchant == "" {
		merchant = sourceTx.Description
	}

	// Calculate billing interval
	interval := 1
	if billingCycle == "yearly" {
		interval = 12
	} else if billingCycle == "quarterly" {
		interval = 3
	}

	sub := entity.Subscription{
		ID:              uuid.New(),
		UserID:          userID,
		MerchantName:    merchant,
		Amount:          math.Abs(sourceTx.Amount),
		Currency:        sourceTx.Currency,
		BillingCycle:    billingCycle,
		BillingInterval: interval,
		Status:          entity.SubscriptionStatusActive,
		CategoryID:      sourceTx.CategoryID,
		LastOccurrence:  &sourceTx.BookingDate,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// 4. Create in DB and backfill this transaction
	created, err := s.subRepo.CreateWithBackfill(ctx, sub, []string{txnHash})
	if err != nil {
		return entity.Subscription{}, err
	}

	_ = s.subRepo.LogEvent(ctx, entity.SubscriptionEvent{
		ID:             uuid.New(),
		SubscriptionID: created.ID,
		UserID:         userID,
		EventType:      "subscription_created",
		Title:          "Manual Creation",
		Content:        fmt.Sprintf("Subscription created manually from transaction: %s", sourceTx.Description),
		CreatedAt:      time.Now(),
	})

	return created, nil
}

func (s *DiscoveryService) PreviewCancellation(ctx context.Context, userID, subID uuid.UUID, language string) (port.CancellationLetterResult, error) {
	// 1. Fetch Subscription
	sub, err := s.subRepo.GetByID(ctx, subID, userID)
	if err != nil {
		return port.CancellationLetterResult{}, err
	}

	// 2. Fetch User
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return port.CancellationLetterResult{}, err
	}

	// 3. Generate Draft
	endDate := ""
	if sub.ContractEndDate != nil {
		endDate = sub.ContractEndDate.Format("2006-01-02")
	}

	if language == "" {
		language = "DE" // Default
	}

	custNum := ""
	if sub.CustomerNumber != nil {
		custNum = *sub.CustomerNumber
	}
	noticeDays := 30
	if sub.NoticePeriodDays != nil {
		noticeDays = *sub.NoticePeriodDays
	}

	req := port.CancellationLetterRequest{
		UserFullName:     user.FullName,
		UserEmail:        user.Email,
		MerchantName:     sub.MerchantName,
		CustomerNumber:   custNum,
		ContractEndDate:  endDate,
		NoticePeriodDays: noticeDays,
		Language:         language,
	}

	return s.letterGen.GenerateCancellationLetter(ctx, userID, req)
}

func (s *DiscoveryService) CancelSubscription(ctx context.Context, userID, subID uuid.UUID, subject, body string) error {
	// 1. Fetch Subscription
	sub, err := s.subRepo.GetByID(ctx, subID, userID)
	if err != nil {
		return err
	}

	if sub.ContactEmail == nil || *sub.ContactEmail == "" {
		return errors.New("merchant contact email is missing")
	}

	// 2. Send Email
	err = s.email.Send(ctx, userID, *sub.ContactEmail, subject, body)
	if err != nil {
		return fmt.Errorf("failed to send cancellation email: %w", err)
	}

	// 3. Update Status
	sub.Status = entity.SubscriptionStatusCancellationPending
	sub.UpdatedAt = time.Now()
	_, err = s.subRepo.Update(ctx, sub)
	if err != nil {
		s.Logger.Error("Failed to update subscription status after cancellation", "error", err, "sub_id", subID)
	}

	// 4. Log Event
	event := entity.SubscriptionEvent{
		SubscriptionID: subID,
		UserID:         userID,
		EventType:      "cancellation_sent",
		Title:          "Cancellation Email Sent",
		Content:        fmt.Sprintf("To: %s\nSubject: %s\n\n%s", *sub.ContactEmail, subject, body),
	}
	_ = s.subRepo.LogEvent(ctx, event)

	return nil
}

func (s *DiscoveryService) DeleteSubscription(ctx context.Context, userID, subID uuid.UUID) error {
	return s.subRepo.Delete(ctx, subID, userID)
}

func (s *DiscoveryService) GetSubscriptionEvents(ctx context.Context, userID, subID uuid.UUID) ([]entity.SubscriptionEvent, error) {
	return s.subRepo.GetEvents(ctx, subID, userID)
}

func (s *DiscoveryService) MatchTransactions(ctx context.Context, userID uuid.UUID, txns []entity.Transaction) error {
	// 1. Fetch all active subscriptions for the user
	subs, err := s.subRepo.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user subscriptions: %w", err)
	}

	if len(subs) == 0 {
		return nil
	}

	// 2. Build a map of normalized merchant names to subscriptions
	merchantMap := make(map[string][]entity.Subscription)
	for _, sub := range subs {
		norm := normalizeDescription(sub.MerchantName)
		merchantMap[norm] = append(merchantMap[norm], sub)
	}

	// 3. Match each transaction
	for _, tx := range txns {
		normDesc := normalizeDescription(tx.Description)
		if candidates, ok := merchantMap[normDesc]; ok {
			for _, sub := range candidates {
				// Match if amount is close (within 10%)
				if isAmountClose(tx.Amount, sub.Amount, 0.1) {
					s.Logger.Info("Matching transaction to subscription", "tx_hash", tx.ContentHash, "sub_id", sub.ID, "user_id", userID)
					err := s.txRepo.UpdateTransactionSubscription(ctx, tx.ContentHash, &sub.ID, userID)
					if err != nil {
						s.Logger.Error("Failed to link transaction to subscription", "error", err, "tx_hash", tx.ContentHash, "sub_id", sub.ID)
					}
					break // Linked to the first matching candidate
				}
			}
		}
	}

	return nil
}

func isAmountClose(a, b, tolerancePercent float64) bool {
	if b == 0 {
		return math.Abs(a) < 0.01
	}
	diff := math.Abs(a - b)
	return diff <= math.Abs(b)*tolerancePercent
}

// Internal helper for pattern detection
type discoveryPattern struct {
	template         entity.Transaction
	interval         int // months
	matchingHashes   []string
	baseTransactions []entity.BaseTransaction
}

func (s *DiscoveryService) detectRecurring(history []entity.Transaction, amountTolerance float64, minTxGeneric int, dateTolerance float64) []discoveryPattern {
	// 1. Initial Grouping: IBAN-first with Description-fallback
	// This ensures that all transactions from the same account are grouped regardless of name noise.
	
	type groupNode struct {
		txs []entity.Transaction
	}
	
	// Composite key: IBAN or Normalized Description
	groups := make(map[string][]entity.Transaction)
	ibanToGroup := make(map[string]string)
	
	for _, tx := range history {
		if tx.Amount >= 0 {
			continue
		}
		
		normDesc := normalizeDescription(tx.Description)
		groupKey := normDesc
		
		if tx.CounterpartyIban != "" {
			if root, ok := ibanToGroup[tx.CounterpartyIban]; ok {
				groupKey = root
			} else {
				// New IBAN, link it to this description group
				ibanToGroup[tx.CounterpartyIban] = normDesc
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
					amtDiff := math.Abs(tx.Amount - last.Amount)
					maxAllowed := math.Max(math.Abs(last.Amount)*amountTolerance, 10.0)

					if amtDiff <= maxAllowed {
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
		// Use normalized template description for merging
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

			// Merging criteria: Same normalized name OR Same IBAN (implicit in clustering but good to double-check)
			// We prioritize name merge to fulfill "same name should never result in multiple suggestions"
			if normName == existingNorm {
				// Merge: Keep the one with more transactions (more likely to be accurate)
				// or the more recent one if counts are equal.
				if len(p.matchingHashes) > len(existing.matchingHashes) || 
				   (len(p.matchingHashes) == len(existing.matchingHashes) && p.template.BookingDate.After(existing.template.BookingDate)) {
					// Transfer matching hashes to the winner if we want to be exhaustive
					// For now, we just pick the better template
					result[i] = p
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


