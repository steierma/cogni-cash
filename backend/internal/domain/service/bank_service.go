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
	repo             port.BankRepository
	stmtRepo         port.BankStatementRepository
	settingsRepo     port.SettingsRepository
	plannedTxService port.PlannedTransactionUseCase
	discoveryService port.DiscoveryUseCase
	provider         port.BankProvider
	logger           *slog.Logger
}

func NewBankService(repo port.BankRepository, stmtRepo port.BankStatementRepository, settingsRepo port.SettingsRepository, plannedTxService port.PlannedTransactionUseCase, provider port.BankProvider, logger *slog.Logger) *BankService {
	return &BankService{
		repo:             repo,
		stmtRepo:         stmtRepo,
		settingsRepo:     settingsRepo,
		plannedTxService: plannedTxService,
		provider:         provider,
		logger:           logger,
	}
}

func (s *BankService) WithDiscoveryService(svc port.DiscoveryUseCase) *BankService {
	s.discoveryService = svc
	return s
}

func (s *BankService) GetInstitutions(ctx context.Context, userID uuid.UUID, countryCode string, isSandbox bool) ([]entity.BankInstitution, error) {
	return s.provider.GetInstitutions(ctx, userID, countryCode, isSandbox)
}

func (s *BankService) CreateConnection(ctx context.Context, userID uuid.UUID, institutionID string, institutionName string, country string, redirectURL string, isSandbox bool) (*entity.BankConnection, error) {
	referenceID := uuid.New().String()
	s.logger.Info("Initiating new bank connection", "institution_id", institutionID, "institution_name", institutionName, "country", country, "user_id", userID)
	conn, err := s.provider.CreateRequisition(ctx, userID, institutionID, institutionName, country, redirectURL, referenceID, isSandbox)
	if err != nil {
		return nil, err
	}

	// Capture current provider
	provider, _ := s.settingsRepo.Get(ctx, "bank_provider", userID)
	if provider == "" {
		provider = "enablebanking"
	}
	conn.Provider = provider
	conn.UserID = userID

	if err := s.repo.CreateConnection(ctx, conn); err != nil {
		return nil, err
	}

	s.logger.Info("Bank connection initialized", "id", conn.ID, "requisition_id", conn.RequisitionID, "user_id", userID)
	return conn, nil
}

func (s *BankService) FinishConnection(ctx context.Context, userID uuid.UUID, requisitionID string, code string) error {
	s.logger.Info("Finishing bank connection", "requisition_id", requisitionID, "user_id", userID)
	conn, err := s.repo.GetConnectionByRequisition(ctx, requisitionID, userID)
	if err != nil {
		return err
	}
	if conn == nil {
		return fmt.Errorf("connection not found for requisition: %s", requisitionID)
	}

	if conn.UserID != userID {
		return fmt.Errorf("unauthorized connection access")
	}

	exchangeValue := requisitionID
	if code != "" {
		exchangeValue = code
	}

	// Exchange code for session
	sessionID, err := s.provider.ExchangeCodeForSession(ctx, userID, exchangeValue)
	if err != nil {
		return fmt.Errorf("failed to exchange code for session: %w", err)
	}

	// If sessionID changed (Enable Banking), update it in the database
	if sessionID != requisitionID {
		s.logger.Info("Updating session ID for connection", "old", requisitionID, "new", sessionID)
		if err := s.repo.UpdateRequisitionID(ctx, conn.ID, sessionID, userID); err != nil {
			return fmt.Errorf("failed to update session id: %w", err)
		}
		requisitionID = sessionID
	}

	// Enable Banking uses standard OAuth; exchanging the code guarantees authorization.
	status := entity.StatusLinked

	if status == entity.StatusLinked {
		s.logger.Info("fetching accounts from provider", "requisition_id", requisitionID, "user_id", userID)
		accounts, err := s.provider.FetchAccounts(ctx, userID, requisitionID)
		if err != nil {
			s.logger.Error("failed to fetch accounts from provider", "requisition_id", requisitionID, "user_id", userID, "error", err)
			return err
		}
		s.logger.Info("Fetched accounts for connection", "count", len(accounts), "connection_id", conn.ID, "user_id", conn.UserID)
		for i := range accounts {
			accounts[i].ConnectionID = conn.ID
		}
		s.logger.Info("upserting accounts to database", "count", len(accounts), "connection_id", conn.ID, "user_id", userID)
		if err := s.repo.UpsertAccounts(ctx, accounts, userID); err != nil {
			s.logger.Error("failed to upsert accounts to database", "connection_id", conn.ID, "user_id", userID, "error", err)
			return err
		}
		s.logger.Info("successfully upserted accounts to database", "count", len(accounts), "connection_id", conn.ID, "user_id", userID)
	}

	s.logger.Info("Bank connection finalized", "connection_id", conn.ID, "status", status, "user_id", conn.UserID)
	return s.repo.UpdateConnectionStatus(ctx, conn.ID, status, userID)
}

func (s *BankService) GetConnections(ctx context.Context, userID uuid.UUID) ([]entity.BankConnection, error) {
	conns, err := s.repo.GetConnectionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range conns {
		accs, err := s.repo.GetAccountsByConnectionID(ctx, conns[i].ID, userID)
		if err == nil {
			conns[i].Accounts = accs
		}
	}

	return conns, nil
}

func (s *BankService) DeleteConnection(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	s.logger.Info("Deleting bank connection", "id", id, "user_id", userID)
	return s.repo.DeleteConnection(ctx, id, userID)
}

func (s *BankService) SyncAccount(ctx context.Context, accountID uuid.UUID, userID uuid.UUID) error {
	s.logger.Info("Manually syncing account", "account_id", accountID, "user_id", userID)
	acc, err := s.repo.GetAccountByID(ctx, accountID, userID)
	if err != nil {
		s.logger.Error("failed to get account by ID", "account_id", accountID, "user_id", userID, "error", err)
		return err
	}
	if acc == nil {
		s.logger.Warn("account not found for sync", "account_id", accountID, "user_id", userID)
		return fmt.Errorf("account not found: %s", accountID)
	}

	historyDaysStr, _ := s.settingsRepo.Get(ctx, "bank_sync_history_days", userID)
	historyDays := 14
	if historyDaysStr != "" {
		if parsed, err := strconv.Atoi(historyDaysStr); err == nil {
			historyDays = parsed
		}
	}
	dateFrom := time.Now().AddDate(0, 0, -historyDays)
	s.logger.Info("Syncing account with lookback", "account_id", accountID, "days", historyDays, "from", dateFrom.Format("2006-01-02"), "user_id", userID)

	txns, balance, err := s.provider.FetchTransactions(ctx, userID, acc.ProviderAccountID, &dateFrom, nil)
	if err != nil {
		s.logger.Error("provider failed to fetch transactions", "account_id", accountID, "provider_id", acc.ProviderAccountID, "user_id", userID, "error", err)
		errStr := err.Error()
		_ = s.repo.UpdateAccountBalance(ctx, acc.ID, acc.Balance, time.Now(), &errStr, userID)
		return err
	}

	s.logger.Info("Fetched transactions from provider", "account_id", accountID, "count", len(txns), "balance", balance, "user_id", userID)

	for i := range txns {
		txns[i].UserID = userID
		txns[i].BankAccountID = &acc.ID
		txns[i].StatementType = acc.AccountType
		txns[i].ContentHash = hash.ForTransaction(acc.IBAN, txns[i])
	}

	if err := s.stmtRepo.CreateTransactions(ctx, txns); err != nil {
		s.logger.Error("failed to save transactions to database", "account_id", accountID, "user_id", userID, "error", err)
		errStr := "Failed to save transactions: " + err.Error()
		_ = s.repo.UpdateAccountBalance(ctx, acc.ID, balance, time.Now(), &errStr, userID)
		return err
	}

	if s.discoveryService != nil && len(txns) > 0 {
		if err := s.discoveryService.MatchTransactions(ctx, userID, txns); err != nil {
			s.logger.Error("failed to match subscription transactions in sync", "user_id", userID, "error", err)
		}
	}

	s.logger.Info("Sync completed for account", "account_id", acc.ID, "new_txns", len(txns), "new_balance", balance, "user_id", userID)
	return s.repo.UpdateAccountBalance(ctx, acc.ID, balance, time.Now(), nil, userID)
}

func (s *BankService) SyncAllAccounts(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("=== BANK SYNC START ===", "user_id", userID)
	conns, err := s.repo.GetConnectionsByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get connections for user", "user_id", userID, "error", err)
		return err
	}

	s.logger.Info("Found bank connections for sync", "count", len(conns), "user_id", userID)

	historyDaysStr, _ := s.settingsRepo.Get(ctx, "bank_sync_history_days", userID)
	historyDays := 14
	if historyDaysStr != "" {
		if parsed, err := strconv.Atoi(historyDaysStr); err == nil {
			historyDays = parsed
		}
	}
	dateFrom := time.Now().AddDate(0, 0, -historyDays)
	s.logger.Info("Syncing all accounts with lookback", "days", historyDays, "from", dateFrom.Format("2006-01-02"), "user_id", userID)

	for _, conn := range conns {
		if conn.Status != entity.StatusLinked {
			s.logger.Info("Skipping connection in status", "connection_id", conn.ID, "status", conn.Status, "user_id", userID)
			continue
		}

		s.logger.Info("Syncing connection", "connection_id", conn.ID, "institution_name", conn.InstitutionName, "institution_id", conn.InstitutionID, "user_id", userID)
		accs, err := s.repo.GetAccountsByConnectionID(ctx, conn.ID, userID)
		if err != nil {
			s.logger.Error("failed to get accounts for connection", "connection_id", conn.ID, "user_id", userID, "error", err)
			continue
		}

		s.logger.Info("Found accounts for connection", "count", len(accs), "connection_id", conn.ID, "user_id", userID)

		for _, acc := range accs {
			s.logger.Info("Syncing account", "account_id", acc.ID, "provider_id", acc.ProviderAccountID, "user_id", userID)
			txns, balance, err := s.provider.FetchTransactions(ctx, userID, acc.ProviderAccountID, &dateFrom, nil)
			if err != nil {
				s.logger.Error("failed to fetch transactions for account", "account_id", acc.ID, "user_id", userID, "error", err)
				errStr := err.Error()
				_ = s.repo.UpdateAccountBalance(ctx, acc.ID, acc.Balance, time.Now(), &errStr, userID)
				continue
			}
			s.logger.Info("Fetched transactions for account", "account_id", acc.ID, "count", len(txns), "balance", balance, "user_id", userID)

			// Add AccountID, StatementType and ContentHash to transactions
			for i := range txns {
				txns[i].UserID = userID
				txns[i].BankAccountID = &acc.ID
				txns[i].StatementType = acc.AccountType
				txns[i].ContentHash = hash.ForTransaction(acc.IBAN, txns[i])
			}

			if err := s.stmtRepo.CreateTransactions(ctx, txns); err != nil {
				s.logger.Error("failed to save synced transactions", "account_id", acc.ID, "user_id", userID, "error", err)
				errStr := "Failed to save: " + err.Error()
				_ = s.repo.UpdateAccountBalance(ctx, acc.ID, balance, time.Now(), &errStr, userID)
				continue
			}

			if s.discoveryService != nil && len(txns) > 0 {
				if err := s.discoveryService.MatchTransactions(ctx, userID, txns); err != nil {
					s.logger.Error("failed to match subscription transactions in scheduled sync", "user_id", userID, "error", err)
				}
			}

			if err := s.repo.UpdateAccountBalance(ctx, acc.ID, balance, time.Now(), nil, userID); err != nil {
				s.logger.Error("failed to update account balance", "account_id", acc.ID, "user_id", userID, "error", err)
			}
			s.logger.Info("Sync completed for account", "account_id", acc.ID, "user_id", userID)
		}
	}

	s.logger.Info("Finished sync process for all accounts", "user_id", userID)
	return nil
}

func (s *BankService) UpdateAccountType(ctx context.Context, accountID uuid.UUID, accType entity.StatementType, userID uuid.UUID) error {
	s.logger.Info("Updating account type", "account_id", accountID, "type", accType, "user_id", userID)
	return s.repo.UpdateAccountType(ctx, accountID, accType, userID)
}
