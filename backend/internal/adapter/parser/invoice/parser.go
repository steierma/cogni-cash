// Package invoice provides a file-to-text extractor for invoice documents.
// Supported formats: PDF (via ledongthuc/pdf).
// For image-based PDFs or unsupported formats the caller should fall back to
// the LLM multimodal path.
package invoice

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

// Parser implements port.InvoiceParser by extracting plain text from files.
type Parser struct{}

// NewParser constructs a new invoice file parser.
func NewParser() *Parser { return &Parser{} }

// Extract reads the file at filePath and returns its plain-text content.
// mimeType is used as a hint; the file extension takes precedence.
func (p *Parser) Extract(_ context.Context, filePath, _ string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".pdf":
		return extractPDF(filePath)
	default:
		return "", fmt.Errorf("invoice parser: unsupported file type %q", ext)
	}
}

// extractPDF pulls all readable text out of a PDF file using ledongthuc/pdf.
func extractPDF(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("invoice parser: open pdf: %w", err)
	}
	defer f.Close()

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

