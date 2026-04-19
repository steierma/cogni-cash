package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

type DocumentService struct {
	repo        port.DocumentRepository
	aiParser    port.DocumentAIParser
	payslipRepo port.PayslipRepository
	invoiceRepo port.InvoiceRepository
}

func NewDocumentService(repo port.DocumentRepository, aiParser port.DocumentAIParser, payslipRepo port.PayslipRepository, invoiceRepo port.InvoiceRepository) *DocumentService {
	return &DocumentService{
		repo:        repo,
		aiParser:    aiParser,
		payslipRepo: payslipRepo,
		invoiceRepo: invoiceRepo,
	}
}

func (s *DocumentService) Upload(ctx context.Context, req entity.DocumentUploadRequest) (entity.Document, error) {
	// 1. Calculate content hash (SHA-256)
	hash := sha256.Sum256(req.FileContent)
	contentHash := hex.EncodeToString(hash[:])

	// 2. Deduplicate
	exists, err := s.repo.ExistsByHash(ctx, req.UserID, contentHash)
	if err != nil {
		return entity.Document{}, err
	}
	if exists {
		return entity.Document{}, entity.ErrDocumentDuplicate
	}

	var docType entity.DocumentType
	var metadata map[string]interface{}
	var extractedText string

	if req.SkipAI {
		// 3a. Manual classification and metadata
		docType = req.ManualType
		metadata = make(map[string]interface{})
		if !req.DocumentDate.IsZero() {
			metadata["date"] = req.DocumentDate.Format("2006-01-02")
			metadata["year"] = float64(req.DocumentDate.Year())
		}
	} else {
		// 3b. AI Classification and Metadata Extraction
		docType, metadata, extractedText, err = s.aiParser.ClassifyAndExtract(ctx, req.UserID, req.FileName, req.ContentType, req.FileContent)
		if err != nil {
			// Log error but proceed with default classification if LLM fails
			docType = entity.DocTypeOther
			if metadata == nil {
				metadata = make(map[string]interface{})
			}
		}
	}

	// 4. Create document entity
	doc := entity.Document{
		UserID:              req.UserID,
		Type:                docType,
		OriginalFileName:    req.FileName,
		ContentHash:         contentHash,
		MimeType:            req.ContentType,
		OriginalFileContent: req.FileContent,
		Metadata:            metadata,
		ExtractedText:       extractedText,
	}

	// 5. Save to repository
	return s.repo.Save(ctx, doc)
}

func (s *DocumentService) List(ctx context.Context, filter entity.DocumentFilter) ([]entity.Document, error) {
	return s.repo.FindAll(ctx, filter)
}

func (s *DocumentService) GetDetail(ctx context.Context, id, userID uuid.UUID) (entity.Document, error) {
	return s.repo.FindByID(ctx, id, userID)
}

func (s *DocumentService) Update(ctx context.Context, id, userID uuid.UUID, req entity.DocumentUpdateRequest) (entity.Document, error) {
	doc, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		return entity.Document{}, err
	}

	if req.OriginalFileName != nil {
		doc.OriginalFileName = *req.OriginalFileName
	}
	if req.Type != nil {
		doc.Type = *req.Type
	}
	if req.DocumentDate != nil {
		if doc.Metadata == nil {
			doc.Metadata = make(map[string]interface{})
		}
		doc.Metadata["date"] = req.DocumentDate.Format("2006-01-02")
		doc.Metadata["year"] = float64(req.DocumentDate.Year())
	}
	if req.Metadata != nil {
		if doc.Metadata == nil {
			doc.Metadata = make(map[string]interface{})
		}
		for k, v := range req.Metadata {
			doc.Metadata[k] = v
		}
	}

	return s.repo.Update(ctx, doc) // Assumes Update is defined in your DocumentRepository interface
}

func (s *DocumentService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *DocumentService) Download(ctx context.Context, id, userID uuid.UUID) (content []byte, mimeType string, fileName string, err error) {
	doc, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, "", "", err
	}
	return doc.OriginalFileContent, doc.MimeType, doc.OriginalFileName, nil
}

func (s *DocumentService) GetTaxYearSummary(ctx context.Context, userID uuid.UUID, year int) (entity.TaxYearSummary, error) {
	// 1. Fetch all general documents for the user
	docs, err := s.repo.FindAll(ctx, entity.DocumentFilter{UserID: userID})
	if err != nil {
		return entity.TaxYearSummary{}, err
	}

	summary := entity.TaxYearSummary{
		Year:      year,
		Documents: []entity.Document{},
	}

	for _, doc := range docs {
		// 2. Determine document year (metadata override, then fallback to CreatedAt)
		docYear := doc.CreatedAt.Year()
		if y, ok := doc.Metadata["year"].(float64); ok {
			docYear = int(y)
		} else if y, ok := doc.Metadata["year"].(int); ok {
			docYear = y
		}

		if docYear != year {
			continue
		}

		// 3. Document belongs to the requested year
		summary.Documents = append(summary.Documents, doc)

		// 4. Aggregate data based on document type
		switch doc.Type {
		case entity.DocTypeTaxCertificate:
			// Values typically extracted by AI into metadata
			if v, ok := doc.Metadata["gross_income"].(float64); ok {
				summary.TotalGrossIncome += v
			}
			if v, ok := doc.Metadata["net_income"].(float64); ok {
				summary.TotalNetIncome += v
			}
			if v, ok := doc.Metadata["income_tax"].(float64); ok {
				summary.TotalIncomeTax += v
			}

		case entity.DocTypeReceipt:
			// Check if document was marked as tax deductible
			isDeductible, _ := doc.Metadata["tax_deductible"].(bool)
			if isDeductible {
				if v, ok := doc.Metadata["amount"].(float64); ok {
					summary.TotalDeductible += v
				}
			}
		}
	}

	// 5. Fetch and Aggregate Payslips for the year
	payslips, err := s.payslipRepo.FindAll(ctx, entity.PayslipFilter{UserID: userID, Year: year})
	if err == nil {
		for _, p := range payslips {
			summary.TotalGrossIncome += p.GrossPay
			summary.TotalNetIncome += p.NetPay
			// If income tax is not explicitly stored, estimate it as the difference (coarse estimate)
			summary.TotalIncomeTax += (p.GrossPay - p.NetPay)
		}
	}

	// 6. Fetch and Aggregate Invoices for the year (assuming all uploaded invoices are deductible)
	invoices, err := s.invoiceRepo.FindAll(ctx, entity.InvoiceFilter{UserID: userID, Year: year, IncludeShared: true})
	if err == nil {
		for _, inv := range invoices {
			summary.TotalDeductible += inv.Amount
		}
	}

	// 7. Sort documents by date (latest first)
	sort.Slice(summary.Documents, func(i, j int) bool {
		dateI := summary.Documents[i].CreatedAt
		if d, ok := summary.Documents[i].Metadata["date"].(string); ok {
			if t, err := time.Parse("2006-01-02", d); err == nil {
				dateI = t
			}
		}

		dateJ := summary.Documents[j].CreatedAt
		if d, ok := summary.Documents[j].Metadata["date"].(string); ok {
			if t, err := time.Parse("2006-01-02", d); err == nil {
				dateJ = t
			}
		}

		if dateI.Equal(dateJ) {
			return summary.Documents[i].CreatedAt.After(summary.Documents[j].CreatedAt)
		}
		return dateI.After(dateJ)
	})

	return summary, nil
}
