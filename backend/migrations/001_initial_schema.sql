-- =============================================================================
-- Squashed Initial Schema (001–017)
-- Defines the complete, final database state for a fresh installation.
-- Replace all individual migration files for new environments.
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ============================================================
-- Users
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    username      TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    email         TEXT        NOT NULL UNIQUE,
    full_name     TEXT        NOT NULL DEFAULT '',
    address       TEXT        NOT NULL DEFAULT '',
    role          TEXT        NOT NULL DEFAULT 'manager',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Bootstrap Admin User
-- ============================================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin') THEN
        INSERT INTO users (username, password_hash, email, role, full_name)
        VALUES ('admin', 'pending_setup', 'admin@localhost', 'admin', 'Administrator');
    END IF;
END $$;

UPDATE users SET role = 'admin' WHERE username = 'admin';

-- ============================================================
-- Password Reset Tokens
-- ============================================================
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reset_tokens_hash    ON password_reset_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_reset_tokens_expires ON password_reset_tokens(expires_at);

-- ============================================================
-- Categories
-- ============================================================
CREATE TABLE IF NOT EXISTS categories (
    id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID        REFERENCES users(id) ON DELETE CASCADE,
    name                 TEXT        NOT NULL,
    color                TEXT        NOT NULL DEFAULT '#6366f1',
    is_variable_spending BOOLEAN     NOT NULL DEFAULT false,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT categories_name_user_unique UNIQUE (name, user_id)
);

INSERT INTO categories (name, color, user_id)
SELECT name, color, (SELECT id FROM users WHERE username = 'admin' LIMIT 1)
FROM (VALUES
    ('Haus und Hausrat',                         '#f59e0b'),
    ('Bildung, Gesundheit, Beauty und Wellness', '#ec4899'),
    ('Sonstige Ausgaben',                        '#6b7280'),
    ('Einkommen',                                '#22c55e')
) AS v(name, color)
ON CONFLICT (name, user_id) DO NOTHING;

-- ============================================================
-- Settings (per-user)
-- ============================================================
CREATE TABLE IF NOT EXISTS settings (
    key     TEXT NOT NULL,
    value   TEXT NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (key, user_id)
);

INSERT INTO settings (key, value, user_id)
SELECT 'base_currency', 'EUR', id FROM users WHERE username = 'admin'
ON CONFLICT (key, user_id) DO NOTHING;

-- ============================================================
-- Bank Connections
-- ============================================================
CREATE TABLE IF NOT EXISTS bank_connections (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    institution_id   TEXT        NOT NULL,
    institution_name TEXT        NOT NULL DEFAULT '',
    requisition_id   TEXT        NOT NULL UNIQUE,
    reference_id     TEXT        NOT NULL UNIQUE,
    status           TEXT        NOT NULL DEFAULT 'initialized',
    provider         TEXT        NOT NULL DEFAULT 'enablebanking',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at       TIMESTAMPTZ
);

-- ============================================================
-- Bank Accounts
-- ============================================================
CREATE TABLE IF NOT EXISTS bank_accounts (
    id                  UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id       UUID          NOT NULL REFERENCES bank_connections(id) ON DELETE CASCADE,
    provider_account_id TEXT          NOT NULL UNIQUE,
    iban                TEXT          NOT NULL DEFAULT '',
    name                TEXT          NOT NULL DEFAULT '',
    currency            TEXT          NOT NULL DEFAULT 'EUR',
    balance             NUMERIC(15,2) NOT NULL DEFAULT 0,
    account_type        TEXT          NOT NULL DEFAULT 'giro',
    last_synced_at      TIMESTAMPTZ,
    last_sync_error     TEXT
);

-- ============================================================
-- Bank Statements
-- ============================================================
CREATE TABLE IF NOT EXISTS bank_statements (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID          REFERENCES users(id) ON DELETE CASCADE,
    bank_account_id UUID          REFERENCES bank_accounts(id) ON DELETE SET NULL,
    account_holder  TEXT          NOT NULL DEFAULT '',
    iban            TEXT          NOT NULL DEFAULT '',
    statement_date  DATE,
    statement_no    INT           NOT NULL DEFAULT 0,
    old_balance     NUMERIC(15,2) NOT NULL DEFAULT 0,
    new_balance     NUMERIC(15,2) NOT NULL DEFAULT 0,
    currency        TEXT          NOT NULL DEFAULT 'EUR',
    source_file     TEXT          NOT NULL DEFAULT '',
    imported_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    content_hash    TEXT          NOT NULL,
    original_file   BYTEA,
    statement_type  TEXT          NOT NULL DEFAULT 'giro',
    CONSTRAINT bank_statements_user_content_hash_unique UNIQUE (content_hash, user_id)
);

CREATE INDEX IF NOT EXISTS idx_bank_statements_bank_account_id ON bank_statements(bank_account_id);

-- ============================================================
-- Reconciliations
-- ============================================================
CREATE TABLE IF NOT EXISTS reconciliations (
    id                          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                     UUID          REFERENCES users(id) ON DELETE CASCADE,
    settlement_transaction_hash TEXT          NOT NULL,
    target_transaction_hash     TEXT,
    amount                      NUMERIC(15,2) NOT NULL,
    reconciled_at               TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CONSTRAINT reconciliations_user_settlement_unique UNIQUE (settlement_transaction_hash, user_id),
    CONSTRAINT reconciliations_user_target_unique     UNIQUE (target_transaction_hash, user_id)
);

CREATE INDEX IF NOT EXISTS idx_reconciliations_target_transaction ON reconciliations(target_transaction_hash);

-- ============================================================
-- Transactions
-- ============================================================
CREATE TABLE IF NOT EXISTS transactions (
    id                    UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               UUID          REFERENCES users(id) ON DELETE CASCADE,
    bank_statement_id     UUID          REFERENCES bank_statements(id) ON DELETE CASCADE,
    bank_account_id       UUID          REFERENCES bank_accounts(id) ON DELETE SET NULL,
    booking_date          DATE          NOT NULL,
    valuta_date           DATE          NOT NULL,
    description           TEXT          NOT NULL DEFAULT '',
    amount                NUMERIC(15,2) NOT NULL,
    currency              TEXT          NOT NULL DEFAULT 'EUR',
    transaction_type      TEXT          NOT NULL,
    reference             TEXT          NOT NULL DEFAULT '',
    category_id           UUID          REFERENCES categories(id) ON DELETE SET NULL,
    content_hash          TEXT          NOT NULL,
    is_reconciled         BOOL          NOT NULL DEFAULT false,
    reconciliation_id     UUID          REFERENCES reconciliations(id) ON DELETE SET NULL,
    statement_type        TEXT,
    location              TEXT,
    reviewed              BOOLEAN       NOT NULL DEFAULT false,
    counterparty_name     TEXT,
    counterparty_iban     TEXT,
    bank_transaction_code TEXT,
    mandate_reference     TEXT,
    skip_forecasting      BOOLEAN       NOT NULL DEFAULT false,
    is_payslip_verified   BOOLEAN       NOT NULL DEFAULT false,
    CONSTRAINT transactions_user_content_hash_unique UNIQUE (content_hash, user_id)
);

CREATE INDEX IF NOT EXISTS idx_transactions_statement_id      ON transactions(bank_statement_id);
CREATE INDEX IF NOT EXISTS idx_transactions_booking_date      ON transactions(booking_date);
CREATE INDEX IF NOT EXISTS idx_transactions_category_id       ON transactions(category_id);
CREATE INDEX IF NOT EXISTS idx_transactions_reconciliation_id ON transactions(reconciliation_id)
    WHERE reconciliation_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_bank_account_id   ON transactions(bank_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_description_trgm  ON transactions USING gin (description gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_transactions_counterparty_trgm ON transactions USING gin (counterparty_name gin_trgm_ops);

-- ============================================================
-- Invoices
-- ============================================================
CREATE TABLE IF NOT EXISTS invoices (
    id                    UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               UUID          REFERENCES users(id) ON DELETE CASCADE,
    vendor                TEXT          NOT NULL DEFAULT '',
    amount                NUMERIC(15,2) NOT NULL DEFAULT 0,
    currency              TEXT          NOT NULL DEFAULT 'EUR',
    invoice_date          DATE,
    description           TEXT          NOT NULL DEFAULT '',
    content_hash          TEXT,
    original_file_name    VARCHAR(255),
    original_file_content BYTEA,
    category_id           UUID          REFERENCES categories(id) ON DELETE SET NULL,
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CONSTRAINT invoices_user_content_hash_unique UNIQUE (content_hash, user_id)
);

-- ============================================================
-- Payslips
-- ============================================================
CREATE TABLE IF NOT EXISTS payslips (
    id                    UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               UUID          REFERENCES users(id) ON DELETE CASCADE,
    original_file_name    VARCHAR(255)  NOT NULL,
    original_file_content BYTEA,
    content_hash          VARCHAR(64)   NOT NULL,
    period_month_num      INT,
    period_year           INT           NOT NULL,
    employer_name         VARCHAR(100)  NOT NULL DEFAULT 'Unknown',
    tax_class             VARCHAR(10),
    tax_id                VARCHAR(50),
    gross_pay             NUMERIC(12,2) NOT NULL,
    net_pay               NUMERIC(12,2) NOT NULL,
    payout_amount         NUMERIC(12,2) NOT NULL,
    custom_deductions     NUMERIC(12,2) NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ   DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT payslips_user_content_hash_unique    UNIQUE (content_hash, user_id),
    CONSTRAINT payslips_user_period_employer_unique UNIQUE (user_id, period_month_num, period_year, employer_name)
);

-- ============================================================
-- Payslip Bonuses
-- ============================================================
CREATE TABLE IF NOT EXISTS payslip_bonuses (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    payslip_id  UUID          NOT NULL REFERENCES payslips(id) ON DELETE CASCADE,
    description VARCHAR(512)  NOT NULL,
    amount      NUMERIC(12,2) NOT NULL,
    created_at  TIMESTAMPTZ   DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payslip_bonuses_payslip_id ON payslip_bonuses(payslip_id);

-- ============================================================
-- Planned Transactions
-- ============================================================
CREATE TABLE IF NOT EXISTS planned_transactions (
    id                     UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount                 NUMERIC(15,2) NOT NULL,
    date                   DATE          NOT NULL,
    description            TEXT          NOT NULL DEFAULT '',
    category_id            UUID          REFERENCES categories(id) ON DELETE SET NULL,
    status                 TEXT          NOT NULL DEFAULT 'pending',
    matched_transaction_id UUID          REFERENCES transactions(id) ON DELETE SET NULL,
    created_at             TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_planned_transactions_user_id ON planned_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_planned_transactions_date    ON planned_transactions(date);
CREATE INDEX IF NOT EXISTS idx_planned_transactions_status  ON planned_transactions(status);

-- ============================================================
-- Excluded Forecasts
-- ============================================================
CREATE TABLE IF NOT EXISTS excluded_forecasts (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    forecast_id UUID        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_excluded_forecasts_user_id_forecast_id ON excluded_forecasts(user_id, forecast_id);

-- ============================================================
-- Pattern Exclusions
-- ============================================================
CREATE TABLE IF NOT EXISTS pattern_exclusions (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    match_term TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pattern_exclusions_user_term ON pattern_exclusions(user_id, match_term);
