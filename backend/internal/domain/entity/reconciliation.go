package entity

import (
	"time"

	"github.com/google/uuid"
)

type Reconciliation struct {
	ID                               uuid.UUID `json:"id"`
	UserID                           uuid.UUID `json:"user_id"`
	SettlementTransactionHash        string    `json:"settlement_transaction_hash"`
	TargetTransactionHash            string    `json:"target_transaction_hash"`
	SettlementTransactionDescription string    `json:"settlement_transaction_description,omitempty"`
	TargetTransactionDescription     string    `json:"target_transaction_description,omitempty"`
	SettlementBookingDate            time.Time `json:"settlement_booking_date,omitempty"`
	TargetBookingDate                time.Time `json:"target_booking_date,omitempty"`
	SettlementStatementType          string    `json:"settlement_statement_type,omitempty"`
	TargetStatementType              string    `json:"target_statement_type,omitempty"`
	Amount                           float64   `json:"amount"`
	ReconciledAt                     time.Time `json:"reconciled_at"`
}
