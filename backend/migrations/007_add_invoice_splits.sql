-- Migration: 007_add_invoice_splits.sql
-- Description: Adds support for multi-category line items for invoices with strict tenancy.

CREATE TABLE IF NOT EXISTS invoice_line_items (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    amount DECIMAL(19,4) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Idempotent fix for dirty local databases that applied the first version of this migration
ALTER TABLE invoice_line_items ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE;

-- We can safely set NOT NULL because this table is guaranteed to be empty on machines where the first version was applied
ALTER TABLE invoice_line_items ALTER COLUMN user_id SET NOT NULL;

-- Performance and Multi-Tenancy
CREATE INDEX IF NOT EXISTS idx_invoice_line_items_invoice_id ON invoice_line_items(invoice_id);
CREATE INDEX IF NOT EXISTS idx_invoice_line_items_user_id ON invoice_line_items(user_id);
