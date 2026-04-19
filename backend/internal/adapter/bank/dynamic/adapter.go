package dynamic

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"cogni-cash/internal/adapter/bank/enablebanking"
	"cogni-cash/internal/adapter/bank/mock"
	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

type Adapter struct {
	settingsRepo port.SettingsRepository
	logger       *slog.Logger

	mu sync.Mutex

	ebAdapters   map[string]*enablebanking.Adapter
	ebPrivateKey *rsa.PrivateKey

	mockProvider *mock.MockBankProvider
	AllowMocks   bool
}

func NewAdapter(settingsRepo port.SettingsRepository, ebPrivateKey *rsa.PrivateKey, logger *slog.Logger) *Adapter {
	return &Adapter{
		settingsRepo: settingsRepo,
		ebPrivateKey: ebPrivateKey,
		logger:       logger,
		ebAdapters:   make(map[string]*enablebanking.Adapter),
		mockProvider: mock.NewMockBankProvider(),
	}
}

func (a *Adapter) getProvider(ctx context.Context, userID uuid.UUID) (port.BankProvider, error) {
	providerType, _ := a.settingsRepo.Get(ctx, "bank_provider", userID)

	if providerType == "" {
		providerType = "enablebanking"
	}

	if providerType == "enablebanking" {
		appID, _ := a.settingsRepo.Get(ctx, "enablebanking_app_id", userID)

		if appID == "" {
			return nil, fmt.Errorf("enablebanking configured but missing app_id in settings")
		}

		// Prüfen, ob der Key beim App-Start erfolgreich geladen wurde
		if a.ebPrivateKey == nil {
			return nil, fmt.Errorf("enablebanking configured but private key was not loaded from environment")
		}

		a.mu.Lock()
		defer a.mu.Unlock()

		if adapter, ok := a.ebAdapters[appID]; ok {
			return adapter, nil
		}

		// Erstelle neuen Adapter für diese AppID
		adapter := enablebanking.NewAdapter(appID, a.ebPrivateKey, a.logger)
		a.ebAdapters[appID] = adapter
		return adapter, nil
	}

	return nil, fmt.Errorf("unsupported bank provider: %s", providerType)
}

func (a *Adapter) GetInstitutions(ctx context.Context, userID uuid.UUID, countryCode string, isSandbox bool) ([]entity.BankInstitution, error) {
	p, err := a.getProvider(ctx, userID)
	if err != nil {
		return nil, err
	}
	return p.GetInstitutions(ctx, userID, countryCode, isSandbox)
}

func (a *Adapter) CreateRequisition(ctx context.Context, userID uuid.UUID, institutionID, institutionName, country, redirectURL, referenceID string, isSandbox bool) (*entity.BankConnection, error) {
	p, err := a.getProvider(ctx, userID)
	if err != nil {
		return nil, err
	}
	return p.CreateRequisition(ctx, userID, institutionID, institutionName, country, redirectURL, referenceID, isSandbox)
}

func (a *Adapter) ExchangeCodeForSession(ctx context.Context, userID uuid.UUID, code string) (string, error) {
	p, err := a.getProvider(ctx, userID)
	if err != nil {
		return "", err
	}
	return p.ExchangeCodeForSession(ctx, userID, code)
}

func (a *Adapter) GetRequisitionStatus(ctx context.Context, userID uuid.UUID, requisitionID string) (entity.ConnectionStatus, error) {
	if a.AllowMocks && strings.HasPrefix(requisitionID, "mock_") {
		return a.mockProvider.GetRequisitionStatus(ctx, userID, requisitionID)
	}

	p, err := a.getProvider(ctx, userID)
	if err != nil {
		return entity.StatusFailed, err
	}
	return p.GetRequisitionStatus(ctx, userID, requisitionID)
}

func (a *Adapter) FetchAccounts(ctx context.Context, userID uuid.UUID, requisitionID string) ([]entity.BankAccount, error) {
	if a.AllowMocks && strings.HasPrefix(requisitionID, "mock_") {
		return a.mockProvider.FetchAccounts(ctx, userID, requisitionID)
	}

	p, err := a.getProvider(ctx, userID)
	if err != nil {
		return nil, err
	}
	return p.FetchAccounts(ctx, userID, requisitionID)
}

func (a *Adapter) FetchTransactions(ctx context.Context, userID uuid.UUID, providerAccountID string, dateFrom *time.Time, dateTo *time.Time) ([]entity.Transaction, float64, error) {
	if a.AllowMocks && (providerAccountID == "dummy_acc_id" || strings.HasPrefix(providerAccountID, "mock_")) {
		return a.mockProvider.FetchTransactions(ctx, userID, providerAccountID, dateFrom, dateTo)
	}

	p, err := a.getProvider(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	return p.FetchTransactions(ctx, userID, providerAccountID, dateFrom, dateTo)
}
