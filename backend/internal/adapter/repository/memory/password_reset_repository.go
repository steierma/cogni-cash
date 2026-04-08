package memory

import (
	"cogni-cash/internal/domain/entity"
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type PasswordResetRepository struct {
	mu     sync.RWMutex
	tokens map[uuid.UUID]entity.PasswordResetToken
}

func NewPasswordResetRepository() *PasswordResetRepository {
	return &PasswordResetRepository{
		tokens: make(map[uuid.UUID]entity.PasswordResetToken),
	}
}

func (r *PasswordResetRepository) Create(_ context.Context, token entity.PasswordResetToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[token.ID] = token
	return nil
}

func (r *PasswordResetRepository) FindByHash(_ context.Context, tokenHash string) (entity.PasswordResetToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, t := range r.tokens {
		if t.TokenHash == tokenHash {
			return t, nil
		}
	}
	return entity.PasswordResetToken{}, entity.ErrResetTokenInvalid
}

func (r *PasswordResetRepository) DeleteByUserID(_ context.Context, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, t := range r.tokens {
		if t.UserID == userID {
			delete(r.tokens, id)
		}
	}
	return nil
}

func (r *PasswordResetRepository) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tokens, id)
	return nil
}

func (r *PasswordResetRepository) CleanupExpired(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for id, t := range r.tokens {
		if now.After(t.ExpiresAt) {
			delete(r.tokens, id)
		}
	}
	return nil
}
