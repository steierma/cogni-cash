package port

import (
	"cogni-cash/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

type PayslipRepository interface {
	Save(ctx context.Context, payslip *entity.Payslip) error
	ExistsByHash(ctx context.Context, hash string, userID uuid.UUID) (bool, error)
	ExistsByOriginalFileName(ctx context.Context, originalFileName string, userID uuid.UUID) (bool, error)

	// Historical analytics & listing
	FindAll(ctx context.Context, filter entity.PayslipFilter) ([]entity.Payslip, error)
	FindByID(ctx context.Context, id string, userID uuid.UUID) (entity.Payslip, error)

	// Editing and Deletion
	Update(ctx context.Context, payslip *entity.Payslip) error
	UpdateBaseAmount(ctx context.Context, id string, baseGross, baseNet, basePayout float64, currency string, userID uuid.UUID) error
	Delete(ctx context.Context, id string, userID uuid.UUID) error

	// File Download
	GetOriginalFile(ctx context.Context, id string, userID uuid.UUID) ([]byte, string, string, error)

	// Aggregations
	GetSummary(ctx context.Context, userID uuid.UUID) (entity.PayslipSummary, error)
}
