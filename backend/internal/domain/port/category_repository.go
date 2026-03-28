package port

import (
	"cogni-cash/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

// CategoryRepository is the output port for category persistence.
type CategoryRepository interface {
	Save(ctx context.Context, category entity.Category) (entity.Category, error)
	Update(ctx context.Context, category entity.Category) (entity.Category, error)
	FindByID(ctx context.Context, id uuid.UUID) (entity.Category, error)
	FindAll(ctx context.Context) ([]entity.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
