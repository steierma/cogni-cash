-- Migration 007: Add reviewed field to transactions
-- This field tracks if a user has acknowledged a new transaction.
-- Default is true for existing transactions to avoid a backlog.
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS reviewed BOOLEAN NOT NULL DEFAULT true;

-- For future inserts, we want it to be true by default if not specified,
-- but since our business logic will handle syncing, we'll set the column default to true
-- AFTER seeding existing ones as reviewed.
ALTER TABLE transactions ALTER COLUMN reviewed SET DEFAULT false;
