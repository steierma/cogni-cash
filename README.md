# 💰 CogniCash: Your Private AI Financial Center

CogniCash is a **privacy-first, self-hosted** financial engine that transforms raw banking data and documents into actionable insights. By combining **Strict Hexagonal Architecture** with **Local AI (Ollama)**, it provides a high-integrity platform for managing your entire financial life without ever leaking data to the cloud.

### 🌟 Why CogniCash?

*   **🧠 Local-First AI Intelligence:** Leverage local LLMs (like Llama 3 via Ollama) to automatically parse and categorize **Invoices, Payslips, and Bank Statements**. Includes a **Hybrid Matcher** that first checks for high-confidence (90%+) historical matches in your local database before calling the AI.
*   **🏦 Seamless Bank Integration:** Synchronize directly with real-world bank accounts via **Enable Banking (PSD2)** or import files from major providers like **ING, Amazon Visa, and VW/CARIAD**.
*   **📊 Precision Analytics & Review:** Master your cash flow with deep-dive analytics, a dedicated **Review Mode (Inbox)** for new transactions, and a smart **Reconciliation Wizard** to link internal transfers and prevent double-counting.
*   **👥 Multi-Tenant by Design:** Built up for **Full User Tenancy**, allowing multiple users to manage their isolated financial data on a single instance.
*   **📱 Offline-First Mobile App:** High-fidelity Flutter app with a **Cache-First (Isar)** architecture and **Mutation Outbox** for seamless financial management without a network connection.
*   **🥧 Raspberry Pi 5 Ready:** Fully compatible with **ARM64** architectures (tested on Raspberry Pi 5).

## Intro Video

https://github.com/user-attachments/assets/44a99551-3589-4b79-b353-bb4d597fd291

-----

## Table of Contents

- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Environment Variables](#environment-variables)
- [System Configuration (Web UI)](#system-configuration-web-ui)
- [Frontend Pages & Capabilities](#frontend-pages--capabilities)
- [Internationalisation (i18n)](#internationalisation-i18n)
- [API Reference](#api-reference)
- [Make Commands](#make-commands)
- [Database](#database)

-----

## Architecture

The project follows **Strict Hexagonal (Ports and Adapters)** architecture. The core domain has zero dependencies on frameworks, HTTP clients, or databases. **A running PostgreSQL database is strictly required.**

```text
┌────────────────────────────────────────────────────────────────┐
│                          Domain (Core)                         │
│   Entities: Invoice, Vendor, Category, BankStatement,          │
│             Transaction, Payslip, User, Reconciliation,        │
│             Setting, ReconciliationPairSuggestion,             │
│             BankConnection, BankAccount                        │
│   Errors:   entity/errors.go (all sentinel errors centralised) │
│   Services: Invoice, BankStatement, Transaction,               │
│             Payslip, Settings, Auth, User, Reconciliation,     │
│             Bank                                               │
│   Ports:    Repos (Invoice, BankStmt, Payslip, Category,       │
│             User, Reconciliation, Settings, Bank), Parsers,    │
│             LLMClient, JobTracker, PayslipParser,              │
│             InvoiceParser, BankProvider                        │
│   Use-Case Ports (Driving): AuthUseCase, UserUseCase,          │
│             InvoiceUseCase, BankStatementUseCase,              │
│             TransactionUseCase, ReconciliationUseCase,         │
│             SettingsUseCase, PayslipUseCase, BankUseCase       │
└────────────┬──────────────────────────────────────────────────┬┘
             │ implements                                       │ implements
   ┌─────────▼───────────┐                          ┌───────────▼────────────┐
   │   Input Adapters    │                          │   Output Adapters      │
   │  (Driving side)     │                          │  (Driven side)         │
   │  • REST API (chi)   │                          │  • PostgresRepo (pgx)  │
   │    depends on port  │                          │  • OllamaAdapter       │
   │    interfaces only  │                          │  • ING PDF/CSV Parser  │
   │  • Dir watcher      │                          │  • Amazon Visa Parser  │
   │  • File upload      │                          │  • VW / CARIAD Parser  │
   │  • JSON cron        │                          │  • AI Fallback Parser  │
   │  • CLI tools        │                          │  • Invoice PDF Parser  │
   │                     │                          │                        │
   └─────────────────────┘                          └────────────────────────┘
   ┌─────────────────────┐
   │  Reverse Proxy      │
   │  • Caddy (HTTPS)    │
   │  • Nginx (SPA)      │
   └─────────────────────┘
```

All feature code is written **test-first (TDD)**. Domain logic is tested with mocks — no database or network connection needed to run the unit tests. The CI pipeline strictly enforces a **filtered logic coverage of >36.8%** (excluding non-logic files) and a targeted **domain service coverage of >68.4%**.

-----

## Tech Stack

| Layer                 | Technology                                    |
|-----------------------|-----------------------------------------------|
| **Backend language** | Go 1.26                                       |
| **HTTP router** | chi v5                                        |
| **Database driver** | pgx v5                                        |
| **PDF parsing** | `ledongthuc/pdf`                              |
| **XLS parsing** | `extrame/xls`                                 |
| **LLM** | Ollama (Llama 3) — configurable via UI or ENV |
| **Frontend** | React 19 + TypeScript + Vite                  |
| **Styling** | Tailwind CSS v4                               |
| **Data fetching** | TanStack Query v5 + Axios                     |
| **Charts** | Recharts 3 (Web) / fl_chart (Mobile)         |
| **Icons** | Lucide React (Web) / Material Icons (Mobile) |
| **Routing** | React Router v7 (Web) / GoRouter (Mobile)    |
| **i18n** | `i18next` (Web) / `L10n` (Mobile)             |
| **Database** | PostgreSQL 16 (Server) / Isar (Mobile Cache)  |
| **Offline Sync** | Mutation Outbox Pattern (Mobile)              |
| **Container runtime** | Docker + Compose                              |
| **CI/CD** | Forgejo Actions → GHCR                        |

-----

## Project Structure

```text
.
├── .forgejo/workflows/
│   ├── ci-cd.yml              # CI/CD: test → build → push to GHCR → deploy
│   └── publish-github.yml     # Sync public-main branch to GitHub mirror
├── backend/
│   ├── cmd/
│   │   ├── healthcheck/       # Distroless-compatible health-check binary
│   │   ├── migrate/           # Standalone DB migration runner
│   │   └── resetpw/           # CLI tool to reset user passwords
│   ├── internal/
│   │   ├── domain/
│   │   │   ├── entity/        # Invoice, Vendor, Category, BankStatement, Transaction,
│   │   │   │                  #   Payslip, User, Reconciliation, Setting,
│   │   │   │                  #   ReconciliationPairSuggestion, errors.go
│   │   │   ├── hash/          # Deterministic SHA-256 content hashing (idempotency keys)
│   │   │   ├── port/          # Repository, parser & use-case interfaces (ports):
│   │   │   │                  #   repos, LLMClient, PayslipParser, InvoiceParser,
│   │   │   │                  #   JobTracker, use_cases.go (8 driving-side ports)
│   │   │   └── service/       # Invoice, BankStatement, Transaction, Payslip,
│   │   │                      #   Auth, User, Reconciliation, Settings, JobManager
│   │   └── adapter/
│   │       ├── http/          # chi REST handler + JWT/RBAC middleware (depends on port interfaces, not service structs)
│   │       ├── ollama/        # LLMClient implementation (Ollama / Gemini)
│   │       ├── parser/
│   │       │   ├── bank_statement/  # ing/, ingcsv/, amazonvisa/, vw/, ai/
│   │       │   ├── invoice/         # PDF text extractor (ledongthuc/pdf)
│   │       │   └── payslip/         # cariad/, ai/
│   │       └── repository/
│   │           └── postgres/  # pgx-based repository implementations
├── migrations/            # Versioned SQL files applied in lexicographic order
│   ├── 001_initial_schema.sql              # Consolidated base schema
│   ├── 002_add_invoice_content_hash.sql    # File storage + dedup columns on invoices
│   ├── 003_add_bank_integration.sql        # Bank API connection tables
│   ├── 004_add_bank_provider.sql           # Multi-provider support
│   ├── 005_add_bank_account_type.sql       # Account type configuration
│   ├── 006_add_transaction_location.sql    # Extracted location column
│   ├── 007_add_transaction_reviewed.sql    # Transaction review status (Inbox)
│   ├── 008_add_password_reset_tokens.sql   # Secure forgot-password flow
│   ├── 009_add_bank_account_sync_error.sql # Transparent sync error reporting
│   ├── 010_add_user_tenancy.sql            # Full multi-tenant isolation
│   ├── 011_enrich_transactions.sql         # Transaction enrichment (Counterparty)
│   └── 012_add_fuzzy_matching.sql          # pg_trgm fuzzy matcher + schema cleanup
├── balance/               # Sample bank statement files for testing
│   ├── payslips/              # Local drop-zone (default PAYSLIP_HOST_DIR)
│   │   ├── payslips_import.json   # Drop here to trigger JSON bulk import
│   │   └── history/               # Permanent archive, organised by year
│   ├── scripts/
│   │   ├── organize_payslip_history.py  # Moves history/ PDFs into year subdirs
│   │   └── testdata/          # Fixture generators (gen_ing_pdf.go, gen_csv.go, gen_amazon_visa.py)
│   ├── .env.example           # Template — copy to .env and fill in values
│   └── Dockerfile             # Multi-stage: golang:1.26-alpine → distroless
├── frontend/
│   ├── src/
│   │   ├── i18n/              # i18n bootstrap & translation catalogues
│   │   │   ├── index.ts       # i18next initialisation (detector + fallback)
│   │   │   └── locales/
│   │   │       ├── en/translation.json   # English (source of truth)
│   │   │       ├── de/translation.json   # German
│   │   │       ├── es/translation.json   # Spanish
│   │   │       └── fr/translation.json   # French
│   │   ├── pages/             # React page components (11 pages)
│   │   ├── components/        # Reusable UI (CategoryBadge, Layout, payslips/, transactions/)
│   │   ├── api/               # Axios client + TypeScript types (client.ts, types.ts)
│   │   └── utils/             # Locale-aware formatters (formatters.ts)
│   ├── nginx.conf             # SPA fallback + /api proxy + security headers
│   └── Dockerfile             # node:22-alpine build → nginx:1.27-alpine serve
├── caddy/
│   └── Caddyfile              # Reverse proxy: HTTPS termination → frontend + backend
├── docs/                      # Conceptual design and architecture docs
├── scripts/
│   └── setup-server.sh        # One-command server bootstrap
├── DATABASE_SCHEMA.md         # Detailed schema documentation
├── docker-compose.yml         # Production: pulls pre-built images from GHCR
├── docker-compose.override.yml # Local dev: builds from source, exposes ports
└── Makefile
```

-----

## Getting Started

### Prerequisites

**Local development (Option A)**

* Go 1.26+
* Node.js 22+
* `psql` CLI — `brew install libpq` (for `make db-*` commands)
* PostgreSQL and Ollama reachable on your network

**Docker deployment (Option B)** — Docker Engine 24+ with the Compose plugin is all that's needed.

### Option A — Local Full Stack (Real DB + Ollama)

```bash
cp backend/.env.example backend/.env
# Edit backend/.env — set POSTGRES_PASSWORD and OLLAMA_URL at minimum
make db-migrate
make backend-run      # Go backend on :8080
make frontend-dev     # Vite dev server on :5173
```

### Option B — Docker (recommended for any server)

```bash
cp backend/.env.example backend/.env
# Copy the local build override
cp docker-compose.override.yml.example docker-compose.override.yml
# Edit backend/.env — set POSTGRES_PASSWORD and DOMAIN_NAME at minimum
make build            # builds backend + frontend images locally
make up               # starts postgres → migrate → backend → frontend → caddy
```

The production `docker-compose.yml` pulls pre-built images from GHCR. For local development, `docker-compose.override.yml` overrides this to build from source and expose ports directly (`:8080` for backend, `:3000` for frontend), disabling the bundled Caddy proxy so it doesn't conflict with your own reverse proxy.

The backend container automatically mounts a payslip drop-zone directory at `/app/payslips`. The host path defaults to `./backend/payslips` and can be overridden via `PAYSLIP_HOST_DIR` in a `.env` file next to `docker-compose.yml`:

```dotenv
# /opt/cogni-cash/.env  (server)
PAYSLIP_HOST_DIR=/tmp/payslips
```

Drop a `payslips_import.json` + PDF files into that directory and the background cron imports them within one interval tick.

-----

## Environment Variables

All variables live in `backend/.env`. Base infrastructure variables are required, while application-specific settings (like LLM prompts or import directories) can be dynamically managed via the Web UI.

### Infrastructure & Security

| Variable               | Default                            | Description                                                            |
|------------------------|------------------------------------|------------------------------------------------------------------------|
| `SERVER_ADDR`          | `:8080`                            | HTTP listen address                                                    |
| `JWT_SECRET`           | *(generated by `setup-server.sh`)* | Secret key used to sign JWTs — **change this in production** |
| `ADMIN_USERNAME`       | `admin`                            | Username for the initial admin web UI login                            |
| `ADMIN_PASSWORD`       | *(generated by `setup-server.sh`)* | Password for the initial admin — seeded on first startup, rotated if changed |
| `DOMAIN_NAME`          | `localhost`                        | Public domain for Caddy auto-HTTPS. Use `localhost` for local testing  |
| `ALLOWED_ORIGINS`      | *(deny all)* | Comma-separated list of allowed CORS origins (e.g. `http://localhost:5173`) |

### Database

| Variable               | Default       | Description                                                    |
|------------------------|---------------|----------------------------------------------------------------|
| `DB_TYPE`              | `postgres`    | Storage mode: `postgres` or `memory` (in-memory for dev)       |
| `POSTGRES_USER`        | —             | Database username                                              |
| `POSTGRES_PASSWORD`    | —             | Database password — **required** |
| `POSTGRES_DB`        | —             | Database name                                                  |
| `DB_TYPE`            | `postgres`    | Storage mode: `postgres` or `memory` (in-memory for dev)       |
| `DEMO_MODE`          | `false`       | Set to `true` to enable mock bank provider interception        |
| `DATABASE_HOST`      | `localhost`   | DB hostname — use `postgres` inside Docker Compose             |

| `DATABASE_PORT`        | `5432`        | DB port                                                        |

### LLM & AI

| Variable               | Default                                | Description                               |
|------------------------|----------------------------------------|-------------------------------------------|
| `OLLAMA_URL`           | `http://localhost:11434`               | Ollama API base URL                       |
| `OLLAMA_MODEL`         | `llama3`                               | Default LLM model name                    |
| `EXAMPLES_PER_CAT`     | `20`                                   | Historical examples per category for few-shot learning |

### Background Workers

| Variable                   | Default              | Description                                                            |
|----------------------------|----------------------|------------------------------------------------------------------------|
| `IMPORT_DIR`               | *(empty)* | Directory to watch for auto-importing bank statement files             |
| `IMPORT_INTERVAL`          | `1h`                 | Polling interval for the directory watcher (Go duration)               |
| `PAYSLIP_IMPORT_JSON_PATH` | *(empty)* | Path to `payslips_import.json` manifest. Worker imports entries, deletes imported PDFs, keeps JSON. |
| `PAYSLIP_IMPORT_INTERVAL`  | `1h`                 | Polling interval for the payslip JSON cron (Go duration)               |
| `PAYSLIP_HOST_DIR`         | `./backend/payslips` | **Docker Compose only.** Host path bind-mounted to `/app/payslips`     |

### Email (Future — SMTP)

| Variable         | Default | Description                      |
|------------------|---------|----------------------------------|
| `SMTP_HOST`      | —       | SMTP server hostname             |
| `SMTP_PORT`      | `587`   | SMTP server port                 |
| `SMTP_USER`      | —       | SMTP authentication username     |
| `SMTP_PASSWORD`  | —       | SMTP authentication password     |
| `SMTP_FROM_EMAIL`| —       | Sender address for outgoing mail |

-----

## System Configuration (Web UI)

Instead of hardcoding functionality in `.env` files, the **Settings Page** (`/settings`) allows real-time configuration of the core application features. Changes are saved directly to the database and take effect immediately.

### 1\. LLM Configuration

* **API URL & Token:** Point the application to any local or remote Ollama instance (e.g., `http://localhost:11434`).
* **Model Name:** Define which model to use (e.g., `llama3`, `deepseek-r1`).
* **Prompt Engineering:** Directly edit the system prompts used for **Single Categorization**, **Batch Categorization**, **Bank Statement Parsing**, and **Payslip Parsing**. Supports dynamic placeholders (`{{CATEGORIES}}`, `{{TEXT}}`, `{{DATA}}`).

### 2\. Background Automation

* **Auto-Import:** Define an absolute directory path for the backend to watch. The system will automatically import any recognized PDF, CSV, or XLS files found here at the designated polling interval (e.g., `1h`).
* **Auto-Categorization:** Enable or disable background transaction processing. Configure the polling interval (e.g., `5m`) and control the LLM load by setting a **Batch Size** (the number of transactions sent to the LLM per prompt). Enhance accuracy by configuring **Learning Examples per Category** (the number of unique historical categorizations provided to the LLM for few-shot learning).
* **Smart Bank Sync:** Automatically synchronizes all connected bank accounts every **second day**. To ensure reliable data fetching and simulate natural usage patterns, the execution time is **randomized** between **11:00 and 13:00**. The schedule is persistent across restarts.
* **Payslip JSON Import:** Drop a `payslips_import.json` manifest and the referenced PDF files into the payslip drop-zone directory
 (default `./backend/payslips`, configurable via `PAYSLIP_HOST_DIR`). The background cron worker picks up the JSON on the next tick, reads and stores the binary content of each PDF that is present on disk, imports every entry into the database (skipping duplicates), **deletes each successfully imported PDF**, and **keeps the JSON manifest** permanently so it can be extended with new entries at any time.

### 3\. Appearance & Language

* **Theme:** System / light / dark.
* **Default Currency:** Used for analytics display.
* **UI Language:** Select from English, German, Spanish, or French. Applied instantly and persisted to the database.

-----

## Frontend Pages & Capabilities

The React frontend has been built to provide deep analytics and efficient batch management.

| Route              | Page             | Key Capabilities                                                                                                                                                                                                                                                                                              |
|--------------------|------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `/`                | **Dashboard** | View dynamic KPIs (Income, Expenses, Net Savings), a scrollable cash flow timeline, top spending categories visually represented by progress bars, and recent transactions. Global toggle to show/hide reconciled settlement payments.                                                                       |
| `/analytics`       | **Analytics** | Advanced visualizations including period-specific KPIs, monthly income/expense trends, and category-based spending breakdowns with negative filtering.                                                                                                                                                       |
| `/transactions`    | **Transactions** | Comprehensive data table with advanced filtering (date ranges, amounts, search, statement, category) and location extraction. **"Unreviewed Only" default inbox** for newly synced data. Perform manual single, **batch category assignment**, or **batch confirmation**. Features inline visual charts based on active filters and controls to manually trigger/cancel the background LLM Auto-Categorization job. |
| `/invoices`        | **Invoices** | Full invoice management. **Drag & Drop / click-to-upload** PDF import with automatic LLM categorization. Submit raw text directly via the categorize panel. Sortable, searchable table with filters for category, date range, and amount. Inline **edit** of vendor, category, amount, currency, date, and description. **Download** the original uploaded file. **Batch delete** with multi-select. Visual analytics panel showing total invoice spend, top vendors chart, and monthly invoice trend. |
| `/categories`      | **Categories** | Centralized category management. Create, rename, or delete categories and assign custom hexadecimal colors used across charts and badges.                                                                                                                                                                    |
| `/bank-statements` | **Statements** | List all imported files with integrated **Drag & Drop file upload** for PDF, CSV, and XLS statements. Distinguishes visually between Giro, Credit Card, and Extra Account statements. View transaction counts and period balances. |
| `/bank-connections` | **Connections** | Manage live bank API connections. Search for and link banks via **Enable Banking**. View linked accounts with **real-time balances**, human-readable institution names, and **transparent sync error reporting (tooltips)**. Configure **Account Types** (Giro, Credit Card, Extra Account) per account for accurate reconciliation. Sync all accounts in background. |
| `/payslips`        | **Payslips** | Full HR document management. **Quick drag-and-drop** single PDF upload with a **Force AI Parsing** toggle and a **Manual Override modal** to force-correct any field. **Batch upload** of multiple PDFs at once. **PDF preview** inline in-browser and **download** of the stored original. View/Edit modal for all structured fields including bonuses. KPI cards for latest gross/net/adjusted net with month-over-month trend. Salary trend and yearly charts (Gross, Total Net, Adjusted Net, Payout lines) with configurable **bonus exclusion** controls. Filterable by period range, employee, and tax class. Column visibility persisted to settings. Payslips imported via JSON manifest show a grayed-out preview/download button. |
| `/reconcile`       | **Reconcile** | Dedicated 1:1 transaction reconciliation wizard. Globally scans all pending accounts to find exact matching internal transfers (where a debit and credit sum to zero) and links them to prevent double-counting in analytics. Supports both statement-based and **Live Feed** transactions with a **floating action button** for rapid batch linking. |
| `/users`           | **Users** | Manage system access and profiles. View user details, create new users, modify roles (Admin or Manager), and delete accounts. This route is strictly protected via RBAC (Admins only). |
| `/settings`        | **Settings** | Configure LLM parameters, edit system prompts, manage background auto-import/categorization intervals, change themes, **select UI language**, update passwords, and persist UI preferences. |
| `/login`           | **Login** | JWT-based authentication. Redirects to Dashboard on success. |


-----

## Internationalisation (i18n)

The frontend supports **four** display languages via **`i18next`** and **`react-i18next`**:

| Language | Locale | Status |
|----------|--------|--------|
| English  | `en`   | Source of truth |
| German   | `de`   | Fully translated |
| Spanish  | `es`   | Fully translated |
| French   | `fr`   | Fully translated |

The active language is auto-detected from the browser on first visit and persisted to the database (settings key `ui_language`) so it survives browser clears and roams across devices.

All pages and components use `useTranslation()` — zero hard-coded user-visible strings remain in the JSX.

-----

## API Reference

The backend exposes a RESTful API under the `/api/v1` namespace. For a detailed list of all endpoints and parameters, see the [API Reference Documentation](docs/API_REFERENCE.md).

-----
| `GET` | `/api/v1/bank/connections` | List active bank connections (includes nested accounts and balances) |
| `DELETE` | `/api/v1/bank/connections/{id}` | Delete a bank connection |
| `POST` | `/api/v1/bank/sync` | Manually trigger background account sync |

### Payslips

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/payslips/` | List all payslips |
| `GET` | `/api/v1/payslips/{id}` | Get payslip by ID |
| `GET` | `/api/v1/payslips/{id}/download` | Download original PDF |
| `POST` | `/api/v1/payslips/import` | Import single payslip (multipart) |
| `POST` | `/api/v1/payslips/import/batch` | Batch import multiple payslips |
| `PUT` | `/api/v1/payslips/{id}` | Update payslip (JSON or multipart) |
| `PATCH` | `/api/v1/payslips/{id}` | Partial update payslip |
| `DELETE` | `/api/v1/payslips/{id}` | Delete payslip |

-----

## Make Commands

```bash
# Infrastructure
make up                 # Start all containers (creates .env from example if missing)
make down               # Stop all containers
make restart            # Restart all containers
make logs               # Tail logs (all services, or: make logs backend)
make ps                 # Show container status

# Docker Images
make build              # Build both backend + frontend images
make build-backend      # Build only the backend image
make build-frontend     # Build only the frontend image
make build TAG=v1.0.0   # Build with a version tag

# Deploy
make deploy             # Full deploy: build → rsync → pull → up on remote
make setup-server       # Run setup-server.sh on DEPLOY_HOST as root

# Backend
make dev                # Local dev server
make backend-run        # Production server (reads backend/.env)
make backend-build      # Compile binary → backend/bin/server
make backend-test       # Run all Go tests
make gen-testdata       # Regenerate anonymised parser fixture files (PDF/CSV/XLS)

# Frontend
make frontend-dev       # Vite dev server on :5173
make frontend-dev-prod  # Vite dev server proxying to production backend
make frontend-build     # Production build

# Database
make db-migrate         # Apply pending SQL migrations
make db-truncate        # Wipe all rows, keep schema
make db-nuke            # Drop all tables + re-migrate
make db-reset           # Drop + recreate database + re-migrate
make db-shell         # Open psql shell
make db-seed          # Insert dummy data from balance/dummy-data.sql
make db-dump-local    # Create a compressed backup of your local database
make db-reset-password USERNAME=admin PASSWORD=newpass  # Reset a user's password

make db-dump-remote     # Dump production DB to local backups/ directory
make db-restore-remote FILE=backups/dump.sql.gz  # Restore a dump to production
make db-restore-local FILE=backups/prod_db_20260326_134748.sql.gz  # Restore a dump to local DB (reads backend/.env for connection)

# Cleanup
make clean              # Remove compiled Go binary
make clean-all          # Remove binary + frontend dist + node_modules
make nuke               # ⚠️ Stop containers + wipe volumes + remove images + clean artifacts
```

-----

## Database

Migrations are plain SQL files in `backend/migrations/`, applied in lexicographic order by the `cogni-cash-migrate` binary. The binary runs automatically as a one-shot container before the backend starts (see `docker-compose.yml`).

| Migration | Description                                                                                                                                                                                                                           |
|---|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `001_initial_schema.sql` | Consolidated schema: all tables (`users`, `categories`, `bank_statements`, `reconciliations`, `transactions`, `invoices`, `settings`, `payslips`, `payslip_bonuses`) with all columns, indexes, and constraints in their final state. |
| `002_add_invoice_content_hash.sql` | Adds `description`, `content_hash`, `original_file_name`, `original_file_mime`, `original_file_size`, `original_file_content` columns to `invoices` for file deduplication and binary storage. |
| `003_add_bank_integration.sql` | Adds `bank_connections` and `bank_accounts` tables for real API integration. |
| `004_add_bank_provider.sql` | Adds `provider` column to `bank_connections` to support multiple aggregators. |
| `005_add_bank_account_type.sql` | Adds `account_type` to `bank_accounts` and `statement_type` to `transactions` table to support cross-account reconciliation for API feeds. |
| `006_add_transaction_location.sql` | Adds `location` column to `transactions` table. |
| `007_add_transaction_reviewed.sql` | Adds `reviewed` boolean column to `transactions` table to support the "Inbox" workflow. |
| `008_add_password_reset_tokens.sql` | Adds `password_reset_tokens` table for the secure forgot-password flow. |
| `009_add_bank_account_sync_error.sql` | Adds `last_sync_error` column to `bank_accounts` for UI visibility. |
| `010_add_user_tenancy.sql` | Implements **Full User Tenancy** (Isolation) across all tables and entities. |
| `011_enrich_transactions.sql` | Adds `counterparty_name`, `counterparty_iban`, `bank_transaction_code` and `mandate_reference` to transactions for improved categorization and reconciliation. |
| `012_add_fuzzy_matching.sql` | Enables `pg_trgm` extension and adds GIN indexes for high-performance fuzzy matching. Includes schema cleanup (removes redundant fields across all document tables). |

-----

## Target Architecture & Roadmap

### 1. Frictionless Deployment (CI/CD & Pre-built Images) ✅

Pre-built Docker images are published to GHCR (`ghcr.io/steierma/cogni-cash-backend`, `ghcr.io/steierma/cogni-cash-frontend`) via the CI/CD pipeline. The production `docker-compose.yml` pulls these images directly — no local build tools required.

* [x] **Targeted Deployment:** Optimized CI/CD to only sync `docker-compose.yml` and pull latest images, preserving all server-side state like `.env` and `docker-compose.override.yml`.
* [x] **State Preservation:** Eliminated `rsync --delete` to ensure manually managed keys (e.g., Enable Banking PEMs) are never accidentally removed.

### 2. External Security & HTTPS (Caddy Proxy) ✅

The stack includes a bundled Caddy reverse proxy that handles automatic HTTPS via Let's Encrypt. The `DOMAIN_NAME` environment variable controls the certificate domain. For local development, `docker-compose.override.yml` disables Caddy so it doesn't conflict with your own proxy. Both the Nginx SPA container and the Go backend enforce security headers (CSP, HSTS, X-Frame-Options, etc.).

### 3. Complete Invoice Use Case ✅

* [x] **Entity separation:** `Category`, `Vendor` extracted into dedicated files.
* [x] **Errors:** `ErrInvoiceDuplicate` and `ErrInvoiceNotFound` added to `entity/errors.go`.
* [x] **Port:** `InvoiceParser` output port for pluggable file-to-text extraction.
* [x] **Port:** `InvoiceUseCase` driving-side port — covers `ImportFromFile`, `CategorizeDocument`, `GetAll`, `GetByID`, `Update`, `Delete`, `GetOriginalFile`.
* [x] **Service:** `InvoiceService` — orchestrates SHA-256 dedup, text extraction, LLM categorization, CRUD, and original-file retrieval.
* [x] **Postgres adapter:** `InvoiceRepository` fully implements the new port with binary storage support.
* [x] **Frontend:** `InvoicesPage` rewritten — drag-and-drop import, sortable/filterable table, inline edit modal, batch delete, analytics visuals panel.

### 4. Real Bank API Integration (PSD2 / Multi-Provider) ✅

* [x] **Entity:** `BankConnection` and `BankAccount` domain entities.
* [x] **Port:** `BankProvider` interface for PSD2 aggregators.
* [x] **Adapter:** Enable Banking implementation (multi-provider support).
* [x] **Service:** `BankService` for connection management and background transaction syncing.
* [x] **UI:** `BankConnectionsPage` with institution search and OAuth redirect flow.
* [x] **UI:** Settings-backed provider configuration.
* [x] **Deduplication:** Automatic `ContentHash` generation for synced transactions.
* [x] **Metadata Enrichment:** Extraction and storage of `counterparty_name`, `counterparty_iban`, `bank_transaction_code`, and `mandate_reference` for all synced transactions.
* [x] **Smarter AI:** Automatically passing all enriched metadata to the LLM for significantly higher categorization accuracy.

### 5. Email & Notifications (SMTP) ✅

SMTP integration is fully implemented for automated user notifications and system alerts. Configuration can be managed via environment variables for initial seeding or directly through the Web UI Settings page.

* [x] Implement `EmailProvider` port and `SMTPAdapter` (using `net/smtp`).
* [x] Create `NotificationUseCase` and `NotificationService` for domain-level alerts.
* [x] **Welcome Emails:** Automatically sent asynchronously upon new user creation.
* [x] **Password Reset:**
    * [x] **Architecture & Design:** Token-based flow documented in `docs/PASSWORD_RESET_CONCEPT.md`.
    * [x] **Database:** Implementation of `password_reset_tokens` table (Migration `008`).
    * [x] **Backend:** Secure token generation (CSPRNG), SHA-256 hashing, and verification.
    * [x] **Frontend:** `ForgotPasswordPage` and `ResetPasswordPage` UI with token validation.
* [x] **UI Configuration:** SMTP host, port, credentials, and sender email managed via `/settings`.
* [x] **i18n:** Fully translated across EN, DE, ES, and FR.

### 6. Full User Tenancy ✅

* [x] **Database Isolation:** Migration `010` adds `user_id` to all data-owning tables.
* [x] **Backend Tenancy:** All repositories and services filtered by `UserID`.

### 7. Hybrid AI Categorization ✅

The system uses a sophisticated hybrid approach to transaction categorization that balances speed, cost, and intelligence.

* [x] **DB-First Matching:** Before calling the AI, the system checks the local database for exact or high-confidence (65%+) historical matches using trigram similarity indexes.
* [x] **Few-Shot Learning:** If no high-confidence match is found, the system calls the local LLM (Ollama), providing a total of 20 unique historical examples from the user's data to improve accuracy.
* [x] **Performance Optimization:** Implemented `pg_trgm` and GIN indexes to ensure fuzzy matching remains instantaneous as data grows.

### 8. Resilient Backup Strategy (Proxmox + rclone)

* [ ] Implement off-site encrypted backups of PostgreSQL dumps and the `payslips` directory to Google Drive via `rclone`.
* [ ] Document local manual backup procedures and enforce manual Proxmox snapshots prior to major upgrades.
* [ ] `make db-dump-remote` and `make db-restore-remote` are available for manual backup/restore operations.
