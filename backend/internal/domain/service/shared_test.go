package service_test

import (
	"context"
	"errors"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type mockCategoryRepo struct {
	saved []entity.Category
	err   error
}

func (m *mockCategoryRepo) Save(_ context.Context, cat entity.Category) (entity.Category, error) {
	if m.err != nil {
		return entity.Category{}, m.err
	}
	for i, existing := range m.saved {
		if existing.Name == cat.Name && existing.UserID == cat.UserID {
			return existing, nil
		}
		if existing.ID == cat.ID {
			m.saved[i] = cat
			return cat, nil
		}
	}
	if cat.ID == uuid.Nil {
		cat.ID = uuid.New()
	}
	m.saved = append(m.saved, cat)
	return cat, nil
}

func (m *mockCategoryRepo) FindAll(_ context.Context, _ uuid.UUID) ([]entity.Category, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.saved, nil
}

func (m *mockCategoryRepo) FindByID(_ context.Context, id uuid.UUID, _ uuid.UUID) (entity.Category, error) {
	if m.err != nil {
		return entity.Category{}, m.err
	}
	for _, c := range m.saved {
		if c.ID == id {
			return c, nil
		}
	}
	return entity.Category{}, errors.New("not found")
}

func (m *mockCategoryRepo) Update(_ context.Context, cat entity.Category) (entity.Category, error) {
	if m.err != nil {
		return entity.Category{}, m.err
	}
	for i, existing := range m.saved {
		if existing.ID == cat.ID {
			m.saved[i] = cat
			return cat, nil
		}
	}
	return cat, nil
}

func (m *mockCategoryRepo) Delete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return m.err
}

func (m *mockCategoryRepo) FindMatchingCategory(_ context.Context, _ port.TransactionToCategorize) *uuid.UUID {
	return nil
}

func (m *mockCategoryRepo) FindMatchingCategoryWithThreshold(_ context.Context, _ port.TransactionToCategorize, _ float64) *uuid.UUID {
	return nil
}

func (m *mockCategoryRepo) GetCategorizationExamples(_ context.Context, _ uuid.UUID, _ int) ([]entity.CategorizationExample, error) {
	return nil, nil
}

type mockSettingsRepo struct {
	mock.Mock
}

func (m *mockSettingsRepo) Get(ctx context.Context, key string, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, key, userID)
	return args.String(0), args.Error(1)
}
func (m *mockSettingsRepo) GetAll(ctx context.Context, userID uuid.UUID) (map[string]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(map[string]string), args.Error(1)
}
func (m *mockSettingsRepo) Set(ctx context.Context, key, value string, userID uuid.UUID, isSensitive bool) error {
	args := m.Called(ctx, key, value, userID, isSensitive)
	return args.Error(0)
}
