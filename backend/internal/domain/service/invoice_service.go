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
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

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
	invoiceRepo     port.InvoiceRepository
	categoryRepo    port.CategoryRepository
	sharingRepo     port.SharingRepository
	notificationSvc port.NotificationUseCase
	userRepo        port.UserRepository
	parser          port.InvoiceParser // extracts text from invoice files (PDF, …)
	aiCategorizer   port.InvoiceAICategorizer
	logger          *slog.Logger
}

// NewInvoiceService creates a new InvoiceService.
func NewInvoiceService(
	invoiceRepo port.InvoiceRepository,
	categoryRepo port.CategoryRepository,
	sharingRepo port.SharingRepository,
	notificationSvc port.NotificationUseCase,
	userRepo port.UserRepository,
	parser port.InvoiceParser,
	aiCategorizer port.InvoiceAICategorizer,
	logger *slog.Logger,
) *InvoiceService {
	if logger == nil {
		logger = slog.Default()
	}
	return &InvoiceService{
		invoiceRepo:     invoiceRepo,
		categoryRepo:    categoryRepo,
		sharingRepo:     sharingRepo,
		notificationSvc: notificationSvc,
		userRepo:        userRepo,
		parser:          parser,
		aiCategorizer:   aiCategorizer,
		logger:          logger,
	}
}

// ── Sharing ────────────────────────────────────────────────────────────────

func (s *InvoiceService) ShareInvoice(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID, permission string) error {
	if ownerID == sharedWithID {
		return fmt.Errorf("cannot share invoice with yourself")
	}

	// Verify ownership
	inv, err := s.invoiceRepo.FindByID(ctx, invoiceID, ownerID)
	if err != nil {
		return fmt.Errorf("invoice not found or not owned by you: %w", err)
	}
	if inv.UserID != ownerID {
		return fmt.Errorf("only the owner can share an invoice")
	}

	if err := s.sharingRepo.ShareInvoice(ctx, invoiceID, ownerID, sharedWithID, permission); err != nil {
		return err
	}

	// Notify the user
	if s.notificationSvc != nil && s.userRepo != nil {
		recipient, err := s.userRepo.FindByID(ctx, sharedWithID)
		if err == nil {
			s.logger.Info("Sending invoice sharing notification", "to", recipient.Email, "invoice_id", invoiceID)
			// Using SendTestEmail as a placeholder. In a production app, we would use a dedicated
			// SendSharingInvitation method.
			_ = s.notificationSvc.SendTestEmail(ctx, recipient.Email, ownerID)
		}
	}

	return nil
}

func (s *InvoiceService) RevokeInvoiceShare(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID) error {
	inv, err := s.invoiceRepo.FindByID(ctx, invoiceID, ownerID)
	if err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	// Either the owner or the recipient can revoke/remove the share
	if inv.UserID != ownerID && ownerID != sharedWithID {
		return fmt.Errorf("unauthorized to revoke this share")
	}

	return s.sharingRepo.RevokeInvoiceShare(ctx, invoiceID, inv.UserID, sharedWithID)
}

func (s *InvoiceService) ListInvoiceShares(ctx context.Context, invoiceID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	_, err := s.invoiceRepo.FindByID(ctx, invoiceID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found or unauthorized: %w", err)
	}

	return s.sharingRepo.ListInvoiceShares(ctx, invoiceID, ownerID)
}

// WithCategories is deprecated and removed as categories are now fetched per user.

// ── File-based import ──────────────────────────────────────────────────────

// ImportFromFile reads the file bytes, checks for duplicates via SHA-256, extracts
// text from the file, calls the LLM for categorization, and persists the invoice.
// Returns ErrInvoiceDuplicate when the same file was already imported.
// Image files (JPEG, PNG, GIF, WEBP) bypass text extraction and are sent
// directly to the LLM via the multimodal CategorizeInvoiceImage path.
func (s *InvoiceService) ImportFromFile(
	ctx context.Context,
	userID uuid.UUID,
	fileName, mimeType string,
	fileBytes []byte,
	categoryID *uuid.UUID,
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

	if imageMIMETypes[normMIME] {
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

	// 4. Caller-supplied category overrides LLM result
	if categoryID != nil {
		invoice.CategoryID = categoryID
	}

	// 5. Attach file metadata
	invoice.UserID = userID
	invoice.ContentHash = contentHash
	invoice.OriginalFileName = fileName
	invoice.OriginalFileContent = fileBytes

	// 6. Persist
	if err := s.invoiceRepo.Save(ctx, invoice); err != nil {
		return entity.Invoice{}, fmt.Errorf("invoice service: save: %w", err)
	}

	s.logger.Info("Invoice imported successfully", "id", invoice.ID, "file", fileName, "category", invoice.CategoryID, "user_id", userID)
	return invoice, nil
}

// ── Raw-text categorization (legacy / direct API) ─────────────────────────

// CategorizeDocument processes raw document text, calls the LLM, persists, and
// returns the invoice. Satisfies port.InvoiceCategorizationUseCase.
func (s *InvoiceService) CategorizeDocument(ctx context.Context, userID uuid.UUID, rawText string) (entity.Invoice, error) {
	if rawText == "" {
		return entity.Invoice{}, ErrEmptyRawText
	}
	invoice, err := s.categorize(ctx, userID, rawText)
	if err != nil {
		return entity.Invoice{}, err
	}
	invoice.UserID = userID
	if err := s.invoiceRepo.Save(ctx, invoice); err != nil {
		return entity.Invoice{}, fmt.Errorf("invoice service: save: %w", err)
	}
	s.logger.Info("Invoice categorized and saved", "id", invoice.ID, "user_id", userID)
	return invoice, nil
}

// ── CRUD ───────────────────────────────────────────────────────────────────

// GetAll returns every invoice ordered by issued_at desc.
func (s *InvoiceService) GetAll(ctx context.Context, filter entity.InvoiceFilter) ([]entity.Invoice, error) {
	return s.invoiceRepo.FindAll(ctx, filter)
}

// GetByID returns a single invoice or ErrInvoiceNotFound.
func (s *InvoiceService) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Invoice, error) {
	inv, err := s.invoiceRepo.FindByID(ctx, id, userID)
	if err != nil {
		return entity.Invoice{}, fmt.Errorf("%w: %w", ErrInvoiceNotFound, err)
	}
	return inv, nil
}

// Update overwrites the mutable fields of an existing invoice.
func (s *InvoiceService) Update(ctx context.Context, invoice entity.Invoice) (entity.Invoice, error) {
	s.logger.Info("Updating invoice", "id", invoice.ID, "user_id", invoice.UserID)
	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return entity.Invoice{}, fmt.Errorf("invoice service: update: %w", err)
	}
	return s.invoiceRepo.FindByID(ctx, invoice.ID, invoice.UserID)
}

// Delete removes an invoice by ID.
func (s *InvoiceService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	s.logger.Info("Deleting invoice", "id", id, "user_id", userID)
	return s.invoiceRepo.Delete(ctx, id, userID)
}

// GetOriginalFile returns the raw bytes, MIME type, and file name of the stored file.
func (s *InvoiceService) GetOriginalFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) ([]byte, string, string, error) {
	return s.invoiceRepo.GetOriginalFile(ctx, id, userID)
}

// ── internal helpers ───────────────────────────────────────────────────────

// categorize calls the LLM and builds an Invoice entity (without persisting it).
func (s *InvoiceService) categorize(ctx context.Context, userID uuid.UUID, rawText string) (entity.Invoice, error) {
	categories, err := s.categoryRepo.FindAll(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to fetch categories for LLM", "user_id", userID, "error", err)
		// continue with empty category list if we must, or return error?
		// Let's return error to be safe.
		return entity.Invoice{}, fmt.Errorf("fetch categories: %w", err)
	}

	catNames := make([]string, len(categories))
	for i, c := range categories {
		catNames[i] = c.Name
	}

	s.logger.Info("Calling LLM for invoice categorization", "categories", catNames, "user_id", userID)
	result, err := s.aiCategorizer.CategorizeInvoice(ctx, userID, port.CategorizationRequest{
		RawText:    rawText,
		Categories: catNames,
	})
	if err != nil {
		s.logger.Error("LLM categorization failed", "error", err)
		return entity.Invoice{}, err
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

// resolveCategory finds the matching Category entity by name, or returns an
// "Uncategorized" placeholder when the LLM returns an unknown name.
func (s *InvoiceService) resolveCategory(name string, categories []entity.Category) entity.Category {
	for _, c := range categories {
		if c.Name == name {
			return c
		}
	}
	return entity.Category{ID: uuid.New(), Name: "Uncategorized"}
}
