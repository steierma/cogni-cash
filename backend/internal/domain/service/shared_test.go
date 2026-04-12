package service_test

import (
	"context"
	"errors"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
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
