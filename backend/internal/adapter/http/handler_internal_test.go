package http

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDiscoveryUseCase struct {
	mock.Mock
}

func (m *mockDiscoveryUseCase) ListSubscriptions(ctx context.Context, userID uuid.UUID) ([]entity.Subscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.Subscription), args.Error(1)
}
func (m *mockDiscoveryUseCase) GetSubscription(ctx context.Context, id, userID uuid.UUID) (entity.Subscription, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(entity.Subscription), args.Error(1)
}
func (m *mockDiscoveryUseCase) UpdateSubscription(ctx context.Context, sub entity.Subscription) (entity.Subscription, error) {
	args := m.Called(ctx, sub)
	return args.Get(0).(entity.Subscription), args.Error(1)
}
func (m *mockDiscoveryUseCase) DeleteSubscription(ctx context.Context, userID, id uuid.UUID) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}
func (m *mockDiscoveryUseCase) GetSuggestedSubscriptions(ctx context.Context, userID uuid.UUID) ([]entity.SuggestedSubscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.SuggestedSubscription), args.Error(1)
}
func (m *mockDiscoveryUseCase) ApproveSubscription(ctx context.Context, userID uuid.UUID, suggestion entity.SuggestedSubscription) (entity.Subscription, error) {
	args := m.Called(ctx, userID, suggestion)
	return args.Get(0).(entity.Subscription), args.Error(1)
}
func (m *mockDiscoveryUseCase) DeclineSuggestion(ctx context.Context, userID uuid.UUID, merchantName string) error {
	return m.Called(ctx, userID, merchantName).Error(0)
}
func (m *mockDiscoveryUseCase) GetDiscoveryFeedback(ctx context.Context, userID uuid.UUID) ([]entity.DiscoveryFeedback, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.DiscoveryFeedback), args.Error(1)
}
func (m *mockDiscoveryUseCase) RemoveDiscoveryFeedback(ctx context.Context, userID uuid.UUID, merchantName string) error {
	return m.Called(ctx, userID, merchantName).Error(0)
}
func (m *mockDiscoveryUseCase) AllowSuggestion(ctx context.Context, userID uuid.UUID, merchantName string) error {
	return m.Called(ctx, userID, merchantName).Error(0)
}
func (m *mockDiscoveryUseCase) EnrichSubscription(ctx context.Context, userID, id uuid.UUID) (entity.Subscription, error) {
	args := m.Called(ctx, userID, id)
	return args.Get(0).(entity.Subscription), args.Error(1)
}
func (m *mockDiscoveryUseCase) CreateSubscriptionFromTransaction(ctx context.Context, userID uuid.UUID, transactionHash string, merchantName string, billingCycle string, billingInterval int) (entity.Subscription, error) {
	args := m.Called(ctx, userID, transactionHash, merchantName, billingCycle, billingInterval)
	return args.Get(0).(entity.Subscription), args.Error(1)
}
func (m *mockDiscoveryUseCase) PreviewCancellation(ctx context.Context, userID, id uuid.UUID, language string) (port.CancellationLetterResult, error) {
	args := m.Called(ctx, userID, id, language)
	return args.Get(0).(port.CancellationLetterResult), args.Error(1)
}
func (m *mockDiscoveryUseCase) CancelSubscription(ctx context.Context, userID, id uuid.UUID, emailBody string, contactEmail string) error {
	return m.Called(ctx, userID, id, emailBody, contactEmail).Error(0)
}
func (m *mockDiscoveryUseCase) GetSubscriptionEvents(ctx context.Context, id, userID uuid.UUID) ([]entity.SubscriptionEvent, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).([]entity.SubscriptionEvent), args.Error(1)
}
func (m *mockDiscoveryUseCase) MatchTransactions(ctx context.Context, userID uuid.UUID, txns []entity.Transaction) error {
	return m.Called(ctx, userID, txns).Error(0)
}
func (m *mockDiscoveryUseCase) LinkTransaction(ctx context.Context, userID, subID uuid.UUID, txnHash string) error {
	return m.Called(ctx, userID, subID, txnHash).Error(0)
}
func (m *mockDiscoveryUseCase) UnlinkTransaction(ctx context.Context, userID, subID uuid.UUID, txnHash string) error {
	return m.Called(ctx, userID, subID, txnHash).Error(0)
}
func (m *mockDiscoveryUseCase) LinkTransactions(ctx context.Context, userID, subID uuid.UUID, txnHashes []string) error {
	return m.Called(ctx, userID, subID, txnHashes).Error(0)
}

type mockCategoryUseCase struct {
	mock.Mock
}

func (m *mockCategoryUseCase) GetAll(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.Category), args.Error(1)
}
func (m *mockCategoryUseCase) GetByID(ctx context.Context, id, userID uuid.UUID) (entity.Category, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(entity.Category), args.Error(1)
}
func (m *mockCategoryUseCase) Create(ctx context.Context, cat entity.Category) (entity.Category, error) {
	args := m.Called(ctx, cat)
	return args.Get(0).(entity.Category), args.Error(1)
}
func (m *mockCategoryUseCase) Update(ctx context.Context, cat entity.Category) (entity.Category, error) {
	args := m.Called(ctx, cat)
	return args.Get(0).(entity.Category), args.Error(1)
}
func (m *mockCategoryUseCase) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}
func (m *mockCategoryUseCase) ShareCategory(ctx context.Context, id, ownerID, sharedWithID uuid.UUID, perm string) error {
	return m.Called(ctx, id, ownerID, sharedWithID, perm).Error(0)
}
func (m *mockCategoryUseCase) RevokeShare(ctx context.Context, id, ownerID, sharedWithID uuid.UUID) error {
	return m.Called(ctx, id, ownerID, sharedWithID).Error(0)
}
func (m *mockCategoryUseCase) ListShares(ctx context.Context, id, ownerID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, id, ownerID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

type mockBankUseCase struct {
	mock.Mock
}

func (m *mockBankUseCase) GetInstitutions(ctx context.Context, userID uuid.UUID, country string, sandbox bool) ([]entity.BankInstitution, error) {
	args := m.Called(ctx, userID, country, sandbox)
	return args.Get(0).([]entity.BankInstitution), args.Error(1)
}
func (m *mockBankUseCase) CreateConnection(ctx context.Context, userID uuid.UUID, instID, instName, country, redirect string, sandbox bool, ip, ua string) (*entity.BankConnection, error) {
	args := m.Called(ctx, userID, instID, instName, country, redirect, sandbox, ip, ua)
	return args.Get(0).(*entity.BankConnection), args.Error(1)
}
func (m *mockBankUseCase) FinishConnection(ctx context.Context, userID uuid.UUID, reqID, code string) error {
	return m.Called(ctx, userID, reqID, code).Error(0)
}
func (m *mockBankUseCase) GetConnections(ctx context.Context, userID uuid.UUID) ([]entity.BankConnection, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.BankConnection), args.Error(1)
}
func (m *mockBankUseCase) DeleteConnection(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}
func (m *mockBankUseCase) SyncAccount(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}
func (m *mockBankUseCase) SyncAllAccounts(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *mockBankUseCase) UpdateAccountType(ctx context.Context, id uuid.UUID, accType entity.StatementType, userID uuid.UUID) error {
	return m.Called(ctx, id, accType, userID).Error(0)
}
func (m *mockBankUseCase) CreateVirtualAccount(ctx context.Context, account *entity.BankAccount) error {
	return m.Called(ctx, account).Error(0)
}
func (m *mockBankUseCase) ShareAccount(ctx context.Context, id, ownerID, sharedWithID uuid.UUID, perm string) error {
	return m.Called(ctx, id, ownerID, sharedWithID, perm).Error(0)
}
func (m *mockBankUseCase) RevokeShare(ctx context.Context, id, ownerID, sharedWithID uuid.UUID) error {
	return m.Called(ctx, id, ownerID, sharedWithID).Error(0)
}
func (m *mockBankUseCase) ListShares(ctx context.Context, id, ownerID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, id, ownerID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

type mockInvoiceUseCase struct {
	mock.Mock
}

func (m *mockInvoiceUseCase) ImportFromFile(ctx context.Context, userID uuid.UUID, fileName, mimeType string, fileBytes []byte, overrides port.ImportOverrides) (entity.Invoice, error) {
	args := m.Called(ctx, userID, fileName, mimeType, fileBytes, overrides)
	return args.Get(0).(entity.Invoice), args.Error(1)
}
func (m *mockInvoiceUseCase) ImportManual(ctx context.Context, userID uuid.UUID, invoice entity.Invoice) (entity.Invoice, error) {
	args := m.Called(ctx, userID, invoice)
	return args.Get(0).(entity.Invoice), args.Error(1)
}
func (m *mockInvoiceUseCase) CategorizeDocument(ctx context.Context, userID uuid.UUID, rawText string) (entity.Invoice, error) {
	args := m.Called(ctx, userID, rawText)
	return args.Get(0).(entity.Invoice), args.Error(1)
}
func (m *mockInvoiceUseCase) GetAll(ctx context.Context, filter entity.InvoiceFilter) ([]entity.Invoice, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Invoice), args.Error(1)
}
func (m *mockInvoiceUseCase) GetByID(ctx context.Context, id, userID uuid.UUID) (entity.Invoice, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(entity.Invoice), args.Error(1)
}
func (m *mockInvoiceUseCase) Update(ctx context.Context, invoice entity.Invoice) (entity.Invoice, error) {
	args := m.Called(ctx, invoice)
	return args.Get(0).(entity.Invoice), args.Error(1)
}
func (m *mockInvoiceUseCase) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}
func (m *mockInvoiceUseCase) GetOriginalFile(ctx context.Context, id, userID uuid.UUID) ([]byte, string, string, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).([]byte), args.String(1), args.String(2), args.Error(3)
}
func (m *mockInvoiceUseCase) ShareInvoice(ctx context.Context, id, ownerID, sharedWithID uuid.UUID, perm string) error {
	return m.Called(ctx, id, ownerID, sharedWithID, perm).Error(0)
}
func (m *mockInvoiceUseCase) RevokeInvoiceShare(ctx context.Context, id, ownerID, sharedWithID uuid.UUID) error {
	return m.Called(ctx, id, ownerID, sharedWithID).Error(0)
}
func (m *mockInvoiceUseCase) ListInvoiceShares(ctx context.Context, id, ownerID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, id, ownerID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func TestDiscoveryHandler(t *testing.T) {
	userID := uuid.New()
	mockSvc := new(mockDiscoveryUseCase)
	handler := NewHandler(nil, nil, nil, nil, nil, slog.Default(), "memory", "localhost", nil, context.Background(), nil).
		WithDiscoveryService(mockSvc)

	t.Run("ListSubscriptions - Success", func(t *testing.T) {
		subs := []entity.Subscription{{ID: uuid.New(), MerchantName: "Netflix"}}
		mockSvc.On("ListSubscriptions", mock.Anything, userID).Return(subs, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/subscriptions/", nil)
		ctx := req.Context()
		ctx = context.WithValue(ctx, userIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.ListSubscriptions(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp []entity.Subscription
		json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Len(t, resp, 1)
		assert.Equal(t, "Netflix", resp[0].MerchantName)
	})

	t.Run("ApproveSubscription - Success", func(t *testing.T) {
		suggestion := entity.SuggestedSubscription{MerchantName: "Spotify"}
		sub := entity.Subscription{ID: uuid.New(), MerchantName: "Spotify"}
		mockSvc.On("ApproveSubscription", mock.Anything, userID, suggestion).Return(sub, nil).Once()

		body, _ := json.Marshal(ApproveRequest{Suggestion: suggestion})
		req := httptest.NewRequest("POST", "/api/v1/subscriptions/approve/", bytes.NewBuffer(body))
		ctx := req.Context()
		ctx = context.WithValue(ctx, userIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.ApproveSubscription(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		var resp entity.Subscription
		json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal(t, "Spotify", resp.MerchantName)
	})
}

func TestCategoryHandler(t *testing.T) {
	userID := uuid.New()
	mockSvc := new(mockCategoryUseCase)
	handler := NewHandler(nil, nil, nil, nil, nil, slog.Default(), "memory", "localhost", nil, context.Background(), nil).
		WithCategoryService(mockSvc)

	t.Run("listCategories - Success", func(t *testing.T) {
		cats := []entity.Category{{ID: uuid.New(), Name: "Food"}}
		mockSvc.On("GetAll", mock.Anything, userID).Return(cats, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/categories/", nil)
		ctx := req.Context()
		ctx = context.WithValue(ctx, userIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.listCategories(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp []entity.Category
		json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Len(t, resp, 1)
	})

	t.Run("createCategory - Success", func(t *testing.T) {
		cat := entity.Category{Name: "Travel", Color: "#blue"}
		mockSvc.On("Create", mock.Anything, mock.MatchedBy(func(c entity.Category) bool {
			return c.Name == "Travel" && c.UserID == userID
		})).Return(cat, nil).Once()

		body, _ := json.Marshal(cat)
		req := httptest.NewRequest("POST", "/api/v1/categories/", bytes.NewBuffer(body))
		ctx := req.Context()
		ctx = context.WithValue(ctx, userIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.createCategory(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})
}

func TestBankHandler(t *testing.T) {
	userID := uuid.New()
	mockSvc := new(mockBankUseCase)
	handler := NewHandler(nil, nil, nil, nil, nil, slog.Default(), "memory", "localhost", nil, context.Background(), nil).
		WithBankService(mockSvc)

	t.Run("listBankConnections - Success", func(t *testing.T) {
		conns := []entity.BankConnection{{ID: uuid.New(), InstitutionName: "Sparkasse"}}
		mockSvc.On("GetConnections", mock.Anything, userID).Return(conns, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/bank/connections/", nil)
		ctx := req.Context()
		ctx = context.WithValue(ctx, userIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.listBankConnections(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("syncAllBankAccounts - Success", func(t *testing.T) {
		mockSvc.On("SyncAllAccounts", mock.Anything, userID).Return(nil).Once()

		req := httptest.NewRequest("POST", "/api/v1/bank/sync/", nil)
		ctx := req.Context()
		ctx = context.WithValue(ctx, userIDKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.syncAllBankAccounts(rr, req)

		assert.Equal(t, http.StatusAccepted, rr.Code)
	})
}

func TestInvoiceHandler(t *testing.T) {
	userID := uuid.New()
	invoiceID := uuid.New()
	mockSvc := new(mockInvoiceUseCase)
	handler := NewHandler(nil, nil, nil, nil, nil, slog.Default(), "memory", "localhost", nil, context.Background(), nil).
		WithInvoiceService(mockSvc)

	t.Run("updateInvoice - Success with ISO Date", func(t *testing.T) {
		issuedAt := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
		mockSvc.On("GetByID", mock.Anything, invoiceID, userID).Return(entity.Invoice{ID: invoiceID, UserID: userID}, nil).Once()
		mockSvc.On("Update", mock.Anything, mock.MatchedBy(func(inv entity.Invoice) bool {
			return inv.ID == invoiceID && inv.IssuedAt.Equal(issuedAt)
		})).Return(entity.Invoice{ID: invoiceID, IssuedAt: issuedAt}, nil).Once()

		body := map[string]interface{}{
			"vendor_name": "Test Vendor",
			"issued_at":   "2026-04-20T00:00:00Z",
			"amount":      100.50,
			"currency":    "EUR",
		}
		jsonBody, _ := json.Marshal(body)
		
		req := httptest.NewRequest("PUT", "/api/v1/invoices/"+invoiceID.String()+"/", bytes.NewBuffer(jsonBody))
		ctx := req.Context()
		ctx = context.WithValue(ctx, userIDKey, userID)
		// Mock chi URL param
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", invoiceID.String())
		ctx = context.WithValue(ctx, chi.RouteCtxKey, chiCtx)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.updateInvoice(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("updateInvoice - Failure with YYYY-MM-DD Date", func(t *testing.T) {
		body := map[string]interface{}{
			"issued_at": "2026-04-20",
		}
		jsonBody, _ := json.Marshal(body)
		
		req := httptest.NewRequest("PUT", "/api/v1/invoices/"+invoiceID.String()+"/", bytes.NewBuffer(jsonBody))
		ctx := req.Context()
		ctx = context.WithValue(ctx, userIDKey, userID)
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", invoiceID.String())
		ctx = context.WithValue(ctx, chi.RouteCtxKey, chiCtx)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.updateInvoice(rr, req)

		// This confirms that the backend currently FAILS with plain dates,
		// which is why we fixed the frontend to send ISO strings.
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid request body")
	})

	t.Run("updateInvoice - Clear Splits", func(t *testing.T) {
		mockSvc.On("GetByID", mock.Anything, invoiceID, userID).Return(entity.Invoice{
			ID:     invoiceID,
			UserID: userID,
			Splits: []entity.InvoiceSplit{{ID: uuid.New()}}, // Has splits
		}, nil).Once()

		mockSvc.On("Update", mock.Anything, mock.MatchedBy(func(inv entity.Invoice) bool {
			return inv.ID == invoiceID && len(inv.Splits) == 0
		})).Return(entity.Invoice{ID: invoiceID, Splits: []entity.InvoiceSplit{}}, nil).Once()

		body := map[string]interface{}{
			"splits": []interface{}{}, // Empty array to clear
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PUT", "/api/v1/invoices/"+invoiceID.String()+"/", bytes.NewBuffer(jsonBody))
		ctx := req.Context()
		ctx = context.WithValue(ctx, userIDKey, userID)
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", invoiceID.String())
		ctx = context.WithValue(ctx, chi.RouteCtxKey, chiCtx)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.updateInvoice(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockSvc.AssertExpectations(t)
	})
}
