package port

import "context"

// FileExtractor is the output port for document text extraction.
type FileExtractor interface {
	Extract(ctx context.Context, filePath string) (string, error)
}

