package entity

import (
	"time"

	"github.com/google/uuid"
)

// Invoice is the domain entity produced by the categorization engine.
// It records the vendor, amount, and the category assigned to it by the LLM.
type Invoice struct {
	ID          uuid.UUID      `json:"id"`
	UserID      uuid.UUID      `json:"user_id"`
	Vendor      Vendor         `json:"vendor"`
	CategoryID  *uuid.UUID     `json:"category_id"`
	Amount      float64        `json:"amount"`
	Currency    string         `json:"currency"`
	BaseAmount  float64        `json:"base_amount"`   // Snapshotted converted amount
	BaseCurrency string        `json:"base_currency"` // Snapshotted base currency
	IssuedAt    time.Time      `json:"issued_at"`
	Description string         `json:"description"`
	Splits      []InvoiceSplit `json:"splits,omitempty"`

	// Deduplication & file storage — populated on file-upload imports (migration 002).
	ContentHash         string `json:"content_hash,omitempty"`
	OriginalFileName    string `json:"original_file_name,omitempty"`
	OriginalFileContent []byte `json:"-"` // excluded from JSON responses

	// Sharing metadata (added for Collaborative Finance Expansion)
	IsShared   bool        `json:"is_shared"`
	SharedWith []uuid.UUID `json:"shared_with"`
	OwnerID    uuid.UUID   `json:"owner_id"`
}

// InvoiceSplit represents an individual line item or split of an invoice
// into a specific category with a sub-amount.
type InvoiceSplit struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	InvoiceID   uuid.UUID `json:"invoice_id"`
	CategoryID  uuid.UUID `json:"category_id"`
	Amount      float64   `json:"amount"`
	BaseAmount  float64   `json:"base_amount"` // Snapshotted converted amount
	Description string    `json:"description"`
}

// InvoiceFilter defines query parameters for listing invoices.
type InvoiceFilter struct {
	UserID        uuid.UUID
	IncludeShared bool   // Collaborative Finance: Include invoices shared with the user
	Source        string // "all", "mine", "shared"
	Year          int    // Optional: Filter by invoice date year
	Limit         int
	Offset        int
}
