package port

import (
	"cogni-cash/internal/domain/entity"
	"context"
	"time"

	"github.com/google/uuid"
)

type BankProvider interface {
	GetInstitutions(ctx context.Context, userID uuid.UUID, countryCode string, isSandbox bool) ([]entity.BankInstitution, error)
	CreateRequisition(ctx context.Context, userID uuid.UUID, institutionID, institutionName, country, redirectURL, referenceID string, isSandbox bool, ip string, userAgent string) (*entity.BankConnection, error)
	GenerateReauthLink(ctx context.Context, userID uuid.UUID, institutionID, country, redirectURL, referenceID string, isSandbox bool, ip string, userAgent string) (string, string, error)
	ExchangeCodeForSession(ctx context.Context, userID uuid.UUID, code string) (string, error)
	GetRequisitionStatus(ctx context.Context, userID uuid.UUID, requisitionID string) (entity.ConnectionStatus, error)
	FetchAccounts(ctx context.Context, userID uuid.UUID, requisitionID string) ([]entity.BankAccount, error)
	FetchTransactions(ctx context.Context, userID uuid.UUID, providerAccountID string, dateFrom *time.Time, dateTo *time.Time) ([]entity.Transaction, float64, error)
}
