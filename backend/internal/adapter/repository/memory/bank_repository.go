package memory

import (
	"context"
	"sync"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

type BankRepository struct {
	mu          sync.RWMutex
	connections map[uuid.UUID]entity.BankConnection
	accounts    map[uuid.UUID]entity.BankAccount
}

func NewBankRepository() *BankRepository {
	return &BankRepository{
		connections: make(map[uuid.UUID]entity.BankConnection),
		accounts:    make(map[uuid.UUID]entity.BankAccount),
	}
}

func (r *BankRepository) CreateConnection(ctx context.Context, conn *entity.BankConnection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if conn.ID == uuid.Nil {
		conn.ID = uuid.New()
	}
	r.connections[conn.ID] = *conn
	return nil
}

func (r *BankRepository) GetConnection(ctx context.Context, id uuid.UUID) (*entity.BankConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	conn, ok := r.connections[id]
	if !ok {
		return nil, entity.ErrBankConnectionNotFound
	}
	return &conn, nil
}

func (r *BankRepository) GetConnectionByRequisition(ctx context.Context, requisitionID string) (*entity.BankConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, conn := range r.connections {
		if conn.RequisitionID == requisitionID {
			return &conn, nil
		}
	}
	return nil, entity.ErrBankConnectionNotFound
}

func (r *BankRepository) GetConnectionsByUserID(ctx context.Context, userID uuid.UUID) ([]entity.BankConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var conns []entity.BankConnection
	for _, conn := range r.connections {
		if conn.UserID == userID {
			conns = append(conns, conn)
		}
	}
	return conns, nil
}

func (r *BankRepository) UpdateConnectionStatus(ctx context.Context, id uuid.UUID, status entity.ConnectionStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	conn, ok := r.connections[id]
	if !ok {
		return entity.ErrBankConnectionNotFound
	}
	conn.Status = status
	r.connections[id] = conn
	return nil
}

func (r *BankRepository) UpdateRequisitionID(ctx context.Context, id uuid.UUID, requisitionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	conn, ok := r.connections[id]
	if !ok {
		return entity.ErrBankConnectionNotFound
	}
	conn.RequisitionID = requisitionID
	r.connections[id] = conn
	return nil
}

func (r *BankRepository) DeleteConnection(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.connections[id]; !ok {
		return entity.ErrBankConnectionNotFound
	}
	delete(r.connections, id)
	return nil
}

func (r *BankRepository) UpsertAccounts(ctx context.Context, accounts []entity.BankAccount) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range accounts {
		acc := accounts[i]
		if acc.ID == uuid.Nil {
			// Try to find by provider account ID
			found := false
			for _, existing := range r.accounts {
				if existing.ProviderAccountID == acc.ProviderAccountID {
					acc.ID = existing.ID
					found = true
					break
				}
			}
			if !found {
				acc.ID = uuid.New()
			}
		}
		r.accounts[acc.ID] = acc
		accounts[i] = acc
	}
	return nil
}

func (r *BankRepository) GetAccountByID(ctx context.Context, id uuid.UUID) (*entity.BankAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	acc, ok := r.accounts[id]
	if !ok {
		return nil, entity.ErrBankAccountNotFound
	}
	return &acc, nil
}

func (r *BankRepository) GetAccountsByConnectionID(ctx context.Context, connectionID uuid.UUID) ([]entity.BankAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var accs []entity.BankAccount
	for _, acc := range r.accounts {
		if acc.ConnectionID == connectionID {
			accs = append(accs, acc)
		}
	}
	return accs, nil
}

func (r *BankRepository) GetAccountByProviderID(ctx context.Context, providerAccountID string) (*entity.BankAccount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, acc := range r.accounts {
		if acc.ProviderAccountID == providerAccountID {
			return &acc, nil
		}
	}
	return nil, entity.ErrBankAccountNotFound
}

func (r *BankRepository) UpdateAccountBalance(ctx context.Context, id uuid.UUID, balance float64, syncedAt interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	acc, ok := r.accounts[id]
	if !ok {
		return entity.ErrBankAccountNotFound
	}
	acc.Balance = balance
	// In-memory, we don't strictly need to handle the syncedAt interface unless it's a specific type we want to store
	r.accounts[id] = acc
	return nil
}

func (r *BankRepository) UpdateAccountType(ctx context.Context, id uuid.UUID, accType entity.StatementType) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	acc, ok := r.accounts[id]
	if !ok {
		return entity.ErrBankAccountNotFound
	}
	acc.AccountType = accType
	r.accounts[id] = acc
	return nil
}

var _ port.BankRepository = (*BankRepository)(nil)
