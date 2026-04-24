package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/service"
)

// --- Mocks ---

type mockPTRepo struct {
	mock.Mock
}

func (m *mockPTRepo) Create(ctx context.Context, pt *entity.PlannedTransaction) error {
	args := m.Called(ctx, pt)
	return args.Error(0)
}

func (m *mockPTRepo) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.PlannedTransaction, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PlannedTransaction), args.Error(1)
}

func (m *mockPTRepo) Update(ctx context.Context, pt *entity.PlannedTransaction) error {
	args := m.Called(ctx, pt)
	return args.Error(0)
}

func (m *mockPTRepo) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *mockPTRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.PlannedTransaction), args.Error(1)
}

func (m *mockPTRepo) FindPendingByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.PlannedTransaction), args.Error(1)
}

// --- Tests ---

func TestPlannedTransactionService_UpdatePreservation(t *testing.T) {
	ctx := context.Background()
	repo := new(mockPTRepo)
	svc := service.NewPlannedTransactionService(repo)

	ptID := uuid.New()
	userID := uuid.New()
	bankID := uuid.New()

	existing := &entity.PlannedTransaction{
		ID:            ptID,
		UserID:        userID,
		Amount:        100.0,
		Currency:      "EUR",
		Status:        entity.PlannedTransactionStatusPending,
		BankAccountID: &bankID,
		CreatedAt:     time.Now().Add(-1 * time.Hour),
	}

	t.Run("Preserve fields when updating with partial data", func(t *testing.T) {
		updateReq := &entity.PlannedTransaction{
			ID:     ptID,
			UserID: userID,
			Amount: 150.0, // only amount changed
		}

		repo.On("GetByID", ctx, ptID, userID).Return(existing, nil)
		repo.On("Update", ctx, mock.MatchedBy(func(p *entity.PlannedTransaction) bool {
			return p.Amount == 150.0 && 
				   p.Status == entity.PlannedTransactionStatusPending && 
				   p.Currency == "EUR" && 
				   p.BankAccountID != nil && *p.BankAccountID == bankID &&
				   p.CreatedAt.Equal(existing.CreatedAt)
		})).Return(nil)

		err := svc.Update(ctx, updateReq)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

func TestPlannedTransactionService_MatchAndSpawn(t *testing.T) {
	ctx := context.Background()
	repo := new(mockPTRepo)
	svc := service.NewPlannedTransactionService(repo)

	userID := uuid.New()
	accID := uuid.New()
	ptID := uuid.New()
	date := time.Now().Truncate(24 * time.Hour)

	pt := entity.PlannedTransaction{
		ID:             ptID,
		UserID:         userID,
		Amount:         -50.0,
		Currency:       "EUR",
		BaseAmount:     -50.0,
		BaseCurrency:   "EUR",
		Date:           date,
		Status:         entity.PlannedTransactionStatusPending,
		IntervalMonths: 1,
		BankAccountID:  &accID,
	}

	tx := entity.Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		Amount:      -50.0,
		BookingDate: date,
	}

	t.Run("Match and Spawn Next Recurring", func(t *testing.T) {
		repo.On("FindPendingByUserID", ctx, userID).Return([]entity.PlannedTransaction{pt}, nil)
		
		// Expect update of current PT to 'matched'
		repo.On("Update", ctx, mock.MatchedBy(func(p *entity.PlannedTransaction) bool {
			return p.ID == ptID && p.Status == entity.PlannedTransactionStatusMatched && *p.MatchedTransactionID == tx.ID
		})).Return(nil)

		// Expect creation of next recurring instance
		repo.On("Create", ctx, mock.MatchedBy(func(p *entity.PlannedTransaction) bool {
			expectedDate := date.AddDate(0, 1, 0)
			return p.Date.Equal(expectedDate) && 
				   p.Amount == -50.0 && 
				   p.Currency == "EUR" &&
				   p.BankAccountID != nil && *p.BankAccountID == accID &&
				   p.Status == entity.PlannedTransactionStatusPending
		})).Return(nil)

		err := svc.MatchTransactions(ctx, userID, []entity.Transaction{tx})
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

func TestPlannedTransactionService_CRUD(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	ptID := uuid.New()

	t.Run("Create", func(t *testing.T) {
		repo := new(mockPTRepo)
		svc := service.NewPlannedTransactionService(repo)
		pt := &entity.PlannedTransaction{
			UserID:      userID, 
			Amount:      100, 
			Description: "Valid PT",
			Date:        time.Now(),
		}
		repo.On("Create", ctx, pt).Return(nil)
		err := svc.Create(ctx, pt)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("Delete", func(t *testing.T) {
		repo := new(mockPTRepo)
		svc := service.NewPlannedTransactionService(repo)
		repo.On("Delete", ctx, ptID, userID).Return(nil)
		err := svc.Delete(ctx, ptID, userID)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("FindByUserID", func(t *testing.T) {
		repo := new(mockPTRepo)
		svc := service.NewPlannedTransactionService(repo)
		repo.On("FindByUserID", ctx, userID).Return([]entity.PlannedTransaction{{ID: ptID}}, nil)
		pts, err := svc.FindByUserID(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, pts, 1)
		repo.AssertExpectations(t)
	})
}
