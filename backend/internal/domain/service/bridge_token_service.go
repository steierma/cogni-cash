package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

type BridgeAccessTokenService struct {
	repo port.BridgeAccessTokenRepository
}

func NewBridgeAccessTokenService(repo port.BridgeAccessTokenRepository) *BridgeAccessTokenService {
	return &BridgeAccessTokenService{repo: repo}
}

func (s *BridgeAccessTokenService) CreateToken(ctx context.Context, userID uuid.UUID, name string) (entity.CreateBridgeTokenResponse, error) {
	token, err := generateRandomToken(32)
	if err != nil {
		return entity.CreateBridgeTokenResponse{}, fmt.Errorf("failed to generate token: %w", err)
	}

	hash := hashToken(token)
	tokenID := uuid.New()
	createdAt := time.Now()

	bat := entity.BridgeAccessToken{
		ID:        tokenID,
		UserID:    userID,
		Name:      name,
		TokenHash: hash,
		CreatedAt: createdAt,
	}

	if err := s.repo.Save(ctx, bat); err != nil {
		return entity.CreateBridgeTokenResponse{}, fmt.Errorf("failed to save token: %w", err)
	}

	return entity.CreateBridgeTokenResponse{
		Token: token,
		Info:  bat,
	}, nil
}

func (s *BridgeAccessTokenService) ListTokens(ctx context.Context, userID uuid.UUID) ([]entity.BridgeAccessToken, error) {
	return s.repo.FindAll(ctx, userID)
}

func (s *BridgeAccessTokenService) RevokeToken(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *BridgeAccessTokenService) ValidateToken(ctx context.Context, token string) (uuid.UUID, error) {
	hash := hashToken(token)
	bat, err := s.repo.FindByHash(ctx, hash)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	if err := s.repo.UpdateLastUsed(ctx, bat.ID, bat.UserID); err != nil {
		// Log error but don't fail validation
		fmt.Printf("Warning: failed to update last_used for token %s: %v\n", bat.ID, err)
	}

	return bat.UserID, nil
}

func generateRandomToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}
