---
name: tenancy-auditor
description: Audits and enforces user-level data isolation (multi-tenancy) in CogniCash. Use when modifying backend repositories, services, or adding new database entities.
---

# Tenancy Auditor

Ensure that every piece of data is strictly isolated to the owning user.

## Rules
1. **Query Scoping:** EVERY `SELECT`, `UPDATE`, and `DELETE` query in the repository must include a `user_id = $n` filter.
2. **Entity Consistency:** All domain entities must include a `UserID uuid.UUID` field.
3. **Insert Validation:** When saving new data, ensure the `user_id` is populated from the authenticated context.
4. **Unique Constraints:** Table constraints must include `user_id` (e.g., `UNIQUE(name, user_id)`) to allow duplicate names across different users.

## Workflows

### 1. Repository Audit
When modifying a repository:
1. Run the auditor: `python3 scripts/check_tenancy.py`.
2. Check each SQL query manually if the script misses something.
2. Verify that the `FindByID`, `Update`, and `Delete` methods take `userID` as a parameter.

### 2. Service-Layer Propagation
1. Ensure the `UseCases` (Ports) include `userID` in their method signatures.
2. Verify that services pass the `userID` from the driving adapter (e.g., HTTP handler) down to the repository.

## Audit Checklist
- [ ] Does the SQL query filter by `user_id`?
- [ ] Does the database table have a `user_id` column?
- [ ] Does the domain entity have a `UserID` field?
- [ ] Are unique indexes multi-tenant aware?
