package mock

import (
	"context"
	"fmt"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

type MockBankProvider struct{}

func NewMockBankProvider() *MockBankProvider {
	return &MockBankProvider{}
}

func (m *MockBankProvider) GetInstitutions(ctx context.Context, countryCode string, isSandbox bool) ([]entity.BankInstitution, error) {
	return []entity.BankInstitution{
		{ID: "SANDBOX_ID", Name: "Mock Bank DE", Bic: "MOCKDEFF", Country: "DE"},
		{ID: "SANDBOX_ID_2", Name: "Mock Bank DE 2", Bic: "MOCKDEFF2", Country: "DE"},
	}, nil
}

func (m *MockBankProvider) CreateRequisition(ctx context.Context, institutionID, country, redirectURL, referenceID string, isSandbox bool) (*entity.BankConnection, error) {
	return &entity.BankConnection{
		ID:              uuid.New(),
		Provider:        "mock",
		InstitutionID:   institutionID,
		InstitutionName: "Mock Bank",
		RequisitionID:   "mock_requisition_" + uuid.New().String(),
		ReferenceID:     referenceID,
		Status:          entity.StatusInitialized,
		AuthLink:        redirectURL + "?code=mock_code",
	}, nil
}

func (m *MockBankProvider) ExchangeCodeForSession(ctx context.Context, code string) (string, error) {
	return "mock_session_" + uuid.New().String(), nil
}

func (m *MockBankProvider) GetRequisitionStatus(ctx context.Context, requisitionID string) (entity.ConnectionStatus, error) {
	return entity.StatusLinked, nil
}

func (m *MockBankProvider) FetchAccounts(ctx context.Context, requisitionID string) ([]entity.BankAccount, error) {
	return []entity.BankAccount{
		{
			ID:                uuid.New(),
			ProviderAccountID: "mock_acc_1",
			IBAN:              "DE1234567890",
			Name:              "Mock Checking Account",
			Currency:          "EUR",
			Balance:           1234.56,
			AccountType:       entity.StatementTypeGiro,
		},
	}, nil
}

func (m *MockBankProvider) FetchTransactions(ctx context.Context, providerAccountID string, dateFrom *time.Time, dateTo *time.Time) ([]entity.Transaction, float64, error) {
	now := time.Now()
	txns := []entity.Transaction{
		{
			ID:            uuid.New(),
			BookingDate:   now.AddDate(0, 0, -1),
			Description:   "Mock Transaction 1",
			Amount:        -50.00,
			Currency:      "EUR",
			Type:          entity.TransactionTypeDebit,
			ContentHash:   fmt.Sprintf("mock_hash_%d", now.Unix()),
			StatementType: entity.StatementTypeGiro,
			Reviewed:      false,
		},
	}
	return txns, 1184.56, nil
}

var _ port.BankProvider = (*MockBankProvider)(nil)
