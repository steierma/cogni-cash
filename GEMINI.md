# Gemini CLI Foundational Mandates: CogniCash

These instructions are foundational mandates for Gemini CLI. They take absolute precedence over general workflows and tool defaults.

## 1. Architectural Integrity (Hexagonal)
- **Strict Hexagonal Architecture:** Maintain a clean separation between Core Domain, Ports (interfaces), and Adapters (infrastructure).
- **Zero Dependencies in Core:** Core domain must have NO dependencies on external frameworks, databases, or HTTP clients.
- **Port-Based Communication:** Driving-side adapters must depend exclusively on use-case interfaces (ports), never on concrete services.
- **Centralized Errors:** All sentinel errors are in `internal/domain/entity/errors.go`.

## 2. Development Methodology (TDD)
- **TDD is Mandatory:** No feature code is written without a failing test first.
- **Black-Box Testing:** Test domain logic using mocks and `package service_test`.
- **Validation:** Use `docs/AI_CONSISTENCY_CHECKLIST.md` for any structural or API change.

## 3. Project Navigation Guide
- `backend/`: Go implementation (Hexagonal Architecture).
  - `internal/domain/entity/`: Core data models.
  - `internal/domain/port/`: Use-case and repository interfaces.
  - `internal/domain/service/`: Business logic.
  - `internal/adapter/`: Infrastructure (HTTP, Postgres, LLM, SMTP).
- `frontend/`: React (TypeScript, Tailwind, i18next).
  - `src/api/`: API Client and types.
  - `src/i18n/locales/`: Translations (EN, DE, ES, FR).
- `mobile/`: Flutter (Clean Architecture, Riverpod, Isar).
  - `lib/domain/`: Entities and Use Cases.
  - `lib/data/`: Repositories and Isar Schemas.
  - `lib/presentation/`: UI features and State Management.
- `docs/`: Technical concepts and guides (History, API, DB, etc.).

## 4. Common Troubleshooting & Pitfalls
- **Trailing Slashes:** All API routes MUST have a trailing slash (e.g., `/api/v1/transactions/`).
- **Isar Mapping (Mobile):** Manual `fromJson`/`toJson` is mandatory for Flutter entities.
- **i18n Completeness:** Every new UI string MUST be added to all 4 languages (EN, DE, ES, FR).
- **CORS:** Backend supports `ALLOWED_ORIGINS=*` in dev, but check `backend/.env` for production.

## 5. Specialized Tools Reference
- **i18n Tool:** `python3 scripts/support/i18n_tool.py check` (Verifies translation completeness).
- **Isar Inspector:** `scripts/support/inspect_isar.sh` (Diagnoses mobile database exports).
- **DB Squash:** `make db-squash-upgrade` (Safely upgrades existing databases to v2.0.0+).
- **Dev Mode:** `DB_TYPE=memory` for zero-dependency backend development.

## 6. Operational Maintenance & Memory
- **Mandatory Synchronization:** After every significant change:
  1. Update **`MEMORY.md`** (keep it lean, move history to `docs/HISTORY.md`).
  2. Update **`DATABASE_SCHEMA.md`** after any database migration.
  3. Update **`README.md`** and **`INSTALL.md`** if features or setup steps change.
  4. Update **`backend/balance/dummy-data.sql`** to maintain test data integrity.
