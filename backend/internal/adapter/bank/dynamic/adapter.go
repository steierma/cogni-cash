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
	"cogni-cash/internal/adapter/bank/gocardless"
	"cogni-cash/internal/adapter/bank/mock"
	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
)

type Adapter struct {
	settingsRepo port.SettingsRepository
	logger       *slog.Logger

	mu       sync.Mutex
	cachedGC *gocardless.Adapter
	gcID     string
	gcKey    string

	cachedEB     *enablebanking.Adapter
	ebAppID      string
	ebPrivateKey *rsa.PrivateKey

	mockProvider *mock.MockBankProvider
	AllowMocks   bool
}

func NewAdapter(settingsRepo port.SettingsRepository, ebPrivateKey *rsa.PrivateKey, logger *slog.Logger) *Adapter {
	return &Adapter{
		settingsRepo: settingsRepo,
		ebPrivateKey: ebPrivateKey,
		logger:       logger,
		mockProvider: mock.NewMockBankProvider(),
	}
}

func (a *Adapter) getProvider(ctx context.Context) (port.BankProvider, error) {
	providerType, _ := a.settingsRepo.Get(ctx, "bank_provider")

	a.mu.Lock()
	defer a.mu.Unlock()

	if providerType == "enablebanking" {
		appID, _ := a.settingsRepo.Get(ctx, "enablebanking_app_id")

		if appID == "" {
			return nil, fmt.Errorf("enablebanking configured but missing app_id in settings")
		}

		// Prüfen, ob der Key beim App-Start erfolgreich geladen wurde
		if a.ebPrivateKey == nil {
			return nil, fmt.Errorf("enablebanking configured but private key was not loaded from environment")
		}

		if a.cachedEB == nil || a.ebAppID != appID {
			a.cachedEB = enablebanking.NewAdapter(appID, a.ebPrivateKey, a.logger)
			a.ebAppID = appID
		}
		return a.cachedEB, nil
	}

	// Default to GoCardless
	secretID, _ := a.settingsRepo.Get(ctx, "gocardless_secret_id")
	secretKey, _ := a.settingsRepo.Get(ctx, "gocardless_secret_key")

	if secretID == "" || secretKey == "" {
		return nil, fmt.Errorf("gocardless configured but missing secret_id or secret_key")
	}

	if a.cachedGC == nil || a.gcID != secretID || a.gcKey != secretKey {
		a.cachedGC = gocardless.NewAdapter(secretID, secretKey, a.logger)
		a.gcID = secretID
		a.gcKey = secretKey
	}
	return a.cachedGC, nil
}

func (a *Adapter) GetInstitutions(ctx context.Context, countryCode string, isSandbox bool) ([]entity.BankInstitution, error) {
	p, err := a.getProvider(ctx)
	if err != nil {
		return nil, err
	}
	return p.GetInstitutions(ctx, countryCode, isSandbox)
}

func (a *Adapter) CreateRequisition(ctx context.Context, institutionID, country, redirectURL, referenceID string, isSandbox bool) (*entity.BankConnection, error) {
	p, err := a.getProvider(ctx)
	if err != nil {
		return nil, err
	}
	return p.CreateRequisition(ctx, institutionID, country, redirectURL, referenceID, isSandbox)
}

func (a *Adapter) ExchangeCodeForSession(ctx context.Context, code string) (string, error) {
	p, err := a.getProvider(ctx)
	if err != nil {
		return "", err
	}
	return p.ExchangeCodeForSession(ctx, code)
}

func (a *Adapter) GetRequisitionStatus(ctx context.Context, requisitionID string) (entity.ConnectionStatus, error) {
	if a.AllowMocks && strings.HasPrefix(requisitionID, "mock_") {
		return a.mockProvider.GetRequisitionStatus(ctx, requisitionID)
	}

	p, err := a.getProvider(ctx)
	if err != nil {
		return entity.StatusFailed, err
	}
	return p.GetRequisitionStatus(ctx, requisitionID)
}

func (a *Adapter) FetchAccounts(ctx context.Context, requisitionID string) ([]entity.BankAccount, error) {
	if a.AllowMocks && strings.HasPrefix(requisitionID, "mock_") {
		return a.mockProvider.FetchAccounts(ctx, requisitionID)
	}

	p, err := a.getProvider(ctx)
	if err != nil {
		return nil, err
	}
	return p.FetchAccounts(ctx, requisitionID)
}

func (a *Adapter) FetchTransactions(ctx context.Context, providerAccountID string, dateFrom *time.Time, dateTo *time.Time) ([]entity.Transaction, float64, error) {
	if a.AllowMocks && (providerAccountID == "dummy_acc_id" || strings.HasPrefix(providerAccountID, "mock_")) {
		return a.mockProvider.FetchTransactions(ctx, providerAccountID, dateFrom, dateTo)
	}

	p, err := a.getProvider(ctx)
	if err != nil {
		return nil, 0, err
	}
	return p.FetchTransactions(ctx, providerAccountID, dateFrom, dateTo)
}
