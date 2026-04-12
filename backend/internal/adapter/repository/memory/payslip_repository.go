package memory

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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

	// Simple sort by period DESC
	for i := 0; i < len(payslips); i++ {
		for j := i + 1; j < len(payslips); j++ {
			if payslips[i].PeriodYear < payslips[j].PeriodYear || (payslips[i].PeriodYear == payslips[j].PeriodYear && payslips[i].PeriodMonthNum < payslips[j].PeriodMonthNum) {
				payslips[i], payslips[j] = payslips[j], payslips[i]
			}
		}
	}

	if filter.Offset >= len(payslips) {
		return []entity.Payslip{}, nil
	}

	end := len(payslips)
	if filter.Limit > 0 && filter.Offset+filter.Limit < end {
		end = filter.Offset + filter.Limit
	}

	return payslips[filter.Offset:end], nil
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
	
	mimeType := "application/octet-stream"
	if len(p.OriginalFileContent) > 0 {
		mimeType = http.DetectContentType(p.OriginalFileContent)
		if idx := strings.IndexByte(mimeType, ';'); idx >= 0 {
			mimeType = mimeType[:idx]
		}
	}
	
	return p.OriginalFileContent, mimeType, p.OriginalFileName, nil
}

func (r *PayslipRepository) GetSummary(ctx context.Context, userID uuid.UUID) (entity.PayslipSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	summary := entity.PayslipSummary{
		Trends: []entity.PayslipTrend{},
	}

	var userPayslips []entity.Payslip
	for _, p := range r.payslips {
		if p.UserID == userID {
			userPayslips = append(userPayslips, p)
		}
	}

	if len(userPayslips) == 0 {
		return summary, nil
	}

	summary.PayslipCount = len(userPayslips)

	var latest, previous *entity.Payslip

	for i := range userPayslips {
		p := &userPayslips[i]
		summary.TotalGross += p.GrossPay
		summary.TotalNet += p.NetPay
		summary.TotalPayout += p.PayoutAmount
		for _, b := range p.Bonuses {
			summary.TotalBonuses += b.Amount
		}

		if latest == nil || p.PeriodYear > latest.PeriodYear || (p.PeriodYear == latest.PeriodYear && p.PeriodMonthNum > latest.PeriodMonthNum) {
			previous = latest
			latest = p
		} else if previous == nil || p.PeriodYear > previous.PeriodYear || (p.PeriodYear == previous.PeriodYear && p.PeriodMonthNum > previous.PeriodMonthNum) {
			previous = p
		}
	}

	if latest != nil {
		summary.LatestNetPay = latest.NetPay
		summary.LatestPeriod = fmt.Sprintf("%04d-%02d", latest.PeriodYear, latest.PeriodMonthNum)

		if previous != nil && previous.NetPay > 0 {
			summary.NetPayTrend = ((latest.NetPay - previous.NetPay) / previous.NetPay) * 100
		}
	}

	// For simple memory repo, just take last 12 from whatever order we have
	// (In a real scenario, we'd sort them properly)
	trendCount := 0
	for i := len(userPayslips) - 1; i >= 0 && trendCount < 12; i-- {
		p := userPayslips[i]
		summary.Trends = append([]entity.PayslipTrend{{
			Period: fmt.Sprintf("%04d-%02d", p.PeriodYear, p.PeriodMonthNum),
			Gross:  p.GrossPay,
			Net:    p.NetPay,
		}}, summary.Trends...)
		trendCount++
	}

	return summary, nil
}

var _ port.PayslipRepository = (*PayslipRepository)(nil)
