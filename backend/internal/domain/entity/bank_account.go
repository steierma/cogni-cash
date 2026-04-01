package entity

import (
	"time"

	"github.com/google/uuid"
)

type BankAccount struct {
	ID                 uuid.UUID     `json:"id"`
	ConnectionID       uuid.UUID     `json:"connection_id"`
	ProviderAccountID  string        `json:"provider_account_id"`
	IBAN               string        `json:"iban"`
	Name               string        `json:"name"`
	Currency           string        `json:"currency"`
	Balance            float64       `json:"balance"`
	LastSyncedAt       time.Time     `json:"last_synced_at"`
	LastSyncError      *string       `json:"last_sync_error,omitempty"`
	AccountType        StatementType `json:"account_type"`
}
