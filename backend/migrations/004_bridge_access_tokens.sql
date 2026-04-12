-- Bridge Access Tokens (BAT) for Standalone Mobile Sync (Hermit)
CREATE TABLE IF NOT EXISTS bridge_access_tokens (
    id           UUID PRIMARY KEY,
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT        NOT NULL, -- e.g., "My iPhone 15"
    token_hash   TEXT        NOT NULL,
    last_used_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Ensure fast lookup by user and token hash
CREATE INDEX IF NOT EXISTS idx_bridge_access_tokens_user_id ON bridge_access_tokens(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_bridge_access_tokens_token_hash ON bridge_access_tokens(token_hash);
