package ai

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

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

// Parse detects the file MIME type from magic bytes and calls ParsePayslipDocument
// with the correct MIME so the LLM adapter can choose the right path
// (multimodal for images, text-based for PDFs).
func (p *PayslipParser) Parse(ctx context.Context, userID uuid.UUID, fileBytes []byte) (entity.Payslip, error) {
	mimeType := http.DetectContentType(fileBytes)

	// Trim charset info if any
	if idx := strings.IndexByte(mimeType, ';'); idx >= 0 {
		mimeType = mimeType[:idx]
	}

	p.logger.Info("Sending document to LLM for payslip parsing",
		"size_bytes", len(fileBytes),
		"mime_type", mimeType,
		"user_id", userID,
	)

	return p.llm.ParsePayslipDocument(ctx, userID, mimeType, fileBytes)
}
