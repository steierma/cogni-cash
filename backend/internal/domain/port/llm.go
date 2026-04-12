package port

import (
	"cogni-cash/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

// --- Domain-Specific AI Parsers ---

// BankStatementAIParser handles AI extraction exclusively for Bank Statements.
type BankStatementAIParser interface {
	ParseBankStatement(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, fileBytes []byte) (entity.BankStatement, error)
}

// PayslipAIParser handles AI extraction exclusively for Payslips.
type PayslipAIParser interface {
	ParsePayslip(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, fileBytes []byte) (entity.Payslip, error)
}

// --- Invoice & Transaction Interfaces ---

type CategorizationRequest struct {
	// RawText is the extracted text from a document.
	RawText string
	// Categories is the list of valid category names the LLM may choose from.
	Categories []string
}

type InvoiceCategorizationResult struct {
	InvoiceName string
	VendorName  string
	Amount      float64
	Currency    string
	InvoiceDate string
	Description string
}

type InvoiceAICategorizer interface {
	CategorizeInvoice(ctx context.Context, userID uuid.UUID, req CategorizationRequest) (InvoiceCategorizationResult, error)
	// CategorizeInvoiceImage sends raw image bytes directly to the LLM.
	CategorizeInvoiceImage(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, imageBytes []byte, categories []string) (InvoiceCategorizationResult, error)
}

// TransactionToCategorize holds the data needed to categorize a single transaction.
type TransactionToCategorize struct {
	Hash                string `json:"hash"`
	Description         string `json:"description"`
	Reference           string `json:"reference"`
	CounterpartyName    string `json:"counterparty_name,omitempty"`
	CounterpartyIban    string `json:"counterparty_iban,omitempty"`
	BankTransactionCode string `json:"bank_transaction_code,omitempty"`
	MandateReference    string `json:"mandate_reference,omitempty"`
}

type CategorizedTransaction struct {
	Hash     string `json:"hash"`
	Category string `json:"category"`
}

type TransactionCategorizer interface {
	CategorizeTransactionsBatch(ctx context.Context, userID uuid.UUID, txns []TransactionToCategorize, categories []string, examples []entity.CategorizationExample) ([]CategorizedTransaction, error)
}
