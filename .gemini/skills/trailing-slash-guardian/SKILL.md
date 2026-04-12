---
name: trailing-slash-guardian
description: Enforces the architectural mandate that all API routes and client calls MUST have a trailing slash. Use when modifying backend handlers, frontend API clients, or mobile repositories.
---

# Trailing Slash Guardian

This skill ensures compliance with the CogniCash architectural mandate: **"All API routes MUST have a trailing slash."**

## Mandate
Every API endpoint definition in the backend and every API call in the frontend/mobile apps must terminate with a `/`.

- ✅ `/api/v1/transactions/`
- ❌ `/api/v1/transactions`

## Workflow

1.  **Verification**: After making changes to any API-related code, run the bundled validation script to check for violations.
    ```bash
    bash .gemini/skills/trailing-slash-guardian/scripts/check.sh
    ```

2.  **Surgical Fixes**: If violations are found, apply surgical updates to the affected files:
    - **Backend**: Update `backend/internal/adapter/http/handler.go`.
    - **Frontend**: Update `frontend/src/api/client.ts`.
    - **Mobile**: Update files in `mobile/lib/data/repositories/`.

3.  **Pattern Matching**:
    - Backend: `r.Get("/path/", ...)`
    - Frontend: `api.get('path/')`
    - Mobile: `_dio.get('path/')`

## Exceptions
- The `/health` endpoint (used for infrastructure checks).
- Root-level router group definitions like `/api/v1`.
- Dynamic parameters that are already followed by a slash: `api.get('users/${id}/')`.
