package port

import (
	"context"

	"github.com/google/uuid"
)

// InvoiceParser is the output port for extracting raw text from invoice files
// (PDF, image, etc.) before they are sent to the LLM for categorization.
type InvoiceParser interface {
	// Extract reads the file content and returns its plain-text content.
	// The mimeType hint may be used to choose the extraction strategy.
	Extract(ctx context.Context, userID uuid.UUID, fileBytes []byte, mimeType string) (string, error)
}
