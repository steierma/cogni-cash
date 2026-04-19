# Unpushed Changes and Local Modifications

This document tracks changes in the `cogni-cash` repository that have not yet been pushed to the remote repository.

## Unpushed Commits

The following commits are on the local `main` branch but have not been pushed to `origin/main`:

### 1. `3632246` - feat(documents): enhance document vault with manual overrides and AI skip
**Date:** Wed Apr 15 22:41:24 2026 +0200
- **Backend:** Support for manual file names, document types, and AI parsing toggle in `DocumentService`.
- **Backend:** Unit tests for `DocumentService` and `DocumentRepository`.
- **Frontend:** Refactored API client into modular services (auth, bank, category, invoice, document, etc.).
- **Frontend:** Shared Document and Invoice form/modal components.
- **Frontend:** Enhanced `DocumentVaultPage` with quick upload, manual metadata editing, and preview.
- **i18n:** Updated DE, EN, ES, and FR translations.
- **Misc:** Updated `.gitignore` and version increment to 2.0.1.

### 2. `3bc8021` - feat(documents): implement document vault for tax certificates and contracts
**Date:** Wed Apr 15 16:07:40 2026 +0200
- Initial implementation of the document vault.

---

## Uncommitted Changes (Working Directory)

These changes are present in the working directory but have not been staged or committed yet.

### Frontend Refactoring
A large-scale refactoring is in progress, primarily focusing on splitting the monolithic `api/types.ts` into modular type files.

- **Deleted:** `frontend/src/api/types.ts`
- **Modified (Import Updates):**
  - Updated all components and pages to use granular imports from `src/api/types/` (e.g., `category.ts`, `transaction.ts`, `system.ts`, etc.) instead of the central `api/types.ts`.
  - Affected directories: `src/api/services/`, `src/components/`, `src/pages/`.
- **Modified (Service Refactor):**
  - `src/api/client.ts` and various services in `src/api/services/` have pending modifications related to this refactor.

### Backend Tests
- `backend/internal/adapter/repository/postgres/document_repository_test.go`: Added new test cases for duplicate hashes, search functionality, and filtering.
- `backend/internal/domain/service/document_service_test.go`: Updates to document service tests.

---

## Untracked Files

The following files are not tracked by git:

- `backend/document_repo_coverage.out`
- `backend/document_svc_coverage.out`

---
*Generated on: Wednesday, April 15, 2026*
