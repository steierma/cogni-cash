// Package invoice provides a file-to-text extractor for invoice documents.
// Supported formats: PDF (via ledongthuc/pdf).
// For image-based PDFs or unsupported formats the caller should fall back to
// the LLM multimodal path.
package invoice

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

// Parser implements port.InvoiceParser by extracting plain text from files.
type Parser struct{}

// NewParser constructs a new invoice file parser.
func NewParser() *Parser { return &Parser{} }

// Extract reads the file content and returns its plain-text content.
// mimeType is used as a hint to determine the extraction strategy.
func (p *Parser) Extract(_ context.Context, _ uuid.UUID, fileBytes []byte, mimeType string) (string, error) {
	// If it's a PDF, extract text directly
	if strings.Contains(strings.ToLower(mimeType), "application/pdf") {
		return extractPDF(fileBytes)
	}

	return "", fmt.Errorf("invoice parser: unsupported mime type %q", mimeType)
}

// extractPDF pulls all readable text out of a PDF byte slice using ledongthuc/pdf.
func extractPDF(fileBytes []byte) (string, error) {
	readerAt := bytes.NewReader(fileBytes)
	r, err := pdf.NewReader(readerAt, int64(len(fileBytes)))
	if err != nil {
		return "", fmt.Errorf("invoice parser: create pdf reader: %w", err)
	}

	var buf bytes.Buffer
	totalPage := r.NumPage()
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		rows, err := p.GetTextByRow()
		if err != nil {
			continue // skip unreadable pages
		}
		for _, row := range rows {
			for _, word := range row.Content {
				buf.WriteString(word.S)
				buf.WriteRune(' ')
			}
			buf.WriteRune('\n')
		}
	}
	return strings.TrimSpace(buf.String()), nil
}
