package service

import (
	"context"
	"fmt"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

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
