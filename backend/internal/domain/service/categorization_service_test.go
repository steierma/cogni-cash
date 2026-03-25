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

// ---- Mock LLMClient --------------------------------------------------------

type mockLLMClient struct {
	result port.CategorizationResult
	err    error
}

func (m *mockLLMClient) Categorize(_ context.Context, _ port.CategorizationRequest) (port.CategorizationResult, error) {
	return m.result, m.err
}

// ---- Mock InvoiceRepository ------------------------------------------------

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

func (m *mockInvoiceRepo) FindByID(_ context.Context, id uuid.UUID) (entity.Invoice, error) {
	for _, inv := range m.saved {
		if inv.ID == id {
			return inv, nil
		}
	}
	return entity.Invoice{}, errors.New("not found")
}

func (m *mockInvoiceRepo) FindAll(_ context.Context) ([]entity.Invoice, error) {
	return m.saved, m.err
}

func (m *mockInvoiceRepo) Delete(_ context.Context, id uuid.UUID) error {
	return m.err
}

// ---- Test helpers ----------------------------------------------------------

var defaultCategories = []entity.Category{
	{ID: uuid.New(), Name: "Utilities"},
	{ID: uuid.New(), Name: "Software"},
	{ID: uuid.New(), Name: "Travel"},
}

// ---- Tests -----------------------------------------------------------------

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
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	svc := service.NewCategorizationService(llm, repo, defaultCategories, logger)

	inv, err := svc.CategorizeDocument(context.Background(), "Invoice for annual software license – €99.99")
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	svc := service.NewCategorizationService(&mockLLMClient{}, &mockInvoiceRepo{}, defaultCategories, logger)

	_, err := svc.CategorizeDocument(context.Background(), "")
	if !errors.Is(err, service.ErrEmptyRawText) {
		t.Errorf("expected ErrEmptyRawText, got %v", err)
	}
}

func TestCategorizeDocument_LLMError_PropagatesError(t *testing.T) {
	llmErr := errors.New("llm unavailable")
	llm := &mockLLMClient{err: llmErr}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	svc := service.NewCategorizationService(llm, &mockInvoiceRepo{}, defaultCategories, logger)

	_, err := svc.CategorizeDocument(context.Background(), "some invoice text")
	if !errors.Is(err, llmErr) {
		t.Errorf("expected llm error to propagate, got %v", err)
	}
}

func TestCategorizeDocument_UncategorizedFallback(t *testing.T) {
	llm := &mockLLMClient{
		result: port.CategorizationResult{},
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	svc := service.NewCategorizationService(llm, &mockInvoiceRepo{}, defaultCategories, logger)

	inv, err := svc.CategorizeDocument(context.Background(), "some invoice text")
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	svc := service.NewCategorizationService(llm, repo, defaultCategories, logger)

	_, err := svc.CategorizeDocument(context.Background(), "some invoice text")
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error to propagate, got %v", err)
	}
}
