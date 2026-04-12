---
name: sync-memory
description: Automates the synchronization of project documentation (MEMORY.md, README.md, DATABASE_SCHEMA.md). Use after completing a feature, bug fix, or schema change.
---

# Sync Memory

Keep CogniCash documentation synchronized with the implementation state.

## Rules
1. **Holistic Sync:** After every task, update all three core documentation files: `MEMORY.md`, `README.md`, and `DATABASE_SCHEMA.md`.
2. **Persistence:** Never delete history from `MEMORY.md`. Only add new entries or mark existing ones as complete.
3. **Public Appeal:** Ensure `README.md` updates focus on value and visual impact.

## Workflows

### 1. Feature Completion
- Update **`MEMORY.md`**: Add the new feature to "Post-Feature Maintenance" with a brief technical summary.
- Update **`README.md`**: If it's a user-facing feature, add it to the roadmap or "Core Features" section.
- Sync **`DATABASE_SCHEMA.md`**: If a migration was added, append it to the "Migration History" and update the "Current Schema" tables.

### 2. Schema Maintenance
After a migration:
1. Copy the table definitions from the `.sql` file to `DATABASE_SCHEMA.md`.
2. Update the `backend/balance/dummy-data.sql` to include the new fields or tables.

## Synchronization Checklist
- [ ] Is **`MEMORY.md`** updated with latest project state?
- [ ] Is **`README.md`** synchronized with new features?
- [ ] Is **`DATABASE_SCHEMA.md`** up-to-date with recent migrations?
- [ ] Is **`backend/balance/dummy-data.sql`** aligned with the current schema?
