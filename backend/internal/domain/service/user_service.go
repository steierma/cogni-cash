package service

import (
	"context"
	"errors"
	"log/slog"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo   port.UserRepository
	logger *slog.Logger
}

func NewUserService(repo port.UserRepository, logger *slog.Logger) *UserService {
	if logger == nil {
		logger = slog.Default()
	}
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}

func (s *UserService) ListUsers(ctx context.Context, search string) ([]entity.User, error) {
	return s.repo.FindAll(ctx, search)
}

func (s *UserService) GetUser(ctx context.Context, idStr string) (entity.User, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return entity.User{}, errors.New("invalid user ID")
	}
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, req entity.User, plainPassword string) (entity.User, error) {
	if plainPassword == "" {
		return entity.User{}, errors.New("password is required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return entity.User{}, errors.New("failed to hash password")
	}

	req.ID = uuid.New()
	req.PasswordHash = string(hash)

	// Set a default role if none provided
	if req.Role == "" {
		req.Role = "manager"
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return entity.User{}, err // Let the HTTP layer handle unique constraint errors
	}

	s.logger.Info("New user created", "username", req.Username, "id", req.ID, "role", req.Role)
	return req, nil
}

// UpdateUser updates profile information. Note: Password changes should remain in AuthService.
func (s *UserService) UpdateUser(ctx context.Context, idStr string, updates entity.User) (entity.User, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return entity.User{}, errors.New("invalid user ID")
	}

	existingUser, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return entity.User{}, err
	}

	// Apply updates
	existingUser.Username = updates.Username
	existingUser.Email = updates.Email
	existingUser.FullName = updates.FullName
	existingUser.Address = updates.Address
	existingUser.Role = updates.Role

	if err := s.repo.Update(ctx, existingUser); err != nil {
		return entity.User{}, err
	}

	s.logger.Info("User profile updated", "id", id, "username", existingUser.Username)
	return existingUser, nil
}

func (s *UserService) DeleteUser(ctx context.Context, idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return errors.New("invalid user ID")
	}
	s.logger.Info("Deleting user", "id", id)
	return s.repo.Delete(ctx, id)
}
