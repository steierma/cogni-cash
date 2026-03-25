# Database Schema

> **Host:** configured via `DATABASE_HOST` / `DATABASE_PORT` in `backend/.env`
> **Database:** configured via `POSTGRES_DB` in `backend/.env`
> **Engine:** PostgreSQL 16

---

## Table of Contents

1. [Entity-Relationship Overview](#entity-relationship-overview)
2. [Tables](#tables)
    - [schema_migrations](#schema_migrations)
    - [users](#users)
    - [categories](#categories)
    - [bank_statements](#bank_statements)
    - [transactions](#transactions)
    - [reconciliations](#reconciliations)
    - [invoices](#invoices)
    - [settings](#settings)
    - [payslips](#payslips)
    - [payslip_bonuses](#payslip_bonuses)
3. [Indexes & Constraints](#indexes--constraints)
4. [Triggers](#triggers)
5. [Deduplication via Content Hash](#deduplication-via-content-hash)
6. [Reconciliation](#reconciliation)
7. [Migration History](#migration-history)

---

## Entity-Relationship Overview

```text
┌─────────────────────────┐          ┌──────────────────────────────────┐
│       categories        │          │         bank_statements          │
├─────────────────────────┤          ├──────────────────────────────────┤
│ id (PK)                 │          │ id (PK)                          │
│ name                    │          │ account_holder                   │
│ color                   │          │ iban                             │
│ created_at              │          │ ...                              │
└────────────┬────────────┘          └─────────────────┬────────────────┘
             │ 1                                       │ 1
             │                                         │
             │ ∞                                       │ ∞
┌────────────▼─────────────────────────────────────────▼────────────┐
│                           transactions                            │
├───────────────────────────────────────────────────────────────────┤
│ id (PK)                                                           │
│ bank_statement_id (FK)                                            │
│ category_id (FK)                                                  │
│ booking_date                                                      │
│ amount                                                            │
│ content_hash (UNIQUE)                                             │
│ is_reconciled                                                     │
│ reconciliation_id (FK)                                            │
└────────────┬──────────────────────────────────────────────────────┘
             │ 1 (source tx) & 1 (target tx)
             │
             │ 1
┌────────────▼──────────────┐
│      reconciliations      │
├───────────────────────────┤
│ id (PK)                   │
│ settlement_transaction_hash (UNIQUE) │
│ target_transaction_hash (UNIQUE)     │
│ amount                    │
│ reconciled_at             │
└───────────────────────────┘
```

---

## Tables

### `schema_migrations`

Managed by `golang-migrate/migrate`. Tracks applied migrations.

### `users`

System accounts and RBAC roles.

| Column          | Type          | Notes                |
|:----------------|:--------------|:---------------------|
| `id`            | `UUID`        | PK                   |
| `username`      | `TEXT`        | `UNIQUE`             |
| `password_hash` | `TEXT`        | bcrypt hash          |
| `email`         | `TEXT`        | `UNIQUE`             |
| `full_name`     | `TEXT`        |                      |
| `address`       | `TEXT`        |                      |
| `role`          | `TEXT`        | `admin` or `manager` |
| `created_at`    | `TIMESTAMPTZ` |                      |

### `categories`

User-defined buckets for transactions and invoices.

| Column       | Type          | Notes                        |
|:-------------|:--------------|:-----------------------------|
| `id`         | `UUID`        | PK                           |
| `name`       | `TEXT`        | `UNIQUE` (e.g., 'Groceries') |
| `color`      | `TEXT`        | Hex code (e.g., '#ef4444')   |
| `created_at` | `TIMESTAMPTZ` |                              |

### `bank_statements`

Headers for imported files (PDF/CSV).

| Column           | Type            | Notes                                  |
|:-----------------|:----------------|:---------------------------------------|
| `id`             | `UUID`          | PK                                     |
| `account_holder` | `TEXT`          |                                        |
| `iban`           | `TEXT`          |                                        |
| `bic`            | `TEXT`          |                                        |
| `account_number` | `TEXT`          |                                        |
| `statement_date` | `DATE`          |                                        |
| `statement_no`   | `INT`           |                                        |
| `old_balance`    | `NUMERIC(15,2)` |                                        |
| `new_balance`    | `NUMERIC(15,2)` |                                        |
| `currency`       | `TEXT`          |                                        |
| `source_file`    | `TEXT`          | Filename                               |
| `original_file`  | `BYTEA`         | Binary blob of PDF/CSV                 |
| `imported_at`    | `TIMESTAMPTZ`   |                                        |
| `content_hash`   | `TEXT`          | `UNIQUE` (Deduplication)               |
| `statement_type` | `TEXT`          | `giro`, `credit_card`, `extra_account` |

### `transactions`

Line items belonging to a bank statement.

| Column                 | Type            | Notes                                 |
|:-----------------------|:----------------|:--------------------------------------|
| `id`                   | `UUID`          | PK                                    |
| `bank_statement_id`    | `UUID`          | FK → `bank_statements(id)` `CASCADE`  |
| `booking_date`         | `DATE`          |                                       |
| `valuta_date`          | `DATE`          |                                       |
| `description`          | `TEXT`          |                                       |
| `amount`               | `NUMERIC(15,2)` | Negative for expenses                 |
| `currency`             | `TEXT`          |                                       |
| `transaction_type`     | `TEXT`          | `credit` or `debit`                   |
| `reference`            | `TEXT`          |                                       |
| `category_id`          | `UUID`          | FK → `categories(id)` `SET NULL`      |
| `content_hash`         | `TEXT`          | `UNIQUE` (Deduplication)              |
| `is_reconciled`        | `BOOLEAN`       | Excludes from analytics if `true`     |
| `reconciliation_id`    | `UUID`          | FK → `reconciliations(id)` `SET NULL` |
| `exchange_rate`        | `NUMERIC(10,6)` |                                       |
| `amount_base_currency` | `NUMERIC(15,2)` |                                       |

### `reconciliations`

1:1 Mapping between internal transfers (e.g., paying off a credit card from a giro account).

| Column                        | Type            | Notes                          |
|:------------------------------|:----------------|:-------------------------------|
| `id`                          | `UUID`          | PK                             |
| `settlement_transaction_hash` | `TEXT`          | `UNIQUE` (Source Giro Debit)   |
| `target_transaction_hash`     | `TEXT`          | `UNIQUE` (Destination Credit)  |
| `amount`                      | `NUMERIC(15,2)` | Absolute value of the transfer |
| `reconciled_at`               | `TIMESTAMPTZ`   |                                |

### `invoices`

Standalone documents parsed by the LLM.

| Column                 | Type            | Notes                            |
|:-----------------------|:----------------|:---------------------------------|
| `id`                   | `UUID`          | PK                               |
| `raw_text`             | `TEXT`          | OCR / parsed text                |
| `vendor`               | `TEXT`          | Extracted by LLM                 |
| `amount`               | `NUMERIC(15,2)` |                                  |
| `currency`             | `TEXT`          |                                  |
| `invoice_date`         | `DATE`          |                                  |
| `created_at`           | `TIMESTAMPTZ`   |                                  |
| `category_id`          | `UUID`          | FK → `categories(id)` `SET NULL` |
| `exchange_rate`        | `NUMERIC(10,6)` |                                  |
| `amount_base_currency` | `NUMERIC(15,2)` |                                  |

### `settings`

Key-Value store for runtime configuration editable in the UI.

| Column  | Type   | Notes |
|:--------|:-------|:------|
| `key`   | `TEXT` | PK    |
| `value` | `TEXT` |       |

### `payslips`

Structured data from monthly payslip PDFs.

| Column                  | Type            | Notes                           |
|-------------------------|-----------------|---------------------------------|
| `id`                    | `UUID`          | PK                              |
| `source_file`           | `VARCHAR(255)`  | Path/URI where it was read from |
| `original_file_name`    | `VARCHAR(255)`  | Exact name of the uploaded PDF  |
| `original_file_mime`    | `VARCHAR(100)`  | `application/pdf`               |
| `original_file_size`    | `BIGINT`        | Bytes                           |
| `original_file_content` | `BYTEA`         | Full binary copy of the PDF     |
| `content_hash`          | `VARCHAR(64)`   | SHA256 (Deduplication)          |
| `period_month_num`      | `INT`           | 1-12                            |
| `period_year`           | `INT`           | e.g. 2024                       |
| `employee_name`         | `VARCHAR(100)`  |                                 |
| `tax_class`             | `VARCHAR(10)`   | e.g. "3" or "4"                 |
| `tax_id`                | `VARCHAR(50)`   |                                 |
| `gross_pay`             | `NUMERIC(12,2)` |                                 |
| `net_pay`               | `NUMERIC(12,2)` |                                 |
| `payout_amount`         | `NUMERIC(12,2)` |                                 |
| `custom_deductions`     | `NUMERIC(12,2)` | e.g. Car Leasing                |
| `created_at`            | `TIMESTAMPTZ`   |                                 |

### `payslip_bonuses`

Stores arbitrary bonus payments linked to a specific payslip.

| Column        | Type            | Notes                                   |
|---------------|-----------------|-----------------------------------------|
| `id`          | `UUID`          | PK                                      |
| `payslip_id`  | `UUID`          | FK → `payslips(id)` `ON DELETE CASCADE` |
| `description` | `VARCHAR(512)`  | e.g. 'Annual Bonus'                     |
| `amount`      | `NUMERIC(12,2)` |                                         |
| `created_at`  | `TIMESTAMPTZ`   |                                         |

---

## Indexes & Constraints

* **Cascading Deletes:** Deleting a `bank_statement` deletes all its `transactions`. Deleting a `payslip` deletes all
  its `payslip_bonuses`.
* **Nullification:** Deleting a `category` sets `category_id = NULL` on related transactions/invoices. Deleting a
  `reconciliation` sets `reconciliation_id = NULL` on linked transactions.
* **Unique Hashes:** `content_hash` is `UNIQUE` across `bank_statements`, `transactions`, and `payslips` to enforce
  strict deduplication at the DB level.
* **Payslip Uniqueness:** Enforced by `UNIQUE (period_month_num, period_year, employee_name)`.

---

## Deduplication via Content Hash

The application guarantees idempotency during file imports.

1. **Statement Level:** `content_hash` = `SHA256(IBAN + StatementDate + StatementNo + NewBalance)`
2. **Transaction Level:** `content_hash` = `SHA256(BookingDate + Amount + Description + Reference + CountIndex)` (
   CountIndex handles identical transactions on the same day).

If a user uploads the exact same PDF or CSV twice, the DB rejects the `content_hash`. If a statement overlaps (e.g., CSV
exports), existing transactions are skipped.

---

## Reconciliation

The system uses a **strict 1:1 transaction mapping** to link internal cash flows (e.g., paying a credit card bill from a
Giro account).

1. A Girokonto transaction (Amount `< 0`) is paired with a target statement transaction (Amount `> 0`).
2. Their absolute amounts must equal each other (sum to `0`).
3. Both transactions are updated in the DB with `is_reconciled = true` and share the same `reconciliation_id`.
4. The Analytics and Dashboard views exclude any transaction where `is_reconciled = true`, preventing internal money
   movements from artificially inflating Income or Expense KPIs.

---

## Migration History

| File                     | Description                                                                                                                                                                                                                                                                                 |
|--------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `001_initial_schema.sql` | Consolidated schema: all tables (`users`, `categories`, `bank_statements`, `reconciliations`, `transactions`, `invoices`, `settings`, `payslips`, `payslip_bonuses`) including the 1:1 Transaction Mapping (`target_transaction_hash`). Replaces all previous incremental migration chains. |


## sample_data.sql

```sql
DO $$
DECLARE
    cat_income UUID;
    cat_housing UUID;
    cat_misc UUID;
    cat_tech UUID;
    cat_groceries UUID;
    cat_utilities UUID;
    stmt_giro_id UUID;
    stmt_savings_id UUID;
    curr_date DATE := '2016-01-01';
    end_date DATE := '2026-03-01';
    
    -- Variables for randomization
    recon_amount NUMERIC(15,2);
    salary_gross NUMERIC(15,2);
    salary_net NUMERIC(15,2);
    invoice_amount NUMERIC(15,2);
    groceries_amount NUMERIC(15,2);
    utilities_amount NUMERIC(15,2);
BEGIN
    -- 1. Fetch existing core categories
    SELECT id INTO cat_income FROM categories WHERE name = 'Einkommen';
    SELECT id INTO cat_housing FROM categories WHERE name = 'Haus und Hausrat';
    SELECT id INTO cat_misc FROM categories WHERE name = 'Sonstige Ausgaben';

    -- 2. Create and fetch new custom categories for a realistic distribution
    INSERT INTO categories (name, color) VALUES ('Tech & Software', '#3b82f6') 
    ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name RETURNING id INTO cat_tech;
    
    INSERT INTO categories (name, color) VALUES ('Groceries & Food', '#10b981') 
    ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name RETURNING id INTO cat_groceries;

    INSERT INTO categories (name, color) VALUES ('Utilities & Internet', '#f97316') 
    ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name RETURNING id INTO cat_utilities;

    -- 3. Loop month by month for 10 years
    WHILE curr_date <= end_date LOOP
        -- Generate random amounts for this month
        recon_amount := round((random() * 400 + 100)::numeric, 2); 
        salary_gross := round((random() * 800 + 6000)::numeric, 2); 
        salary_net := salary_gross - round((random() * 200 + 2000)::numeric, 2); 
        invoice_amount := round((random() * 150 + 20)::numeric, 2); 
        groceries_amount := round((random() * 250 + 150)::numeric, 2); -- Between €150 and €400
        utilities_amount := round((random() * 50 + 80)::numeric, 2); -- Between €80 and €130

        -- Create a Giro Statement
        INSERT INTO bank_statements (id, account_holder, iban, statement_date, content_hash, statement_type)
        VALUES (gen_random_uuid(), 'Max Mustermann', 'DE12345678901234567890', curr_date + interval '28 days', md5(gen_random_uuid()::text), 'giro')
        RETURNING id INTO stmt_giro_id;

        -- Create a Savings/Credit Statement
        INSERT INTO bank_statements (id, account_holder, iban, statement_date, content_hash, statement_type)
        VALUES (gen_random_uuid(), 'Max Mustermann', 'DE09876543210987654321', curr_date + interval '28 days', md5(gen_random_uuid()::text), 'credit')
        RETURNING id INTO stmt_savings_id;

        -- ---------------------------------------------------------
        -- Insert Categorized Transactions
        -- ---------------------------------------------------------

        -- 1. Income (Einkommen)
        INSERT INTO transactions (bank_statement_id, booking_date, valuta_date, description, amount, transaction_type, category_id, content_hash, is_reconciled)
        VALUES (stmt_giro_id, curr_date + interval '1 day', curr_date + interval '1 day', 'Salary Mustermann GmbH', salary_net, 'credit', cat_income, md5(gen_random_uuid()::text), false);
        
        -- 2. Rent (Haus und Hausrat)
        INSERT INTO transactions (bank_statement_id, booking_date, valuta_date, description, amount, transaction_type, category_id, content_hash, is_reconciled)
        VALUES (stmt_giro_id, curr_date + interval '3 days', curr_date + interval '3 days', 'Rent Payment', -1200.00, 'debit', cat_housing, md5(gen_random_uuid()::text), false);

        -- 3. Utilities (Utilities & Internet)
        INSERT INTO transactions (bank_statement_id, booking_date, valuta_date, description, amount, transaction_type, category_id, content_hash, is_reconciled)
        VALUES (stmt_giro_id, curr_date + interval '4 days', curr_date + interval '4 days', 'Telekom Internet & Power', -utilities_amount, 'debit', cat_utilities, md5(gen_random_uuid()::text), false);

        -- 4. Groceries (Groceries & Food)
        INSERT INTO transactions (bank_statement_id, booking_date, valuta_date, description, amount, transaction_type, category_id, content_hash, is_reconciled)
        VALUES (stmt_giro_id, curr_date + interval '10 days', curr_date + interval '10 days', 'REWE Supermarket', -groceries_amount, 'debit', cat_groceries, md5(gen_random_uuid()::text), false);

        -- 5. Tech & Subscriptions (Tech & Software)
        INSERT INTO transactions (bank_statement_id, booking_date, valuta_date, description, amount, transaction_type, category_id, content_hash, is_reconciled)
        VALUES (stmt_giro_id, curr_date + interval '12 days', curr_date + interval '12 days', 'Hetzner Online GmbH', -invoice_amount, 'debit', cat_tech, md5(gen_random_uuid()::text), false);

        -- ---------------------------------------------------------
        -- OPEN RECONCILIATIONS (1:1 Transfers)
        -- ---------------------------------------------------------
        INSERT INTO transactions (bank_statement_id, booking_date, valuta_date, description, amount, transaction_type, category_id, content_hash, is_reconciled)
        VALUES (stmt_giro_id, curr_date + interval '15 days', curr_date + interval '15 days', 'Internal Transfer to Savings', -recon_amount, 'debit', cat_misc, md5(gen_random_uuid()::text), false);

        INSERT INTO transactions (bank_statement_id, booking_date, valuta_date, description, amount, transaction_type, category_id, content_hash, is_reconciled)
        VALUES (stmt_savings_id, curr_date + interval '16 days', curr_date + interval '16 days', 'Internal Transfer from Giro', recon_amount, 'credit', cat_misc, md5(gen_random_uuid()::text), false);

        -- ---------------------------------------------------------
        -- Standalone Invoices & Payslips
        -- ---------------------------------------------------------
        INSERT INTO invoices (raw_text, vendor, amount, invoice_date, category_id)
        VALUES ('Hetzner Online GmbH Cloud Server Instance...', 'Hetzner Online GmbH', invoice_amount, curr_date + interval '5 days', cat_tech);

        INSERT INTO payslips (original_file_name, original_file_size, content_hash, period_month_num, period_year, employee_name, tax_class, tax_id, gross_pay, net_pay, payout_amount)
        VALUES (
            'Entgeltnachweis_' || to_char(curr_date, 'YYYY_MM') || '.pdf', 
            45000, 
            md5(gen_random_uuid()::text), 
            EXTRACT(MONTH FROM curr_date), 
            EXTRACT(YEAR FROM curr_date), 
            'Max Mustermann', 
            '3', 
            '12345678901',
            salary_gross, 
            salary_net, 
            salary_net
        );

        -- Advance one month
        curr_date := curr_date + interval '1 month';
    END LOOP;
END $$;
```