package port

import (
	"context"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

// DocumentRepository defines the persistence port for generic documents in the vault.
type DocumentRepository interface {
	// Save persists a document to the repository.
	Save(ctx context.Context, doc entity.Document) (entity.Document, error)

	// FindAll returns a list of documents matching the filter.
	FindAll(ctx context.Context, filter entity.DocumentFilter) ([]entity.Document, error)

	// FindByID retrieves a single document by its ID and owner.
	FindByID(ctx context.Context, id, userID uuid.UUID) (entity.Document, error)

	// Update modifies an existing document in the repository.
	Update(ctx context.Context, doc entity.Document) (entity.Document, error)

	// Delete removes a document from the repository.
	Delete(ctx context.Context, id, userID uuid.UUID) error

	// ExistsByHash checks if a document with the same content hash exists for a user.
	ExistsByHash(ctx context.Context, userID uuid.UUID, contentHash string) (bool, error)
}
