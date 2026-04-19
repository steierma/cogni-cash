package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
)

type CategoryService struct {
	repo        port.CategoryRepository
	sharingRepo port.SharingRepository
	logger      *slog.Logger
}

func NewCategoryService(repo port.CategoryRepository, sharingRepo port.SharingRepository, logger *slog.Logger) *CategoryService {
	return &CategoryService{
		repo:        repo,
		sharingRepo: sharingRepo,
		logger:      logger,
	}
}

func (s *CategoryService) GetAll(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
	return s.repo.FindAll(ctx, userID)
}

func (s *CategoryService) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.Category, error) {
	return s.repo.FindByID(ctx, id, userID)
}

func (s *CategoryService) Create(ctx context.Context, cat entity.Category) (entity.Category, error) {
	return s.repo.Save(ctx, cat)
}

func (s *CategoryService) Update(ctx context.Context, cat entity.Category) (entity.Category, error) {
	return s.repo.Update(ctx, cat)
}

func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *CategoryService) ShareCategory(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID, permission string) error {
	if ownerID == sharedWithID {
		return fmt.Errorf("cannot share category with yourself")
	}

	// Verify ownership
	cat, err := s.repo.FindByID(ctx, categoryID, ownerID)
	if err != nil {
		return fmt.Errorf("category not found or not owned by you: %w", err)
	}
	if cat.UserID != ownerID {
		return fmt.Errorf("only the owner can share a category")
	}

	return s.sharingRepo.ShareCategory(ctx, categoryID, ownerID, sharedWithID, permission)
}

func (s *CategoryService) RevokeShare(ctx context.Context, categoryID, ownerID, sharedWithID uuid.UUID) error {
	// Verify ownership or being the person it was shared with
	cat, err := s.repo.FindByID(ctx, categoryID, ownerID)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// Either the owner or the recipient can revoke/remove the share
	if cat.UserID != ownerID && ownerID != sharedWithID {
		return fmt.Errorf("unauthorized to revoke this share")
	}

	return s.sharingRepo.RevokeShare(ctx, categoryID, cat.UserID, sharedWithID)
}

func (s *CategoryService) ListShares(ctx context.Context, categoryID, ownerID uuid.UUID) ([]uuid.UUID, error) {
	// Verify ownership: only the owner can list who they shared with
	_, err := s.repo.FindByID(ctx, categoryID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("category not found or unauthorized: %w", err)
	}

	return s.sharingRepo.ListShares(ctx, categoryID, ownerID)
}
