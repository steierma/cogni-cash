package port

import (
	"cogni-cash/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

type InvoiceRepository interface {
	Save(ctx context.Context, invoice entity.Invoice) error
	FindByID(ctx context.Context, id uuid.UUID) (entity.Invoice, error)
	FindAll(ctx context.Context) ([]entity.Invoice, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
