package invoice

import (
	"context"
	"fmt"
	"strings"

	"cogni-cash/internal/adapter/parser/pdfutil"

	"github.com/google/uuid"
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
		return pdfutil.ExtractText(fileBytes)
	}

	// Images: return empty string — caller will use multimodal LLM
	if supportedImageMIMETypes[normalised] {
		return "", nil
	}

	return "", fmt.Errorf("invoice parser: unsupported mime type %q", mimeType)
}
