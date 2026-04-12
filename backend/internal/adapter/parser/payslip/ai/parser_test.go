package ai_test

import (
	"context"
	"log/slog"
	"testing"

	"cogni-cash/internal/adapter/parser/payslip/ai"
	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

type mockLLMPayslipParser struct {
	lastMime string
	lastData []byte
	payslip  entity.Payslip
	err      error
}

func (m *mockLLMPayslipParser) ParsePayslipDocument(ctx context.Context, userID uuid.UUID, mimeType string, data []byte) (entity.Payslip, error) {
	m.lastMime = mimeType
	m.lastData = data
	return m.payslip, m.err
}

func TestParser_Parse_MimeDetection(t *testing.T) {
	mockLLM := &mockLLMPayslipParser{
		payslip: entity.Payslip{GrossPay: 1000},
	}
	parser := ai.NewPayslipParser(mockLLM, slog.Default())

	userID := uuid.New()

	t.Run("Detect PNG", func(t *testing.T) {
		// PNG Magic number: 89 50 4E 47 0D 0A 1A 0A
		pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		_, err := parser.Parse(context.Background(), userID, pngData)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mockLLM.lastMime != "image/png" {
			t.Errorf("expected image/png, got %s", mockLLM.lastMime)
		}
	})

	t.Run("Detect PDF", func(t *testing.T) {
		// PDF Magic number: %PDF-
		pdfData := []byte("%PDF-1.4")
		_, err := parser.Parse(context.Background(), userID, pdfData)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mockLLM.lastMime != "application/pdf" {
			t.Errorf("expected application/pdf, got %s", mockLLM.lastMime)
		}
	})
}
