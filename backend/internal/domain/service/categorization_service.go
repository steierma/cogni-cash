package service

import (
	"context"
	"errors"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"log/slog"

	"github.com/google/uuid"
)

// ErrEmptyRawText is returned when a document has no extractable text.
var ErrEmptyRawText = errors.New("categorization service: raw text must not be empty")

// CategorizationService orchestrates the classification of an invoice document.
// It depends only on ports (interfaces) and has no infrastructure imports.
type CategorizationService struct {
	llm         port.LLMClient
	invoiceRepo port.InvoiceRepository
	categories  []entity.Category
	Logger      *slog.Logger // Structured logger
}

// NewCategorizationService creates a new CategorizationService.
func NewCategorizationService(llm port.LLMClient, invoiceRepo port.InvoiceRepository, categories []entity.Category, logger *slog.Logger) *CategorizationService {
	return &CategorizationService{
		llm:         llm,
		invoiceRepo: invoiceRepo,
		categories:  categories,
		Logger:      logger,
	}
}

// CategorizeDocument processes raw document text, calls the LLM, builds an Invoice
// entity, persists it, and returns the saved invoice.
func (s *CategorizationService) CategorizeDocument(ctx context.Context, rawText string) (entity.Invoice, error) {
	if rawText == "" {
		s.Logger.Warn("Empty raw text provided to CategorizeDocument")
		return entity.Invoice{}, ErrEmptyRawText
	}

	categoryNames := make([]string, len(s.categories))
	for i, c := range s.categories {
		categoryNames[i] = c.Name
	}

	s.Logger.Info("Calling LLM for categorization", "categories", categoryNames)
	result, err := s.llm.Categorize(ctx, port.CategorizationRequest{
		RawText:    rawText,
		Categories: categoryNames,
	})
	if err != nil {
		s.Logger.Error("LLM categorization failed", "error", err)
		return entity.Invoice{}, err
	}

	// Resolve the matched category entity.
	category := s.resolveCategory(result.CategoryName)
	vendor := entity.Vendor{ID: uuid.New(), Name: result.VendorName}

	invoice := entity.Invoice{
		ID:          uuid.New(),
		Vendor:      vendor,
		CategoryID:  &category.ID,
		Amount:      result.Amount,
		Currency:    result.Currency,
		IssuedAt:    time.Now().UTC(),
		Description: result.Description,
		RawText:     rawText,
	}

	s.Logger.Info("Saving categorized invoice", "id", invoice.ID, "category_id", invoice.CategoryID)
	if err := s.invoiceRepo.Save(ctx, invoice); err != nil {
		s.Logger.Error("Failed to save invoice", "id", invoice.ID, "error", err)
		return entity.Invoice{}, err
	}

	s.Logger.Info("Invoice categorized and saved successfully", "id", invoice.ID)
	return invoice, nil
}

// resolveCategory finds the matching Category entity by name, or creates an
// "Uncategorized" placeholder when the LLM returns an unknown name.
func (s *CategorizationService) resolveCategory(name string) entity.Category {
	for _, c := range s.categories {
		if c.Name == name {
			return c
		}
	}
	return entity.Category{ID: uuid.New(), Name: "Uncategorized"}
}
