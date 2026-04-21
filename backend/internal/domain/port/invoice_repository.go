package port

import (
	"cogni-cash/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

type InvoiceRepository interface {
	Save(ctx context.Context, invoice entity.Invoice) error
	Update(ctx context.Context, invoice entity.Invoice) error
	UpdateBaseAmount(ctx context.Context, id uuid.UUID, baseAmount float64, baseCurrency string, userID uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Invoice, error)
	FindAll(ctx context.Context, filter entity.InvoiceFilter) ([]entity.Invoice, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// Split management
	DeleteSplits(ctx context.Context, invoiceID, userID uuid.UUID) error

	// Deduplication
	ExistsByContentHash(ctx context.Context, hash string, userID uuid.UUID) (bool, error)

	// File download
	GetOriginalFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) (content []byte, mimeType string, fileName string, err error)
}
