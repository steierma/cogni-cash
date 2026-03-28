package port

import (
	"cogni-cash/internal/domain/entity"
	"context"
	"time"
)

type BankProvider interface {
	GetInstitutions(ctx context.Context, countryCode string, isSandbox bool) ([]entity.BankInstitution, error)
	CreateRequisition(ctx context.Context, institutionID, country, redirectURL, referenceID string, isSandbox bool) (*entity.BankConnection, error)
	ExchangeCodeForSession(ctx context.Context, code string) (string, error)
	GetRequisitionStatus(ctx context.Context, requisitionID string) (entity.ConnectionStatus, error)
	FetchAccounts(ctx context.Context, requisitionID string) ([]entity.BankAccount, error)
	FetchTransactions(ctx context.Context, providerAccountID string, dateFrom *time.Time, dateTo *time.Time) ([]entity.Transaction, float64, error)
}
