package entity

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a named classification for transactions.
type Category struct {
	ID                 uuid.UUID `json:"id"`
	UserID             uuid.UUID `json:"user_id"`
	Name               string    `json:"name"`
	Color              string    `json:"color"`                // hex color, e.g. "#6366f1"
	IsVariableSpending bool      `json:"is_variable_spending"` // If true, forecast uses monthly burn rate
	CreatedAt          time.Time `json:"created_at"`           // set by the DB on insert
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

