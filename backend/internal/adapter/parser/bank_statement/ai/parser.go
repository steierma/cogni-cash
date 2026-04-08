package ai

import (
	"bytes"
	"context"
	"log/slog"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

// LLMStatementParser defines the method we expect the LLM adapter to fulfill.
type LLMStatementParser interface {
	ParseBankStatementText(ctx context.Context, userID uuid.UUID, text string) (entity.BankStatement, error)
}

// Parser implements port.BankStatementParser using an LLM.
type Parser struct {
	llm    LLMStatementParser
	logger *slog.Logger
}

func NewParser(llm LLMStatementParser, logger *slog.Logger) *Parser {
	return &Parser{
		llm:    llm,
		logger: logger,
	}
}

func (p *Parser) Parse(ctx context.Context, userID uuid.UUID, fileBytes []byte) (entity.BankStatement, error) {
	// 1. Try to extract PDF text first, fallback to raw text
	rawText, err := extractPDFText(fileBytes)
	if err != nil {
		// Fallback: Treat as plain text
		rawText = string(fileBytes)
	}

	// Truncate to save context window tokens if the file is absurdly huge
	if len(rawText) > 60000 {
		rawText = rawText[:60000]
	}

	p.logger.Info("Sending extracted text to LLM for bank statement parsing", "text_length", len(rawText))
	return p.llm.ParseBankStatementText(ctx, userID, rawText)
}

func extractPDFText(fileBytes []byte) (string, error) {
	readerAt := bytes.NewReader(fileBytes)
	r, err := pdf.NewReader(readerAt, int64(len(fileBytes)))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
}
