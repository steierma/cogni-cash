package service

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
)

type plannedTransactionService struct {
	repo port.PlannedTransactionRepository
}

// NewPlannedTransactionService creates a new instance of PlannedTransactionUseCase.
func NewPlannedTransactionService(repo port.PlannedTransactionRepository) port.PlannedTransactionUseCase {
	return &plannedTransactionService{repo: repo}
}

func (s *plannedTransactionService) Create(ctx context.Context, pt *entity.PlannedTransaction) error {
	if pt.ID == uuid.Nil {
		pt.ID = uuid.New()
	}
	if pt.UserID == uuid.Nil {
		return entity.ErrInvalidPlannedTransaction
	}
	if pt.Amount == 0 {
		return entity.ErrInvalidPlannedTransaction
	}
	if pt.Description == "" {
		return entity.ErrInvalidPlannedTransaction
	}
	pt.Status = entity.PlannedTransactionStatusPending
	pt.CreatedAt = time.Now().UTC()

	return s.repo.Create(ctx, pt)
}

func (s *plannedTransactionService) Update(ctx context.Context, pt *entity.PlannedTransaction) error {
	existing, err := s.repo.GetByID(ctx, pt.ID, pt.UserID)
	if err != nil {
		return err
	}

	// Preserve fields that shouldn't change
	pt.CreatedAt = existing.CreatedAt

	return s.repo.Update(ctx, pt)
}

func (s *plannedTransactionService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *plannedTransactionService) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *plannedTransactionService) MatchTransactions(ctx context.Context, userID uuid.UUID, txns []entity.Transaction) error {
	pending, err := s.repo.FindPendingByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if len(pending) == 0 {
		return nil
	}

	for _, tx := range txns {
		for i := range pending {
			pt := &pending[i]
			if pt.Status != entity.PlannedTransactionStatusPending {
				continue
			}

			// Amount check (±5% tolerance)
			diff := math.Abs(pt.Amount - tx.Amount)
			tolerance := math.Abs(pt.Amount * 0.05)
			if diff > tolerance {
				continue
			}

			// Date check (±7 days)
			dateDiff := tx.BookingDate.Sub(pt.Date)
			if dateDiff < 0 {
				dateDiff = -dateDiff
			}
			if dateDiff > 7*24*time.Hour {
				continue
			}

			// Match found!
			pt.Status = entity.PlannedTransactionStatusMatched
			pt.MatchedTransactionID = &tx.ID
			if err := s.repo.Update(ctx, pt); err != nil {
				return err
			}

			// Rolling Head: Spawn next instance if recurring
			if pt.IntervalMonths > 0 {
				nextDate := pt.Date.AddDate(0, pt.IntervalMonths, 0)

				// Only spawn if not past end date
				if pt.EndDate == nil || !nextDate.After(*pt.EndDate) {
					nextPT := &entity.PlannedTransaction{
						ID:             uuid.New(),
						UserID:         pt.UserID,
						Amount:         pt.Amount,
						Date:           nextDate,
						Description:    pt.Description,
						CategoryID:     pt.CategoryID,
						Status:         entity.PlannedTransactionStatusPending,
						IntervalMonths: pt.IntervalMonths,
						EndDate:        pt.EndDate,
						CreatedAt:      time.Now().UTC(),
					}
					if err := s.repo.Create(ctx, nextPT); err != nil {
						return err
					}
				}
			}

			// Once matched, we don't want to match it again in this loop
			break
		}
	}
	return nil
}
