package ai

import (
	"context"
	"log/slog"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

// LLMPayslipParser defines the method we expect the LLM adapter to fulfill.
type LLMPayslipParser interface {
	// Refactored to accept the raw file data and a MIME type
	ParsePayslipDocument(ctx context.Context, userID uuid.UUID, mimeType string, data []byte) (entity.Payslip, error)
}

// PayslipParser implements port.PayslipParser using an LLM.
type PayslipParser struct {
	llm    LLMPayslipParser
	logger *slog.Logger
}

func NewPayslipParser(llm LLMPayslipParser, logger *slog.Logger) *PayslipParser {
	return &PayslipParser{
		llm:    llm,
		logger: logger,
	}
}

func (p *PayslipParser) Parse(ctx context.Context, userID uuid.UUID, fileBytes []byte) (entity.Payslip, error) {
	// 1. Determine the MIME type (assume PDF for now, but in a real app this would be passed or detected)
	mimeType := "application/pdf"

	p.logger.Info("Sending document to LLM for payslip parsing",
		"size_bytes", len(fileBytes),
		"mime_type", mimeType,
		"user_id", userID,
	)

	// 2. Pass the blob data to the LLM adapter
	payslip, err := p.llm.ParsePayslipDocument(ctx, userID, mimeType, fileBytes)
	if err != nil {
		return entity.Payslip{}, err
	}

	return payslip, nil
}
