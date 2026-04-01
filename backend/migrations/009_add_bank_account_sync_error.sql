-- Add last_sync_error to bank_accounts
ALTER TABLE bank_accounts ADD COLUMN last_sync_error TEXT;
