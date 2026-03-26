package entity

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a named classification for transactions.
type Category struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`      // hex color, e.g. "#6366f1"
	CreatedAt time.Time `json:"created_at"` // set by the DB on insert
}

