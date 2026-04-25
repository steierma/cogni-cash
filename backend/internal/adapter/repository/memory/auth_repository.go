package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"

	"cogni-cash/internal/domain/entity"
)

// AuthRepository is an in-memory implementation of port.AuthRepository.
type AuthRepository struct {
	mu            sync.RWMutex
	refreshTokens map[string]entity.RefreshToken
}

// NewAuthRepository creates a new AuthRepository.
func NewAuthRepository() *AuthRepository {
	return &AuthRepository{
		refreshTokens: make(map[string]entity.RefreshToken),
	}
}

func (r *AuthRepository) SaveRefreshToken(_ context.Context, token entity.RefreshToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refreshTokens[token.TokenHash] = token
	return nil
}

func (r *AuthRepository) FindRefreshToken(_ context.Context, tokenHash string) (entity.RefreshToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	token, ok := r.refreshTokens[tokenHash]
	if !ok {
		return entity.RefreshToken{}, errors.New("refresh token not found")
	}
	return token, nil
}

func (r *AuthRepository) RevokeRefreshToken(_ context.Context, id uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for hash, token := range r.refreshTokens {
		if token.ID == id && token.UserID == userID {
			token.Revoked = true
			r.refreshTokens[hash] = token
			break
		}
	}
	return nil
}

func (r *AuthRepository) RevokeAllRefreshTokens(_ context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for hash, token := range r.refreshTokens {
		if token.UserID == userID {
			token.Revoked = true
			r.refreshTokens[hash] = token
		}
	}
	return nil
}

func (r *AuthRepository) CleanupExpiredRefreshTokens(_ context.Context) error {
	// Not implemented for in-memory as it's not strictly necessary for tests/sandbox
	return nil
}
