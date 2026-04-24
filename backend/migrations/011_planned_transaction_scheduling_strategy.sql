-- Migration: Add scheduling strategy to planned transactions
-- Purpose: Support "Last Bank Day" scheduling for salary and recurring payments.

-- 1. Add scheduling_strategy column with default 'fixed_day'
ALTER TABLE planned_transactions 
ADD COLUMN IF NOT EXISTS scheduling_strategy VARCHAR(50) NOT NULL DEFAULT 'fixed_day';

-- 2. Add comment for documentation
COMMENT ON COLUMN planned_transactions.scheduling_strategy IS 'Projection strategy: fixed_day (default) or last_bank_day';
