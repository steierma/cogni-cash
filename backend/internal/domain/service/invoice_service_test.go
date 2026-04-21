package service_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
)

// ── Mock AICategorizer ───────────────────────────────────────────────────────

type mockAICategorizer struct {
	result      port.InvoiceCategorizationResult
	imageResult port.InvoiceCategorizationResult
	err         error
	imageErr    error
}

func (m *mockAICategorizer) CategorizeInvoice(_ context.Context, _ uuid.UUID, _ port.CategorizationRequest) (port.InvoiceCategorizationResult, error) {
	return m.result, m.err
}

func (m *mockAICategorizer) CategorizeInvoiceImage(_ context.Context, _ uuid.UUID, _ string, _ string, _ []byte, _ []string) (port.InvoiceCategorizationResult, error) {
	if m.imageErr != nil {
		return port.InvoiceCategorizationResult{}, m.imageErr
	}
	if m.imageResult != (port.InvoiceCategorizationResult{}) {
		return m.imageResult, nil
	}
	return m.result, m.err
}

// ── Mock InvoiceParser ───────────────────────────────────────────────────────

type mockInvoiceParser struct {
	text string
	err  error
}

func (m *mockInvoiceParser) Extract(_ context.Context, _ uuid.UUID, _ []byte, _ string) (string, error) {
	return m.text, m.err
}

// ── Mock Invoice Repository ───────────────────────────────────────────────────

type mockInvoiceRepo struct {
	saved []entity.Invoice
	err   error
}

func (m *mockInvoiceRepo) Save(_ context.Context, inv entity.Invoice) error {
	if m.err != nil {
		return m.err
	}
	// Simple mock: if it exists, update it; else, append
	for i, existing := range m.saved {
		if existing.ID == inv.ID {
			m.saved[i] = inv
			return nil
		}
	}
	m.saved = append(m.saved, inv)
	return nil
}

func (m *mockInvoiceRepo) Update(_ context.Context, inv entity.Invoice) error {
	if m.err != nil {
		return m.err
	}
	for i, existing := range m.saved {
		if existing.ID == inv.ID {
			m.saved[i] = inv
			return nil
		}
	}
	return entity.ErrInvoiceNotFound
}

func (m *mockInvoiceRepo) UpdateBaseAmount(_ context.Context, _ uuid.UUID, _ float64, _ string, _ uuid.UUID) error {
	return nil
}

func (m *mockInvoiceRepo) FindByID(_ context.Context, id uuid.UUID, _ uuid.UUID) (entity.Invoice, error) {
	for _, inv := range m.saved {
		if inv.ID == id {
			return inv, nil
		}
	}
	return entity.Invoice{}, entity.ErrInvoiceNotFound
}

func (m *mockInvoiceRepo) FindAll(_ context.Context, _ entity.InvoiceFilter) ([]entity.Invoice, error) {
	return m.saved, nil
}

func (m *mockInvoiceRepo) Delete(_ context.Context, id uuid.UUID, _ uuid.UUID) error {
	if m.err != nil {
		return m.err
	}
	for i, inv := range m.saved {
		if inv.ID == id {
			m.saved = append(m.saved[:i], m.saved[i+1:]...)
			return nil
		}
	}
	return entity.ErrInvoiceNotFound
}

func (m *mockInvoiceRepo) DeleteSplits(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return nil
}

func (m *mockInvoiceRepo) ExistsByContentHash(_ context.Context, hash string, _ uuid.UUID) (bool, error) {
	for _, inv := range m.saved {
		if inv.ContentHash == hash {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockInvoiceRepo) GetOriginalFile(_ context.Context, id uuid.UUID, _ uuid.UUID) ([]byte, string, string, error) {
	for _, inv := range m.saved {
		if inv.ID == id {
			return inv.OriginalFileContent, "application/pdf", inv.OriginalFileName, nil
		}
	}
	return nil, "", "", entity.ErrInvoiceNotFound
}

// ── Mock Category Repository ───────────────────────────────────────────────────

type mockCatRepo struct {
	cats []entity.Category
}

func (m *mockCatRepo) FindAll(_ context.Context, _ uuid.UUID) ([]entity.Category, error) {
	return m.cats, nil
}
func (m *mockCatRepo) Save(_ context.Context, cat entity.Category) (entity.Category, error) {
	return cat, nil
}
func (m *mockCatRepo) Update(_ context.Context, cat entity.Category) (entity.Category, error) {
	return cat, nil
}
func (m *mockCatRepo) FindByID(_ context.Context, id uuid.UUID, _ uuid.UUID) (entity.Category, error) {
	for _, c := range m.cats {
		if c.ID == id {
			return c, nil
		}
	}
	return entity.Category{}, errors.New("not found")
}
func (m *mockCatRepo) Delete(_ context.Context, id uuid.UUID, _ uuid.UUID) error { return nil }

// ── Test helpers ─────────────────────────────────────────────────────────────

var defaultCategories = []entity.Category{
	{ID: uuid.New(), Name: "Utilities"},
	{ID: uuid.New(), Name: "Software"},
	{ID: uuid.New(), Name: "Travel"},
}

var dummyUserID = uuid.New()

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

type mockNotification struct{}

func (m *mockNotification) SendWelcomeEmail(_ context.Context, _ entity.User) error { return nil }
func (m *mockNotification) SendPasswordResetEmail(_ context.Context, _ entity.User, _ string) error {
	return nil
}
func (m *mockNotification) SendTestEmail(_ context.Context, _ string, _ uuid.UUID) error { return nil }
func (m *mockNotification) SendInvoiceShareNotification(_ context.Context, _ entity.Invoice, _ uuid.UUID, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockNotification) SendCategoryShareNotification(_ context.Context, _ entity.Category, _ uuid.UUID, _ uuid.UUID, _ string) error {
	return nil
}

type mockUserRepo struct{}

func (m *mockUserRepo) FindByUsername(_ context.Context, _ string) (entity.User, error) {
	return entity.User{}, nil
}
func (m *mockUserRepo) FindByID(_ context.Context, _ uuid.UUID) (entity.User, error) {
	return entity.User{}, nil
}
func (m *mockUserRepo) GetAdminID(_ context.Context) (uuid.UUID, error)               { return uuid.Nil, nil }
func (m *mockUserRepo) FindAll(_ context.Context, _ string) ([]entity.User, error)    { return nil, nil }
func (m *mockUserRepo) Create(_ context.Context, _ entity.User) error                 { return nil }
func (m *mockUserRepo) Update(_ context.Context, _ entity.User) error                 { return nil }
func (m *mockUserRepo) Upsert(_ context.Context, _ entity.User) error                 { return nil }
func (m *mockUserRepo) UpdatePassword(_ context.Context, _ uuid.UUID, _ string) error { return nil }
func (m *mockUserRepo) Delete(_ context.Context, _ uuid.UUID) error                   { return nil }

func newTestInvoiceSvc(aiCategorizer port.InvoiceAICategorizer, repo *mockInvoiceRepo) *service.InvoiceService {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	catRepo := &mockCatRepo{cats: defaultCategories}
	return service.NewInvoiceService(repo, aiCategorizer, &mockInvoiceParser{text: "extracted text"}, catRepo, &mockSharingRepo{}, logger)
}

// ── CategorizeDocument tests ─────────────────────────────────────────────────

func TestCategorizeDocument_HappyPath(t *testing.T) {
	aiCategorizer := &mockAICategorizer{
		result: port.InvoiceCategorizationResult{
			InvoiceName: "Software",
			VendorName:  "JetBrains",
			Amount:      199.99,
			Currency:    "EUR",
			Description: "IDE Subscription",
		},
	}
	repo := &mockInvoiceRepo{}
	svc := newTestInvoiceSvc(aiCategorizer, repo)

	inv, err := svc.CategorizeDocument(context.Background(), dummyUserID, "IDE receipt text...")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inv.Vendor.Name != "JetBrains" {
		t.Errorf("expected vendor 'JetBrains', got '%s'", inv.Vendor.Name)
	}
	if inv.Amount != 199.99 {
		t.Errorf("expected amount 199.99, got %f", inv.Amount)
	}
	if inv.Description != "IDE Subscription" {
		t.Errorf("expected description 'IDE Subscription', got '%s'", inv.Description)
	}
}

// ── ImportFromFile tests ─────────────────────────────────────────────────────

func TestImportFromFile_HappyPath(t *testing.T) {
	aiCategorizer := &mockAICategorizer{
		result: port.InvoiceCategorizationResult{InvoiceName: "Software", VendorName: "v", Amount: 1, Currency: "EUR"},
	}
	repo := &mockInvoiceRepo{}
	svc := newTestInvoiceSvc(aiCategorizer, repo)

	fileBytes := []byte("fake pdf content")
	inv, err := svc.ImportFromFile(context.Background(), dummyUserID, "invoice.pdf", "application/pdf", fileBytes, port.ImportOverrides{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.OriginalFileName != "invoice.pdf" {
		t.Errorf("expected original file name 'invoice.pdf', got '%s'", inv.OriginalFileName)
	}
	if len(repo.saved) != 1 {
		t.Errorf("expected 1 invoice saved, got %d", len(repo.saved))
	}
}

func TestImportFromFile_DuplicateHash_ReturnsError(t *testing.T) {
	// Pre-seed an invoice with a known hash
	fileBytes := []byte("same pdf content")
	repo := &mockInvoiceRepo{
		saved: []entity.Invoice{
			{ID: uuid.New(), UserID: dummyUserID, ContentHash: "badbadbadbad"}, // hash won't match below
		},
	}
	svc := newTestInvoiceSvc(&mockAICategorizer{}, repo)

	_, _ = svc.ImportFromFile(context.Background(), dummyUserID, "a.pdf", "application/pdf", fileBytes, port.ImportOverrides{})

	// Try to import the exact same bytes again
	_, err := svc.ImportFromFile(context.Background(), dummyUserID, "a.pdf", "application/pdf", fileBytes, port.ImportOverrides{})
	if !errors.Is(err, service.ErrInvoiceDuplicate) {
		t.Errorf("expected ErrInvoiceDuplicate on second import, got %v", err)
	}
}

func TestImportFromFile_CallerCategoryOverridesLLM(t *testing.T) {
	aiCategorizer := &mockAICategorizer{
		result: port.InvoiceCategorizationResult{InvoiceName: "Software", VendorName: "v", Amount: 1, Currency: "EUR"},
	}
	repo := &mockInvoiceRepo{}
	svc := newTestInvoiceSvc(aiCategorizer, repo)

	overrideCatID := uuid.New()
	inv, err := svc.ImportFromFile(context.Background(), dummyUserID, "inv.pdf", "application/pdf", []byte("pdf"), port.ImportOverrides{CategoryID: &overrideCatID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.CategoryID == nil || *inv.CategoryID != overrideCatID {
		t.Errorf("expected caller-supplied category %v, got %v", overrideCatID, inv.CategoryID)
	}
}

// ── CRUD tests ───────────────────────────────────────────────────────────────

func TestUpdate_ChangesVendorAndAmount(t *testing.T) {
	repo := &mockInvoiceRepo{}
	svc := newTestInvoiceSvc(&mockAICategorizer{
		result: port.InvoiceCategorizationResult{VendorName: "v", Amount: 1, Currency: "EUR"},
	}, repo)

	inv, _ := svc.CategorizeDocument(context.Background(), dummyUserID, "text")

	inv.Vendor.Name = "New Vendor"
	inv.Amount = 500.0

	updated, err := svc.Update(context.Background(), inv)
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	if updated.Vendor.Name != "New Vendor" {
		t.Errorf("expected vendor 'New Vendor', got '%s'", updated.Vendor.Name)
	}
	if updated.Amount != 500.0 {
		t.Errorf("expected amount 500.0, got %f", updated.Amount)
	}
}

func TestDelete_RemovesFromRepo(t *testing.T) {
	repo := &mockInvoiceRepo{}
	svc := newTestInvoiceSvc(&mockAICategorizer{
		result: port.InvoiceCategorizationResult{VendorName: "v", Amount: 1, Currency: "EUR"},
	}, repo)

	inv, _ := svc.CategorizeDocument(context.Background(), dummyUserID, "text")

	if err := svc.Delete(context.Background(), inv.ID, dummyUserID); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
	if len(repo.saved) != 0 {
		t.Errorf("expected 0 invoices after delete, got %d", len(repo.saved))
	}
}

func TestDelete_NotFound_ReturnsError(t *testing.T) {
	svc := newTestInvoiceSvc(&mockAICategorizer{}, &mockInvoiceRepo{})

	err := svc.Delete(context.Background(), uuid.New(), dummyUserID)
	if !errors.Is(err, entity.ErrInvoiceNotFound) {
		t.Errorf("expected ErrInvoiceNotFound, got %v", err)
	}
}

func TestGetAll_ReturnsSavedInvoices(t *testing.T) {
	repo := &mockInvoiceRepo{}
	svc := newTestInvoiceSvc(&mockAICategorizer{
		result: port.InvoiceCategorizationResult{VendorName: "v", Amount: 1, Currency: "EUR"},
	}, repo)

	for i := 0; i < 3; i++ {
		_, _ = svc.CategorizeDocument(context.Background(), dummyUserID, "text")
	}

	all, err := svc.GetAll(context.Background(), entity.InvoiceFilter{UserID: dummyUserID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 invoices, got %d", len(all))
	}
}
