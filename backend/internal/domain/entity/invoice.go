package entity

import (
	"time"

	"github.com/google/uuid"
)

// Invoice is the domain entity produced by the categorization engine.
// It records the vendor, amount, and the category assigned to it by the LLM.
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

	// Deduplication & file storage — populated on file-upload imports (migration 002).
	ContentHash         string `json:"content_hash,omitempty"`
	OriginalFileName    string `json:"original_file_name,omitempty"`
	OriginalFileMime    string `json:"original_file_mime,omitempty"`
	OriginalFileSize    int64  `json:"original_file_size,omitempty"`
	OriginalFileContent []byte `json:"-"` // excluded from JSON responses
}
