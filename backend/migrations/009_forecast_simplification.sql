-- 009_forecast_simplification.sql
-- Story 6: Drop obsolete forecasting exclusion tables and columns.
-- The forecast engine now uses subscriptions as the source of truth.

-- Drop exclusion tables (pattern exclusions + individual forecast exclusions)
DROP TABLE IF EXISTS pattern_exclusions;
DROP TABLE IF EXISTS excluded_forecasts;

-- Drop the skip_forecasting column from transactions
ALTER TABLE transactions DROP COLUMN IF EXISTS skip_forecasting;

-- Drop the is_superseded column from planned_transactions
ALTER TABLE planned_transactions DROP COLUMN IF EXISTS is_superseded;

-- Drop the index on is_superseded if it exists
DROP INDEX IF EXISTS idx_planned_transactions_is_superseded;

