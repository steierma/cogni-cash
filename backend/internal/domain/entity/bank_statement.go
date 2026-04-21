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
	ID                  uuid.UUID       `json:"id"`
	UserID              uuid.UUID       `json:"user_id"`
	BankStatementID     *uuid.UUID      `json:"bank_statement_id,omitempty"`
	BankAccountID       *uuid.UUID      `json:"bank_account_id,omitempty"`
	BookingDate         time.Time       `json:"booking_date"`
	ValutaDate          time.Time       `json:"valuta_date"`
	Description         string          `json:"description"`
	CounterpartyName    string          `json:"counterparty_name,omitempty"`
	CounterpartyIban    string          `json:"counterparty_iban,omitempty"`
	BankTransactionCode string          `json:"bank_transaction_code,omitempty"`
	MandateReference    string          `json:"mandate_reference,omitempty"`
	Location            string          `json:"location,omitempty"` // <-- NEW Optional Field
	Amount              float64         `json:"amount"`
	Currency            string          `json:"currency"`
	BaseAmount          float64         `json:"base_amount"`   // Snapshotted converted amount
	BaseCurrency        string          `json:"base_currency"` // Snapshotted base currency
	Type                TransactionType `json:"type"`
	Reference           string          `json:"reference"`
	CategoryID          *uuid.UUID      `json:"category_id"`
	ContentHash         string          `json:"content_hash"`
	IsReconciled        bool            `json:"is_reconciled"`
	ReconciliationID    *uuid.UUID      `json:"reconciliation_id,omitempty"`
	Reviewed            bool            `json:"reviewed"`
	StatementType       StatementType   `json:"statement_type"`
	IsPrediction        bool            `json:"is_prediction"`       // <-- NEW Field
	SkipForecasting     bool            `json:"skip_forecasting"`    // <-- NEW Field: Exclude historical pattern
	IsPayslipVerified   bool            `json:"is_payslip_verified"` // <-- NEW Field: Verified against payslip
	IsShared            bool            `json:"is_shared"`           // Collaborative Finance: Is category shared
	OwnerID             *uuid.UUID      `json:"owner_id,omitempty"`  // Collaborative Finance: Original owner of the transaction
	SubscriptionID      *uuid.UUID      `json:"subscription_id,omitempty"`
}

// BankStatement is the top-level entity representing one parsed Kontoauszug.
type BankStatement struct {
	ID                  uuid.UUID     `json:"id"`
	UserID              uuid.UUID     `json:"user_id"`
	AccountHolder       string        `json:"account_holder"`
	IBAN                string        `json:"iban"`
	StatementDate       time.Time     `json:"statement_date"`
	StatementNo         int           `json:"statement_no"`
	OldBalance          float64       `json:"old_balance"`
	NewBalance          float64       `json:"new_balance"`
	Currency            string        `json:"currency"`
	StatementType       StatementType `json:"statement_type"` // "giro" | "credit_card" — set by the parser
	Transactions        []Transaction `json:"transactions"`
	SkippedTransactions []Transaction `json:"skipped_transactions,omitempty"` // Used to store skipped duplicates for a potential future cleanup table
	OriginalFile        []byte        `json:"-"`                              // Exclude from JSON payload
	ImportedAt          time.Time     `json:"imported_at"`                    // set by the DB on insert
	ContentHash         string        `json:"content_hash"`                   // SHA-256 over stable statement fields — used to prevent duplicate imports
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
	HasOriginalFile  bool          `json:"has_original_file"`
}

// TransactionFilter holds the query parameters to search and filter transactions server-side.
type TransactionFilter struct {
	UserID             uuid.UUID
	StatementID        *uuid.UUID
	CategoryID         *uuid.UUID // Replaces Category string
	Type               string     // "credit", "debit", "all"
	Search             string
	FromDate           *time.Time
	ToDate             *time.Time
	MinAmount          *float64
	MaxAmount          *float64
	IsReconciled       *bool          // Filter for reconciliation status
	Reviewed           *bool          // Filter for reviewed status
	StatementType      *StatementType // Added to support filtering by statement type (e.g., extra_account)
	IncludePredictions bool           // <-- NEW Field
	IncludeShared      bool           // Collaborative Finance: Include transactions from shared categories
	SubscriptionID     *uuid.UUID     // Filter by linked subscription
	Limit              int            // Pagination limit
	Offset             int            // Pagination offset
}

// CategorizationExample represents a historical categorization for few-shot learning.
type CategorizationExample struct {
	Description         string `json:"description"`
	Reference           string `json:"reference"`
	CounterpartyName    string `json:"counterparty_name,omitempty"`
	CounterpartyIban    string `json:"counterparty_iban,omitempty"`
	BankTransactionCode string `json:"bank_transaction_code,omitempty"`
	MandateReference    string `json:"mandate_reference,omitempty"`
	Category            string `json:"category"`
}

// ---- Analytics DTOs for the Frontend ----

// CategoryTotal represents aggregated category data for charts (e.g., Pie/Doughnut charts).
type CategoryTotal struct {
	CategoryID string  `json:"category_id"`
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"` // Absolute amount
	Type       string  `json:"type"`   // "income" or "expense"
	Color      string  `json:"color"`  // Hex color fetched from category repository
}

// TimeSeriesPoint represents aggregated data for a specific time period (e.g., Line/Bar charts).
type TimeSeriesPoint struct {
	Date            string             `json:"date"`             // Format dynamically adjusted (YYYY-MM-DD, YYYY-MM, or YYYY)
	Income          float64            `json:"income"`           // Total income for the period
	Expense         float64            `json:"expense"`          // Total expense for the period
	CategoryAmounts map[string]float64 `json:"category_amounts"` // Map of CategoryID -> Amount for this period
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
