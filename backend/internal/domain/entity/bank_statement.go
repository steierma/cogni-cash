package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrDuplicate is returned when a statement or transaction has already been imported.
var ErrDuplicate = errors.New("statement already imported (duplicate)")

// TransactionType classifies whether money came in or went out.
type TransactionType string

const (
	TransactionTypeCredit TransactionType = "credit"
	TransactionTypeDebit  TransactionType = "debit"
)

// StatementType classifies the source of a bank statement for reconciliation purposes.
type StatementType string

const (
	StatementTypeGiro         StatementType = "giro"
	StatementTypeCreditCard   StatementType = "credit_card"
	StatementTypeExtraAccount StatementType = "extra_account"
)

// Transaction represents a single line-item on a bank statement.
type Transaction struct {
	BookingDate        time.Time       `json:"booking_date"`
	ValutaDate         time.Time       `json:"valuta_date"`
	Description        string          `json:"description"`
	Amount             float64         `json:"amount"`
	Currency           string          `json:"currency"`
	Type               TransactionType `json:"type"`
	Reference          string          `json:"reference"`
	CategoryID         *uuid.UUID      `json:"category_id"`
	ContentHash        string          `json:"content_hash"`
	IsReconciled       bool            `json:"is_reconciled"`
	ReconciliationID   *uuid.UUID      `json:"reconciliation_id,omitempty"`
	StatementType      StatementType   `json:"statement_type"`
	ExchangeRate       float64         `json:"exchange_rate"`        // e.g., 1.0 for base currency, 1.08 for EUR to USD
	AmountBaseCurrency float64         `json:"amount_base_currency"` // The calculated value used strictly for analytics
}

// BankStatement is the top-level entity representing one parsed Kontoauszug.
type BankStatement struct {
	ID                  uuid.UUID     `json:"id"`
	AccountHolder       string        `json:"account_holder"`
	IBAN                string        `json:"iban"`
	BIC                 string        `json:"bic"`
	AccountNumber       string        `json:"account_number"`
	StatementDate       time.Time     `json:"statement_date"`
	StatementNo         int           `json:"statement_no"`
	OldBalance          float64       `json:"old_balance"`
	NewBalance          float64       `json:"new_balance"`
	Currency            string        `json:"currency"`
	StatementType       StatementType `json:"statement_type"` // "giro" | "credit_card" — set by the parser
	Transactions        []Transaction `json:"transactions"`
	SkippedTransactions []Transaction `json:"skipped_transactions,omitempty"` // Used to store skipped duplicates for a potential future cleanup table
	SourceFile          string        `json:"source_file"`
	OriginalFile        []byte        `json:"-"`            // Exclude from JSON payload
	ImportedAt          time.Time     `json:"imported_at"`  // set by the DB on insert
	ContentHash         string        `json:"content_hash"` // SHA-256 over stable statement fields — used to prevent duplicate imports
}

func (b *BankStatement) IsValid() error {
	if b.IBAN == "" {
		return errors.New("invalid statement: missing IBAN")
	}
	if b.StatementDate.IsZero() {
		return errors.New("invalid statement: missing statement date")
	}
	if len(b.Transactions) == 0 {
		return errors.New("invalid statement: zero transactions extracted")
	}

	// Removed the calculatedSum == 0 check because zero-sum months are valid.

	return nil
}

// BankStatementSummary is a lightweight representation for list views.
type BankStatementSummary struct {
	ID               uuid.UUID     `json:"id"`
	StatementNo      int           `json:"statement_no"`
	PeriodLabel      string        `json:"period_label"`
	IBAN             string        `json:"iban"`
	Currency         string        `json:"currency"`
	NewBalance       float64       `json:"new_balance"`
	StartDate        time.Time     `json:"start_date"`
	EndDate          time.Time     `json:"end_date"`
	TransactionCount int           `json:"transaction_count"`
	StatementType    StatementType `json:"statement_type"`
}

// TransactionFilter holds the query parameters to search and filter transactions server-side.
type TransactionFilter struct {
	StatementID   *uuid.UUID
	CategoryID    *uuid.UUID // Replaces Category string
	Type          string     // "credit", "debit", "all"
	Search        string
	FromDate      *time.Time
	ToDate        *time.Time
	MinAmount     *float64
	MaxAmount     *float64
	IsReconciled  *bool          // Filter for reconciliation status
	StatementType *StatementType // Added to support filtering by statement type (e.g., extra_account)
}

// ---- Analytics DTOs for the Frontend ----

// CategoryTotal represents aggregated category data for charts (e.g., Pie/Doughnut charts).
type CategoryTotal struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"` // Absolute amount
	Type     string  `json:"type"`   // "income" or "expense"
	Color    string  `json:"color"`  // Hex color fetched from category repository
}

// TimeSeriesPoint represents aggregated data for a specific time period (e.g., Line/Bar charts).
type TimeSeriesPoint struct {
	Date    string  `json:"date"`    // Format dynamically adjusted (YYYY-MM-DD, YYYY-MM, or YYYY)
	Income  float64 `json:"income"`  // Total income for the period
	Expense float64 `json:"expense"` // Total expense for the period
}

// MerchantTotal represents an aggregated amount per payee/merchant.
type MerchantTotal struct {
	Merchant string  `json:"merchant"`
	Amount   float64 `json:"amount"` // Absolute amount
}

// TransactionAnalytics is the top-level DTO returned to the frontend for graphs.
type TransactionAnalytics struct {
	TotalIncome    float64           `json:"total_income"`
	TotalExpense   float64           `json:"total_expense"`
	NetSavings     float64           `json:"net_savings"`
	CategoryTotals []CategoryTotal   `json:"category_totals"`
	TimeSeries     []TimeSeriesPoint `json:"time_series"`
	TopMerchants   []MerchantTotal   `json:"top_merchants"`
}
