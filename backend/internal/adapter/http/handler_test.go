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
	"testing"
	"time"

	apphttp "cogni-cash/internal/adapter/http"
	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"cogni-cash/internal/domain/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

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
func (m *mockAuthRepo) RevokeRefreshToken(_ context.Context, id uuid.UUID) error { return nil }
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
	authSvc := service.NewAuthService(repo, nil, &mockAuthRepo{}, nil, nil, "test-secret-key", setupLogger())

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

func (m *mockBankStmtRepo) UpdateTransactionSkipForecasting(_ context.Context, _ string, _ bool, _ uuid.UUID) error {
	return nil
}

func (m *mockBankStmtRepo) LinkTransactionToStatement(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ uuid.UUID) error {
	return nil
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
	authResp, _ := authSvc.Login(context.Background(), "admin", "password")
	token := authResp.Token
	return authSvc, token
}

func TestSettingsAccessControl(t *testing.T) {
	adminAuth, adminToken := setupTestUser(t, "admin", "admin")
	managerAuth, managerToken := setupTestUser(t, "manager", "manager")

	dummyPinger := func(ctx context.Context) error { return nil }

	runReq := func(authSvc *service.AuthService, token string) int {
		handler := apphttp.NewHandler(authSvc, nil, nil, &realMockSettingsSvc{}, nil, setupLogger(), "memory", "localhost", dummyPinger)
		userSvc := service.NewUserService(authSvc.GetRepo_ForTest().(port.UserRepository), setupLogger())
		handler.WithUserService(userSvc)

		r := chi.NewRouter()
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/settings/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		return rr.Code
	}

	adminCode := runReq(adminAuth, adminToken)
	if adminCode == http.StatusUnauthorized || adminCode == http.StatusForbidden {
		t.Errorf("expected admin to bypass middleware, got %d", adminCode)
	}

	managerCode := runReq(managerAuth, managerToken)
	if managerCode != http.StatusUnauthorized && managerCode != http.StatusForbidden {
		t.Errorf("expected manager to be blocked, got %d", managerCode)
	}
}

type realMockSettingsSvc struct {
	port.SettingsUseCase
}

func (m *realMockSettingsSvc) GetAll(ctx context.Context, userID uuid.UUID) (map[string]string, error) {
	return map[string]string{}, nil
}

// HIER IST DIE KORREKTUR: isAdmin Parameter hinzugefügt
func (m *realMockSettingsSvc) GetAllMasked(ctx context.Context, userID uuid.UUID, isAdmin bool) (map[string]string, error) {
	return map[string]string{}, nil
}

func TestHealthCheck(t *testing.T) {
	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(nil, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `{"status":"ok"}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestChangePassword(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	payload := []byte(`{"old_password":"password", "new_password":"newpassword123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password/", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("expected 204 No Content for successful password change, got %v", status)
	}

	badPayload := []byte(`{"old_password":"wrongpassword", "new_password":"newpassword123"}`)
	badReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password/", bytes.NewBuffer(badPayload))
	badReq.Header.Set("Authorization", "Bearer "+token)
	badRr := httptest.NewRecorder()

	r.ServeHTTP(badRr, badReq)

	if status := badRr.Code; status != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for invalid old password, got %v", status)
	}
}

func TestGetTransactionAnalytics(t *testing.T) {
	authSvc, token, userID := setupTestAuth(t)

	cat1 := uuid.New()
	cat2 := uuid.New()

	repo := &mockBankStmtRepo{
		txns: []entity.Transaction{
			{UserID: userID, BookingDate: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), Amount: -50, CategoryID: &cat1, Description: "Supermarket"},
			{UserID: userID, BookingDate: time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC), Amount: 100, CategoryID: &cat2, Description: "Employer"},
		},
	}

	txSvc := service.NewTransactionService(repo, nil, nil, nil, setupLogger())

	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).
		WithTransactionService(txSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/analytics/?from=2026-01-01&to=2026-01-31", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v", status)
	}

	var res entity.TransactionAnalytics
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}

	if res.TotalExpense != 50 {
		t.Errorf("expected 50 expense, got %f", res.TotalExpense)
	}
	if res.TotalIncome != 100 {
		t.Errorf("expected 100 income, got %f", res.TotalIncome)
	}
	if len(res.TopMerchants) != 1 {
		t.Fatalf("expected 1 top merchant, got %d", len(res.TopMerchants))
	}
	if res.TopMerchants[0].Merchant != "Supermarket" {
		t.Errorf("expected Supermarket, got %s", res.TopMerchants[0].Merchant)
	}
}

type mockSharingRepo struct{}

func (m *mockSharingRepo) ShareCategory(_ context.Context, _, _, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockSharingRepo) RevokeShare(_ context.Context, _, _, _ uuid.UUID) error { return nil }
func (m *mockSharingRepo) ListShares(_ context.Context, _, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (m *mockSharingRepo) ShareInvoice(_ context.Context, _, _, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockSharingRepo) RevokeInvoiceShare(_ context.Context, _, _, _ uuid.UUID) error { return nil }
func (m *mockSharingRepo) ListInvoiceShares(_ context.Context, _, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func TestListCategories_Empty(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	mockRepo := &mockCategoryRepo{}
	mockSharing := &mockSharingRepo{}
	catSvc := service.NewCategoryService(mockRepo, mockSharing, setupLogger())

	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).WithCategoryService(catSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected 200 OK, got %v", status)
	}

	var response []entity.Category
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if len(response) != 0 {
		t.Errorf("expected 0 categories, got %d", len(response))
	}
}

func TestCreateCategory(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	mockRepo := &mockCategoryRepo{}
	mockSharing := &mockSharingRepo{}
	catSvc := service.NewCategoryService(mockRepo, mockSharing, setupLogger())

	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).WithCategoryService(catSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	payload := []byte(`{"name":"Groceries", "color":"#123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories/", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("expected 201 Created, got %v", status)
	}

	if len(mockRepo.categories) != 1 {
		t.Fatalf("expected 1 category saved to mock repo, got %d", len(mockRepo.categories))
	}
	if mockRepo.categories[0].Name != "Groceries" {
		t.Errorf("expected name Groceries, got %s", mockRepo.categories[0].Name)
	}
}

func TestDeleteBankStatement(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	repo := &mockBankStmtRepo{}

	svc := service.NewBankStatementService(repo, setupLogger())

	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(authSvc, nil, svc, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).
		WithBankStatementRepository(repo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	statementID := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/bank-statements/"+statementID+"/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %v", status)
	}
}

// --- Mock Payslip Repository ---

type mockPayslipRepo struct {
	payslips map[string]entity.Payslip
}

func newMockPayslipRepo() *mockPayslipRepo {
	return &mockPayslipRepo{payslips: make(map[string]entity.Payslip)}
}

func (m *mockPayslipRepo) Save(_ context.Context, p *entity.Payslip) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	m.payslips[p.ID] = *p
	return nil
}
func (m *mockPayslipRepo) ExistsByHash(_ context.Context, _ string, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockPayslipRepo) ExistsByOriginalFileName(_ context.Context, _ string, _ uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockPayslipRepo) FindAll(_ context.Context, filter entity.PayslipFilter) ([]entity.Payslip, error) {
	result := make([]entity.Payslip, 0, len(m.payslips))
	for _, p := range m.payslips {
		if filter.Employer != "" && p.EmployerName != filter.Employer {
			continue
		}
		result = append(result, p)
	}
	return result, nil
}
func (m *mockPayslipRepo) FindByID(_ context.Context, id string, _ uuid.UUID) (entity.Payslip, error) {
	p, ok := m.payslips[id]
	if !ok {
		return entity.Payslip{}, errors.New("payslip not found")
	}
	return p, nil
}
func (m *mockPayslipRepo) GetAll(ctx context.Context, filter entity.PayslipFilter) ([]entity.Payslip, error) {
	return m.FindAll(ctx, filter)
}
func (m *mockPayslipRepo) GetByID(ctx context.Context, id string, userID uuid.UUID) (entity.Payslip, error) {
	return m.FindByID(ctx, id, userID)
}
func (m *mockPayslipRepo) Update(_ context.Context, p *entity.Payslip) error {
	if _, ok := m.payslips[p.ID]; !ok {
		return errors.New("payslip not found")
	}
	m.payslips[p.ID] = *p
	return nil
}
func (m *mockPayslipRepo) Delete(_ context.Context, id string, _ uuid.UUID) error {
	delete(m.payslips, id)
	return nil
}
func (m *mockPayslipRepo) Import(ctx context.Context, userID uuid.UUID, fileName, mimeType string, fileBytes []byte, overrides *entity.Payslip, useAI bool) (*entity.Payslip, error) {
	return nil, nil
}
func (m *mockPayslipRepo) GetOriginalFile(_ context.Context, _ string, _ uuid.UUID) ([]byte, string, string, error) {
	return []byte("pdf"), "application/pdf", "test.pdf", nil
}
func (m *mockPayslipRepo) GetSummary(_ context.Context, _ uuid.UUID) (entity.PayslipSummary, error) {
	return entity.PayslipSummary{}, nil
}

func TestUpdatePayslip_WithBonuses(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)

	repo := newMockPayslipRepo()
	existing := entity.Payslip{
		ID:             uuid.New().String(),
		PeriodMonthNum: 6,
		PeriodYear:     2026,
		TaxClass:       "1",
		GrossPay:       5000.00,
		NetPay:         3200.00,
		PayoutAmount:   3200.00,
		Bonuses: []entity.Bonus{
			{Description: "Einmalzahlung", Amount: 600.00},
		},
	}
	_ = repo.Save(context.Background(), &existing)

	payslipSvc := service.NewPayslipService(repo, nil, nil, setupLogger())

	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).
		WithPayslipService(payslipSvc).
		WithPayslipRepository(repo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	payload, _ := json.Marshal(entity.Payslip{
		PeriodMonthNum: 6,
		PeriodYear:     2026,
		TaxClass:       "3",
		GrossPay:       5000.00,
		NetPay:         3200.00,
		PayoutAmount:   3200.00,
		Bonuses: []entity.Bonus{
			{Description: "TZUG Zusatzbetrag", Amount: 398.50},
			{Description: "TZUG-Tarifl. Zusatzgeld", Amount: 1581.60},
			{Description: "Urlaubsgeld", Amount: 793.38},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/v1/payslips/"+existing.ID+"/", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v — body: %s", status, rr.Body.String())
	}

	var result entity.Payslip
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.TaxClass != "3" {
		t.Errorf("expected TaxClass '3', got '%s'", result.TaxClass)
	}
	if len(result.Bonuses) != 3 {
		t.Fatalf("expected 3 Bonuses in response, got %d", len(result.Bonuses))
	}
	expectedBonuses := []struct {
		desc   string
		amount float64
	}{
		{"TZUG Zusatzbetrag", 398.50},
		{"TZUG-Tarifl. Zusatzgeld", 1581.60},
		{"Urlaubsgeld", 793.38},
	}
	for i, exp := range expectedBonuses {
		if result.Bonuses[i].Description != exp.desc {
			t.Errorf("Bonus[%d]: expected '%s', got '%s'", i, exp.desc, result.Bonuses[i].Description)
		}
		if result.Bonuses[i].Amount != exp.amount {
			t.Errorf("Bonus[%d]: expected %.2f, got %.2f", i, exp.amount, result.Bonuses[i].Amount)
		}
	}
}

func TestGetPayslip_IncludesBonuses(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)

	repo := newMockPayslipRepo()
	p := entity.Payslip{
		ID:             uuid.New().String(),
		PeriodMonthNum: 3,
		PeriodYear:     2026,
		GrossPay:       6000.00,
		NetPay:         3800.00,
		PayoutAmount:   3800.00,
		Bonuses: []entity.Bonus{
			{Description: "Urlaubsgeld", Amount: 793.38},
			{Description: "TZUG Zusatzbetrag", Amount: 398.50},
		},
	}
	_ = repo.Save(context.Background(), &p)

	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).
		WithPayslipRepository(repo).
		WithPayslipService(repo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/payslips/"+p.ID+"/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v", status)
	}

	var result entity.Payslip
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result.Bonuses) != 2 {
		t.Fatalf("expected 2 Bonuses, got %d", len(result.Bonuses))
	}
	if result.Bonuses[0].Description != "Urlaubsgeld" {
		t.Errorf("expected 'Urlaubsgeld', got '%s'", result.Bonuses[0].Description)
	}
}

func TestImportPayslipsBatch(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)
	repo := newMockPayslipRepo()

	staticParser := &mockPayslipParser{}
	aiParser := &mockPayslipParser{}
	payslipSvc := service.NewPayslipService(repo, staticParser, aiParser, setupLogger())

	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).
		WithPayslipService(payslipSvc).
		WithPayslipRepository(repo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for i := 1; i <= 2; i++ {
		part, err := writer.CreateFormFile("files", fmt.Sprintf("payslip%d.pdf", i))
		if err != nil {
			t.Fatal(err)
		}
		part.Write([]byte(fmt.Sprintf("pdf content %d", i)))
	}
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/payslips/import/batch/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Successful []entity.Payslip `json:"successful"`
		Failed     []interface{}    `json:"failed"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if len(resp.Successful) != 2 {
		t.Errorf("expected 2 successful imports, got %d. Body: %s", len(resp.Successful), w.Body.String())
	}
}

type mockPayslipParser struct{}

func (m *mockPayslipParser) Parse(_ context.Context, _ uuid.UUID, _ []byte) (entity.Payslip, error) {
	return entity.Payslip{
		PeriodMonthNum: 3,
		PeriodYear:     2026,
		EmployerName:   "Test Corp",
		GrossPay:       1000,
		NetPay:         800,
		PayoutAmount:   800,
	}, nil
}

func (m *mockPayslipParser) ParsePayslip(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, fileBytes []byte) (entity.Payslip, error) {
	return m.Parse(ctx, userID, fileBytes)
}

// --- Mock Sharing Service ---
type mockSharingSvc struct {
	dashboard entity.SharingDashboard
	err       error
}

func (m *mockSharingSvc) GetDashboard(ctx context.Context, userID uuid.UUID) (entity.SharingDashboard, error) {
	return m.dashboard, m.err
}

func TestGetSharingDashboard_Success(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)

	expectedDash := entity.SharingDashboard{
		SharedCategories: []entity.SharedCategorySummary{
			{Permissions: "view"},
			{Permissions: "owner"},
		},
	}

	mockSvc := &mockSharingSvc{dashboard: expectedDash}
	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).
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

	var res entity.SharingDashboard
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}

	if len(res.SharedCategories) != 2 {
		t.Errorf("expected 2 shared categories, got %d", len(res.SharedCategories))
	}
}

func TestGetSharingDashboard_ServiceUnavailable(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)

	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger)

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

	mockSvc := &mockSharingSvc{err: errors.New("database connection failed")}
	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).
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

// --- NEW ENRICHED TESTS ---

// --- Mock Reconciliation Service ---
type mockReconciliationSvc struct {
	suggestions []entity.ReconciliationPairSuggestion
	rec         entity.Reconciliation
	err         error
}

func (m *mockReconciliationSvc) SuggestReconciliations(ctx context.Context, userID uuid.UUID, matchWindowDays int) ([]entity.ReconciliationPairSuggestion, error) {
	return m.suggestions, m.err
}

func (m *mockReconciliationSvc) ReconcileStatements(ctx context.Context, userID uuid.UUID, settlementHash, targetHash string) (entity.Reconciliation, error) {
	return m.rec, m.err
}

func (m *mockReconciliationSvc) DeleteReconciliation(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return m.err
}

func TestGetReconciliationSuggestions(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)

	mockSvc := &mockReconciliationSvc{
		suggestions: []entity.ReconciliationPairSuggestion{
			{TargetTransaction: entity.Transaction{CounterpartyName: "abc"}, SourceTransaction: entity.Transaction{CounterpartyName: "def"}, MatchScore: 0.95},
		},
	}

	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).
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

	mockSvc := &mockReconciliationSvc{
		suggestions: []entity.ReconciliationPairSuggestion{
			{TargetTransaction: entity.Transaction{CounterpartyName: "abc"}, SourceTransaction: entity.Transaction{CounterpartyName: "def"}, MatchScore: 0.95},
		},
	}

	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).
		WithReconciliationService(mockSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	// Test singular alias and window_days parameter
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reconciliation/suggestions/?window_days=14", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for singular route, got %v", rr.Code)
	}

	var res []entity.ReconciliationPairSuggestion
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}

	if len(res) != 1 {
		t.Errorf("expected 1 suggestion, got %d", len(res))
	}
}

// --- Mock Planned Transaction Service ---
type mockPlannedTxSvc struct {
	txs []entity.PlannedTransaction
	err error
}

func (m *mockPlannedTxSvc) MatchTransactions(ctx context.Context, userID uuid.UUID, txns []entity.Transaction) error {
	return nil
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

func TestCreatePlannedTransaction(t *testing.T) {
	authSvc, token, userID := setupTestAuth(t)

	mockSvc := &mockPlannedTxSvc{}
	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).
		WithPlannedTransactionService(mockSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	payload := []byte(`{"amount": 1500.50, "description": "Rent", "date": "2026-05-01T00:00:00Z"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/planned-transactions/", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %v. Body: %s", status, rr.Body.String())
	}

	if len(mockSvc.txs) != 1 {
		t.Fatalf("expected 1 planned transaction to be created, got %d", len(mockSvc.txs))
	}
	if mockSvc.txs[0].UserID != userID {
		t.Errorf("expected UserID to match auth context")
	}
	if mockSvc.txs[0].Amount != 1500.50 {
		t.Errorf("expected amount 1500.50, got %f", mockSvc.txs[0].Amount)
	}
}

func TestGetSystemInfo(t *testing.T) {
	authSvc, token, userID := setupTestAuth(t)

	// Create a mock user service so adminMiddleware can verify the role
	userRepo := &mockUserRepo{user: entity.User{ID: userID, Username: "testadmin", Role: "admin"}}
	userSvc := service.NewUserService(userRepo, setupLogger())

	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "postgres", "db.internal", dummyPinger).
		WithUserService(userSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/info/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v", status)
	}

	var info map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&info); err != nil {
		t.Fatal(err)
	}

	if info["db_state"] != "connected" {
		t.Errorf("expected db_state 'connected', got '%s'", info["db_state"])
	}
	if info["storage_mode"] != "postgres" {
		t.Errorf("expected storage_mode 'postgres', got '%s'", info["storage_mode"])
	}
	if info["db_host"] != "db.internal" {
		t.Errorf("expected db_host 'db.internal', got '%s'", info["db_host"])
	}
}

// --- Mock Forecasting Service ---
type mockForecastingSvc struct {
	err error
}

func (m *mockForecastingSvc) GetCashFlowForecast(ctx context.Context, userID uuid.UUID, fromDate time.Time, toDate time.Time) (entity.CashFlowForecast, error) {
	return entity.CashFlowForecast{}, m.err
}

func (m *mockForecastingSvc) ExcludeForecast(ctx context.Context, userID uuid.UUID, forecastID uuid.UUID) error {
	return m.err
}

func (m *mockForecastingSvc) IncludeForecast(ctx context.Context, userID uuid.UUID, forecastID uuid.UUID) error {
	return m.err
}

func (m *mockForecastingSvc) ExcludePattern(ctx context.Context, userID uuid.UUID, matchTerm string) error {
	return m.err
}

func (m *mockForecastingSvc) IncludePattern(ctx context.Context, userID uuid.UUID, matchTerm string) error {
	return m.err
}

func (m *mockForecastingSvc) ListPatternExclusions(ctx context.Context, userID uuid.UUID) ([]entity.PatternExclusion, error) {
	return nil, m.err
}

func (m *mockForecastingSvc) CalculateCategoryAverage(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID, strategy string) (float64, error) {
	return 123.45, m.err
}

func TestGetForecast(t *testing.T) {
	authSvc, token, _ := setupTestAuth(t)

	mockSvc := &mockForecastingSvc{}
	dummyPinger := func(ctx context.Context) error { return nil }
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).
		WithForecastingService(mockSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/forecast/?from=2026-05-01&to=2026-05-31", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v", status)
	}
}
