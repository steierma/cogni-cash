package memory

import (
	"context"
	"errors"
	"strings"
	"sync"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

const maxUsers = 50

type UserRepository struct {
	mu    sync.RWMutex
	users map[uuid.UUID]entity.User
	order []uuid.UUID
}

func NewUserRepository() *UserRepository {
	r := &UserRepository{
		users: make(map[uuid.UUID]entity.User),
		order: make([]uuid.UUID, 0, maxUsers),
	}
	r.seedData()
	return r
}

func (r *UserRepository) seedData() {
	// The deterministic UUID that connects all our seeded memory data
	testUserID := uuid.MustParse("12345678-1234-1234-1234-123456789012")

	// 1. Seed the primary Test User
	testUser := entity.User{
		ID:       testUserID,
		Username: "test",
		Email:    "test@example.com",
		FullName: "John Doe", // Matches the AccountHolder in BankStatement
		// Pre-computed bcrypt hash for the password "password"
		// This allows you to easily log in during development
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
	}

	r.users[testUserID] = testUser
	r.order = append(r.order, testUserID)

	// 2. Seed an Admin User (required for GetAdminID to function properly)
	adminID := uuid.New()
	adminUser := entity.User{
		ID:           adminID,
		Username:     "admin",
		Email:        "admin@example.com",
		FullName:     "System Admin",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // password
	}

	r.users[adminID] = adminUser
	r.order = append(r.order, adminID)
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return entity.User{}, entity.ErrUserNotFound
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return entity.User{}, entity.ErrUserNotFound
	}
	return user, nil
}

func (r *UserRepository) GetAdminID(ctx context.Context) (uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Username == "admin" {
			return u.ID, nil
		}
	}
	return uuid.Nil, errors.New("admin user not found")
}

func (r *UserRepository) FindAll(ctx context.Context, search string) ([]entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var users []entity.User
	searchLower := strings.ToLower(search)
	for _, u := range r.users {
		if search == "" ||
			strings.Contains(strings.ToLower(u.Username), searchLower) ||
			strings.Contains(strings.ToLower(u.Email), searchLower) ||
			strings.Contains(strings.ToLower(u.FullName), searchLower) {
			users = append(users, u)
		}
	}
	return users, nil
}

func (r *UserRepository) Create(ctx context.Context, user entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	if _, exists := r.users[user.ID]; !exists {
		if len(r.order) >= maxUsers {
			// Find a non-admin/non-test user to evict
			evictedIdx := -1
			for i, id := range r.order {
				u := r.users[id]
				if u.Username != "admin" && u.Username != "test" {
					evictedIdx = i
					delete(r.users, id)
					break
				}
			}
			if evictedIdx != -1 {
				r.order = append(r.order[:evictedIdx], r.order[evictedIdx+1:]...)
			}
		}
		r.order = append(r.order, user.ID)
	}

	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) Update(ctx context.Context, user entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[user.ID]; !ok {
		return entity.ErrUserNotFound
	}
	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) Upsert(ctx context.Context, user entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if user.ID == uuid.Nil {
		// Try to find by username for upsert
		for _, u := range r.users {
			if u.Username == user.Username {
				user.ID = u.ID
				break
			}
		}
		if user.ID == uuid.Nil {
			user.ID = uuid.New()
		}
	}

	if _, exists := r.users[user.ID]; !exists {
		if len(r.order) >= maxUsers {
			// Evict oldest non-protected
			evictedIdx := -1
			for i, id := range r.order {
				u := r.users[id]
				if u.Username != "admin" && u.Username != "test" {
					evictedIdx = i
					delete(r.users, id)
					break
				}
			}
			if evictedIdx != -1 {
				r.order = append(r.order[:evictedIdx], r.order[evictedIdx+1:]...)
			}
		}
		r.order = append(r.order, user.ID)
	}

	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, newHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.users[userID]
	if !ok {
		return entity.ErrUserNotFound
	}
	user.PasswordHash = newHash
	r.users[userID] = user
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[id]; !ok {
		return entity.ErrUserNotFound
	}
	delete(r.users, id)
	return nil
}

var _ port.UserRepository = (*UserRepository)(nil)
