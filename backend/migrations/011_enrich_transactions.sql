-- ============================================================
-- Add additional columns to the transactions table to store enriched transaction data such as counterparty name, IBAN, bank transaction code, and mandate reference.
-- ============================================================
ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS counterparty_name TEXT,
    ADD COLUMN IF NOT EXISTS counterparty_iban TEXT,
    ADD COLUMN IF NOT EXISTS bank_transaction_code TEXT,
    ADD COLUMN IF NOT EXISTS mandate_reference TEXT;
