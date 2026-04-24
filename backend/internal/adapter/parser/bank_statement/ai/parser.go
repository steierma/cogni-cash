// Package ai provides an LLM-backed BankStatementParser that handles both
// text-based documents (PDF, CSV) and image files (JPEG, PNG, GIF, WEBP).
package ai

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"cogni-cash/internal/adapter/parser/pdfutil"
	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

// LLMStatementParser defines the single method the LLM adapter must fulfil.
// ParseBankStatementDocument mirrors ParsePayslipDocument in design: the caller
// passes the detected MIME type and raw bytes; the adapter decides the LLM path
// (multimodal for images, text-prompt for PDFs/CSV).
type LLMStatementParser interface {
	ParseBankStatement(ctx context.Context, userID uuid.UUID, fileName string, mimeType string, data []byte) (entity.BankStatement, error)
}

// imageMIMETypes is the set of MIME types that bypass text extraction.
var imageMIMETypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// Parser implements port.BankStatementParser using an LLM.
type Parser struct {
	llm    LLMStatementParser
	logger *slog.Logger
}

func NewParser(llm LLMStatementParser, logger *slog.Logger) *Parser {
	return &Parser{llm: llm, logger: logger}
}

// Parse detects the file type from magic bytes and routes accordingly:
//   - Image → passes raw bytes + detected MIME to ParseBankStatement (multimodal)
//   - PDF/text → extracts text first, then passes as text/plain to ParseBankStatement
func (p *Parser) Parse(ctx context.Context, userID uuid.UUID, fileBytes []byte) (entity.BankStatement, error) {
	detected := http.DetectContentType(fileBytes)

	// Trim charset info if any
	if idx := strings.IndexByte(detected, ';'); idx >= 0 {
		detected = detected[:idx]
	}

	if imageMIMETypes[detected] {
		p.logger.Info("Image file detected in AI bank statement parser, using multimodal path",
			"mime_type", detected, "size_bytes", len(fileBytes), "user_id", userID)
		return p.llm.ParseBankStatement(ctx, userID, "ai_extracted_image", detected, fileBytes)
	}

	// Text/PDF path: extract readable text, then forward as plain-text bytes
	var rawText string
	if detected == "application/pdf" {
		if text, err := pdfutil.ExtractText(fileBytes); err == nil {
			rawText = text
		}
	}
	if rawText == "" {
		rawText = string(fileBytes)
	}

	if len(rawText) > 60000 {
		rawText = rawText[:60000]
	}

	p.logger.Info("Sending extracted text to LLM for bank statement parsing", "text_length", len(rawText))
	return p.llm.ParseBankStatement(ctx, userID, "ai_extracted_text", "text/plain", []byte(rawText))
}
