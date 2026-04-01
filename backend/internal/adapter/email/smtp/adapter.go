package smtp

import (
	"cogni-cash/internal/domain/port"
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"github.com/google/uuid"
)

type Adapter struct {
	settingsRepo port.SettingsRepository
	logger       *slog.Logger
}

func NewAdapter(settingsRepo port.SettingsRepository, logger *slog.Logger) *Adapter {
	return &Adapter{
		settingsRepo: settingsRepo,
		logger:       logger,
	}
}

func (a *Adapter) Send(ctx context.Context, userID uuid.UUID, to, subject, body string) error {
	host, _ := a.settingsRepo.Get(ctx, "smtp_host", userID)
	portStr, _ := a.settingsRepo.Get(ctx, "smtp_port", userID)
	user, _ := a.settingsRepo.Get(ctx, "smtp_user", userID)
	password, _ := a.settingsRepo.Get(ctx, "smtp_password", userID)
	from, _ := a.settingsRepo.Get(ctx, "smtp_from_email", userID)

	if host == "" {
		a.logger.Warn("SMTP not configured (smtp_host is empty). Email sending skipped.", "to", to, "subject", subject)
		return nil
	}

	addr := fmt.Sprintf("%s:%s", host, portStr)
	auth := smtp.PlainAuth("", user, password, host)

	header := make(map[string]string)
	header["From"] = from
	header["To"] = to
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""

	var message strings.Builder
	for k, v := range header {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(message.String()))
	if err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	a.logger.Info("Email sent successfully", "to", to, "subject", subject)
	return nil
}
