package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	apphttp "cogni-cash/internal/adapter/http"
	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	portmock "cogni-cash/internal/domain/port/mock"
	"cogni-cash/internal/domain/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// --- Test Router Helper ---
func NewTestRouter(authSvc port.AuthUseCase, userSvc port.UserUseCase, invoiceSvc port.InvoiceUseCase, bankStatementSvc port.BankStatementUseCase, settingsSvc port.SettingsUseCase, bankSvc port.BankUseCase, documentSvc port.DocumentUseCase, logger *slog.Logger, storageMode string, dbHost string, dbPinger func(context.Context) error,
	appCtx context.Context,
	wg *sync.WaitGroup,
) *apphttp.Router {
	// User service is mandatory for some middlewares (admin check)
	if userSvc == nil {
		userSvc = &portmock.MockUserUseCase{}
	}

	r := apphttp.NewRouter(logger, appCtx, wg, authSvc, userSvc, nil)
	r.Auth = apphttp.NewAuthHandler(logger, authSvc, nil, userSvc)
	r.System = apphttp.NewSystemHandler(&slog.LevelVar{}, logger, dbHost, dbPinger, nil, settingsSvc, storageMode, userSvc)
	r.User = apphttp.NewUserHandler(appCtx, logger, wg, nil, userSvc)

	if invoiceSvc != nil {
		r.Invoice = apphttp.NewInvoiceHandler(logger, invoiceSvc)
	}
	if bankStatementSvc != nil {
		r.BankStatement = apphttp.NewBankStatementHandler(logger, bankStatementSvc, nil, nil, settingsSvc, nil)
	}
	if bankSvc != nil {
		r.Bank = apphttp.NewBankHandler(appCtx, logger, wg, bankSvc)
	}
	if documentSvc != nil {
		r.Document = apphttp.NewDocumentHandler(logger, documentSvc)
	}
	return r
}

// setupLogger provides a no-op logger to avoid cluttering test output
func setupLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// --- Mock User Repository ---
type mockUserRepo struct {
	user entity.User
}

func (m *mockUserRepo) FindByUsername(_ context.Context, username string) (entity.User, error) {
	if username == m.user.Username {
		return m.user, nil
	}
	return entity.User{}, errors.New("user not found")
}

func (m *mockUserRepo) FindByID(_ context.Context, id uuid.UUID) (entity.User, error) {
	if id == m.user.ID {
		return m.user, nil
	}
	return entity.User{}, errors.New("user not found")
}

func (m *mockUserRepo) GetAdminID(_ context.Context) (uuid.UUID, error) {
	if m.user.Role == "admin" || m.user.Username == "admin" {
		return m.user.ID, nil
	}
	return uuid.Nil, errors.New("admin not found")
}

func (m *mockUserRepo) FindAll(_ context.Context, _ string) ([]entity.User, error) {
	return []entity.User{m.user}, nil
}

func (m *mockUserRepo) Create(_ context.Context, user entity.User) error {
	m.user = user
	return nil
}

func (m *mockUserRepo) Update(_ context.Context, user entity.User) error {
	m.user = user
	return nil
}

func (m *mockUserRepo) Upsert(_ context.Context, user entity.User) error {
	m.user = user
	return nil
}

func (m *mockUserRepo) UpdatePassword(_ context.Context, userID uuid.UUID, newHash string) error {
	if userID == m.user.ID {
		m.user.PasswordHash = newHash
		return nil
	}
	return errors.New("user not found")
}

func (m *mockUserRepo) Delete(_ context.Context, id uuid.UUID) error {
	if id == m.user.ID {
		return nil
	}
	return errors.New("user not found")
}

type mockAuthRepo struct {
	tokens map[string]entity.RefreshToken
}

func (m *mockAuthRepo) SaveRefreshToken(_ context.Context, t entity.RefreshToken) error { return nil }
func (m *mockAuthRepo) FindRefreshToken(_ context.Context, hash string) (entity.RefreshToken, error) {
	return entity.RefreshToken{}, errors.New("not found")
}
func (m *mockAuthRepo) RevokeRefreshToken(_ context.Context, id uuid.UUID, userID uuid.UUID) error { return nil }
func (m *mockAuthRepo) RevokeAllRefreshTokens(_ context.Context, userID uuid.UUID) error {
	return nil
}
func (m *mockAuthRepo) CleanupExpiredRefreshTokens(_ context.Context) error { return nil }

// setupTestAuth creates a real AuthService with a mock repo and returns a valid JWT token
func setupTestAuth(t *testing.T) (*service.AuthService, string, uuid.UUID) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	userID := uuid.New()
	user := entity.User{
		ID:           userID,
		Username:     "testadmin",
		PasswordHash: string(hash),
		Role:         "admin",
	}

	repo := &mockUserRepo{user: user}
	authSvc := service.NewAuthService(repo, nil, &mockAuthRepo{}, nil, nil, "test-secret", setupLogger())

	authResp, err := authSvc.Login(context.Background(), "testadmin", "password")
	if err != nil {
		t.Fatalf("failed to login and get token: %v", err)
	}

	return authSvc, authResp.Token, userID
}

// --- Mock Category Repository ---
type mockCategoryRepo struct {
	categories []entity.Category
}

func (m *mockCategoryRepo) Save(_ context.Context, cat entity.Category) (entity.Category, error) {
	cat.ID = uuid.New()
	m.categories = append(m.categories, cat)
	return cat, nil
}

func (m *mockCategoryRepo) Update(_ context.Context, cat entity.Category) (entity.Category, error) {
	return cat, nil
}

func (m *mockCategoryRepo) FindByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) (entity.Category, error) {
	return entity.Category{}, nil
}

func (m *mockCategoryRepo) FindAll(_ context.Context, _ uuid.UUID) ([]entity.Category, error) {
	return m.categories, nil
}

func (m *mockCategoryRepo) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

// --- Mock Bank Statement Repository (For Analytics Tests) ---
type mockBankStmtRepo struct {
	txns []entity.Transaction
}

func (m *mockBankStmtRepo) FindTransactions(_ context.Context, _ entity.TransactionFilter) ([]entity.Transaction, error) {
	return m.txns, nil
}
func (m *mockBankStmtRepo) Save(_ context.Context, stmt entity.BankStatement) error {
	return nil
}
func (m *mockBankStmtRepo) FindByID(_ context.Context, id uuid.UUID, _ uuid.UUID) (entity.BankStatement, error) {
	return entity.BankStatement{ID: id}, nil
}
func (m *mockBankStmtRepo) FindAll(_ context.Context, _ uuid.UUID) ([]entity.BankStatement, error) {
	return nil, nil
}
func (m *mockBankStmtRepo) FindSummaries(_ context.Context, _ uuid.UUID) ([]entity.BankStatementSummary, error) {
	return nil, nil
}
func (m *mockBankStmtRepo) SearchTransactions(_ context.Context, f entity.TransactionFilter) ([]entity.Transaction, error) {
	return nil, nil
}
func (m *mockBankStmtRepo) GetCategorizationExamples(_ context.Context, _ uuid.UUID, _ int) ([]entity.CategorizationExample, error) {
	return nil, nil
}
func (m *mockBankStmtRepo) FindMatchingCategory(_ context.Context, _ uuid.UUID, _ port.TransactionToCategorize) (*uuid.UUID, error) {
	return nil, nil
}
func (m *mockBankStmtRepo) UpdateTransactionCategory(_ context.Context, hash string, categoryID *uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockBankStmtRepo) UpdateTransactionSubscription(_ context.Context, contentHash string, subscriptionID *uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockBankStmtRepo) MarkTransactionReconciled(_ context.Context, _ string, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockBankStmtRepo) MarkTransactionReviewed(_ context.Context, _ string, _ uuid.UUID) error {
	return nil
}

func (m *mockBankStmtRepo) MarkTransactionsReviewedBulk(_ context.Context, _ []string, _ uuid.UUID) error {
	return nil
}

func (m *mockBankStmtRepo) UpdateTransactionBaseAmount(_ context.Context, _ string, _ float64, _ string, _ uuid.UUID) error {
	return nil
}

func (m *mockBankStmtRepo) LinkTransactionToStatement(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockBankStmtRepo) UpdateStatementAccount(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockBankStmtRepo) GetTransactionsByAccountID(_ context.Context, _ uuid.UUID, _ uuid.UUID) ([]entity.Transaction, error) {
	return nil, nil
}

func (m *mockBankStmtRepo) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockBankStmtRepo) CreateTransactions(_ context.Context, txns []entity.Transaction) error {
	m.txns = append(m.txns, txns...)
	return nil
}

// setupTestUser creates a real AuthService with a mock repo for a specific user role
func setupTestUser(t *testing.T, username, role string) (*service.AuthService, string) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	user := entity.User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
	}
	repo := &mockUserRepo{user: user}
	authSvc := service.NewAuthService(repo, nil, &mockAuthRepo{}, nil, nil, "test-secret", setupLogger())
	// Login with the user we just created
	authResp, err := authSvc.Login(context.Background(), username, "password")
	if err != nil {
		t.Fatalf("setupTestUser login failed: %v", err)
	}
	return authSvc, authResp.Token
}

// --- ACTUAL TESTS ---

func TestSettingsAccessControl(t *testing.T) {
	dummyPinger := func(ctx context.Context) error { return nil }

	runReq := func(role string) int {
		mockAuth := &portmock.MockAuthUseCase{}
		userID := uuid.New()
		token := "valid-token"
		mockAuth.On("ValidateToken", token).Return(userID.String(), nil)

		userRepo := &mockUserRepo{user: entity.User{ID: userID, Username: "testuser", Role: role}}
		userSvc := service.NewUserService(userRepo, setupLogger())

		handler := NewTestRouter(mockAuth, userSvc, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil)
		handler.System = apphttp.NewSystemHandler(&slog.LevelVar{}, setupLogger(), "localhost", dummyPinger, nil, &realMockSettingsSvc{}, "memory", userSvc)

		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/settings/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		return rr.Code
	}

	adminCode := runReq("admin")
	if adminCode != http.StatusOK {
		t.Errorf("expected admin to get 200 OK, got %d", adminCode)
	}

	managerCode := runReq("manager")
	if managerCode != http.StatusOK {
		t.Errorf("expected manager to also get 200 OK (masked), got %d", managerCode)
	}
}

func TestGetMe(t *testing.T) {
	authSvc, token, userID := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	// Setup mock user service
	userRepo := &mockUserRepo{user: entity.User{ID: userID, Username: "testuser", Role: "manager"}}
	userSvc := service.NewUserService(userRepo, setupLogger())

	handler := NewTestRouter(authSvc, userSvc, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v", rr.Code)
	}

	var res entity.User
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}

	if res.Username != "testuser" {
		t.Errorf("expected username testuser, got %v", res.Username)
	}
}

func TestHealthCheck(t *testing.T) {
	dummyPinger := func(ctx context.Context) error { return nil }
	handler := NewTestRouter(nil, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"status":"ok"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestChangePassword(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }
	handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	payload := `{"old_password":"password", "new_password":"newpassword123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password/", strings.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v - %s", status, http.StatusNoContent, rr.Body.String())
	}
}

func TestGetTransactionAnalytics(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	mockSvc := &portmock.MockTransactionUseCase{}
	mockSvc.On("GetTransactionAnalytics", mock.Anything, mock.Anything).Return(entity.TransactionAnalytics{}, nil)

	handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).
		WithTransactionService(mockSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/analytics/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestListCategories_Empty(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }
	catSvc := service.NewCategoryService(&mockCategoryRepo{}, nil, setupLogger())
	handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).WithCategoryService(catSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v", rr.Code)
	}

	var res []entity.Category
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}

	if len(res) != 0 {
		t.Errorf("expected 0 categories, got %d", len(res))
	}
}

func TestCreateCategory(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }
	catSvc := service.NewCategoryService(&mockCategoryRepo{}, nil, setupLogger())
	handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).WithCategoryService(catSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	payload := `{"name":"Food", "color":"#ff0000"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories/", strings.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %v", rr.Code)
	}
}

func TestDeleteBankStatement(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }
	svc := &portmock.MockBankStatementUseCase{}
	svc.On("DeleteStatement", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	handler := NewTestRouter(authSvc, nil, nil, svc, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).
		WithBankStatementRepository(&mockBankStmtRepo{})

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	stmtID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/bank-statements/"+stmtID.String()+"/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content, got %v", rr.Code)
	}
}

func TestUpdatePayslip_WithBonuses(t *testing.T) {
	authSvc, token, userID := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	payslipID := uuid.New().String()
	mockSvc := &portmock.MockPayslipUseCase{}
	mockSvc.On("Update", mock.Anything, mock.MatchedBy(func(p *entity.Payslip) bool {
		return p.ID == payslipID && len(p.Bonuses) == 1 && p.Bonuses[0].Amount == 500.0
	})).Return(nil)

	handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).
		WithPayslipService(mockSvc).
		WithPayslipRepository(&mockPayslipRepo{payslips: []entity.Payslip{{ID: payslipID, UserID: userID}}})

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	payload := fmt.Sprintf(`{
		"id": "%s",
		"user_id": "%s",
		"bonuses": [{"description": "Performance", "amount": 500.0}]
	}`, payslipID, userID)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/payslips/"+payslipID+"/", strings.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v - %s", rr.Code, rr.Body.String())
	}
}

func TestGetPayslip_IncludesBonuses(t *testing.T) {
	authSvc, token, userID := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	payslipID := uuid.New().String()
	mockSvc := &portmock.MockPayslipUseCase{}
	mockSvc.On("GetByID", mock.Anything, payslipID, userID).Return(entity.Payslip{
		ID:      payslipID,
		UserID:  userID,
		Bonuses: []entity.Bonus{{Description: "Yearly", Amount: 1000.0}},
	}, nil)

	handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).
		WithPayslipService(mockSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/payslips/"+payslipID+"/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v", rr.Code)
	}

	var res entity.Payslip
	json.NewDecoder(rr.Body).Decode(&res)
	if len(res.Bonuses) != 1 || res.Bonuses[0].Amount != 1000.0 {
		t.Errorf("expected bonus amount 1000.0, got %v", res.Bonuses)
	}
}

func TestImportPayslipsBatch(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	mockSvc := &portmock.MockPayslipUseCase{}
	mockSvc.On("Import", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&entity.Payslip{ID: uuid.New().String()}, nil)

	handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).
		WithPayslipService(mockSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("files", "test1.pdf")
	part.Write([]byte("fake pdf content"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/payslips/import/batch/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v", rr.Code)
	}
}

func TestGetSharingDashboard_Success(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	mockSvc := &mockSharingSvc{}
	handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).
		WithSharingService(mockSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sharing/dashboard/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v", rr.Code)
	}
}

func TestGetSharingDashboard_ServiceUnavailable(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }
	handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sharing/dashboard/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 Service Unavailable, got %v", rr.Code)
	}
}

func TestGetSharingDashboard_InternalServerError(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	mockSvc := &mockSharingSvc{err: errors.New("database error")}
	handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).
		WithSharingService(mockSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sharing/dashboard/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 Internal Server Error, got %v", rr.Code)
	}
}

func TestGetReconciliationSuggestions(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)

	mockSvc := &portmock.MockReconciliationUseCase{}
	mockSvc.On("SuggestReconciliations", mock.Anything, mock.Anything, 14).Return([]entity.ReconciliationPairSuggestion{
		{TargetTransaction: entity.Transaction{CounterpartyName: "abc"}, SourceTransaction: entity.Transaction{CounterpartyName: "def"}, MatchScore: 0.95},
	}, nil)

	dummyPinger := func(ctx context.Context) error { return nil }
	handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).
		WithReconciliationService(mockSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reconciliations/suggestions/?window=14", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v", rr.Code)
	}

	var res []entity.ReconciliationPairSuggestion
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}

	if len(res) != 1 {
		t.Errorf("expected 1 suggestion, got %d", len(res))
	}
	if res[0].MatchScore != 0.95 {
		t.Errorf("expected confidence 0.95, got %f", res[0].MatchScore)
	}
}

func TestReconciliationRouteAlias(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)

	mockSvc := &portmock.MockReconciliationUseCase{}
	mockSvc.On("SuggestReconciliations", mock.Anything, mock.Anything, 14).Return([]entity.ReconciliationPairSuggestion{
		{TargetTransaction: entity.Transaction{CounterpartyName: "abc"}, SourceTransaction: entity.Transaction{CounterpartyName: "def"}, MatchScore: 0.95},
	}, nil)

	dummyPinger := func(ctx context.Context) error { return nil }
	handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).
		WithReconciliationService(mockSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	// Test singular alias and window_days parameter
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reconciliation/suggestions/?window_days=14", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v", rr.Code)
	}
}

func TestCreatePlannedTransaction(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	mockSvc := &mockPlannedTxSvc{}
	handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).
		WithPlannedTransactionService(mockSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	payload := `{"description":"Rent", "amount":-1000.0, "category_id":"00000000-0000-0000-0000-000000000000", "booking_day":1, "occurrence":"monthly"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/planned-transactions/", strings.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %v", rr.Code)
	}
}

func TestGetSystemInfo(t *testing.T) {
	authSvc, token, userID := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	mockUserSvc := &portmock.MockUserUseCase{}
	mockUserSvc.On("GetUser", mock.Anything, userID.String()).Return(entity.User{ID: userID, Role: "admin"}, nil)

	handler := NewTestRouter(authSvc, mockUserSvc, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/info/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v", rr.Code)
	}

	var res map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&res)
	if _, ok := res["version"]; !ok {
		t.Errorf("expected version in response")
	}
}

func TestGetForecast(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	mockSvc := &mockForecastingSvc{}
	handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).
		WithForecastingService(mockSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/forecast/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v", rr.Code)
	}
}

func TestBankHandler(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	t.Run("ListInstitutions_Success", func(t *testing.T) {
		mockBankSvc := &portmock.MockBankUseCase{}
		mockBankSvc.On("GetInstitutions", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entity.BankInstitution{{ID: "ing", Name: "ING"}}, nil)

		handler := NewTestRouter(authSvc, nil, nil, nil, nil, mockBankSvc, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil)
		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/bank/institutions/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %v", rr.Code)
		}
	})

	t.Run("SyncAllAccounts_Success", func(t *testing.T) {
		mockBankSvc := &portmock.MockBankUseCase{}
		mockBankSvc.On("SyncAllAccounts", mock.Anything, mock.Anything).Return(nil)

		handler := NewTestRouter(authSvc, nil, nil, nil, nil, mockBankSvc, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil)
		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/bank/sync/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusAccepted {
			t.Fatalf("expected 202 Accepted, got %v", rr.Code)
		}
	})
}

func TestAuthHandler(t *testing.T) {
	dummyPinger := func(ctx context.Context) error { return nil }
	
	t.Run("Login_Success", func(t *testing.T) {
		mockAuth := &portmock.MockAuthUseCase{}
		mockAuth.On("Login", mock.Anything, "user", "pass").Return(entity.AuthResponse{Token: "jwt-token"}, nil)
		
		handler := NewTestRouter(mockAuth, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil)
		r := chi.NewRouter()
		handler.RegisterRoutes(r)
		
		payload := []byte(`{"username":"user", "password":"pass"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/login/", bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()
		
		r.ServeHTTP(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %v", rr.Code)
		}
	})
}

func TestUserCRUD(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	t.Run("CreateUser_Success", func(t *testing.T) {
		mockUserSvc := &portmock.MockUserUseCase{}
		mockUserSvc.On("CreateUser", mock.Anything, mock.Anything, "password123").Return(entity.User{ID: uuid.New(), Username: "newuser"}, nil)
		// Middleware needs a user to check admin role
		mockUserSvc.On("GetUser", mock.Anything, mock.Anything).Return(entity.User{Role: "admin"}, nil)

		handler := NewTestRouter(authSvc, mockUserSvc, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil)
		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		payload := []byte(`{"username":"newuser", "password":"password123", "role":"manager"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users/", bytes.NewBuffer(payload))
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %v", rr.Code)
		}
	})
}

func TestSettingsUpdate(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	t.Run("UpdateSettings_Success", func(t *testing.T) {
		mockSettingsSvc := &portmock.MockSettingsUseCase{}
		mockSettingsSvc.On("UpdateMultiple", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockUserSvc := &portmock.MockUserUseCase{}
		mockUserSvc.On("GetUser", mock.Anything, mock.Anything).Return(entity.User{Role: "admin"}, nil)

		handler := NewTestRouter(authSvc, mockUserSvc, nil, nil, mockSettingsSvc, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil)
		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		payload := []byte(`{"bank_provider":"mock"}`)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/settings/", bytes.NewBuffer(payload))
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Fatalf("expected 204 No Content, got %v", rr.Code)
		}
	})
}

func TestDiscoveryHandler(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	t.Run("GetSuggestedSubscriptions_Success", func(t *testing.T) {
		mockDiscoverySvc := &portmock.MockDiscoveryUseCase{}
		mockDiscoverySvc.On("GetSuggestedSubscriptions", mock.Anything, mock.Anything).Return([]entity.SuggestedSubscription{{MerchantName: "Netflix"}}, nil)

		handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).
			WithDiscoveryService(mockDiscoverySvc)
		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/suggested/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %v", rr.Code)
		}
	})

	t.Run("GetDiscoveryFeedback_Success", func(t *testing.T) {
		mockDiscoverySvc := &portmock.MockDiscoveryUseCase{}
		mockDiscoverySvc.On("GetDiscoveryFeedback", mock.Anything, mock.Anything).Return([]entity.DiscoveryFeedback{{MerchantName: "Amazon", Status: "ignored"}}, nil)

		handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).
			WithDiscoveryService(mockDiscoverySvc)
		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/feedback/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %v", rr.Code)
		}
	})

	t.Run("ApproveSubscription_Success", func(t *testing.T) {
		mockDiscoverySvc := &portmock.MockDiscoveryUseCase{}
		mockDiscoverySvc.On("ApproveSubscription", mock.Anything, mock.Anything, mock.Anything).Return(entity.Subscription{ID: uuid.New()}, nil)

		handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).
			WithDiscoveryService(mockDiscoverySvc)
		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		payload := []byte(`{"merchant_name":"Netflix", "estimated_amount":15.99}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/approve/", bytes.NewBuffer(payload))
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %v", rr.Code)
		}
	})
}

func TestBankStatementHandler_Transactions(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	t.Run("ListTransactions_Success", func(t *testing.T) {
		mockSvc := &portmock.MockTransactionUseCase{}
		mockSvc.On("ListTransactions", mock.Anything, mock.Anything).Return([]entity.Transaction{{Description: "Coffee", Amount: -3.5}}, nil)

		handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).
			WithTransactionService(mockSvc)
		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %v", rr.Code)
		}
	})

	t.Run("UpdateTransactionCategory_Success", func(t *testing.T) {
		mockSvc := &portmock.MockTransactionUseCase{}
		mockSvc.On("UpdateCategory", mock.Anything, "hash123", mock.Anything, mock.Anything).Return(nil)

		handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil).
			WithTransactionService(mockSvc)
		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		catID := uuid.New().String()
		payload := []byte(fmt.Sprintf(`{"category_id":"%s"}`, catID))
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/transactions/hash123/category/", bytes.NewBuffer(payload))
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Fatalf("expected 204 No Content, got %v", rr.Code)
		}
	})
}

func TestInvoiceHandler_List(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	mockSvc := &portmock.MockInvoiceUseCase{}
	mockSvc.On("GetAll", mock.Anything, mock.Anything).Return([]entity.Invoice{{Vendor: entity.Vendor{Name: "AWS"}, Amount: 42.0}}, nil)

	handler := NewTestRouter(authSvc, nil, mockSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/invoices/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v", rr.Code)
	}
}

func TestUserHandler_Extended(t *testing.T) {
	authSvc, token, userID := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	t.Run("ListUsers_Success", func(t *testing.T) {
		mockUserSvc := &portmock.MockUserUseCase{}
		mockUserSvc.On("ListUsers", mock.Anything, mock.Anything).Return([]entity.User{{ID: userID, Username: "admin", Role: "admin"}}, nil)
		mockUserSvc.On("GetUser", mock.Anything, mock.Anything).Return(entity.User{Role: "admin"}, nil)

		handler := NewTestRouter(authSvc, mockUserSvc, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil)
		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %v", rr.Code)
		}
	})

	t.Run("DeleteUser_Success", func(t *testing.T) {
		mockUserSvc := &portmock.MockUserUseCase{}
		mockUserSvc.On("DeleteUser", mock.Anything, mock.Anything).Return(nil)
		mockUserSvc.On("GetUser", mock.Anything, mock.Anything).Return(entity.User{Role: "admin"}, nil)

		handler := NewTestRouter(authSvc, mockUserSvc, nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil)
		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+uuid.New().String()+"/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Fatalf("expected 204 No Content, got %v", rr.Code)
		}
	})
}

// --- Mocks ---

type realMockSettingsSvc struct {
	port.SettingsUseCase
}

func (m *realMockSettingsSvc) GetAll(ctx context.Context, userID uuid.UUID) (map[string]string, error) {
	return map[string]string{}, nil
}

func (m *realMockSettingsSvc) GetAllMasked(ctx context.Context, userID uuid.UUID, isAdmin bool) (map[string]string, error) {
	return map[string]string{}, nil
}

type mockSharingRepo struct{}

func (m *mockSharingRepo) GetDashboard(_ context.Context, _ uuid.UUID) (entity.SharingDashboard, error) {
	return entity.SharingDashboard{}, nil
}

type mockPayslipRepo struct {
	payslips []entity.Payslip
}

func (m *mockPayslipRepo) Save(_ context.Context, p *entity.Payslip) error {
	m.payslips = append(m.payslips, *p)
	return nil
}
func (m *mockPayslipRepo) ExistsByHash(_ context.Context, _ string, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockPayslipRepo) ExistsByOriginalFileName(_ context.Context, _ string, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockPayslipRepo) FindAll(_ context.Context, _ entity.PayslipFilter) ([]entity.Payslip, error) {
	return m.payslips, nil
}
func (m *mockPayslipRepo) FindByID(_ context.Context, id string, _ uuid.UUID) (entity.Payslip, error) {
	for _, p := range m.payslips {
		if p.ID == id {
			return p, nil
		}
	}
	return entity.Payslip{}, errors.New("not found")
}
func (m *mockPayslipRepo) Delete(_ context.Context, _ string, _ uuid.UUID) error { return nil }
func (m *mockPayslipRepo) GetSummary(_ context.Context, _ uuid.UUID) (entity.PayslipSummary, error) {
	return entity.PayslipSummary{}, nil
}
func (m *mockPayslipRepo) Update(_ context.Context, p *entity.Payslip) error { return nil }
func (m *mockPayslipRepo) UpdateBaseAmount(_ context.Context, _ string, _, _, _ float64, _ string, _ uuid.UUID) error {
	return nil
}
func (m *mockPayslipRepo) GetOriginalFile(_ context.Context, _ string, _ uuid.UUID) ([]byte, string, string, error) {
	return nil, "", "", nil
}

type mockPayslipParser struct{}

func (m *mockPayslipParser) Parse(_ []byte) (entity.Payslip, error) {
	return entity.Payslip{PeriodMonthNum: 1, PeriodYear: 2024, NetPay: 3000}, nil
}

type mockSharingSvc struct {
	err error
}

func (m *mockSharingSvc) GetSharingDashboard(ctx context.Context, userID uuid.UUID) (entity.SharingDashboard, error) {
	if m.err != nil {
		return entity.SharingDashboard{}, m.err
	}
	return entity.SharingDashboard{}, nil
}

func (m *mockSharingSvc) GetDashboard(ctx context.Context, userID uuid.UUID) (entity.SharingDashboard, error) {
	return m.GetSharingDashboard(ctx, userID)
}

func (m *mockSharingSvc) ShareInvoice(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID, permission string) error {
	return nil
}

func (m *mockSharingSvc) RevokeInvoiceShare(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID) error {
	return nil
}

func (m *mockSharingSvc) ListInvoiceShares(ctx context.Context, invoiceID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

type mockPlannedTxSvc struct {
	txs []entity.PlannedTransaction
	err error
}

func (m *mockPlannedTxSvc) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error) {
	return m.txs, m.err
}

func (m *mockPlannedTxSvc) Create(ctx context.Context, pt *entity.PlannedTransaction) error {
	pt.ID = uuid.New()
	m.txs = append(m.txs, *pt)
	return m.err
}

func (m *mockPlannedTxSvc) Update(ctx context.Context, pt *entity.PlannedTransaction) error {
	return m.err
}

func (m *mockPlannedTxSvc) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return m.err
}

func (m *mockPlannedTxSvc) MatchTransactions(ctx context.Context, userID uuid.UUID, txns []entity.Transaction) error {
	return m.err
}

func (m *mockPlannedTxSvc) MatchAndSpawn(ctx context.Context, tx entity.Transaction) error {
	return m.err
}

type mockForecastingSvc struct {
	forecast entity.CashFlowForecast
	err      error
}

func (m *mockForecastingSvc) GetCashFlowForecast(ctx context.Context, userID uuid.UUID, from, to time.Time) (entity.CashFlowForecast, error) {
	return m.forecast, m.err
}

func (m *mockForecastingSvc) CalculateCategoryAverage(ctx context.Context, userID, categoryID uuid.UUID, strategy string) (float64, error) {
	return 0, nil
}

func TestDocumentHandler(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }

	t.Run("ListDocuments_Success", func(t *testing.T) {
		mockSvc := &portmock.MockDocumentUseCase{}
		mockSvc.On("List", mock.Anything, mock.Anything).Return([]entity.Document{{OriginalFileName: "doc.pdf"}}, nil)

		handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, mockSvc, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil)
		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %v", rr.Code)
		}
	})

	t.Run("DownloadDocument_Success", func(t *testing.T) {
		mockSvc := &portmock.MockDocumentUseCase{}
		docID := uuid.New()
		mockSvc.On("Download", mock.Anything, docID, mock.Anything).Return([]byte("pdf-content"), "application/pdf", "doc.pdf", nil)

		handler := NewTestRouter(authSvc, nil, nil, nil, nil, nil, mockSvc, setupLogger(), "memory", "localhost", dummyPinger, context.Background(), nil)
		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/"+docID.String()+"/download/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %v", rr.Code)
		}
		if rr.Body.String() != "pdf-content" {
			t.Errorf("expected pdf-content, got %v", rr.Body.String())
		}
	})
}
