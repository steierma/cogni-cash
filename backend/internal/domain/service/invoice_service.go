package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

// ErrInvoiceDuplicate is returned when an invoice with the same content hash already exists.
var ErrInvoiceDuplicate = entity.ErrInvoiceDuplicate

// ErrInvoiceNotFound is returned when a referenced invoice cannot be located.
var ErrInvoiceNotFound = entity.ErrInvoiceNotFound

// ErrEmptyRawText is returned when a document has no extractable text.
var ErrEmptyRawText = entity.ErrEmptyRawText

// InvoiceService orchestrates all invoice operations:
//   - file-based import (text extraction → LLM categorization → dedup → persist)
//   - raw-text categorization (legacy / API path)
//   - CRUD (list, get, update, delete)
//   - original-file download
//
// It depends exclusively on ports and has no infrastructure imports.
type InvoiceService struct {
	invoiceRepo port.InvoiceRepository
	parser      port.InvoiceParser // extracts text from invoice files (PDF, …)
	llm         port.LLMClient
	categories  []entity.Category
	logger      *slog.Logger
}

// NewInvoiceService creates a new InvoiceService.
func NewInvoiceService(
	invoiceRepo port.InvoiceRepository,
	parser port.InvoiceParser,
	llm port.LLMClient,
	categories []entity.Category,
	logger *slog.Logger,
) *InvoiceService {
	if logger == nil {
		logger = slog.Default()
	}
	return &InvoiceService{
		invoiceRepo: invoiceRepo,
		parser:      parser,
		llm:         llm,
		categories:  categories,
		logger:      logger,
	}
}

// WithCategories replaces the category list used for LLM prompt building.
// This allows the handler/main to refresh categories after service construction.
func (s *InvoiceService) WithCategories(cats []entity.Category) {
	s.categories = cats
}

// ── File-based import ──────────────────────────────────────────────────────

// ImportFromFile reads the file bytes, checks for duplicates via SHA-256, extracts
// text from the file, calls the LLM for categorization, and persists the invoice.
// Returns ErrInvoiceDuplicate when the same file was already imported.
func (s *InvoiceService) ImportFromFile(
	ctx context.Context,
	filePath, fileName, mimeType string,
	fileBytes []byte,
	categoryID *uuid.UUID,
) (entity.Invoice, error) {
	// 1. Content-hash for deduplication
	h := sha256.New()
	h.Write(fileBytes)
	contentHash := hex.EncodeToString(h.Sum(nil))

	exists, err := s.invoiceRepo.ExistsByContentHash(ctx, contentHash)
	if err != nil {
		return entity.Invoice{}, fmt.Errorf("invoice service: check hash: %w", err)
	}
	if exists {
		return entity.Invoice{}, fmt.Errorf("%w: %s", ErrInvoiceDuplicate, contentHash)
	}

	// 2. Write to a temp file so the parser can open it by path
	tmp, err := os.CreateTemp("", "invoice-*"+fileExtFromName(fileName))
	if err != nil {
		return entity.Invoice{}, fmt.Errorf("invoice service: create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(fileBytes); err != nil {
		tmp.Close()
		return entity.Invoice{}, fmt.Errorf("invoice service: write temp file: %w", err)
	}
	tmp.Close()

	// 3. Extract text
	rawText, err := s.parser.Extract(ctx, tmp.Name(), mimeType)
	if err != nil || rawText == "" {
		s.logger.Warn("Text extraction failed or returned empty; using filename as fallback raw text",
			"file", fileName, "error", err)
		rawText = fileName // minimal fallback so LLM still gets something
	}

	// 4. Categorize via LLM
	invoice, err := s.categorize(ctx, rawText)
	if err != nil {
		return entity.Invoice{}, err
	}

	// 5. Caller-supplied category overrides LLM result
	if categoryID != nil {
		invoice.CategoryID = categoryID
	}

	// 6. Attach file metadata
	invoice.ContentHash = contentHash
	invoice.OriginalFileName = fileName
	invoice.OriginalFileMime = mimeType
	invoice.OriginalFileSize = int64(len(fileBytes))
	invoice.OriginalFileContent = fileBytes

	// 7. Persist
	if err := s.invoiceRepo.Save(ctx, invoice); err != nil {
		return entity.Invoice{}, fmt.Errorf("invoice service: save: %w", err)
	}

	s.logger.Info("Invoice imported successfully", "id", invoice.ID, "file", fileName, "category", invoice.CategoryID)
	return invoice, nil
}

// ── Raw-text categorization (legacy / direct API) ─────────────────────────

// CategorizeDocument processes raw document text, calls the LLM, persists, and
// returns the invoice. Satisfies port.InvoiceCategorizationUseCase.
func (s *InvoiceService) CategorizeDocument(ctx context.Context, rawText string) (entity.Invoice, error) {
	if rawText == "" {
		return entity.Invoice{}, ErrEmptyRawText
	}
	invoice, err := s.categorize(ctx, rawText)
	if err != nil {
		return entity.Invoice{}, err
	}
	if err := s.invoiceRepo.Save(ctx, invoice); err != nil {
		return entity.Invoice{}, fmt.Errorf("invoice service: save: %w", err)
	}
	s.logger.Info("Invoice categorized and saved", "id", invoice.ID)
	return invoice, nil
}

// ── CRUD ───────────────────────────────────────────────────────────────────

// GetAll returns every invoice ordered by issued_at desc.
func (s *InvoiceService) GetAll(ctx context.Context) ([]entity.Invoice, error) {
	return s.invoiceRepo.FindAll(ctx)
}

// GetByID returns a single invoice or ErrInvoiceNotFound.
func (s *InvoiceService) GetByID(ctx context.Context, id uuid.UUID) (entity.Invoice, error) {
	inv, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return entity.Invoice{}, fmt.Errorf("%w: %w", ErrInvoiceNotFound, err)
	}
	return inv, nil
}

// Update overwrites the mutable fields of an existing invoice.
func (s *InvoiceService) Update(ctx context.Context, invoice entity.Invoice) (entity.Invoice, error) {
	s.logger.Info("Updating invoice", "id", invoice.ID)
	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return entity.Invoice{}, fmt.Errorf("invoice service: update: %w", err)
	}
	return s.invoiceRepo.FindByID(ctx, invoice.ID)
}

// Delete removes an invoice by ID.
func (s *InvoiceService) Delete(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("Deleting invoice", "id", id)
	return s.invoiceRepo.Delete(ctx, id)
}

// GetOriginalFile returns the raw bytes, MIME type, and file name of the stored file.
func (s *InvoiceService) GetOriginalFile(ctx context.Context, id uuid.UUID) ([]byte, string, string, error) {
	return s.invoiceRepo.GetOriginalFile(ctx, id)
}

// ── internal helpers ───────────────────────────────────────────────────────

// categorize calls the LLM and builds an Invoice entity (without persisting it).
func (s *InvoiceService) categorize(ctx context.Context, rawText string) (entity.Invoice, error) {
	catNames := make([]string, len(s.categories))
	for i, c := range s.categories {
		catNames[i] = c.Name
	}

	s.logger.Info("Calling LLM for invoice categorization", "categories", catNames)
	result, err := s.llm.Categorize(ctx, port.CategorizationRequest{
		RawText:    rawText,
		Categories: catNames,
	})
	if err != nil {
		s.logger.Error("LLM categorization failed", "error", err)
		return entity.Invoice{}, err
	}

	cat := s.resolveCategory(result.CategoryName)
	vendor := entity.Vendor{ID: uuid.New(), Name: result.VendorName}

	return entity.Invoice{
		ID:          uuid.New(),
		Vendor:      vendor,
		CategoryID:  &cat.ID,
		Amount:      result.Amount,
		Currency:    result.Currency,
		IssuedAt:    time.Now().UTC(),
		Description: result.Description,
		RawText:     rawText,
	}, nil
}

// resolveCategory finds the matching Category entity by name, or returns an
// "Uncategorized" placeholder when the LLM returns an unknown name.
func (s *InvoiceService) resolveCategory(name string) entity.Category {
	for _, c := range s.categories {
		if c.Name == name {
			return c
		}
	}
	return entity.Category{ID: uuid.New(), Name: "Uncategorized"}
}

func fileExtFromName(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return name[i:]
		}
	}
	return ""
}
