package entity

import (
	"time"

	"github.com/google/uuid"
)

// Invoice is the domain entity produced by the categorization engine.
// It records the vendor, amount, and the category assigned to it by the LLM.
type Invoice struct {
	ID                 uuid.UUID  `json:"id"`
	UserID             uuid.UUID  `json:"user_id"`
	Vendor             Vendor     `json:"vendor"`
	CategoryID         *uuid.UUID `json:"category_id"`
	Amount             float64    `json:"amount"`
	Currency           string     `json:"currency"`
	IssuedAt           time.Time  `json:"issued_at"`
	Description        string     `json:"description"`

	// Deduplication & file storage — populated on file-upload imports (migration 002).
	ContentHash         string `json:"content_hash,omitempty"`
	OriginalFileName    string `json:"original_file_name,omitempty"`
	OriginalFileContent []byte `json:"-"` // excluded from JSON responses
}
