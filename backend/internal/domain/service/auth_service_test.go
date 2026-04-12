package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// --- Mock User Repository for Auth Tests ---

type mockUserRepoForAuth struct {
	users    map[string]entity.User // keyed by username
	usersID  map[uuid.UUID]entity.User
	createFn func(entity.User) error
}

func newMockUserRepoForAuth() *mockUserRepoForAuth {
	return &mockUserRepoForAuth{
		users:   make(map[string]entity.User),
		usersID: make(map[uuid.UUID]entity.User),
	}
}

func (m *mockUserRepoForAuth) addUser(u entity.User) {
	m.users[u.Username] = u
	m.usersID[u.ID] = u
}

func (m *mockUserRepoForAuth) FindByUsername(_ context.Context, username string) (entity.User, error) {
	if u, ok := m.users[username]; ok {
		return u, nil
	}
	return entity.User{}, errors.New("user not found")
}

func (m *mockUserRepoForAuth) FindByID(_ context.Context, id uuid.UUID) (entity.User, error) {
	if u, ok := m.usersID[id]; ok {
		return u, nil
	}
	return entity.User{}, errors.New("user not found")
}

func (m *mockUserRepoForAuth) GetAdminID(_ context.Context) (uuid.UUID, error) {
	for _, u := range m.users {
		if u.Username == "admin" {
			return u.ID, nil
		}
	}
	return uuid.Nil, errors.New("admin not found")
}

func (m *mockUserRepoForAuth) FindAll(_ context.Context, _ string) ([]entity.User, error) {
	var result []entity.User
	for _, u := range m.users {
		result = append(result, u)
	}
	return result, nil
}

func (m *mockUserRepoForAuth) Create(_ context.Context, user entity.User) error {
	if m.createFn != nil {
		return m.createFn(user)
	}
	m.users[user.Username] = user
	m.usersID[user.ID] = user
	return nil
}

func (m *mockUserRepoForAuth) Update(_ context.Context, user entity.User) error {
	m.users[user.Username] = user
	m.usersID[user.ID] = user
	return nil
}

func (m *mockUserRepoForAuth) Upsert(_ context.Context, user entity.User) error {
	m.users[user.Username] = user
	m.usersID[user.ID] = user
	return nil
}

func (m *mockUserRepoForAuth) UpdatePassword(_ context.Context, userID uuid.UUID, newHash string) error {
	if u, ok := m.usersID[userID]; ok {
		u.PasswordHash = newHash
		m.usersID[userID] = u
		m.users[u.Username] = u
		return nil
	}
	return errors.New("user not found")
}

func (m *mockUserRepoForAuth) Delete(_ context.Context, id uuid.UUID) error {
	if u, ok := m.usersID[id]; ok {
		delete(m.usersID, id)
		delete(m.users, u.Username)
		return nil
	}
	return errors.New("user not found")
}

// --- Mock Password Reset Repository ---

type mockResetRepo struct {
	tokens  map[uuid.UUID]entity.PasswordResetToken
	hashMap map[string]entity.PasswordResetToken
}

func newMockResetRepo() *mockResetRepo {
	return &mockResetRepo{
		tokens:  make(map[uuid.UUID]entity.PasswordResetToken),
		hashMap: make(map[string]entity.PasswordResetToken),
	}
}

func (m *mockResetRepo) Create(_ context.Context, t entity.PasswordResetToken) error {
	m.tokens[t.ID] = t
	m.hashMap[t.TokenHash] = t
	return nil
}

func (m *mockResetRepo) FindByHash(_ context.Context, hash string) (entity.PasswordResetToken, error) {
	if t, ok := m.hashMap[hash]; ok {
		return t, nil
	}
	return entity.PasswordResetToken{}, entity.ErrResetTokenInvalid
}

func (m *mockResetRepo) DeleteByUserID(_ context.Context, userID uuid.UUID) error {
	for id, t := range m.tokens {
		if t.UserID == userID {
			delete(m.tokens, id)
			delete(m.hashMap, t.TokenHash)
		}
	}
	return nil
}

func (m *mockResetRepo) Delete(_ context.Context, id uuid.UUID) error {
	if t, ok := m.tokens[id]; ok {
		delete(m.tokens, id)
		delete(m.hashMap, t.TokenHash)
	}
	return nil
}

func (m *mockResetRepo) CleanupExpired(_ context.Context) error { return nil }

// --- Mock Auth Repository ---

type mockAuthRepo struct {
	tokens map[string]entity.RefreshToken
}

func newMockAuthRepo() *mockAuthRepo {
	return &mockAuthRepo{tokens: make(map[string]entity.RefreshToken)}
}

func (m *mockAuthRepo) SaveRefreshToken(_ context.Context, t entity.RefreshToken) error {
	m.tokens[t.TokenHash] = t
	return nil
}

func (m *mockAuthRepo) FindRefreshToken(_ context.Context, hash string) (entity.RefreshToken, error) {
	if t, ok := m.tokens[hash]; ok {
		return t, nil
	}
	return entity.RefreshToken{}, errors.New("not found")
}

func (m *mockAuthRepo) RevokeRefreshToken(_ context.Context, id uuid.UUID) error {
	for hash, t := range m.tokens {
		if t.ID == id {
			t.Revoked = true
			m.tokens[hash] = t
		}
	}
	return nil
}

func (m *mockAuthRepo) RevokeAllRefreshTokens(_ context.Context, userID uuid.UUID) error {
	for hash, t := range m.tokens {
		if t.UserID == userID {
			t.Revoked = true
			m.tokens[hash] = t
		}
	}
	return nil
}

func (m *mockAuthRepo) CleanupExpiredRefreshTokens(_ context.Context) error { return nil }

// --- Mock Notification Service ---

type mockNotificationSvc struct {
	sentResetEmail bool
	lastResetURL   string
}

func (m *mockNotificationSvc) SendWelcomeEmail(_ context.Context, _ entity.User) error { return nil }
func (m *mockNotificationSvc) SendPasswordResetEmail(_ context.Context, _ entity.User, url string) error {
	m.sentResetEmail = true
	m.lastResetURL = url
	return nil
}
func (m *mockNotificationSvc) SendTestEmail(_ context.Context, _ string, _ uuid.UUID) error { return nil }

// --- Mock Settings Repository ---

type mockSettingsRepoForAuth struct {
	settings map[string]string
}

func (m *mockSettingsRepoForAuth) Get(_ context.Context, key string, _ uuid.UUID) (string, error) {
	return m.settings[key], nil
}
func (m *mockSettingsRepoForAuth) GetAll(_ context.Context, _ uuid.UUID) (map[string]string, error) {
	return m.settings, nil
}
func (m *mockSettingsRepoForAuth) Set(_ context.Context, key, value string, _ uuid.UUID) error {
	m.settings[key] = value
	return nil
}

func nopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestAuthService_PasswordReset(t *testing.T) {
	userRepo := newMockUserRepoForAuth()
	resetRepo := newMockResetRepo()
	notifSvc := &mockNotificationSvc{}
	settingsRepo := &mockSettingsRepoForAuth{settings: make(map[string]string)}
	logger := nopLogger()
	secret := "test-secret"

	user := entity.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}
	userRepo.addUser(user)

	svc := service.NewAuthService(userRepo, resetRepo, newMockAuthRepo(), notifSvc, settingsRepo, secret, logger)

	t.Run("RequestPasswordReset success", func(t *testing.T) {
		err := svc.RequestPasswordReset(context.Background(), "test@example.com")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !notifSvc.sentResetEmail {
			t.Error("expected reset email to be sent")
		}
	})

	t.Run("ConfirmPasswordReset success", func(t *testing.T) {
		// 1. Request to get a token
		_ = svc.RequestPasswordReset(context.Background(), "test@example.com")
		
		// Extract token from mock URL (e.g. http://localhost:3000/reset-password?token=...)
		url := notifSvc.lastResetURL
		token := url[len(url)-64:] // 32 bytes hex = 64 chars

		// 2. Confirm reset
		newPwd := "new-secure-password"
		err := svc.ConfirmPasswordReset(context.Background(), token, newPwd)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// 3. Verify password was updated
		updatedUser, _ := userRepo.FindByID(context.Background(), user.ID)
		err = bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte(newPwd))
		if err != nil {
			t.Error("expected password hash to match new password")
		}
	})

	t.Run("ValidateResetToken invalid", func(t *testing.T) {
		valid, err := svc.ValidateResetToken(context.Background(), "invalid-token")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if valid {
			t.Error("expected token to be invalid")
		}
	})
}

func hashPassword(t *testing.T, plain string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(hash)
}

// --- Login Tests ---

func TestAuthService_Login_Success(t *testing.T) {
	repo := newMockUserRepoForAuth()
	user := entity.User{
		ID:           uuid.New(),
		Username:     "admin",
		PasswordHash: hashPassword(t, "secret123"),
	}
	repo.addUser(user)

	svc := service.NewAuthService(repo, nil, newMockAuthRepo(), nil, nil, "test-jwt-secret", nopLogger())

	authResp, err := svc.Login(context.Background(), "admin", "secret123")
	if err != nil {
		t.Fatalf("expected successful login, got error: %v", err)
	}
	if authResp.Token == "" {
		t.Error("expected non-empty JWT token")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	repo := newMockUserRepoForAuth()
	user := entity.User{
		ID:           uuid.New(),
		Username:     "admin",
		PasswordHash: hashPassword(t, "correct"),
	}
	repo.addUser(user)

	svc := service.NewAuthService(repo, nil, newMockAuthRepo(), nil, nil, "test-jwt-secret", nopLogger())

	_, err := svc.Login(context.Background(), "admin", "wrong")
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := newMockUserRepoForAuth()
	svc := service.NewAuthService(repo, nil, newMockAuthRepo(), nil, nil, "test-jwt-secret", nopLogger())

	_, err := svc.Login(context.Background(), "nonexistent", "password")
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_NilRepo(t *testing.T) {
	svc := service.NewAuthService(nil, nil, newMockAuthRepo(), nil, nil, "test-jwt-secret", nopLogger())

	_, err := svc.Login(context.Background(), "admin", "password")
	if err == nil {
		t.Error("expected error when repo is nil")
	}
}

// --- ValidateToken Tests ---

func TestAuthService_ValidateToken_Success(t *testing.T) {
	repo := newMockUserRepoForAuth()
	user := entity.User{
		ID:           uuid.New(),
		Username:     "admin",
		PasswordHash: hashPassword(t, "secret123"),
	}
	repo.addUser(user)

	svc := service.NewAuthService(repo, nil, newMockAuthRepo(), nil, nil, "test-jwt-secret", nopLogger())

	authResp, err := svc.Login(context.Background(), "admin", "secret123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	sub, err := svc.ValidateToken(authResp.Token)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}
	if sub != user.ID.String() {
		t.Errorf("expected subject %q, got %q", user.ID.String(), sub)
	}
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	svc := service.NewAuthService(nil, nil, newMockAuthRepo(), nil, nil, "test-jwt-secret", nopLogger())

	_, err := svc.ValidateToken("invalid.token.string")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestAuthService_ValidateToken_WrongSecret(t *testing.T) {
	repo := newMockUserRepoForAuth()
	user := entity.User{
		ID:           uuid.New(),
		Username:     "admin",
		PasswordHash: hashPassword(t, "secret123"),
	}
	repo.addUser(user)

	svc1 := service.NewAuthService(repo, nil, newMockAuthRepo(), nil, nil, "secret-1", nopLogger())
	authResp1, _ := svc1.Login(context.Background(), "admin", "secret123")

	svc2 := service.NewAuthService(repo, nil, newMockAuthRepo(), nil, nil, "secret-2", nopLogger())
	_, err := svc2.ValidateToken(authResp1.Token)
	if err == nil {
		t.Error("expected error when validating token signed with different secret")
	}
}

// --- ChangePassword Tests ---

func TestAuthService_ChangePassword_Success(t *testing.T) {
	repo := newMockUserRepoForAuth()
	userID := uuid.New()
	user := entity.User{
		ID:           userID,
		Username:     "admin",
		PasswordHash: hashPassword(t, "oldpass"),
	}
	repo.addUser(user)

	svc := service.NewAuthService(repo, nil, newMockAuthRepo(), nil, nil, "test-jwt-secret", nopLogger())

	err := svc.ChangePassword(context.Background(), userID.String(), "oldpass", "newpass")
	if err != nil {
		t.Fatalf("expected successful password change, got: %v", err)
	}

	// Verify new password works
	_, err = svc.Login(context.Background(), "admin", "newpass")
	if err != nil {
		t.Error("expected login with new password to succeed")
	}
}

func TestAuthService_ChangePassword_WrongOldPassword(t *testing.T) {
	repo := newMockUserRepoForAuth()
	userID := uuid.New()
	user := entity.User{
		ID:           userID,
		Username:     "admin",
		PasswordHash: hashPassword(t, "correct"),
	}
	repo.addUser(user)

	svc := service.NewAuthService(repo, nil, newMockAuthRepo(), nil, nil, "test-jwt-secret", nopLogger())

	err := svc.ChangePassword(context.Background(), userID.String(), "wrong", "newpass")
	if err == nil {
		t.Error("expected error for wrong old password")
	}
}

func TestAuthService_ChangePassword_InvalidUserID(t *testing.T) {
	svc := service.NewAuthService(newMockUserRepoForAuth(), nil, newMockAuthRepo(), nil, nil, "test-jwt-secret", nopLogger())

	err := svc.ChangePassword(context.Background(), "not-a-uuid", "old", "new")
	if err == nil {
		t.Error("expected error for invalid user ID")
	}
}

func TestAuthService_ChangePassword_UserNotFound(t *testing.T) {
	svc := service.NewAuthService(newMockUserRepoForAuth(), nil, newMockAuthRepo(), nil, nil, "test-jwt-secret", nopLogger())

	err := svc.ChangePassword(context.Background(), uuid.New().String(), "old", "new")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

// --- EnsureAdminUser Tests ---

func TestAuthService_EnsureAdminUser_CreatesNewUser(t *testing.T) {
	repo := newMockUserRepoForAuth()
	svc := service.NewAuthService(repo, nil, newMockAuthRepo(), nil, nil, "test-jwt-secret", nopLogger())

	err := svc.EnsureAdminUser(context.Background(), "admin", "password123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify admin was created and can login
	_, err = svc.Login(context.Background(), "admin", "password123")
	if err != nil {
		t.Error("expected admin user to be created and loginable")
	}
}

func TestAuthService_EnsureAdminUser_SkipsWhenAlreadyCorrect(t *testing.T) {
	repo := newMockUserRepoForAuth()
	svc := service.NewAuthService(repo, nil, newMockAuthRepo(), nil, nil, "test-jwt-secret", nopLogger())

	// Create admin first
	err := svc.EnsureAdminUser(context.Background(), "admin", "password123")
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Second call should be a no-op
	err = svc.EnsureAdminUser(context.Background(), "admin", "password123")
	if err != nil {
		t.Fatalf("expected no error on idempotent call, got: %v", err)
	}
}

func TestAuthService_EnsureAdminUser_RotatesPassword(t *testing.T) {
	repo := newMockUserRepoForAuth()
	svc := service.NewAuthService(repo, nil, newMockAuthRepo(), nil, nil, "test-jwt-secret", nopLogger())

	// Create admin with initial password
	err := svc.EnsureAdminUser(context.Background(), "admin", "initial")
	if err != nil {
		t.Fatalf("initial creation failed: %v", err)
	}

	// Rotate password
	err = svc.EnsureAdminUser(context.Background(), "admin", "rotated")
	if err != nil {
		t.Fatalf("password rotation failed: %v", err)
	}

	// Verify new password works
	_, err = svc.Login(context.Background(), "admin", "rotated")
	if err != nil {
		t.Error("expected login with rotated password to succeed")
	}
}

func TestAuthService_EnsureAdminUser_EmptyCredentials(t *testing.T) {
	svc := service.NewAuthService(newMockUserRepoForAuth(), nil, newMockAuthRepo(), nil, nil, "test-jwt-secret", nopLogger())

	err := svc.EnsureAdminUser(context.Background(), "", "")
	if err != nil {
		t.Errorf("expected nil (skip), got: %v", err)
	}

	err = svc.EnsureAdminUser(context.Background(), "admin", "")
	if err != nil {
		t.Errorf("expected nil (skip), got: %v", err)
	}
}

