package entity

import (
	"time"

	"github.com/google/uuid"
)

// PlannedTransactionStatus defines the current state of a manual forecast.
type PlannedTransactionStatus string

const (
	PlannedTransactionStatusPending PlannedTransactionStatus = "pending"
	PlannedTransactionStatusMatched PlannedTransactionStatus = "matched"
	PlannedTransactionStatusExpired PlannedTransactionStatus = "expired"
)

// SchedulingStrategy defines how future occurrences are calculated.
type SchedulingStrategy string

const (
	SchedulingStrategyFixedDay    SchedulingStrategy = "fixed_day"
	SchedulingStrategyLastBankDay SchedulingStrategy = "last_bank_day"
)

// PlannedTransaction represents a user-defined future transaction for forecasting.
type PlannedTransaction struct {
	ID                   uuid.UUID                `json:"id"`
	UserID               uuid.UUID                `json:"user_id"`
	Amount               float64                  `json:"amount"`
	Currency             string                   `json:"currency"`
	BaseAmount           float64                  `json:"base_amount"`   // Snapshotted converted amount
	BaseCurrency         string                   `json:"base_currency"` // Snapshotted base currency
	Date                 time.Time                `json:"date"`
	Description          string                   `json:"description"`
	CategoryID           *uuid.UUID               `json:"category_id"`
	Status               PlannedTransactionStatus `json:"status"`
	MatchedTransactionID *uuid.UUID               `json:"matched_transaction_id"`
	IntervalMonths       int                      `json:"interval_months"`
	SchedulingStrategy   SchedulingStrategy       `json:"scheduling_strategy"`
	EndDate              *time.Time               `json:"end_date"`
	BankAccountID        *uuid.UUID               `json:"bank_account_id,omitempty"`
	IsShared             bool                     `json:"is_shared"`
	IsSuperseded         bool                     `json:"is_superseded"`
	OwnerID              uuid.UUID                `json:"owner_id"`
	CreatedAt            time.Time                `json:"created_at"`
}
