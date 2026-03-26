# AI Agents & System Architecture: Local AI Financial Manager

This document defines the roles, responsibilities, and technical constraints for the invoice & bank-statement management
application.

## 1. System Philosophy & Constraints

* **Architecture:** **Strict Hexagonal (Ports and Adapters)**. The core domain logic has zero dependencies on external
  frameworks, databases, or HTTP clients.
* **Development Methodology:** **Test-Driven Development (TDD)**. No feature code is written without a failing test
  first. Domain logic is tested with mocks — no database or network connection required.
* **Runtime Environment:** **Docker + Compose** for all services (Postgres, Backend, Frontend).
* **Persistence:** **PostgreSQL 16** (containerised via Docker Compose). *Note: In-memory fallback has been deprecated;
  a running database is strictly required.*
* **Secrets:** No credentials are hardcoded. All secrets live in a `.env` file (gitignored).
* **Internationalisation (i18n):** **All user-visible strings in the frontend must use `react-i18next`.**
  No hard-coded English (or any language) strings are permitted in `.tsx` / `.ts` UI files. Every new page,
  component, or modal must call `useTranslation()` and reference a key from the translation catalogues.
  Translation keys must be added to **both** `frontend/src/i18n/locales/en/translation.json` (English, source
  of truth) and `frontend/src/i18n/locales/de/translation.json` (German and all others) before a feature is considered
  complete. See **Section 7** for the full coding standard.

### 💡 Test Setup & Migration Constraints

To maintain a fast and reliable TDD cycle, the integration tests (`setup_test.go`) apply SQL migrations directly from
the `backend/migrations/` directory.

* **No Migration Tool Annotations:** The test runner executes migration files as raw SQL blocks via `pool.Exec`. It *
  *does not** parse tool-specific annotations like `-- +goose Up` or `-- +goose Down`.
* **Idempotency is Mandatory:** All migration files must use `IF NOT EXISTS` or `ON CONFLICT` guards.
* **Single-Action Files:** Avoid including "Down" migrations or multiple conflicting structural changes in a single
  file. If a file contains both an `ADD COLUMN` and a `DROP COLUMN` (even if commented with migration tags), the test
  runner will execute both, leaving the schema in an inconsistent state.
* **Isolation:** When testing repositories, use unique IDs, far-future dates (e.g., year 2099), or specific
  `content_hash` lookups to prevent data pollution between parallel test runs.

---

## 2. Hexagonal Structure (The "Ports")

### The Core (Domain)

* **Entities:** `Invoice`, `Category`, `BankStatement`, `Transaction`, `Reconciliation`, `JobState`, `Payslip`,
  `User`, `ReconciliationPairSuggestion`.
* **Domain Errors:** All sentinel errors (`ErrDuplicate`, `ErrPayslipDuplicate`, `ErrTransactionNotFound`,
  `ErrSameAccount`, `ErrEmptyRawText`, `ErrJobAlreadyRunning`, `ErrNothingToCategorize`,
  `ErrInvalidCredentials`) are centralised in `entity/errors.go`. Services re-export them for backward
  compatibility; adapters import them from the `entity` package only.
* **Domain Services:**
    * `CategorizationService` — sends raw text to the LLM and persists the resulting invoice.
    * `BankStatementService` — orchestrates file parsing, duplicate detection, category matching, and reconciliation
      logic (including marking statements as finished).
    * `PayslipService` — orchestrates HR document parsing, gross/net extraction, and bonus handling.
    * `UserService` — handles profile updates, secure password hashing, and account creation.
    * `AuthService` — manages JWT generation, login verification, and bootstrap admin seeding.
* **Repository Ports (Interfaces):** `InvoiceRepository`, `CategoryRepository`, `BankStatementRepository`,
  `ReconciliationRepository`, `PayslipRepository`, `UserRepository`, `SettingsRepository`.
* **External Ports (Interfaces):** `LLMClient`, `BankStatementParser`, `PayslipParser`, `JobTracker`.
* **Driving-side (Use-Case) Ports:** `AuthUseCase`, `UserUseCase`, `InvoiceCategorizationUseCase`,
  `BankStatementUseCase`, `TransactionUseCase`, `ReconciliationUseCase`, `SettingsUseCase`,
  `PayslipUseCase` — defined in `port/use_cases.go`. The HTTP adapter depends on these interfaces,
  never on concrete service structs.

### The Adapters (Infrastructure)

* **Input Adapters (Driving):**
    * REST API (`chi` v5) — `internal/adapter/http` with integrated JWT and RBAC middlewares.
    * Background directory watcher — auto-imports files from `IMPORT_DIR`.
    * Background JSON cron — polls `PAYSLIP_IMPORT_JSON_PATH` and bulk-imports structured payslip records from `payslips_import.json`.

* **Output Adapters (Driven):**
    * `PostgresRepository` — (pgx v5) serves as the strict, single source of truth for all structured data
      persistence.
    * `OllamaAdapter` — `internal/adapter/ollama` (LLMClient implementation).
    * `Parsers` — Specialized logic for ING (PDF/CSV), Amazon Visa (XLS), VW/CARIAD (Payslips), and AI fallback
      routines.

---

## 3. Agent & Component Definitions

### A. The Ingestion Agent (Driving Adapter)

* **Responsibility:** Triggers the import workflow via HTTP or the directory watcher.
* **Duplicate detection:** SHA-256 content hash prevents re-importing the same statement or payslip payload.

### B. The Reconciliation Agent

* **Responsibility:** Identifies matching transactions between Giro and Credit Card/Extra-Konto statements.
* **State Management:** Tracks `IsReconciled` status at both the **Transaction** level (settlement payments) and the *
  *Statement** level (marking a target statement as "Done").

### C. The Categorization Agent (Ollama / Llama 3)

* **Responsibility:** Classifies invoice and transaction text into categories, and optionally acts as a fallback parser
  for dynamic payslip formats.

### D. The Access Control Agent (RBAC Middleware)

* **Responsibility:** Secures API routes by validating JWT tokens and intercepting requests based on administrative
  roles. Ensures standard managers cannot modify peer users or escalate privileges.

---

## 4. Development Roadmap

### Phase 5: Credit Card & Statement Reconciliation ✅

> Prevents double-counting and tracks the completion status of target statements (Visa, Extra-Konto). Reconciled
> settlements are excluded from cashflow totals to maintain accurate Net Savings metrics.

* [x] **Migration:** `003_reconciliation.sql` — `statement_type` on statements; `is_reconciled` on
  transactions.
* [x] **Migration:** `007_extrakonto_reconciliation.sql` — generalized `target_statement_id`.
* [x] **Migration:** `008_statement_reconciliation_status.sql` — `is_reconciled` flag for `bank_statements`.
* [x] **Service Logic:** `ReconcileStatements` links individual Giro debits to target statements.
* [x] **Service Logic:** `FinishStatementReconciliation` marks an entire statement as reconciled.
* [x] **Analytics:** Repository-level filtering ensures reconciled settlement payments do not inflate
  `TotalExpense`.
* [x] **API:** `PATCH /api/v1/bank-statements/:id/reconcile` to finalize statement status.
* [x] **Frontend:** Reconciliation wizard, suggestion engine, and "Reconciled" status badges in the statement
  list.

### Phase 6: Async Batch Processing ✅

* [x] **Port:** `JobTracker` and `JobState` for async tracking.
* [x] **HTTP Adapter:** Expose `/start`, `/status`, and `/cancel` endpoints.
* [x] **Frontend:** Polling-based progress bar and mid-flight cancellation.

### Phase 7: User Management & RBAC ✅

* [x] **Migration:** `009_enhance_users_table.sql` — extends the `users` table with email, name, address, and role
  fields.
* [x] **Security:** Enforce `adminMiddleware` to protect user-mutation API endpoints.
* [x] **Service Logic:** `DeleteUser` and secure password hashing integration during account creation.
* [x] **Frontend:** Admin-only UI navigation filtering and interactive user-management table.

### Phase 8: Payslip JSON Bulk Import ✅

> Enables headless, zero-UI bulk ingestion of structured payslip data via a drop-zone JSON file — useful for
> migrating historical HR records without uploading individual PDFs.

* [x] **Port:** Extended `PayslipRepository` with `ExistsByOriginalFileName` for filename-based duplicate detection.
* [x] **Service Logic:** `PayslipService.ImportFromJSONFile` reads `payslips_import.json`, imports all entries, skips duplicates by `original_file_name`, **deletes each successfully imported PDF** from the same directory, and **keeps the JSON manifest** permanently. Logs a warning on per-record errors without aborting the batch.
* [x] **Background Cron:** A dedicated goroutine in `main.go` polls `PAYSLIP_IMPORT_JSON_PATH` on a configurable `PAYSLIP_IMPORT_INTERVAL` tick — same pattern as the bank-statement directory watcher.
* [x] **Docker Mount:** `./backend/payslips` is bind-mounted to `/app/payslips` in the backend container — drop a file on the host and the container picks it up with no rebuild.
* [x] **Configuration:** `PAYSLIP_IMPORT_JSON_PATH` and `PAYSLIP_IMPORT_INTERVAL` in `backend/.env` and `.env.example`.
* [x] **Sample file:** `backend/payslips/payslips_import.json` with realistic test entries ready to use.
* [x] **TDD:** Four unit tests covering no-op (file absent), happy path + file deletion, filename-dedup skip, and error-keeps-file.

### Phase 9: Multi-Language Frontend (i18n) 🚧

> Introduces `i18next` / `react-i18next` so the UI can be rendered in multiple languages. The active language
> is auto-detected from the browser and persisted per-user via the `ui_language` settings key.

* [ ] **Dependencies:** Install `i18next`, `react-i18next`, `i18next-browser-languagedetector`.
* [ ] **Bootstrap:** Create `frontend/src/i18n/index.ts`; side-effect import in `main.tsx`.
* [ ] **Catalogues:** Author `locales/en/translation.json` (source of truth), `locales/de/translation.json`, `locales/es/translation.json` and `locales/fr/translation.json`.
* [ ] **App wiring:** Read `ui_language` from settings in `App.tsx`; call `i18n.changeLanguage` on change.
* [ ] **Pages & components:** Replace all hard-coded strings with `t('...')` across all 18 affected files.
* [ ] **Settings UI:** Language selector dropdown in `SettingsPage.tsx`; persist via `PATCH /api/v1/settings/`.
* [ ] **Number/date formatting:** Update `frontend/src/utils/formatters.ts` to use `i18n.language` as locale for `Intl.NumberFormat` / `Intl.DateTimeFormat`.
* [ ] **Backend:** No changes required — `ui_language` is a regular settings key.

### Phase 10: Backend Hexagonal Architecture Hardening ✅

> Enforces strict hexagonal boundaries in the Go backend: the HTTP adapter (driving side) now depends
> exclusively on port interfaces — never on concrete service structs. Domain-level sentinel errors are
> centralised in the entity layer. All test files use black-box `package service_test` conventions.
> Domain service test coverage raised from 61 % to 81 %.

* [x] **Port:** `PayslipParser` moved from inline definition in `payslip_service.go` to `port/payslip_parser.go`.
* [x] **Port:** Eight driving-side use-case interfaces (`AuthUseCase`, `UserUseCase`,
  `InvoiceCategorizationUseCase`, `BankStatementUseCase`, `TransactionUseCase`,
  `ReconciliationUseCase`, `SettingsUseCase`, `PayslipUseCase`) added in `port/use_cases.go`.
* [x] **Entity:** `ReconciliationPairSuggestion` moved from `reconciliation_service.go` to
  `entity/reconciliation_suggestion.go`.
* [x] **Entity:** Centralised domain errors in `entity/errors.go`; services re-export via `var Err = entity.Err`.
  HTTP handlers now import `entity`, not `service`.
* [x] **Handler:** `Handler` struct fields changed from `*service.XxxService` to `port.XxxUseCase` interfaces.
  All production handler files no longer import the `service` package.
* [x] **Tests (black-box):** `payslip_service_test.go` converted from `package service` to
  `package service_test`. `JSONPayslipEntry` exported to support this.
* [x] **TDD — AuthService:** 15 unit tests (Login, ValidateToken, ChangePassword, EnsureAdminUser).
* [x] **TDD — UserService:** 13 unit tests (ListUsers, GetUser, CreateUser, UpdateUser, DeleteUser).
* [x] **TDD — SettingsService:** 5 unit tests (GetAll, UpdateMultiple, nil-logger safety).
* [x] **TDD — ReconciliationService:** 9 additional tests (DeleteReconciliation, SameAccountType,
  NilRepo, TargetNotFound, DefaultWindow, SkipsSameType, NilLogger).
* [x] **TDD — PayslipService:** 4 additional tests (Update success, Update duplicate hash, Delete
  success, Delete repo error).
* [x] **TDD — JobManager:** 5 tests (Cancel, DoubleStart, FinishThenRestart, CancelWhenIdle,
  TransactionService.CancelJob).

---

## 5. Environment & Secrets

| Variable                     | Description                                                                            |
|------------------------------|----------------------------------------------------------------------------------------|
| `JWT_SECRET`                 | Secret key for signing JWTs.                                                 |
| `ADMIN_USERNAME`             | The initial username seeded into the DB on startup.                                    |
| `ADMIN_PASSWORD`             | The initial password seeded into the DB on startup.                          |
| `POSTGRES_PASSWORD`          | Database password.                                                           |
| `OLLAMA_URL`                 | Ollama API base URL (default `http://localhost:11434`).                      |
| `PAYSLIP_IMPORT_JSON_PATH`   | Absolute path to `payslips_import.json` manifest. Worker imports all entries, deletes each imported PDF, and keeps the JSON. Leave empty to disable. |
| `PAYSLIP_IMPORT_INTERVAL`    | Polling interval for the payslip JSON cron (Go duration, e.g. `1m`, `1h`). Default `1h`. |

---

## 6. Local Development

```bash
# Start full stack with real DB dependency
make db-migrate     # Applies migrations (lexicographical order)
make backend-run    # Boots standard REST server
```

---

## 7. Frontend i18n Coding Standard

> **Mandatory for all new and modified frontend code.**
> `InvoicesPage.tsx` is the canonical reference implementation — follow it exactly.

### 7.1 Library stack

| Package | Role |
|---|---|
| `i18next` | Core translation engine |
| `react-i18next` | React bindings (`useTranslation`, `<Trans>`) |
| `i18next-browser-languagedetector` | Auto-detects language from browser / `localStorage` |

### 7.2 Bootstrap (done once)

`frontend/src/i18n/index.ts` initialises i18next and must be imported as a side-effect in `main.tsx` **before**
the React tree is rendered:

```ts
// frontend/src/i18n/index.ts
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import en from './locales/en/translation.json';
import de from './locales/de/translation.json';
import es from './locales/es/translation.json';
import fr from './locales/fr/translation.json';

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: { en: { translation: en }, de: { translation: de }, es: { translation: es }, fr: { translation: fr } },
    fallbackLng: 'en',
    interpolation: { escapeValue: false },
  });

export default i18n;
```

```ts
// frontend/src/main.tsx  — first import
import './i18n';
```

### 7.3 Using translations in components — the canonical pattern

Every page and component that renders user-visible text **must** follow this exact pattern (taken directly from
`InvoicesPage.tsx`):

```tsx
import { useTranslation } from 'react-i18next';

export default function MyPage() {
    const { t } = useTranslation();

    return (
        <div>
            <h1>{t('myPage.title')}</h1>
            <p>{t('myPage.subtitle')}</p>

            {/* Fallback / empty-state strings */}
            <span>{t('myPage.emptyState')}</span>

            {/* Inline conditional strings */}
            <td>{item.vendor?.name || t('myPage.unknownVendor')}</td>
            <td>{item.description || t('myPage.emptyDescription')}</td>
        </div>
    );
}
```

**Rules:**
1. `useTranslation()` is called **once per component** at the top of the function body.
2. Every user-visible string — titles, labels, table headers, placeholder text, empty states, tooltips,
   button text, error messages — must be wrapped in `t('...')`.
3. **Never** pass a raw string literal as JSX text or as a `title`/`placeholder`/`aria-label` prop.
4. Dynamic values (numbers, dates, names) are passed as i18next interpolation variables:
   `t('myPage.itemCount', { count: items.length })` — never built by string concatenation.

### 7.4 Translation key naming convention

Use a **page-first, dot-separated** hierarchy:

```
<page | component>.<section?>.<element>
```

| Prefix | Used for |
|---|---|
| `common.*` | Shared labels reused across many pages (`save`, `cancel`, `delete`, `loading`, `error`) |
| `nav.*` | Sidebar / top-bar navigation labels |
| `dashboard.*` | Dashboard page strings |
| `transactions.*` | Transactions page strings |
| `bankStatements.*` | Bank Statements page strings |
| `payslips.*` | Payslips page strings |
| `invoices.*` | Invoices page strings |
| `categories.*` | Categories page strings |
| `reconcile.*` | Reconciliation wizard strings |
| `users.*` | User management page strings |
| `settings.*` | Settings page strings |
| `login.*` | Login page strings |

**Example catalogue shape** (`locales/en/translation.json`):

```json
{
  "common": {
    "save": "Save",
    "cancel": "Cancel",
    "delete": "Delete",
    "edit": "Edit",
    "loading": "Loading…",
    "error": "An error occurred"
  },
  "invoices": {
    "title": "Invoices",
    "subtitle": "Documents processed by the LLM categorization engine.",
    "vendor": "Vendor",
    "category": "Category",
    "date": "Date",
    "description": "Description",
    "amount": "Amount",
    "actions": "Actions",
    "unknownVendor": "Unknown",
    "emptyDescription": "—",
    "noInvoices": "No invoices found."
  }
}
```

The German, Spain and French catalogue (`locales/de/translation.json`, `locales/es/translation.json`, `locales/fr/translation.json`) must mirror every key exactly.

### 7.5 Number & date formatting

`i18next` does **not** localise numbers or dates. Use the browser-native `Intl` API via the helpers in
`frontend/src/utils/formatters.ts`, which must respect `i18n.language` as the active locale:

```ts
import i18n from '../i18n';

// Currency — locale-aware decimal separator and symbol placement
export const fmtCurrency = (v: number, currency: string, locale = i18n.language) =>
    new Intl.NumberFormat(locale, { style: 'currency', currency }).format(v ?? 0);

// Date — locale-aware month/day order
export const fmtDate = (iso: string, style: 'short' | 'long' = 'short', locale = i18n.language) =>
    new Intl.DateTimeFormat(locale, { dateStyle: style }).format(new Date(iso));
```

### 7.6 Language persistence

The active language is stored in the database as the `ui_language` settings key (type `string`, e.g. `"en"`,
`"de"`). It is read once on app startup in `App.tsx` and applied via `i18n.changeLanguage(lang)`. Users can
change it in `SettingsPage.tsx` via a dropdown, which simultaneously calls `i18n.changeLanguage` (instant
live-switch) and persists via `PATCH /api/v1/settings/`.

No backend changes are required — `ui_language` is stored as a regular key in the existing `settings` table.

### 7.7 Checklist for every new frontend file

Before marking a frontend task as done, verify:

- [ ] `useTranslation()` is imported from `react-i18next` and called at the top of the component.
- [ ] Zero hard-coded user-visible strings remain in the JSX or props.
- [ ] All new translation keys are added to `locales/en/translation.json`.
- [ ] All new translation keys are mirrored in `locales/de/translation.json`.
- [ ] Any new number or date output uses `fmtCurrency` / `fmtDate` from `formatters.ts` (not raw `Intl` calls).
- [ ] Dynamic values use i18next interpolation syntax, not string concatenation.