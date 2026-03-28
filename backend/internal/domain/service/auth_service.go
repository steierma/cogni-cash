package service

import (
	"cogni-cash/internal/domain/entity"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"cogni-cash/internal/domain/port"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = entity.ErrInvalidCredentials

type AuthService struct {
	repo      port.UserRepository
	jwtSecret []byte
	logger    *slog.Logger
}

func NewAuthService(repo port.UserRepository, secret string, logger *slog.Logger) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: []byte(secret),
		logger:    logger,
	}
}

// Login verifies credentials and returns a signed JWT.
func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	if s.repo == nil {
		return "", errors.New("user repository not available")
	}

	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		s.logger.Warn("Login failed: user not found", "username", username)
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.Warn("Login failed: invalid password", "username", username)
		return "", ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		s.logger.Error("Failed to sign JWT", "error", err)
		return "", err
	}

	s.logger.Info("User logged in successfully", "username", username, "user_id", user.ID)
	return tokenString, nil
}

// ValidateToken parses and validates a JWT, returning the user ID (subject).
func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	sub, err := claims.GetSubject()
	if err != nil {
		return "", errors.New("missing subject in token")
	}

	s.logger.Info("JWT token validated", "user_id", sub)
	return sub, nil
}


func (s *AuthService) ChangePassword(ctx context.Context, userIDStr string, oldPassword, newPassword string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	// 1. Fetch the user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	// 2. Verify the old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		s.logger.Warn("Password change failed: invalid old password", "userID", userIDStr)
		return errors.New("incorrect current password")
	}

	// 3. Hash the new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash new password", "error", err)
		return errors.New("internal server error")
	}

	// 4. Save to the database
	if err := s.repo.UpdatePassword(ctx, userID, string(newHash)); err != nil {
		return errors.New("failed to update password")
	}

	s.logger.Info("Password updated successfully", "userID", userIDStr)
	return nil
}

// EnsureAdminUser creates or updates the admin user with the supplied
// plain-text password. It is safe to call on every startup: if the user
// already exists with the same credentials nothing changes; if the password
// differs (e.g. after a credential rotation) it is updated.
func (s *AuthService) EnsureAdminUser(ctx context.Context, username, plainPassword string) error {
	if username == "" || plainPassword == "" {
		s.logger.Warn("EnsureAdminUser: ADMIN_USERNAME or ADMIN_PASSWORD not set — skipping admin seeding")
		return nil
	}

	// Check if the admin user already exists with the correct password to
	// avoid an unnecessary bcrypt hash + DB write on every startup.
	existing, err := s.repo.FindByUsername(ctx, username)
	if err == nil {
		if bcrypt.CompareHashAndPassword([]byte(existing.PasswordHash), []byte(plainPassword)) == nil {
			s.logger.Info("Admin user already exists with correct password — skipping seed", "username", username)
			return nil
		}
		// Password changed — update only the password hash.
		newHash, hashErr := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
		if hashErr != nil {
			return fmt.Errorf("EnsureAdminUser: hash password: %w", hashErr)
		}
		if updateErr := s.repo.UpdatePassword(ctx, existing.ID, string(newHash)); updateErr != nil {
			return fmt.Errorf("EnsureAdminUser: update password: %w", updateErr)
		}
		s.logger.Info("Admin password rotated", "username", username)
		return nil
	}

	// User does not exist — create a fresh record.
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("EnsureAdminUser: hash password: %w", err)
	}

	user := entity.User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: string(hash),
		Email:        username + "@localhost",
		FullName:     "System Administrator",
		Address:      "Not Provided",
		Role:         "admin",
	}

	if err := s.repo.Upsert(ctx, user); err != nil {
		return fmt.Errorf("EnsureAdminUser: upsert: %w", err)
	}

	s.logger.Info("Admin user created", "username", username)
	return nil
}
