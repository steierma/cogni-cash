package port

import (
	"cogni-cash/internal/domain/entity"
	"context"
)

type PayslipRepository interface {
	Save(ctx context.Context, payslip *entity.Payslip) error
	ExistsByHash(ctx context.Context, hash string) (bool, error)
	ExistsByOriginalFileName(ctx context.Context, originalFileName string) (bool, error)

	// Historical analytics & listing
	FindAll(ctx context.Context) ([]entity.Payslip, error)
	FindByID(ctx context.Context, id string) (entity.Payslip, error)

	// Editing and Deletion
	Update(ctx context.Context, payslip *entity.Payslip) error
	Delete(ctx context.Context, id string) error

	// File Download
	GetOriginalFile(ctx context.Context, id string) ([]byte, string, string, error)
}
