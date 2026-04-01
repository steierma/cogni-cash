# AI Codebase Consistency Review Plan & Checklist

## Objective
Provide a robust framework for an AI to perform a holistic review of the Cogni-Cash codebase after any structural change (e.g., adding a database column, changing an interface). The goal is to ensure that changes propagate through all layers of the Hexagonal Architecture and that the system remains stable, tested, and documented.

---

## 1. Database & Persistence Layer
When a database change is introduced:
- [ ] **Migrations**: Verify the new migration file in `backend/migrations/` is idempotent (`IF NOT EXISTS`).
- [ ] **Postgres Adapter**: Ensure `backend/internal/adapter/repository/postgres/` implementations are updated with the new column in `SELECT`, `INSERT`, and `UPDATE` queries.
- [ ] **Memory Adapter**: Update the thread-safe in-memory implementation in `backend/internal/adapter/repository/memory/` to maintain parity with Postgres.
- [ ] **Dummy Data**: Update `backend/balance/dummy-data.sql` to include the new column for development/test environments.
- [ ] **Schema Documentation**: Reflect the change in `docs/DATABASE_SCHEMA.md`.

## 2. Domain & Port Layer (Core)
- [ ] **Entities**: Add the new field to the relevant struct in `backend/internal/domain/entity/`.
- [ ] **Ports**: If the change affects method signatures, update the interfaces in `backend/internal/domain/port/`.
- [ ] **Error Handling**: If the change introduces new failure modes, add sentinel errors to `backend/internal/domain/entity/errors.go`.

## 3. Application & Service Layer
- [ ] **Services**: Update business logic in `backend/internal/domain/service/` to handle the new field/parameter.
- [ ] **Use Cases**: Ensure the service still correctly implements the ports defined in the previous step.
- [ ] **Validation**: Add validation logic for the new field if it has constraints (e.g., non-empty, specific format).

## 4. Driving Adapters (HTTP / API)
- [ ] **DTOs/Handlers**: Update the JSON response/request structures in `backend/internal/adapter/http/`.
- [ ] **Mappers**: If the service entity differs from the HTTP response, update the mapping logic.
- [ ] **Status Codes**: Ensure appropriate HTTP status codes are returned for new validation errors.

## 5. Frontend (React)
- [ ] **API Types**: Update `frontend/src/api/types.ts` to mirror the backend changes.
- [ ] **API Client**: Ensure `frontend/src/api/client.ts` correctly handles the new data.
- [ ] **UI Components**: Update relevant components to display or input the new field.
- [ ] **i18n**: If new labels are needed, update `frontend/src/i18n/locales/` for EN, DE, ES, and FR.
- [ ] **Formatters**: Use `frontend/src/utils/formatters.ts` for any new currency or date fields.

## 6. Mobile (Flutter)
- [ ] **Entities/Models**: Manually update `fromJson` and `toJson` in `mobile/lib/domain/` or `data/` layers.
- [ ] **Null Safety**: Ensure the model handles potential `null` values from the API with defaults (e.g., `?? 0.0`).
- [ ] **UI/State**: Update Riverpod `StateNotifier` and UI widgets to reflect the change.

## 7. Mocks, Testing & Validation
- [ ] **Interface Mocks**: 
    - Identify all interfaces affected in `backend/internal/domain/port/`.
    - Regenerate or manually update mocks used in `*_test.go` files.
- [ ] **Unit Tests**:
    - Add a failing test case for the new functionality/field in the service layer.
    - Verify that service tests use `package service_test` (Black-Box testing).
- [ ] **Code Coverage**:
    - Run `go test ./... -coverprofile=coverage.out && grep -vE 'internal/domain/entity|internal/domain/port|cmd/|migrations/|main\.go' coverage.out > coverage_filtered.out && go tool cover -func=coverage_filtered.out` in `backend/`.
    - Ensure the new code is covered. The CI pipeline currently enforces a minimum **overall coverage of 35%** (excluding non-logic files like entities and ports), but the long-term target is **higher than 75%**.
    - The CI will fail if the coverage drops below the enforced 35% threshold or if the `domain/service` layer coverage drops below **65%**.
- [ ] **Integration Tests**:
    - Update repository tests to verify the new column is correctly persisted and retrieved from the DB.
- [ ] **SHA-256 Hashing**: If the field affects data integrity (e.g., in a transaction or bank statement), ensure the `content_hash` calculation is updated and verified.
- [ ] **Test Isolation**: Ensure new tests use unique IDs or far-future dates to avoid collisions.

---

## 8. Final Synchronization
- [ ] **MEMORY.md**: Update the project state and completed tasks.
- [ ] **README.md**: Sync features and roadmap if applicable.
- [ ] **Local Secrets**: Check if `LOCAL_SECRETS.md` needs any new configuration keys.
