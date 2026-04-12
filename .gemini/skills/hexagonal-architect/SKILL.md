---
name: hexagonal-architect
description: Audits and enforces Hexagonal Architecture in the CogniCash backend. Use when creating or modifying backend domain services, repositories, or ports.
---

# Hexagonal Architect

Enforce a strict separation between Core Domain, Ports, and Adapters.

## Rules
1. **Zero Dependencies in Core:** `internal/domain/entity` and `internal/domain/service` must NOT import external frameworks (e.g., `gin`, `pgx`), databases, or infrastructure adapters.
2. **Port-Based Communication:** Services depend on Port interfaces (`internal/domain/port`), never concrete adapter implementations.
3. **Driving-Side Isolation:** HTTP handlers (Driving Adapters) depend on `internal/domain/port/use_cases.go`, not on service structs directly.
4. **Sentinels:** Centralized errors in `internal/domain/entity/errors.go` are the only exported errors to use across layers.

## Workflows

### 1. New Service Creation
- Define Port interfaces first in `internal/domain/port/`.
- Implement Core Logic in `internal/domain/service/`.
- Use Mock-Based testing (`shared_test.go`) for domain verification.

### 2. Leak Audit
When adding a dependency to a Go file in the domain:
1. Scan imports. If any infrastructure-related package (database, networking, JSON-parsing for APIs) is present, REFACTOR into a Port.
2. Confirm that the service only knows about `entity` and `port` packages.

## Compliance
- All new logic requires a failing test before implementation (TDD).
- Integration tests must be in `service_test` package to enforce black-box testing.
