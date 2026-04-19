package entity

import "github.com/google/uuid"

// Vendor represents the issuer of an invoice.
type Vendor struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
