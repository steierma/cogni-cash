package ai

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"cogni-cash/internal/domain/entity"
)

// LLMPayslipParser defines the method we expect the LLM adapter to fulfill.
type LLMPayslipParser interface {
	// Refactored to accept the raw file data and a MIME type
	ParsePayslipDocument(ctx context.Context, mimeType string, data []byte) (entity.Payslip, error)
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

func (p *PayslipParser) Parse(ctx context.Context, filePath string) (entity.Payslip, error) {
	// 1. Read the raw bytes of the file directly
	data, err := os.ReadFile(filePath)
	if err != nil {
		return entity.Payslip{}, fmt.Errorf("ai payslip parser: failed to read file %s: %w", filepath.Base(filePath), err)
	}

	// 2. Determine the MIME type based on the extension
	ext := strings.ToLower(filepath.Ext(filePath))
	mimeType := "text/plain" // Default fallback

	switch ext {
	case ".pdf":
		mimeType = "application/pdf"
	case ".png":
		mimeType = "image/png"
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	}

	p.logger.Info("Sending document to LLM for payslip parsing",
		"file", filepath.Base(filePath),
		"size_bytes", len(data),
		"mime_type", mimeType,
	)

	// 3. Pass the blob data to the LLM adapter
	payslip, err := p.llm.ParsePayslipDocument(ctx, mimeType, data)
	if err != nil {
		return entity.Payslip{}, err
	}

	payslip.SourceFile = filePath
	return payslip, nil
}
