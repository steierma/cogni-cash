package service_test

import (
	"context"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockDocumentRepoDocSvc struct {
	saveFunc         func(ctx context.Context, doc entity.Document) (entity.Document, error)
	findAllFunc      func(ctx context.Context, filter entity.DocumentFilter) ([]entity.Document, error)
	findByIDFunc     func(ctx context.Context, id, userID uuid.UUID) (entity.Document, error)
	updateFunc       func(ctx context.Context, doc entity.Document) (entity.Document, error)
	deleteFunc       func(ctx context.Context, id, userID uuid.UUID) error
	existsByHashFunc func(ctx context.Context, userID uuid.UUID, contentHash string) (bool, error)
}

func (m *mockDocumentRepoDocSvc) Save(ctx context.Context, doc entity.Document) (entity.Document, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, doc)
	}
	return doc, nil
}

func (m *mockDocumentRepoDocSvc) FindAll(ctx context.Context, filter entity.DocumentFilter) ([]entity.Document, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx, filter)
	}
	return nil, nil
}

func (m *mockDocumentRepoDocSvc) FindByID(ctx context.Context, id, userID uuid.UUID) (entity.Document, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id, userID)
	}
	return entity.Document{}, nil
}

func (m *mockDocumentRepoDocSvc) Update(ctx context.Context, doc entity.Document) (entity.Document, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, doc)
	}
	return doc, nil
}

func (m *mockDocumentRepoDocSvc) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id, userID)
	}
	return nil
}

func (m *mockDocumentRepoDocSvc) ExistsByHash(ctx context.Context, userID uuid.UUID, contentHash string) (bool, error) {
	if m.existsByHashFunc != nil {
		return m.existsByHashFunc(ctx, userID, contentHash)
	}
	return false, nil
}

type mockDocumentAIParserDocSvc struct {
	classifyAndExtractFunc func(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, fileBytes []byte) (docType entity.DocumentType, metadata map[string]interface{}, extractedText string, err error)
}

func (m *mockDocumentAIParserDocSvc) ClassifyAndExtract(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, fileBytes []byte) (docType entity.DocumentType, metadata map[string]interface{}, extractedText string, err error) {
	if m.classifyAndExtractFunc != nil {
		return m.classifyAndExtractFunc(ctx, userID, fileName, mimeType, fileBytes)
	}
	return entity.DocTypeOther, nil, "", nil
}

type mockPayslipRepoDocSvc struct {
	findAllFunc func(context.Context, entity.PayslipFilter) ([]entity.Payslip, error)
}

func (m *mockPayslipRepoDocSvc) Save(context.Context, *entity.Payslip) error { return nil }
func (m *mockPayslipRepoDocSvc) FindAll(ctx context.Context, filter entity.PayslipFilter) ([]entity.Payslip, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx, filter)
	}
	return nil, nil
}
func (m *mockPayslipRepoDocSvc) FindByID(context.Context, string, uuid.UUID) (entity.Payslip, error) {
	return entity.Payslip{}, nil
}
func (m *mockPayslipRepoDocSvc) Update(context.Context, *entity.Payslip) error   { return nil }
func (m *mockPayslipRepoDocSvc) UpdateBaseAmount(context.Context, string, float64, float64, float64, string, uuid.UUID) error {
	return nil
}
func (m *mockPayslipRepoDocSvc) Delete(context.Context, string, uuid.UUID) error { return nil }
func (m *mockPayslipRepoDocSvc) ExistsByHash(context.Context, string, uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockPayslipRepoDocSvc) ExistsByOriginalFileName(context.Context, string, uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockPayslipRepoDocSvc) GetOriginalFile(context.Context, string, uuid.UUID) ([]byte, string, string, error) {
	return nil, "", "", nil
}
func (m *mockPayslipRepoDocSvc) GetSummary(context.Context, uuid.UUID) (entity.PayslipSummary, error) {
	return entity.PayslipSummary{}, nil
}

type mockInvoiceRepoDocSvc struct {
	findAllFunc func(context.Context, entity.InvoiceFilter) ([]entity.Invoice, error)
}

func (m *mockInvoiceRepoDocSvc) Save(context.Context, entity.Invoice) error { return nil }
func (m *mockInvoiceRepoDocSvc) FindAll(ctx context.Context, filter entity.InvoiceFilter) ([]entity.Invoice, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc(ctx, filter)
	}
	return nil, nil
}
func (m *mockInvoiceRepoDocSvc) FindByID(context.Context, uuid.UUID, uuid.UUID) (entity.Invoice, error) {
	return entity.Invoice{}, nil
}
func (m *mockInvoiceRepoDocSvc) Update(context.Context, entity.Invoice) error { return nil }
func (m *mockInvoiceRepoDocSvc) UpdateCategoriesBulk(context.Context, []uuid.UUID, *uuid.UUID, uuid.UUID) error {
	return nil
}
func (m *mockInvoiceRepoDocSvc) UpdateBaseAmount(context.Context, uuid.UUID, float64, string, uuid.UUID) error {
	return nil
}
func (m *mockInvoiceRepoDocSvc) Delete(context.Context, uuid.UUID, uuid.UUID) error { return nil }

func (m *mockInvoiceRepoDocSvc) DeleteSplits(context.Context, uuid.UUID, uuid.UUID) error        { return nil }
func (m *mockInvoiceRepoDocSvc) ExistsByContentHash(context.Context, string, uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockInvoiceRepoDocSvc) GetOriginalFile(context.Context, uuid.UUID, uuid.UUID) ([]byte, string, string, error) {
	return nil, "", "", nil
}
func (m *mockInvoiceRepoDocSvc) ToggleTaxDeductible(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

// --- Tests ---

func TestDocumentService_Upload(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()

	t.Run("Upload with SkipAI = true", func(t *testing.T) {
		repo := &mockDocumentRepoDocSvc{
			existsByHashFunc: func(ctx context.Context, userID uuid.UUID, contentHash string) (bool, error) {
				return false, nil
			},
			saveFunc: func(ctx context.Context, doc entity.Document) (entity.Document, error) {
				return doc, nil
			},
		}
		aiParser := &mockDocumentAIParserDocSvc{}
		svc := service.NewDocumentService(repo, aiParser, &mockPayslipRepoDocSvc{}, &mockInvoiceRepoDocSvc{})

		docDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		req := entity.DocumentUploadRequest{
			UserID:       userID,
			FileContent:  []byte("test content"),
			FileName:     "test.pdf",
			ContentType:  "application/pdf",
			SkipAI:       true,
			ManualType:   entity.DocTypeTaxCertificate,
			DocumentDate: docDate,
		}

		doc, err := svc.Upload(ctx, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if doc.Type != entity.DocTypeTaxCertificate {
			t.Errorf("expected type %s, got %s", entity.DocTypeTaxCertificate, doc.Type)
		}

		if doc.Metadata["date"] != "2024-01-15" {
			t.Errorf("expected metadata date 2024-01-15, got %v", doc.Metadata["date"])
		}

		if doc.Metadata["year"] != 2024.0 {
			t.Errorf("expected metadata year 2024.0, got %v", doc.Metadata["year"])
		}
	})

	t.Run("Upload with SkipAI = false (AI used)", func(t *testing.T) {
		repo := &mockDocumentRepoDocSvc{
			existsByHashFunc: func(ctx context.Context, userID uuid.UUID, contentHash string) (bool, error) {
				return false, nil
			},
			saveFunc: func(ctx context.Context, doc entity.Document) (entity.Document, error) {
				return doc, nil
			},
		}
		aiParser := &mockDocumentAIParserDocSvc{
			classifyAndExtractFunc: func(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, fileBytes []byte) (entity.DocumentType, map[string]interface{}, string, error) {
				return entity.DocTypeReceipt, map[string]interface{}{"vendor": "Amazon"}, "Extracted Text", nil
			},
		}
		svc := service.NewDocumentService(repo, aiParser, &mockPayslipRepoDocSvc{}, &mockInvoiceRepoDocSvc{})

		req := entity.DocumentUploadRequest{
			UserID:      userID,
			FileContent: []byte("test content"),
			FileName:    "receipt.pdf",
			ContentType: "application/pdf",
			SkipAI:      false,
		}

		doc, err := svc.Upload(ctx, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if doc.Type != entity.DocTypeReceipt {
			t.Errorf("expected type %s, got %s", entity.DocTypeReceipt, doc.Type)
		}

		if doc.Metadata["vendor"] != "Amazon" {
			t.Errorf("expected metadata vendor Amazon, got %v", doc.Metadata["vendor"])
		}

		if doc.ExtractedText != "Extracted Text" {
			t.Errorf("expected extracted text, got %s", doc.ExtractedText)
		}
	})

	t.Run("Duplicate Upload Rejection", func(t *testing.T) {
		repo := &mockDocumentRepoDocSvc{
			existsByHashFunc: func(ctx context.Context, userID uuid.UUID, contentHash string) (bool, error) {
				return true, nil
			},
		}
		svc := service.NewDocumentService(repo, &mockDocumentAIParserDocSvc{}, &mockPayslipRepoDocSvc{}, &mockInvoiceRepoDocSvc{})

		req := entity.DocumentUploadRequest{
			UserID:      userID,
			FileContent: []byte("test content"),
		}

		_, err := svc.Upload(ctx, req)
		if err != entity.ErrDocumentDuplicate {
			t.Fatalf("expected ErrDocumentDuplicate, got %v", err)
		}
	})

	t.Run("Upload with AI Failure (Fallback to Other)", func(t *testing.T) {
		repo := &mockDocumentRepoDocSvc{
			existsByHashFunc: func(ctx context.Context, userID uuid.UUID, contentHash string) (bool, error) {
				return false, nil
			},
			saveFunc: func(ctx context.Context, doc entity.Document) (entity.Document, error) {
				return doc, nil
			},
		}
		aiParser := &mockDocumentAIParserDocSvc{
			classifyAndExtractFunc: func(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, fileBytes []byte) (entity.DocumentType, map[string]interface{}, string, error) {
				return "", nil, "", context.DeadlineExceeded
			},
		}
		svc := service.NewDocumentService(repo, aiParser, &mockPayslipRepoDocSvc{}, &mockInvoiceRepoDocSvc{})

		req := entity.DocumentUploadRequest{
			UserID:      userID,
			FileContent: []byte("test content"),
			FileName:    "receipt.pdf",
			ContentType: "application/pdf",
			SkipAI:      false,
		}

		doc, err := svc.Upload(ctx, req)
		if err != nil {
			t.Fatalf("expected no error (fallback), got %v", err)
		}

		if doc.Type != entity.DocTypeOther {
			t.Errorf("expected type %s on AI failure, got %s", entity.DocTypeOther, doc.Type)
		}
		if doc.Metadata == nil {
			t.Error("expected non-nil metadata even on AI failure")
		}
	})
}

func TestDocumentService_Update(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	ctx := context.Background()

	t.Run("Update multiple fields including nil metadata init", func(t *testing.T) {
		initialDoc := entity.Document{
			ID:               docID,
			UserID:           userID,
			Type:             entity.DocTypeOther,
			OriginalFileName: "old.pdf",
			Metadata:         nil, // Test nil metadata handling
		}

		repo := &mockDocumentRepoDocSvc{
			findByIDFunc: func(ctx context.Context, id, uID uuid.UUID) (entity.Document, error) {
				return initialDoc, nil
			},
			updateFunc: func(ctx context.Context, doc entity.Document) (entity.Document, error) {
				return doc, nil
			},
		}
		svc := service.NewDocumentService(repo, &mockDocumentAIParserDocSvc{}, &mockPayslipRepoDocSvc{}, &mockInvoiceRepoDocSvc{})

		newName := "new.pdf"
		newType := entity.DocTypeContract
		newDate := time.Date(2025, 5, 20, 0, 0, 0, 0, time.UTC)

		req := entity.DocumentUpdateRequest{
			OriginalFileName: &newName,
			Type:             &newType,
			DocumentDate:     &newDate,
			Metadata:         map[string]interface{}{"new_key": "new_val"},
		}

		updatedDoc, err := svc.Update(ctx, docID, userID, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if updatedDoc.OriginalFileName != newName {
			t.Errorf("expected name %s, got %s", newName, updatedDoc.OriginalFileName)
		}
		if updatedDoc.Metadata["date"] != "2025-05-20" {
			t.Errorf("expected metadata date 2025-05-20, got %v", updatedDoc.Metadata["date"])
		}
		if updatedDoc.Metadata["new_key"] != "new_val" {
			t.Errorf("expected new_key to be added, got %v", updatedDoc.Metadata["new_key"])
		}
	})

	t.Run("Update with Re-upload (Trigger AI)", func(t *testing.T) {
		initialDoc := entity.Document{
			ID:               docID,
			UserID:           userID,
			Type:             entity.DocTypeOther,
			OriginalFileName: "old.pdf",
			ContentHash:      "old-hash",
			MimeType:         "application/pdf",
		}

		repo := &mockDocumentRepoDocSvc{
			findByIDFunc: func(ctx context.Context, id, uID uuid.UUID) (entity.Document, error) {
				return initialDoc, nil
			},
			updateFunc: func(ctx context.Context, doc entity.Document) (entity.Document, error) {
				return doc, nil
			},
		}

		aiParser := &mockDocumentAIParserDocSvc{
			classifyAndExtractFunc: func(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, fileBytes []byte) (entity.DocumentType, map[string]interface{}, string, error) {
				return entity.DocTypeReceipt, map[string]interface{}{"vendor": "Re-uploaded Inc"}, "Fresh AI Text", nil
			},
		}

		svc := service.NewDocumentService(repo, aiParser, &mockPayslipRepoDocSvc{}, &mockInvoiceRepoDocSvc{})

		newContent := []byte("new file content")
		newMime := "application/pdf"
		req := entity.DocumentUpdateRequest{
			OriginalFileContent: newContent,
			MimeType:            &newMime,
		}

		updatedDoc, err := svc.Update(ctx, docID, userID, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify content hash changed
		if updatedDoc.ContentHash == "old-hash" {
			t.Error("expected content hash to change after re-upload")
		}

		// Verify AI was triggered
		if updatedDoc.Type != entity.DocTypeReceipt {
			t.Errorf("expected re-classification to %s, got %s", entity.DocTypeReceipt, updatedDoc.Type)
		}
		if updatedDoc.Metadata["vendor"] != "Re-uploaded Inc" {
			t.Errorf("expected updated metadata, got %v", updatedDoc.Metadata["vendor"])
		}
		if updatedDoc.ExtractedText != "Fresh AI Text" {
			t.Errorf("expected updated extracted text, got %s", updatedDoc.ExtractedText)
		}
	})
}

func TestDocumentService_SimpleWrappers(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	ctx := context.Background()

	repo := &mockDocumentRepoDocSvc{
		findAllFunc: func(ctx context.Context, filter entity.DocumentFilter) ([]entity.Document, error) {
			return []entity.Document{{ID: docID, UserID: userID}}, nil
		},
		findByIDFunc: func(ctx context.Context, id, uID uuid.UUID) (entity.Document, error) {
			if id == uuid.Nil {
				return entity.Document{}, entity.ErrDocumentNotFound
			}
			return entity.Document{ID: id, UserID: uID, OriginalFileContent: []byte("content"), MimeType: "application/pdf", OriginalFileName: "test.pdf"}, nil
		},
		deleteFunc: func(ctx context.Context, id, uID uuid.UUID) error {
			return nil
		},
	}
	svc := service.NewDocumentService(repo, &mockDocumentAIParserDocSvc{}, &mockPayslipRepoDocSvc{}, &mockInvoiceRepoDocSvc{})

	t.Run("List", func(t *testing.T) {
		res, err := svc.List(ctx, entity.DocumentFilter{UserID: userID})
		if err != nil || len(res) != 1 {
			t.Errorf("List failed: %v, len: %d", err, len(res))
		}
	})

	t.Run("GetDetail", func(t *testing.T) {
		res, err := svc.GetDetail(ctx, docID, userID)
		if err != nil || res.ID != docID {
			t.Errorf("GetDetail failed: %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := svc.Delete(ctx, docID, userID)
		if err != nil {
			t.Errorf("Delete failed: %v", err)
		}
	})

	t.Run("Download Success", func(t *testing.T) {
		content, mime, name, err := svc.Download(ctx, docID, userID)
		if err != nil || string(content) != "content" || mime != "application/pdf" || name != "test.pdf" {
			t.Errorf("Download failed: %v", err)
		}
	})

	t.Run("Download Not Found", func(t *testing.T) {
		_, _, _, err := svc.Download(ctx, uuid.Nil, userID)
		if err == nil {
			t.Error("expected error for non-existent document")
		}
	})
}

func TestDocumentService_GetTaxYearSummary(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()
	year := 2024

	repo := &mockDocumentRepoDocSvc{
		findAllFunc: func(ctx context.Context, filter entity.DocumentFilter) ([]entity.Document, error) {
			return []entity.Document{
				{
					Type:      entity.DocTypeTaxCertificate,
					Metadata:  map[string]interface{}{"year": float64(year), "gross_income": 50000.0, "net_income": 35000.0, "income_tax": 15000.0},
					CreatedAt: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					Type:      entity.DocTypeReceipt,
					Metadata:  map[string]interface{}{"year": year, "tax_deductible": true, "amount": 200.0}, // Testing int year
					CreatedAt: time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					Type:      entity.DocTypeOther,
					Metadata:  nil, // Testing CreatedAt fallback
					CreatedAt: time.Date(year, 5, 10, 0, 0, 0, 0, time.UTC),
				},
				{
					Type:      entity.DocTypeOther,
					Metadata:  map[string]interface{}{"year": float64(year - 1)}, // Different year
					CreatedAt: time.Date(year-1, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			}, nil
		},
	}

	pRepo := &mockPayslipRepoDocSvc{
		findAllFunc: func(ctx context.Context, filter entity.PayslipFilter) ([]entity.Payslip, error) {
			if filter.Year == year {
				return []entity.Payslip{
					{GrossPay: 3000.0, NetPay: 2000.0},
				}, nil
			}
			return nil, nil
		},
	}

	iRepo := &mockInvoiceRepoDocSvc{
		findAllFunc: func(ctx context.Context, filter entity.InvoiceFilter) ([]entity.Invoice, error) {
			if filter.Year == year {
				return []entity.Invoice{
					{Amount: 150.0},
				}, nil
			}
			return nil, nil
		},
	}

	svc := service.NewDocumentService(repo, &mockDocumentAIParserDocSvc{}, pRepo, iRepo)

	summary, err := svc.GetTaxYearSummary(ctx, userID, year)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if summary.Year != year {
		t.Errorf("expected year %d, got %d", year, summary.Year)
	}

	// 50000 (Tax Cert) + 3000 (Payslip) = 53000
	if summary.TotalGrossIncome != 53000.0 {
		t.Errorf("expected gross income 53000, got %f", summary.TotalGrossIncome)
	}

	// 200 (Deductible Receipt) + 150 (Invoice) = 350
	if summary.TotalDeductible != 350.0 {
		t.Errorf("expected deductible 350, got %f", summary.TotalDeductible)
	}

	if len(summary.Documents) != 3 {
		t.Errorf("expected 3 documents in summary, got %d", len(summary.Documents))
	}
}
