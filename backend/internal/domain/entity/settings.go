package entity

import "github.com/google/uuid"

// Setting represents a key-value configuration pair stored in the database.
type Setting struct {
	UserID      uuid.UUID `json:"user_id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	IsSensitive bool      `json:"is_sensitive"`
}
