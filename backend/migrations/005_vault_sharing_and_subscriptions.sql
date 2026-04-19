-- =============================================================================
-- V2 Feature Squash (005)
-- Consolidates all features added after the v2.0.0 public release (001-004).
-- Includes: Shared Categories/Invoices, Document Vault, Hardened Settings,
-- and Subscription Management.
-- =============================================================================

-- 1. Infrastructure & User Table Adjustments
ALTER TABLE users ADD COLUMN IF NOT EXISTS address TEXT NOT NULL DEFAULT '';

-- 2. Sharing (Categories & Invoices)
CREATE TABLE IF NOT EXISTS shared_categories (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id         UUID        NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    owner_user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shared_with_user_id UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_level    TEXT        NOT NULL DEFAULT 'view' CHECK (permission_level IN ('view', 'edit')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT shared_categories_unique UNIQUE (owner_user_id, category_id, shared_with_user_id),
    CONSTRAINT shared_categories_no_self_share CHECK (owner_user_id <> shared_with_user_id)
);

CREATE INDEX IF NOT EXISTS idx_shared_categories_owner_user_id       ON shared_categories(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_shared_categories_shared_with_user_id ON shared_categories(shared_with_user_id);
CREATE INDEX IF NOT EXISTS idx_shared_categories_category_id          ON shared_categories(category_id);

CREATE TABLE IF NOT EXISTS shared_invoices (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id          UUID        NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    owner_user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shared_with_user_id UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_level    TEXT        NOT NULL DEFAULT 'view' CHECK (permission_level IN ('view', 'edit')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT shared_invoices_unique UNIQUE (owner_user_id, invoice_id, shared_with_user_id),
    CONSTRAINT shared_invoices_no_self_share CHECK (owner_user_id <> shared_with_user_id)
);

CREATE INDEX IF NOT EXISTS idx_shared_invoices_owner_user_id       ON shared_invoices(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_shared_invoices_shared_with_user_id ON shared_invoices(shared_with_user_id);
CREATE INDEX IF NOT EXISTS idx_shared_invoices_invoice_id          ON shared_invoices(invoice_id);

-- 3. Category & Forecasting Updates
ALTER TABLE categories ADD COLUMN IF NOT EXISTS forecast_strategy VARCHAR(50) NOT NULL DEFAULT '3y';
ALTER TABLE planned_transactions ADD COLUMN IF NOT EXISTS interval_months INTEGER NOT NULL DEFAULT 0;
ALTER TABLE planned_transactions ADD COLUMN IF NOT EXISTS end_date DATE;
ALTER TABLE planned_transactions ADD COLUMN IF NOT EXISTS is_superseded BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX IF NOT EXISTS idx_planned_transactions_is_superseded ON planned_transactions(is_superseded);

-- 4. Document Vault
CREATE TABLE IF NOT EXISTS documents (
    id                    UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    document_type         TEXT          NOT NULL,
    original_file_name    TEXT          NOT NULL,
    original_file_content BYTEA,
    content_hash          TEXT          NOT NULL,
    mime_type             TEXT,
    extracted_text        TEXT,
    metadata              JSONB         NOT NULL DEFAULT '{}',
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CONSTRAINT documents_user_hash_unique UNIQUE (user_id, content_hash),
    CONSTRAINT documents_type_check CHECK (document_type IN ('tax_certificate', 'receipt', 'contract', 'other'))
);

CREATE INDEX IF NOT EXISTS idx_documents_user_id ON documents(user_id);
CREATE INDEX IF NOT EXISTS idx_documents_type    ON documents(document_type);
CREATE INDEX IF NOT EXISTS idx_documents_created ON documents(created_at);
CREATE INDEX IF NOT EXISTS idx_documents_extracted_text_trgm ON documents USING gin (extracted_text gin_trgm_ops);

-- 5. Settings Hardening (Sensitive Flags & BYTEA Storage)
ALTER TABLE settings ADD COLUMN IF NOT EXISTS is_sensitive BOOLEAN DEFAULT FALSE;
ALTER TABLE settings ALTER COLUMN value TYPE BYTEA USING value::bytea;

UPDATE settings SET is_sensitive = TRUE 
WHERE key ILIKE '%password%' 
   OR key ILIKE '%token%' 
   OR key ILIKE '%secret%' 
   OR key ILIKE '%key%';

-- 6. Bank Account Uniqueness (Tenancy-aware)
ALTER TABLE bank_accounts DROP CONSTRAINT IF EXISTS bank_accounts_provider_account_id_key;
ALTER TABLE bank_accounts DROP CONSTRAINT IF EXISTS bank_accounts_connection_id_provider_account_id_key;
ALTER TABLE bank_accounts ADD CONSTRAINT bank_accounts_connection_id_provider_account_id_key UNIQUE (connection_id, provider_account_id);

-- 7. Subscriptions Management
CREATE TABLE IF NOT EXISTS subscriptions (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    merchant_name      TEXT          NOT NULL,
    amount             NUMERIC(15,2) NOT NULL,
    currency           TEXT          NOT NULL DEFAULT 'EUR',
    billing_cycle      TEXT          NOT NULL DEFAULT 'monthly',
    billing_interval   INT           NOT NULL DEFAULT 1,
    category_id        UUID          REFERENCES categories(id) ON DELETE SET NULL,
    customer_number    TEXT,
    contact_email      TEXT,
    contact_phone      TEXT,
    contact_website    TEXT,
    support_url        TEXT,
    cancellation_url   TEXT,
    status             TEXT          NOT NULL DEFAULT 'active',
    notice_period_days INT,
    contract_end_date  DATE,
    is_trial           BOOLEAN       NOT NULL DEFAULT false,
    payment_method     TEXT,
    last_occurrence    DATE,
    next_occurrence    DATE,
    notes             TEXT,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status  ON subscriptions(status);

ALTER TABLE transactions ADD COLUMN IF NOT EXISTS subscription_id UUID REFERENCES subscriptions(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_subscription_id ON transactions(subscription_id) WHERE subscription_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS subscription_events (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID          NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    user_id         UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type      TEXT          NOT NULL,
    title           TEXT          NOT NULL,
    content         TEXT          NOT NULL,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscription_events_sub_id ON subscription_events(subscription_id);
CREATE INDEX IF NOT EXISTS idx_subscription_events_user_id ON subscription_events(user_id);

-- 8. Unified Discovery Feedback
DROP TABLE IF EXISTS allowed_subscription_suggestions;
DROP TABLE IF EXISTS declined_subscription_suggestions;

CREATE TABLE IF NOT EXISTS subscription_discovery_feedback (
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    merchant_name TEXT NOT NULL,
    status        VARCHAR(20) NOT NULL, -- 'ALLOWED', 'DECLINED', 'AI_REJECTED'
    source        VARCHAR(20) DEFAULT 'USER' NOT NULL,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, merchant_name)
);

CREATE INDEX IF NOT EXISTS idx_sub_feedback_lookup ON subscription_discovery_feedback (user_id, merchant_name);
