-- =============================================================================
-- Squash Catch-up (002)
-- Brings any existing database up to the full squashed schema state.
--
-- Safe to apply at ANY migration level (001–017) because every statement
-- uses IF NOT EXISTS / ADD COLUMN IF NOT EXISTS guards.
--
-- Fresh installs: 001_initial_schema already creates everything — this is a
-- harmless no-op.
-- Existing installs: applies whatever is missing from migrations 013–017.
-- =============================================================================

-- -------------------------------------------------------
-- 013: is_variable_spending flag on categories
-- -------------------------------------------------------
ALTER TABLE categories ADD COLUMN IF NOT EXISTS is_variable_spending BOOLEAN NOT NULL DEFAULT false;

-- -------------------------------------------------------
-- 014: planned_transactions table
-- -------------------------------------------------------
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

-- -------------------------------------------------------
-- 015: forecast exclusions + skip_forecasting flag
-- -------------------------------------------------------
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS skip_forecasting BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS excluded_forecasts (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    forecast_id UUID        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_excluded_forecasts_user_id_forecast_id ON excluded_forecasts(user_id, forecast_id);

-- -------------------------------------------------------
-- 016: pattern exclusions
-- -------------------------------------------------------
CREATE TABLE IF NOT EXISTS pattern_exclusions (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    match_term TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pattern_exclusions_user_term ON pattern_exclusions(user_id, match_term);

-- -------------------------------------------------------
-- 017: is_payslip_verified flag on transactions
-- -------------------------------------------------------
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS is_payslip_verified BOOLEAN NOT NULL DEFAULT false;

