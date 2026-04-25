package service

import (
	"context"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

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
