package memory

import (
	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type PayslipRepository struct {
	payslips map[string]entity.Payslip
	mu       sync.RWMutex
}

func NewPayslipRepository() *PayslipRepository {
	r := &PayslipRepository{
		payslips: make(map[string]entity.Payslip),
	}
	r.seedData()
	return r
}

func (r *PayslipRepository) seedData() {
	userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
	baseGross := 4500.00
	baseNet := 2900.00

	// Seed 4 years: 2021 to 2024
	for year := 2021; year <= 2024; year++ {
		for month := 1; month <= 12; month++ {
			id := uuid.New().String()
			gross := baseGross
			net := baseNet
			var bonuses []entity.Bonus

			// June and November Bonuses
			if month == 6 {
				b := baseGross * 0.5 // Half month holiday bonus
				bonuses = append(bonuses, entity.Bonus{Description: "Holiday Bonus", Amount: b, BaseAmount: b})
				gross += b
				net += (b * 0.55) // Approximated net impact
			} else if month == 11 {
				b := baseGross * 0.8 // Christmas bonus
				bonuses = append(bonuses, entity.Bonus{Description: "Christmas Bonus", Amount: b, BaseAmount: b})
				gross += b
				net += (b * 0.55)
			}

			p := entity.Payslip{
				ID:               id,
				UserID:           userID,
				PeriodMonthNum:   month,
				PeriodYear:       year,
				EmployerName:     "Acme Corp",
				Currency:         "EUR",
				GrossPay:         gross,
				BaseGrossPay:     baseGross,
				NetPay:           net,
				BaseNetPay:       baseNet,
				PayoutAmount:     net,
				BasePayoutAmount: baseNet,
				Bonuses:          bonuses,
				CreatedAt:        time.Date(year, time.Month(month), 28, 8, 0, 0, 0, time.UTC),
			}
			r.payslips[id] = p
		}
		// Yearly salary increase of ~2% (falls within 1.5 - 3%)
		baseGross *= 1.02
		baseNet *= 1.02
	}
}

func (r *PayslipRepository) Save(ctx context.Context, payslip *entity.Payslip) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if payslip.ID == "" {
		payslip.ID = uuid.New().String()
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

func (r *PayslipRepository) ExistsByOriginalFileName(ctx context.Context, fileName string, userID uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.payslips {
		if p.OriginalFileName == fileName && p.UserID == userID {
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
			payslips = append(payslips, p)
		}
	}

	// Simple pagination
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

func (r *PayslipRepository) UpdateBaseAmount(ctx context.Context, id string, baseGross, baseNet, basePayout float64, currency string, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.payslips[id]
	if !ok || p.UserID != userID {
		return entity.ErrPayslipNotFound
	}
	p.BaseGrossPay = baseGross
	p.BaseNetPay = baseNet
	p.BasePayoutAmount = basePayout
	p.Currency = currency
	r.payslips[id] = p
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
			mimeType = mimeType[0:idx]
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
		summary.TotalGross += p.BaseGrossPay
		summary.TotalNet += p.BaseNetPay
		summary.TotalPayout += p.BasePayoutAmount
		for _, b := range p.Bonuses {
			summary.TotalBonuses += b.BaseAmount
		}

		if latest == nil || p.PeriodYear > latest.PeriodYear || (p.PeriodYear == latest.PeriodYear && p.PeriodMonthNum > latest.PeriodMonthNum) {
			previous = latest
			latest = p
		} else if previous == nil || p.PeriodYear > previous.PeriodYear || (p.PeriodYear == previous.PeriodYear && p.PeriodMonthNum > previous.PeriodMonthNum) {
			previous = p
		}
	}

	if latest != nil {
		summary.LatestNetPay = latest.BaseNetPay
		summary.LatestPeriod = fmt.Sprintf("%04d-%02d", latest.PeriodYear, latest.PeriodMonthNum)

		if previous != nil && previous.BaseNetPay > 0 {
			summary.NetPayTrend = ((latest.BaseNetPay - previous.BaseNetPay) / previous.BaseNetPay) * 100
		}
	}

	// For simple memory repo, just take last 12 from whatever order we have
	trendCount := 0
	for i := len(userPayslips) - 1; i >= 0 && trendCount < 12; i-- {
		p := userPayslips[i]
		summary.Trends = append([]entity.PayslipTrend{{
			Period: fmt.Sprintf("%04d-%02d", p.PeriodYear, p.PeriodMonthNum),
			Gross:  p.BaseGrossPay,
			Net:    p.BaseNetPay,
		}}, summary.Trends...)
		trendCount++
	}

	return summary, nil
}

var _ port.PayslipRepository = (*PayslipRepository)(nil)
