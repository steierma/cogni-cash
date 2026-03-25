# Test Fixture Generators

These scripts regenerate the anonymised fixture files used by the parser unit tests.
All fixtures are stored under `backend/balance/` and are **fully synthetic** — no real
personal data is present.

## Fixtures

| File | Generator | Description |
|---|---|---|
| `balance/Girokonto_5437817550_Kontoauszug_20260301.pdf` | `gen_ing_pdf.go` | ING Girokonto PDF — Feb 2026, 8 transactions |
| `balance/Umsatzanzeige_02_2026.csv` | *(legacy — kept for ING CSV parser tests)* | ING Girokonto CSV — Feb 2026, 58 transactions |
| `balance/Umsatzanzeige_03_2026.csv` | `gen_csv.go` | ING Girokonto CSV — **Mar 2026, 40 transactions** — includes the Amazon Visa settlement debit |
| `balance/Amazon_Visa_25_11_2025_bis_13_03_2026.xls` | `gen_amazon_visa.py` | Amazon Visa XLS — Nov 2025 – Mar 2026, **30 transactions summing to exactly 2 500,00 EUR** |

## Reconciliation demo pair

`Umsatzanzeige_03_2026.csv` and `Amazon_Visa_25_11_2025_bis_13_03_2026.xls` form a
**matching pair** that demonstrates the credit-card reconciliation feature:

| Statement | File | Key transaction |
|---|---|---|
| Girokonto Mar 2026 | `Umsatzanzeige_03_2026.csv` | Row 1: `14.03.2026  Amazon Visa Kreditkartenabrechnung  -2.500,00 EUR` |
| Amazon Visa | `Amazon_Visa_25_11_2025_bis_13_03_2026.xls` | 30 individual purchases summing to **2 500,00 EUR** |

### How to use

1. Import **both** files via the Import page (or `POST /api/v1/bank-statements/import`).
2. On the Import page, note the **statement UUID** shown under the Amazon Visa XLS result.
3. Open the **Transactions** page, filter to the Girokonto March statement, and copy the
   `ContentHash` of the `14.03.2026 − −2.500,00` row (hover the row or use the API).
4. Reconcile:
   ```bash
   curl -X POST http://localhost:8080/api/v1/reconciliations \
     -H "Authorization: Bearer <token>" \
     -H "Content-Type: application/json" \
     -d '{
       "settlement_tx_hash":      "<content_hash from step 3>",
       "credit_card_statement_id": "<uuid from step 2>"
     }'
   ```
5. The Girokonto settlement row is now flagged **Reconciled** (amber badge in the UI)
   and excluded from `TotalExpense` / `NetSavings` — the individual Visa purchases
   remain visible in category and merchant charts.

## Regenerating

```bash
# From backend/
go run scripts/testdata/gen_ing_pdf.go   # requires github.com/jung-kurt/gofpdf (temp module)
go run scripts/testdata/gen_csv.go
python3 -m venv /tmp/xlvenv && /tmp/xlvenv/bin/pip install xlwt -q && \
  /tmp/xlvenv/bin/python3 scripts/testdata/gen_amazon_visa.py
```

Or use the Makefile target:
```bash
make gen-testdata
```
