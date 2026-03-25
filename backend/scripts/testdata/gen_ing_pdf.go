package main

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/jung-kurt/gofpdf"
)

func main() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Helvetica", "", 9)

	writeLine := func(s string) {
		pdf.Cell(0, 5, s)
		pdf.Ln(5)
	}

	// ── Page 1 ──────────────────────────────────────────────────────────────
	pdf.AddPage()

	// Address block — fully anonymised.
	// extractAccountHolder walks backwards from "ING-DiBa AG" and returns the
	// first candidate that looks like a name, so the name must be the line
	// immediately before the ING-DiBa line (address lines follow the name here).
	writeLine("Musterstrasse 1")
	writeLine("12345 Musterstadt")
	writeLine("Erika Musterfrau")
	writeLine("und Max Mustermann")

	// Main header (triggers header-stripping logic up to "Valuta")
	writeLine("ING-DiBa AG  60628 Frankfurt am Main")
	writeLine("Datum")
	writeLine("01.03.2026")
	writeLine("Auszugsnummer")
	writeLine("2")
	writeLine("Eingerumte Kontoberziehung")
	writeLine("2.000,00 Euro")
	writeLine("Alter Saldo")
	writeLine("1.758,24 Euro")
	writeLine("Neuer Saldo")
	writeLine("3.503,82 Euro")
	writeLine("IBAN")
	writeLine("DE89 3704 0044 0532 0130 00")
	writeLine("BIC")
	writeLine("INGDDEFFXXX")
	writeLine("Seite")
	writeLine("1 von 2")

	// Legal footer (boilerplate, stripped by parser)
	writeLine("ING-DiBa AG  Theodor-Heuss-Allee 2  60486 Frankfurt am Main  Vorstand:")
	writeLine("Lars Stoy (Vorsitzender),")
	writeLine("Steuernummer: 014 220 2800 4  USt-IdNr.: DE 114 103 475  Mitglied im Einlagensicherungsfonds")

	// Column headers (boilerplate, skipped by isBoilerplate)
	writeLine("Buchung")
	writeLine("Buchung / Verwendungszweck")
	writeLine("Betrag (EUR)")
	writeLine("Valuta")

	// TX 1: Rent debit  01.02.2026
	writeLine("01.02.2026")
	writeLine("Dauerauftrag/Terminueberw.")
	writeLine("Wohnungsvermietung Muster GmbH")
	writeLine("-950,00")
	writeLine("01.02.2026")
	writeLine("Kaltmiete Februar 2026")

	// TX 2: Groceries debit  03.02.2026
	writeLine("03.02.2026")
	writeLine("Lastschrift")
	writeLine("REWE SAGT DANKE")
	writeLine("-112,35")
	writeLine("03.02.2026")
	writeLine("NR XXXX 0000 MUSTERSTADT DE KAUFUMSATZ 01.02 112.35")

	// TX 3: Insurance debit  05.02.2026
	writeLine("05.02.2026")
	writeLine("Lastschrift")
	writeLine("Muster Krankenversicherung AG")
	writeLine("-79,86")
	writeLine("05.02.2026")
	writeLine("Krankenversicherung Vertrag 0000000")
	writeLine("Mandat:")
	writeLine("MUSTER-MANDAT-001")
	writeLine("Referenz:")
	writeLine("MUSTER-REF-001")

	// TX 4: Streaming debit  10.02.2026
	writeLine("10.02.2026")
	writeLine("Lastschrift")
	writeLine("Streaming Dienst GmbH")
	writeLine("-17,99")
	writeLine("10.02.2026")
	writeLine("Mandat:")
	writeLine("MUSTER-MANDAT-002")
	writeLine("Referenz:")
	writeLine("MUSTER-REF-002")

	// ── Page 2 ──────────────────────────────────────────────────────────────
	pdf.AddPage()

	// Mini page header (stripped by stripAllPageHeaders up to "Valuta")
	writeLine("Girokonto Nummer 0532013000")
	writeLine("Kontoauszug Februar 2026")
	writeLine("Datum")
	writeLine("01.03.2026")
	writeLine("Seite")
	writeLine("2 von 2")
	writeLine("Buchung")
	writeLine("Buchung / Verwendungszweck")
	writeLine("Betrag (EUR)")
	writeLine("Valuta")

	// TX 5: Transport debit  12.02.2026
	writeLine("12.02.2026")
	writeLine("Lastschrift")
	writeLine("Bahn Muster AG")
	writeLine("-89,90")
	writeLine("12.02.2026")
	writeLine("Fahrkarte Musterstadt 12.02.2026")

	// TX 6: Utilities debit  15.02.2026
	writeLine("15.02.2026")
	writeLine("Lastschrift")
	writeLine("STADTWERKE MUSTERSTADT ENERGIE")
	writeLine("-135,00")
	writeLine("15.02.2026")
	writeLine("Strom Februar 2026")
	writeLine("Mandat:")
	writeLine("MUSTER-MANDAT-003")
	writeLine("Referenz:")
	writeLine("MUSTER-REF-003")

	// TX 7: Music school debit  17.02.2026
	writeLine("17.02.2026")
	writeLine("Lastschrift")
	writeLine("Musikschule Musterstadt e. V.")
	writeLine("-69,00")
	writeLine("17.02.2026")
	writeLine("Unterrichtsgebuehren Monat Februar 2026")
	writeLine("Mandat:")
	writeLine("MUSTER-MANDAT-004")
	writeLine("Referenz:")
	writeLine("MUSTER-REF-004")

	// TX 8: Salary credit  27.02.2026
	writeLine("27.02.2026")
	writeLine("Gehalt/Rente")
	writeLine("ARBEITGEBER MUSTER GMBH")
	writeLine("4.752,01")
	writeLine("27.02.2026")
	writeLine("LOHN GEHALT 02/2026")
	writeLine("Referenz:")
	writeLine("MUSTER-REF-005")

	// Footer sentinel — cutAtFooter trims everything from here
	writeLine("Kunden-Information")
	writeLine("Bitte beachten Sie die nachstehenden Hinweise.")

	_, srcFile, _, _ := runtime.Caller(0)
	out := filepath.Join(filepath.Dir(srcFile), "..", "..", "balance", "Girokonto_5437817550_Kontoauszug_20260301.pdf")
	if err := pdf.OutputFileAndClose(out); err != nil {
		log.Fatalf("pdf write error: %v", err)
	}
	log.Printf("written: %s", out)
}
