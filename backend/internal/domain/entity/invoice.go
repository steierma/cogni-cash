package entity

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a named classification for invoices and transactions.
type Category struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`      // hex color, e.g. "#6366f1"
	CreatedAt time.Time `json:"created_at"` // set by the DB on insert
}

// Vendor represents the issuer of an invoice.
type Vendor struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// Invoice is the central domain entity.
type Invoice struct {
	ID                 uuid.UUID  `json:"id"`
	Vendor             Vendor     `json:"vendor"`
	CategoryID         *uuid.UUID `json:"category_id"`
	Amount             float64    `json:"amount"`
	Currency           string     `json:"currency"`
	IssuedAt           time.Time  `json:"issued_at"`
	Description        string     `json:"description"`
	RawText            string     `json:"raw_text"`
	ExchangeRate       float64    `json:"exchange_rate"`        // e.g., 1.0 for base currency, 1.08 for EUR to USD
	AmountBaseCurrency float64    `json:"amount_base_currency"` // The calculated value used strictly for analytics
}
