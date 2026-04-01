-- ============================================================
-- Bank Integration (PSD2 / Enable Banking)
-- ============================================================

CREATE TABLE IF NOT EXISTS bank_connections (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    institution_id   TEXT          NOT NULL,
    institution_name TEXT          NOT NULL DEFAULT '',
    requisition_id   TEXT          NOT NULL UNIQUE,
    reference_id     TEXT          NOT NULL UNIQUE,
    status           TEXT          NOT NULL DEFAULT 'initialized',
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    expires_at       TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS bank_accounts (
    id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id       UUID          NOT NULL REFERENCES bank_connections(id) ON DELETE CASCADE,
    provider_account_id TEXT          NOT NULL UNIQUE,
    iban                TEXT          NOT NULL DEFAULT '',
    name                TEXT          NOT NULL DEFAULT '',
    currency            TEXT          NOT NULL DEFAULT 'EUR',
    balance             NUMERIC(15,2) NOT NULL DEFAULT 0,
    last_synced_at      TIMESTAMPTZ
);

-- Adjust transactions to allow linking to a bank_account instead of a static statement
ALTER TABLE transactions 
    ALTER COLUMN bank_statement_id DROP NOT NULL;

DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='transactions' AND column_name='bank_account_id') THEN
        ALTER TABLE transactions ADD COLUMN bank_account_id UUID REFERENCES bank_accounts(id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_transactions_bank_account_id ON transactions(bank_account_id);

DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='bank_statements' AND column_name='bank_account_id') THEN
        ALTER TABLE bank_statements ADD COLUMN bank_account_id UUID REFERENCES bank_accounts(id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_bank_statements_bank_account_id ON bank_statements(bank_account_id);
