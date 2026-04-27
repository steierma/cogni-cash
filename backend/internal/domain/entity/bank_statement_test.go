package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBankStatement_IsValid(t *testing.T) {
	uID := uuid.New()
	t1 := Transaction{ID: uuid.New(), Amount: 100}

	tests := []struct {
		name    string
		s       BankStatement
		wantErr bool
	}{
		{
			name: "valid statement",
			s: BankStatement{
				IBAN:          "DE123",
				StatementDate: time.Now(),
				Transactions:  []Transaction{t1},
			},
			wantErr: false,
		},
		{
			name: "missing IBAN",
			s: BankStatement{
				StatementDate: time.Now(),
				Transactions:  []Transaction{t1},
			},
			wantErr: true,
		},
		{
			name: "missing date",
			s: BankStatement{
				IBAN:         "DE123",
				Transactions: []Transaction{t1},
			},
			wantErr: true,
		},
		{
			name: "no transactions",
			s: BankStatement{
				IBAN:          "DE123",
				StatementDate: time.Now(),
				Transactions:  []Transaction{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.UserID = uID
			if err := tt.s.IsValid(); (err != nil) != tt.wantErr {
				t.Errorf("BankStatement.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
