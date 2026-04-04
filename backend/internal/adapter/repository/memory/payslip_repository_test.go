package memory

import (
	"context"
	"testing"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

func TestPayslipRepository_GetSummary(t *testing.T) {
	ctx := context.Background()
	repo := NewPayslipRepository()
	userID := uuid.New()

	p1 := &entity.Payslip{
		ID:             "p1",
		UserID:         userID,
		PeriodMonthNum: 1,
		PeriodYear:     2026,
		GrossPay:       5000,
		NetPay:         3500,
		PayoutAmount:   3400,
		Bonuses: []entity.Bonus{
			{Description: "Bonus 1", Amount: 500},
		},
	}

	p2 := &entity.Payslip{
		ID:             "p2",
		UserID:         userID,
		PeriodMonthNum: 2,
		PeriodYear:     2026,
		GrossPay:       5000,
		NetPay:         3700,
		PayoutAmount:   3600,
	}

	repo.Save(ctx, p1)
	repo.Save(ctx, p2)

	summary, err := repo.GetSummary(ctx, userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if summary.PayslipCount != 2 {
		t.Errorf("expected 2 payslips, got %d", summary.PayslipCount)
	}
	if summary.TotalGross != 10000 {
		t.Errorf("expected 10000 gross, got %f", summary.TotalGross)
	}
	if summary.TotalBonuses != 500 {
		t.Errorf("expected 500 bonuses, got %f", summary.TotalBonuses)
	}
	if summary.LatestNetPay != 3700 {
		t.Errorf("expected latest net 3700, got %f", summary.LatestNetPay)
	}
	// (3700 - 3500) / 3500 = 5.714...%
	if summary.NetPayTrend < 5.7 || summary.NetPayTrend > 5.8 {
		t.Errorf("expected trend ~5.7, got %f", summary.NetPayTrend)
	}
}
