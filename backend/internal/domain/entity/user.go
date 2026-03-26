package entity

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Prevents accidental exposure in API responses
	Email        string    `json:"email"`
	FullName     string    `json:"full_name"`
	Address      string    `json:"address"`
	Role         string    `json:"role"`
}
