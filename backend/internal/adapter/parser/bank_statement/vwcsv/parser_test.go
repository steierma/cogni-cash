package vwcsv

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestParser_ParseAnonymized(t *testing.T) {
	csvData := `Kontoinhaber;Max Mustermann;;;;;;;;;;;;;
;;;;;;;;;;;;
;Plus Konto online Nr. 1234567890;;;;;;;;;;;;;
Saldo (EUR);1219,75;;;;;;;;;;;;;
Zeitraum;01.01.2025 - 01.11.2025;;;;;;;;;;;;;
;;;;;;;;;;;;
Nr.;Buchungsdatum;Umsatzart;Umsatzinformation;UCI;Mandat ID;Abweichender Debitor;Abweichender Kreditor;Referenznummer;Wertstellung;Soll (EUR);Haben (EUR)
1;"30.10.2025";"Habenzinsen";"Interest calculation summary...";"";"";"";"";"";"30.10.2025";"";"1,38"
2;"20.10.2025";"Gutschrift";"M. Test description with fake IBAN: DE50721608180000000000";"";"";"";"";"NOTPROVIDED";"20.10.2025";"";"115,00"
5;"17.09.2025";"Telebanking Belastung";"Transfer to another account";"";"";"";"";"";"18.09.2025";"500,00";""
`

	p := NewParser()
	stmt, err := p.Parse(context.Background(), uuid.New(), []byte(csvData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if stmt.IBAN != "1234567890" {
		t.Errorf("Expected IBAN 1234567890, got %s", stmt.IBAN)
	}

	if len(stmt.Transactions) != 3 {
		t.Fatalf("Expected 3 transactions, got %d", len(stmt.Transactions))
	}

	// First transaction (Credit)
	tx1 := stmt.Transactions[0]
	if tx1.Amount != 1.38 {
		t.Errorf("Expected amount 1.38, got %f", tx1.Amount)
	}
	if tx1.Description != "Habenzinsen" {
		t.Errorf("Expected description Habenzinsen, got %s", tx1.Description)
	}

	// Last transaction (Debit)
	tx3 := stmt.Transactions[2]
	if tx3.Amount != -500.00 {
		t.Errorf("Expected amount -500.00, got %f", tx3.Amount)
	}
	if tx3.Description != "Telebanking Belastung" {
		t.Errorf("Expected description Telebanking Belastung, got %s", tx3.Description)
	}
}
