-- Add expiry_notified_at to bank_connections
ALTER TABLE bank_connections ADD COLUMN expiry_notified_at TIMESTAMP;
