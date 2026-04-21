package service_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"cogni-cash/internal/domain/service"
)

type mockCatSvcRepo struct {
	port.CategoryRepository
	FindAllFunc  func(ctx context.Context, userID uuid.UUID) ([]entity.Category, error)
	FindByIDFunc func(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Category, error)
	SaveFunc     func(ctx context.Context, cat entity.Category) (entity.Category, error)
	UpdateFunc   func(ctx context.Context, cat entity.Category) (entity.Category, error)
	DeleteFunc   func(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

func (m *mockCatSvcRepo) FindAll(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
	return m.FindAllFunc(ctx, userID)
}
func (m *mockCatSvcRepo) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Category, error) {
	return m.FindByIDFunc(ctx, id, userID)
}
func (m *mockCatSvcRepo) Save(ctx context.Context, cat entity.Category) (entity.Category, error) {
	return m.SaveFunc(ctx, cat)
}
func (m *mockCatSvcRepo) Update(ctx context.Context, cat entity.Category) (entity.Category, error) {
	return m.UpdateFunc(ctx, cat)
}
func (m *mockCatSvcRepo) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return m.DeleteFunc(ctx, id, userID)
}

type mockCatSvcSharingRepo struct {
	port.SharingRepository
	ShareCategoryFunc func(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID, permission string) error
	RevokeShareFunc   func(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID) error
	ListSharesFunc    func(ctx context.Context, categoryID, ownerID uuid.UUID) ([]uuid.UUID, error)
}

func (m *mockCatSvcSharingRepo) ShareCategory(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID, permission string) error {
	return m.ShareCategoryFunc(ctx, categoryID, ownerID, sharedWithID, permission)
}
func (m *mockCatSvcSharingRepo) RevokeShare(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID) error {
	return m.RevokeShareFunc(ctx, categoryID, ownerID, sharedWithID)
}
func (m *mockCatSvcSharingRepo) ListShares(ctx context.Context, categoryID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	return m.ListSharesFunc(ctx, categoryID, ownerID)
}

func TestCategoryService(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	userID := uuid.New()
	catID := uuid.New()

	t.Run("CRUD operations", func(t *testing.T) {
		repo := &mockCatSvcRepo{
			FindAllFunc: func(ctx context.Context, uID uuid.UUID) ([]entity.Category, error) {
				return []entity.Category{{ID: catID, Name: "Test"}}, nil
			},
			FindByIDFunc: func(ctx context.Context, id, uID uuid.UUID) (entity.Category, error) {
				return entity.Category{ID: id, Name: "Test"}, nil
			},
			SaveFunc: func(ctx context.Context, cat entity.Category) (entity.Category, error) {
				return cat, nil
			},
		}

		svc := service.NewCategoryService(repo, nil, logger)

		cats, err := svc.GetAll(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, cats, 1)

		cat, err := svc.GetByID(ctx, catID, userID)
		require.NoError(t, err)
		assert.Equal(t, "Test", cat.Name)

		_, err = svc.Create(ctx, entity.Category{Name: "New"})
		require.NoError(t, err)
	})

	t.Run("Sharing operations", func(t *testing.T) {
		sharedWithID := uuid.New()
		repo := &mockCatSvcRepo{
			FindByIDFunc: func(ctx context.Context, id, uID uuid.UUID) (entity.Category, error) {
				return entity.Category{ID: id, UserID: uID, Name: "Test"}, nil
			},
		}

		sharingRepo := &mockCatSvcSharingRepo{
			ShareCategoryFunc: func(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID, permission string) error {
				return nil
			},
			ListSharesFunc: func(ctx context.Context, categoryID, ownerID uuid.UUID) ([]uuid.UUID, error) {
				return []uuid.UUID{sharedWithID}, nil
			},
		}

		svc := service.NewCategoryService(repo, sharingRepo, logger)

		err := svc.ShareCategory(ctx, catID, userID, sharedWithID, "view")
		require.NoError(t, err)

		shares, err := svc.ListShares(ctx, catID, userID)
		require.NoError(t, err)
		assert.Len(t, shares, 1)
	})
}
