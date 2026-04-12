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
    - [refresh_tokens](#refresh_tokens)
    - [planned_transactions](#planned_transactions)
    - [excluded_forecasts](#excluded_forecasts)
    - [pattern_exclusions](#pattern_exclusions)
    - [bridge_access_tokens](#bridge_access_tokens)
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
             │                       │ currency                         │
             │                       │ source_file                      │
             │                       │ original_file                    │
             │                       │ content_hash (UQ w/ user_id)     │
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
             │                       │ content_hash (UQ w/ user_id)     │
             │                       │ is_reconciled                    │
             │                       │ reconciliation_id (FK)           │
             │                       │ reviewed                         │
             │                       │ statement_type                   │
             │                       │ skip_forecasting                 │
             │                       │ is_payslip_verified              │
             │                       └─────────────────┬────────────────┘
             │                                         │
             │ (1:N)                 ┌─────────────────▼────────────────┐
             │                       │          reconciliations         │
             ├───────────────────────┤──────────────────────────────────┤
             │                       │ id (PK)                          │
             │                       │ user_id (FK)                     │
             │                       │ settlement_transaction_hash (FK) │
             │                       │ target_transaction_hash (FK)     │
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
             │                       │ content_hash (UQ w/ user_id)     │
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
                                     │ content_hash (UQ w/ user_id)     │
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
             │                       │      bridge_access_tokens        │
             ├───────────────────────┤──────────────────────────────────┤
             │                       │ id (PK)                          │
             │                       │ user_id (FK)                     │
             │                       │ name                             │
             │                       │ token_hash (UQ)                  │
             │                       │ last_used_at                     │
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
Internal table managed by the lightweight Go migration script (`cmd/migrate/main.go`). Tracks applied SQL files by filename stem and content hash.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `version` | `TEXT` | `PK` | Filename stem (e.g. `001_initial_schema`) |
| `applied_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT NOW()` | When the migration was last applied |
| `content_hash` | `TEXT` | `NOT NULL DEFAULT ''` | SHA-256 of the SQL file — triggers re-run if changed |

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
| `color` | `TEXT` | `NOT NULL`, `DEFAULT '#6366f1'` | Hex color code for UI badges. Constraint: `~* '^#[a-fA-F0-9]{6}$'` |
| `is_variable_spending` | `BOOLEAN` | `NOT NULL`, `DEFAULT false` | If true, forecast uses monthly burn rate instead of discrete events |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |
| `deleted_at` | `TIMESTAMPTZ` | `NULL` | Soft-delete timestamp |

**Constraints:** `UNIQUE (name, user_id)`

### `bank_statements`
Represents imported bank statement files (PDF, CSV, XLS).

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `user_id` | `UUID` | `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `bank_account_id`| `UUID` | `FK (bank_accounts.id)`, `ON DELETE SET NULL` | Parent bank account (API connection) |
| `account_holder`| `TEXT` | `NOT NULL` | Name on the account |
| `iban` | `TEXT` | `NOT NULL` | Standardized account number |
| `statement_date`| `DATE` | | Closing date of the statement |
| `statement_no` | `INT` | `NOT NULL` | Sequential statement number |
| `old_balance` | `NUMERIC(15,2)`| `NOT NULL` | Opening balance for the period |
| `new_balance` | `NUMERIC(15,2)`| `NOT NULL` | Closing balance for the period |
| `currency` | `VARCHAR(3)` | `NOT NULL`, `DEFAULT 'EUR'` | 3-letter currency code (ISO 4217) |
| `source_file` | `TEXT` | `NOT NULL` | Original filename |
| `imported_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |
| `content_hash` | `TEXT` | `NOT NULL`, `UNIQUE w/ user_id` | SHA-256 hash for deduplication |
| `original_file`| `BYTEA` | | Raw binary payload of the imported file |
| `statement_type`| `TEXT` | `NOT NULL`, `DEFAULT 'giro'` | 'giro', 'credit_card', or 'extra_account' |

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
| `amount` | `NUMERIC(15,2)`| `NOT NULL` | Positive (income) or negative (expense) |
| `currency` | `VARCHAR(3)` | `NOT NULL`, `DEFAULT 'EUR'` | 3-letter currency code (ISO 4217) |
| `transaction_type`| `TEXT` | `NOT NULL` | 'credit' (in) or 'debit' (out) |
| `reference` | `TEXT` | `NOT NULL`, `DEFAULT ''` | Optional reference ID from bank |
| `category_id` | `UUID` | `FK (categories.id)`, `ON DELETE SET NULL` | Assigned category |
| `content_hash` | `TEXT` | `NOT NULL`, `UNIQUE w/ user_id` | SHA-256 hash for deduplication |
| `is_reconciled` | `BOOLEAN` | `NOT NULL`, `DEFAULT false` | True if part of an internal transfer |
| `reconciliation_id` | `UUID` | `FK (reconciliations.id)`, `ON DELETE SET NULL` | Linked reconciliation record |
| `statement_type`| `TEXT` | | 'giro', 'credit_card', or 'extra_account' |
| `location` | `TEXT` | | Optional extracted city, region, or country |
| `reviewed` | `BOOLEAN` | `NOT NULL`, `DEFAULT false` | True if user has acknowledged the transaction |
| `counterparty_name`| `TEXT` | | Extracted metadata |
| `counterparty_iban`| `TEXT` | | Extracted metadata |
| `bank_transaction_code`| `TEXT` | | Extracted metadata |
| `mandate_reference`| `TEXT` | | Extracted metadata |
| `skip_forecasting`| `BOOLEAN` | `NOT NULL`, `DEFAULT false` | If true, ignored by the forecast pattern detector |
| `is_payslip_verified`| `BOOLEAN` | `NOT NULL`, `DEFAULT false` | True if verified against a payslip |

### `reconciliations`
Links a settlement payment (from a Giro account) to a target transaction (Credit Card or Extra Account) via content hashes.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `user_id` | `UUID` | `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `settlement_transaction_hash` | `TEXT` | `NOT NULL`, `UNIQUE w/ user_id` | The hash of the debit transaction |
| `target_transaction_hash` | `TEXT` | `UNIQUE w/ user_id` | The hash of the credit transaction |
| `amount` | `NUMERIC(15,2)`| `NOT NULL` | The absolute transferred amount |
| `reconciled_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |

### `invoices`
Stores parsed and auto-categorized invoices.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `user_id` | `UUID` | `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `vendor` | `TEXT` | `NOT NULL` | Vendor or company name extracted by LLM |
| `amount` | `NUMERIC(15,2)`| `NOT NULL`, `DEFAULT 0` | Total invoice amount |
| `currency` | `VARCHAR(3)` | `NOT NULL`, `DEFAULT 'EUR'` | 3-letter currency code (ISO 4217) |
| `invoice_date` | `DATE` | | Extracted invoice date |
| `description` | `TEXT` | `NOT NULL`, `DEFAULT ''` | Manual or extracted description |
| `content_hash` | `TEXT` | `UNIQUE w/ user_id` | SHA-256 hash for deduplication |
| `original_file_name` | `VARCHAR(255)` | | Filename of the imported invoice |
| `original_file_content`| `BYTEA` | | Binary content of the invoice file |
| `category_id` | `UUID` | `FK (categories.id)`, `ON DELETE SET NULL` | Assigned category |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |

### `settings`
Key-value store for application configuration. Per-user settings.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `key` | `TEXT` | `PK` (part) | The setting identifier string |
| `user_id` | `UUID` | `PK`, `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `value` | `TEXT` | `NOT NULL` | The configured value |

**Constraints:** `PRIMARY KEY (key, user_id)`

### `payslips`
Structured payroll information parsed from HR documents.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `user_id` | `UUID` | `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `original_file_name` | `VARCHAR(255)`| `NOT NULL` | Stored filename |
| `original_file_content`| `BYTEA` | | Raw binary payload of the PDF/document |
| `content_hash` | `VARCHAR(64)` | `NOT NULL`, `UNIQUE w/ user_id` | SHA-256 hash for deduplication |
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
| `requisition_id` | `TEXT` | `NOT NULL`, `UNIQUE` | Session ID from provider |
| `reference_id` | `TEXT` | `NOT NULL`, `UNIQUE` | Internal tracking ID |
| `status` | `TEXT` | `NOT NULL`, `DEFAULT 'initialized'` | initialized, linked, expired, failed |
| `provider` | `TEXT` | `NOT NULL`, `DEFAULT 'enablebanking'` | Aggregator name |
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
| `currency` | `VARCHAR(3)` | `NOT NULL`, `DEFAULT 'EUR'` | 3-letter currency code (ISO 4217) |
| `balance` | `NUMERIC(15,2)`| `NOT NULL`, `DEFAULT 0` | Cached balance |
| `account_type` | `TEXT` | `NOT NULL`, `DEFAULT 'giro'` | 'giro', 'credit_card', or 'extra_account' |
| `last_synced_at` | `TIMESTAMPTZ` | | Last successful data fetch |
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

### `refresh_tokens`
Stores rotation tokens for JWT session renewal and revocation.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK` | Unique identifier |
| `user_id` | `UUID` | `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `token_hash` | `TEXT` | `NOT NULL` | Hashed refresh token |
| `expires_at` | `TIMESTAMPTZ` | `NOT NULL` | Expiration date |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |
| `revoked` | `BOOLEAN` | `NOT NULL`, `DEFAULT FALSE` | Revocation status |

### `planned_transactions`
Contains manual entries for expected future transactions to improve forecasting accuracy.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `user_id` | `UUID` | `NOT NULL`, `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `amount` | `NUMERIC(15,2)`| `NOT NULL` | Expected amount |
| `date` | `DATE` | `NOT NULL` | Expected occurrence date |
| `description` | `TEXT` | `NOT NULL`, `DEFAULT ''` | Reference or purpose |
| `category_id` | `UUID` | `FK (categories.id)`, `ON DELETE SET NULL` | Assigned category |
| `status` | `TEXT` | `NOT NULL`, `DEFAULT 'pending'` | 'pending', 'matched', or 'cancelled' |
| `matched_transaction_id` | `UUID` | `FK (transactions.id)`, `ON DELETE SET NULL` | The actual transaction this was resolved to |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | Record creation timestamp |

### `excluded_forecasts`
Stores UUIDs of future projections that the user has explicitly chosen to ignore/delete from their view.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `user_id` | `UUID` | `NOT NULL`, `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `forecast_id` | `UUID` | `NOT NULL` | The deterministic UUID of the excluded projection |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | When the exclusion was created |

### `pattern_exclusions`
Stores rules for ignoring entire recurring patterns based on their normalized description or counterparty.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK`, `DEFAULT gen_random_uuid()` | Unique identifier |
| `user_id` | `UUID` | `NOT NULL`, `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `match_term` | `TEXT` | `NOT NULL` | The normalized term (e.g. first 25 chars of description) to exclude |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL`, `DEFAULT NOW()` | When the exclusion was created |

**Constraints:** `UNIQUE (user_id, match_term)`

### `bridge_access_tokens`
Stores Bridge Access Tokens (BAT) for standalone mobile sync (Hermit). These are long-lived tokens for devices that do not use standard JWT-based authentication.

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | `UUID` | `PK` | Unique identifier |
| `user_id` | `UUID` | `NOT NULL`, `FK (users.id)`, `ON DELETE CASCADE` | Owning user |
| `name` | `TEXT` | `NOT NULL` | Device name (e.g., "iPhone 15") |
| `token_hash` | `TEXT` | `NOT NULL`, `UNIQUE` | SHA-256 hash of the token |
| `last_used_at`| `TIMESTAMPTZ`| | Last time the token was used for sync |
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
    - `content_hash` scoped to `user_id` on `bank_statements`, `transactions`, `invoices`, and `payslips`
    - `reconciliations.settlement_transaction_hash` and `reconciliations.target_transaction_hash` scoped to `user_id`
    - `payslips(user_id, period_month_num, period_year, employer_name)` composite unique constraint.
- **Performance Indexes:**
    - `idx_transactions_statement_id` on `transactions(bank_statement_id)`
    - `idx_transactions_booking_date` on `transactions(booking_date)`
    - `idx_transactions_category_id` on `transactions(category_id)`
    - `idx_transactions_reconciliation_id` on `transactions(reconciliation_id)`
    - `idx_reconciliations_target_transaction` on `reconciliations(target_transaction_hash)`
    - `idx_transactions_bank_account_id` on `transactions(bank_account_id)`
    - `idx_bank_statements_bank_account_id` on `bank_statements(bank_account_id)`
    - `idx_reset_tokens_hash` on `password_reset_tokens(token_hash)`
    - `idx_reset_tokens_expires` on `password_reset_tokens(expires_at)`
    - `idx_refresh_tokens_user_id` on `refresh_tokens(user_id)`
    - `idx_refresh_tokens_token_hash` on `refresh_tokens(token_hash)`
    - `idx_payslip_bonuses_payslip_id` on `payslip_bonuses(payslip_id)`
    - `idx_transactions_description_trgm` on `transactions(description)` (GIN trigram)
    - `idx_transactions_counterparty_trgm` on `transactions(counterparty_name)` (GIN trigram)
    - `idx_planned_transactions_user_id` on `planned_transactions(user_id)`
    - `idx_categories_user_id` on `categories(user_id)`
    - `idx_settings_user_id` on `settings(user_id)`
    - `idx_bank_connections_user_id` on `bank_connections(user_id)`
    - `idx_bank_statements_user_id` on `bank_statements(user_id)`
    - `idx_reconciliations_user_id` on `reconciliations(user_id)`
    - `idx_transactions_user_id` on `transactions(user_id)`
    - `idx_invoices_user_id` on `invoices(user_id)`
    - `idx_payslips_user_id` on `payslips(user_id)`
    - `idx_planned_transactions_date` on `planned_transactions(date)`
    - `idx_planned_transactions_status` on `planned_transactions(status)`
    - `idx_excluded_forecasts_user_id_forecast_id` on `excluded_forecasts(user_id, forecast_id)`
    - `idx_pattern_exclusions_user_term` on `pattern_exclusions(user_id, match_term)`
    - `idx_bridge_access_tokens_user_id` on `bridge_access_tokens(user_id)`
    - `idx_bridge_access_tokens_token_hash` on `bridge_access_tokens(token_hash)` (UNIQUE)

### Custom Constraints
- **Categories**: `check_color_hex` (`color ~* '^#[a-fA-F0-9]{6}$'`)
- **Currency**: `check_currency_len` (`length(currency) = 3`) on `bank_accounts`, `bank_statements`, `transactions`, and `invoices`.

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

Migrations are plain SQL files in `backend/migrations/`, applied in **lexicographic order** by the Go migration runner (`cmd/migrate`). The runner tracks each file by filename stem and content hash in a `schema_migrations` table — only changed or new files are (re-)applied.

The schema history was **squashed** at v2.0.0 (consolidating migrations 001–017). There are now two files:

| File | Purpose |
|---|---|
| `001_initial_schema.sql` | **Full schema for fresh installs.** Creates all extensions, tables, indexes, constraints, and default seed data in one idempotent script. Use this as the authoritative reference for the complete database structure. |
| `002_squash_catchup.sql` | **Catch-up for existing databases.** Safely applies everything from the original migrations 013–017 (variable spending flag, planned transactions, forecast exclusions, pattern exclusions, `is_payslip_verified`). Uses `IF NOT EXISTS` / `ADD COLUMN IF NOT EXISTS` guards — a no-op on fresh installs. |
| `003_database_improvements.sql` | **Post-squash optimizations.** Adds missing multi-tenancy indexes, color/currency validation constraints, soft deletes for categories, and `refresh_tokens` for JWT revocation. |
| `004_bridge_access_tokens.sql` | **Hermit Sync Bridge support.** Adds the `bridge_access_tokens` table for standalone mobile synchronization without JWT. |

### Upgrading an Existing Installation to v2.0.0+

If your database was created before the squash (i.e., it has entries like `002_add_invoice_content_hash` in `schema_migrations`), run the following **once** before deploying:

```bash
# 1. Back up first (recommended)
make db-dump-local

# 2. Rewrite the migration history and apply the catch-up
make db-squash-upgrade

# 3. Deploy / start normally
make up
```

`make db-squash-upgrade` removes all old numbered migration entries from `schema_migrations` and runs `make db-migrate`, which stamps the real content hashes for both squashed files and applies any missing columns or tables via `002_squash_catchup.sql`.

### Adding Future Migrations

Add new files starting at `003_...sql`. The two squashed files (`001`, `002`) are frozen — all future schema changes go into new numbered files.