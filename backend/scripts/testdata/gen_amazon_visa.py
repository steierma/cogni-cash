#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Generate a synthetic Amazon Visa BIFF8 XLS fixture for tests.

All personal data is anonymised. Merchant names, amounts, and dates are
synthetic.

Transaction layout (31 rows total):
  Rows 1-30  : individual purchases, summing to exactly -2 500,00 EUR
  Row  31    : Girokonto payment credit of +2 500,00 EUR (14.03.2026)
               This mirrors the -2.500,00 debit in Umsatzanzeige_03_2026.csv

RECONCILIATION NOTE
-------------------
The +2.500,00 credit row on the Visa XLS and the -2.500,00 debit row on the
Girokonto CSV are two sides of the same money movement.  Import both files,
then reconcile via POST /api/v1/reconciliations to prevent double-counting.

After reconciliation:
  - The Girokonto -2.500,00 debit is excluded from TotalExpense / NetSavings.
  - All individual Visa purchases remain in CategoryTotals / TopMerchants.
  - The +2.500,00 credit on the Visa is visible but does NOT inflate income
    (it is a liability repayment, not real income).

Requirements:
    pip install xlwt
"""
import sys
import os

try:
    import xlwt
except ImportError:
    print("ERROR: xlwt not installed. Run: pip3 install xlwt", file=sys.stderr)
    sys.exit(1)

OUT = os.path.join(os.path.dirname(__file__), "..", "..", "balance", "Amazon_Visa_25_11_2025_bis_13_03_2026.xls")

wb = xlwt.Workbook(encoding="utf-8")
ws = wb.add_sheet("Umsaetze")

# ── metadata rows (0-based) ──────────────────────────────────────────────────
ws.write(0, 0, "")
ws.write(1, 0, "Amazon Visa - Umsaetze")
ws.write(2, 0, "")
ws.write(3, 0, "Datum der Belastung:")
ws.write(3, 1, "13.03.2026, 10:00 Uhr")
ws.write(4, 0, "Karteninhaber:")
ws.write(4, 1, "Max Mustermann")
ws.write(5, 0, "Referenzkonto:")
ws.write(5, 1, "DE89 3704 0044 0532 0130 00")   # matches Girokonto IBAN
ws.write(6, 0, "Zeitraum der Bewegung:")
ws.write(6, 1, "25.11.2025 - 18.03.2026")
ws.write(7, 0, "Kreditkartenlimit:")
ws.write(7, 1, "5.000,00 EUR")
ws.write(8, 0, "Verbraucht:")
ws.write(8, 1, "-2.500,00 \u20ac")   # total spent (purchases only, unchanged)
ws.write(9, 0, "")

# column headers (row 10)
for col, h in enumerate(["Datum", "Zeit", "Karte", "Beschreibung", "Umsatzkategorie", "", "Betrag", "Punkte"]):
    ws.write(10, col, h)
ws.write(11, 0, "")  # blank spacer row

# ── transactions ─────────────────────────────────────────────────────────────
# Rows 1-30: purchases, amount_cents is positive (written as negative EUR).
# Row  31  : payment credit, amount_cents is negative sentinel (-250000)
#            written as "+2.500,00 €".
#
# Categories use canonical German names so they map directly to the seeded
# categories table.

CARD = "************0000"
TIME = "10:00 Uhr"

# (date, description, category, amount_cents)
# Positive cents = purchase (debit on card), written as negative EUR.
# Negative cents = payment (credit on card), written as positive EUR.
TX_DATA = [
    # Nov 2025 — purchases
    ("25.11.2025", "Lebensmittelmarkt Muster",          "Lebensmittel und Drogerie",          8749),
    ("28.11.2025", "Tankstelle Muster",                 "Reise und Transport",                5520),
    ("30.11.2025", "Online Versand Muster GmbH",        "Handel und Geschaefte",              12399),
    # Dec 2025 — purchases
    ("03.12.2025", "Restaurant Zur Linde",              "Restaurants und Lokale",             6750),
    ("07.12.2025", "Elektronik Muster GmbH",            "Handel und Geschaefte",              18990),
    ("10.12.2025", "Apotheke am Marktplatz",            "Gesundheit",                         3499),
    ("14.12.2025", "Modehaus Muster",                   "Handel und Geschaefte",              8950),
    ("17.12.2025", "Buchhandel Muster",                 "Bildung",                            4299),
    ("20.12.2025", "Supermarkt Muster",                 "Lebensmittel und Drogerie",          7640),
    ("23.12.2025", "Spielzeug Muster",                  "Handel und Geschaefte",              5599),
    ("27.12.2025", "Haushalt Muster",                   "Handel und Geschaefte",              3150),
    ("30.12.2025", "Kino Musterstadt",                  "Unterhaltung und Medien",            2400),
    # Jan 2026 — purchases
    ("03.01.2026", "Drogerie Muster",                   "Lebensmittel und Drogerie",          4210),
    ("06.01.2026", "Sportartikel Muster GmbH",          "Handel und Geschaefte",              9900),
    ("10.01.2026", "Tiernahrung Fachmarkt",             "Lebensmittel und Drogerie",          6781),
    ("13.01.2026", "Fahrradladen Muster",               "Reise und Transport",                14900),
    ("17.01.2026", "Blumenmarkt Muster",                "Handel und Geschaefte",              2199),
    ("20.01.2026", "Baeckerei Muster",                  "Restaurants und Lokale",             1350),
    ("24.01.2026", "Schreibwaren Muster",               "Bildung",                            1890),
    ("28.01.2026", "Moebel Muster GmbH",                "Handel und Geschaefte",              22900),
    # Feb 2026 — purchases
    ("02.02.2026", "Garten Fachmarkt Muster",           "Handel und Geschaefte",              4799),
    ("05.02.2026", "Lebensmittelmarkt Muster",          "Lebensmittel und Drogerie",          9320),
    ("09.02.2026", "Restaurant Zur Linde",              "Restaurants und Lokale",             5450),
    ("12.02.2026", "Tankstelle Muster",                 "Reise und Transport",                4810),
    ("16.02.2026", "Online Versand Muster GmbH",        "Handel und Geschaefte",              7699),
    ("20.02.2026", "Optik Muster",                      "Gesundheit",                         11900),
    ("23.02.2026", "Heimwerker Muster",                 "Handel und Geschaefte",              6250),
    ("26.02.2026", "Bekleidung Muster GmbH",            "Handel und Geschaefte",              8400),
    # Mar 2026 — last purchase (absorbs rounding remainder to hit TARGET_CENTS exactly)
    ("03.03.2026", "Supermarkt Muster",                 "Lebensmittel und Drogerie",          6140),
    ("12.03.2026", "Elektromarkt Muster",               "Handel und Geschaefte",              0),   # placeholder
    # ── PAYMENT ROW ──────────────────────────────────────────────────────────
    # 14.03.2026: Girokonto repayment credit of +2 500,00 EUR.
    # This is the mirror of the -2.500,00 debit in Umsatzanzeige_03_2026.csv.
    # amount_cents is negative here (sentinel) -- written as a positive EUR string.
    ("15.03.2026", "Girokonto-Zahlung Kreditkartenabrechnung", "Ueberweisungen, Bankkosten und Darlehen", -250000),
]

TARGET_PURCHASE_CENTS = 250000  # 2 500,00 EUR — purchases only

# Fill the remainder into the placeholder purchase row (index 29)
sum_purchases = sum(amt for _, _, _, amt in TX_DATA[:29])
remainder     = TARGET_PURCHASE_CENTS - sum_purchases
TX_DATA[29]   = (TX_DATA[29][0], TX_DATA[29][1], TX_DATA[29][2], remainder)

purchase_total = sum(amt for _, _, _, amt in TX_DATA if amt > 0)
assert purchase_total == TARGET_PURCHASE_CENTS, \
    f"Purchase total mismatch: {purchase_total} != {TARGET_PURCHASE_CENTS}"
assert len(TX_DATA) == 31, f"Expected 31 rows, got {len(TX_DATA)}"

for i, (date_str, desc, cat, amt_cents) in enumerate(TX_DATA):
    excel_row = 12 + i
    if amt_cents < 0:
        # Credit / payment row: positive EUR amount
        amount = abs(amt_cents) / 100.0
        betrag = "+{:.2f} \u20ac".format(amount).replace(".", ",", 1)
        # betrag format: "+2.500,00 €"
        betrag = "+{} \u20ac".format("{:.2f}".format(amount).replace(".", ","))
        pts    = "0"
    else:
        # Purchase / debit row: negative EUR amount
        amount = -amt_cents / 100.0
        betrag = "{:.2f}".format(amount).replace(".", ",") + " \u20ac"
        pts    = "+{}".format(int(amt_cents // 100))

    ws.write(excel_row, 0, date_str)
    ws.write(excel_row, 1, TIME)
    ws.write(excel_row, 2, CARD)
    ws.write(excel_row, 3, desc)
    ws.write(excel_row, 4, cat)
    ws.write(excel_row, 5, "")
    ws.write(excel_row, 6, betrag)
    ws.write(excel_row, 7, pts)

wb.save(OUT)
print(f"Written {len(TX_DATA)} transactions to {OUT}")
print(f"Purchase total : {purchase_total / 100:.2f} EUR  (must be 2500.00)")
print(f"Payment credit : +2500.00 EUR on 14.03.2026  (mirrors Girokonto debit)")
print()
print("Reconciliation instructions:")
print("  1. Import this file  ->  note the statement UUID")
print("  2. Import Umsatzanzeige_03_2026.csv")
print("  3. Find ContentHash of the 14.03.2026 -2500.00 row in TransactionsPage")
print("  4. POST /api/v1/reconciliations")
print('       { "settlement_tx_hash": "<hash>",')
print('         "credit_card_statement_id": "<uuid>" }')
