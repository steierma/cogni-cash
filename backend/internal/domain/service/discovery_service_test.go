package service_test

import (
	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"cogni-cash/internal/domain/port/mock"
	"cogni-cash/internal/domain/service"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	mockpkg "github.com/stretchr/testify/mock"
)

func mockDiscoverySettings(m *mock.MockSettingsRepository, ctx context.Context, userID uuid.UUID) {
	m.On("Get", ctx, "subscription_lookback_years", userID).Return("3", nil).Maybe()
	m.On("Get", ctx, "subscription_discovery_amount_tolerance", userID).Return("0.10", nil).Maybe()
	m.On("Get", ctx, "subscription_discovery_min_transactions_generic", userID).Return("3", nil).Maybe()
	m.On("Get", ctx, "subscription_discovery_date_tolerance", userID).Return("3.0", nil).Maybe()
	m.On("Get", ctx, "ui_language", userID).Return("DE", nil).Maybe()
}

// --- Tests ---

func TestDiscoveryService_EnrichSubscription(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	subID := uuid.New()

	t.Run("successful enrichment", func(t *testing.T) {
		mockSubRepo := new(mock.MockSubscriptionRepository)
		mockDiscoveryBankStmtRepo := new(mock.MockBankStatementRepository)
		mockUserRepoForDiscovery := new(mock.MockUserRepository)
		mockLLM := new(mock.MockSubscriptionEnricher)
		mockLetterGen := new(mock.MockCancellationLetterGenerator)
		mockEmail := new(mock.MockEmailProvider)
		mockSettingsRepo := new(mock.MockSettingsRepository)
		mockDiscoverySettings(mockSettingsRepo, ctx, userID)
		svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, mockUserRepoForDiscovery, mockSettingsRepo, mockLLM, mockLetterGen, mockEmail, slog.Default())

		existingSub := entity.Subscription{
			ID:           subID,
			UserID:       userID,
			MerchantName: "Netflix",
			Amount:       17.99,
		}

		mockSubRepo.On("GetByID", ctx, subID, userID).Return(existingSub, nil)

		linkedTxns := []entity.Transaction{
			{Description: "Netflix.com", Amount: -17.99, Reference: "SUB-123"},
		}
		mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mockpkg.MatchedBy(func(f entity.TransactionFilter) bool {
			return f.SubscriptionID != nil && *f.SubscriptionID == subID
		})).Return(linkedTxns, nil)

		enrichmentResult := port.SubscriptionEnrichmentResult{
			MerchantName:    "Netflix Inc.",
			CustomerNumber:  "C-12345",
			ContactWebsite:  "https://netflix.com",
			CancellationURL: "https://netflix.com/cancel",
			Notes:           "Streaming service",
		}
		mockLLM.On("EnrichSubscription", ctx, userID, "Netflix", []string{"Netflix.com (Ref: SUB-123)"}, "DE").
			Return(enrichmentResult, nil)

		mockSubRepo.On("Update", ctx, mockpkg.MatchedBy(func(s entity.Subscription) bool {
			return s.MerchantName == "Netflix Inc." &&
				s.CustomerNumber != nil && *s.CustomerNumber == "C-12345" &&
				s.Notes != nil && *s.Notes == "Streaming service"
		})).Return(entity.Subscription{MerchantName: "Netflix Inc."}, nil)

		mockSubRepo.On("LogEvent", ctx, mockpkg.MatchedBy(func(e entity.SubscriptionEvent) bool {
			return e.EventType == "subscription_enriched" && e.SubscriptionID == subID
		})).Return(nil)

		result, err := svc.EnrichSubscription(ctx, userID, subID)

		assert.NoError(t, err)
		assert.Equal(t, "Netflix Inc.", result.MerchantName)
		mockLLM.AssertExpectations(t)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("subscription not found", func(t *testing.T) {
		mockSubRepo := new(mock.MockSubscriptionRepository)
		mockDiscoveryBankStmtRepo := new(mock.MockBankStatementRepository)
		mockUserRepoForDiscovery := new(mock.MockUserRepository)
		mockLLM := new(mock.MockSubscriptionEnricher)
		mockLetterGen := new(mock.MockCancellationLetterGenerator)
		mockEmail := new(mock.MockEmailProvider)
		mockSettingsRepo := new(mock.MockSettingsRepository)
		mockDiscoverySettings(mockSettingsRepo, ctx, userID)
		svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, mockUserRepoForDiscovery, mockSettingsRepo, mockLLM, mockLetterGen, mockEmail, slog.Default())

		mockSubRepo.On("GetByID", ctx, subID, userID).Return(entity.Subscription{}, errors.New("not found"))

		_, err := svc.EnrichSubscription(ctx, userID, subID)

		assert.Error(t, err)
	})
}

func TestDiscoveryService_CancelSubscription(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	subID := uuid.New()

	t.Run("successful cancellation", func(t *testing.T) {
		mockSubRepo := new(mock.MockSubscriptionRepository)
		mockDiscoveryBankStmtRepo := new(mock.MockBankStatementRepository)
		mockUserRepoForDiscovery := new(mock.MockUserRepository)
		mockLLM := new(mock.MockSubscriptionEnricher)
		mockLetterGen := new(mock.MockCancellationLetterGenerator)
		mockEmail := new(mock.MockEmailProvider)
		mockSettingsRepo := new(mock.MockSettingsRepository)
		mockDiscoverySettings(mockSettingsRepo, ctx, userID)
		svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, mockUserRepoForDiscovery, mockSettingsRepo, mockLLM, mockLetterGen, mockEmail, slog.Default())

		email := "support@netflix.com"
		sub := entity.Subscription{
			ID:           subID,
			UserID:       userID,
			MerchantName: "Netflix",
			ContactEmail: &email,
		}

		mockSubRepo.On("GetByID", ctx, subID, userID).Return(sub, nil)
		mockEmail.On("Send", ctx, userID, "support@netflix.com", "Cancel my sub", "Please cancel").Return(nil)
		mockSubRepo.On("Update", ctx, mockpkg.MatchedBy(func(s entity.Subscription) bool {
			return s.Status == entity.SubscriptionStatusCancellationPending
		})).Return(entity.Subscription{}, nil)
		mockSubRepo.On("LogEvent", ctx, mockpkg.MatchedBy(func(e entity.SubscriptionEvent) bool {
			return e.EventType == "cancellation_sent" && e.SubscriptionID == subID
		})).Return(nil)

		err := svc.CancelSubscription(ctx, userID, subID, "Cancel my sub", "Please cancel")

		assert.NoError(t, err)
		mockEmail.AssertExpectations(t)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("missing contact email", func(t *testing.T) {
		mockSubRepo := new(mock.MockSubscriptionRepository)
		mockDiscoveryBankStmtRepo := new(mock.MockBankStatementRepository)
		mockUserRepoForDiscovery := new(mock.MockUserRepository)
		mockLLM := new(mock.MockSubscriptionEnricher)
		mockLetterGen := new(mock.MockCancellationLetterGenerator)
		mockEmail := new(mock.MockEmailProvider)
		mockSettingsRepo := new(mock.MockSettingsRepository)
		mockDiscoverySettings(mockSettingsRepo, ctx, userID)
		svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, mockUserRepoForDiscovery, mockSettingsRepo, mockLLM, mockLetterGen, mockEmail, slog.Default())

		sub := entity.Subscription{
			ID:           subID,
			UserID:       userID,
			MerchantName: "Netflix",
			ContactEmail: nil,
		}

		mockSubRepo.On("GetByID", ctx, subID, userID).Return(sub, nil)

		err := svc.CancelSubscription(ctx, userID, subID, "Cancel my sub", "Please cancel")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "contact email is missing")
	})
}

func TestDiscoveryService_DeleteSubscription(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	subID := uuid.New()

	t.Run("successful deletion", func(t *testing.T) {
		mockSubRepo := new(mock.MockSubscriptionRepository)
		mockDiscoveryBankStmtRepo := new(mock.MockBankStatementRepository)
		mockUserRepoForDiscovery := new(mock.MockUserRepository)
		mockLLM := new(mock.MockSubscriptionEnricher)
		mockLetterGen := new(mock.MockCancellationLetterGenerator)
		mockEmail := new(mock.MockEmailProvider)
		mockSettingsRepo := new(mock.MockSettingsRepository)
		mockDiscoverySettings(mockSettingsRepo, ctx, userID)
		svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, mockUserRepoForDiscovery, mockSettingsRepo, mockLLM, mockLetterGen, mockEmail, slog.Default())

		mockSubRepo.On("Delete", ctx, subID, userID).Return(nil)

		err := svc.DeleteSubscription(ctx, userID, subID)

		assert.NoError(t, err)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockSubRepo := new(mock.MockSubscriptionRepository)
		mockDiscoveryBankStmtRepo := new(mock.MockBankStatementRepository)
		mockUserRepoForDiscovery := new(mock.MockUserRepository)
		mockLLM := new(mock.MockSubscriptionEnricher)
		mockLetterGen := new(mock.MockCancellationLetterGenerator)
		mockEmail := new(mock.MockEmailProvider)
		mockSettingsRepo := new(mock.MockSettingsRepository)
		mockDiscoverySettings(mockSettingsRepo, ctx, userID)
		svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, mockUserRepoForDiscovery, mockSettingsRepo, mockLLM, mockLetterGen, mockEmail, slog.Default())

		mockSubRepo.On("Delete", ctx, subID, userID).Return(errors.New("delete failed"))

		err := svc.DeleteSubscription(ctx, userID, subID)

		assert.Error(t, err)
		assert.Equal(t, "delete failed", err.Error())
	})
}

func TestDiscoveryService_CreateSubscriptionFromTransaction(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	txnHash := "test-hash"

	t.Run("creates subscription with correct negative amount for debit", func(t *testing.T) {
		mockSubRepo := new(mock.MockSubscriptionRepository)
		mockTxRepo := new(mock.MockBankStatementRepository)
		mockSettingsRepo := new(mock.MockSettingsRepository)
		mockDiscoverySettings(mockSettingsRepo, ctx, userID)
		svc := service.NewDiscoveryService(mockTxRepo, mockSubRepo, nil, mockSettingsRepo, nil, nil, nil, slog.Default())

		sourceTx := entity.Transaction{
			ContentHash:      txnHash,
			Amount:           -49.99,
			Currency:         "EUR",
			CounterpartyName: "Test Merchant",
			BookingDate:      time.Now(),
		}

		mockTxRepo.On("FindTransactions", ctx, mockpkg.MatchedBy(func(f entity.TransactionFilter) bool {
			return f.UserID == userID && f.IncludeShared == true
		})).Return([]entity.Transaction{sourceTx}, nil)

		mockSubRepo.On("CreateWithBackfill", ctx, mockpkg.MatchedBy(func(s entity.Subscription) bool {
			// CRITICAL: Ensure amount is -49.99, NOT 49.99
			return s.MerchantName == "Test Merchant" && s.Amount == -49.99
		}), []string{txnHash}).Return(entity.Subscription{Amount: -49.99}, nil)

		mockSubRepo.On("LogEvent", ctx, mockpkg.MatchedBy(func(e entity.SubscriptionEvent) bool {
			return e.EventType == "subscription_created"
		})).Return(nil)

		result, err := svc.CreateSubscriptionFromTransaction(ctx, userID, txnHash, "", "monthly", 1)

		assert.NoError(t, err)
		assert.Equal(t, -49.99, result.Amount)
		mockSubRepo.AssertExpectations(t)
		mockTxRepo.AssertExpectations(t)
		mockSettingsRepo.AssertExpectations(t)
	})
}

func TestDiscoveryService_ApproveSubscription(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("successful approval logs creation event, backfills, and enriches", func(t *testing.T) {
		mockSubRepo := new(mock.MockSubscriptionRepository)
		mockDiscoveryBankStmtRepo := new(mock.MockBankStatementRepository)
		mockSettingsRepo := new(mock.MockSettingsRepository)
		mockLLM := new(mock.MockSubscriptionEnricher)
		svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, nil, mockSettingsRepo, mockLLM, nil, nil, slog.Default())

		mockSettingsRepo.On("Get", ctx, "subscription_discovery_amount_tolerance", userID).Return("0.1", nil)

		suggestion := entity.SuggestedSubscription{
			MerchantName:    "Netflix",
			EstimatedAmount: -17.99,
			Currency:        "EUR",
			MatchingHashes:  []string{"hash1"},
		}

		subID := uuid.New()
		createdSub := entity.Subscription{ID: subID, UserID: userID, MerchantName: "Netflix", Amount: -17.99}

		// Mock broad history finding an extra unlinked transaction
		history := []entity.Transaction{
			{ContentHash: "hash1", Description: "Netflix", Amount: -17.99, SubscriptionID: nil},
			{ContentHash: "hash-extra", Description: "Lastschrift Netflix", Amount: -17.99, SubscriptionID: nil},
			{ContentHash: "hash-ignore-credit", Description: "Netflix", Amount: 17.99, SubscriptionID: nil},
		}
		mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mockpkg.MatchedBy(func(f entity.TransactionFilter) bool {
			return f.UserID == userID && f.IncludeShared == true
		})).Return(history, nil).Once()

		// CreateWithBackfill should now receive BOTH hashes
		mockSubRepo.On("CreateWithBackfill", ctx, mockpkg.MatchedBy(func(s entity.Subscription) bool {
			return s.MerchantName == "Netflix" && s.Amount == -17.99
		}), mockpkg.MatchedBy(func(hashes []string) bool {
			return len(hashes) == 2 && hashes[0] == "hash1" && hashes[1] == "hash-extra"
		})).Return(createdSub, nil)

		mockSubRepo.On("LogEvent", ctx, mockpkg.MatchedBy(func(e entity.SubscriptionEvent) bool {
			return e.EventType == "subscription_created" && e.Title == "Subscription Tracked"
		})).Return(nil)

		result, err := svc.ApproveSubscription(ctx, userID, suggestion)

		assert.NoError(t, err)
		assert.Equal(t, "Netflix", result.MerchantName)
		mockSubRepo.AssertExpectations(t)
		mockDiscoveryBankStmtRepo.AssertExpectations(t)
		mockLLM.AssertExpectations(t)
	})
}

func TestDiscoveryService_UpdateSubscription(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	subID := uuid.New()

	mockSubRepo := new(mock.MockSubscriptionRepository)
	svc := service.NewDiscoveryService(nil, mockSubRepo, nil, nil, nil, nil, nil, slog.Default())

	currentSub := entity.Subscription{
		ID:           subID,
		UserID:       userID,
		MerchantName: "Old Name",
		Amount:       10.0,
		Status:       "active",
	}

	t.Run("successful update with status change", func(t *testing.T) {
		updateData := entity.Subscription{
			ID:           subID,
			UserID:       userID,
			MerchantName: "New Name",
			Amount:       15.0,
			Status:       "cancelled",
		}

		mockSubRepo.On("GetByID", ctx, subID, userID).Return(currentSub, nil).Once()
		mockSubRepo.On("LogEvent", ctx, mockpkg.MatchedBy(func(ev entity.SubscriptionEvent) bool {
			return ev.SubscriptionID == subID && ev.EventType == "status_changed"
		})).Return(nil).Once()
		mockSubRepo.On("Update", ctx, mockpkg.MatchedBy(func(s entity.Subscription) bool {
			return s.MerchantName == "New Name" && s.Status == "cancelled" && s.Amount == 15.0
		})).Return(updateData, nil).Once()

		updated, err := svc.UpdateSubscription(ctx, updateData)

		assert.NoError(t, err)
		assert.Equal(t, "cancelled", string(updated.Status))
		mockSubRepo.AssertExpectations(t)
	})
}

func TestDiscoveryService_DiscoveryLogic(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	setup := func() (*mock.MockSubscriptionRepository, *mock.MockBankStatementRepository, *mock.MockSubscriptionEnricher, *mock.MockSettingsRepository, *service.DiscoveryService) {
		mockSubRepo := new(mock.MockSubscriptionRepository)
		mockDiscoveryBankStmtRepo := new(mock.MockBankStatementRepository)
		mockLLM := new(mock.MockSubscriptionEnricher)
		mockSettingsRepo := new(mock.MockSettingsRepository)
		mockDiscoverySettings(mockSettingsRepo, ctx, userID)
		svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, nil, mockSettingsRepo, mockLLM, nil, nil, slog.Default())

		mockSubRepo.On("FindByUserID", ctx, userID).Return([]entity.Subscription{}, nil)

		pDate := func(s string) time.Time {
			t, _ := time.Parse("2006-01-02", s)
			return t
		}
		history := []entity.Transaction{
			{Description: "Netflix.com", Amount: -15.99, BookingDate: pDate("2024-01-01")},
			{Description: "Netflix.com", Amount: -15.99, BookingDate: pDate("2024-02-01")},
			{Description: "Netflix.com", Amount: -15.99, BookingDate: pDate("2024-03-01")},
		}
		mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mockpkg.MatchedBy(func(f entity.TransactionFilter) bool {
			return f.UserID == userID
		})).Return(history, nil)

		return mockSubRepo, mockDiscoveryBankStmtRepo, mockLLM, mockSettingsRepo, svc
	}

	t.Run("bypasses AI if whitelisted", func(t *testing.T) {
		mockSubRepo, _, mockLLM, _, svc := setup()
		mockSubRepo.On("GetDiscoveryFeedback", ctx, userID).Return([]entity.DiscoveryFeedback{
			{MerchantName: "Netflix.com", Status: entity.DiscoveryStatusAllowed},
		}, nil).Once()

		suggestions, err := svc.GetSuggestedSubscriptions(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, suggestions, 1)
		mockLLM.AssertNotCalled(t, "VerifySubscriptionSuggestion", mockpkg.Anything, mockpkg.Anything, mockpkg.Anything, mockpkg.Anything, mockpkg.Anything, mockpkg.Anything)
	})

	t.Run("caches positive result from AI", func(t *testing.T) {
		mockSubRepo, _, mockLLM, _, svc := setup()
		mockSubRepo.On("GetDiscoveryFeedback", ctx, userID).Return([]entity.DiscoveryFeedback{}, nil).Once()
		mockLLM.On("VerifySubscriptionSuggestion", ctx, userID, "Netflix.com", -15.99, "EUR", "monthly").Return(true, nil).Once()
		mockSubRepo.On("SetDiscoveryFeedback", ctx, userID, "Netflix.com", entity.DiscoveryStatusAllowed, "AI").Return(nil).Once()

		suggestions, err := svc.GetSuggestedSubscriptions(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, suggestions, 1)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("falls back if AI fails", func(t *testing.T) {
		mockSubRepo, _, mockLLM, _, svc := setup()
		mockSubRepo.On("GetDiscoveryFeedback", ctx, userID).Return([]entity.DiscoveryFeedback{}, nil).Once()
		mockLLM.On("VerifySubscriptionSuggestion", ctx, userID, "Netflix.com", -15.99, "EUR", "monthly").Return(false, errors.New("ai down")).Once()

		suggestions, err := svc.GetSuggestedSubscriptions(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, suggestions, 1)
		mockSubRepo.AssertNotCalled(t, "SetDiscoveryFeedback", mockpkg.Anything, mockpkg.Anything, mockpkg.Anything, mockpkg.Anything, mockpkg.Anything, mockpkg.Anything)
	})
}

func TestDiscoveryService_AllowSuggestion(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	merchant := "Netflix.com"

	mockSubRepo := new(mock.MockSubscriptionRepository)
	svc := service.NewDiscoveryService(nil, mockSubRepo, nil, nil, nil, nil, nil, slog.Default())

	t.Run("adds to whitelist", func(t *testing.T) {
		mockSubRepo.On("SetDiscoveryFeedback", ctx, userID, merchant, entity.DiscoveryStatusAllowed, "USER").Return(nil).Once()

		err := svc.AllowSuggestion(ctx, userID, merchant)
		assert.NoError(t, err)
		mockSubRepo.AssertExpectations(t)
	})
}

func TestDiscoveryService_DiscoveryLogic_IBANGrouping(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockSubRepo := new(mock.MockSubscriptionRepository)
	mockDiscoveryBankStmtRepo := new(mock.MockBankStatementRepository)
	mockLLM := new(mock.MockSubscriptionEnricher)
	mockSettingsRepo := new(mock.MockSettingsRepository)
	mockDiscoverySettings(mockSettingsRepo, ctx, userID)
	svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, nil, mockSettingsRepo, mockLLM, nil, nil, slog.Default())

	mockSubRepo.On("FindByUserID", ctx, userID).Return([]entity.Subscription{}, nil)
	mockSubRepo.On("GetDiscoveryFeedback", ctx, userID).Return([]entity.DiscoveryFeedback{}, nil)

	pDate := func(s string) time.Time {
		t, _ := time.Parse("2006-01-02", s)
		return t
	}

	// Two transactions with different descriptions but same IBAN
	history := []entity.Transaction{
		{
			Description:      "Dauerauftrag/Terminueberw. Max Mustermann",
			CounterpartyName: "Max Mustermann",
			CounterpartyIban: "DE00123456789012345678",
			Amount:           -500.0,
			BookingDate:      pDate("2024-01-01"),
			ContentHash:      "h1",
		},
		{
			Description:      "Miete",
			CounterpartyName: "Max Mustermann",
			CounterpartyIban: "DE00123456789012345678",
			Amount:           -500.0,
			BookingDate:      pDate("2024-02-01"),
			ContentHash:      "h2",
		},
		{
			Description:      "Miete",
			CounterpartyName: "Max Mustermann",
			CounterpartyIban: "DE00123456789012345678",
			Amount:           -500.0,
			BookingDate:      pDate("2024-03-01"),
			ContentHash:      "h3",
		},
	}
	mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mockpkg.Anything).Return(history, nil)

	// AI verification mock - should be called for the preferred name "Max Mustermann"
	mockLLM.On("VerifySubscriptionSuggestion", ctx, userID, "Max Mustermann", -500.0, "EUR", "monthly").Return(true, nil)
	mockSubRepo.On("SetDiscoveryFeedback", ctx, userID, "Max Mustermann", entity.DiscoveryStatusAllowed, "AI").Return(nil)

	suggestions, err := svc.GetSuggestedSubscriptions(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, suggestions, 1, "Should have exactly one suggestion due to IBAN grouping")
	assert.Equal(t, "Max Mustermann", suggestions[0].MerchantName, "Should prefer CounterpartyName")
	assert.Len(t, suggestions[0].MatchingHashes, 3, "Should include all 3 transactions in the group")
}

func TestDiscoveryService_DiscoveryLogic_FuzzyAmountGrouping(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockSubRepo := new(mock.MockSubscriptionRepository)
	mockDiscoveryBankStmtRepo := new(mock.MockBankStatementRepository)
	mockLLM := new(mock.MockSubscriptionEnricher)
	mockSettingsRepo := new(mock.MockSettingsRepository)
	mockDiscoverySettings(mockSettingsRepo, ctx, userID)
	svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, nil, mockSettingsRepo, mockLLM, nil, nil, slog.Default())

	mockSubRepo.On("FindByUserID", ctx, userID).Return([]entity.Subscription{}, nil)
	mockSubRepo.On("GetDiscoveryFeedback", ctx, userID).Return([]entity.DiscoveryFeedback{}, nil)

	pDate := func(s string) time.Time {
		t, _ := time.Parse("2006-01-02", s)
		return t
	}

	// Two transactions with slightly different amounts
	history := []entity.Transaction{
		{
			Description:      "Barmenia Versicherung",
			CounterpartyName: "Barmenia",
			Amount:           -50.00,
			BookingDate:      pDate("2024-01-01"),
			ContentHash:      "h1",
		},
		{
			Description:      "Barmenia Versicherung",
			CounterpartyName: "Barmenia",
			Amount:           -51.50, // +3%, within the 10% tolerance
			BookingDate:      pDate("2024-02-01"),
			ContentHash:      "h2",
		},
		{
			Description:      "Barmenia Versicherung",
			CounterpartyName: "Barmenia",
			Amount:           -52.00, // +4%, within the 10% tolerance
			BookingDate:      pDate("2024-03-01"),
			ContentHash:      "h3",
		},
	}
	mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mockpkg.Anything).Return(history, nil)

	// AI verification mock
	mockLLM.On("VerifySubscriptionSuggestion", ctx, userID, "Barmenia", -52.0, "EUR", "monthly").Return(true, nil)
	mockSubRepo.On("SetDiscoveryFeedback", ctx, userID, "Barmenia", entity.DiscoveryStatusAllowed, "AI").Return(nil)

	suggestions, err := svc.GetSuggestedSubscriptions(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, suggestions, 1, "Should have exactly one suggestion due to fuzzy amount grouping")
	assert.Equal(t, "Barmenia", suggestions[0].MerchantName)
	assert.Equal(t, -52.0, suggestions[0].EstimatedAmount, "Should keep the most recent amount")
	assert.Len(t, suggestions[0].MatchingHashes, 3, "Should include all 3 transactions in the group")
}

func TestDiscoveryService_DiscoveryLogic_MultipleSubscriptionsSameMerchant(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockSubRepo := new(mock.MockSubscriptionRepository)
	mockDiscoveryBankStmtRepo := new(mock.MockBankStatementRepository)
	mockLLM := new(mock.MockSubscriptionEnricher)
	mockSettingsRepo := new(mock.MockSettingsRepository)

	// Mock settings
	mockSettingsRepo.On("Get", ctx, "subscription_lookback_years", userID).Return("3", nil)
	mockSettingsRepo.On("Get", ctx, "subscription_discovery_amount_tolerance", userID).Return("0.1", nil)
	mockSettingsRepo.On("Get", ctx, "subscription_discovery_min_transactions_generic", userID).Return("3", nil)
	mockSettingsRepo.On("Get", ctx, "subscription_discovery_date_tolerance", userID).Return("3", nil)

	svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, nil, mockSettingsRepo, mockLLM, nil, nil, slog.Default())

	mockSubRepo.On("FindByUserID", ctx, userID).Return([]entity.Subscription{}, nil)
	mockSubRepo.On("GetDiscoveryFeedback", ctx, userID).Return([]entity.DiscoveryFeedback{}, nil)

	pDate := func(s string) time.Time {
		t, _ := time.Parse("2006-01-02", s)
		return t
	}

	// Two distinct sequences for Barmenia with very different amounts
	history := []entity.Transaction{
		// Sequence 1: 50.00 EUR
		{Description: "Barmenia Vertrag 1", CounterpartyName: "Barmenia", Amount: -50.00, BookingDate: pDate("2024-01-01"), ContentHash: "h1_1"},
		{Description: "Barmenia Vertrag 1", CounterpartyName: "Barmenia", Amount: -50.00, BookingDate: pDate("2024-02-01"), ContentHash: "h1_2"},
		{Description: "Barmenia Vertrag 1", CounterpartyName: "Barmenia", Amount: -50.00, BookingDate: pDate("2024-03-01"), ContentHash: "h1_3"},

		// Sequence 2: 150.00 EUR
		{Description: "Barmenia Vertrag 2", CounterpartyName: "Barmenia", Amount: -150.00, BookingDate: pDate("2024-01-05"), ContentHash: "h2_1"},
		{Description: "Barmenia Vertrag 2", CounterpartyName: "Barmenia", Amount: -150.00, BookingDate: pDate("2024-02-05"), ContentHash: "h2_2"},
		{Description: "Barmenia Vertrag 2", CounterpartyName: "Barmenia", Amount: -150.00, BookingDate: pDate("2024-03-05"), ContentHash: "h2_3"},
	}
	mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mockpkg.Anything).Return(history, nil)

	// AI verification mock - should be called for both if they are distinct
	mockLLM.On("VerifySubscriptionSuggestion", ctx, userID, "Barmenia", -50.0, "EUR", "monthly").Return(true, nil)
	mockLLM.On("VerifySubscriptionSuggestion", ctx, userID, "Barmenia", -150.0, "EUR", "monthly").Return(true, nil)

	mockSubRepo.On("SetDiscoveryFeedback", ctx, userID, "Barmenia", entity.DiscoveryStatusAllowed, "AI").Return(nil)

	suggestions, err := svc.GetSuggestedSubscriptions(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(suggestions), "Should have TWO suggestions for Barmenia since they have different amounts")
}

func TestDiscoveryService_DiscoveryLogic_MultipleSubscriptionsSameMerchantSameAmount(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockSubRepo := new(mock.MockSubscriptionRepository)
	mockDiscoveryBankStmtRepo := new(mock.MockBankStatementRepository)
	mockLLM := new(mock.MockSubscriptionEnricher)
	mockSettingsRepo := new(mock.MockSettingsRepository)

	// Mock settings
	mockSettingsRepo.On("Get", ctx, "subscription_lookback_years", userID).Return("3", nil)
	mockSettingsRepo.On("Get", ctx, "subscription_discovery_amount_tolerance", userID).Return("0.1", nil)
	mockSettingsRepo.On("Get", ctx, "subscription_discovery_min_transactions_generic", userID).Return("3", nil)
	mockSettingsRepo.On("Get", ctx, "subscription_discovery_date_tolerance", userID).Return("3", nil)

	svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, nil, mockSettingsRepo, mockLLM, nil, nil, slog.Default())

	mockSubRepo.On("FindByUserID", ctx, userID).Return([]entity.Subscription{}, nil)
	mockSubRepo.On("GetDiscoveryFeedback", ctx, userID).Return([]entity.DiscoveryFeedback{}, nil)

	pDate := func(s string) time.Time {
		t, _ := time.Parse("2006-01-02", s)
		return t
	}

	// Two distinct sequences for Barmenia with SAME amounts but DIFFERENT contract numbers in description
	history := []entity.Transaction{
		// Sequence 1: Vertrag 123
		{Description: "Barmenia Vertrag 123", CounterpartyName: "Barmenia", Amount: -50.00, BookingDate: pDate("2024-01-01"), ContentHash: "h1_1"},
		{Description: "Barmenia Vertrag 123", CounterpartyName: "Barmenia", Amount: -50.00, BookingDate: pDate("2024-02-01"), ContentHash: "h1_2"},
		{Description: "Barmenia Vertrag 123", CounterpartyName: "Barmenia", Amount: -50.00, BookingDate: pDate("2024-03-01"), ContentHash: "h1_3"},

		// Sequence 2: Vertrag 456
		{Description: "Barmenia Vertrag 456", CounterpartyName: "Barmenia", Amount: -50.00, BookingDate: pDate("2024-01-01"), ContentHash: "h2_1"},
		{Description: "Barmenia Vertrag 456", CounterpartyName: "Barmenia", Amount: -50.00, BookingDate: pDate("2024-02-01"), ContentHash: "h2_2"},
		{Description: "Barmenia Vertrag 456", CounterpartyName: "Barmenia", Amount: -50.00, BookingDate: pDate("2024-03-01"), ContentHash: "h2_3"},
	}
	mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mockpkg.Anything).Return(history, nil)

	// AI verification mock
	mockLLM.On("VerifySubscriptionSuggestion", ctx, userID, "Barmenia", -50.0, "EUR", "monthly").Return(true, nil)

	mockSubRepo.On("SetDiscoveryFeedback", ctx, userID, "Barmenia", entity.DiscoveryStatusAllowed, "AI").Return(nil)

	suggestions, err := svc.GetSuggestedSubscriptions(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(suggestions), "Should have TWO suggestions for Barmenia since they have different contract numbers even if amount is same")
}

func TestDiscoveryService_DiscoveryLogic_MergingSimilarDescriptions(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockSubRepo := new(mock.MockSubscriptionRepository)
	mockDiscoveryBankStmtRepo := new(mock.MockBankStatementRepository)
	mockLLM := new(mock.MockSubscriptionEnricher)
	mockSettingsRepo := new(mock.MockSettingsRepository)

	// Mock settings
	mockSettingsRepo.On("Get", ctx, "subscription_lookback_years", userID).Return("3", nil)
	mockSettingsRepo.On("Get", ctx, "subscription_discovery_amount_tolerance", userID).Return("0.1", nil)
	mockSettingsRepo.On("Get", ctx, "subscription_discovery_min_transactions_generic", userID).Return("3", nil)
	mockSettingsRepo.On("Get", ctx, "subscription_discovery_date_tolerance", userID).Return("3", nil)

	svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, nil, mockSettingsRepo, mockLLM, nil, nil, slog.Default())

	mockSubRepo.On("FindByUserID", ctx, userID).Return([]entity.Subscription{}, nil)
	mockSubRepo.On("GetDiscoveryFeedback", ctx, userID).Return([]entity.DiscoveryFeedback{}, nil)

	pDate := func(s string) time.Time {
		t, _ := time.Parse("2006-01-02", s)
		return t
	}

	history := []entity.Transaction{
		{Description: "Netflix", CounterpartyName: "Netflix", Amount: -17.99, BookingDate: pDate("2024-01-01"), ContentHash: "n1_1"},
		{Description: "Netflix", CounterpartyName: "Netflix", Amount: -17.99, BookingDate: pDate("2024-02-01"), ContentHash: "n1_2"},
		{Description: "Netflix", CounterpartyName: "Netflix", Amount: -17.99, BookingDate: pDate("2024-03-01"), ContentHash: "n1_3"},

		{Description: "Lastschrift Netflix", CounterpartyName: "Netflix", Amount: -17.99, BookingDate: pDate("2024-01-05"), ContentHash: "n2_1"},
		{Description: "Lastschrift Netflix", CounterpartyName: "Netflix", Amount: -17.99, BookingDate: pDate("2024-02-05"), ContentHash: "n2_2"},
		{Description: "Lastschrift Netflix", CounterpartyName: "Netflix", Amount: -17.99, BookingDate: pDate("2024-03-05"), ContentHash: "n2_3"},
	}
	mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mockpkg.Anything).Return(history, nil)

	mockLLM.On("VerifySubscriptionSuggestion", ctx, userID, "Netflix", -17.99, "EUR", "monthly").Return(true, nil)
	mockSubRepo.On("SetDiscoveryFeedback", ctx, userID, "Netflix", entity.DiscoveryStatusAllowed, "AI").Return(nil)

	suggestions, err := svc.GetSuggestedSubscriptions(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(suggestions), "Should merge Netflix suggestions because they normalize to the same name and description")
}

func TestDiscoveryService_DiscoveryLogic_ConsolidateMandateWithVaryingAmounts(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockSubRepo := new(mock.MockSubscriptionRepository)
	mockDiscoveryBankStmtRepo := new(mock.MockBankStatementRepository)
	mockLLM := new(mock.MockSubscriptionEnricher)
	mockSettingsRepo := new(mock.MockSettingsRepository)
	mockDiscoverySettings(mockSettingsRepo, ctx, userID)
	svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, nil, mockSettingsRepo, mockLLM, nil, nil, slog.Default())

	mockSubRepo.On("FindByUserID", ctx, userID).Return([]entity.Subscription{}, nil)
	mockSubRepo.On("GetDiscoveryFeedback", ctx, userID).Return([]entity.DiscoveryFeedback{}, nil)

	pDate := func(s string) time.Time {
		t, _ := time.Parse("2006-01-02", s)
		return t
	}

	mandate := "DE000201000200000000000000011781284"
	history := []entity.Transaction{
		{Description: "Telekom", Amount: -74.95, BookingDate: pDate("2026-04-16"), MandateReference: mandate, ContentHash: "h1"},
		{Description: "Telekom", Amount: -74.95, BookingDate: pDate("2026-03-16"), MandateReference: mandate, ContentHash: "h2"},
		{Description: "Telekom", Amount: -90.26, BookingDate: pDate("2026-02-16"), MandateReference: mandate, ContentHash: "h3"},
		{Description: "Telekom", Amount: -49.94, BookingDate: pDate("2026-01-16"), MandateReference: mandate, ContentHash: "h4"},
	}
	mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mockpkg.Anything).Return(history, nil)

	mockLLM.On("VerifySubscriptionSuggestion", ctx, userID, "Telekom", -74.95, "EUR", "monthly").Return(true, nil)
	mockSubRepo.On("SetDiscoveryFeedback", ctx, userID, "Telekom", entity.DiscoveryStatusAllowed, "AI").Return(nil)

	suggestions, err := svc.GetSuggestedSubscriptions(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, suggestions, 1, "Should have exactly ONE suggestion for the same mandate despite varying amounts")
}

func TestDiscoveryService_MatchTransactions_Deterministic(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	subID := uuid.New()

	mockSubRepo := new(mock.MockSubscriptionRepository)
	mockDiscoveryBankStmtRepo := new(mock.MockBankStatementRepository)
	svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, nil, nil, nil, nil, nil, slog.Default())

	sub := entity.Subscription{
		ID:             subID,
		UserID:         userID,
		MerchantName:   "Barmenia",
		Amount:         -62.65,
		LinkedMandates: []string{"MANDATE_123"},
		LinkedIbans:    []string{"DE1234567890"},
		MatchingHashes: []string{"MANUAL_HASH"},
		IgnoredHashes:  []string{"IGNORED_HASH"},
	}

	mockSubRepo.On("FindByUserID", ctx, userID).Return([]entity.Subscription{sub}, nil)

	txns := []entity.Transaction{
		// Priority 1: Manual Link
		{ContentHash: "MANUAL_HASH", Description: "Something else entirely", Amount: -1000.0, MandateReference: "MANDATE_WRONG"},

		// Priority 2: Mandate (Price changed, name same)
		{ContentHash: "MANDATE_MATCH", Description: "Barmenia", Amount: -65.00, MandateReference: "MANDATE_123"},

		// Priority 2: IBAN (Price same, name same)
		{ContentHash: "IBAN_MATCH", Description: "Barmenia", Amount: -62.65, CounterpartyIban: "DE1234567890"},

		// Priority 3: Fuzzy (Price close, name same)
		{ContentHash: "FUZZY_MATCH", Description: "Barmenia", Amount: -62.00},

		// Ignored: Should NOT match even if mandate is correct
		{ContentHash: "IGNORED_HASH", Description: "Barmenia", Amount: -62.65, MandateReference: "MANDATE_123"},

		// No match
		{ContentHash: "NO_MATCH", Description: "Something else", Amount: -10.00},
	}

	// Expectations
	mockDiscoveryBankStmtRepo.On("UpdateTransactionSubscription", ctx, "MANUAL_HASH", &subID, userID).Return(nil)
	mockDiscoveryBankStmtRepo.On("UpdateTransactionSubscription", ctx, "MANDATE_MATCH", &subID, userID).Return(nil)
	mockDiscoveryBankStmtRepo.On("UpdateTransactionSubscription", ctx, "IBAN_MATCH", &subID, userID).Return(nil)
	mockDiscoveryBankStmtRepo.On("UpdateTransactionSubscription", ctx, "FUZZY_MATCH", &subID, userID).Return(nil)

	err := svc.MatchTransactions(ctx, userID, txns)

	assert.NoError(t, err)
	mockDiscoveryBankStmtRepo.AssertExpectations(t)
	mockSubRepo.AssertExpectations(t)
}

func TestDiscoveryService_MatchTransactions_Fuzzy(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	subID := uuid.New()

	mockSubRepo := new(mock.MockSubscriptionRepository)
	mockDiscoveryBankStmtRepo := new(mock.MockBankStatementRepository)
	svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, nil, nil, nil, nil, nil, slog.Default())

	sub := entity.Subscription{
		ID:           subID,
		UserID:       userID,
		MerchantName: "Netflix",
		Amount:       -17.99,
	}

	mockSubRepo.On("FindByUserID", ctx, userID).Return([]entity.Subscription{sub}, nil)

	txns := []entity.Transaction{
		{ContentHash: "h1", Description: "Netflix", Amount: -17.99},
	}

	mockDiscoveryBankStmtRepo.On("UpdateTransactionSubscription", ctx, "h1", mockpkg.MatchedBy(func(id *uuid.UUID) bool {
		return id != nil && *id == subID
	}), userID).Return(nil)

	err := svc.MatchTransactions(ctx, userID, txns)

	assert.NoError(t, err)
	mockDiscoveryBankStmtRepo.AssertExpectations(t)
}
