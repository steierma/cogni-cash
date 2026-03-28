-- ============================================================
-- Add Account Type to Bank Accounts
-- ============================================================

DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='bank_accounts' AND column_name='account_type') THEN
        ALTER TABLE bank_accounts ADD COLUMN account_type TEXT NOT NULL DEFAULT 'giro';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='transactions' AND column_name='statement_type') THEN
        ALTER TABLE transactions ADD COLUMN statement_type TEXT;
    END IF;
END $$;
