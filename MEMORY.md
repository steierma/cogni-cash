# Gemini CLI Project Memory: CogniCash

This file tracks persistent project state, maintenance requirements, and synchronization tasks. It must be updated by Gemini CLI regularly.

## 1. Post-Feature Maintenance (Recent)
- [x] **URL State Synchronization (Deep Linking)**: All frontend filters, sorting, and UI toggles (Transactions, Analytics, Forecasting, Invoices) are now synchronized with URL query parameters for bookmarking and history support.
- [x] **Multi-token Authentication (JWT + Refresh Tokens)**: Implemented refresh token rotation, secure cookie handling (web), and revocation logic (Migration `003`).
- [x] **Soft-delete Support for Categories**: Categories can now be soft-deleted to preserve historical transaction integrity (Migration `003`).
- [x] **Database Optimizations**: Added missing multi-tenancy indexes and stricter schema constraints for color hex and currency codes (Migration `003`).
- [x] **Transaction Forecasting (Phase 2 - Advanced & Stable)**: Stable UI (UUID v3), extended range (1-12 months), Burn Rate logic, and hybrid exclusion strategy.
- [x] **Payslip-Assisted Transaction Decomposition**: Normalizes salary/bonus bundles (Migration `017`).
- [x] **Net Bonus Forecasting**: Realistic net payout projections using `netFactor`.
- [x] **Mobile Cash Flow Forecasting**: Advanced forecasting UI ported to Flutter with `fl_chart`.
- [x] **API Consistency Fix (Trailing Slashes)**: Resolved 404 errors for Forecast Pattern Exclusions and performed a global audit/fix of all API routes and client calls to enforce trailing slashes.
- [x] **Trailing Slash Guardian Skill**: Created and installed a new skill (`trailing-slash-guardian`) with an automated check script to prevent future regressions.
- [x] **Pattern-Level Forecast Exclusion**: Support for "muting" an entire recurring series (Migration `016`).
- [x] **Repository Sanitization**: Deep-scrubbed personal data and updated git persona.
- [x] **Vault Mode**: License-enforced read-only mode for Mobile.
- [x] **Mobile i18n Completeness**: Resolved missing translations and hardcoded strings in the Flutter forecasting view; added `fore.restorePatternConfirm` and enforced fallbacks to prevent raw keys (like `common.restore`) from appearing in the UI.
- [x] **Enhanced Mobile AI Configuration**: Ported the detailed AI configuration UI from Hermit to the main CogniCash mobile app. This includes a new `AIConfigView` with prompt template editors and "Insert Default" logic, which synchronizes directly with the backend settings (`llm_api_url`, `llm_prompts`, etc.).
- [x] **Mobile Biometric Authentication**: Integrated the `local_auth` logic from Hermit into the main CogniCash mobile app. This includes a global `LockScreen` wrapper that protects the app with Fingerprint/FaceID/PIN and a new "Biometric Lock" toggle in the Security settings.
- [x] **Android Biometric Compatibility**: Updated `MainActivity.kt` to inherit from `FlutterFragmentActivity` and added `USE_BIOMETRIC` permission to `AndroidManifest.xml` to resolve `PlatformException(no_fragment_activity)` errors.

*Note: For older feature history, see [docs/HISTORY.md](docs/HISTORY.md).*

## 2. Technical Debt & Roadmap
- **Testing & Quality (High Priority)**:
  - [ ] **Frontend Test Suite**: Configure Vitest/Jest for React.
  - [ ] **Frontend Linting**: Resolve 30+ ESLint errors (primarily `no-explicit-any`).
  - [ ] **E2E Testing**: Implement Playwright/Cypress flows for critical paths.
- **Bank Integration (SimpleFIN) (High Priority)**:
  - [ ] **Backend Adapter**: Implement `internal/adapter/bank/simplefin`.
  - [ ] **Data Mapping & UI**: Token-based link flow in Bank Connections.
- **Advanced Forecasting (Phase 3)**:
  - [ ] **Planned Transactions UI**: Implement React forms for manual future entries.
  - [ ] **Auto-Resolution Matching**: Fuzzy match imported Tx against pending planned Tx.
- **Mobile Roadmap**:
  - [ ] **Push Notifications**: FCM integration for alerts.
  - [ ] **Native Document Camera**: `google_ml_kit` for auto-cropping/edge detection.
  - [ ] **Batch Actions**: Long-press multi-select for transactions.

## 3. Active Context & State
- **Current Phase:** Security & Performance Hardening ✅
- **Database Status:** v2.1.0 Optimized Schema (`003_database_improvements.sql`).
- **Recent Significant Changes:**
  - Implemented JWT Refresh Token rotation and revocation.
  - Added soft-delete support for Categories.
  - Enforced ISO 4217 currency codes and Hex color constraints.
  - Added multi-tenancy indexes across all tables.
  - Fixed frontend Axios 401 interceptor build issues.

## 4. Critical Pending Tasks
- [ ] **Forgejo Cleanup:** Force-push the scrubbed git history to remote. (Command: `git push forgejo --force --all && git push forgejo --force --tags`)
