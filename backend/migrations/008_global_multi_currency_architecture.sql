-- migration 008: Global Multi-Currency Architecture
-- This migration adds snapshot mapping columns for unified currency analytics.

-- 1. Transactions Expansion
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS base_amount NUMERIC(15,2) NOT NULL DEFAULT 0;
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS base_currency VARCHAR(3) NOT NULL DEFAULT 'EUR';

-- 2. Invoices Expansion
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS base_amount NUMERIC(15,2) NOT NULL DEFAULT 0;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS base_currency VARCHAR(3) NOT NULL DEFAULT 'EUR';

-- 3. Invoice Line Items Expansion (amount-only as currency is inherited)
ALTER TABLE invoice_line_items ADD COLUMN IF NOT EXISTS base_amount NUMERIC(15,2) NOT NULL DEFAULT 0;

-- 4. Planned Transactions Expansion
ALTER TABLE planned_transactions ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'EUR';
ALTER TABLE planned_transactions ADD COLUMN IF NOT EXISTS base_amount NUMERIC(15,2) NOT NULL DEFAULT 0;
ALTER TABLE planned_transactions ADD COLUMN IF NOT EXISTS base_currency VARCHAR(3) NOT NULL DEFAULT 'EUR';

-- 5. Payslips & Bonuses Expansion
ALTER TABLE payslips ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'EUR';
ALTER TABLE payslips ADD COLUMN IF NOT EXISTS base_gross_pay NUMERIC(15,2) NOT NULL DEFAULT 0;
ALTER TABLE payslips ADD COLUMN IF NOT EXISTS base_net_pay NUMERIC(15,2) NOT NULL DEFAULT 0;
ALTER TABLE payslips ADD COLUMN IF NOT EXISTS base_payout_amount NUMERIC(15,2) NOT NULL DEFAULT 0;

ALTER TABLE payslip_bonuses ADD COLUMN IF NOT EXISTS base_amount NUMERIC(15,2) NOT NULL DEFAULT 0;

-- 6. Backfill existing data
-- Initially, we assume base_amount = amount. Correct rates will be backfilled by a background task.
UPDATE transactions SET base_amount = amount, base_currency = currency;
UPDATE invoices SET base_amount = amount, base_currency = currency;
UPDATE invoice_line_items SET base_amount = amount;
UPDATE planned_transactions SET base_amount = amount, base_currency = currency;
UPDATE payslips SET base_gross_pay = gross_pay, base_net_pay = net_pay, base_payout_amount = payout_amount;
UPDATE payslip_bonuses SET base_amount = amount;

-- 7. Default Setting for current users
INSERT INTO settings (key, user_id, value, is_sensitive)
SELECT 'BASE_DISPLAY_CURRENCY', id, 'EUR'::bytea, false
FROM users
ON CONFLICT (key, user_id) DO NOTHING;
