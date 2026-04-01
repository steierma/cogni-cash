# Gemini CLI Foundational Mandates: CogniCash

These instructions are foundational mandates for Gemini CLI. They take absolute precedence over general workflows and tool defaults.

## 1. Architectural Integrity (Hexagonal)
- **Strict Hexagonal Architecture:** Maintain a clean separation between Core Domain, Ports (interfaces), and Adapters (infrastructure).
- **Zero Dependencies in Core:** The core domain logic must have NO dependencies on external frameworks, databases, or HTTP clients.
- **Port-Based Communication:** Driving-side adapters (e.g., HTTP handlers) must depend exclusively on use-case interfaces defined in `internal/domain/port/use_cases.go`, never on concrete service structs.
- **Centralized Errors:** All sentinel errors are centralized in `internal/domain/entity/errors.go`. Services re-export them; adapters import them from the `entity` package only.

## 2. Development Methodology (TDD)
- **TDD is Mandatory:** No feature code is written without a failing test first.
- **Mock-Based Domain Testing:** Test domain logic using mocks to ensure zero reliance on DB or network.
- **Black-Box Testing:** Test files should use `package service_test` conventions to ensure they only use exported APIs.
- **Test Isolation:** Use unique IDs, far-future dates (e.g., year 2099), or specific `content_hash` lookups to prevent data pollution between parallel test runs.

## 3. Backend Implementation & Migrations
- **Persistence:** PostgreSQL 16 is the strict source of truth. Use `pgx/v5`.
- **In-Memory Mode:** Support `DB_TYPE=memory` for zero-dependency development. All repositories must have a thread-safe in-memory implementation with FIFO eviction policies to prevent memory leaks.
- **Demo Gating:** Mock banking providers must only be active when `DEMO_MODE=true` is explicitly set.
- **Migration Constraints:**
  - **Idempotency:** All migration files must use `IF NOT EXISTS` or `ON CONFLICT` guards.
  - **Raw SQL Execution:** Integration tests execute migrations as raw SQL. Do NOT use tool-specific annotations (e.g., `-- +goose Up`).
  - **Single-Action Files:** Do not include "Down" migrations or conflicting changes in a single file.
- **Secrets:** Use `.env` files; never hardcode credentials.

## 4. Frontend i18n Coding Standard
- **MANDATORY:** All user-visible strings must use `react-i18next`.
- **Library Stack:** `i18next`, `react-i18next`, `i18next-browser-languagedetector`.
- **Catalogue Updates:** Any new key must be added to `en` (Source of Truth), `de`, `es`, and `fr` in `frontend/src/i18n/locales/`.
- **Canonical Pattern:**
    ```tsx
    import { useTranslation } from 'react-i18next';
    const { t } = useTranslation();
    // Use: t('page.key')
    ```
- **Naming Convention:** Page-first, dot-separated hierarchy (e.g., `common.*`, `invoices.*`).
- **Formatting:** Use `fmtCurrency` and `fmtDate` from `frontend/src/utils/formatters.ts`. These must respect `i18n.language` for locale-aware output.
- **Persistence:** UI language is stored in the `settings` table as `ui_language` and applied via `i18n.changeLanguage`.

## 5. Mobile Development (Flutter)
- **Layered Integrity:** Strictly maintain `core/`, `data/`, `domain/`, and `features/` (UI/Controllers) layers.
- **Manual Entity Mapping:** Due to library instability, `fromJson` and `toJson` MUST be implemented manually. Avoid using code generators for entities until further notice.
- **Resilient Models:** Factories MUST handle `null` values from the API gracefully using default values (e.g., `?? 0.0`) to prevent runtime crashes.
- **Dio Provider:** All network calls MUST flow through the central `dioProvider` to inherit JWT and base URL configuration.
- **State Management:** Use Riverpod `StateNotifier` for all business logic.

## 6. Duplicate Detection
- **SHA-256 Content Hashing:** Mandatory for bank statements, transactions, payslips, and invoices to prevent re-importing identical payloads.

## 7. Security & Identity
- **Secure Password Resets:**
  - Tokens must be generated using CSPRNG (`crypto/rand`).
  - Tokens must be **hashed (SHA-256)** at rest.
  - Reset flow must be idempotent and short-lived (max 1 hour).
  - Use generic response messages for "Forgot Password" to prevent email enumeration.

## 8. Email & Notifications
- **SMTP-First:** All system-generated messages (welcome, reset, alerts) must use the standard SMTP adapter.
- **Async Delivery:** Outgoing emails must be sent asynchronously to prevent blocking the main request-response cycle.

## 9. Security & Infrastructure
- **Secret Management:** The application MUST NOT start with default or empty secrets for JWT or bank integrations.
- **Access Control:** All sensitive system configuration endpoints MUST be protected by strict RBAC (adminMiddleware).
- **Session & Infrastructure:** Prioritize session security (HttpOnly where possible) and container isolation (restricted port exposure).
- **Validation:** Use strict schemas and system-level boundaries for all untrusted document or LLM data.

## 10. Operational Maintenance & Memory
- **Local Reconnaissance:** During initialization, you MUST check if `LOCAL_SECRETS.md` exists and read its content. This file contains local-only configuration that is never committed to git.
- **Mandatory Synchronization:** After every significant change or feature completion, you MUST:
  1. Update **`MEMORY.md`** with the latest project state and completed tasks.
  2. Sync **`README.md`** to reflect new features or roadmap progress.
  3. Update **`docs/DATABASE_SCHEMA.md`** after any database migration.
  4. Update **`backend/balance/dummy-data.sql`** after schema changes to maintain test data integrity.
