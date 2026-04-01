package port

import (
	"cogni-cash/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

// PayslipParser is the output port for parsing payslip documents.
// Each payslip format (CARIAD, AI fallback, …) provides its own adapter.
type PayslipParser interface {
	Parse(ctx context.Context, userID uuid.UUID, filePath string) (entity.Payslip, error)
}

