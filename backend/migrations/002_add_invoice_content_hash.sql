-- Deduplication hash
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS content_hash TEXT UNIQUE;

-- Original file storage
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS original_file_name VARCHAR(255);
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS original_file_mime VARCHAR(100);
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS original_file_size BIGINT;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS original_file_content BYTEA;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
