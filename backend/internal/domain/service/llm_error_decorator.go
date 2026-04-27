package service

import (
	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"context"
	"log/slog"
	"strings"

	"github.com/google/uuid"
)

type LLMAdapterPorts interface {
	port.BankStatementAIParser
	port.PayslipAIParser
	port.InvoiceAICategorizer
	port.TransactionCategorizer
	port.DocumentAIParser
	port.SubscriptionEnricher
	port.CancellationLetterGenerator
}

type LLMNotifierDecorator struct {
	base     LLMAdapterPorts
	notifier port.NotificationUseCase
	logger   *slog.Logger
}

func NewLLMNotifierDecorator(base LLMAdapterPorts, notifier port.NotificationUseCase, logger *slog.Logger) *LLMNotifierDecorator {
	if logger == nil {
		logger = slog.Default()
	}
	return &LLMNotifierDecorator{
		base:     base,
		notifier: notifier,
		logger:   logger,
	}
}

func (d *LLMNotifierDecorator) handleError(err error, operation string) {
	if err == nil {
		return
	}

	errStr := strings.ToLower(err.Error())
	// Alert on rate limit, token limit, or generic external server errors.
	if strings.Contains(errStr, "status: 429") ||
		strings.Contains(errStr, "status: 400") ||
		strings.Contains(errStr, "token limit") ||
		strings.Contains(errStr, "status: 500") ||
		strings.Contains(errStr, "status: 502") ||
		strings.Contains(errStr, "status: 503") ||
		strings.Contains(errStr, "status: 504") {

		d.logger.Warn("Critical LLM error intercepted by decorator", "operation", operation, "error", err)

		// Dispatch notification asynchronously so we don't delay the HTTP response
		go func(errCtx error, op string) {
			// Using background context since the original HTTP context might be cancelled
			alertCtx := context.Background()
			alertSubject := "Cogni-Cash LLM Alert: " + op
			if alertErr := d.notifier.SendAdminAlert(alertCtx, alertSubject, errCtx.Error()); alertErr != nil {
				d.logger.Error("Failed to dispatch LLM alert email", "error", alertErr)
			}
		}(err, operation)
	}
}

func (d *LLMNotifierDecorator) ParseBankStatement(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, fileBytes []byte) (entity.BankStatement, error) {
	res, err := d.base.ParseBankStatement(ctx, userID, fileName, mimeType, fileBytes)
	d.handleError(err, "ParseBankStatement")
	return res, err
}

func (d *LLMNotifierDecorator) ParsePayslip(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, fileBytes []byte) (entity.Payslip, error) {
	res, err := d.base.ParsePayslip(ctx, userID, fileName, mimeType, fileBytes)
	d.handleError(err, "ParsePayslip")
	return res, err
}

func (d *LLMNotifierDecorator) CategorizeInvoice(ctx context.Context, userID uuid.UUID, req port.CategorizationRequest) (port.InvoiceCategorizationResult, error) {
	res, err := d.base.CategorizeInvoice(ctx, userID, req)
	d.handleError(err, "CategorizeInvoice")
	return res, err
}

func (d *LLMNotifierDecorator) CategorizeInvoiceImage(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, imageBytes []byte, categories []string) (port.InvoiceCategorizationResult, error) {
	res, err := d.base.CategorizeInvoiceImage(ctx, userID, fileName, mimeType, imageBytes, categories)
	d.handleError(err, "CategorizeInvoiceImage")
	return res, err
}

func (d *LLMNotifierDecorator) CategorizeTransactionsBatch(ctx context.Context, userID uuid.UUID, txns []port.TransactionToCategorize, categories []string, examples []entity.CategorizationExample) ([]port.CategorizedTransaction, error) {
	res, err := d.base.CategorizeTransactionsBatch(ctx, userID, txns, categories, examples)
	d.handleError(err, "CategorizeTransactionsBatch")
	return res, err
}

func (d *LLMNotifierDecorator) ClassifyAndExtract(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, fileBytes []byte) (entity.DocumentType, map[string]interface{}, string, error) {
	docType, metadata, text, err := d.base.ClassifyAndExtract(ctx, userID, fileName, mimeType, fileBytes)
	d.handleError(err, "ClassifyAndExtract")
	return docType, metadata, text, err
}

func (d *LLMNotifierDecorator) VerifySubscriptionSuggestion(ctx context.Context, userID uuid.UUID, merchantName string, amount float64, currency string, billingCycle string) (bool, error) {
	res, err := d.base.VerifySubscriptionSuggestion(ctx, userID, merchantName, amount, currency, billingCycle)
	d.handleError(err, "VerifySubscriptionSuggestion")
	return res, err
}

func (d *LLMNotifierDecorator) EnrichSubscription(ctx context.Context, userID uuid.UUID, merchantName string, transactionDescriptions []string, language string) (port.SubscriptionEnrichmentResult, error) {
	res, err := d.base.EnrichSubscription(ctx, userID, merchantName, transactionDescriptions, language)
	d.handleError(err, "EnrichSubscription")
	return res, err
}

func (d *LLMNotifierDecorator) GenerateCancellationLetter(ctx context.Context, userID uuid.UUID, req port.CancellationLetterRequest) (port.CancellationLetterResult, error) {
	res, err := d.base.GenerateCancellationLetter(ctx, userID, req)
	d.handleError(err, "GenerateCancellationLetter")
	return res, err
}
