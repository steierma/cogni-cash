package entity

import (
	"time"

	"github.com/google/uuid"
)

type BankAccount struct {
	ID                uuid.UUID     `json:"id"`
	UserID            uuid.UUID     `json:"user_id"`
	ConnectionID      *uuid.UUID    `json:"connection_id,omitempty"`
	ProviderAccountID string        `json:"provider_account_id"`
	IBAN              string        `json:"iban"`
	Name              string        `json:"name"`
	Currency          string        `json:"currency"`
	Balance           float64       `json:"balance"`
	LastSyncedAt      time.Time     `json:"last_synced_at"`
	LastSyncError     *string       `json:"last_sync_error,omitempty"`
	AccountType       StatementType `json:"account_type"`
	IsShared          bool          `json:"is_shared"`
	SharedWith        []uuid.UUID   `json:"shared_with,omitempty"`
	OwnerID           uuid.UUID     `json:"owner_id"`
}
