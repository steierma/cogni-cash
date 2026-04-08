package service

import (
	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type NotificationService struct {
	emailProvider port.EmailProvider
	userRepo      port.UserRepository
	logger        *slog.Logger
}

func NewNotificationService(emailProvider port.EmailProvider, userRepo port.UserRepository, logger *slog.Logger) *NotificationService {
	return &NotificationService{
		emailProvider: emailProvider,
		userRepo:      userRepo,
		logger:        logger,
	}
}

func (s *NotificationService) getAdminID(ctx context.Context) uuid.UUID {
	admin, err := s.userRepo.FindByUsername(ctx, "admin")
	if err != nil {
		s.logger.Error("Failed to find admin user for system email", "error", err)
		return uuid.Nil
	}
	return admin.ID
}

func (s *NotificationService) SendWelcomeEmail(ctx context.Context, user entity.User) error {
	if user.Email == "" {
		s.logger.Warn("Cannot send welcome email: user has no email address", "user_id", user.ID)
		return nil
	}

	adminID := s.getAdminID(ctx)
	subject := "Welcome to Cogni-Cash!"
	body := fmt.Sprintf("Hello %s,\n\nWelcome to Cogni-Cash, your local AI financial manager.\n\nBest regards,\nThe Cogni-Cash Team", user.FullName)

	if err := s.emailProvider.Send(ctx, adminID, user.Email, subject, body); err != nil {
		s.logger.Error("Failed to send welcome email", "user_id", user.ID, "error", err)
		return err
	}

	return nil
}

func (s *NotificationService) SendPasswordResetEmail(ctx context.Context, user entity.User, resetURL string) error {
	if user.Email == "" {
		s.logger.Warn("Cannot send password reset email: user has no email address", "user_id", user.ID)
		return nil
	}

	adminID := s.getAdminID(ctx)
	subject := "Reset your Cogni-Cash password"
	body := fmt.Sprintf("Hello %s,\n\nYou requested to reset your password. Please use the following link:\n\n%s\n\nIf you did not request this, please ignore this email.\n\nBest regards,\nThe Cogni-Cash Team", user.FullName, resetURL)

	if err := s.emailProvider.Send(ctx, adminID, user.Email, subject, body); err != nil {
		s.logger.Error("Failed to send password reset email", "user_id", user.ID, "error", err)
		return err
	}

	return nil
}

func (s *NotificationService) SendTestEmail(ctx context.Context, to string, userID uuid.UUID) error {
	subject := "Cogni-Cash SMTP Test"
	body := "This is a test email from your Cogni-Cash instance. If you received this, your SMTP configuration is working correctly!"

	if err := s.emailProvider.Send(ctx, userID, to, subject, body); err != nil {
		s.logger.Error("Failed to send test email", "to", to, "user_id", userID, "error", err)
		return err
	}

	return nil
}
