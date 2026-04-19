package entity

import (
	"time"

	"github.com/google/uuid"
)

type ConnectionStatus string

const (
	StatusInitialized ConnectionStatus = "initialized"
	StatusLinked      ConnectionStatus = "linked"
	StatusExpired     ConnectionStatus = "expired"
	StatusFailed      ConnectionStatus = "failed"
)

type BankConnection struct {
	ID              uuid.UUID        `json:"id"`
	UserID          uuid.UUID        `json:"user_id"`
	Provider        string           `json:"provider"` // 'enablebanking'
	InstitutionID   string           `json:"institution_id"`
	InstitutionName string           `json:"institution_name"`
	RequisitionID   string           `json:"requisition_id"` // Provider session ID
	ReferenceID     string           `json:"reference_id"`   // Our internal match ID
	Status          ConnectionStatus `json:"status"`
	AuthLink        string           `json:"auth_link,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	ExpiresAt       *time.Time       `json:"expires_at,omitempty"`
	Accounts        []BankAccount    `json:"accounts,omitempty"`
}

type BankInstitution struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Bic     string `json:"bic"`
	Logo    string `json:"logo,omitempty"`
	Country string `json:"country"`
}
