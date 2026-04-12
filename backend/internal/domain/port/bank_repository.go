// bank_repository.go
package port

import (
	"cogni-cash/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

type BankRepository interface {
	CreateConnection(ctx context.Context, conn *entity.BankConnection) error
	GetConnection(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.BankConnection, error)
	GetConnectionByRequisition(ctx context.Context, requisitionID string, userID uuid.UUID) (*entity.BankConnection, error)
	GetConnectionsByUserID(ctx context.Context, userID uuid.UUID) ([]entity.BankConnection, error)
	UpdateConnectionStatus(ctx context.Context, id uuid.UUID, status entity.ConnectionStatus, userID uuid.UUID) error
	UpdateRequisitionID(ctx context.Context, id uuid.UUID, requisitionID string, userID uuid.UUID) error
	DeleteConnection(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	UpsertAccounts(ctx context.Context, accounts []entity.BankAccount, userID uuid.UUID) error
	GetAccountByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*entity.BankAccount, error)
	GetAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]entity.BankAccount, error)
	GetAccountsByConnectionID(ctx context.Context, connectionID uuid.UUID, userID uuid.UUID) ([]entity.BankAccount, error)
	GetAccountByProviderID(ctx context.Context, providerAccountID string, userID uuid.UUID) (*entity.BankAccount, error)
	UpdateAccountBalance(ctx context.Context, id uuid.UUID, balance float64, syncedAt interface{}, errorMsg *string, userID uuid.UUID) error
	UpdateAccountType(ctx context.Context, id uuid.UUID, accType entity.StatementType, userID uuid.UUID) error
}
