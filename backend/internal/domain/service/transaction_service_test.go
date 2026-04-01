package service_test

import (
	"context"
	"testing"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
)

func TestTransactionService_ListTransactions(t *testing.T) {
	repo := &mockRepo{
		existingTxns: []entity.Transaction{
			{ContentHash: "h1", Description: "Tx 1"},
			{ContentHash: "h2", Description: "Tx 2"},
		},
	}
	svc := service.NewTransactionService(repo, nil, nil, nil, setupLogger())

	filter := entity.TransactionFilter{UserID: uuid.New()}
	txns, err := svc.ListTransactions(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txns) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(txns))
	}
}

func TestTransactionService_UpdateCategory(t *testing.T) {
	repo := &mockRepo{
		existingTxns: []entity.Transaction{
			{ContentHash: "h1", Description: "Tx 1"},
		},
	}
	svc := service.NewTransactionService(repo, nil, nil, nil, setupLogger())

	catID := uuid.New()
	userID := uuid.New()
	err := svc.UpdateCategory(context.Background(), "h1", &catID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.existingTxns[0].CategoryID == nil || *repo.existingTxns[0].CategoryID != catID {
		t.Errorf("expected category ID %v, got %v", catID, repo.existingTxns[0].CategoryID)
	}
}

func TestTransactionService_MarkAsReviewed(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewTransactionService(repo, nil, nil, nil, setupLogger())

	err := svc.MarkAsReviewed(context.Background(), "h1", uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
