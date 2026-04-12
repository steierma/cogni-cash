package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
)

type BridgeAccessTokenRepository struct {
	mu     sync.RWMutex
	tokens map[uuid.UUID]entity.BridgeAccessToken
}

func NewBridgeAccessTokenRepository() *BridgeAccessTokenRepository {
	return &BridgeAccessTokenRepository{
		tokens: make(map[uuid.UUID]entity.BridgeAccessToken),
	}
}

func (r *BridgeAccessTokenRepository) Save(ctx context.Context, token entity.BridgeAccessToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[token.ID] = token
	return nil
}

func (r *BridgeAccessTokenRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]entity.BridgeAccessToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var tokens []entity.BridgeAccessToken
	for _, t := range r.tokens {
		if t.UserID == userID {
			tokens = append(tokens, t)
		}
	}
	return tokens, nil
}

func (r *BridgeAccessTokenRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.BridgeAccessToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tokens[id]
	if !ok || t.UserID != userID {
		return entity.BridgeAccessToken{}, errors.New("bridge token not found")
	}
	return t, nil
}

func (r *BridgeAccessTokenRepository) FindByHash(ctx context.Context, hash string) (entity.BridgeAccessToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, t := range r.tokens {
		if t.TokenHash == hash {
			return t, nil
		}
	}
	return entity.BridgeAccessToken{}, errors.New("bridge token not found")
}

func (r *BridgeAccessTokenRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tokens[id]
	if !ok || t.UserID != userID {
		return errors.New("bridge token not found")
	}
	now := time.Now()
	t.LastUsedAt = &now
	r.tokens[id] = t
	return nil
}

func (r *BridgeAccessTokenRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tokens[id]
	if !ok || t.UserID != userID {
		return errors.New("bridge token not found")
	}
	delete(r.tokens, id)
	return nil
}
