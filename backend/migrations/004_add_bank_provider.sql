-- ============================================================
-- Add provider to bank_connections
-- =============================================

DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='bank_connections' AND column_name='provider') THEN
        ALTER TABLE bank_connections ADD COLUMN provider TEXT NOT NULL DEFAULT 'enablebanking';
    END IF;
END $$;
