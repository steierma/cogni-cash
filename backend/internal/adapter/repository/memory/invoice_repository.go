package memory

import (
	"context"
	"net/http"
	"strings"
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

	// Ensure splits have IDs
	for i := range invoice.Splits {
		if invoice.Splits[i].ID == uuid.Nil {
			invoice.Splits[i].ID = uuid.New()
		}
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

	// Ensure splits have IDs
	for i := range invoice.Splits {
		if invoice.Splits[i].ID == uuid.Nil {
			invoice.Splits[i].ID = uuid.New()
		}
	}

	r.invoices[invoice.ID] = invoice
	return nil
}

func (r *InvoiceRepository) UpdateBaseAmount(ctx context.Context, id uuid.UUID, baseAmount float64, baseCurrency string, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	inv, ok := r.invoices[id]
	if !ok || inv.UserID != userID {
		return entity.ErrInvoiceNotFound
	}
	inv.BaseAmount = baseAmount
	inv.BaseCurrency = baseCurrency
	r.invoices[id] = inv
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

func (r *InvoiceRepository) FindAll(ctx context.Context, filter entity.InvoiceFilter) ([]entity.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var invoices []entity.Invoice
	for _, id := range r.order {
		inv := r.invoices[id]
		if inv.UserID == filter.UserID {
			invoices = append(invoices, inv)
		}
	}

	if filter.Offset >= len(invoices) {
		return []entity.Invoice{}, nil
	}

	end := len(invoices)
	if filter.Limit > 0 && filter.Offset+filter.Limit < end {
		end = filter.Offset + filter.Limit
	}

	return invoices[filter.Offset:end], nil
}

func (r *InvoiceRepository) UpdateCategoriesBulk(ctx context.Context, ids []uuid.UUID, categoryID *uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, id := range ids {
		if inv, ok := r.invoices[id]; ok && inv.UserID == userID {
			inv.CategoryID = categoryID
			r.invoices[id] = inv
		}
	}
	return nil
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

func (r *InvoiceRepository) DeleteSplits(ctx context.Context, invoiceID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if inv, ok := r.invoices[invoiceID]; ok && inv.UserID == userID {
		inv.Splits = nil
		r.invoices[invoiceID] = inv
	}
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

func (r *InvoiceRepository) GetOriginalFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) ([]byte, string, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	inv, ok := r.invoices[id]
	if !ok || inv.UserID != userID {
		return nil, "", "", entity.ErrInvoiceNotFound
	}

	mimeType := "application/octet-stream"
	if len(inv.OriginalFileContent) > 0 {
		mimeType = http.DetectContentType(inv.OriginalFileContent)
		if idx := strings.IndexByte(mimeType, ';'); idx >= 0 {
			mimeType = mimeType[:idx]
		}
	}

	return inv.OriginalFileContent, mimeType, inv.OriginalFileName, nil
}

var _ port.InvoiceRepository = (*InvoiceRepository)(nil)
