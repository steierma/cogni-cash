package service

import (
	"context"
	"log/slog"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

type CurrencyService struct {
	ratePort     port.CurrencyExchangeRatePort
	settingsRepo port.SettingsRepository
	txnRepo      port.BankStatementRepository
	invoiceRepo  port.InvoiceRepository
	payslipRepo  port.PayslipRepository
	logger       *slog.Logger
}

func NewCurrencyService(
	ratePort port.CurrencyExchangeRatePort,
	settingsRepo port.SettingsRepository,
	txnRepo port.BankStatementRepository,
	invoiceRepo port.InvoiceRepository,
	payslipRepo port.PayslipRepository,
	logger *slog.Logger,
) *CurrencyService {
	if logger == nil {
		logger = slog.Default()
	}
	return &CurrencyService{
		ratePort:     ratePort,
		settingsRepo: settingsRepo,
		txnRepo:      txnRepo,
		invoiceRepo:  invoiceRepo,
		payslipRepo:  payslipRepo,
		logger:       logger,
	}
}

// UpdateBaseAmountsForUser fetches the user's base currency and updates all records that need conversion.
func (s *CurrencyService) UpdateBaseAmountsForUser(ctx context.Context, userID uuid.UUID) error {
	baseCurrency, err := s.settingsRepo.Get(ctx, "BASE_DISPLAY_CURRENCY", userID)
	if err != nil || baseCurrency == "" {
		baseCurrency = "EUR" // Default fallback
	}

	// 1. Transactions
	txns, err := s.txnRepo.FindTransactions(ctx, entity.TransactionFilter{UserID: userID})
	if err == nil {
		for _, tx := range txns {
			if tx.Currency != baseCurrency && (tx.BaseCurrency != baseCurrency || tx.BaseAmount == 0) {
				rate, err := s.ratePort.GetRate(ctx, tx.Currency, baseCurrency, tx.BookingDate)
				if err != nil {
					s.logger.Error("Failed to fetch rate for transaction", "hash", tx.ContentHash, "error", err)
					continue
				}
				baseAmount := tx.Amount * rate
				if err := s.txnRepo.UpdateTransactionBaseAmount(ctx, tx.ContentHash, baseAmount, baseCurrency, userID); err != nil {
					s.logger.Error("Failed to update transaction base amount", "hash", tx.ContentHash, "error", err)
				}
			} else if tx.Currency == baseCurrency && (tx.BaseCurrency != baseCurrency || tx.BaseAmount != tx.Amount) {
				// Fast path for same currency
				if err := s.txnRepo.UpdateTransactionBaseAmount(ctx, tx.ContentHash, tx.Amount, baseCurrency, userID); err != nil {
					s.logger.Error("Failed to update transaction base amount (same currency)", "hash", tx.ContentHash, "error", err)
				}
			}
		}
	}

	// 2. Invoices
	invoices, err := s.invoiceRepo.FindAll(ctx, entity.InvoiceFilter{UserID: userID})
	if err == nil {
		for _, inv := range invoices {
			if inv.Currency != baseCurrency && (inv.BaseCurrency != baseCurrency || inv.BaseAmount == 0) {
				rate, err := s.ratePort.GetRate(ctx, inv.Currency, baseCurrency, inv.IssuedAt)
				if err != nil {
					s.logger.Error("Failed to fetch rate for invoice", "id", inv.ID, "error", err)
					continue
				}
				baseAmount := inv.Amount * rate
				if err := s.invoiceRepo.UpdateBaseAmount(ctx, inv.ID, baseAmount, baseCurrency, userID); err != nil {
					s.logger.Error("Failed to update invoice base amount", "id", inv.ID, "error", err)
				}
			} else if inv.Currency == baseCurrency && (inv.BaseCurrency != baseCurrency || inv.BaseAmount != inv.Amount) {
				if err := s.invoiceRepo.UpdateBaseAmount(ctx, inv.ID, inv.Amount, baseCurrency, userID); err != nil {
					s.logger.Error("Failed to update invoice base amount (same currency)", "id", inv.ID, "error", err)
				}
			}
		}
	}

	// 3. Payslips
	payslips, err := s.payslipRepo.FindAll(ctx, entity.PayslipFilter{UserID: userID})
	if err == nil {
		for _, p := range payslips {
			// Payslips use middle of the period for rate fetching if date is not explicit
			rateDate := time.Date(p.PeriodYear, time.Month(p.PeriodMonthNum), 15, 0, 0, 0, 0, time.UTC)
			if p.Currency != baseCurrency && (p.BasePayoutAmount == 0) {
				rate, err := s.ratePort.GetRate(ctx, p.Currency, baseCurrency, rateDate)
				if err != nil {
					s.logger.Error("Failed to fetch rate for payslip", "id", p.ID, "error", err)
					continue
				}
				baseGross := p.GrossPay * rate
				baseNet := p.NetPay * rate
				basePayout := p.PayoutAmount * rate
				if err := s.payslipRepo.UpdateBaseAmount(ctx, p.ID, baseGross, baseNet, basePayout, baseCurrency, userID); err != nil {
					s.logger.Error("Failed to update payslip base amount", "id", p.ID, "error", err)
				}
			} else if p.Currency == baseCurrency && (p.BasePayoutAmount != p.PayoutAmount) {
				if err := s.payslipRepo.UpdateBaseAmount(ctx, p.ID, p.GrossPay, p.NetPay, p.PayoutAmount, baseCurrency, userID); err != nil {
					s.logger.Error("Failed to update payslip base amount (same currency)", "id", p.ID, "error", err)
				}
			}
		}
	}

	return nil
}
