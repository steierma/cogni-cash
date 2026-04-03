package entity

import (
	"time"

	"github.com/google/uuid"
)

// Payslip represents a parsed monthly salary statement (generalized for international use).
type Payslip struct {
	ID         string    `json:"id"`
	UserID     uuid.UUID `json:"user_id"`

	OriginalFileName    string `json:"original_file_name,omitempty"`
	OriginalFileContent []byte `json:"-"`
	ContentHash         string `json:"content_hash,omitempty"`

	PeriodMonthNum int    `json:"period_month_num"`
	PeriodYear     int    `json:"period_year"`
	EmployerName   string `json:"employer_name"`
	TaxClass       string `json:"tax_class,omitempty"`
	TaxID          string `json:"tax_id,omitempty"`

	// Internationalized financial fields
	GrossPay         float64 `json:"gross_pay"`
	NetPay           float64 `json:"net_pay"`
	PayoutAmount     float64 `json:"payout_amount"`
	CustomDeductions float64 `json:"custom_deductions"`

	Bonuses   []Bonus   `json:"bonuses"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// Bonus replaces Sonderzahlung for international compatibility.
type Bonus struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

// PayslipFilter defines optional filters for listing payslips.
type PayslipFilter struct {
	UserID   uuid.UUID
	Employer string // Optional: Filter by employer name
}
