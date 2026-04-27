package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

// imageMIMETypes is the set of image MIME types that bypass text extraction
// and go directly to the LLM via the multimodal CategorizeInvoiceImage path.
var imageMIMETypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

var allowedMIMETypes = map[string]bool{
	"application/pdf": true,
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
}

var ErrInvoiceDuplicate = fmt.Errorf("duplicate invoice")

type InvoiceService struct {
	invoiceRepo     port.InvoiceRepository
	aiCategorizer   port.InvoiceAICategorizer
	parser          port.InvoiceParser
	categoryRepo    port.CategoryRepository
	sharingRepo     port.SharingRepository
	currencyService *CurrencyService
	logger          *slog.Logger
}

func NewInvoiceService(
	invoiceRepo port.InvoiceRepository,
	aiCategorizer port.InvoiceAICategorizer,
	parser port.InvoiceParser,
	categoryRepo port.CategoryRepository,
	sharingRepo port.SharingRepository,
	logger *slog.Logger,
) *InvoiceService {
	return &InvoiceService{
		invoiceRepo:   invoiceRepo,
		aiCategorizer: aiCategorizer,
		parser:        parser,
		categoryRepo:  categoryRepo,
		sharingRepo:   sharingRepo,
		logger:        logger,
	}
}

func (s *InvoiceService) WithCurrencyService(svc *CurrencyService) *InvoiceService {
	s.currencyService = svc
	return s
}

func (s *InvoiceService) GetAll(ctx context.Context, filter entity.InvoiceFilter) ([]entity.Invoice, error) {
	return s.invoiceRepo.FindAll(ctx, filter)
}

func (s *InvoiceService) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Invoice, error) {
	return s.invoiceRepo.FindByID(ctx, id, userID)
}

func (s *InvoiceService) Update(ctx context.Context, invoice entity.Invoice) (entity.Invoice, error) {
	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return entity.Invoice{}, err
	}
	return s.invoiceRepo.FindByID(ctx, invoice.ID, invoice.UserID)
}

func (s *InvoiceService) UpdateCategoriesBulk(ctx context.Context, ids []uuid.UUID, categoryID *uuid.UUID, userID uuid.UUID) error {
	return s.invoiceRepo.UpdateCategoriesBulk(ctx, ids, categoryID, userID)
}

func (s *InvoiceService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.invoiceRepo.Delete(ctx, id, userID)
}

func (s *InvoiceService) GetOriginalFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) ([]byte, string, string, error) {
	return s.invoiceRepo.GetOriginalFile(ctx, id, userID)
}

func (s *InvoiceService) ShareInvoice(ctx context.Context, invoiceID, ownerID, sharedWithUserID uuid.UUID, permission string) error {
	// Verify ownership/existence
	_, err := s.invoiceRepo.FindByID(ctx, invoiceID, ownerID)
	if err != nil {
		return fmt.Errorf("invoice not found or unauthorized: %w", err)
	}

	return s.sharingRepo.ShareInvoice(ctx, invoiceID, ownerID, sharedWithUserID, permission)
}

func (s *InvoiceService) RevokeInvoiceShare(ctx context.Context, invoiceID, ownerID, sharedWithUserID uuid.UUID) error {
	return s.sharingRepo.RevokeInvoiceShare(ctx, invoiceID, ownerID, sharedWithUserID)
}

func (s *InvoiceService) ListInvoiceShares(ctx context.Context, invoiceID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	// Verify ownership
	_, err := s.invoiceRepo.FindByID(ctx, invoiceID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found or unauthorized: %w", err)
	}

	return s.sharingRepo.ListInvoiceShares(ctx, invoiceID, ownerID)
}

// ── File-based import ──────────────────────────────────────────────────────

// ImportFromFile reads the file bytes, checks for duplicates via SHA-256, extracts
// text from the file, calls the LLM for categorization, and persists the invoice.
func (s *InvoiceService) ImportFromFile(
	ctx context.Context,
	userID uuid.UUID,
	fileName, mimeType string,
	fileBytes []byte,
	overrides port.ImportOverrides,
) (entity.Invoice, error) {
	// 1. Content-hash for deduplication
	h := sha256.New()
	h.Write(fileBytes)
	contentHash := hex.EncodeToString(h.Sum(nil))

	exists, err := s.invoiceRepo.ExistsByContentHash(ctx, contentHash, userID)
	if err != nil {
		return entity.Invoice{}, fmt.Errorf("invoice service: check hash: %w", err)
	}
	if exists {
		return entity.Invoice{}, fmt.Errorf("%w: %s", ErrInvoiceDuplicate, contentHash)
	}

	normMIME := strings.ToLower(strings.TrimSpace(mimeType))

	var invoice entity.Invoice

	if len(overrides.Splits) > 0 {
		// 2. If splits are provided manually, skip AI categorization but still link the file
		s.logger.Info("Manual splits provided, skipping AI categorization", "file", fileName, "user_id", userID)

		totalAmount := 0.0
		for _, sp := range overrides.Splits {
			totalAmount += sp.Amount
		}

		invoice = entity.Invoice{
			ID:          uuid.New(),
			UserID:      userID,
			Vendor:      entity.Vendor{ID: uuid.New(), Name: "Manual Import"}, // Placeholder
			Amount:      totalAmount,
			Currency:    "EUR", // Default fallback
			IssuedAt:    time.Now().UTC(),
			Splits:      overrides.Splits,
		}
	} else if imageMIMETypes[normMIME] {
		// 2a. Image path — skip text extraction, go directly to multimodal LLM
		s.logger.Info("Image file detected, using multimodal LLM path",
			"file", fileName, "mime", mimeType, "user_id", userID)

		categories, err := s.categoryRepo.FindAll(ctx, userID)
		if err != nil {
			return entity.Invoice{}, fmt.Errorf("fetch categories: %w", err)
		}
		catNames := make([]string, len(categories))
		for i, c := range categories {
			catNames[i] = c.Name
		}

		result, err := s.aiCategorizer.CategorizeInvoiceImage(ctx, userID, fileName, normMIME, fileBytes, catNames)
		if err != nil {
			s.logger.Error("Multimodal image categorization failed", "error", err, "user_id", userID)
			return entity.Invoice{}, fmt.Errorf("invoice service: image categorization: %w", err)
		}

		cat := s.resolveCategory(result.InvoiceName, categories)
		vendor := entity.Vendor{ID: uuid.New(), Name: result.VendorName}

		invoice = entity.Invoice{
			ID:          uuid.New(),
			UserID:      userID,
			Vendor:      vendor,
			CategoryID:  &cat.ID,
			Amount:      result.Amount,
			Currency:    result.Currency,
			IssuedAt:    time.Now().UTC(),
			Description: result.Description,
		}
	} else {
		// 2b. Non-image path — extract text then categorize via text prompt
		rawText, err := s.parser.Extract(ctx, userID, fileBytes, mimeType)
		if err != nil || rawText == "" {
			s.logger.Warn("Text extraction failed or returned empty; using filename as fallback raw text",
				"file", fileName, "error", err, "user_id", userID)
			rawText = fileName // minimal fallback so LLM still gets something
		}

		// 3. CategorizeInvoice via LLM
		invoice, err = s.categorize(ctx, userID, rawText)
		if err != nil {
			return entity.Invoice{}, err
		}
	}

	// 4. Manual overrides from caller
	if overrides.VendorName != nil && *overrides.VendorName != "" {
		invoice.Vendor.Name = *overrides.VendorName
	}
	if overrides.Amount != nil {
		invoice.Amount = *overrides.Amount
	}
	if overrides.Currency != nil && *overrides.Currency != "" {
		invoice.Currency = *overrides.Currency
	}
	if overrides.IssuedAt != nil {
		invoice.IssuedAt = *overrides.IssuedAt
	}
	if overrides.CategoryID != nil {
		invoice.CategoryID = overrides.CategoryID
	}

	// 5. Attach file metadata
	invoice.UserID = userID
	invoice.ContentHash = contentHash
	invoice.OriginalFileName = fileName
	invoice.OriginalFileContent = fileBytes

	// Set split IDs
	for i := range invoice.Splits {
		if invoice.Splits[i].ID == uuid.Nil {
			invoice.Splits[i].ID = uuid.New()
		}
		invoice.Splits[i].InvoiceID = invoice.ID
		invoice.Splits[i].UserID = userID
	}

	// 6. Persist
	if err := s.invoiceRepo.Save(ctx, invoice); err != nil {
		return entity.Invoice{}, fmt.Errorf("invoice service: save: %w", err)
	}

	// Trigger asynchronous currency conversion
	if s.currencyService != nil {
		go func() {
			cCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			if err := s.currencyService.UpdateBaseAmountsForUser(cCtx, userID); err != nil {
				s.logger.Error("Background currency conversion failed for invoice", "user_id", userID, "error", err)
			}
		}()
	}

	s.logger.Info("Invoice imported successfully", "id", invoice.ID, "file", fileName, "category", invoice.CategoryID, "user_id", userID)
	return invoice, nil
}

// ImportManual persists an invoice with manually provided metadata and splits.
func (s *InvoiceService) ImportManual(ctx context.Context, userID uuid.UUID, invoice entity.Invoice) (entity.Invoice, error) {
	invoice.UserID = userID
	if invoice.ID == uuid.Nil {
		invoice.ID = uuid.New()
	}

	if invoice.Currency == "" {
		invoice.Currency = "EUR"
	}

	// Use provided file content if present to generate hash
	if len(invoice.OriginalFileContent) > 0 {
		h := sha256.New()
		h.Write(invoice.OriginalFileContent)
		invoice.ContentHash = hex.EncodeToString(h.Sum(nil))

		exists, err := s.invoiceRepo.ExistsByContentHash(ctx, invoice.ContentHash, userID)
		if err != nil {
			return entity.Invoice{}, err
		}
		if exists {
			return entity.Invoice{}, fmt.Errorf("%w: %s", ErrInvoiceDuplicate, invoice.ContentHash)
		}
	}

	// Set split IDs
	for i := range invoice.Splits {
		if invoice.Splits[i].ID == uuid.Nil {
			invoice.Splits[i].ID = uuid.New()
		}
		invoice.Splits[i].InvoiceID = invoice.ID
		invoice.Splits[i].UserID = userID
	}

	if err := s.invoiceRepo.Save(ctx, invoice); err != nil {
		return entity.Invoice{}, fmt.Errorf("invoice service: manual save: %w", err)
	}

	return s.invoiceRepo.FindByID(ctx, invoice.ID, userID)
}

// ── Raw-text categorization (legacy / direct API) ─────────────────────────

// CategorizeDocument processes raw document text, calls the LLM, persists, and
// returns the invoice. Satisfies port.InvoiceCategorizationUseCase.
func (s *InvoiceService) CategorizeDocument(ctx context.Context, userID uuid.UUID, rawText string) (entity.Invoice, error) {
	invoice, err := s.categorize(ctx, userID, rawText)
	if err != nil {
		return entity.Invoice{}, err
	}
	if err := s.invoiceRepo.Save(ctx, invoice); err != nil {
		return entity.Invoice{}, fmt.Errorf("invoice service: save manual: %w", err)
	}

	// Trigger asynchronous currency conversion
	if s.currencyService != nil {
		go func() {
			cCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			if err := s.currencyService.UpdateBaseAmountsForUser(cCtx, userID); err != nil {
				s.logger.Error("Background currency conversion failed for manual invoice", "user_id", userID, "error", err)
			}
		}()
	}

	return invoice, nil
}

// ── Private helpers ──────────────────────────────────────────────────────────

func (s *InvoiceService) categorize(ctx context.Context, userID uuid.UUID, rawText string) (entity.Invoice, error) {
	categories, err := s.categoryRepo.FindAll(ctx, userID)
	if err != nil {
		return entity.Invoice{}, fmt.Errorf("fetch categories: %w", err)
	}
	catNames := make([]string, len(categories))
	for i, c := range categories {
		catNames[i] = c.Name
	}

	result, err := s.aiCategorizer.CategorizeInvoice(ctx, userID, port.CategorizationRequest{
		RawText:    rawText,
		Categories: catNames,
	})
	if err != nil {
		return entity.Invoice{}, fmt.Errorf("invoice service: categorize: %w", err)
	}

	cat := s.resolveCategory(result.InvoiceName, categories)
	vendor := entity.Vendor{ID: uuid.New(), Name: result.VendorName}

	return entity.Invoice{
		ID:          uuid.New(),
		UserID:      userID,
		Vendor:      vendor,
		CategoryID:  &cat.ID,
		Amount:      result.Amount,
		Currency:    result.Currency,
		IssuedAt:    time.Now().UTC(),
		Description: result.Description,
	}, nil
}

func (s *InvoiceService) resolveCategory(name string, categories []entity.Category) entity.Category {
	for _, c := range categories {
		if c.Name == name {
			return c
		}
	}
	return entity.Category{ID: uuid.New(), Name: "Uncategorized"}
}
