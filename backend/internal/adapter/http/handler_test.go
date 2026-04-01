package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
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

// setupTestAuth creates a real AuthService with a mock repo and returns a valid JWT token
func setupTestAuth(t *testing.T) (*service.AuthService, string) {
	// Use MinCost so tests run fast
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	user := entity.User{
		ID:           uuid.New(),
		Username:     "testadmin",
		PasswordHash: string(hash),
	}

	repo := &mockUserRepo{user: user}
	authSvc := service.NewAuthService(repo, nil, nil, nil, "test-secret-key", setupLogger())

	token, err := authSvc.Login(context.Background(), "testadmin", "password")
	if err != nil {
		t.Fatalf("failed to login and get token: %v", err)
	}

	return authSvc, token
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

// FIX: Added userID parameter
func (m *mockCategoryRepo) FindByID(_ context.Context, _ uuid.UUID, _ uuid.UUID) (entity.Category, error) {
	return entity.Category{}, nil
}

// FIX: Added userID parameter
func (m *mockCategoryRepo) FindAll(_ context.Context, _ uuid.UUID) ([]entity.Category, error) {
	return m.categories, nil
}

// FIX: Added userID parameter
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
	// Return a stub to avoid "not found" errors in the service layer when finishing reconciliations
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
func (m *mockBankStmtRepo) UpdateTransactionCategory(_ context.Context, hash string, categoryID *uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockBankStmtRepo) MarkTransactionReconciled(_ context.Context, _ string, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockBankStmtRepo) MarkTransactionReviewed(_ context.Context, _ string, _ uuid.UUID) error {
	return nil
}

func (m *mockBankStmtRepo) LinkTransactionToStatement(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockBankStmtRepo) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

// Added missing method to satisfy port.BankStatementRepository
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
	authSvc := service.NewAuthService(repo, nil, nil, nil, "test-secret-key", setupLogger())
	token, _ := authSvc.Login(context.Background(), username, "password")
	return authSvc, token
}

func TestSettingsAccessControl(t *testing.T) {
	adminAuth, adminToken := setupTestUser(t, "admin", "admin")
	managerAuth, managerToken := setupTestUser(t, "manager", "manager")

	dummyPinger := func(ctx context.Context) error { return nil }

	// Helper to run request
	runReq := func(authSvc *service.AuthService, token string) int {
		// Use a local mock that implements GetAll to avoid nil dereference
		handler := apphttp.NewHandler(authSvc, nil, nil, &realMockSettingsSvc{}, nil, setupLogger(), "memory", "localhost", dummyPinger)
		// We need to inject the user service so adminMiddleware can fetch the user to check the role
		// In the test, authSvc and userSvc can use the same repo/logic
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

	// 1. Admin should succeed (or get 200 via our no-op mock)
	adminCode := runReq(adminAuth, adminToken)
	if adminCode == http.StatusUnauthorized || adminCode == http.StatusForbidden {
		t.Errorf("expected admin to bypass middleware, got %d", adminCode)
	}

	// 2. Manager should be forbidden
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

func TestHealthCheck(t *testing.T) {
	dummyPinger := func(ctx context.Context) error { return nil }
	// Added nil for bankSvc (5th argument)
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
	authSvc, token := setupTestAuth(t)
	dummyPinger := func(ctx context.Context) error { return nil }
	// Added nil for bankSvc (5th argument)
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	// Test successful password change
	payload := []byte(`{"old_password":"password", "new_password":"newpassword123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("expected 204 No Content for successful password change, got %v", status)
	}

	// Test failed password change (wrong old password)
	badPayload := []byte(`{"old_password":"wrongpassword", "new_password":"newpassword123"}`)
	badReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewBuffer(badPayload))
	badReq.Header.Set("Authorization", "Bearer "+token)
	badRr := httptest.NewRecorder()

	r.ServeHTTP(badRr, badReq)

	if status := badRr.Code; status != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for invalid old password, got %v", status)
	}
}

func TestGetTransactionAnalytics(t *testing.T) {
	authSvc, token := setupTestAuth(t)

	cat1 := uuid.New()
	cat2 := uuid.New()

	repo := &mockBankStmtRepo{
		txns: []entity.Transaction{
			{BookingDate: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), Amount: -50, CategoryID: &cat1, Description: "Supermarket"},
			{BookingDate: time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC), Amount: 100, CategoryID: &cat2, Description: "Employer"},
		},
	}

	// Inject the new TransactionService for Analytics
	txSvc := service.NewTransactionService(repo, nil, nil, nil, setupLogger())

	dummyPinger := func(ctx context.Context) error { return nil }
	// Added nil for bankSvc (5th argument)
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).
		WithTransactionService(txSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/analytics?from=2026-01-01&to=2026-01-31", nil)
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

func TestListCategories_Empty(t *testing.T) {
	authSvc, token := setupTestAuth(t)
	mockRepo := &mockCategoryRepo{}

	dummyPinger := func(ctx context.Context) error { return nil }
	// Added nil for bankSvc (5th argument)
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).WithCategoryRepository(mockRepo)

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
	authSvc, token := setupTestAuth(t)
	mockRepo := &mockCategoryRepo{}

	dummyPinger := func(ctx context.Context) error { return nil }
	// Added nil for bankSvc (5th argument)
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).WithCategoryRepository(mockRepo)

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
	authSvc, token := setupTestAuth(t)
	repo := &mockBankStmtRepo{}

	// Updated signature for BankStatementService
	svc := service.NewBankStatementService(repo, setupLogger())

	dummyPinger := func(ctx context.Context) error { return nil }
	// Added nil for bankSvc (5th argument)
	handler := apphttp.NewHandler(authSvc, nil, svc, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).
		WithBankStatementRepository(repo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	statementID := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/bank-statements/"+statementID, nil)
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
func (m *mockPayslipRepo) FindAll(_ context.Context, _ uuid.UUID) ([]entity.Payslip, error) {
	result := make([]entity.Payslip, 0, len(m.payslips))
	for _, p := range m.payslips {
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
func (m *mockPayslipRepo) GetOriginalFile(_ context.Context, _ string, _ uuid.UUID) ([]byte, string, string, error) {
	return []byte("pdf"), "application/pdf", "test.pdf", nil
}

func TestUpdatePayslip_WithBonuses(t *testing.T) {
	authSvc, token := setupTestAuth(t)

	repo := newMockPayslipRepo()
	existing := entity.Payslip{
		ID:             uuid.New().String(),
		PeriodMonthNum: 6,
		PeriodYear:     2026,
		EmployeeName:   "Max Mustermann",
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
	// Added nil for bankSvc (5th argument)
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).
		WithPayslipService(payslipSvc).
		WithPayslipRepository(repo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	payload, _ := json.Marshal(entity.Payslip{
		PeriodMonthNum: 6,
		PeriodYear:     2026,
		EmployeeName:   "Max Mustermann",
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

	req := httptest.NewRequest(http.MethodPut, "/api/v1/payslips/"+existing.ID, bytes.NewBuffer(payload))
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
	authSvc, token := setupTestAuth(t)

	repo := newMockPayslipRepo()
	p := entity.Payslip{
		ID:             uuid.New().String(),
		PeriodMonthNum: 3,
		PeriodYear:     2026,
		EmployeeName:   "Erika Musterfrau",
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
	// Added nil for bankSvc (5th argument)
	handler := apphttp.NewHandler(authSvc, nil, nil, nil, nil, setupLogger(), "memory", "localhost", dummyPinger).
		WithPayslipRepository(repo)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/payslips/"+p.ID, nil)
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
