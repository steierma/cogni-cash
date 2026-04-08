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
    - [bank_connections](#bank_connections)
    - [bank_accounts](#bank_accounts)
    - [password_reset_tokens](#password_reset_tokens)
3. [Indexes & Constraints](#indexes--constraints)
4. [Deduplication via Content Hash](#deduplication-via-content-hash)
5. [Reconciliation](#reconciliation)
6. [Migration History](#migration-history)

---

## Entity-Relationship Overview

```text
┌─────────────────────────┐          ┌──────────────────────────────────┐
│       categories        │          │         bank_statements          │
├─────────────────────────┤          ├──────────────────────────────────┤
│ id (PK)                 │          │ id (PK)                          │
│ user_id (FK)            │          │ user_id (FK)                     │
│ name                    │          │ account_holder                   │
│ color                   │          │ iban                             │
│ created_at              │          │ statement_date                   │
└────────────┬────────────┘          │ statement_no                     │
             │                       │ old_balance                      │
             │                       │ new_balance                      │
             │                       │ source_file                      │
             │                       │ content_hash (UQ)                │
             │                       │ statement_type                   │
             │                       │ bank_account_id (FK)             │
             │                       │ imported_at                      │
             │                       └─────────────────┬────────────────┘
             │                                         │ (1:N)
             │ (1:N)                 ┌─────────────────▼────────────────┐
             │                       │           transactions           │
             │                       ├──────────────────────────────────┤
             │                       │ id (PK)                          │
             │                       │ user_id (FK)                     │
             │                       │ bank_statement_id (FK)           │
             │                       │ bank_account_id (FK)             │
             │                       │ booking_date                     │
             │                       │ valuta_date                      │
             │                       │ description                      │
             │                       │ counterparty_name                │
             │                       │ counterparty_iban                │
             │                       │ bank_transaction_code            │
             │                       │ mandate_reference                │
             │                       │ location                         │
             │                       │ amount                           │
             │                       │ currency                         │
             │                       │ transaction_type                 │
             │                       │ reference                        │
             │                       │ category_id (FK)                 │
             │                       │ content_hash (UQ)                │
             │                       │ is_reconciled                    │
             │                       │ reconciliation_id (FK)           │
             │                       │ reviewed                         │
             │                       │ statement_type                   │
             │                       └─────────────────┬────────────────┘
             │                                         │
             │ (1:N)                 ┌─────────────────▼────────────────┐
             │                       │          reconciliations         │
             ├───────────────────────┤──────────────────────────────────┤
             │                       │ id (PK)                          │
             │                       │ user_id (FK)                     │
             │                       │ settlement_transaction_hash (FK)  │
             │                       │ target_transaction_hash (FK)      │
             │                       │ amount                           │
             │                       │ reconciled_at                    │
             │                       └──────────────────────────────────┘
             │
             │ (1:N)                 ┌──────────────────────────────────┐
             │                       │             invoices             │
             ├───────────────────────┤──────────────────────────────────┤
             │                       │ id (PK)                          │
             │                       │ user_id (FK)                     │
             │                       │ vendor                           │
             │                       │ category_id (FK)                 │
             │                       │ amount                           │
             │                       │ currency                         │
             │                       │ invoice_date                     │
             │                       │ description                      │
             │                       │ content_hash (UQ)                │
             │                       │ original_file_name               │
             │                       │ original_file_content            │
             │                       │ created_at                       │
             │                       └──────────────────────────────────┘
             │
             │ (1:N)                 ┌──────────────────────────────────┐
             │                       │             payslips             │
             └───────────────────────┤──────────────────────────────────┤
                                     │ id (PK)                          │
                                     │ user_id (FK)                     │
                                     │ original_file_name               │
                                     │ original_file_content            │
                                     │ content_hash (UQ)                │
                                     │ period_month_num                 │
                                     │ period_year                      │
                                     │ employer_name                    │
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
                                     └─────────────────┬────────────────┘

┌─────────────────────────┐          ┌─────────────────────────┐
│          users          │          │        settings         │
├─────────────────────────┤          ├─────────────────────────┤
│ id (PK)                 │          │ key (PK, part)          │
│ username (UQ)           │          │ user_id (PK, FK)        │
│ password_hash           │          │ value                   │
│ email (UQ)              │          └─────────────────────────┘
│ full_name               │
│ address                 │
│ role                    │
│ created_at              │
└────────────┬────────────┘
             │
             │ (1:N)                 ┌──────────────────────────────────┐
             │                       │      password_reset_tokens       │
             ├───────────────────────┤──────────────────────────────────┤
             │                       │ id (PK)                          │
             │                       │ user_id (FK)                     │
             │                       │ token_hash (UQ)                  │
             │                       │ expires_at                       │
             │                       │ created_at                       │
             │                       └──────────────────────────────────┘
             │
             │ (1:N)                 ┌──────────────────────────────────┐
             │                       │         bank_connections         │
             └───────────────────────┤──────────────────────────────────┘
                                     │ id (PK)                          │
                                     │ user_id (FK)                     │
                                     │ institution_id                   │
                                     │ institution_name                 │
                                     │ provider                         │
                                     │ requisition_id (UQ)              │
                                     │ reference_id (UQ)                │
                                     │ status                           │
                                     │ created_at                       │
                                     │ expires_at                       │
                                     └─────────────────┬────────────────┘
                                                       │ (1:N)
                                     ┌─────────────────▼────────────────┐
                                     │          bank_accounts           │
                                     ├──────────────────────────────────┤
                                     │ id (PK)                          │
                                     │ connection_id (FK)               │
                                     │ provider_account_id (UQ)         │
                                     │ iban                             │
                                     │ name                             │
                                     │ currency                         │
                                     │ balance                          │
                                     │ last_synced_at                   │
                                     │ account_type                     │
                                     │ last_sync_error                  │
                                     └──────────────────────────────────┘
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
Stores classification tags used across the application. Per-user namespace.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `user_id` | `UUID` | `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `name` | `TEXT` | `NOT NULL` | The category name (e.g., 'Groceries') |
| `color` | `TEXT` | `NOT NULL`, `DEFAULT '#6366f1'` | Hex color code for UI badges |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |

**Constraints:** `UNIQUE (name, user_id)`

### `bank_statements`
Represents imported bank statement files (PDF, CSV, XLS).

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `user_id` | `UUID` | `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `account_holder`| `TEXT` | `NOT NULL` | Name on the account |
| `iban` | `TEXT` | `NOT NULL` | Standardized account number |
| `statement_date`| `DATE` | `NOT NULL` | Closing date of the statement |
| `statement_no` | `INT` | `NOT NULL` | Sequential statement number |
| `old_balance` | `NUMERIC(15,2)`| `NOT NULL` | Opening balance for the period |
| `new_balance` | `NUMERIC(15,2)`| `NOT NULL` | Closing balance for the period |
| `content_hash` | `VARCHAR(64)` | `NOT NULL`, `UNIQUE` | SHA-256 hash for deduplication |
| `statement_type`| `VARCHAR(20)` | `NOT NULL`, `DEFAULT 'giro'` | 'giro', 'credit_card', or 'extra_account' |
| `bank_account_id`| `UUID` | `FK (bank_accounts.id)`, `ON DELETE SET NULL` | Parent bank account (API connection) |
| `imported_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |

### `transactions`
Contains both file-imported and API-synced financial transactions.

| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `user_id` | `UUID` | `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `bank_statement_id`| `UUID` | `FK (bank_statements.id)`, `ON DELETE CASCADE` | Parent statement (nullable for API syncs) |
| `bank_account_id`| `UUID` | `FK (bank_accounts.id)`, `ON DELETE SET NULL` | Parent bank account (for API syncs) |
| `booking_date` | `DATE` | `NOT NULL` | Date transaction was booked |
| `valuta_date` | `DATE` | `NOT NULL` | Value date |
| `description` | `TEXT` | `NOT NULL` | Original reference/purpose string |
| `location` | `TEXT` | | Optional extracted city, region, or country |
| `amount` | `NUMERIC(15,2)`| `NOT NULL` | Positive (income) or negative (expense) |
| `currency` | `TEXT` | `NOT NULL`, `DEFAULT 'EUR'` | Currency code |
| `transaction_type`| `TEXT` | `NOT NULL` | 'credit' (in) or 'debit' (out) |
| `reference` | `TEXT` | | Optional reference ID from bank |
| `category_id` | `UUID` | `FK (categories.id)`, `ON DELETE SET NULL` | Assigned category |
| `content_hash` | `VARCHAR(64)` | `NOT NULL`, `UNIQUE` | SHA-256 hash for deduplication |
| `is_reconciled` | `BOOLEAN` | `NOT NULL`, `DEFAULT false` | True if part of an internal transfer |
| `reconciliation_id` | `UUID` | `FK (reconciliations.id)`, `ON DELETE SET NULL` | Linked reconciliation record |
| `reviewed` | `BOOLEAN` | `NOT NULL`, `DEFAULT false` | True if user has acknowledged the transaction |
| `statement_type`| `TEXT` | | 'giro', 'credit_card', or 'extra_account' |

### `reconciliations`
Links a settlement payment (from a Giro account) to a target transaction (Credit Card or Extra Account) via content hashes.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `user_id` | `UUID` | `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `settlement_transaction_hash` | `TEXT` | `NOT NULL`, `UNIQUE` | The hash of the debit transaction |
| `target_transaction_hash` | `TEXT` | `UNIQUE` | The hash of the credit transaction |
| `amount` | `NUMERIC(15,2)`| `NOT NULL` | The absolute transferred amount |
| `reconciled_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |

### `invoices`
Stores parsed and auto-categorized invoices.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `user_id` | `UUID` | `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `vendor` | `TEXT` | `NOT NULL` | Vendor or company name extracted by LLM |
| `category_id` | `UUID` | `FK (categories.id)`, `ON DELETE SET NULL` | Assigned category |
| `amount` | `NUMERIC(12,2)`| `NOT NULL` | Total invoice amount |
| `currency` | `TEXT` | `NOT NULL`, `DEFAULT 'EUR'` | Currency code |
| `invoice_date` | `DATE` | | Extracted invoice date |
| `description` | `TEXT` | `NOT NULL`, `DEFAULT ''` | Manual or extracted description |
| `content_hash` | `TEXT` | `UNIQUE` | SHA-256 hash for deduplication |
| `original_file_name` | `VARCHAR(255)` | | Filename of the imported invoice |
| `original_file_content`| `BYTEA` | | Binary content of the invoice file |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |

### `settings`
Key-value store for application configuration. Per-user settings.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `key` | `TEXT` | `PK` (part) | The setting identifier string |
| `user_id` | `UUID` | `PK`, `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `value` | `TEXT" | `NOT NULL` | The configured value |

**Constraints:** `PRIMARY KEY (key, user_id)`

### `payslips`
Structured payroll information parsed from HR documents.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `user_id` | `UUID` | `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `original_file_name` | `VARCHAR(255)`| `NOT NULL` | Stored filename |
| `original_file_content`| `BYTEA` | | Raw binary payload of the PDF/document |
| `content_hash` | `VARCHAR(64)` | `NOT NULL`, `UNIQUE` | SHA-256 hash for deduplication |
| `period_month_num` | `INT` | | 1-12 representing the payroll month |
| `period_year` | `INT` | `NOT NULL` | Payroll year |
| `employer_name` | `VARCHAR(100)`| `NOT NULL`, `DEFAULT 'Unknown'` | Extracted employer name |
| `tax_class` | `VARCHAR(10)` | | E.g., '1', '3', '4' |
| `tax_id` | `VARCHAR(50)` | | Tax identification number |
| `gross_pay` | `NUMERIC(12,2)`| `NOT NULL` | Total gross salary |
| `net_pay` | `NUMERIC(12,2)`| `NOT NULL` | Net salary before deductions |
| `payout_amount` | `NUMERIC(12,2)`| `NOT NULL` | Final transferred amount to bank account |
| `custom_deductions`| `NUMERIC(12,2)`| `NOT NULL`, `DEFAULT 0` | E.g., Leasing rates |
| `created_at` | `TIMESTAMPTZ` | `DEFAULT CURRENT_TIMESTAMP` | Record creation timestamp |

**Constraints:** `UNIQUE (user_id, period_month_num, period_year, employer_name)`

### `payslip_bonuses`
Variable compensation components extracted from payslips.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `payslip_id` | `UUID` | `FK (payslips.id)`, `ON DELETE CASCADE` | The parent payslip |
| `description` | `VARCHAR(512)`| `NOT NULL` | Reason (e.g., 'Urlaubsgeld', 'Bonus') |
| `amount` | `NUMERIC(12,2)`| `NOT NULL` | Gross amount of the bonus |
| `created_at` | `TIMESTAMPTZ` | `DEFAULT CURRENT_TIMESTAMP` | Record creation timestamp |

### `bank_connections`
Stores external bank API authorization state.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `user_id` | `UUID` | `NOT NULL`, `FK (users.id)` | Owning user |
| `institution_id` | `TEXT` | `NOT NULL` | Provider's bank ID |
| `institution_name` | `TEXT` | `NOT NULL`, `DEFAULT ''` | Readable bank name |
| `provider` | `TEXT` | `NOT NULL`, `DEFAULT 'enablebanking'` | Aggregator name |
| `requisition_id` | `TEXT` | `NOT NULL`, `UNIQUE` | Session ID from provider |
| `reference_id` | `TEXT` | `NOT NULL`, `UNIQUE` | Internal tracking ID |
| `status` | `TEXT` | `NOT NULL`, `DEFAULT 'initialized'` | initialized, linked, expired, failed |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Creation timestamp |
| `expires_at` | `TIMESTAMPTZ` | | Consent expiration |

### `bank_accounts`
Stores individual accounts under a connection.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `connection_id` | `UUID` | `NOT NULL`, `FK (bank_connections.id)` | Parent connection |
| `provider_account_id` | `TEXT` | `NOT NULL`, `UNIQUE` | Account ID from provider |
| `iban` | `TEXT` | `NOT NULL`, `DEFAULT ''` | International Account Number |
| `name` | `TEXT` | `NOT NULL`, `DEFAULT ''` | Friendly name |
| `currency` | `TEXT` | `NOT NULL`, `DEFAULT 'EUR'` | Currency code |
| `balance` | `NUMERIC(15,2)`| `NOT NULL`, `DEFAULT 0` | Cached balance |
| `last_synced_at` | `TIMESTAMPTZ` | | Last successful data fetch |
| `account_type` | `TEXT` | `NOT NULL`, `DEFAULT 'giro'` | 'giro', 'credit_card', or 'extra_account' |
| `last_sync_error` | `TEXT` | | Error message from last failed sync |

### `password_reset_tokens`
Stores temporary, hashed security tokens for the "Forgot Password" flow.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `user_id` | `UUID` | `NOT NULL`, `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `token_hash` | `TEXT` | `NOT NULL`, `UNIQUE` | SHA-256 hash of the random token |
| `expires_at` | `TIMESTAMPTZ` | `NOT NULL` | Token expiration timestamp |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |

---

## Indexes & Constraints

To ensure data integrity and optimal query performance:

- **Primary Keys:** Every table uses a `UUID` generated via `gen_random_uuid()` (except `settings` which uses a composite key of `key` and `user_id`).
- **Foreign Keys:**
    - `transactions.bank_statement_id` -> `bank_statements(id)` (`ON DELETE CASCADE`)
    - `transactions.bank_account_id` -> `bank_accounts(id)` (`ON DELETE SET NULL`)
    - `bank_statements.bank_account_id` -> `bank_accounts(id)` (`ON DELETE SET NULL`)
    - `bank_connections.user_id` -> `users(id)` (`ON DELETE CASCADE`)
    - `bank_accounts.connection_id` -> `bank_connections(id)` (`ON DELETE CASCADE`)
    - `payslip_bonuses.payslip_id` -> `payslips(id)` (`ON DELETE CASCADE`)
    - `transactions.category_id` -> `categories(id)` (`ON DELETE SET NULL`)
    - `invoices.category_id` -> `categories(id)` (`ON DELETE SET NULL`)
    - `transactions.reconciliation_id` -> `reconciliations(id)` (`ON DELETE SET NULL`)
    - All tenant-specific tables point to `users(id)` (`ON DELETE CASCADE`).
- **Uniqueness:**
    - `users.username` and `users.email`
    - `categories(name, user_id)` composite uniqueness
    - `content_hash` on `bank_statements`, `transactions`, `invoices`, and `payslips`
    - `reconciliations.settlement_transaction_hash` and `reconciliations.target_transaction_hash`
    - `payslips(user_id, period_month_num, period_year, employer_name)` composite unique constraint.
- **Performance Indexes:**
    - `idx_transactions_date_amt` on `transactions(booking_date, amount)`
    - `idx_payslips_period` on `payslips(period_year, period_month_num)`
    - `idx_transactions_bank_account_id` on `transactions(bank_account_id)`
    - `idx_bank_statements_bank_account_id` on `bank_statements(bank_account_id)`
    - `idx_reset_tokens_hash` on `password_reset_tokens(token_hash)`
    - `idx_transactions_description_trgm` on `transactions(description)` (GIN trigram)
    - `idx_transactions_counterparty_trgm` on `transactions(counterparty_name)` (GIN trigram)

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

1. **Pairing**: A `reconciliations` record links a debit from a `giro` account (e.g., settling a credit card bill) to a credit on a `credit_card` account using their unique `content_hash`.
2. **Transaction State**: Both transactions have their `is_reconciled` flag set to `TRUE` and their `reviewed` flag set to `TRUE`. The transactions are linked to the reconciliation record via `reconciliation_id`.
3. **Statement State**: When a `credit_card` or `extra_account` statement has all of its incoming settlement payments linked, its `is_reconciled` flag can be set to `TRUE`.

---

## Migration History

Migrations are plain SQL files in `backend/migrations/`, applied in lexicographic order.

1.  **`001_initial_schema.sql`**: Baseline schema containing all initial tables, constraints, and indices.
2.  **`002_add_invoice_content_hash.sql`**: Adds file storage fields, `description`, and `content_hash` to the `invoices` table.
3.  **`003_add_bank_integration.sql`**: Adds `bank_connections` and `bank_accounts` tables. Modifies `transactions` and `bank_statements` to support linking to live accounts.
4.  **`004_add_bank_provider.sql`**: Adds `provider` column to `bank_connections` for multi-aggregator support.
5.  **`005_add_bank_account_type.sql`**: Adds `account_type` to `bank_accounts` table and `statement_type` to `transactions`.
6.  **`006_add_transaction_location.sql`**: Adds `location` column to `transactions` table.
7.  **`007_add_transaction_reviewed.sql`**: Adds `reviewed` boolean column to `transactions` table for user acknowledgement.
8.  **`008_add_password_reset_tokens.sql`**: Adds `password_reset_tokens` table for the secure forgot-password flow.
9.  **`009_add_bank_account_sync_error.sql`**: Adds `last_sync_error` to `bank_accounts` table for transparent error reporting.
10. **`010_add_user_tenancy.sql`**: Implements multi-tenancy by adding `user_id` to all relevant tables and updating constraints.
11. **`011_enrich_transactions.sql`**: Adds `counterparty_name`, `counterparty_iban`, `bank_transaction_code` and mandate reference to transactions.
12. **`012_add_fuzzy_matching.sql`**: Enables `pg_trgm` extension and adds GIN indexes for high-performance fuzzy matching of descriptions and counterparty names. Also removes redundant columns across `bank_statements`, `transactions`, `invoices`, and `payslips` to streamline the schema.
