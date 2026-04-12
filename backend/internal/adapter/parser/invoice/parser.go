// Package invoice provides a file-to-text extractor for invoice documents.
// Supported formats: PDF (via ledongthuc/pdf), JPEG, PNG, GIF, WEBP.
// For image files, Extract returns an empty string — the caller (InvoiceService)
// detects this and falls back to the LLM multimodal path via CategorizeInvoiceImage.
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

// supportedImageMIMETypes lists all image MIME types accepted by the parser.
// The parser returns an empty string for these so that InvoiceService falls
// back to the multimodal LLM path (CategorizeInvoiceImage).
var supportedImageMIMETypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// Parser implements port.InvoiceParser by extracting plain text from files.
type Parser struct{}

// NewParser constructs a new invoice file parser.
func NewParser() *Parser { return &Parser{} }

// Extract reads the file content and returns its plain-text content.
// mimeType is used as a hint to determine the extraction strategy.
// For image files (JPEG, PNG, GIF, WEBP) it intentionally returns an empty
// string without an error, signalling to the caller to use the multimodal LLM path.
func (p *Parser) Extract(_ context.Context, _ uuid.UUID, fileBytes []byte, mimeType string) (string, error) {
	normalised := strings.ToLower(strings.TrimSpace(mimeType))

	// PDF: extract text directly
	if strings.Contains(normalised, "application/pdf") {
		return extractPDF(fileBytes)
	}

	// Images: return empty string — caller will use multimodal LLM
	if supportedImageMIMETypes[normalised] {
		return "", nil
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
