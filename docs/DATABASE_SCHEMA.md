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
│ created_at              │          │ bic                              │
└────────────┬────────────┘          │ account_number                   │
             │                       │ statement_date                   │
             │                       │ statement_no                     │
             │                       │ old_balance                      │
             │                       │ new_balance                      │
             │                       │ source_file                      │
             │                       │ content_hash (UQ)                │
             │                       │ statement_type                   │
             │                       │ is_reconciled                    │
             │                       │ created_at                       │
             │                       └─────────────────┬────────────────┘
             │                                         │ (1:N)
             │ (1:N)                 ┌─────────────────▼────────────────┐
             │                       │           transactions           │
             │                       ├──────────────────────────────────┤
             │                       │ id (PK)                          │
             │                       │ bank_statement_id (FK)           │
             │                       │ booking_date                     │
             │                       │ valuta_date                      │
             │                       │ description                      │
             │                       │ amount                           │
             │                       │ currency                         │
             │                       │ transaction_type                 │
             │                       │ reference                        │
             │                       │ category_id (FK)                 │
             │                       │ content_hash (UQ)                │
             │                       │ is_reconciled                    │
             │                       │ target_statement_id (FK)         │
             │                       │ created_at                       │
             │                       └─────────────────┬────────────────┘
             │                                         │
             │ (1:N)                 ┌─────────────────▼────────────────┐
             │                       │          reconciliations         │
             ├───────────────────────┤──────────────────────────────────┤
             │                       │ id (PK)                          │
             │                       │ settlement_tx_id (FK)            │
             │                       │ target_tx_id (FK)                │
             │                       │ amount                           │
             │                       │ created_at                       │
             │                       └──────────────────────────────────┘
             │
             │ (1:N)                 ┌──────────────────────────────────┐
             │                       │             invoices             │
             ├───────────────────────┤──────────────────────────────────┤
             │                       │ id (PK)                          │
             │                       │ vendor                           │
             │                       │ category_id (FK)                 │
             │                       │ amount                           │
             │                       │ currency                         │
             │                       │ invoice_date                     │
             │                       │ description                      │
             │                       │ raw_text                         │
             │                       │ content_hash (UQ)                │
             │                       │ original_file_name               │
             │                       │ original_file_mime               │
             │                       │ original_file_size               │
             │                       │ original_file_content            │
             │                       │ created_at                       │
             │                       └──────────────────────────────────┘
             │
             │ (1:N)                 ┌──────────────────────────────────┐
             │                       │             payslips             │
             └───────────────────────┤──────────────────────────────────┤
                                     │ id (PK)                          │
                                     │ source_file                      │
                                     │ original_file_name               │
                                     │ original_file_mime               │
                                     │ original_file_size               │
                                     │ original_file_content            │
                                     │ content_hash (UQ)                │
                                     │ period_month_num                 │
                                     │ period_year                      │
                                     │ employee_name                    │
                                     │ tax_class                        │
                                     │ tax_id                           │
                                     │ gross_pay                        │
                                     │ net_pay                          │
                                     │ payout_amount                    │
                                     │ custom_deductions                │
                                     │ created_at                       │
                                     └─────────────────┬────────────────┘
                                                       │ (1:N)
                                     ┌─────────────────▼────────────────┐
                                     │         payslip_bonuses          │
                                     ├──────────────────────────────────┤
                                     │ id (PK)                          │
                                     │ payslip_id (FK)                  │
                                     │ description                      │
                                     │ amount                           │
                                     │ created_at                       │
                                     └──────────────────────────────────┘

┌─────────────────────────┐          ┌─────────────────────────┐
│          users          │          │        settings         │
├─────────────────────────┤          ├─────────────────────────┤
│ id (PK)                 │          │ key (PK)                │
│ username (UQ)           │          │ value                   │
│ password_hash           │          │ updated_at              │
│ email (UQ)              │          └─────────────────────────┘
│ full_name               │
│ address                 │
│ role                    │
│ created_at              │
└─────────────────────────┘
```

---

## Tables

### `schema_migrations`
Internal table managed by the lightweight Go migration script (`cmd/migrate/main.go`). Tracks applied SQL files.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `version` | `VARCHAR(255)` | `PK` | The filename (e.g. `001_initial_schema.sql`) |
| `applied_at` | `TIMESTAMPTZ` | `DEFAULT CURRENT_TIMESTAMP` | When the migration was applied |

### `users`
Manages system access, authentication (JWT), and role-based permissions (RBAC).

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `username` | `TEXT` | `NOT NULL`, `UNIQUE` | Login username |
| `password_hash` | `TEXT` | `NOT NULL` | bcrypt hashed password |
| `email` | `TEXT` | `NOT NULL`, `UNIQUE` | User's email address |
| `full_name` | `TEXT` | `NOT NULL`, `DEFAULT ''` | Display name |
| `address` | `TEXT` | `NOT NULL`, `DEFAULT ''` | Physical address / contact info |
| `role` | `TEXT` | `NOT NULL`, `DEFAULT 'manager'` | System role (e.g., 'admin', 'manager') |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |

### `categories`
Stores classification tags used across the application.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `name` | `TEXT` | `NOT NULL`, `UNIQUE` | The category name (e.g., 'Groceries') |
| `color` | `TEXT` | `NOT NULL`, `DEFAULT '#6366f1'` | Hex color code for UI badges |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |

### `bank_statements`
Represents imported bank statement files (PDF, CSV, XLS).

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `account_holder`| `TEXT` | `NOT NULL` | Name on the account |
| `iban` | `TEXT` | `NOT NULL` | Standardized account number |
| `bic` | `TEXT` | `NOT NULL` | Bank identifier code |
| `account_number`| `TEXT` | `NOT NULL` | Internal account number string |
| `statement_date`| `DATE` | `NOT NULL` | Closing date of the statement |
| `statement_no` | `INT` | `NOT NULL` | Sequential statement number |
| `old_balance` | `NUMERIC(15,2)`| `NOT NULL` | Opening balance for the period |
| `new_balance` | `NUMERIC(15,2)`| `NOT NULL` | Closing balance for the period |
| `source_file` | `TEXT` | `NOT NULL` | Original filename |
| `content_hash` | `VARCHAR(64)` | `NOT NULL`, `UNIQUE` | SHA-256 hash for deduplication |
| `statement_type`| `VARCHAR(20)` | `NOT NULL`, `DEFAULT 'giro'` | 'giro', 'credit_card', or 'extra_account' |
| `is_reconciled` | `BOOLEAN` | `NOT NULL`, `DEFAULT false` | True if all balance transfers are settled |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |

### `transactions`
Individual line items belonging to a bank statement.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `bank_statement_id`| `UUID` | `FK (bank_statements.id)`, `ON DELETE CASCADE` | Parent statement |
| `booking_date` | `DATE` | `NOT NULL` | Date transaction was booked |
| `valuta_date` | `DATE` | `NOT NULL` | Value date |
| `description` | `TEXT` | `NOT NULL` | Original reference/purpose string |
| `amount` | `NUMERIC(15,2)`| `NOT NULL` | Positive (income) or negative (expense) |
| `currency` | `TEXT` | `NOT NULL`, `DEFAULT 'EUR'` | Currency code |
| `transaction_type`| `TEXT` | `NOT NULL` | 'credit' (in) or 'debit' (out) |
| `reference` | `TEXT` | | Optional reference ID from bank |
| `category_id` | `UUID` | `FK (categories.id)`, `ON DELETE SET NULL` | Assigned category |
| `content_hash` | `VARCHAR(64)` | `NOT NULL`, `UNIQUE` | SHA-256 hash for deduplication |
| `is_reconciled` | `BOOLEAN` | `NOT NULL`, `DEFAULT false` | True if part of an internal transfer |
| `target_statement_id`| `UUID`| `FK (bank_statements.id)`, `ON DELETE SET NULL` | Target statement if reconciled |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |

### `reconciliations`
Links a settlement payment (from a Giro account) to a target transaction (Credit Card or Extra Account).

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `settlement_tx_id` | `UUID`| `FK (transactions.id)`, `ON DELETE CASCADE` | The debit (outgoing) transaction |
| `target_tx_id` | `UUID` | `FK (transactions.id)`, `ON DELETE CASCADE` | The credit (incoming) transaction |
| `amount` | `NUMERIC(12,2)`| `NOT NULL` | The absolute transferred amount |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |

### `invoices`
Stores parsed and auto-categorized invoices.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `vendor` | `TEXT` | `NOT NULL` | Vendor or company name extracted by LLM |
| `category_id` | `UUID` | `FK (categories.id)`, `ON DELETE SET NULL` | Assigned category |
| `amount` | `NUMERIC(12,2)`| `NOT NULL` | Total invoice amount |
| `currency` | `TEXT` | `NOT NULL`, `DEFAULT 'EUR'` | Currency code |
| `invoice_date` | `DATE` | | Extracted invoice date |
| `description` | `TEXT` | `NOT NULL`, `DEFAULT ''` | Manual or extracted description |
| `raw_text` | `TEXT` | `NOT NULL` | Raw text payload sent to LLM |
| `content_hash` | `TEXT` | `UNIQUE` | SHA-256 hash for deduplication |
| `original_file_name` | `VARCHAR(255)` | | Filename of the imported invoice |
| `original_file_mime` | `VARCHAR(100)` | | MIME type (e.g., application/pdf) |
| `original_file_size` | `BIGINT` | | Size of the file in bytes |
| `original_file_content`| `BYTEA` | | Binary content of the invoice file |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |

### `settings`
Key-value store for application configuration (e.g., LLM prompts, import intervals).

| Column | Type | Constraints | Description |
|---|---|---|---|
| `key` | `TEXT` | `PK` | The setting identifier string |
| `value` | `TEXT` | `NOT NULL` | The configured value |
| `updated_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Timestamp of last change |

### `payslips`
Structured payroll information parsed from HR documents.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `source_file` | `VARCHAR(255)`| | Name or path of the original file |
| `original_file_name` | `VARCHAR(255)`| `NOT NULL` | Stored filename |
| `original_file_mime` | `VARCHAR(100)`| | MIME type of the stored document |
| `original_file_size` | `BIGINT` | `NOT NULL` | Size of the stored document in bytes |
| `original_file_content`| `BYTEA` | | Raw binary payload of the PDF/document |
| `content_hash` | `VARCHAR(64)` | `NOT NULL`, `UNIQUE` | SHA-256 hash for deduplication |
| `period_month_num` | `INT` | | 1-12 representing the payroll month |
| `period_year` | `INT` | `NOT NULL` | Payroll year |
| `employee_name` | `VARCHAR(100)`| `NOT NULL` | Extracted employee name |
| `tax_class` | `VARCHAR(10)` | | E.g., '1', '3', '4' |
| `tax_id` | `VARCHAR(50)` | | Tax identification number |
| `gross_pay` | `NUMERIC(12,2)`| `NOT NULL` | Total gross salary |
| `net_pay` | `NUMERIC(12,2)`| `NOT NULL` | Net salary before deductions |
| `payout_amount` | `NUMERIC(12,2)`| `NOT NULL` | Final transferred amount to bank account |
| `custom_deductions`| `NUMERIC(12,2)`| `NOT NULL`, `DEFAULT 0` | E.g., Leasing rates |
| `created_at` | `TIMESTAMPTZ` | `DEFAULT CURRENT_TIMESTAMP` | Record creation timestamp |

### `payslip_bonuses`
Variable compensation components extracted from payslips.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `payslip_id` | `UUID` | `FK (payslips.id)`, `ON DELETE CASCADE` | The parent payslip |
| `description` | `VARCHAR(512)`| `NOT NULL` | Reason (e.g., 'Urlaubsgeld', 'Bonus') |
| `amount` | `NUMERIC(12,2)`| `NOT NULL` | Gross amount of the bonus |
| `created_at` | `TIMESTAMPTZ` | `DEFAULT CURRENT_TIMESTAMP` | Record creation timestamp |

---

## Indexes & Constraints

To ensure data integrity and optimal query performance:

- **Primary Keys:** Every table uses a `UUID` generated via `gen_random_uuid()` (except `settings` which uses a TEXT key).
- **Foreign Keys:**
    - `transactions.bank_statement_id` -> `bank_statements(id)` (`ON DELETE CASCADE`)
    - `payslip_bonuses.payslip_id` -> `payslips(id)` (`ON DELETE CASCADE`)
    - `transactions.category_id` -> `categories(id)` (`ON DELETE SET NULL`)
    - `invoices.category_id` -> `categories(id)` (`ON DELETE SET NULL`)
    - `reconciliations.settlement_tx_id` -> `transactions(id)` (`ON DELETE CASCADE`)
    - `reconciliations.target_tx_id` -> `transactions(id)` (`ON DELETE CASCADE`)
- **Uniqueness:**
    - `users.username` and `users.email`
    - `categories.name`
    - `content_hash` on `bank_statements`, `transactions`, `invoices`, and `payslips`
    - `payslips(period_month_num, period_year, employee_name)` composite unique constraint.
- **Performance Indexes:**
    - `idx_transactions_date_amt` on `transactions(booking_date, amount)`
    - `idx_payslips_period` on `payslips(period_year, period_month_num)`

---

## Triggers

1. **`trg_update_settings_updated_at`**: Automatically fires before any `UPDATE` on the `settings` table to set the `updated_at` column to `NOW()`.

---

## Deduplication via Content Hash

To allow safe, idempotent imports, the system calculates a deterministic SHA-256 hash for records prior to insertion.

- **`bank_statements`**: Hashed over `iban` + `statement_date` + `statement_no` + `new_balance`.
- **`transactions`**: Hashed over `iban` + `booking_date` + `valuta_date` + `amount` + `description`.
- **`invoices`**: Hashed over the original file content or explicitly set upon insert.
- **`payslips`**: Hashed over the original file content. When importing via JSON, duplicate checks rely on `original_file_name`.

If a `UNIQUE` constraint violation occurs on `content_hash`, the import routine marks the item as "skipped/duplicate" without failing the entire batch.

---

## Reconciliation

Reconciliation prevents internal transfers from inflating analytics metrics.

1. **Pairing**: A `reconciliations` record links a debit from a `giro` account (e.g., settling a credit card bill) to a credit on a `credit_card` account.
2. **Transaction State**: Both transactions have their `is_reconciled` flag set to `TRUE`. The debit transaction is also updated with `target_statement_id`.
3. **Statement State**: When a `credit_card` or `extra_account` statement has all of its incoming settlement payments linked, its `is_reconciled` flag can be set to `TRUE`, effectively marking it "done" and hiding it from the pending reconciliation wizard.

---

## Migration History

Since moving to a strictly consolidated initialization approach, older fragmented migration scripts were merged. Current actively tracked migrations:

1. **`001_initial_schema.sql`**: The baseline schema containing all tables, constraints, triggers, and indices.
2. **`002_add_invoice_content_hash.sql`**: Adds file storage fields, `description`, and `content_hash` to the `invoices` table to support PDF/image uploading and deduplication.