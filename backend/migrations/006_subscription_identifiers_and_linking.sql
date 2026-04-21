-- =============================================================================
-- Subscription Identifiers and Manual Linking (006)
-- Adds support for tracking matching/ignored hashes, SEPA Mandate References,
-- and Counterparty IBANs to enable deterministic transaction matching.
-- =============================================================================

-- 1. Manual Linking support
-- Tracks hashes that should ALWAYS be linked to this subscription.
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS matching_hashes TEXT[] DEFAULT '{}';

-- Tracks hashes that should NEVER be linked to this subscription (manual unlink override).
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS ignored_hashes TEXT[] DEFAULT '{}';

-- 2. Deterministic Identifiers
-- Tracks SEPA Mandate References that are bound to this subscription.
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS linked_mandates TEXT[] DEFAULT '{}';

-- Tracks Counterparty IBANs that are bound to this subscription.
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS linked_ibans TEXT[] DEFAULT '{}';
