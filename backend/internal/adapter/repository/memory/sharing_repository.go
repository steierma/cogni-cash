package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type sharedItem struct {
	ItemID          uuid.UUID
	OwnerID         uuid.UUID
	SharedWithID    uuid.UUID
	PermissionLevel string
}

type SharingRepository struct {
	mu             sync.RWMutex
	categoryShares map[string]sharedItem
	invoiceShares  map[string]sharedItem
}

func NewSharingRepository() *SharingRepository {
	return &SharingRepository{
		categoryShares: make(map[string]sharedItem),
		invoiceShares:  make(map[string]sharedItem),
	}
}

func (r *SharingRepository) ShareCategory(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID, permission string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := fmt.Sprintf("%s:%s:%s", categoryID, ownerID, sharedWithID)
	r.categoryShares[key] = sharedItem{
		ItemID:          categoryID,
		OwnerID:         ownerID,
		SharedWithID:    sharedWithID,
		PermissionLevel: permission,
	}
	return nil
}

func (r *SharingRepository) RevokeShare(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := fmt.Sprintf("%s:%s:%s", categoryID, ownerID, sharedWithID)
	if _, ok := r.categoryShares[key]; !ok {
		return fmt.Errorf("share not found")
	}
	delete(r.categoryShares, key)
	return nil
}

func (r *SharingRepository) ListShares(ctx context.Context, categoryID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var userIDs []uuid.UUID
	for _, share := range r.categoryShares {
		if share.ItemID == categoryID && share.OwnerID == ownerID {
			userIDs = append(userIDs, share.SharedWithID)
		}
	}
	return userIDs, nil
}

func (r *SharingRepository) ShareInvoice(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID, permission string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := fmt.Sprintf("%s:%s:%s", invoiceID, ownerID, sharedWithID)
	r.invoiceShares[key] = sharedItem{
		ItemID:          invoiceID,
		OwnerID:         ownerID,
		SharedWithID:    sharedWithID,
		PermissionLevel: permission,
	}
	return nil
}

func (r *SharingRepository) RevokeInvoiceShare(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := fmt.Sprintf("%s:%s:%s", invoiceID, ownerID, sharedWithID)
	if _, ok := r.invoiceShares[key]; !ok {
		return fmt.Errorf("invoice share not found")
	}
	delete(r.invoiceShares, key)
	return nil
}

func (r *SharingRepository) ListInvoiceShares(ctx context.Context, invoiceID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var userIDs []uuid.UUID
	for _, share := range r.invoiceShares {
		if share.ItemID == invoiceID && share.OwnerID == ownerID {
			userIDs = append(userIDs, share.SharedWithID)
		}
	}
	return userIDs, nil
}
