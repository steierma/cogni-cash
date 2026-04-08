package service_test

import (
	"context"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
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

func TestTransactionService_AutoCategorize_HybridMatching(t *testing.T) {
	catGroceries := uuid.New()
	userID := uuid.New()

	tx1 := entity.Transaction{ContentHash: "h1", Description: "REWE", Amount: -10.0} // Will match in DB
	tx2 := entity.Transaction{ContentHash: "h2", Description: "New Vendor", Amount: -20.0} // Will NOT match in DB

	repo := &mockRepo{
		existingTxns: []entity.Transaction{tx1, tx2},
		findMatchingID: &catGroceries, // Simulate DB match for all queries
	}

	// We override FindMatchingCategory to only match 'h1'
	repo.findMatchFunc = func(txn port.TransactionToCategorize) *uuid.UUID {
		if txn.Hash == "h1" {
			return &catGroceries
		}
		return nil
	}

	catRepo := &mockCategoryRepo{
		saved: []entity.Category{
			{ID: catGroceries, Name: "Groceries"},
		},
	}

	llm := &mockCategorizer{
		results: []port.CategorizedTransaction{
			{Hash: "h2", Category: "Groceries"},
		},
	}

	svc := service.NewTransactionService(repo, catRepo, nil, llm, setupLogger())

	err := svc.StartAutoCategorizeAsync(context.Background(), userID, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait for completion
	timeout := time.After(1 * time.Second)
	for {
		if svc.GetJobStatus().Status == "completed" {
			break
		}
		select {
		case <-timeout:
			t.Fatal("timed out")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Verify LLM was only called for h2
	if llm.calls != 1 {
		t.Errorf("expected 1 LLM call, got %d", llm.calls)
	}

	// Verify categories updated in repo
	if repo.existingTxns[0].CategoryID == nil || *repo.existingTxns[0].CategoryID != catGroceries {
		t.Error("h1 category not updated from DB match")
	}
	if repo.existingTxns[1].CategoryID == nil || *repo.existingTxns[1].CategoryID != catGroceries {
		t.Error("h2 category not updated from LLM result")
	}
}

