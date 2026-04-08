-- ============================================================
-- Add pg_trgm extension for fuzzy matching of transaction descriptions.
-- ============================================================
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Index for description and counterparty_name to speed up similarity searches
CREATE INDEX IF NOT EXISTS idx_transactions_description_trgm ON transactions USING gin (description gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_transactions_counterparty_trgm ON transactions USING gin (counterparty_name gin_trgm_ops);

-- Remove redundant fields from bank_statements
ALTER TABLE bank_statements DROP COLUMN IF EXISTS account_number;
ALTER TABLE bank_statements DROP COLUMN IF EXISTS bic;

-- Remove redundant fields from transactions
ALTER TABLE transactions DROP COLUMN IF EXISTS amount_base_currency;
ALTER TABLE transactions DROP COLUMN IF EXISTS exchange_rate;

-- Remove redundant fields from invoices
ALTER TABLE invoices DROP COLUMN IF EXISTS raw_text;
ALTER TABLE invoices DROP COLUMN IF EXISTS exchange_rate;
ALTER TABLE invoices DROP COLUMN IF EXISTS amount_base_currency;
ALTER TABLE invoices DROP COLUMN IF EXISTS original_file_mime;
ALTER TABLE invoices DROP COLUMN IF EXISTS original_file_size;

-- Remove redundant fields from payslips
DO $$
BEGIN
    -- Drop the old constraint that included employee_name
    IF EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'payslips_user_period_employer_unique' AND table_name = 'payslips') THEN
        ALTER TABLE payslips DROP CONSTRAINT payslips_user_period_employer_unique;
    END IF;

    -- Add the new constraint without employee_name
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'payslips_user_period_employer_unique' AND table_name = 'payslips') THEN
        ALTER TABLE payslips ADD CONSTRAINT payslips_user_period_employer_unique UNIQUE (user_id, period_month_num, period_year, employer_name);
    END IF;
END $$;

ALTER TABLE payslips DROP COLUMN IF EXISTS employee_name;
ALTER TABLE payslips DROP COLUMN IF EXISTS source_file;
ALTER TABLE payslips DROP COLUMN IF EXISTS original_file_mime;
ALTER TABLE payslips DROP COLUMN IF EXISTS original_file_size;
