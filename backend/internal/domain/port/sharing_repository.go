package port

import (
	"context"

	"github.com/google/uuid"
)

// SharingRepository handles the persistence of category sharing relationships.
type SharingRepository interface {
	// ShareCategory creates a sharing relationship.
	ShareCategory(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID, permission string) error
	// RevokeShare removes a sharing relationship.
	RevokeShare(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID) error
	// ListShares returns all user IDs a category is shared with.
	// Requires ownerID to ensure only the owner can list participants.
	ListShares(ctx context.Context, categoryID, ownerID uuid.UUID) ([]uuid.UUID, error)

	// ShareInvoice creates a sharing relationship for an invoice.
	ShareInvoice(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID, permission string) error
	// RevokeInvoiceShare removes a sharing relationship for an invoice.
	RevokeInvoiceShare(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID) error
	// ListInvoiceShares returns all user IDs an invoice is shared with.
	ListInvoiceShares(ctx context.Context, invoiceID, ownerID uuid.UUID) ([]uuid.UUID, error)
}
