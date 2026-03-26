package port

import "context"

// CategorizationRequest is the input to the LLM categorization call.
type CategorizationRequest struct {
	// RawText is the extracted text from a document.
	RawText string
	// Categories is the list of valid category names the LLM may choose from.
	Categories []string
}

// CategorizationResult is the structured output from the LLM.
type CategorizationResult struct {
	CategoryName string
	VendorName   string
	Amount       float64
	Currency     string
	Description  string
}

// LLMClient is the output port for the AI categorization adapter.
type LLMClient interface {
	Categorize(ctx context.Context, req CategorizationRequest) (CategorizationResult, error)
}

// TransactionToCategorize holds the data needed to categorize a single transaction.
type TransactionToCategorize struct {
	Hash        string `json:"hash"`
	Description string `json:"description"`
	Reference   string `json:"reference"`
}

type CategorizedTransaction struct {
	Hash     string `json:"hash"`
	Category string `json:"category"`
}

// TransactionCategorizer handles batch LLM operations for transactions.
type TransactionCategorizer interface {
	CategorizeBatch(ctx context.Context, txns []TransactionToCategorize, categories []string) ([]CategorizedTransaction, error)
}
