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

type MockPlannedTransactionRepository struct {
	mock.Mock
}

func (m *MockPlannedTransactionRepository) Create(ctx context.Context, pt *entity.PlannedTransaction) error {
	args := m.Called(ctx, pt)
	return args.Error(0)
}

func (m *MockPlannedTransactionRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.PlannedTransaction, error) {
	args := m.Called(ctx, id, userID)
	if pt := args.Get(0); pt != nil {
		return pt.(*entity.PlannedTransaction), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPlannedTransactionRepository) Update(ctx context.Context, pt *entity.PlannedTransaction) error {
	args := m.Called(ctx, pt)
	return args.Error(0)
}

func (m *MockPlannedTransactionRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockPlannedTransactionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error) {
	args := m.Called(ctx, userID)
	if pts := args.Get(0); pts != nil {
		return pts.([]entity.PlannedTransaction), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPlannedTransactionRepository) FindPendingByUserID(ctx context.Context, userID uuid.UUID) ([]entity.PlannedTransaction, error) {
	args := m.Called(ctx, userID)
	if pts := args.Get(0); pts != nil {
		return pts.([]entity.PlannedTransaction), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestPlannedTransactionService_Create(t *testing.T) {
	repo := new(MockPlannedTransactionRepository)
	svc := service.NewPlannedTransactionService(repo)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		pt := &entity.PlannedTransaction{
			UserID:      userID,
			Amount:      100.0,
			Date:        time.Now(),
			Description: "Test PT",
		}
		repo.On("Create", ctx, mock.AnythingOfType("*entity.PlannedTransaction")).Return(nil).Once()

		err := svc.Create(ctx, pt)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, pt.ID)
		assert.Equal(t, entity.PlannedTransactionStatusPending, pt.Status)
		repo.AssertExpectations(t)
	})

	t.Run("validation_error_no_user", func(t *testing.T) {
		pt := &entity.PlannedTransaction{
			Amount:      100.0,
			Description: "Test PT",
		}
		err := svc.Create(ctx, pt)
		assert.ErrorIs(t, err, entity.ErrInvalidPlannedTransaction)
	})
	
	t.Run("validation_error_no_amount", func(t *testing.T) {
		pt := &entity.PlannedTransaction{
			UserID:      userID,
			Description: "Test PT",
		}
		err := svc.Create(ctx, pt)
		assert.ErrorIs(t, err, entity.ErrInvalidPlannedTransaction)
	})
	
	t.Run("validation_error_no_description", func(t *testing.T) {
		pt := &entity.PlannedTransaction{
			UserID: userID,
			Amount: 100.0,
		}
		err := svc.Create(ctx, pt)
		assert.ErrorIs(t, err, entity.ErrInvalidPlannedTransaction)
	})
}

func TestPlannedTransactionService_Update(t *testing.T) {
	repo := new(MockPlannedTransactionRepository)
	svc := service.NewPlannedTransactionService(repo)
	ctx := context.Background()
	userID := uuid.New()
	ptID := uuid.New()

	t.Run("success", func(t *testing.T) {
		createdAt := time.Now().Add(-time.Hour)
		existingPT := &entity.PlannedTransaction{
			ID:          ptID,
			UserID:      userID,
			Amount:      50.0,
			Description: "Old",
			CreatedAt:   createdAt,
		}
		
		pt := &entity.PlannedTransaction{
			ID:          ptID,
			UserID:      userID,
			Amount:      150.0,
			Description: "New",
			CreatedAt:   time.Now(), // Should be overridden
		}
		
		repo.On("GetByID", ctx, ptID, userID).Return(existingPT, nil).Once()
		repo.On("Update", ctx, pt).Return(nil).Once()

		err := svc.Update(ctx, pt)
		assert.NoError(t, err)
		assert.Equal(t, createdAt, pt.CreatedAt)
		repo.AssertExpectations(t)
	})
	
	t.Run("not_found", func(t *testing.T) {
		pt := &entity.PlannedTransaction{
			ID:     ptID,
			UserID: userID,
		}
		
		repo.On("GetByID", ctx, ptID, userID).Return(nil, entity.ErrPlannedTransactionNotFound).Once()

		err := svc.Update(ctx, pt)
		assert.ErrorIs(t, err, entity.ErrPlannedTransactionNotFound)
		repo.AssertExpectations(t)
	})
}

func TestPlannedTransactionService_Delete(t *testing.T) {
	repo := new(MockPlannedTransactionRepository)
	svc := service.NewPlannedTransactionService(repo)
	ctx := context.Background()
	userID := uuid.New()
	ptID := uuid.New()

	repo.On("Delete", ctx, ptID, userID).Return(nil).Once()

	err := svc.Delete(ctx, ptID, userID)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestPlannedTransactionService_FindByUserID(t *testing.T) {
	repo := new(MockPlannedTransactionRepository)
	svc := service.NewPlannedTransactionService(repo)
	ctx := context.Background()
	userID := uuid.New()

	expected := []entity.PlannedTransaction{
		{ID: uuid.New(), UserID: userID},
	}
	repo.On("FindByUserID", ctx, userID).Return(expected, nil).Once()

	result, err := svc.FindByUserID(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}
