package ai

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"cogni-cash/internal/domain/entity"

	"github.com/ledongthuc/pdf"
)

// LLMStatementParser defines the method we expect the LLM adapter to fulfill.
type LLMStatementParser interface {
	ParseBankStatementText(ctx context.Context, text string) (entity.BankStatement, error)
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

func (p *Parser) Parse(ctx context.Context, filePath string) (entity.BankStatement, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	var rawText string
	var err error

	if ext == ".pdf" {
		rawText, err = extractPDFText(filePath)
	} else {
		// Treat CSV and XLS (which are often just CSV/HTML dumps) as plain text
		var b []byte
		b, err = os.ReadFile(filePath)
		rawText = string(b)
	}

	if err != nil {
		return entity.BankStatement{}, fmt.Errorf("ai parser: failed to read file %s: %w", filepath.Base(filePath), err)
	}

	// Truncate to save context window tokens if the file is absurdly huge
	if len(rawText) > 60000 {
		rawText = rawText[:60000]
	}

	p.logger.Info("Sending extracted text to LLM for bank statement parsing", "file", filepath.Base(filePath), "text_length", len(rawText))
	return p.llm.ParseBankStatementText(ctx, rawText)
}

func extractPDFText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
}
