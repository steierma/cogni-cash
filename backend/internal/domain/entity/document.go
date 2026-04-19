package entity

import (
	"time"

	"github.com/google/uuid"
)

type DocumentType string

const (
	DocTypeTaxCertificate DocumentType = "tax_certificate"
	DocTypeReceipt        DocumentType = "receipt"
	DocTypeContract       DocumentType = "contract"
	DocTypeOther          DocumentType = "other"
)

// Document is the generic domain entity for the Document Vault.
// It stores file metadata, OCR-extracted text, and flexible JSONB metadata.
type Document struct {
	ID                  uuid.UUID              `json:"id"`
	UserID              uuid.UUID              `json:"user_id"`
	Type                DocumentType           `json:"type"`
	OriginalFileName    string                 `json:"file_name"`
	ContentHash         string                 `json:"content_hash"`
	MimeType            string                 `json:"mime_type"`
	Metadata            map[string]interface{} `json:"metadata"` // JSONB in DB
	ExtractedText       string                 `json:"-"`        // OCR Output for Search
	OriginalFileContent []byte                 `json:"-"`        // Encrypted at rest
	CreatedAt           time.Time              `json:"created_at"`
}

// DocumentFilter defines query parameters for listing documents in the vault.
type DocumentFilter struct {
	UserID uuid.UUID
	Type   DocumentType
	Search string // Full-text search string for extracted_text and file_name
}

// DocumentUploadRequest defines the parameters for uploading a document to the vault.
type DocumentUploadRequest struct {
	UserID       uuid.UUID
	FileContent  []byte
	FileName     string
	ContentType  string
	SkipAI       bool
	ManualType   DocumentType
	DocumentDate time.Time
}

// DocumentUpdateRequest defines the mutable fields for a document.
type DocumentUpdateRequest struct {
	OriginalFileName *string                `json:"file_name"`
	Type             *DocumentType          `json:"type"`
	DocumentDate     *time.Time             `json:"document_date"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// TaxYearSummary provides an aggregated view of tax-relevant documents for a specific year.
type TaxYearSummary struct {
	Year             int        `json:"year"`
	Documents        []Document `json:"documents"`
	TotalGrossIncome float64    `json:"total_gross_income"`
	TotalNetIncome   float64    `json:"total_net_income"`
	TotalIncomeTax   float64    `json:"total_income_tax"`
	TotalDeductible  float64    `json:"total_deductible"`
}
