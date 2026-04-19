package entity

import (
	"time"

	"github.com/google/uuid"
)

// BridgeAccessToken is a specialized, long-lived credential generated in the Web UI
// to allow the Hermit app to sync without needing to manually extract JWT tokens.
type BridgeAccessToken struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	Name       string     `json:"name"`
	TokenHash  string     `json:"-"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// CreateBridgeTokenResponse contains the plaintext token and its metadata.
type CreateBridgeTokenResponse struct {
	Token string            `json:"token"`
	Info  BridgeAccessToken `json:"info"`
}
