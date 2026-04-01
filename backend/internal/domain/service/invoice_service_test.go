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

// ── Mock LLMClient ───────────────────────────────────────────────────────────

type mockLLMClient struct {
	result port.CategorizationResult
	err    error
}

func (m *mockLLMClient) Categorize(_ context.Context, _ uuid.UUID, _ port.CategorizationRequest) (port.CategorizationResult, error) {
	return m.result, m.err
}

// ── Mock InvoiceParser ───────────────────────────────────────────────────────

type mockInvoiceParser struct {
	text string
	err  error
}

func (m *mockInvoiceParser) Extract(_ context.Context, _ uuid.UUID, _, _ string) (string, error) {
	return m.text, m.err
}

// ── Mock InvoiceRepository ───────────────────────────────────────────────────

type mockInvoiceRepo struct {
	saved []entity.Invoice
	err   error
}

func (m *mockInvoiceRepo) Save(_ context.Context, inv entity.Invoice) error {
	if m.err != nil {
		return m.err
	}
	m.saved = append(m.saved, inv)
	return nil
}

func (m *mockInvoiceRepo) Update(_ context.Context, inv entity.Invoice) error {
	if m.err != nil {
		return m.err
	}
	for i, s := range m.saved {
		if s.ID == inv.ID {
			m.saved[i] = inv
			return nil
		}
	}
	return entity.ErrInvoiceNotFound
}

func (m *mockInvoiceRepo) FindByID(_ context.Context, id uuid.UUID, _ uuid.UUID) (entity.Invoice, error) {
	for _, inv := range m.saved {
		if inv.ID == id {
			return inv, nil
		}
	}
	return entity.Invoice{}, entity.ErrInvoiceNotFound
}

func (m *mockInvoiceRepo) FindAll(_ context.Context, _ uuid.UUID) ([]entity.Invoice, error) {
	return m.saved, m.err
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
			return inv.OriginalFileContent, inv.OriginalFileMime, inv.OriginalFileName, nil
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
func (m *mockCatRepo) Save(_ context.Context, cat entity.Category) (entity.Category, error) { return cat, nil }
func (m *mockCatRepo) Update(_ context.Context, cat entity.Category) (entity.Category, error) { return cat, nil }
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

func newTestInvoiceSvc(llm port.LLMClient, repo *mockInvoiceRepo) *service.InvoiceService {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	catRepo := &mockCatRepo{cats: defaultCategories}
	return service.NewInvoiceService(repo, catRepo, &mockInvoiceParser{text: "extracted text"}, llm, logger)
}

// ── CategorizeDocument tests ─────────────────────────────────────────────────

func TestCategorizeDocument_HappyPath(t *testing.T) {
	llm := &mockLLMClient{
		result: port.CategorizationResult{
			CategoryName: "Software",
			VendorName:   "Acme Corp",
			Amount:       99.99,
			Currency:     "EUR",
			Description:  "Annual license",
		},
	}
	repo := &mockInvoiceRepo{}
	svc := newTestInvoiceSvc(llm, repo)

	inv, err := svc.CategorizeDocument(context.Background(), dummyUserID, "Invoice for annual software license – €99.99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.CategoryID == nil || *inv.CategoryID != defaultCategories[1].ID {
		t.Errorf("expected category ID for 'Software' (%v), got %v", defaultCategories[1].ID, inv.CategoryID)
	}
	if inv.Vendor.Name != "Acme Corp" {
		t.Errorf("expected vendor 'Acme Corp', got '%s'", inv.Vendor.Name)
	}
	if inv.Amount != 99.99 {
		t.Errorf("expected amount 99.99, got %f", inv.Amount)
	}
	if inv.Currency != "EUR" {
		t.Errorf("expected currency 'EUR', got '%s'", inv.Currency)
	}
	if len(repo.saved) != 1 {
		t.Errorf("expected 1 saved invoice, got %d", len(repo.saved))
	}
}

func TestCategorizeDocument_EmptyRawText_ReturnsError(t *testing.T) {
	svc := newTestInvoiceSvc(&mockLLMClient{}, &mockInvoiceRepo{})

	_, err := svc.CategorizeDocument(context.Background(), dummyUserID, "")
	if !errors.Is(err, service.ErrEmptyRawText) {
		t.Errorf("expected ErrEmptyRawText, got %v", err)
	}
}

func TestCategorizeDocument_LLMError_PropagatesError(t *testing.T) {
	llmErr := errors.New("llm unavailable")
	svc := newTestInvoiceSvc(&mockLLMClient{err: llmErr}, &mockInvoiceRepo{})

	_, err := svc.CategorizeDocument(context.Background(), dummyUserID, "some invoice text")
	if !errors.Is(err, llmErr) {
		t.Errorf("expected llm error to propagate, got %v", err)
	}
}

func TestCategorizeDocument_UncategorizedFallback(t *testing.T) {
	svc := newTestInvoiceSvc(&mockLLMClient{result: port.CategorizationResult{}}, &mockInvoiceRepo{})

	inv, err := svc.CategorizeDocument(context.Background(), dummyUserID, "some invoice text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.CategoryID == nil {
		t.Error("expected a generated CategoryID for 'Uncategorized', got nil")
	}
}

func TestCategorizeDocument_RepositoryError_PropagatesError(t *testing.T) {
	llm := &mockLLMClient{
		result: port.CategorizationResult{CategoryName: "Software", VendorName: "v", Amount: 1, Currency: "EUR"},
	}
	repoErr := errors.New("db write error")
	repo := &mockInvoiceRepo{err: repoErr}
	svc := newTestInvoiceSvc(llm, repo)

	_, err := svc.CategorizeDocument(context.Background(), dummyUserID, "some invoice text")
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error to propagate, got %v", err)
	}
}

// ── ImportFromFile tests ─────────────────────────────────────────────────────

func TestImportFromFile_HappyPath(t *testing.T) {
	llm := &mockLLMClient{
		result: port.CategorizationResult{
			CategoryName: "Software",
			VendorName:   "Acme Corp",
			Amount:       49.99,
			Currency:     "EUR",
			Description:  "SaaS subscription",
		},
	}
	repo := &mockInvoiceRepo{}
	svc := newTestInvoiceSvc(llm, repo)

	fileBytes := []byte("fake pdf content")
	inv, err := svc.ImportFromFile(context.Background(), dummyUserID, "dummy/path", "invoice.pdf", "application/pdf", fileBytes, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.OriginalFileName != "invoice.pdf" {
		t.Errorf("expected original file name 'invoice.pdf', got '%s'", inv.OriginalFileName)
	}
	if inv.ContentHash == "" {
		t.Error("expected content hash to be set")
	}
	if inv.OriginalFileSize != int64(len(fileBytes)) {
		t.Errorf("expected file size %d, got %d", len(fileBytes), inv.OriginalFileSize)
	}
	if len(repo.saved) != 1 {
		t.Errorf("expected 1 saved invoice, got %d", len(repo.saved))
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
	svc := newTestInvoiceSvc(&mockLLMClient{}, repo)

	// Import once to get the real hash stored
	_, _ = svc.ImportFromFile(context.Background(), dummyUserID, "dummy/path", "a.pdf", "application/pdf", fileBytes, nil)

	// Try to import the exact same bytes again
	_, err := svc.ImportFromFile(context.Background(), dummyUserID, "dummy/path", "a.pdf", "application/pdf", fileBytes, nil)
	if !errors.Is(err, service.ErrInvoiceDuplicate) {
		t.Errorf("expected ErrInvoiceDuplicate on second import, got %v", err)
	}
}

func TestImportFromFile_CallerCategoryOverridesLLM(t *testing.T) {
	llm := &mockLLMClient{
		result: port.CategorizationResult{CategoryName: "Software", VendorName: "v", Amount: 1, Currency: "EUR"},
	}
	repo := &mockInvoiceRepo{}
	svc := newTestInvoiceSvc(llm, repo)

	overrideCatID := uuid.New()
	inv, err := svc.ImportFromFile(context.Background(), dummyUserID, "dummy/path", "inv.pdf", "application/pdf", []byte("pdf"), &overrideCatID)
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
	svc := newTestInvoiceSvc(&mockLLMClient{
		result: port.CategorizationResult{VendorName: "OldCo", Amount: 10, Currency: "EUR"},
	}, repo)

	inv, _ := svc.CategorizeDocument(context.Background(), dummyUserID, "text")

	inv.Vendor.Name = "NewCo"
	inv.Amount = 99.0
	updated, err := svc.Update(context.Background(), inv)
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}
	if updated.Vendor.Name != "NewCo" {
		t.Errorf("expected vendor 'NewCo', got '%s'", updated.Vendor.Name)
	}
	if updated.Amount != 99.0 {
		t.Errorf("expected amount 99.0, got %f", updated.Amount)
	}
}

func TestUpdate_NotFound_ReturnsError(t *testing.T) {
	repo := &mockInvoiceRepo{}
	svc := newTestInvoiceSvc(&mockLLMClient{}, repo)

	_, err := svc.Update(context.Background(), entity.Invoice{ID: uuid.New(), UserID: dummyUserID})
	if !errors.Is(err, entity.ErrInvoiceNotFound) {
		t.Errorf("expected ErrInvoiceNotFound, got %v", err)
	}
}

func TestDelete_RemovesInvoice(t *testing.T) {
	repo := &mockInvoiceRepo{}
	svc := newTestInvoiceSvc(&mockLLMClient{
		result: port.CategorizationResult{VendorName: "v", Amount: 1, Currency: "EUR"},
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
	svc := newTestInvoiceSvc(&mockLLMClient{}, &mockInvoiceRepo{})

	err := svc.Delete(context.Background(), uuid.New(), dummyUserID)
	if !errors.Is(err, entity.ErrInvoiceNotFound) {
		t.Errorf("expected ErrInvoiceNotFound, got %v", err)
	}
}

func TestGetAll_ReturnsSavedInvoices(t *testing.T) {
	repo := &mockInvoiceRepo{}
	svc := newTestInvoiceSvc(&mockLLMClient{
		result: port.CategorizationResult{VendorName: "v", Amount: 1, Currency: "EUR"},
	}, repo)

	for i := 0; i < 3; i++ {
		_, _ = svc.CategorizeDocument(context.Background(), dummyUserID, "text")
	}

	all, err := svc.GetAll(context.Background(), dummyUserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 invoices, got %d", len(all))
	}
}
