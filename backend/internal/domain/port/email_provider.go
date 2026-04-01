package port

import (
	"context"

	"github.com/google/uuid"
)

// EmailProvider defines the driven port for sending emails.
type EmailProvider interface {
	Send(ctx context.Context, userID uuid.UUID, to, subject, body string) error
}
