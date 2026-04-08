package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"log/slog"

	"github.com/google/uuid"
)

var ErrTransactionNotFound = entity.ErrTransactionNotFound
var ErrSameAccount = entity.ErrSameAccount

type ReconciliationService struct {
	transactionRepo    port.BankStatementRepository
	reconciliationRepo port.ReconciliationRepository
	logger             *slog.Logger
}

func NewReconciliationService(txRepo port.BankStatementRepository, recRepo port.ReconciliationRepository, logger *slog.Logger) *ReconciliationService {
	if logger == nil {
		logger = slog.Default()
	}
	return &ReconciliationService{
		transactionRepo:    txRepo,
		reconciliationRepo: recRepo,
		logger:             logger,
	}
}

func (s *ReconciliationService) ReconcileStatements(
	ctx context.Context,
	userID uuid.UUID,
	settlementTxHash string,
	targetTxHash string,
) (entity.Reconciliation, error) {
	if s.reconciliationRepo == nil {
		return entity.Reconciliation{}, errors.New("reconciliation service: repository not configured")
	}

	f := false
	txns, err := s.transactionRepo.SearchTransactions(ctx, entity.TransactionFilter{UserID: userID, IsReconciled: &f})
	if err != nil {
		return entity.Reconciliation{}, fmt.Errorf("reconciliation service: resolve transactions: %w", err)
	}

	var settlementTx, targetTx *entity.Transaction
	for i := range txns {
		if txns[i].ContentHash == settlementTxHash {
			settlementTx = &txns[i]
		}
		if txns[i].ContentHash == targetTxHash {
			targetTx = &txns[i]
		}
	}

	if settlementTx == nil {
		return entity.Reconciliation{}, ErrTransactionNotFound
	}
	if targetTx == nil {
		return entity.Reconciliation{}, errors.New("reconciliation service: target transaction not found")
	}

	// Prevent reconciling two transactions from the exact same bank account or statement.
	if settlementTx.BankAccountID != nil && targetTx.BankAccountID != nil && *settlementTx.BankAccountID == *targetTx.BankAccountID {
		return entity.Reconciliation{}, ErrSameAccount
	}
	if settlementTx.BankStatementID != nil && targetTx.BankStatementID != nil && *settlementTx.BankStatementID == *targetTx.BankStatementID {
		return entity.Reconciliation{}, ErrSameAccount
	}

	rec := entity.Reconciliation{
		ID:                        uuid.New(),
		UserID:                    userID,
		SettlementTransactionHash: settlementTxHash,
		TargetTransactionHash:     targetTxHash,
		Amount:                    math.Abs(settlementTx.Amount),
		ReconciledAt:              time.Now().UTC(),
	}

	saved, err := s.reconciliationRepo.Save(ctx, rec)
	if err != nil {
		return entity.Reconciliation{}, fmt.Errorf("reconciliation service: save reconciliation: %w", err)
	}

	s.logger.Info("1:1 Reconciliation created",
		"reconciliation_id", saved.ID,
		"settlement_tx_hash", settlementTxHash,
		"target_tx_hash", targetTxHash,
		"user_id", userID,
	)
	return saved, nil
}

func (s *ReconciliationService) SuggestReconciliations(ctx context.Context, userID uuid.UUID, matchWindowDays int) ([]entity.ReconciliationPairSuggestion, error) {
	if matchWindowDays <= 0 {
		matchWindowDays = 7
	}

	f := false
	allTxns, err := s.transactionRepo.FindTransactions(ctx, entity.TransactionFilter{
		UserID:       userID,
		IsReconciled: &f,
	})
	if err != nil {
		return nil, fmt.Errorf("find candidate txns: %w", err)
	}

	s.logger.Info("Searching for reconciliation suggestions", "candidate_count", len(allTxns), "user_id", userID)
	var suggestions []entity.ReconciliationPairSuggestion
	usedHashes := make(map[string]bool)

	var debits, credits []entity.Transaction
	for _, tx := range allTxns {
		if tx.Amount < 0 {
			debits = append(debits, tx)
		} else if tx.Amount > 0 {
			credits = append(credits, tx)
		}
	}

	for _, debit := range debits {
		if usedHashes[debit.ContentHash] {
			continue
		}

		for _, credit := range credits {
			if usedHashes[credit.ContentHash] {
				continue
			}

			// Skip if both transactions are from the exact same bank account or statement.
			if debit.BankAccountID != nil && credit.BankAccountID != nil && *debit.BankAccountID == *credit.BankAccountID {
				continue
			}
			if debit.BankStatementID != nil && credit.BankStatementID != nil && *debit.BankStatementID == *credit.BankStatementID {
				continue
			}

			// Skip if the target (credit) date is before the source (debit) date.
			// A settlement payment always debits first; the credit appears later (or same day).
			if credit.BookingDate.Before(debit.BookingDate) {
				continue
			}

			if math.Abs(debit.Amount+credit.Amount) < 0.01 {
				diffHours := math.Abs(credit.BookingDate.Sub(debit.BookingDate).Hours())

				if diffHours <= float64(24*matchWindowDays) {
					suggestions = append(suggestions, entity.ReconciliationPairSuggestion{
						SourceTransaction: debit,
						TargetTransaction: credit,
						MatchScore:        1.0 - (diffHours / float64(24*matchWindowDays)),
					})

					usedHashes[debit.ContentHash] = true
					usedHashes[credit.ContentHash] = true
					break
				}
			}
		}
	}

	return suggestions, nil
}

func (s *ReconciliationService) DeleteReconciliation(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if s.reconciliationRepo == nil {
		return errors.New("reconciliation service: repository not configured")
	}

	err := s.reconciliationRepo.Delete(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("reconciliation service: delete reconciliation: %w", err)
	}

	s.logger.Info("1:1 Reconciliation deleted", "reconciliation_id", id, "user_id", userID)
	return nil
}
