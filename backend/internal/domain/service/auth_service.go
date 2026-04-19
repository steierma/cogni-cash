package service

import (
	"cogni-cash/internal/domain/entity"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cogni-cash/internal/domain/port"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = entity.ErrInvalidCredentials

type AuthService struct {
	repo            port.UserRepository
	resetRepo       port.PasswordResetRepository
	authRepo        port.AuthRepository
	notificationSvc port.NotificationUseCase
	settingsRepo    port.SettingsRepository
	jwtSecret       []byte
	logger          *slog.Logger
}

func NewAuthService(repo port.UserRepository, resetRepo port.PasswordResetRepository, authRepo port.AuthRepository, notificationSvc port.NotificationUseCase, settingsRepo port.SettingsRepository, secret string, logger *slog.Logger) *AuthService {
	return &AuthService{
		repo:            repo,
		resetRepo:       resetRepo,
		authRepo:        authRepo,
		notificationSvc: notificationSvc,
		settingsRepo:    settingsRepo,
		jwtSecret:       []byte(secret),
		logger:          logger,
	}
}

func (s *AuthService) GetRepo_ForTest() any {
	return s.repo
}

// Login verifies credentials and returns an AuthResponse (JWT + Refresh Token).
func (s *AuthService) Login(ctx context.Context, username, password string) (entity.AuthResponse, error) {
	if s.repo == nil {
		return entity.AuthResponse{}, errors.New("user repository not available")
	}

	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		s.logger.Warn("Login failed: user not found", "username", username)
		return entity.AuthResponse{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.Warn("Login failed: invalid password", "username", username)
		return entity.AuthResponse{}, ErrInvalidCredentials
	}

	// Generate Access Token (JWT)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		s.logger.Error("Failed to sign JWT", "error", err)
		return entity.AuthResponse{}, err
	}

	// Generate Refresh Token
	_, plainRT, err := s.generateRefreshToken(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", "error", err)
		return entity.AuthResponse{}, err
	}

	s.logger.Info("User logged in successfully", "username", username, "user_id", user.ID)
	return entity.AuthResponse{
		Token:        tokenString,
		RefreshToken: plainRT,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, plainRT string) (entity.AuthResponse, error) {
	hash := s.hashToken(plainRT)
	rt, err := s.authRepo.FindRefreshToken(ctx, hash)
	if err != nil {
		return entity.AuthResponse{}, errors.New("invalid refresh token")
	}

	if rt.Revoked || time.Now().After(rt.ExpiresAt) {
		return entity.AuthResponse{}, errors.New("refresh token expired or revoked")
	}

	// Token is valid. Issue new JWT.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": rt.UserID.String(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return entity.AuthResponse{}, fmt.Errorf("failed to sign JWT during refresh: %w", err)
	}

	s.logger.Info("JWT refreshed", "user_id", rt.UserID)
	return entity.AuthResponse{
		Token:        tokenString,
		RefreshToken: plainRT, // Return the same refresh token or issue a new one?
		// For simplicity, we keep the same refresh token.
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, plainRT string) error {
	hash := s.hashToken(plainRT)
	rt, err := s.authRepo.FindRefreshToken(ctx, hash)
	if err != nil {
		return nil // Already gone or invalid
	}

	return s.authRepo.RevokeRefreshToken(ctx, rt.ID)
}

func (s *AuthService) generateRefreshToken(ctx context.Context, userID uuid.UUID) (entity.RefreshToken, string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := cryptoRandRead(tokenBytes); err != nil {
		return entity.RefreshToken{}, "", err
	}
	plainRT := fmt.Sprintf("%x", tokenBytes)
	hash := s.hashToken(plainRT)

	rt := entity.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	if err := s.authRepo.SaveRefreshToken(ctx, rt); err != nil {
		return entity.RefreshToken{}, "", err
	}

	return rt, plainRT, nil
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

	// 5. Revoke all refresh tokens for this user
	_ = s.authRepo.RevokeAllRefreshTokens(ctx, userID)

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

// RequestPasswordReset initiates the reset flow for a given email.
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	// 1. Find the user by email
	users, err := s.repo.FindAll(ctx, email) // FindAll supports search
	if err != nil {
		s.logger.Error("Failed to search for user during reset request", "email", email, "error", err)
		return nil // Generic success to prevent enumeration
	}

	var targetUser *entity.User
	for _, u := range users {
		if u.Email == email {
			targetUser = &u
			break
		}
	}

	if targetUser == nil {
		s.logger.Warn("Password reset requested for non-existent email", "email", email)
		return nil // Generic success to prevent enumeration
	}

	// 2. Generate a secure random token
	tokenBytes := make([]byte, 32)
	if _, err := cryptoRandRead(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate secure token: %w", err)
	}
	plainToken := fmt.Sprintf("%x", tokenBytes)
	tokenHash := s.hashToken(plainToken)

	// 3. Store the token hash
	// First, invalidate old tokens
	_ = s.resetRepo.DeleteByUserID(ctx, targetUser.ID)

	resetToken := entity.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    targetUser.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := s.resetRepo.Create(ctx, resetToken); err != nil {
		s.logger.Error("Failed to store reset token", "user_id", targetUser.ID, "error", err)
		return fmt.Errorf("internal server error")
	}

	// 4. Send the email
	admin, _ := s.repo.FindByUsername(ctx, "admin")
	adminID := admin.ID

	domain, _ := s.settingsRepo.Get(ctx, "domain_name", adminID)
	if domain == "" {
		domain = "localhost"
	}
	resetURL := fmt.Sprintf("http://%s/reset-password?token=%s", domain, plainToken)
	if strings.Contains(domain, "localhost") {
		resetURL = fmt.Sprintf("http://localhost:3000/reset-password?token=%s", plainToken)
	}

	if err := s.notificationSvc.SendPasswordResetEmail(ctx, *targetUser, resetURL); err != nil {
		s.logger.Error("Failed to send reset email", "email", email, "error", err)
		// We still return nil to the user as the process started
	}

	s.logger.Info("Password reset link sent", "user_id", targetUser.ID)
	return nil
}

func (s *AuthService) ValidateResetToken(ctx context.Context, token string) (bool, error) {
	tokenHash := s.hashToken(token)
	t, err := s.resetRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, entity.ErrResetTokenInvalid) {
			return false, nil
		}
		return false, err
	}

	if t.IsExpired() {
		_ = s.resetRepo.Delete(ctx, t.ID)
		return false, nil
	}

	return true, nil
}

func (s *AuthService) ConfirmPasswordReset(ctx context.Context, token string, newPassword string) error {
	tokenHash := s.hashToken(token)
	t, err := s.resetRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return entity.ErrResetTokenInvalid
	}

	if t.IsExpired() {
		_ = s.resetRepo.Delete(ctx, t.ID)
		return entity.ErrResetTokenInvalid
	}

	// Hash the new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	if err := s.repo.UpdatePassword(ctx, t.UserID, string(newHash)); err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	// Revoke all refresh tokens for this user
	_ = s.authRepo.RevokeAllRefreshTokens(ctx, t.UserID)

	// Delete the used token
	_ = s.resetRepo.Delete(ctx, t.ID)

	s.logger.Info("Password reset confirmed successfully", "user_id", t.UserID)
	return nil
}

func (s *AuthService) hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// cryptoRandRead is a mockable wrapper for crypto/rand.Read
var cryptoRandRead = cryptoRandReadFunc

func cryptoRandReadFunc(b []byte) (n int, err error) {
	return rand.Read(b)
}
