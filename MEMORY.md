# Gemini CLI Project Memory: CogniCash

This file tracks persistent project state, maintenance requirements, and synchronization tasks. It must be updated by Gemini CLI regularly.

## 1. Post-Feature Maintenance (Recent)
- [x] **Configurable Discovery Algorithm (User Preference)**: Introduced user-configurable parameters for the subscription discovery engine, including Date Tolerance (± X days), Amount Tolerance (%), Min Transactions for generic merchants, and Lookback Period (Years). Added a settings panel directly in the `SubscriptionsPage` for easy access. Supported across all 4 languages (EN, DE, ES, FR). Aligned `ForecastingService` and unit tests with the new date tolerance logic.
- [x] **Manual Subscription Creation from Transactions**: Implemented a workflow to allow users to manually create subscriptions directly from unlinked transactions on the `TransactionsPage`. Added a new backend endpoint `POST /api/v1/subscriptions/from-transaction/`, updated the frontend UI with a "Create Subscription" action and pre-filled modal, and provided full i18n support (EN, DE, ES, FR). (Vikunja Task ID 23).
- [x] **Enable Banking Mapping Enhancement**: Implemented mapping for `CounterpartyIban` and `BankTransactionCode` in the Enable Banking adapter by extracting Creditor/Debtor account details and transaction codes from the API response.
- [x] **LLM Adapter Robustness (Fix)**: Fixed a critical panic (nil pointer dereference) in the Ollama/Gemini adapter caused by ignoring errors from `http.NewRequestWithContext` and `json.Marshal`. Added proper error handling for all API request creation steps.
- [x] **Subscription Suggestion Robustness (Fix)**: Fixed a regression where partial IBAN population or "noise" (unrelated payments to the same IBAN) broke discovery sequences. Implemented Connected Components Grouping to merge Description and IBAN signals, Robust Multi-Sequence Detection to skip noise transactions, and mandatory deduplication to prevent duplicate merchant suggestions. Added stricter (3+) thresholds for generic payment processors (e.g., Stripe, First Data).
- [x] **Subscription Suggestion Consolidation (IBAN Grouping)**: Enhanced discovery logic to group transactions by `CounterpartyIban`, ensuring that subscriptions with varying descriptions (e.g., "Dauerauftrag" vs "Miete") are correctly consolidated into a single suggestion. Improved naming by preferring `CounterpartyName`.
- [x] **Mobile Parity for Subscription Management**: Implemented full-stack parity for the Subscription Management feature in the Flutter app, including manual entity mapping, repository implementation with trailing-slash compliance, and native UI for tracked/suggested subscriptions.
- [x] **Manual Editing of Extended Fields**: Expanded the subscription edit form to include Customer Number, Contact Details (Email, Phone, Website), Support/Cancellation URLs, Trial status, and Notes. (See Vikunja Task ID 21).
- [x] **Broad Historical Backfill**: Updated `ApproveSubscription` to link ALL historical transactions matching the normalized merchant name upon approval, ensuring accurate cumulative spend and history even if the discovery engine's strict pattern matching was broken by anomalies. (See Vikunja Task ID 19).
- [x] **Subscription Activity Log Fix**: Ensured the Activity Log (Aktivitätsprotokoll) is correctly populated by logging initial approval, AI data enrichment, and manual status changes.
- [x] **Subscription Analytics**: Display total cumulative spend for each subscription on its detail page. (See Vikunja Task ID 18).
- [x] **Suggestion Transparency**: Added a hover tooltip to the Discovery Inbox showing the underlying historical transactions (date/amount) for each suggestion. (See Vikunja Task ID 17).
- [x] **AI Enrichment Clarity & Navigation**: Improved user feedback for AI enrichment by adding a dedicated button on the detail page, showing a "Just Enriched!" success badge, and highlighting enriched metadata fields. (See Vikunja Task ID 11).
- [x] **Manual Subscription Deactivation**: Added the ability to manually change subscription status (Active/Canceled/Paused) from the Edit view, allowing users to track past subscriptions without impacting active spend analytics. (See Vikunja Task ID 16).
- [x] **Subscription Suggestion Stability**: Ensured a stable and consistent sort order for suggestions in the Discovery Inbox. (See Vikunja Task ID 15).
- [x] **Strict AI Discovery & Merchant Whitelisting**: Implemented AI source tracking for declined suggestions and a whitelist to prevent AI re-evaluation of false positives. Optimized costs by caching both positive and negative AI results. (See Vikunja Task ID 13).
- [x] **Frontend - Subscriptions Deletion**: Approved subscriptions can now be deleted without triggering cancellation. (See Vikunja Task ID 10).
- [x] **Manual Subscription Editing**: Users can now manually edit subscription details (Name, Amount, Frequency). (See Vikunja Task ID 12).
- [x] **Transaction History Fix**: Corrected linking of historical transactions for subscriptions. (See Vikunja Task ID 14).
- [x] **Configurable Subscription Discovery Lookback**: Added a setting to allow users to configure the discovery lookback period (default 3 years) to detect annual subscriptions. (See Vikunja Task ID 9).
- [x] **Undecline Subscription Suggestions**: Implemented backend logic and frontend UI to restore/undecline previously ignored subscription suggestions. (See Vikunja Task ID 8).
- [x] **Decline Subscription Suggestions**: Added ability to permanently decline/ignore suggested subscriptions with database tracking and frontend UI. (See Vikunja Task ID 7).
- [x] **Subscription Discovery & Back-filling**: Implemented backend logic for pattern recognition and retroactive linking. (See Vikunja Project: "CogniCash - Subscription Management" ID: 2).
- [x] **AI Merchant Profiling**: Implemented AI-driven enrichment for subscriptions, extracting contact details, billing info, and trial status. (See Vikunja Task ID 3).
- [x] **One-Click Cancellation & Dispatch**: Implemented AI draft generation, email dispatch via SMTP, and an audit trail for legal compliance. (See Vikunja Task ID 4).
- [x] **Frontend Dashboard & Discovery Inbox**: Created a dedicated Subscriptions page with AI discovery inbox, analytics, and tracked services list. Integrated summary widget into main Dashboard. (See Vikunja Task ID 5).
- [x] **Detail View & Cancellation UI**: Implemented a deep-dive view for subscriptions with payment history, contact details, and an AI-powered cancellation modal with email dispatch. (See Vikunja Task ID 6).
- [x] **Settings Service Hardening**: Implemented pgcrypto encryption and API masking for sensitive values.
- [x] **Document Vault Expansion**: Unified storage with AI classification and OCR search.
- [x] **Collaborative Finance**: Implemented shared categories and invoices with a joint dashboard.
- [x] **API Consistency**: Enforced trailing slashes globally and installed the `trailing-slash-guardian` skill.
- [x] **Bank Connection UX & Tenancy**: Enabled manager access to bank connections, implemented automatic transaction sync upon successful linking, and fixed a critical multi-tenancy bug by making the bank account unique constraint connection-aware.

*Note: For detailed feature history, see [docs/HISTORY.md](docs/HISTORY.md).*

## 2. Technical Debt & Roadmap
- **Vikunja Integration**: All implementation stories are now managed in [Vikunja](https://vikunja.steierl.org).
  - Main Project: `cogni-cash`
  - Current Feature Project: `CogniCash - Subscription Management`
- **Testing & Quality (High Priority)**:
  - [ ] **Frontend Test Suite**: Configure Vitest/Jest for React.
  - [ ] **Frontend Linting**: Resolve 30+ ESLint errors (primarily `no-explicit-any`).
  - [ ] **E2E Testing**: Implement Playwright/Cypress flows for critical paths.
- **Bank Integration (SimpleFIN) (High Priority)**:
  - [ ] **Backend Adapter**: Implement `internal/adapter/bank/simplefin`.
  - [ ] **Data Mapping & UI**: Token-based link flow in Bank Connections.
- **Mobile Roadmap**:
  - [ ] **Push Notifications**: FCM integration for alerts.
  - [ ] **Native Document Camera**: `google_ml_kit` for auto-cropping/edge detection.
  - [ ] **Batch Actions**: Long-press multi-select for transactions.

## 3. Active Context & State
- **Current Phase:** Subscription Management (Discovery Optimization & Manual Controls).
  - [x] **Manual Transaction Creation**: Enabled 'Create Subscription' directly from the transaction table.
  - [x] **Configurable Discovery**: Added user settings for date/amount tolerance and generic thresholds.
  - [x] **Discovery Consolidation**: Unified grouping logic using Connected Components (IBAN + Description).
  - [x] **Documentation Sync**: Synchronized README, DATABASE_SCHEMA, and MEMORY across all languages.
- **Story Management:** Implementation stories are exclusively managed in Vikunja. Local story files (e.g., `docs/stories/SUBSCRIPTION_STORIES.md`) have been removed.

- **Database Status:** v2.1.0 Optimized Schema (Consolidated Migration 005).
- **Skill Activation:** `trailing-slash-guardian`, `tenancy-auditor`, `i18n-guardian`, `migration-validator`, `sync-memory` are all active and enforced.

## 4. Critical Pending Tasks
*(No critical pending tasks)*
