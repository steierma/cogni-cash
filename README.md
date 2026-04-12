# 💰 CogniCash: Your Private AI Financial Center

CogniCash is a **privacy-first, self-hosted** financial engine that transforms raw banking data and documents into
actionable insights. By combining **Strict Hexagonal Architecture** with **Local AI (Ollama)**, it provides a
high-integrity platform for managing your entire financial life without ever leaking data to the cloud.

### 🌟 Why CogniCash?

* **🧠 Local-First AI Intelligence:** Leverage local LLMs (like Llama 3 via Ollama) to automatically parse and categorize
  **Invoices, Payslips, and Bank Statements**. Supports PDF and image formats (JPEG, PNG, GIF, WEBP) via multimodal
  Gemini or AI-fallback paths. All AI processing happens **locally on your hardware** — your sensitive financial data
  never leaves your network. Includes a **Hybrid Matcher** that first checks for high-confidence (65%+) historical
  matches before calling the AI.
* **🔮 Predictive Intelligence (The End of Surprise Expenses):** Look into your financial future with the **Forecasting
  Engine**. Automatically detects recurring patterns (rent, salary, subscriptions) even up to **Yearly intervals** with
  a fixed 3-year historical lookback. Calculates a **Monthly Burn Rate** for variable categories (groceries, fuel), and
  even projects **Seasonal Bonuses** by decomposing payslip data from your bank transactions. Now includes **Forecast
  Fine-Tuning**, allowing you to manually exclude specific future projections or **Mute an entire recurring pattern** with a single click from the transactions page (and restore them if needed). Supports range options from 1 to 12 months with stable, deterministic

  projections. Learn more in our [User Guide](docs/USER_GUIDE.md).
* **🛡️ Privacy-First Integrity (Deep-Scrubbing Policy):** We maintain a strictly **personal-data-free codebase**. Our
  automated "deep-scrubbing" policy ensures that all test data, logs, and documentation examples are completely
  anonymized or synthetic. Your financial privacy is not just a feature; it's baked into our development lifecycle,
  ensuring no real-world sensitive data ever accidentally enters the version control history.
* **📜 Professional Payslip Management:** Master your HR documents with a dedicated **Payslip Engine**. Automatically
  extract Gross, Net, Payout, and Bonuses. Includes a **Split-View Preview** to compare the original PDF with the
  extracted data for 100% accuracy.
* **🏦 Offline-First Parsers (ING & more):** Built-in, privacy-respecting offline parsers for major providers like *
  *ING (DIBA), Amazon Visa, and VW/CARIAD**. These parsers work entirely without AI for maximum speed and reliability,
  with many more provider-specific parsers currently in development.
* **📈 Precision Analytics:** Master your cash flow with deep-dive analytics, a dedicated **Review Mode (Inbox)** for new
  transactions, and a smart **Reconciliation Wizard** to link internal transfers and prevent double-counting.
* **👥 Multi-Tenant by Design:** Built from the ground up for **Full User Tenancy**, allowing multiple users to manage
  their isolated financial data on a single shared instance.
* **📱 Native Mobile App:** High-fidelity Flutter app with a **Cache-First (Isar)** architecture and **Mutation Outbox**
  for seamless financial management without a network connection.
* **🥧 Raspberry Pi 5 Ready:** Fully compatible with **ARM64** architectures — run your entire financial center on a
  low-power Pi 5.

## Intro Video

https://github.com/user-attachments/assets/44a99551-3589-4b79-b353-bb4d597fd291

-----

## 📱 Mobile Experience (Beta)

Manage your finances on the go with the **CogniCash Mobile App**. Built with Flutter, it offers a seamless,
offline-first experience that syncs perfectly with your self-hosted instance.

### ✨ Key Mobile Features

* **🤖 Local-First AI Recognition:** Snap a photo or upload a PDF of your **payslips, bank statements, or invoices**. The
  app uses local AI logic (via your self-hosted backend) to recognize and categorize your data instantly without cloud
  leaks.
* **📜 Payslip Master View:** Detailed management of your salary history, bonuses, and tax classes with high-fidelity
  charts.
* **🏦 ING DIBA & More:** Native support for the **ING DIBA offline parser**, ensuring maximum privacy and speed for your
  banking imports, with many more providers coming soon.
* **Offline-First:** View and manage your data even without an internet connection.
* **Native Performance:** Smooth 60fps animations and transitions.
* **Document Preview:** View original documents directly in the app while you edit or review extracted fields.
* **Biometric Security:** Protect your financial data with Fingerprint or Face ID.
* **🔏 License Management (Beta):** Transparent trial tracking with a dedicated **License & Trial** view. The app uses a "Silent Registration" mechanism to manage its development status. If your trial expires, the app enters **Vault Mode**, allowing full read-only access to your history while blocking new mutations until renewed. **Note:** This is currently a client-side beta gate and does not yet restrict backend access.
* **Hardware ID:** Conveniently copy your unique, privacy-salted Hardware ID for manual license upgrades or support.

<p align="center">
  <img src="screenshots/CogniCash-portfolio.png" width="100%" alt="CogniCash Portfolio">
</p>

#### Smartphone Experience

<p align="center">
  <img src="screenshots/spartphone_dashboard.jpg" width="32%" alt="Dashboard">
  <img src="screenshots/smartphone_transactions.jpg" width="32%" alt="Transactions">
  <img src="screenshots/spartphone_analytics.jpg" width="32%" alt="Analytics">
</p>

#### Tablet Optimization

<p align="center">
  <img src="screenshots/tablet_dashboard.jpeg" width="49%" alt="Tablet Dashboard">
  <img src="screenshots/tablet_analytics.jpeg" width="49%" alt="Tablet Analytics">
</p>

### 🚀 Interested in the Mobile App?

I am currently preparing for a public release on the **Google Play Store**.

If you are interested in joining the **Beta Program** or want to be notified when the app is available, please **contact
me**:
👉 [support-cogni-cash@steierl.org](mailto:support-cogni-cash@steierl.org?subject=Interest%20in%20CogniCash%20Mobile)

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

The project follows **Strict Hexagonal (Ports and Adapters)** architecture. The core domain has zero dependencies on
frameworks, HTTP clients, or databases. **A running PostgreSQL database is strictly required.**

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
````

All feature code is written **test-first (TDD)**. Domain logic is tested with mocks — no database or network connection
needed to run the unit tests. The CI pipeline strictly enforces a **filtered logic coverage of \>36.8%** (excluding
non-logic files) and a targeted **domain service coverage of \>68.4%**.

-----

## Tech Stack

| Layer                 | Technology                                    |
|-----------------------|-----------------------------------------------|
| **Backend language** | Go 1.26                                       |
| **HTTP router** | chi v5                                        |
| **Database driver** | pgx v5                                        |
| **PDF parsing** | `ledongthuc/pdf`                              |
| **Image parsing** | Google Gemini (multimodal) / AI fallback      |
| **XLS parsing** | `extrame/xls`                                 |
| **LLM** | Ollama (Llama 3) — configurable via UI or ENV |
| **Frontend** | React 19 + TypeScript + Vite                  |
| **Styling** | Tailwind CSS v4                               |
| **Data fetching** | TanStack Query v5 + Axios                     |
| **Charts** | Recharts 3 (Web) / fl\_chart (Mobile)         |
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
│   │   │   ├── port/          # Repository, parser & use-case interfaces (ports)
│   │   │   └── service/       # Domain orchestrators (Invoice, BankStatement, Auth, etc.)
│   │   └── adapter/
│   │       ├── http/          # chi REST handler + JWT/RBAC middleware
│   │       ├── ollama/        # LLMClient implementation (Ollama / Gemini)
│   │       ├── parser/        # Parsers for ing, amazonvisa, vw, pdf invoices, etc.
│   │       └── repository/    # pgx-based repository implementations
├── migrations/                # Versioned SQL files (applied in lexicographic order)
│   ├── 001_initial_schema.sql # Full schema for fresh installs (squashed v2.0.0)
│   └── 002_squash_catchup.sql # Catch-up for existing DBs: applies migrations 013–017
├── balance/                   # Sample bank statement files for testing
│   ├── payslips/              # Local drop-zone (default PAYSLIP_HOST_DIR)
│   │   ├── payslips_import.json   # Drop here to trigger JSON bulk import
│   │   └── history/               # Permanent archive, organised by year
│   ├── scripts/               # Test data generators & organizers
│   ├── .env.example           # Template — copy to .env and fill in values
│   └── Dockerfile             # Multi-stage: golang:1.26-alpine → distroless
├── frontend/
│   ├── src/
│   │   ├── i18n/              # i18n bootstrap & translation catalogues (en, de, es, fr)
│   │   ├── pages/             # React page components (11 pages)
│   │   ├── components/        # Reusable UI (CategoryBadge, Layout, etc.)
│   │   ├── api/               # Axios client + TypeScript types
│   │   └── utils/             # Locale-aware formatters
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

### ⚠️ Upgrading to v2.0.0 (From v1.5.x or older)

In version 2.0.0, the historical database schema (migrations 001 through 017) was consolidated into two files:
- **`001_initial_schema.sql`** — the full target schema for fresh installs
- **`002_squash_catchup.sql`** — a safe, idempotent catch-up for any existing database

If you are upgrading an existing instance, your database's internal migration tracking will conflict with the new
file structure. **Before running `make up` or deploying v2.0.0**, run the one-time sync command:

**Upgrade Steps for Existing Users:**

1. **Take a backup** (recommended):
   ```bash
   make db-dump-local
   ```
2. **Rewrite migration history and apply the catch-up:**
   ```bash
   make db-squash-upgrade
   ```
   This removes all old numbered entries from `schema_migrations`, then runs `make db-migrate` to stamp the correct
   content hashes for both squashed files and apply any missing columns/tables via `002_squash_catchup.sql`.

3. **Start the application normally:**
   ```bash
   make up
   ```

*(Note: If you are installing CogniCash for the very first time, ignore this step entirely.)*

-----

### Prerequisites

**New User Registration:** CogniCash is a private, multi-tenant system. There is no public sign-up page. The first user (Admin) is created via environment variables; all subsequent users must be added by an administrator via the Web UI. See [First Steps](docs/USER_GUIDE.md#🚀-0-getting-started-first-login--setup) for details.

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
# Edit backend/.env — set POSTGRES_PASSWORD and OLLAMA_URL at minimum
# NOTE: Use your host's network IP for OLLAMA_URL, not 'localhost'

# If using Enable Banking, add your private key
touch enable-banking-prod.pem 

make build            # builds backend + frontend images locally
make up               # starts postgres → migrate → backend → frontend → caddy
```

The production `docker-compose.yml` pulls pre-built images from GHCR. For local development,
`docker-compose.override.yml` overrides this to build from source and expose ports directly (`:8080` for backend,
`:3000` for frontend), disabling the bundled Caddy proxy so it doesn't conflict with your own reverse proxy.

The backend container automatically mounts a payslip drop-zone directory at `/app/payslips`. The host path defaults to
`./backend/payslips` and can be overridden via `PAYSLIP_HOST_DIR` in a `.env` file next to `docker-compose.yml`:

```dotenv
# /opt/cogni-cash/.env  (server)
PAYSLIP_HOST_DIR=/tmp/payslips
```

Drop a `payslips_import.json` + PDF files into that directory and the background cron imports them within one interval
tick.

-----

## Environment Variables

All variables live in `backend/.env`. Base infrastructure variables are required, while application-specific settings (
like LLM prompts or import directories) can be dynamically managed via the Web UI.

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

| `DATABASE_PORT`        | `5432`        | DB port |

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

Instead of hardcoding functionality in `.env` files, the **Settings Page** (`/settings`) allows real-time configuration
of the core application features. Changes are saved directly to the database and take effect immediately.

### 1\. LLM Configuration

* **API URL & Token:** Point the application to any local or remote Ollama instance (e.g., `http://localhost:11434`).
* **Model Name:** Define which model to use (e.g., `llama3`, `deepseek-r1`).
* **Prompt Engineering:** Directly edit the system prompts used for **Single Categorization**, **Batch Categorization**,
  **Bank Statement Parsing**, and **Payslip Parsing**. Supports dynamic placeholders (`{{CATEGORIES}}`, `{{TEXT}}`,
  `{{DATA}}`).

### 2\. Background Automation

* **Auto-Import:** Define an absolute directory path for the backend to watch. The system will automatically import any
  recognized PDF, CSV, or XLS files found here at the designated polling interval (e.g., `1h`).
* **Auto-Categorization:** Enable or disable background transaction processing. Configure the polling interval (e.g.,
  `5m`) and control the LLM load by setting a **Batch Size** (the number of transactions sent to the LLM per prompt).
  Enhance accuracy by configuring **Learning Examples per Category** (the number of unique historical categorizations
  provided to the LLM for few-shot learning).
* **Smart Bank Sync:** Automatically synchronizes all connected bank accounts every **second day**. To ensure reliable
  data fetching and simulate natural usage patterns, the execution time is **randomized** between **11:00 and 13:00**.
  The schedule is persistent across restarts.
* **Payslip JSON Import:** Drop a `payslips_import.json` manifest and the referenced PDF files into the payslip
  drop-zone directory
  (default `./backend/payslips`, configurable via `PAYSLIP_HOST_DIR`). The background cron worker picks up the JSON on
  the next tick, reads and stores the binary content of each PDF that is present on disk, imports every entry into the
  database (skipping duplicates), **deletes each successfully imported PDF**, and **keeps the JSON manifest**
  permanently so it can be extended with new entries at any time.

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

The active language is auto-detected from the browser on first visit and persisted to the database (settings key
`ui_language`) so it survives browser clears and roams across devices.

All pages and components use `useTranslation()` — zero hard-coded user-visible strings remain in the JSX.

-----

## 🔗 Deep Linking & URL State Synchronization

CogniCash supports comprehensive **URL state synchronization** across all major frontend pages. This allows you to bookmark specific filtered views, share deep links, and use the browser's back/forward buttons naturally.

* **Transactions:** Persists search terms, category/statement filters, date/amount ranges, and sorting.
* **Analytics:** Persists date ranges, excluded categories, and reconciliation visibility.
* **Forecasting:** Persists forecast range (30-365 days), active tab, and search filters.
* **Invoices:** Persists all filters, sorting, and visual toggle states.

-----

## API Reference

The backend exposes a RESTful API under the `/api/v1` namespace. For a detailed list of all endpoints and parameters,
see the [API Reference Documentation](https://www.google.com/search?q=docs/API_REFERENCE.md).

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
make db-squash-upgrade  # ⚠️ One-time: rewrite migration history + apply catch-up for v2.0.0 upgrade
make db-truncate        # Wipe all rows, keep schema
make db-nuke            # Drop all tables + re-migrate
make db-reset           # Drop + recreate database + re-migrate
make db-shell           # Open psql shell
make db-seed            # Insert dummy data from balance/dummy-data.sql
make db-dump-local      # Create a compressed backup of your local database
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

Migrations are plain SQL files in `backend/migrations/`, applied in lexicographic order by the `cogni-cash-migrate`
binary. The binary runs automatically as a one-shot container before the backend starts (see `docker-compose.yml`).
The runner tracks each file by filename stem and SHA-256 content hash — only changed or new files are (re-)applied.

| Migration | Description |
|---|---|
| `001_initial_schema.sql` | **Full schema for fresh installs.** All tables, indexes, extensions, constraints, and seed data in one idempotent script. |
| `002_squash_catchup.sql` | **Catch-up for existing databases.** Idempotently applies everything from migrations 013–017 (`is_variable_spending`, `planned_transactions`, `excluded_forecasts`, `pattern_exclusions`, `is_payslip_verified`). No-op on fresh installs. |

> **Adding future migrations:** create new files starting at `003_...sql`. The two squashed files are frozen.
>
> See [DATABASE_SCHEMA.md](DATABASE_SCHEMA.md) for the full schema reference and upgrade instructions.

-----

## Target Architecture & Roadmap

### 1\. Frictionless Deployment (CI/CD & Pre-built Images) ✅

Pre-built Docker images are published to GHCR (`ghcr.io/steierma/cogni-cash-backend`,
`ghcr.io/steierma/cogni-cash-frontend`) via the CI/CD pipeline. The production `docker-compose.yml` pulls these images
directly — no local build tools required.

* [x] **Targeted Deployment:** Optimized CI/CD to only sync `docker-compose.yml` and pull latest images, preserving all
  server-side state like `.env` and `docker-compose.override.yml`.
* [x] **State Preservation:** Eliminated `rsync --delete` to ensure manually managed keys (e.g., Enable Banking PEMs)
  are never accidentally removed.

### 2\. External Security & HTTPS (Caddy Proxy) ✅

The stack includes a bundled Caddy reverse proxy that handles automatic HTTPS via Let's Encrypt. The `DOMAIN_NAME`
environment variable controls the certificate domain. For local development, `docker-compose.override.yml` disables
Caddy so it doesn't conflict with your own proxy. Both the Nginx SPA container and the Go backend enforce security
headers (CSP, HSTS, X-Frame-Options, etc.).

### 3\. Complete Invoice Use Case ✅

* [x] **Entity separation:** `Category`, `Vendor` extracted into dedicated files.
* [x] **Errors:** `ErrInvoiceDuplicate` and `ErrInvoiceNotFound` added to `entity/errors.go`.
* [x] **Port:** `InvoiceParser` output port for pluggable file-to-text extraction.
* [x] **Port:** `InvoiceUseCase` driving-side port — covers `ImportFromFile`, `CategorizeDocument`, `GetAll`, `GetByID`,
  `Update`, `Delete`, `GetOriginalFile`.
* [x] **Service:** `InvoiceService` — orchestrates SHA-256 dedup, text extraction, LLM categorization, CRUD, and
  original-file retrieval.
* [x] **Postgres adapter:** `InvoiceRepository` fully implements the new port with binary storage support.
* [x] **Frontend:** `InvoicesPage` rewritten — drag-and-drop import, sortable/filterable table, inline edit modal, batch
  delete, analytics visuals panel.

### 4\. Real Bank API Integration (PSD2 / Multi-Provider) ✅

* [x] **Entity:** `BankConnection` and `BankAccount` domain entities.
* [x] **Port:** `BankProvider` interface for PSD2 aggregators.
* [x] **Adapter:** Enable Banking implementation (multi-provider support).
* [x] **Service:** `BankService` for connection management and background transaction syncing.
* [x] **UI:** `BankConnectionsPage` with institution search and OAuth redirect flow.
* [x] **UI:** Settings-backed provider configuration.
* [x] **Deduplication:** Automatic `ContentHash` generation for synced transactions.
* [x] **Metadata Enrichment:** Extraction and storage of `counterparty_name`, `counterparty_iban`,
  `bank_transaction_code`, and `mandate_reference` for all synced transactions.
* [x] **Smarter AI:** Automatically passing all enriched metadata to the LLM for significantly higher categorization
  accuracy.

### 5\. Email & Notifications (SMTP) ✅

SMTP integration is fully implemented for automated user notifications and system alerts. Configuration can be managed
via environment variables for initial seeding or directly through the Web UI Settings page.

* [x] Implement `EmailProvider` port and `SMTPAdapter` (using `net/smtp`).
* [x] Create `NotificationUseCase` and `NotificationService` for domain-level alerts.
* [x] **Welcome Emails:** Automatically sent asynchronously upon new user creation.
* [x] **Password Reset:**
    * [x] **Architecture & Design:** Token-based flow documented in `docs/PASSWORD_RESET_CONCEPT.md`.
    * [x] **Database:** Secure `password_reset_tokens` table with indexed token hash and expiry.
    * [x] **Backend:** Secure token generation (CSPRNG), SHA-256 hashing, and verification.
    * [x] **Frontend:** `ForgotPasswordPage` and `ResetPasswordPage` UI with token validation.
* [x] **UI Configuration:** SMTP host, port, credentials, and sender email managed via `/settings`.
* [x] **i18n:** Fully translated across EN, DE, ES, and FR.

### 6\. Full User Tenancy ✅

* [x] **Database Isolation:** `user_id` added to all data-owning tables; all unique constraints scoped per user.
* [x] **Backend Tenancy:** All repositories and services filtered by `UserID`.

### 7\. Hybrid AI Categorization ✅

The system uses a sophisticated hybrid approach to transaction categorization that balances speed, cost, and
intelligence.

* [x] **DB-First Matching:** Before calling the AI, the system checks the local database for exact or high-confidence (
  65%+) historical matches using trigram similarity indexes.
* [x] **Few-Shot Learning:** If no high-confidence match is found, the system calls the local LLM (Ollama), providing a
  total of 20 unique historical examples from the user's data to improve accuracy.
* [x] **Performance Optimization:** Implemented `pg_trgm` and GIN indexes to ensure fuzzy matching remains instantaneous
  as data grows.

### 8\. Resilient Backup Strategy (Proxmox + rclone)

* [ ] Implement off-site encrypted backups of PostgreSQL dumps and the `payslips` directory to Google Drive via
  `rclone`.
* [ ] Document local manual backup procedures and enforce manual Proxmox snapshots prior to major upgrades.
* [ ] `make db-dump-remote` and `make db-restore-remote` are available for manual backup/restore operations.

### 9\. Financial Forecasting (Upcoming Transactions) ✅

* [x] **Pattern Detection Engine:** Implemented `detectRecurring` logic in `ForecastingService` to automatically
  identify monthly recurring transactions (rent, salary, subscriptions) based on 3+ occurrences and amount stability.
* [x] **Forecasting Service:** Core domain service to generate `CashFlowForecast` entries for future dates (up to 12
  months) including projected balances and expected cash flow.
* [x] **Variable Spending (Burn Rate):** Implemented monthly burn rate logic for categories like "Groceries" to provide
  realistic residual budget forecasts.
* [x] **Bonus Projections:** Integration with `PayslipRepository` to project seasonal bonuses based on historical data.
* [x] **Stable UI:** Implemented deterministic UUIDs for predicted transactions to ensure a flicker-free, stable user
  experience.
* [x] **API Integration:** Enhanced `GET /api/v1/transactions/` to optionally include predictions and added
  `GET /api/v1/transactions/forecast/` for granular time-series data.
* [x] **Forecast Fine-Tuning:** Users can manually exclude specific future projections and mark historical transactions
  to be ignored by the engine for unmatched accuracy.
* [ ] **Manual Expected Transactions:** Capability for users to manually add one-time future expenses (e.g., vacation,
  taxes) for accurate planning.

<!-- end list -->