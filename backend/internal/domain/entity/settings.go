package entity

import "github.com/google/uuid"

// Setting represents a key-value configuration pair stored in the database.
type Setting struct {
	UserID      uuid.UUID `json:"user_id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	IsSensitive bool      `json:"is_sensitive"`
}

type LLMProfile struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"` // gemini, ollama, openai
	URL      string `json:"url"`
	Token    string `json:"token"`
	Model    string `json:"model"`
	IsActive bool   `json:"is_active"`
}
