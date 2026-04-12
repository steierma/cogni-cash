-- Missing Multi-Tenancy Indexes
CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories(user_id);
CREATE INDEX IF NOT EXISTS idx_settings_user_id ON settings(user_id);
CREATE INDEX IF NOT EXISTS idx_bank_connections_user_id ON bank_connections(user_id);
CREATE INDEX IF NOT EXISTS idx_bank_statements_user_id ON bank_statements(user_id);
CREATE INDEX IF NOT EXISTS idx_reconciliations_user_id ON reconciliations(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_invoices_user_id ON invoices(user_id);
CREATE INDEX IF NOT EXISTS idx_payslips_user_id ON payslips(user_id);

-- Stricter Data Types and Constraints

-- Color check constraint for categories
ALTER TABLE categories ADD CONSTRAINT check_color_hex CHECK (color ~* '^#[a-fA-F0-9]{6}$');

-- Currency length constraint for all relevant tables
-- First, cleanup empty strings that would violate the constraint
UPDATE bank_statements SET currency = 'EUR' WHERE currency = '' OR currency IS NULL;
UPDATE transactions SET currency = 'EUR' WHERE currency = '' OR currency IS NULL;
UPDATE bank_accounts SET currency = 'EUR' WHERE currency = '' OR currency IS NULL;
UPDATE invoices SET currency = 'EUR' WHERE currency = '' OR currency IS NULL;

ALTER TABLE bank_accounts ADD CONSTRAINT check_currency_len CHECK (length(currency) = 3);
ALTER TABLE bank_statements ADD CONSTRAINT check_currency_len CHECK (length(currency) = 3);
ALTER TABLE transactions ADD CONSTRAINT check_currency_len CHECK (length(currency) = 3);
ALTER TABLE invoices ADD CONSTRAINT check_currency_len CHECK (length(currency) = 3);

-- Soft Deletes for Categorization
ALTER TABLE categories ADD COLUMN deleted_at TIMESTAMPTZ;

-- Refresh Tokens for JWT Revocation
CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked    BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
