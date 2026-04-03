package memory

import (
	"context"
	"sync"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

const maxPayslips = 100

type PayslipRepository struct {
	mu       sync.RWMutex
	payslips map[string]entity.Payslip
	order    []string
}

func NewPayslipRepository() *PayslipRepository {
	return &PayslipRepository{
		payslips: make(map[string]entity.Payslip),
		order:    make([]string, 0, maxPayslips),
	}
}

func (r *PayslipRepository) Save(ctx context.Context, payslip *entity.Payslip) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if payslip.ID == "" {
		payslip.ID = uuid.New().String()
	}

	if _, exists := r.payslips[payslip.ID]; !exists {
		if len(r.order) >= maxPayslips {
			// Evict oldest
			oldestID := r.order[0]
			delete(r.payslips, oldestID)
			r.order = r.order[1:]
		}
		r.order = append(r.order, payslip.ID)
	}

	r.payslips[payslip.ID] = *payslip
	return nil
}

func (r *PayslipRepository) ExistsByHash(ctx context.Context, hash string, userID uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.payslips {
		if p.ContentHash == hash && p.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (r *PayslipRepository) ExistsByOriginalFileName(ctx context.Context, originalFileName string, userID uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.payslips {
		if p.OriginalFileName == originalFileName && p.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (r *PayslipRepository) FindAll(ctx context.Context, filter entity.PayslipFilter) ([]entity.Payslip, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var payslips []entity.Payslip
	for _, p := range r.payslips {
		if p.UserID == filter.UserID {
			if filter.Employer != "" && p.EmployerName != filter.Employer {
				continue
			}
			payslips = append(payslips, p)
		}
	}
	return payslips, nil
}

func (r *PayslipRepository) FindByID(ctx context.Context, id string, userID uuid.UUID) (entity.Payslip, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.payslips[id]
	if !ok || p.UserID != userID {
		return entity.Payslip{}, entity.ErrPayslipNotFound
	}
	return p, nil
}

func (r *PayslipRepository) Update(ctx context.Context, payslip *entity.Payslip) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, ok := r.payslips[payslip.ID]
	if !ok || old.UserID != payslip.UserID {
		return entity.ErrPayslipNotFound
	}
	r.payslips[payslip.ID] = *payslip
	return nil
}

func (r *PayslipRepository) Delete(ctx context.Context, id string, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.payslips[id]
	if !ok || p.UserID != userID {
		return entity.ErrPayslipNotFound
	}
	delete(r.payslips, id)
	return nil
}

func (r *PayslipRepository) GetOriginalFile(ctx context.Context, id string, userID uuid.UUID) ([]byte, string, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.payslips[id]
	if !ok || p.UserID != userID {
		return nil, "", "", entity.ErrPayslipNotFound
	}
	return p.OriginalFileContent, "application/pdf", p.OriginalFileName, nil
}

var _ port.PayslipRepository = (*PayslipRepository)(nil)
