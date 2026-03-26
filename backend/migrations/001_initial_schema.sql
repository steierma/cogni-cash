-- =============================================================================
-- Clean Consolidated Schema
-- Defines the final database state without legacy migration/backfill guards.
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

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

-- Ensure the bootstrap admin user gets the admin role
UPDATE users SET role = 'admin' WHERE username = 'admin';

-- ============================================================
-- Categories
-- ============================================================
CREATE TABLE IF NOT EXISTS categories (
                                          id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL UNIQUE,
    color      TEXT        NOT NULL DEFAULT '#6366f1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

INSERT INTO categories (name, color) VALUES
                                         ('Haus und Hausrat',                         '#f59e0b'),
                                         ('Bildung, Gesundheit, Beauty und Wellness', '#ec4899'),
                                         ('Sonstige Ausgaben',                        '#6b7280'),
                                         ('Einkommen',                                '#22c55e')
    ON CONFLICT (name) DO NOTHING;

-- ============================================================
-- Bank Statements
-- ============================================================
CREATE TABLE IF NOT EXISTS bank_statements (
                                               id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    account_holder TEXT          NOT NULL DEFAULT '',
    iban           TEXT          NOT NULL DEFAULT '',
    bic            TEXT          NOT NULL DEFAULT '',
    account_number TEXT          NOT NULL DEFAULT '',
    statement_date DATE,
    statement_no   INT           NOT NULL DEFAULT 0,
    old_balance    NUMERIC(15,2) NOT NULL DEFAULT 0,
    new_balance    NUMERIC(15,2) NOT NULL DEFAULT 0,
    currency       TEXT          NOT NULL DEFAULT 'EUR',
    source_file    TEXT          NOT NULL DEFAULT '',
    imported_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    content_hash   TEXT          NOT NULL UNIQUE,
    original_file  BYTEA,
    statement_type TEXT          NOT NULL DEFAULT 'giro'
    );

-- ============================================================
-- Reconciliations
-- ============================================================
CREATE TABLE IF NOT EXISTS reconciliations (
                                               id                          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    settlement_transaction_hash TEXT          NOT NULL UNIQUE,
    target_transaction_hash     TEXT          UNIQUE,
    amount                      NUMERIC(15,2) NOT NULL,
    reconciled_at               TIMESTAMPTZ   NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_reconciliations_target_transaction
    ON reconciliations(target_transaction_hash);

-- ============================================================
-- Transactions
-- ============================================================
CREATE TABLE IF NOT EXISTS transactions (
                                            id                   UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_statement_id    UUID          NOT NULL REFERENCES bank_statements(id) ON DELETE CASCADE,
    booking_date         DATE          NOT NULL,
    valuta_date          DATE          NOT NULL,
    description          TEXT          NOT NULL DEFAULT '',
    amount               NUMERIC(15,2) NOT NULL,
    currency             TEXT          NOT NULL DEFAULT 'EUR',
    transaction_type     TEXT          NOT NULL,
    reference            TEXT          NOT NULL DEFAULT '',
    category_id          UUID          REFERENCES categories(id) ON DELETE SET NULL,
    content_hash         TEXT          NOT NULL UNIQUE,
    is_reconciled        BOOL          NOT NULL DEFAULT false,
    reconciliation_id    UUID          REFERENCES reconciliations(id) ON DELETE SET NULL,
    exchange_rate        NUMERIC(10,6) DEFAULT 1.0,
    amount_base_currency NUMERIC(15,2)
    );

CREATE INDEX IF NOT EXISTS idx_transactions_statement_id      ON transactions(bank_statement_id);
CREATE INDEX IF NOT EXISTS idx_transactions_booking_date      ON transactions(booking_date);
CREATE INDEX IF NOT EXISTS idx_transactions_category_id       ON transactions(category_id);
CREATE INDEX IF NOT EXISTS idx_transactions_reconciliation_id ON transactions(reconciliation_id)
    WHERE reconciliation_id IS NOT NULL;

-- ============================================================
-- Invoices
-- ============================================================
CREATE TABLE IF NOT EXISTS invoices (
                                        id                   UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    raw_text             TEXT          NOT NULL DEFAULT '',
    vendor               TEXT          NOT NULL DEFAULT '',
    amount               NUMERIC(15,2) NOT NULL DEFAULT 0,
    currency             TEXT          NOT NULL DEFAULT 'EUR',
    invoice_date         DATE,
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    category_id          UUID          REFERENCES categories(id) ON DELETE SET NULL,
    exchange_rate        NUMERIC(10,6) DEFAULT 1.0,
    amount_base_currency NUMERIC(15,2)
    );

-- ============================================================
-- Settings
-- ============================================================
CREATE TABLE IF NOT EXISTS settings (
                                        key   TEXT PRIMARY KEY,
                                        value TEXT NOT NULL
);

INSERT INTO settings (key, value)
VALUES ('base_currency', 'EUR')
    ON CONFLICT (key) DO NOTHING;

-- ============================================================
-- Payslips
-- ============================================================
CREATE TABLE IF NOT EXISTS payslips (
                                        id                    UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    source_file           VARCHAR(255),
    original_file_name    VARCHAR(255) NOT NULL,
    original_file_mime    VARCHAR(100),
    original_file_size    BIGINT       NOT NULL,
    original_file_content BYTEA,
    content_hash          VARCHAR(64)  NOT NULL UNIQUE,
    period_month_num      INT,
    period_year           INT          NOT NULL,
    employee_name         VARCHAR(100) NOT NULL,
    tax_class             VARCHAR(10),
    tax_id                VARCHAR(50),
    gross_pay             NUMERIC(12,2) NOT NULL,
    net_pay               NUMERIC(12,2) NOT NULL,
    payout_amount         NUMERIC(12,2) NOT NULL,
    custom_deductions     NUMERIC(12,2) NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (period_month_num, period_year, employee_name)
    );

-- ============================================================
-- Payslip Bonuses
-- ============================================================
CREATE TABLE IF NOT EXISTS payslip_bonuses (
                                               id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    payslip_id  UUID         NOT NULL REFERENCES payslips(id) ON DELETE CASCADE,
    description VARCHAR(512) NOT NULL,
    amount      NUMERIC(12,2) NOT NULL,
    created_at  TIMESTAMPTZ  DEFAULT CURRENT_TIMESTAMP
    );

CREATE INDEX IF NOT EXISTS idx_payslip_bonuses_payslip_id
    ON payslip_bonuses(payslip_id);