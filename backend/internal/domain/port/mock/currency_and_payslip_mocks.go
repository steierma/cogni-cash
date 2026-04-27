package mock

import (
	"context"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockCurrencyExchangeRatePort struct {
	mock.Mock
}

func (m *MockCurrencyExchangeRatePort) GetRate(ctx context.Context, from, to string, date time.Time) (float64, error) {
	args := m.Called(ctx, from, to, date)
	return args.Get(0).(float64), args.Error(1)
}

type MockPayslipRepository struct {
	mock.Mock
}

func (m *MockPayslipRepository) Save(ctx context.Context, p *entity.Payslip) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPayslipRepository) ExistsByHash(ctx context.Context, hash string, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, hash, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPayslipRepository) ExistsByOriginalFileName(ctx context.Context, originalFileName string, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, originalFileName, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPayslipRepository) FindByID(ctx context.Context, id string, userID uuid.UUID) (entity.Payslip, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(entity.Payslip), args.Error(1)
}

func (m *MockPayslipRepository) FindAll(ctx context.Context, filter entity.PayslipFilter) ([]entity.Payslip, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entity.Payslip), args.Error(1)
}

func (m *MockPayslipRepository) Update(ctx context.Context, payslip *entity.Payslip) error {
	args := m.Called(ctx, payslip)
	return args.Error(0)
}

func (m *MockPayslipRepository) Delete(ctx context.Context, id string, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockPayslipRepository) UpdateBaseAmount(ctx context.Context, id string, baseGross, baseNet, basePayout float64, baseCurrency string, userID uuid.UUID) error {
	args := m.Called(ctx, id, baseGross, baseNet, basePayout, baseCurrency, userID)
	return args.Error(0)
}

func (m *MockPayslipRepository) GetOriginalFile(ctx context.Context, id string, userID uuid.UUID) ([]byte, string, string, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).([]byte), args.String(1), args.String(2), args.Error(3)
}

func (m *MockPayslipRepository) GetSummary(ctx context.Context, userID uuid.UUID) (entity.PayslipSummary, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(entity.PayslipSummary), args.Error(1)
}
