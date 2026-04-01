package service_test

import (
	"context"
	"errors"
	"testing"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
)

// --- Mock User Repository for UserService Tests ---

type mockUserRepoForUserSvc struct {
	users map[uuid.UUID]entity.User
}

func newMockUserRepoForUserSvc() *mockUserRepoForUserSvc {
	return &mockUserRepoForUserSvc{users: make(map[uuid.UUID]entity.User)}
}

func (m *mockUserRepoForUserSvc) addUser(u entity.User) {
	m.users[u.ID] = u
}

func (m *mockUserRepoForUserSvc) FindByUsername(_ context.Context, username string) (entity.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return entity.User{}, errors.New("not found")
}

func (m *mockUserRepoForUserSvc) FindByID(_ context.Context, id uuid.UUID) (entity.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return entity.User{}, errors.New("not found")
}

func (m *mockUserRepoForUserSvc) GetAdminID(_ context.Context) (uuid.UUID, error) {
	for _, u := range m.users {
		if u.Username == "admin" {
			return u.ID, nil
		}
	}
	return uuid.Nil, errors.New("admin not found")
}

func (m *mockUserRepoForUserSvc) FindAll(_ context.Context, search string) ([]entity.User, error) {
	var result []entity.User
	for _, u := range m.users {
		result = append(result, u)
	}
	return result, nil
}

func (m *mockUserRepoForUserSvc) Create(_ context.Context, user entity.User) error {
	for _, u := range m.users {
		if u.Username == user.Username {
			return errors.New("username already exists")
		}
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepoForUserSvc) Update(_ context.Context, user entity.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return errors.New("not found")
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepoForUserSvc) Upsert(_ context.Context, user entity.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepoForUserSvc) UpdatePassword(_ context.Context, userID uuid.UUID, newHash string) error {
	if u, ok := m.users[userID]; ok {
		u.PasswordHash = newHash
		m.users[userID] = u
		return nil
	}
	return errors.New("not found")
}

func (m *mockUserRepoForUserSvc) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := m.users[id]; !ok {
		return errors.New("not found")
	}
	delete(m.users, id)
	return nil
}

// --- Tests ---

func TestUserService_ListUsers(t *testing.T) {
	repo := newMockUserRepoForUserSvc()
	repo.addUser(entity.User{ID: uuid.New(), Username: "alice"})
	repo.addUser(entity.User{ID: uuid.New(), Username: "bob"})

	svc := service.NewUserService(repo, nil)
	users, err := svc.ListUsers(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestUserService_GetUser_Success(t *testing.T) {
	repo := newMockUserRepoForUserSvc()
	userID := uuid.New()
	repo.addUser(entity.User{ID: userID, Username: "alice"})

	svc := service.NewUserService(repo, nil)
	user, err := svc.GetUser(context.Background(), userID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "alice" {
		t.Errorf("expected username 'alice', got %q", user.Username)
	}
}

func TestUserService_GetUser_InvalidID(t *testing.T) {
	svc := service.NewUserService(newMockUserRepoForUserSvc(), nil)
	_, err := svc.GetUser(context.Background(), "not-a-uuid")
	if err == nil {
		t.Error("expected error for invalid UUID")
	}
}

func TestUserService_GetUser_NotFound(t *testing.T) {
	svc := service.NewUserService(newMockUserRepoForUserSvc(), nil)
	_, err := svc.GetUser(context.Background(), uuid.New().String())
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestUserService_CreateUser_Success(t *testing.T) {
	repo := newMockUserRepoForUserSvc()
	svc := service.NewUserService(repo, nil)

	user, err := svc.CreateUser(context.Background(), entity.User{
		Username: "newuser",
		Email:    "new@example.com",
		FullName: "New User",
	}, "password123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID == uuid.Nil {
		t.Error("expected a generated user ID")
	}
	if user.PasswordHash == "" {
		t.Error("expected password hash to be set")
	}
	if user.Role != "manager" {
		t.Errorf("expected default role 'manager', got %q", user.Role)
	}
}

func TestUserService_CreateUser_EmptyPassword(t *testing.T) {
	svc := service.NewUserService(newMockUserRepoForUserSvc(), nil)

	_, err := svc.CreateUser(context.Background(), entity.User{Username: "test"}, "")
	if err == nil {
		t.Error("expected error for empty password")
	}
}

func TestUserService_CreateUser_WithRole(t *testing.T) {
	svc := service.NewUserService(newMockUserRepoForUserSvc(), nil)

	user, err := svc.CreateUser(context.Background(), entity.User{
		Username: "admin",
		Role:     "admin",
	}, "password123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Role != "admin" {
		t.Errorf("expected role 'admin', got %q", user.Role)
	}
}

func TestUserService_UpdateUser_Success(t *testing.T) {
	repo := newMockUserRepoForUserSvc()
	userID := uuid.New()
	repo.addUser(entity.User{
		ID:       userID,
		Username: "alice",
		Email:    "alice@old.com",
		Role:     "manager",
	})

	svc := service.NewUserService(repo, nil)
	updated, err := svc.UpdateUser(context.Background(), userID.String(), entity.User{
		Username: "alice_updated",
		Email:    "alice@new.com",
		FullName: "Alice Wonderland",
		Address:  "123 Main St",
		Role:     "admin",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Username != "alice_updated" {
		t.Errorf("expected updated username, got %q", updated.Username)
	}
	if updated.Email != "alice@new.com" {
		t.Errorf("expected updated email, got %q", updated.Email)
	}
	if updated.Role != "admin" {
		t.Errorf("expected updated role 'admin', got %q", updated.Role)
	}
}

func TestUserService_UpdateUser_InvalidID(t *testing.T) {
	svc := service.NewUserService(newMockUserRepoForUserSvc(), nil)
	_, err := svc.UpdateUser(context.Background(), "bad-id", entity.User{})
	if err == nil {
		t.Error("expected error for invalid UUID")
	}
}

func TestUserService_UpdateUser_NotFound(t *testing.T) {
	svc := service.NewUserService(newMockUserRepoForUserSvc(), nil)
	_, err := svc.UpdateUser(context.Background(), uuid.New().String(), entity.User{Username: "x"})
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestUserService_DeleteUser_Success(t *testing.T) {
	repo := newMockUserRepoForUserSvc()
	userID := uuid.New()
	repo.addUser(entity.User{ID: userID, Username: "alice"})

	svc := service.NewUserService(repo, nil)
	err := svc.DeleteUser(context.Background(), userID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify user was deleted
	_, err = svc.GetUser(context.Background(), userID.String())
	if err == nil {
		t.Error("expected user to be deleted")
	}
}

func TestUserService_DeleteUser_InvalidID(t *testing.T) {
	svc := service.NewUserService(newMockUserRepoForUserSvc(), nil)
	err := svc.DeleteUser(context.Background(), "not-a-uuid")
	if err == nil {
		t.Error("expected error for invalid UUID")
	}
}

func TestUserService_DeleteUser_NotFound(t *testing.T) {
	svc := service.NewUserService(newMockUserRepoForUserSvc(), nil)
	err := svc.DeleteUser(context.Background(), uuid.New().String())
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

