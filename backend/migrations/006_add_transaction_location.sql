-- ============================================================
-- Add location to transactions
-- ============================================================

ALTER TABLE transactions
    ADD COLUMN location TEXT;