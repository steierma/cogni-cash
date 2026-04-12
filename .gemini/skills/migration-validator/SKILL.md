---
name: migration-validator
description: Validates and enforces SQL migration standards for CogniCash. Use when adding or modifying `.sql` files in `backend/migrations/`.
---

# Migration Validator

Enforce idempotent, raw SQL standards for database schema changes.

## Rules
1. **Idempotency:** ALWAYS use `IF NOT EXISTS` for `CREATE TABLE`, `CREATE INDEX`, `CREATE EXTENSION`, and `ON CONFLICT` for `INSERT`.
2. **Single-Action:** Do NOT include "Down" migrations in the same file.
3. **Raw SQL:** Do not use tool-specific annotations (e.g., `-- +goose Up`).
4. **No Conflicting Changes:** Do not drop columns and add them back in the same migration.

## Workflows

### 1. New Migration Creation
- Name the file with a three-digit prefix (e.g., `015_my_new_feature.sql`).
- Ensure every statement is idempotent.
- Add comments explaining the purpose of the change.
- Run the validator: `python3 scripts/validate_sql.py`.

### 2. Documentation Sync
After adding a migration:
1. Update `DATABASE_SCHEMA.md` to reflect the new state.
2. Update `backend/balance/dummy-data.sql` if the schema change affects existing test data.
3. Verify the migration by running `go test ./internal/adapter/repository/postgres/...` (which executes migrations on a test DB).

## Audit
When reviewing a migration:
1. Check for `SERIAL` types (use `UUID PRIMARY KEY DEFAULT gen_random_uuid()` instead).
2. Check for missing indexes on foreign keys.
3. Check for proper `ON DELETE` constraints (usually `CASCADE` or `SET NULL`).
