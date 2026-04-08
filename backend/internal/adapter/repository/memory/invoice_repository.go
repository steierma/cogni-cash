package memory

import (
	"context"
	"sync"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

const maxInvoices = 500

type InvoiceRepository struct {
	mu       sync.RWMutex
	invoices map[uuid.UUID]entity.Invoice
	order    []uuid.UUID
}

func NewInvoiceRepository() *InvoiceRepository {
	return &InvoiceRepository{
		invoices: make(map[uuid.UUID]entity.Invoice),
		order:    make([]uuid.UUID, 0, maxInvoices),
	}
}

func (r *InvoiceRepository) Save(ctx context.Context, invoice entity.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if invoice.ID == uuid.Nil {
		invoice.ID = uuid.New()
	}

	// Check if already exists (for upsert-like behavior)
	if _, exists := r.invoices[invoice.ID]; !exists {
		if len(r.order) >= maxInvoices {
			// Evict oldest
			oldestID := r.order[0]
			delete(r.invoices, oldestID)
			r.order = r.order[1:]
		}
		r.order = append(r.order, invoice.ID)
	}

	r.invoices[invoice.ID] = invoice
	return nil
}

func (r *InvoiceRepository) Update(ctx context.Context, invoice entity.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, ok := r.invoices[invoice.ID]
	if !ok || old.UserID != invoice.UserID {
		return entity.ErrInvoiceNotFound
	}
	r.invoices[invoice.ID] = invoice
	return nil
}

func (r *InvoiceRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	invoice, ok := r.invoices[id]
	if !ok || invoice.UserID != userID {
		return entity.Invoice{}, entity.ErrInvoiceNotFound
	}
	return invoice, nil
}

func (r *InvoiceRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]entity.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var invoices []entity.Invoice
	for _, inv := range r.invoices {
		if inv.UserID == userID {
			invoices = append(invoices, inv)
		}
	}
	return invoices, nil
}

func (r *InvoiceRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	inv, ok := r.invoices[id]
	if !ok || inv.UserID != userID {
		return entity.ErrInvoiceNotFound
	}
	delete(r.invoices, id)
	return nil
}

func (r *InvoiceRepository) ExistsByContentHash(ctx context.Context, hash string, userID uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, inv := range r.invoices {
		if inv.ContentHash == hash && inv.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (r *InvoiceRepository) GetOriginalFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) (content []byte, mimeType string, fileName string, err error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	inv, ok := r.invoices[id]
	if !ok || inv.UserID != userID {
		return nil, "", "", entity.ErrInvoiceNotFound
	}
	return inv.OriginalFileContent, "application/pdf", inv.OriginalFileName, nil
}

var _ port.InvoiceRepository = (*InvoiceRepository)(nil)
