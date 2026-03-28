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
	if _, ok := r.invoices[invoice.ID]; !ok {
		return entity.ErrInvoiceNotFound
	}
	r.invoices[invoice.ID] = invoice
	return nil
}

func (r *InvoiceRepository) FindByID(ctx context.Context, id uuid.UUID) (entity.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	invoice, ok := r.invoices[id]
	if !ok {
		return entity.Invoice{}, entity.ErrInvoiceNotFound
	}
	return invoice, nil
}

func (r *InvoiceRepository) FindAll(ctx context.Context) ([]entity.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var invoices []entity.Invoice
	for _, inv := range r.invoices {
		invoices = append(invoices, inv)
	}
	return invoices, nil
}

func (r *InvoiceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.invoices[id]; !ok {
		return entity.ErrInvoiceNotFound
	}
	delete(r.invoices, id)
	return nil
}

func (r *InvoiceRepository) ExistsByContentHash(ctx context.Context, hash string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, inv := range r.invoices {
		if inv.ContentHash == hash {
			return true, nil
		}
	}
	return false, nil
}

func (r *InvoiceRepository) GetOriginalFile(ctx context.Context, id uuid.UUID) (content []byte, mimeType string, fileName string, err error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	inv, ok := r.invoices[id]
	if !ok {
		return nil, "", "", entity.ErrInvoiceNotFound
	}
	return inv.OriginalFileContent, inv.OriginalFileMime, inv.OriginalFileName, nil
}

var _ port.InvoiceRepository = (*InvoiceRepository)(nil)
