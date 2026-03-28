package port

import (
	"cogni-cash/internal/domain/entity"
	"context"
)

// PayslipParser is the output port for parsing payslip documents.
// Each payslip format (CARIAD, AI fallback, …) provides its own adapter.
type PayslipParser interface {
	Parse(ctx context.Context, filePath string) (entity.Payslip, error)
}

