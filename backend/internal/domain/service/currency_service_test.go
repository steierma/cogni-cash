package service_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"
	portmock "cogni-cash/internal/domain/port/mock"
	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
)

func TestCurrencyService_UpdateBaseAmountsForUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	baseCurrency := "EUR"

	mockRatePort := new(portmock.MockCurrencyExchangeRatePort)
	mockSettingsRepo := new(portmock.MockSettingsRepository)
	mockTxnRepo := new(portmock.MockBankStatementRepository)
	mockInvoiceRepo := new(portmock.MockInvoiceRepository)
	mockPayslipRepo := new(portmock.MockPayslipRepository)

	s := service.NewCurrencyService(mockRatePort, mockSettingsRepo, mockTxnRepo, mockInvoiceRepo, mockPayslipRepo, slog.Default())

	t.Run("converts transactions with different currency", func(t *testing.T) {
		mockSettingsRepo.On("Get", ctx, "BASE_DISPLAY_CURRENCY", userID).Return(baseCurrency, nil).Once()

		txns := []entity.Transaction{
			{
				ContentHash: "hash1",
				Amount:      100,
				Currency:    "USD",
				BookingDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		}
		mockTxnRepo.On("FindTransactions", ctx, entity.TransactionFilter{UserID: userID}).Return(txns, nil).Once()
		mockRatePort.On("GetRate", ctx, "USD", "EUR", txns[0].BookingDate).Return(0.9, nil).Once()
		mockTxnRepo.On("UpdateTransactionBaseAmount", ctx, "hash1", 90.0, "EUR", userID).Return(nil).Once()

		// For invoices and payslips, return empty for simplicity in this subtest
		mockInvoiceRepo.On("FindAll", ctx, testifymock.Anything).Return([]entity.Invoice{}, nil).Once()
		mockPayslipRepo.On("FindAll", ctx, testifymock.Anything).Return([]entity.Payslip{}, nil).Once()

		err := s.UpdateBaseAmountsForUser(ctx, userID)
		assert.NoError(t, err)

		mockSettingsRepo.AssertExpectations(t)
		mockTxnRepo.AssertExpectations(t)
		mockRatePort.AssertExpectations(t)
	})

	t.Run("updates base amount for same currency", func(t *testing.T) {
		mockSettingsRepo.On("Get", ctx, "BASE_DISPLAY_CURRENCY", userID).Return(baseCurrency, nil).Once()

		txns := []entity.Transaction{
			{
				ContentHash: "hash2",
				Amount:      50,
				Currency:    "EUR",
				BaseAmount:  0, // Needs update
			},
		}
		mockTxnRepo.On("FindTransactions", ctx, entity.TransactionFilter{UserID: userID}).Return(txns, nil).Once()
		mockTxnRepo.On("UpdateTransactionBaseAmount", ctx, "hash2", 50.0, "EUR", userID).Return(nil).Once()

		mockInvoiceRepo.On("FindAll", ctx, testifymock.Anything).Return([]entity.Invoice{}, nil).Once()
		mockPayslipRepo.On("FindAll", ctx, testifymock.Anything).Return([]entity.Payslip{}, nil).Once()

		err := s.UpdateBaseAmountsForUser(ctx, userID)
		assert.NoError(t, err)
	})

	t.Run("converts invoices and payslips with different currency", func(t *testing.T) {
		mockSettingsRepo.On("Get", ctx, "BASE_DISPLAY_CURRENCY", userID).Return(baseCurrency, nil).Once()

		mockTxnRepo.On("FindTransactions", ctx, entity.TransactionFilter{UserID: userID}).Return([]entity.Transaction{}, nil).Once()

		invDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
		invoices := []entity.Invoice{
			{
				ID:       uuid.New(),
				Amount:   200,
				Currency: "USD",
				IssuedAt: invDate,
			},
		}
		mockInvoiceRepo.On("FindAll", ctx, testifymock.Anything).Return(invoices, nil).Once()
		mockRatePort.On("GetRate", ctx, "USD", "EUR", invDate).Return(0.8, nil).Once()
		mockInvoiceRepo.On("UpdateBaseAmount", ctx, invoices[0].ID, 160.0, "EUR", userID).Return(nil).Once()

		payslips := []entity.Payslip{
			{
				ID:             uuid.New().String(),
				PeriodYear:     2024,
				PeriodMonthNum: 1,
				Currency:       "USD",
				GrossPay:       1000,
				NetPay:         800,
				PayoutAmount:   800,
			},
		}
		mockPayslipRepo.On("FindAll", ctx, testifymock.Anything).Return(payslips, nil).Once()
		mockRatePort.On("GetRate", ctx, "USD", "EUR", time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)).Return(0.8, nil).Once()
		mockPayslipRepo.On("UpdateBaseAmount", ctx, payslips[0].ID, 800.0, 640.0, 640.0, "EUR", userID).Return(nil).Once()

		err := s.UpdateBaseAmountsForUser(ctx, userID)
		assert.NoError(t, err)
	})

	t.Run("updates base amount for invoices and payslips with same currency", func(t *testing.T) {
		mockSettingsRepo.On("Get", ctx, "BASE_DISPLAY_CURRENCY", userID).Return(baseCurrency, nil).Once()

		mockTxnRepo.On("FindTransactions", ctx, entity.TransactionFilter{UserID: userID}).Return([]entity.Transaction{}, nil).Once()

		invoices := []entity.Invoice{
			{
				ID:         uuid.New(),
				Amount:     150,
				Currency:   "EUR",
				BaseAmount: 0,
			},
		}
		mockInvoiceRepo.On("FindAll", ctx, testifymock.Anything).Return(invoices, nil).Once()
		mockInvoiceRepo.On("UpdateBaseAmount", ctx, invoices[0].ID, 150.0, "EUR", userID).Return(nil).Once()

		payslips := []entity.Payslip{
			{
				ID:               uuid.New().String(),
				Currency:         "EUR",
				GrossPay:         2000,
				NetPay:           1500,
				PayoutAmount:     1500,
				BasePayoutAmount: 0,
			},
		}
		mockPayslipRepo.On("FindAll", ctx, testifymock.Anything).Return(payslips, nil).Once()
		mockPayslipRepo.On("UpdateBaseAmount", ctx, payslips[0].ID, 2000.0, 1500.0, 1500.0, "EUR", userID).Return(nil).Once()

		err := s.UpdateBaseAmountsForUser(ctx, userID)
		assert.NoError(t, err)
	})
}

