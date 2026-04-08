-- ============================================================
-- Add User Tenancy & Scope Deduplication to User
-- ============================================================

-- 0. Ensure Bootstrap Admin Exists
-- Solves the chicken-and-egg problem during a fresh reset so defaults can be assigned.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin') THEN
        INSERT INTO users (username, password_hash, email, role, full_name)
        VALUES ('admin', 'pending_setup', 'admin@localhost', 'admin', 'Administrator');
END IF;
END $$;

-- 1. Categories
ALTER TABLE categories ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;
UPDATE categories SET user_id = (SELECT id FROM users WHERE username = 'admin' LIMIT 1) WHERE user_id IS NULL;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'categories_name_key' AND table_name = 'categories') THEN
ALTER TABLE categories DROP CONSTRAINT categories_name_key;
END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'categories_name_user_unique' AND table_name = 'categories') THEN
ALTER TABLE categories ADD CONSTRAINT categories_name_user_unique UNIQUE (name, user_id);
END IF;
END $$;

-- 2. Invoices
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;
UPDATE invoices SET user_id = (SELECT id FROM users WHERE username = 'admin' LIMIT 1) WHERE user_id IS NULL;

-- Scope Invoice Hash to User
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'invoices_content_hash_key' AND table_name = 'invoices') THEN
ALTER TABLE invoices DROP CONSTRAINT invoices_content_hash_key;
END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'invoices_user_content_hash_unique' AND table_name = 'invoices') THEN
ALTER TABLE invoices ADD CONSTRAINT invoices_user_content_hash_unique UNIQUE (content_hash, user_id);
END IF;
END $$;

-- 3. Payslips
ALTER TABLE payslips ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE payslips ADD COLUMN IF NOT EXISTS employer_name VARCHAR(100) NOT NULL DEFAULT 'Unknown';
UPDATE payslips SET user_id = (SELECT id FROM users WHERE username = 'admin' LIMIT 1) WHERE user_id IS NULL;

-- Safety deduplication before adding constraint
DO $$
BEGIN
WITH duplicates AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY user_id, period_month_num, period_year, employee_name, employer_name ORDER BY created_at) as rn
    FROM payslips
)
UPDATE payslips
SET employer_name = employer_name || ' (duplicate ' || (duplicates.rn - 1) || ')'
    FROM duplicates
WHERE payslips.id = duplicates.id AND duplicates.rn > 1;
END $$;

DO $$
BEGIN
    -- Composite unique for period/employee/employer/user
    IF EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'payslips_period_month_num_period_year_employee_name_key' AND table_name = 'payslips') THEN
ALTER TABLE payslips DROP CONSTRAINT payslips_period_month_num_period_year_employee_name_key;
END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'payslips_user_period_employer_unique' AND table_name = 'payslips') THEN
ALTER TABLE payslips ADD CONSTRAINT payslips_user_period_employer_unique UNIQUE (user_id, period_month_num, period_year, employee_name, employer_name);
END IF;

    -- Scope Payslip Hash to User
    IF EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'payslips_content_hash_key' AND table_name = 'payslips') THEN
ALTER TABLE payslips DROP CONSTRAINT payslips_content_hash_key;
END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'payslips_user_content_hash_unique' AND table_name = 'payslips') THEN
ALTER TABLE payslips ADD CONSTRAINT payslips_user_content_hash_unique UNIQUE (content_hash, user_id);
END IF;
END $$;

-- 4. Bank Statements
ALTER TABLE bank_statements ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;
UPDATE bank_statements SET user_id = (SELECT id FROM users WHERE username = 'admin' LIMIT 1) WHERE user_id IS NULL;

-- Re-calculate user_id for statements that have a bank_account_id
UPDATE bank_statements bs
SET user_id = bc.user_id
    FROM bank_accounts ba
JOIN bank_connections bc ON ba.connection_id = bc.id
WHERE bs.bank_account_id = ba.id AND bs.user_id != bc.user_id;

-- Scope Statement Hash to User
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'bank_statements_content_hash_key' AND table_name = 'bank_statements') THEN
ALTER TABLE bank_statements DROP CONSTRAINT bank_statements_content_hash_key;
END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'bank_statements_user_content_hash_unique' AND table_name = 'bank_statements') THEN
ALTER TABLE bank_statements ADD CONSTRAINT bank_statements_user_content_hash_unique UNIQUE (content_hash, user_id);
END IF;
END $$;

-- 5. Transactions
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;
UPDATE transactions SET user_id = (SELECT id FROM users WHERE username = 'admin' LIMIT 1) WHERE user_id IS NULL;

-- Re-calculate user_id for transactions that have a bank_account_id
UPDATE transactions t
SET user_id = bc.user_id
    FROM bank_accounts ba
JOIN bank_connections bc ON ba.connection_id = bc.id
WHERE t.bank_account_id = ba.id AND t.user_id != bc.user_id;

-- Scope Transaction Hash to User
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'transactions_content_hash_key' AND table_name = 'transactions') THEN
ALTER TABLE transactions DROP CONSTRAINT transactions_content_hash_key;
END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'transactions_user_content_hash_unique' AND table_name = 'transactions') THEN
ALTER TABLE transactions ADD CONSTRAINT transactions_user_content_hash_unique UNIQUE (content_hash, user_id);
END IF;
END $$;

-- 6. Reconciliations
ALTER TABLE reconciliations ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;
UPDATE reconciliations SET user_id = (SELECT id FROM users WHERE username = 'admin' LIMIT 1) WHERE user_id IS NULL;

-- Scope Reconciliation Hashes to User
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'reconciliations_settlement_transaction_hash_key' AND table_name = 'reconciliations') THEN
ALTER TABLE reconciliations DROP CONSTRAINT reconciliations_settlement_transaction_hash_key;
END IF;
    IF EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'reconciliations_target_transaction_hash_key' AND table_name = 'reconciliations') THEN
ALTER TABLE reconciliations DROP CONSTRAINT reconciliations_target_transaction_hash_key;
END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'reconciliations_user_settlement_unique' AND table_name = 'reconciliations') THEN
ALTER TABLE reconciliations ADD CONSTRAINT reconciliations_user_settlement_unique UNIQUE (settlement_transaction_hash, user_id);
END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'reconciliations_user_target_unique' AND table_name = 'reconciliations') THEN
ALTER TABLE reconciliations ADD CONSTRAINT reconciliations_user_target_unique UNIQUE (target_transaction_hash, user_id);
END IF;
END $$;

-- 7. Settings (Transition to per-user settings)
ALTER TABLE settings ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;
UPDATE settings SET user_id = (SELECT id FROM users WHERE username = 'admin' LIMIT 1) WHERE user_id IS NULL;

-- Change Primary Key to (key, user_id)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints tc
        JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
        WHERE tc.table_name = 'settings' AND tc.constraint_type = 'PRIMARY KEY'
        GROUP BY tc.constraint_name HAVING COUNT(*) = 1
    ) THEN
ALTER TABLE settings DROP CONSTRAINT settings_pkey;
ALTER TABLE settings ADD PRIMARY KEY (key, user_id);
END IF;
END $$;