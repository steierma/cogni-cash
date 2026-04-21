package service_test

import (
	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"cogni-cash/internal/domain/service"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type mockEnricher struct {
	mock.Mock
}

func (m *mockEnricher) EnrichSubscription(ctx context.Context, userID uuid.UUID, merchantName string, transactionDescriptions []string) (port.SubscriptionEnrichmentResult, error) {
	args := m.Called(ctx, userID, merchantName, transactionDescriptions)
	return args.Get(0).(port.SubscriptionEnrichmentResult), args.Error(1)
}

func (m *mockEnricher) VerifySubscriptionSuggestion(ctx context.Context, userID uuid.UUID, merchantName string, amount float64, currency string, billingCycle string) (bool, error) {
	args := m.Called(ctx, userID, merchantName, amount, currency, billingCycle)
	return args.Bool(0), args.Error(1)
}

type mockLetterGen struct {
	mock.Mock
}

func (m *mockLetterGen) GenerateCancellationLetter(ctx context.Context, userID uuid.UUID, req port.CancellationLetterRequest) (port.CancellationLetterResult, error) {
	args := m.Called(ctx, userID, req)
	return args.Get(0).(port.CancellationLetterResult), args.Error(1)
}

type mockEmailProviderForDiscovery struct {
	mock.Mock
}

func (m *mockEmailProviderForDiscovery) Send(ctx context.Context, userID uuid.UUID, to, subject, body string) error {
	args := m.Called(ctx, userID, to, subject, body)
	return args.Error(0)
}

type mockUserRepoForDiscovery struct {
	mock.Mock
}

func (m *mockUserRepoForDiscovery) FindByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entity.User), args.Error(1)
}

func (m *mockUserRepoForDiscovery) FindAll(ctx context.Context, search string) ([]entity.User, error) {
	args := m.Called(ctx, search)
	return args.Get(0).([]entity.User), args.Error(1)
}

func (m *mockUserRepoForDiscovery) FindByUsername(ctx context.Context, username string) (entity.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(entity.User), args.Error(1)
}

func (m *mockUserRepoForDiscovery) GetAdminID(ctx context.Context) (uuid.UUID, error) {
	args := m.Called(ctx)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockUserRepoForDiscovery) Create(ctx context.Context, user entity.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *mockUserRepoForDiscovery) Upsert(ctx context.Context, user entity.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *mockUserRepoForDiscovery) Update(ctx context.Context, user entity.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *mockUserRepoForDiscovery) UpdatePassword(ctx context.Context, userID uuid.UUID, newHash string) error {
	return m.Called(ctx, userID, newHash).Error(0)
}

func (m *mockUserRepoForDiscovery) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type mockSettingsRepoForDiscovery struct {
	mock.Mock
}

func (m *mockSettingsRepoForDiscovery) Get(ctx context.Context, key string, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, key, userID)
	return args.String(0), args.Error(1)
}

func (m *mockSettingsRepoForDiscovery) Set(ctx context.Context, key, value string, userID uuid.UUID, isSensitive bool) error {
	return m.Called(ctx, key, value, userID, isSensitive).Error(0)
}

func (m *mockSettingsRepoForDiscovery) GetAll(ctx context.Context, userID uuid.UUID) (map[string]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(map[string]string), args.Error(1)
}

func mockDiscoverySettings(m *mockSettingsRepoForDiscovery, ctx context.Context, userID uuid.UUID) {
	m.On("Get", ctx, "subscription_lookback_years", userID).Return("3", nil).Maybe()
	m.On("Get", ctx, "subscription_discovery_amount_tolerance", userID).Return("0.10", nil).Maybe()
	m.On("Get", ctx, "subscription_discovery_min_transactions_generic", userID).Return("3", nil).Maybe()
	m.On("Get", ctx, "subscription_discovery_date_tolerance", userID).Return("3.0", nil).Maybe()
}

type mockSubscriptionRepo struct {
	mock.Mock
}

func (m *mockSubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Subscription, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(entity.Subscription), args.Error(1)
}

func (m *mockSubscriptionRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.Subscription), args.Error(1)
}

func (m *mockSubscriptionRepo) Create(ctx context.Context, sub entity.Subscription) (entity.Subscription, error) {
	args := m.Called(ctx, sub)
	return args.Get(0).(entity.Subscription), args.Error(1)
}

func (m *mockSubscriptionRepo) CreateWithBackfill(ctx context.Context, sub entity.Subscription, matchingHashes []string) (entity.Subscription, error) {
	args := m.Called(ctx, sub, matchingHashes)
	return args.Get(0).(entity.Subscription), args.Error(1)
}

func (m *mockSubscriptionRepo) Update(ctx context.Context, sub entity.Subscription) (entity.Subscription, error) {
	args := m.Called(ctx, sub)
	return args.Get(0).(entity.Subscription), args.Error(1)
}

func (m *mockSubscriptionRepo) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *mockSubscriptionRepo) LogEvent(ctx context.Context, event entity.SubscriptionEvent) error {
	return m.Called(ctx, event).Error(0)
}

func (m *mockSubscriptionRepo) GetEvents(ctx context.Context, subID uuid.UUID, userID uuid.UUID) ([]entity.SubscriptionEvent, error) {
	args := m.Called(ctx, subID, userID)
	return args.Get(0).([]entity.SubscriptionEvent), args.Error(1)
}

func (m *mockSubscriptionRepo) SetDiscoveryFeedback(ctx context.Context, userID uuid.UUID, merchantName string, status entity.DiscoveryFeedbackStatus, source string) error {
	return m.Called(ctx, userID, merchantName, status, source).Error(0)
}

func (m *mockSubscriptionRepo) GetDiscoveryFeedback(ctx context.Context, userID uuid.UUID) ([]entity.DiscoveryFeedback, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.DiscoveryFeedback), args.Error(1)
}

func (m *mockSubscriptionRepo) DeleteDiscoveryFeedback(ctx context.Context, userID uuid.UUID, merchantName string) error {
	return m.Called(ctx, userID, merchantName).Error(0)
}

type mockDiscoveryBankStmtRepo struct {
	mock.Mock
}

func (m *mockDiscoveryBankStmtRepo) Save(ctx context.Context, stmt entity.BankStatement) error {
	return m.Called(ctx, stmt).Error(0)
}
func (m *mockDiscoveryBankStmtRepo) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.BankStatement, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(entity.BankStatement), args.Error(1)
}
func (m *mockDiscoveryBankStmtRepo) FindAll(ctx context.Context, userID uuid.UUID) ([]entity.BankStatement, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.BankStatement), args.Error(1)
}
func (m *mockDiscoveryBankStmtRepo) FindSummaries(ctx context.Context, userID uuid.UUID) ([]entity.BankStatementSummary, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.BankStatementSummary), args.Error(1)
}
func (m *mockDiscoveryBankStmtRepo) FindTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Transaction), args.Error(1)
}
func (m *mockDiscoveryBankStmtRepo) SearchTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Transaction), args.Error(1)
}
func (m *mockDiscoveryBankStmtRepo) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}
func (m *mockDiscoveryBankStmtRepo) UpdateTransactionCategory(ctx context.Context, hash string, categoryID *uuid.UUID, userID uuid.UUID) error {
	return m.Called(ctx, hash, categoryID, userID).Error(0)
}
func (m *mockDiscoveryBankStmtRepo) UpdateTransactionSubscription(ctx context.Context, hash string, subID *uuid.UUID, userID uuid.UUID) error {
	return m.Called(ctx, hash, subID, userID).Error(0)
}
func (m *mockDiscoveryBankStmtRepo) MarkTransactionReviewed(ctx context.Context, hash string, userID uuid.UUID) error {
	return m.Called(ctx, hash, userID).Error(0)
}
func (m *mockDiscoveryBankStmtRepo) MarkTransactionReconciled(ctx context.Context, hash string, reconID uuid.UUID, userID uuid.UUID) error {
	return m.Called(ctx, hash, reconID, userID).Error(0)
}
func (m *mockDiscoveryBankStmtRepo) LinkTransactionToStatement(ctx context.Context, id uuid.UUID, stmtID uuid.UUID, userID uuid.UUID) error {
	return m.Called(ctx, id, stmtID, userID).Error(0)
}
func (m *mockDiscoveryBankStmtRepo) CreateTransactions(ctx context.Context, txns []entity.Transaction) error {
	return m.Called(ctx, txns).Error(0)
}
func (m *mockDiscoveryBankStmtRepo) GetCategorizationExamples(ctx context.Context, userID uuid.UUID, count int) ([]entity.CategorizationExample, error) {
	args := m.Called(ctx, userID, count)
	return args.Get(0).([]entity.CategorizationExample), args.Error(1)
}
func (m *mockDiscoveryBankStmtRepo) FindMatchingCategory(ctx context.Context, userID uuid.UUID, txn port.TransactionToCategorize) (*uuid.UUID, error) {
	args := m.Called(ctx, userID, txn)
	return args.Get(0).(*uuid.UUID), args.Error(1)
}
func (m *mockDiscoveryBankStmtRepo) UpdateTransactionSkipForecasting(ctx context.Context, hash string, skip bool, userID uuid.UUID) error {
	return m.Called(ctx, hash, skip, userID).Error(0)
}
func (m *mockDiscoveryBankStmtRepo) UpdateTransactionBaseAmount(ctx context.Context, hash string, baseAmount float64, baseCurrency string, userID uuid.UUID) error {
	return m.Called(ctx, hash, baseAmount, baseCurrency, userID).Error(0)
}

// --- Tests ---

func TestDiscoveryService_EnrichSubscription(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	subID := uuid.New()

	t.Run("successful enrichment", func(t *testing.T) {
		mockSubRepo := new(mockSubscriptionRepo)
		mockDiscoveryBankStmtRepo := new(mockDiscoveryBankStmtRepo)
		mockUserRepoForDiscovery := new(mockUserRepoForDiscovery)
		mockLLM := new(mockEnricher)
		mockLetterGen := new(mockLetterGen)
		mockEmail := new(mockEmailProviderForDiscovery)
		mockSettingsRepo := new(mockSettingsRepoForDiscovery)
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
		mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mock.MatchedBy(func(f entity.TransactionFilter) bool {
			return f.SubscriptionID != nil && *f.SubscriptionID == subID
		})).Return(linkedTxns, nil)

		enrichmentResult := port.SubscriptionEnrichmentResult{
			MerchantName:    "Netflix Inc.",
			CustomerNumber:  "C-12345",
			ContactWebsite:  "https://netflix.com",
			CancellationURL: "https://netflix.com/cancel",
			Notes:           "Streaming service",
		}
		mockLLM.On("EnrichSubscription", ctx, userID, "Netflix", []string{"Netflix.com (Ref: SUB-123)"}).
			Return(enrichmentResult, nil)

		mockSubRepo.On("Update", ctx, mock.MatchedBy(func(s entity.Subscription) bool {
			return s.MerchantName == "Netflix Inc." &&
				s.CustomerNumber != nil && *s.CustomerNumber == "C-12345" &&
				s.Notes != nil && *s.Notes == "Streaming service"
		})).Return(entity.Subscription{MerchantName: "Netflix Inc."}, nil)

		mockSubRepo.On("LogEvent", ctx, mock.MatchedBy(func(e entity.SubscriptionEvent) bool {
			return e.EventType == "subscription_enriched" && e.SubscriptionID == subID
		})).Return(nil)

		result, err := svc.EnrichSubscription(ctx, userID, subID)

		assert.NoError(t, err)
		assert.Equal(t, "Netflix Inc.", result.MerchantName)
		mockLLM.AssertExpectations(t)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("subscription not found", func(t *testing.T) {
		mockSubRepo := new(mockSubscriptionRepo)
		mockDiscoveryBankStmtRepo := new(mockDiscoveryBankStmtRepo)
		mockUserRepoForDiscovery := new(mockUserRepoForDiscovery)
		mockLLM := new(mockEnricher)
		mockLetterGen := new(mockLetterGen)
		mockEmail := new(mockEmailProviderForDiscovery)
		mockSettingsRepo := new(mockSettingsRepoForDiscovery)
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
		mockSubRepo := new(mockSubscriptionRepo)
		mockDiscoveryBankStmtRepo := new(mockDiscoveryBankStmtRepo)
		mockUserRepoForDiscovery := new(mockUserRepoForDiscovery)
		mockLLM := new(mockEnricher)
		mockLetterGen := new(mockLetterGen)
		mockEmail := new(mockEmailProviderForDiscovery)
		mockSettingsRepo := new(mockSettingsRepoForDiscovery)
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
		mockSubRepo.On("Update", ctx, mock.MatchedBy(func(s entity.Subscription) bool {
			return s.Status == entity.SubscriptionStatusCancellationPending
		})).Return(entity.Subscription{}, nil)
		mockSubRepo.On("LogEvent", ctx, mock.MatchedBy(func(e entity.SubscriptionEvent) bool {
			return e.EventType == "cancellation_sent" && e.SubscriptionID == subID
		})).Return(nil)

		err := svc.CancelSubscription(ctx, userID, subID, "Cancel my sub", "Please cancel")

		assert.NoError(t, err)
		mockEmail.AssertExpectations(t)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("missing contact email", func(t *testing.T) {
		mockSubRepo := new(mockSubscriptionRepo)
		mockDiscoveryBankStmtRepo := new(mockDiscoveryBankStmtRepo)
		mockUserRepoForDiscovery := new(mockUserRepoForDiscovery)
		mockLLM := new(mockEnricher)
		mockLetterGen := new(mockLetterGen)
		mockEmail := new(mockEmailProviderForDiscovery)
		mockSettingsRepo := new(mockSettingsRepoForDiscovery)
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
		mockSubRepo := new(mockSubscriptionRepo)
		mockDiscoveryBankStmtRepo := new(mockDiscoveryBankStmtRepo)
		mockUserRepoForDiscovery := new(mockUserRepoForDiscovery)
		mockLLM := new(mockEnricher)
		mockLetterGen := new(mockLetterGen)
		mockEmail := new(mockEmailProviderForDiscovery)
		mockSettingsRepo := new(mockSettingsRepoForDiscovery)
		mockDiscoverySettings(mockSettingsRepo, ctx, userID)
		svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, mockUserRepoForDiscovery, mockSettingsRepo, mockLLM, mockLetterGen, mockEmail, slog.Default())

		mockSubRepo.On("Delete", ctx, subID, userID).Return(nil)

		err := svc.DeleteSubscription(ctx, userID, subID)

		assert.NoError(t, err)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockSubRepo := new(mockSubscriptionRepo)
		mockDiscoveryBankStmtRepo := new(mockDiscoveryBankStmtRepo)
		mockUserRepoForDiscovery := new(mockUserRepoForDiscovery)
		mockLLM := new(mockEnricher)
		mockLetterGen := new(mockLetterGen)
		mockEmail := new(mockEmailProviderForDiscovery)
		mockSettingsRepo := new(mockSettingsRepoForDiscovery)
		mockDiscoverySettings(mockSettingsRepo, ctx, userID)
		svc := service.NewDiscoveryService(mockDiscoveryBankStmtRepo, mockSubRepo, mockUserRepoForDiscovery, mockSettingsRepo, mockLLM, mockLetterGen, mockEmail, slog.Default())

		mockSubRepo.On("Delete", ctx, subID, userID).Return(errors.New("delete failed"))

		err := svc.DeleteSubscription(ctx, userID, subID)

		assert.Error(t, err)
		assert.Equal(t, "delete failed", err.Error())
	})
}

func TestDiscoveryService_ApproveSubscription(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("successful approval logs creation event, backfills, and enriches", func(t *testing.T) {
		mockSubRepo := new(mockSubscriptionRepo)
		mockDiscoveryBankStmtRepo := new(mockDiscoveryBankStmtRepo)
		mockSettingsRepo := new(mockSettingsRepoForDiscovery)
		mockLLM := new(mockEnricher)
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
		mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mock.MatchedBy(func(f entity.TransactionFilter) bool {
			return f.UserID == userID && f.SubscriptionID == nil
		})).Return(history, nil).Once()

		// CreateWithBackfill should now receive BOTH hashes
		mockSubRepo.On("CreateWithBackfill", ctx, mock.MatchedBy(func(s entity.Subscription) bool {
			return s.MerchantName == "Netflix" && s.Amount == -17.99
		}), mock.MatchedBy(func(hashes []string) bool {
			return len(hashes) == 2 && hashes[0] == "hash1" && hashes[1] == "hash-extra"
		})).Return(createdSub, nil)

		mockSubRepo.On("LogEvent", ctx, mock.MatchedBy(func(e entity.SubscriptionEvent) bool {
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

	mockSubRepo := new(mockSubscriptionRepo)
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
		mockSubRepo.On("LogEvent", ctx, mock.MatchedBy(func(ev entity.SubscriptionEvent) bool {
			return ev.SubscriptionID == subID && ev.EventType == "status_changed"
		})).Return(nil).Once()
		mockSubRepo.On("Update", ctx, mock.MatchedBy(func(s entity.Subscription) bool {
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

	setup := func() (*mockSubscriptionRepo, *mockDiscoveryBankStmtRepo, *mockEnricher, *mockSettingsRepoForDiscovery, *service.DiscoveryService) {
		mockSubRepo := new(mockSubscriptionRepo)
		mockDiscoveryBankStmtRepo := new(mockDiscoveryBankStmtRepo)
		mockLLM := new(mockEnricher)
		mockSettingsRepo := new(mockSettingsRepoForDiscovery)
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
		mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mock.MatchedBy(func(f entity.TransactionFilter) bool {
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
		mockLLM.AssertNotCalled(t, "VerifySubscriptionSuggestion", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
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
		mockSubRepo.AssertNotCalled(t, "SetDiscoveryFeedback", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestDiscoveryService_AllowSuggestion(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	merchant := "Netflix.com"

	mockSubRepo := new(mockSubscriptionRepo)
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

	mockSubRepo := new(mockSubscriptionRepo)
	mockDiscoveryBankStmtRepo := new(mockDiscoveryBankStmtRepo)
	mockLLM := new(mockEnricher)
	mockSettingsRepo := new(mockSettingsRepoForDiscovery)
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
	mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mock.Anything).Return(history, nil)

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

	mockSubRepo := new(mockSubscriptionRepo)
	mockDiscoveryBankStmtRepo := new(mockDiscoveryBankStmtRepo)
	mockLLM := new(mockEnricher)
	mockSettingsRepo := new(mockSettingsRepoForDiscovery)
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
	mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mock.Anything).Return(history, nil)

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

	mockSubRepo := new(mockSubscriptionRepo)
	mockDiscoveryBankStmtRepo := new(mockDiscoveryBankStmtRepo)
	mockLLM := new(mockEnricher)
	mockSettingsRepo := new(mockSettingsRepoForDiscovery)

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
	mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mock.Anything).Return(history, nil)

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

	mockSubRepo := new(mockSubscriptionRepo)
	mockDiscoveryBankStmtRepo := new(mockDiscoveryBankStmtRepo)
	mockLLM := new(mockEnricher)
	mockSettingsRepo := new(mockSettingsRepoForDiscovery)

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
	mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mock.Anything).Return(history, nil)

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

	mockSubRepo := new(mockSubscriptionRepo)
	mockDiscoveryBankStmtRepo := new(mockDiscoveryBankStmtRepo)
	mockLLM := new(mockEnricher)
	mockSettingsRepo := new(mockSettingsRepoForDiscovery)

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
	mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mock.Anything).Return(history, nil)

	mockLLM.On("VerifySubscriptionSuggestion", ctx, userID, "Netflix", -17.99, "EUR", "monthly").Return(true, nil)
	mockSubRepo.On("SetDiscoveryFeedback", ctx, userID, "Netflix", entity.DiscoveryStatusAllowed, "AI").Return(nil)

	suggestions, err := svc.GetSuggestedSubscriptions(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(suggestions), "Should merge Netflix suggestions because they normalize to the same name and description")
}

func TestDiscoveryService_DiscoveryLogic_ConsolidateMandateWithVaryingAmounts(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockSubRepo := new(mockSubscriptionRepo)
	mockDiscoveryBankStmtRepo := new(mockDiscoveryBankStmtRepo)
	mockLLM := new(mockEnricher)
	mockSettingsRepo := new(mockSettingsRepoForDiscovery)
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
	mockDiscoveryBankStmtRepo.On("FindTransactions", ctx, mock.Anything).Return(history, nil)

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

	mockSubRepo := new(mockSubscriptionRepo)
	mockDiscoveryBankStmtRepo := new(mockDiscoveryBankStmtRepo)
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
