package port

import (
	"context"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (entity.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (entity.User, error)
	FindAll(ctx context.Context, search string) ([]entity.User, error)
	Create(ctx context.Context, user entity.User) error
	Update(ctx context.Context, user entity.User) error
	Upsert(ctx context.Context, user entity.User) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, newHash string) error
	Delete(ctx context.Context, id uuid.UUID) error
}
