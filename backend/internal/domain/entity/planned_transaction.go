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

// PlannedTransaction represents a user-defined future transaction for forecasting.
type PlannedTransaction struct {
	ID                   uuid.UUID                `json:"id"`
	UserID               uuid.UUID                `json:"user_id"`
	Amount               float64                  `json:"amount"`
	Date                 time.Time                `json:"date"`
	Description          string                   `json:"description"`
	CategoryID           *uuid.UUID               `json:"category_id"`
	Status               PlannedTransactionStatus `json:"status"`
	MatchedTransactionID *uuid.UUID               `json:"matched_transaction_id"`
	IntervalMonths       int                      `json:"interval_months"`
	EndDate              *time.Time               `json:"end_date"`
	IsSuperseded         bool                     `json:"is_superseded"`
	CreatedAt            time.Time                `json:"created_at"`
}
