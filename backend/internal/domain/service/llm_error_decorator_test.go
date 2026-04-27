package service_test

import (
	"cogni-cash/internal/domain/port"
	"cogni-cash/internal/domain/port/mock"
	"cogni-cash/internal/domain/service"
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	testifymock "github.com/stretchr/testify/mock"
)

type mockLLMAdapter struct {
	port.BankStatementAIParser
	port.PayslipAIParser
	port.InvoiceAICategorizer
	port.TransactionCategorizer
	port.DocumentAIParser
	port.SubscriptionEnricher
	port.CancellationLetterGenerator

	errToReturn error
}

func (m *mockLLMAdapter) CategorizeInvoice(ctx context.Context, userID uuid.UUID, req port.CategorizationRequest) (port.InvoiceCategorizationResult, error) {
	return port.InvoiceCategorizationResult{}, m.errToReturn
}

func TestLLMNotifierDecorator_AlertsOnCriticalErrors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockNotifier := new(mock.MockNotificationUseCase)
	
	// Expect alert to be called
	mockNotifier.On("SendAdminAlert", testifymock.Anything, "Cogni-Cash LLM Alert: CategorizeInvoice", "groq API error: rate limit exceeded (status: 429)").Return(nil)

	adapter := &mockLLMAdapter{
		errToReturn: errors.New("groq API error: rate limit exceeded (status: 429)"),
	}

	decorator := service.NewLLMNotifierDecorator(adapter, mockNotifier, logger)

	_, err := decorator.CategorizeInvoice(context.Background(), uuid.New(), port.CategorizationRequest{})
	if err == nil {
		t.Fatal("Expected error to be propagated")
	}

	// Give goroutine time to execute
	time.Sleep(50 * time.Millisecond)

	mockNotifier.AssertExpectations(t)
}

func TestLLMNotifierDecorator_IgnoresNonCriticalErrors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockNotifier := new(mock.MockNotificationUseCase)
	
	// Should not be called
	// No mockNotifier.On("SendAdminAlert", ...) setup

	adapter := &mockLLMAdapter{
		errToReturn: errors.New("some business logic error"),
	}

	decorator := service.NewLLMNotifierDecorator(adapter, mockNotifier, logger)

	_, err := decorator.CategorizeInvoice(context.Background(), uuid.New(), port.CategorizationRequest{})
	if err == nil {
		t.Fatal("Expected error to be propagated")
	}

	// Give goroutine time to execute (just in case it was called incorrectly)
	time.Sleep(50 * time.Millisecond)

	mockNotifier.AssertExpectations(t)
}
