//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func main() {
	_, srcFile, _, _ := runtime.Caller(0)
	outPath := filepath.Join(filepath.Dir(srcFile), "..", "..", "balance", "Umsatzanzeige_03_2026.csv")
	// 40 transactions for March 2026, newest first.
	//
	// RECONCILIATION NOTE
	// Row index 0 (14.03.2026) is the Amazon Visa credit-card settlement debit:
	//   Amount: -2.500,00 EUR  Party: "Amazon Visa Kreditkartenabrechnung"
	// This matches the Amazon Visa XLS fixture (total = 2.500,00 EUR exactly).
	// NewBalance = 1.247,55
	type row struct{ booking, valuta, party, btext, zweck, betrag string }
	rows := []row{
		// SETTLEMENT ROW - reconciles Amazon Visa XLS
		{"14.03.2026", "14.03.2026", "Amazon Visa Kreditkartenabrechnung", "Lastschrift", "Abrechnung 25.11.2025-13.03.2026", "-2.500,00"},
		// Regular March income and expenses
		{"27.03.2026", "27.03.2026", "Muster Arbeitgeber GmbH", "Gehalt/Rente", "LOHN GEHALT 03/2026", "4.752,01"},
		{"25.03.2026", "25.03.2026", "REWE SAGT DANKE", "Lastschrift", "KAUFUMSATZ 24.03 REWE", "-91,20"},
		{"24.03.2026", "24.03.2026", "Musterstadt Verkehrsbetriebe", "Lastschrift", "Monatskarte Maerz 2026", "-86,00"},
		{"23.03.2026", "23.03.2026", "Streaming Dienst GmbH", "Lastschrift", "Abo Maerz 2026", "-17,99"},
		{"22.03.2026", "22.03.2026", "Restaurant Zur Linde", "Lastschrift", "KAUFUMSATZ 21.03", "-38,50"},
		{"21.03.2026", "21.03.2026", "LIDL SAGT DANKE", "Lastschrift", "KAUFUMSATZ 20.03 LIDL", "-58,40"},
		{"20.03.2026", "20.03.2026", "Stadtwerke Musterstadt GmbH", "Lastschrift", "Strom Maerz 2026", "-110,00"},
		{"19.03.2026", "19.03.2026", "Online Versand GmbH", "Lastschrift", "Bestellung 2026-0015", "-28,99"},
		{"18.03.2026", "18.03.2026", "Apotheke am Marktplatz", "Lastschrift", "KAUFUMSATZ 18.03", "-22,80"},
		{"17.03.2026", "17.03.2026", "Musterbank Ruecklastschrift", "Lastschrift", "Darl.-Leistung Tilgung", "-167,70"},
		{"16.03.2026", "16.03.2026", "Musikschule Musterstadt e. V.", "Lastschrift", "Unterrichtsgebuehren Maerz 2026", "-69,00"},
		{"15.03.2026", "15.03.2026", "Krankenversicherung AG", "Lastschrift", "Krankenversicherung 03/2026", "-79,86"},
		{"13.03.2026", "13.03.2026", "Familienkasse Bundesagentur", "Gutschrift", "Kindergeld 03/2026", "259,00"},
		{"12.03.2026", "12.03.2026", "Wohnungsvermietung GmbH", "Dauerauftrag", "Kaltmiete Maerz 2026", "-950,00"},
		{"11.03.2026", "11.03.2026", "ALDI SUED", "Lastschrift", "KAUFUMSATZ 10.03 ALDI", "-47,30"},
		{"10.03.2026", "10.03.2026", "DB Bahn GmbH", "Ueberweisung", "Ticket 10.03.2026", "-34,90"},
		{"09.03.2026", "09.03.2026", "Tiernahrung Fachmarkt", "Lastschrift", "KAUFUMSATZ 08.03", "-67,81"},
		{"08.03.2026", "08.03.2026", "Drogerie Muster", "Lastschrift", "KAUFUMSATZ 07.03", "-19,80"},
		{"07.03.2026", "07.03.2026", "Energie Versorgung GmbH", "Lastschrift", "Gas Maerz 2026", "-95,00"},
		{"06.03.2026", "06.03.2026", "Sportverein Muster e. V.", "Lastschrift", "Beitrag 03/2026", "-37,00"},
		{"05.03.2026", "05.03.2026", "Kindergarten Musterstadt", "Ueberweisung", "Beitrag Maerz 2026", "-200,00"},
		{"04.03.2026", "04.03.2026", "PayPal Europe S.a.r.l.", "Lastschrift", "PP.MUSTER.010 Einkauf", "-33,49"},
		{"03.03.2026", "03.03.2026", "Mobilfunk Anbieter GmbH", "Lastschrift", "Rechnung Februar 2026", "-14,99"},
		{"03.03.2026", "03.03.2026", "Telekom Muster GmbH", "Lastschrift", "Festnetz 03/2026", "-74,95"},
		{"02.03.2026", "02.03.2026", "Biokiste Muster GmbH", "Lastschrift", "Rechnung 03/2026", "-48,06"},
		{"02.03.2026", "02.03.2026", "Gemuese Lieferservice", "Lastschrift", "Rechnung Nr. 465011", "-36,42"},
		{"02.03.2026", "02.03.2026", "Energie Solar GmbH", "Lastschrift", "INV01590000", "-52,00"},
		{"01.03.2026", "01.03.2026", "Mustermann M.", "Ueberweisung", "FMT 01.03.2026", "800,00"},
		{"01.03.2026", "01.03.2026", "Bausparkasse Muster AG", "Lastschrift", "Sparrate 03/2026", "-50,00"},
		{"01.03.2026", "01.03.2026", "Ratenzahlung Bank", "Lastschrift", "Ratenzahlung 01/2026", "-78,73"},
		{"01.03.2026", "01.03.2026", "Ratenzahlung Bank", "Lastschrift", "Ratenzahlung 02/2026", "-36,95"},
		{"01.03.2026", "01.03.2026", "Grundschule Musterstadt", "Ueberweisung", "Beitrag Maerz", "-30,00"},
		{"01.03.2026", "01.03.2026", "Mustermann M.", "Dauerauftrag", "Tierkonto", "-115,00"},
		{"01.03.2026", "01.03.2026", "Supermarkt E-Center Muster", "Lastschrift", "KAUFUMSATZ 28.02", "-41,20"},
		{"01.03.2026", "01.03.2026", "Musik Streaming GmbH", "Lastschrift", "Abo 03/2026", "-13,00"},
		{"01.03.2026", "01.03.2026", "Rundfunkanstalt Beitragsservice", "Lastschrift", "Rundfunkbeitrag Maerz 2026", "-18,36"},
		{"01.03.2026", "01.03.2026", "Krankenkasse Zusatz AG", "Gutschrift", "Erstattung Rechnung", "30,00"},
		{"01.03.2026", "01.03.2026", "Finanzamt Musterstadt", "Gutschrift", "Erstattung USt", "14,00"},
		{"01.03.2026", "01.03.2026", "Mustermann E.", "Ueberweisung", "Privatdarlehen Rueckzahlung", "1.300,00"},
	}
	if len(rows) != 40 {
		log.Fatalf("need exactly 40 rows, got %d", len(rows))
	}
	amounts := make([]float64, len(rows))
	for i, r := range rows {
		f, err := parseDE(r.betrag)
		if err != nil {
			log.Fatalf("row %d: bad amount %q: %v", i, r.betrag, err)
		}
		amounts[i] = f
	}
	newBalance := 1247.55
	var total float64
	for _, a := range amounts {
		total += a
	}
	oldBalance := newBalance - total
	saldi := make([]float64, len(rows))
	running := oldBalance
	for i := len(rows) - 1; i >= 0; i-- {
		running += amounts[i]
		saldi[i] = running
	}
	f, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("create: %v", err)
	}
	defer f.Close()
	w := func(s string) { fmt.Fprintln(f, s) }
	w("Umsatzanzeige;Datei erstellt am: 01.04.2026 08:00")
	w("")
	w("IBAN;DE89 3704 0044 0532 0130 00")
	w("Kontoname;Girokonto")
	w("Bank;ING")
	w("Kunde;Erika Mustermann")
	w("Zeitraum;01.03.2026 - 31.03.2026")
	w(fmt.Sprintf("Saldo;%s;EUR", fmtDE(newBalance)))
	w("")
	w("Sortierung;Erstbuchungsdatum, absteigend")
	w("Ihre Umsaetze im angegebenen Zeitraum")
	w("")
	w("Buchung;Wertstellungsdatum;Auftraggeber/Empfaenger;Buchungstext;Verwendungszweck;Saldo;Waehrung;Betrag;Waehrung")
	for i, r := range rows {
		w(fmt.Sprintf("%s;%s;%s;%s;%s;%s;EUR;%s;EUR",
			r.booking, r.valuta, r.party, r.btext, r.zweck,
			fmtDE(saldi[i]), r.betrag))
	}
	log.Printf("written %s (%d transactions, newBalance=%.2f, oldBalance=%.2f, total=%.2f)",
		outPath, len(rows), newBalance, oldBalance, total)
}
func parseDE(s string) (float64, error) {
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	return strconv.ParseFloat(s, 64)
}
func fmtDE(f float64) string {
	s := fmt.Sprintf("%.2f", f)
	return strings.ReplaceAll(s, ".", ",")
}
