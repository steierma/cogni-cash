package service_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/service"

	"github.com/google/uuid"
)

type mockEmailProvider struct {
	sentTo      string
	sentSubject string
	sentBody    string
	err         error
}

func (m *mockEmailProvider) Send(ctx context.Context, userID uuid.UUID, to, subject, body string) error {
	if m.err != nil {
		return m.err
	}
	m.sentTo = to
	m.sentSubject = subject
	m.sentBody = body
	return nil
}

type mockUserRepoForNotification struct {
	admin entity.User
}

func (m *mockUserRepoForNotification) FindByUsername(ctx context.Context, username string) (entity.User, error) {
	if username == "admin" {
		return m.admin, nil
	}
	return entity.User{}, errors.New("not found")
}
func (m *mockUserRepoForNotification) FindByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	return entity.User{}, nil
}
func (m *mockUserRepoForNotification) GetAdminID(ctx context.Context) (uuid.UUID, error) {
	return m.admin.ID, nil
}
func (m *mockUserRepoForNotification) FindAll(ctx context.Context, search string) ([]entity.User, error) {
	return nil, nil
}
func (m *mockUserRepoForNotification) Create(ctx context.Context, user entity.User) error { return nil }
func (m *mockUserRepoForNotification) Update(ctx context.Context, user entity.User) error { return nil }
func (m *mockUserRepoForNotification) Upsert(ctx context.Context, user entity.User) error { return nil }
func (m *mockUserRepoForNotification) UpdatePassword(ctx context.Context, userID uuid.UUID, newHash string) error {
	return nil
}
func (m *mockUserRepoForNotification) Delete(ctx context.Context, id uuid.UUID) error { return nil }

func TestNotificationService_SendWelcomeEmail(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	userRepo := &mockUserRepoForNotification{admin: entity.User{ID: uuid.New(), Username: "admin"}}

	t.Run("success", func(t *testing.T) {
		mockEmail := &mockEmailProvider{}
		svc := service.NewNotificationService(mockEmail, userRepo, logger)

		user := entity.User{
			ID:       uuid.New(),
			Email:    "test@example.com",
			FullName: "Test User",
		}

		err := svc.SendWelcomeEmail(context.Background(), user)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if mockEmail.sentTo != "test@example.com" {
			t.Errorf("expected sentTo to be test@example.com, got %s", mockEmail.sentTo)
		}
	})

	t.Run("missing email", func(t *testing.T) {
		mockEmail := &mockEmailProvider{}
		svc := service.NewNotificationService(mockEmail, userRepo, logger)

		user := entity.User{
			ID:       uuid.New(),
			Email:    "",
			FullName: "Test User",
		}

		err := svc.SendWelcomeEmail(context.Background(), user)
		if err != nil {
			t.Fatalf("expected no error when email is missing, got %v", err)
		}

		if mockEmail.sentTo != "" {
			t.Error("expected no email to be sent")
		}
	})

	t.Run("provider error", func(t *testing.T) {
		mockEmail := &mockEmailProvider{err: errors.New("smtp error")}
		svc := service.NewNotificationService(mockEmail, userRepo, logger)

		user := entity.User{
			ID:       uuid.New(),
			Email:    "test@example.com",
			FullName: "Test User",
		}

		err := svc.SendWelcomeEmail(context.Background(), user)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestNotificationService_SendPasswordResetEmail(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	userRepo := &mockUserRepoForNotification{admin: entity.User{ID: uuid.New(), Username: "admin"}}

	t.Run("success", func(t *testing.T) {
		mockEmail := &mockEmailProvider{}
		svc := service.NewNotificationService(mockEmail, userRepo, logger)

		user := entity.User{
			ID:       uuid.New(),
			Email:    "test@example.com",
			FullName: "Test User",
		}
		resetURL := "https://example.com/reset?token=123"

		err := svc.SendPasswordResetEmail(context.Background(), user, resetURL)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if mockEmail.sentTo != "test@example.com" {
			t.Errorf("expected sentTo to be test@example.com, got %s", mockEmail.sentTo)
		}
	})
}
