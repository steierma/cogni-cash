-- 1. Create shared_bank_accounts table
CREATE TABLE IF NOT EXISTS shared_bank_accounts (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_account_id     UUID        NOT NULL REFERENCES bank_accounts(id) ON DELETE CASCADE,
    owner_user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shared_with_user_id UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_level    TEXT        NOT NULL DEFAULT 'view' CHECK (permission_level IN ('view', 'edit')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT shared_bank_accounts_unique UNIQUE (owner_user_id, bank_account_id, shared_with_user_id),
    CONSTRAINT shared_bank_accounts_no_self_share CHECK (owner_user_id <> shared_with_user_id)
);

CREATE INDEX IF NOT EXISTS idx_shared_bank_accounts_owner_user_id       ON shared_bank_accounts(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_shared_bank_accounts_shared_with_user_id ON shared_bank_accounts(shared_with_user_id);
CREATE INDEX IF NOT EXISTS idx_shared_bank_accounts_bank_account_id    ON shared_bank_accounts(bank_account_id);

-- 2. Support Virtual Bank Accounts (nullable connection_id)
ALTER TABLE bank_accounts ALTER COLUMN connection_id DROP NOT NULL;
ALTER TABLE bank_accounts ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;

-- Backfill user_id from bank_connections for existing accounts
UPDATE bank_accounts
SET user_id = bc.user_id
FROM bank_connections bc
WHERE bank_accounts.connection_id = bc.id AND bank_accounts.user_id IS NULL;

-- 3. Link Subscriptions and Planned Transactions to Bank Accounts
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS bank_account_id UUID REFERENCES bank_accounts(id) ON DELETE SET NULL;
ALTER TABLE planned_transactions ADD COLUMN IF NOT EXISTS bank_account_id UUID REFERENCES bank_accounts(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_subscriptions_bank_account_id ON subscriptions(bank_account_id);
CREATE INDEX IF NOT EXISTS idx_planned_transactions_bank_account_id ON planned_transactions(bank_account_id);
