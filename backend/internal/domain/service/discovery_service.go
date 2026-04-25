package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
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

func (s *DiscoveryService) DeleteSubscription(ctx context.Context, userID, subID uuid.UUID) error {
	return s.subRepo.Delete(ctx, subID, userID)
}

func (s *DiscoveryService) GetSubscriptionEvents(ctx context.Context, userID, subID uuid.UUID) ([]entity.SubscriptionEvent, error) {
	return s.subRepo.GetEvents(ctx, subID, userID)
}
