package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"log/slog"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/hash"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

type BankService struct {
	repo         port.BankRepository
	stmtRepo     port.BankStatementRepository
	settingsRepo port.SettingsRepository
	provider     port.BankProvider
	logger       *slog.Logger
}

func NewBankService(repo port.BankRepository, stmtRepo port.BankStatementRepository, settingsRepo port.SettingsRepository, provider port.BankProvider, logger *slog.Logger) *BankService {
	return &BankService{
		repo:         repo,
		stmtRepo:     stmtRepo,
		settingsRepo: settingsRepo,
		provider:     provider,
		logger:       logger,
	}
}

func (s *BankService) GetInstitutions(ctx context.Context, countryCode string, isSandbox bool) ([]entity.BankInstitution, error) {
	return s.provider.GetInstitutions(ctx, countryCode, isSandbox)
}

func (s *BankService) CreateConnection(ctx context.Context, userID uuid.UUID, institutionID string, country string, redirectURL string, isSandbox bool) (*entity.BankConnection, error) {
	referenceID := uuid.New().String()
	s.logger.Info("Initiating new bank connection", "institution_id", institutionID, "country", country, "user_id", userID)
	conn, err := s.provider.CreateRequisition(ctx, institutionID, country, redirectURL, referenceID, isSandbox)
	if err != nil {
		return nil, err
	}

	// Capture current provider
	provider, _ := s.settingsRepo.Get(ctx, "bank_provider")
	if provider == "" {
		provider = "gocardless"
	}
	conn.Provider = provider
	conn.UserID = userID

	if err := s.repo.CreateConnection(ctx, conn); err != nil {
		return nil, err
	}

	s.logger.Info("Bank connection initialized", "id", conn.ID, "requisition_id", conn.RequisitionID)
	return conn, nil
}

func (s *BankService) FinishConnection(ctx context.Context, requisitionID string, code string) error {
	s.logger.Info("Finishing bank connection", "requisition_id", requisitionID)
	conn, err := s.repo.GetConnectionByRequisition(ctx, requisitionID)
	if err != nil {
		return err
	}
	if conn == nil {
		return fmt.Errorf("connection not found for requisition: %s", requisitionID)
	}

	exchangeValue := requisitionID
	if code != "" {
		exchangeValue = code
	}

	// Exchange code for session
	sessionID, err := s.provider.ExchangeCodeForSession(ctx, exchangeValue)
	if err != nil {
		return fmt.Errorf("failed to exchange code for session: %w", err)
	}

	// If sessionID changed (Enable Banking), update it in the database
	if sessionID != requisitionID {
		s.logger.Info("Updating session ID for connection", "old", requisitionID, "new", sessionID)
		if err := s.repo.UpdateRequisitionID(ctx, conn.ID, sessionID); err != nil {
			return fmt.Errorf("failed to update session id: %w", err)
		}
		requisitionID = sessionID
	}

	var status entity.ConnectionStatus

	// Enable Banking uses standard OAuth; exchanging the code guarantees authorization.
	// GoCardless requires explicitly fetching the requisition status.
	if conn.Provider == "enablebanking" {
		status = entity.StatusLinked
	} else {
		status, err = s.provider.GetRequisitionStatus(ctx, requisitionID)
		if err != nil {
			return err
		}
	}

	if status == entity.StatusLinked {
		accounts, err := s.provider.FetchAccounts(ctx, requisitionID)
		if err != nil {
			return err
		}
		s.logger.Info("Fetched accounts for connection", "count", len(accounts), "connection_id", conn.ID)
		for i := range accounts {
			accounts[i].ConnectionID = conn.ID
		}
		if err := s.repo.UpsertAccounts(ctx, accounts); err != nil {
			return err
		}
	}

	s.logger.Info("Bank connection finalized", "connection_id", conn.ID, "status", status)
	return s.repo.UpdateConnectionStatus(ctx, conn.ID, status)
}

func (s *BankService) GetConnections(ctx context.Context, userID uuid.UUID) ([]entity.BankConnection, error) {
	conns, err := s.repo.GetConnectionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range conns {
		accs, err := s.repo.GetAccountsByConnectionID(ctx, conns[i].ID)
		if err == nil {
			conns[i].Accounts = accs
		}
	}

	return conns, nil
}

func (s *BankService) DeleteConnection(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("Deleting bank connection", "id", id)
	return s.repo.DeleteConnection(ctx, id)
}

func (s *BankService) SyncAccount(ctx context.Context, accountID uuid.UUID) error {
	s.logger.Info("Manually syncing account", "account_id", accountID)
	acc, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return err
	}
	if acc == nil {
		return fmt.Errorf("account not found: %s", accountID)
	}

	historyDaysStr, _ := s.settingsRepo.Get(ctx, "bank_sync_history_days")
	historyDays := 14
	if historyDaysStr != "" {
		if parsed, err := strconv.Atoi(historyDaysStr); err == nil {
			historyDays = parsed
		}
	}
	dateFrom := time.Now().AddDate(0, 0, -historyDays)

	txns, balance, err := s.provider.FetchTransactions(ctx, acc.ProviderAccountID, &dateFrom, nil)
	if err != nil {
		return err
	}

	for i := range txns {
		txns[i].BankAccountID = &acc.ID
		txns[i].StatementType = acc.AccountType
		txns[i].ContentHash = hash.ForTransaction(acc.IBAN, txns[i])
	}

	if err := s.stmtRepo.CreateTransactions(ctx, txns); err != nil {
		return err
	}

	s.logger.Info("Sync completed for account", "account_id", acc.ID, "new_txns", len(txns), "new_balance", balance)
	return s.repo.UpdateAccountBalance(ctx, acc.ID, balance, time.Now())
}

func (s *BankService) SyncAllAccounts(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("Scheduled sync for all accounts", "user_id", userID)
	conns, err := s.repo.GetConnectionsByUserID(ctx, userID)
	if err != nil {
		return err
	}

	historyDaysStr, _ := s.settingsRepo.Get(ctx, "bank_sync_history_days")
	historyDays := 14
	if historyDaysStr != "" {
		if parsed, err := strconv.Atoi(historyDaysStr); err == nil {
			historyDays = parsed
		}
	}
	dateFrom := time.Now().AddDate(0, 0, -historyDays)

	for _, conn := range conns {
		if conn.Status != entity.StatusLinked {
			continue
		}

		accs, err := s.repo.GetAccountsByConnectionID(ctx, conn.ID)
		if err != nil {
			continue
		}

		for _, acc := range accs {
			txns, balance, err := s.provider.FetchTransactions(ctx, acc.ProviderAccountID, &dateFrom, nil)
			if err != nil {
				s.logger.Error("failed to sync account", "account_id", acc.ID, "error", err)
				continue
			}
			s.logger.Info("fetched transactions for account", "account_id", acc.ID, "transaction_count", len(txns), "balance", balance)

			// Add AccountID, StatementType and ContentHash to transactions
			for i := range txns {
				txns[i].BankAccountID = &acc.ID
				txns[i].StatementType = acc.AccountType
				txns[i].ContentHash = hash.ForTransaction(acc.IBAN, txns[i])
			}

			if err := s.stmtRepo.CreateTransactions(ctx, txns); err != nil {
				s.logger.Error("failed to save synced transactions", "account_id", acc.ID, "error", err)
				continue
			}

			if err := s.repo.UpdateAccountBalance(ctx, acc.ID, balance, time.Now()); err != nil {
				s.logger.Error("failed to update account balance", "account_id", acc.ID, "error", err)
			}
		}
	}
	return nil
}

func (s *BankService) UpdateAccountType(ctx context.Context, accountID uuid.UUID, accType entity.StatementType) error {
	s.logger.Info("Updating account type", "account_id", accountID, "type", accType)
	return s.repo.UpdateAccountType(ctx, accountID, accType)
}
