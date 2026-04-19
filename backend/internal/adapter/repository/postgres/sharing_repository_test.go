package postgres

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogni-cash/internal/domain/port"
)

// -- In-Memory Implementation for Testing --

type shareRecord struct {
	OwnerID      uuid.UUID
	SharedWithID uuid.UUID
	Permission   string
}

// InMemorySharingRepository is a thread-safe test-double implementing port.SharingRepository
type InMemorySharingRepository struct {
	mu             sync.RWMutex
	categoryShares map[uuid.UUID]map[uuid.UUID]shareRecord // map[categoryID]map[sharedWithID]record
	invoiceShares  map[uuid.UUID]map[uuid.UUID]shareRecord // map[invoiceID]map[sharedWithID]record
}

func NewInMemorySharingRepository() *InMemorySharingRepository {
	return &InMemorySharingRepository{
		categoryShares: make(map[uuid.UUID]map[uuid.UUID]shareRecord),
		invoiceShares:  make(map[uuid.UUID]map[uuid.UUID]shareRecord),
	}
}

// -- Category Methods --

func (r *InMemorySharingRepository) ShareCategory(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID, permission string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.categoryShares[categoryID] == nil {
		r.categoryShares[categoryID] = make(map[uuid.UUID]shareRecord)
	}
	r.categoryShares[categoryID][sharedWithID] = shareRecord{
		OwnerID:      ownerID,
		SharedWithID: sharedWithID,
		Permission:   permission,
	}
	return nil
}

func (r *InMemorySharingRepository) RevokeShare(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	shares, exists := r.categoryShares[categoryID]
	if !exists {
		return nil // idempotent
	}

	record, ok := shares[sharedWithID]
	if ok && record.OwnerID != ownerID {
		return errors.New("unauthorized: only the owner can revoke")
	}

	delete(r.categoryShares[categoryID], sharedWithID)
	return nil
}

func (r *InMemorySharingRepository) ListShares(ctx context.Context, categoryID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shares, exists := r.categoryShares[categoryID]
	if !exists {
		return []uuid.UUID{}, nil
	}

	var sharedWith []uuid.UUID
	for targetID, record := range shares {
		if record.OwnerID != ownerID {
			return nil, errors.New("unauthorized: only the owner can list shares")
		}
		sharedWith = append(sharedWith, targetID)
	}
	return sharedWith, nil
}

// -- Invoice Methods --

func (r *InMemorySharingRepository) ShareInvoice(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID, permission string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.invoiceShares[invoiceID] == nil {
		r.invoiceShares[invoiceID] = make(map[uuid.UUID]shareRecord)
	}
	r.invoiceShares[invoiceID][sharedWithID] = shareRecord{
		OwnerID:      ownerID,
		SharedWithID: sharedWithID,
		Permission:   permission,
	}
	return nil
}

func (r *InMemorySharingRepository) RevokeInvoiceShare(ctx context.Context, invoiceID, ownerID, sharedWithID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	shares, exists := r.invoiceShares[invoiceID]
	if !exists {
		return nil // idempotent
	}

	record, ok := shares[sharedWithID]
	if ok && record.OwnerID != ownerID {
		return errors.New("unauthorized: only the owner can revoke")
	}

	delete(r.invoiceShares[invoiceID], sharedWithID)
	return nil
}

func (r *InMemorySharingRepository) ListInvoiceShares(ctx context.Context, invoiceID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shares, exists := r.invoiceShares[invoiceID]
	if !exists {
		return []uuid.UUID{}, nil
	}

	var sharedWith []uuid.UUID
	for targetID, record := range shares {
		if record.OwnerID != ownerID {
			return nil, errors.New("unauthorized: only the owner can list shares")
		}
		sharedWith = append(sharedWith, targetID)
	}
	return sharedWith, nil
}

// Interface compliance check
var _ port.SharingRepository = (*InMemorySharingRepository)(nil)

// -- Tests --

func TestInMemorySharingRepository_CategoryShares(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemorySharingRepository()

	catID := uuid.New()
	ownerID := uuid.New()
	friend1 := uuid.New()
	friend2 := uuid.New()
	hacker := uuid.New()

	t.Run("Share Category", func(t *testing.T) {
		err := repo.ShareCategory(ctx, catID, ownerID, friend1, "view")
		require.NoError(t, err)

		err = repo.ShareCategory(ctx, catID, ownerID, friend2, "edit")
		require.NoError(t, err)
	})

	t.Run("List Shares - Success", func(t *testing.T) {
		shares, err := repo.ListShares(ctx, catID, ownerID)
		require.NoError(t, err)
		assert.Len(t, shares, 2)
		assert.Contains(t, shares, friend1)
		assert.Contains(t, shares, friend2)
	})

	t.Run("List Shares - Unauthorized", func(t *testing.T) {
		shares, err := repo.ListShares(ctx, catID, hacker)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
		assert.Nil(t, shares)
	})

	t.Run("Revoke Share - Unauthorized", func(t *testing.T) {
		err := repo.RevokeShare(ctx, catID, hacker, friend1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("Revoke Share - Success", func(t *testing.T) {
		err := repo.RevokeShare(ctx, catID, ownerID, friend1)
		require.NoError(t, err)

		shares, _ := repo.ListShares(ctx, catID, ownerID)
		assert.Len(t, shares, 1)
		assert.Contains(t, shares, friend2)
		assert.NotContains(t, shares, friend1)
	})
}

func TestInMemorySharingRepository_InvoiceShares(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemorySharingRepository()

	invID := uuid.New()
	ownerID := uuid.New()
	accountant := uuid.New()

	t.Run("Share Invoice", func(t *testing.T) {
		err := repo.ShareInvoice(ctx, invID, ownerID, accountant, "view")
		require.NoError(t, err)
	})

	t.Run("List Invoice Shares", func(t *testing.T) {
		shares, err := repo.ListInvoiceShares(ctx, invID, ownerID)
		require.NoError(t, err)
		require.Len(t, shares, 1)
		assert.Equal(t, accountant, shares[0])
	})

	t.Run("Revoke Invoice Share", func(t *testing.T) {
		err := repo.RevokeInvoiceShare(ctx, invID, ownerID, accountant)
		require.NoError(t, err)

		shares, err := repo.ListInvoiceShares(ctx, invID, ownerID)
		require.NoError(t, err)
		assert.Len(t, shares, 0)
	})
}
