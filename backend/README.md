# CogniCash Backend

The core API and processing engine for CogniCash, built with Go 1.26 and PostgreSQL 16.

## 🏗️ Architecture

Following **Strict Hexagonal Architecture**:
- **Core Domain:** `internal/domain/entity/` (Data models) and `internal/domain/service/` (Business logic).
- **Ports:** `internal/domain/port/` (Interfaces for persistence, parsing, and LLM).
- **Adapters:** `internal/adapter/` (Implementation of ports: `postgres`, `http`, `parser`, `ollama`).

## 🚀 Key Features

- **Multi-Parser:** Native support for ING (PDF/CSV) and Amazon Visa (XLS).
- **AI Integration:** Local LLM support via Ollama for document categorization and transaction analysis.
- **Reconciliation:** 1:1 transaction matching to prevent double-counting of internal transfers.
- **Deduplication:** SHA-256 content hashing for all imported documents.
- **Bank Sync:** Live API integration via Enable Banking.
- **Automation:** Background workers for directory watching, auto-categorization, and bank syncing.

## 🛠️ Development

### Prerequisites
- Go 1.26+
- PostgreSQL 16
- Ollama (optional, for AI features)

### Running Locally
```bash
cp .env.example .env
# Update environment variables
go run main.go
```

### Database Migrations
Migrations are applied automatically by the `cmd/migrate` tool.
```bash
go run cmd/migrate/main.go
```

### Testing
```bash
go test ./...
```

## 🔐 Security

- **JWT Authentication:** All API routes (except `/health` and `/login`) require a valid token.
- **RBAC:** Admin-only access for user management and system configuration.
- **Password Recovery:** Secure, hashed, time-bound recovery tokens via SMTP.
- **CORS:** Configurable origins with dynamic reflection for development.
