# Gemini CLI Project Memory: Cogni-Cash

This file tracks persistent project state, maintenance requirements, and synchronization tasks. It must be updated by Gemini CLI regularly.

## 1. Post-Feature Maintenance
- [x] **i18n Cleanup**: Finalized translation files for en, de, es, fr. Added bank integration and provider settings.
- [x] **README.md Sync**: Updated "Target Architecture & Roadmap" and migration history.
- [x] **DATABASE_SCHEMA.md Sync**: Added `provider` column to `bank_connections` and migration `004`.
- [x] **Settings Service Sync**: Added `Get` method to `SettingsUseCase` to support granular retrieval.
- [x] **Transaction Review System**: Added `reviewed` field to transactions, "Unreviewed" default inbox, and "Review All" batch actions.
- [x] **AI Few-Shot Learning**: Enhanced auto-categorization by providing historical examples to the LLM (configurable via settings).
- [x] **Security Audit**: Verified SQL injection safety across the entire PostgreSQL repository layer.

## 2. Technical Debt & Roadmap
- **AI Few-Shot Learning (Phase 14)**: ✅
  - [x] Implement `GetCategorizationExamples` in repositories.
  - [x] Update `TransactionCategorizer` to support `{{EXAMPLES}}` placeholder.
  - [x] Add `auto_categorization_examples_per_category` setting.
  - [x] Detailed Ollama adapter logging for observability.
- **Transaction Review (Phase 13)**: ✅
  - [x] Database migration `007` for `reviewed` column.
  - [x] Backend repository & service support for review status.
  - [x] UI implementation: pulsing dots, "Review All" button, default filtering.
  - [x] i18n alignment across EN, DE, ES, FR.
- **In-Memory Dev Solution (Phase 12)**: ✅
  - [x] Create in-memory repository implementations in `internal/adapter/repository/memory/`.
  - [x] Add missing sentinel errors to `internal/domain/entity/errors.go`.
  - [x] Update `main.go` to support `DB_TYPE` environment variable for switching between `postgres` and `memory`.
  - [x] Refactor PostgreSQL adapters to use port interfaces instead of concrete types for cross-adapter dependencies.
- **Configurable Bank Sync History**: ✅
  - [x] Update `BankProvider` port interface to accept `dateFrom` and `dateTo` parameters.
  - [x] Update `EnableBanking`, `GoCardless`, `Dynamic`, and `Mock` adapters to support the new date parameters and pass them to the respective external APIs.
  - [x] Integrate `SettingsUseCase` into `BankService` to dynamically retrieve the `bank_sync_history_days` setting.
  - [x] Add inline settings control for `bank_sync_history_days` directly to the `BankConnectionsPage.tsx` UI.

## 3. Active Context & State
- **Current Phase:** Enhancing Bank Integration Features ✅
- **Last Database Migration:** `004_add_bank_provider.sql` (Note: In-memory mode does not use migrations)
- **Recent Significant Changes:**
  - Implemented a complete in-memory persistence layer for local development and testing.
  - Added `DB_TYPE` environment variable (defaults to `postgres`) to control the storage adapter.
  - Extended the bank synchronization engine to support configurable historical data fetching via the `bank_sync_history_days` setting.
  - Threaded date boundaries through the core domain down to the HTTP adapters, modifying Enable Banking and GoCardless query parameters.
  - Implemented inline settings mutation directly within the Bank Connections React page to adjust the sync window on the fly.