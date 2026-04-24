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
	current.BillingInterval = sub.BillingInterval
	current.CustomerNumber = sub.CustomerNumber
	current.ContactEmail = sub.ContactEmail
	current.ContactPhone = sub.ContactPhone
	current.ContactWebsite = sub.ContactWebsite
	current.SupportURL = sub.SupportURL
	current.CancellationURL = sub.CancellationURL
	current.IsTrial = sub.IsTrial
	current.Notes = sub.Notes
	current.BankAccountID = sub.BankAccountID
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
		BankAccountID:   suggestion.BankAccountID,
		Status:          entity.SubscriptionStatusActive,
		LastOccurrence:  &suggestion.LastOccurrence,
		NextOccurrence:  &suggestion.NextOccurrence,
	}

	// 1. Broad Historical Backfill
	// The discovery engine is strict to avoid false positives, but once a user confirms
	// "Yes, this is a subscription", we want to link ALL historical transactions that match.
	txns, err := s.txRepo.FindTransactions(ctx, entity.TransactionFilter{
		UserID:        userID,
		IncludeShared: true,
	})
	if err == nil {
		normSub := normalizeDescription(suggestion.MerchantName)
		hashSet := make(map[string]bool)
		for _, h := range suggestion.MatchingHashes {
			hashSet[h] = true
		}

		mandateSet := make(map[string]bool)
		ibanSet := make(map[string]bool)

		// Fetch tolerance (same as GetSuggestedSubscriptions)
		amountTolerance := 0.10
		if val, err := s.settingsRepo.Get(ctx, "subscription_discovery_amount_tolerance", userID); err == nil && val != "" {
			if v, err := strconv.ParseFloat(val, 64); err == nil && v > 0 {
				amountTolerance = v
			}
		}

		for _, tx := range txns {
			// Link if: no sub yet, it's a debit, name matches (CP or Desc) AND amount is similar
			normTxDesc := normalizeDescription(tx.Description)
			normTxCP := ""
			if tx.CounterpartyName != "" {
				normTxCP = normalizeDescription(tx.CounterpartyName)
			}

			isNameMatch := (normTxDesc == normSub || (normTxCP != "" && normTxCP == normSub))
			if tx.SubscriptionID == nil && tx.Amount < 0 &&
				isNameMatch &&
				isAmountClose(tx.Amount, suggestion.EstimatedAmount, amountTolerance) {
				if !hashSet[tx.ContentHash] {
					hashSet[tx.ContentHash] = true
					suggestion.MatchingHashes = append(suggestion.MatchingHashes, tx.ContentHash)
				}

				// Extract deterministic identifiers
				if tx.MandateReference != "" {
					mandateSet[tx.MandateReference] = true
				}
				if tx.CounterpartyIban != "" {
					ibanSet[tx.CounterpartyIban] = true
				}
			}
		}

		// Update subscription with deterministic identifiers
		for m := range mandateSet {
			sub.LinkedMandates = append(sub.LinkedMandates, m)
		}
		for i := range ibanSet {
			sub.LinkedIbans = append(sub.LinkedIbans, i)
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
		IncludeShared:  true,
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

	// 3. Get user language
	language, _ := s.settingsRepo.Get(ctx, "ui_language", userID)
	if language == "" {
		language = "DE" // Default
	}

	// 4. Call AI to enrich
	enrichment, err := s.llm.EnrichSubscription(ctx, userID, sub.MerchantName, descriptions, language)
	if err != nil {
		return entity.Subscription{}, err
	}

	// 5. Update the subscription entity with new data
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

func (s *DiscoveryService) CreateSubscriptionFromTransaction(ctx context.Context, userID uuid.UUID, txnHash string, merchantName, billingCycle string, billingInterval int) (entity.Subscription, error) {
	// 1. Fetch the source transaction
	// We fetch ALL transactions for the user to ensure we find the one with the matching hash,
	// even if it's already reconciled or doesn't match a search proxy.
	txns, err := s.txRepo.FindTransactions(ctx, entity.TransactionFilter{
		UserID:        userID,
		IncludeShared: true,
	})
	if err != nil {
		return entity.Subscription{}, err
	}

	var sourceTx *entity.Transaction
	for i := range txns {
		if txns[i].ContentHash == txnHash {
			sourceTx = &txns[i]
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
	merchant := merchantName
	if merchant == "" {
		merchant = sourceTx.CounterpartyName
		if merchant == "" {
			merchant = sourceTx.Description
		}
	}

	sub := entity.Subscription{
		ID:              uuid.New(),
		UserID:          userID,
		MerchantName:    merchant,
		Amount:          sourceTx.Amount,
		Currency:        sourceTx.Currency,
		BillingCycle:    billingCycle,
		BillingInterval: billingInterval,
		Status:          entity.SubscriptionStatusActive,
		CategoryID:      sourceTx.CategoryID,
		BankAccountID:   sourceTx.BankAccountID,
		LastOccurrence:  &sourceTx.BookingDate,
		LinkedMandates:  []string{},
		LinkedIbans:     []string{},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// 4. Broad Historical Backfill
	// Similar to ApproveSubscription, we link all historical transactions that match this merchant.
	normSub := normalizeDescription(merchant)
	matchingHashes := []string{txnHash}
	hashSet := map[string]bool{txnHash: true}

	mandateSet := make(map[string]bool)
	ibanSet := make(map[string]bool)

	if sourceTx.MandateReference != "" {
		mandateSet[sourceTx.MandateReference] = true
	}
	if sourceTx.CounterpartyIban != "" {
		ibanSet[sourceTx.CounterpartyIban] = true
	}

	// Fetch tolerance
	amountTolerance := 0.10
	if val, err := s.settingsRepo.Get(ctx, "subscription_discovery_amount_tolerance", userID); err == nil && val != "" {
		if v, err := strconv.ParseFloat(val, 64); err == nil && v > 0 {
			amountTolerance = v
		}
	}

	for _, tx := range txns {
		if tx.ContentHash == txnHash {
			continue
		}

		// Link if: no sub yet, it's a debit, name matches (CP or Desc) AND amount is similar
		normTxDesc := normalizeDescription(tx.Description)
		normTxCP := ""
		if tx.CounterpartyName != "" {
			normTxCP = normalizeDescription(tx.CounterpartyName)
		}

		isNameMatch := (normTxDesc == normSub || (normTxCP != "" && normTxCP == normSub))
		if tx.SubscriptionID == nil && tx.Amount < 0 &&
			isNameMatch &&
			isAmountClose(tx.Amount, sourceTx.Amount, amountTolerance) {
			if !hashSet[tx.ContentHash] {
				hashSet[tx.ContentHash] = true
				matchingHashes = append(matchingHashes, tx.ContentHash)
			}

			// Extract deterministic identifiers
			if tx.MandateReference != "" {
				mandateSet[tx.MandateReference] = true
			}
			if tx.CounterpartyIban != "" {
				ibanSet[tx.CounterpartyIban] = true
			}
		}
	}

	// Update subscription with deterministic identifiers
	for m := range mandateSet {
		sub.LinkedMandates = append(sub.LinkedMandates, m)
	}
	for i := range ibanSet {
		sub.LinkedIbans = append(sub.LinkedIbans, i)
	}

	// 5. Create in DB and backfill matched transactions
	created, err := s.subRepo.CreateWithBackfill(ctx, sub, matchingHashes)
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

	// 2. Build maps for deterministic and fuzzy matching
	merchantMap := make(map[string][]entity.Subscription)
	mandateMap := make(map[string]uuid.UUID)
	ibanMap := make(map[string]uuid.UUID)
	hashMap := make(map[string]uuid.UUID)

	for _, sub := range subs {
		norm := normalizeDescription(sub.MerchantName)
		merchantMap[norm] = append(merchantMap[norm], sub)

		for _, h := range sub.MatchingHashes {
			hashMap[h] = sub.ID
		}
		for _, m := range sub.LinkedMandates {
			if m != "" {
				mandateMap[m] = sub.ID
			}
		}
		for _, i := range sub.LinkedIbans {
			if i != "" {
				ibanMap[i] = sub.ID
			}
		}
	}

	// 3. Match each transaction
	for _, tx := range txns {
		// Priority 1: Explicit matching_hashes (manual links)
		if subID, ok := hashMap[tx.ContentHash]; ok {
			s.Logger.Info("Matching transaction to subscription via explicit hash", "tx_hash", tx.ContentHash, "sub_id", subID, "user_id", userID)
			_ = s.txRepo.UpdateTransactionSubscription(ctx, tx.ContentHash, &subID, userID)
			continue
		}

		// Priority 2: Deterministic identifiers (Mandate or IBAN)
		matchedDeterministic := false
		if tx.MandateReference != "" {
			if subID, ok := mandateMap[tx.MandateReference]; ok {
				// Verify not ignored
				if !isIgnored(getSubByID(subs, subID), tx.ContentHash) {
					s.Logger.Info("Matching transaction to subscription via mandate", "tx_hash", tx.ContentHash, "sub_id", subID, "mandate", tx.MandateReference)
					_ = s.txRepo.UpdateTransactionSubscription(ctx, tx.ContentHash, &subID, userID)
					matchedDeterministic = true
				}
			}
		}
		if !matchedDeterministic && tx.CounterpartyIban != "" {
			if subID, ok := ibanMap[tx.CounterpartyIban]; ok {
				// Verify not ignored
				if !isIgnored(getSubByID(subs, subID), tx.ContentHash) {
					s.Logger.Info("Matching transaction to subscription via IBAN", "tx_hash", tx.ContentHash, "sub_id", subID, "iban", tx.CounterpartyIban)
					_ = s.txRepo.UpdateTransactionSubscription(ctx, tx.ContentHash, &subID, userID)
					matchedDeterministic = true
				}
			}
		}
		if matchedDeterministic {
			continue
		}

		// Priority 3: Fuzzy matching (normalized name + amount)
		normDesc := normalizeDescription(tx.Description)
		normCP := ""
		if tx.CounterpartyName != "" {
			normCP = normalizeDescription(tx.CounterpartyName)
		}

		matchedFuzzy := false
		// Try matching by Description first
		if candidates, ok := merchantMap[normDesc]; ok {
			for _, sub := range candidates {
				if isIgnored(sub, tx.ContentHash) {
					continue
				}
				if isAmountClose(tx.Amount, sub.Amount, 0.1) {
					s.Logger.Info("Matching transaction to subscription via fuzzy match (Description)", "tx_hash", tx.ContentHash, "sub_id", sub.ID)
					_ = s.txRepo.UpdateTransactionSubscription(ctx, tx.ContentHash, &sub.ID, userID)
					matchedFuzzy = true
					break
				}
			}
		}

		// If not matched, try matching by CounterpartyName
		if !matchedFuzzy && normCP != "" {
			if candidates, ok := merchantMap[normCP]; ok {
				for _, sub := range candidates {
					if isIgnored(sub, tx.ContentHash) {
						continue
					}
					if isAmountClose(tx.Amount, sub.Amount, 0.1) {
						s.Logger.Info("Matching transaction to subscription via fuzzy match (Counterparty)", "tx_hash", tx.ContentHash, "sub_id", sub.ID)
						_ = s.txRepo.UpdateTransactionSubscription(ctx, tx.ContentHash, &sub.ID, userID)
						matchedFuzzy = true
						break
					}
				}
			}
		}
	}

	return nil
}

func (s *DiscoveryService) LinkTransactions(ctx context.Context, userID, subID uuid.UUID, txnHashes []string) error {
	if len(txnHashes) == 0 {
		return nil
	}

	// 1. Fetch Subscription
	sub, err := s.subRepo.GetByID(ctx, subID, userID)
	if err != nil {
		return err
	}

	// 2. Modify IgnoredHashes and MatchingHashes
	ignoredMap := make(map[string]bool)
	for _, h := range sub.IgnoredHashes {
		ignoredMap[h] = true
	}

	matchingMap := make(map[string]bool)
	for _, h := range sub.MatchingHashes {
		matchingMap[h] = true
	}

	for _, hash := range txnHashes {
		delete(ignoredMap, hash)
		matchingMap[hash] = true
	}

	sub.IgnoredHashes = make([]string, 0, len(ignoredMap))
	for h := range ignoredMap {
		sub.IgnoredHashes = append(sub.IgnoredHashes, h)
	}

	sub.MatchingHashes = make([]string, 0, len(matchingMap))
	for h := range matchingMap {
		sub.MatchingHashes = append(sub.MatchingHashes, h)
	}

	// 3. Update Subscription
	if _, err := s.subRepo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription matching hashes: %w", err)
	}

	// 4. Link All Transactions
	for _, hash := range txnHashes {
		if err := s.txRepo.UpdateTransactionSubscription(ctx, hash, &subID, userID); err != nil {
			s.Logger.Warn("Failed to link transaction in batch", "hash", hash, "error", err)
			// We continue even if one fails, but it shouldn't ideally.
		}
	}

	// 5. Log Event
	_ = s.subRepo.LogEvent(ctx, entity.SubscriptionEvent{
		ID:             uuid.New(),
		SubscriptionID: subID,
		UserID:         userID,
		EventType:      "transaction_linked_manually",
		Title:          "Manual Batch Transaction Link",
		Content:        fmt.Sprintf("%d transactions were manually linked to this subscription.", len(txnHashes)),
		CreatedAt:      time.Now(),
	})

	return nil
}

func (s *DiscoveryService) LinkTransaction(ctx context.Context, userID, subID uuid.UUID, txnHash string) error {
	// 1. Fetch Subscription
	sub, err := s.subRepo.GetByID(ctx, subID, userID)
	if err != nil {
		return err
	}

	// 2. Remove from IgnoredHashes if present
	newIgnored := []string{}
	for _, h := range sub.IgnoredHashes {
		if h != txnHash {
			newIgnored = append(newIgnored, h)
		}
	}
	sub.IgnoredHashes = newIgnored

	// 3. Add to MatchingHashes if not present
	found := false
	for _, h := range sub.MatchingHashes {
		if h == txnHash {
			found = true
			break
		}
	}
	if !found {
		sub.MatchingHashes = append(sub.MatchingHashes, txnHash)
	}

	// 4. Update Subscription
	if _, err := s.subRepo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription matching hashes: %w", err)
	}

	// 5. Link Transaction
	if err := s.txRepo.UpdateTransactionSubscription(ctx, txnHash, &subID, userID); err != nil {
		return fmt.Errorf("failed to link transaction: %w", err)
	}

	// 6. Log Event
	_ = s.subRepo.LogEvent(ctx, entity.SubscriptionEvent{
		ID:             uuid.New(),
		SubscriptionID: subID,
		UserID:         userID,
		EventType:      "transaction_linked_manually",
		Title:          "Manual Transaction Link",
		Content:        fmt.Sprintf("Transaction %s was manually linked to this subscription.", txnHash),
		CreatedAt:      time.Now(),
	})

	return nil
}

func (s *DiscoveryService) UnlinkTransaction(ctx context.Context, userID, subID uuid.UUID, txnHash string) error {
	// 1. Fetch Subscription
	sub, err := s.subRepo.GetByID(ctx, subID, userID)
	if err != nil {
		return err
	}

	// 2. Remove from MatchingHashes if present
	newMatching := []string{}
	for _, h := range sub.MatchingHashes {
		if h != txnHash {
			newMatching = append(newMatching, h)
		}
	}
	sub.MatchingHashes = newMatching

	// 3. Add to IgnoredHashes if not present
	found := false
	for _, h := range sub.IgnoredHashes {
		if h == txnHash {
			found = true
			break
		}
	}
	if !found {
		sub.IgnoredHashes = append(sub.IgnoredHashes, txnHash)
	}

	// 4. Update Subscription
	if _, err := s.subRepo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription ignored hashes: %w", err)
	}

	// 5. Unlink Transaction (set subscription_id to NULL)
	if err := s.txRepo.UpdateTransactionSubscription(ctx, txnHash, nil, userID); err != nil {
		return fmt.Errorf("failed to unlink transaction: %w", err)
	}

	// 6. Log Event
	_ = s.subRepo.LogEvent(ctx, entity.SubscriptionEvent{
		ID:             uuid.New(),
		SubscriptionID: subID,
		UserID:         userID,
		EventType:      "transaction_unlinked_manually",
		Title:          "Manual Transaction Unlink",
		Content:        fmt.Sprintf("Transaction %s was manually unlinked and added to the ignore list.", txnHash),
		CreatedAt:      time.Now(),
	})

	return nil
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

func isIgnored(sub entity.Subscription, hash string) bool {
	if sub.ID == uuid.Nil {
		return false
	}
	for _, h := range sub.IgnoredHashes {
		if h == hash {
			return true
		}
	}
	return false
}

func getSubByID(subs []entity.Subscription, id uuid.UUID) entity.Subscription {
	for _, s := range subs {
		if s.ID == id {
			return s
		}
	}
	return entity.Subscription{}
}
