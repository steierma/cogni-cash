package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
)

// --- Reconciliation-Specific Tests ---

func TestReconciliationService_DeleteReconciliation_Success(t *testing.T) {
	reconcRepo := &mockReconciliationRepo{}
	svc := service.NewReconciliationService(&mockRepo{}, reconcRepo, setupLogger())

	// First create a reconciliation
	rec := entity.Reconciliation{
		ID:                        uuid.New(),
		SettlementTransactionHash: "hash1",
		TargetTransactionHash:     "hash2",
		Amount:                    100.0,
		ReconciledAt:              time.Now().UTC(),
	}
	reconcRepo.saved = append(reconcRepo.saved, rec)

	err := svc.DeleteReconciliation(context.Background(), rec.ID, uuid.New())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestReconciliationService_DeleteReconciliation_NilRepo(t *testing.T) {
	svc := service.NewReconciliationService(&mockRepo{}, nil, setupLogger())

	err := svc.DeleteReconciliation(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Error("expected error when reconciliation repo is nil")
	}
}

func TestReconciliationService_DeleteReconciliation_RepoError(t *testing.T) {
	reconcRepo := &mockReconciliationRepo{err: errors.New("db error")}
	svc := service.NewReconciliationService(&mockRepo{}, reconcRepo, setupLogger())

	err := svc.DeleteReconciliation(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Error("expected error when repo returns error")
	}
}

func TestReconcileStatements_SameAccount(t *testing.T) {
	// Both transactions have the same statement type — should be rejected
	sameId := uuid.New()
	tx1 := entity.Transaction{
		BookingDate:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		Amount:        -100.0,
		ContentHash:   "giro1",
		BankAccountID: &sameId,
		StatementType: entity.StatementTypeGiro,
	}
	tx2 := entity.Transaction{
		BookingDate:   time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
		Amount:        100.0,
		ContentHash:   "giro2",
		BankAccountID: &sameId,
		StatementType: entity.StatementTypeGiro,
	}

	repo := &mockRepo{existingTxns: []entity.Transaction{tx1, tx2}}
	reconcRepo := &mockReconciliationRepo{}
	svc := service.NewReconciliationService(repo, reconcRepo, setupLogger())

	_, err := svc.ReconcileStatements(context.Background(), uuid.New(), "giro1", "giro2")
	if !errors.Is(err, service.ErrSameAccount) {
		t.Errorf("expected ErrSameAccount, got %v", err)
	}
}

func TestReconcileStatements_NilReconciliationRepo(t *testing.T) {
	svc := service.NewReconciliationService(&mockRepo{}, nil, setupLogger())

	_, err := svc.ReconcileStatements(context.Background(), uuid.New(), "hash1", "hash2")
	if err == nil {
		t.Error("expected error when reconciliation repo is nil")
	}
}

func TestReconcileStatements_TargetNotFound(t *testing.T) {
	tx1 := entity.Transaction{
		BookingDate:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		Amount:        -100.0,
		ContentHash:   "exists",
		StatementType: entity.StatementTypeGiro,
	}

	repo := &mockRepo{existingTxns: []entity.Transaction{tx1}}
	reconcRepo := &mockReconciliationRepo{}
	svc := service.NewReconciliationService(repo, reconcRepo, setupLogger())

	_, err := svc.ReconcileStatements(context.Background(), uuid.New(), "exists", "missing")
	if err == nil {
		t.Error("expected error for missing target transaction")
	}
}

func TestReconciliationService_SuggestReconciliations_DefaultWindow(t *testing.T) {
	// Pass 0 for matchWindowDays — should default to 7
	repo := &mockRepo{
		existingTxns: []entity.Transaction{
			{BookingDate: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC), Amount: -200, ContentHash: "giro1", StatementType: entity.StatementTypeGiro},
			{BookingDate: time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC), Amount: 200, ContentHash: "cc1", StatementType: entity.StatementTypeCreditCard},
		},
	}
	svc := service.NewReconciliationService(repo, nil, setupLogger())

	suggestions, err := svc.SuggestReconciliations(context.Background(), uuid.New(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 1 {
		t.Errorf("expected 1 suggestion with default window, got %d", len(suggestions))
	}
}

func TestReconciliationService_SuggestReconciliations_SkipsSameAccount(t *testing.T) {
	// Both are giro — should not match
	sameUuid := uuid.New()
	repo := &mockRepo{
		existingTxns: []entity.Transaction{
			{BookingDate: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC), Amount: -200, ContentHash: "g1", StatementType: entity.StatementTypeGiro, BankAccountID: &sameUuid},
			{BookingDate: time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC), Amount: 200, ContentHash: "g2", StatementType: entity.StatementTypeGiro, BankAccountID: &sameUuid},
		},
	}
	svc := service.NewReconciliationService(repo, nil, setupLogger())

	suggestions, err := svc.SuggestReconciliations(context.Background(), uuid.New(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions for same statement type, got %d", len(suggestions))
	}
}

func TestReconciliationService_NilLogger(t *testing.T) {
	// Should not panic with nil logger
	svc := service.NewReconciliationService(&mockRepo{}, &mockReconciliationRepo{}, nil)
	_, err := svc.SuggestReconciliations(context.Background(), uuid.New(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
