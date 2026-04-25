# Installation Guide

CogniCash is designed to be easy to deploy. The recommended method is using **Docker**.

## ⚡ Option 1: Standalone Docker (Quickest)
The fastest way to run CogniCash. No repository clone or local compilation required.

1. **Prepare Environment**
   ```bash
   mkdir cognicash && cd cognicash
   # Download the standalone compose file and env example
   curl -O https://raw.githubusercontent.com/steierma/cogni-cash/main/docker-compose.standalone.yml
   curl -O https://raw.githubusercontent.com/steierma/cogni-cash/main/backend/.env.example
   mv docker-compose.standalone.yml docker-compose.yml
   cp .env.example .env
   ```
2. **Configure**
   - Edit `.env` and set `JWT_SECRET` and `POSTGRES_PASSWORD`.
   - Set `OLLAMA_URL` (e.g., `http://host.docker.internal:11434` for Mac/Windows).
3. **Run**
   ```bash
   docker compose up -d
   ```
   Access at **http://localhost**. (Default: `admin` / `admin`)

---

## ⚠️ Upgrading to v2.0.0 (From v1.5.x or older)
In version 2.0.0, we consolidated the historical database schema. If you are upgrading an existing instance, you MUST sync your migration history before running `make up`:
```bash
make db-squash-upgrade
```

---

## Option 2: Docker from Source
Recommended if you want to contribute or use the very latest development changes.

1. `git clone https://github.com/steierma/cogni-cash.git && cd cogni-cash`
2. `cp backend/.env.example backend/.env` (and configure it)
3. `make build`
4. `make up`

---

## Option 3: Remote Linux Server
Use our automated script for a production-ready setup with random passwords and CI/CD preparation.
```bash
sudo bash scripts/setup-server.sh
```

---

## Option 4: Local Development (No Containers)
For developers working on the Go/React codebase directly.
1. `make db-migrate`
2. `make backend-run`
3. `make frontend-dev`

---

## Configuration Reference
All variables live in `backend/.env` (or `.env` in standalone mode).

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | | Secret key for JWT signing. |
| `ADMIN_PASSWORD` | | Initial password for the 'admin' user. |
| `POSTGRES_PASSWORD` | | Password for the database user. |
| `OLLAMA_URL` | `http://localhost:11434` | URL of your local Ollama instance. |
| `DATABASE_HOST` | `postgres` | Use `postgres` for Docker, `localhost` for local dev. |

---

## Database Management
```bash
make db-migrate    # Apply pending migrations
make db-reset      # WIPE everything and re-migrate
make db-shell      # Open psql shell
make db-dump-local # Create a backup
```

---

## Live Banking (Enable Banking)
1. Register at [enablebanking.com](https://enablebanking.com).
2. Set `ENABLE_BANKING_APP_ID` in `.env`.
3. Place your RSA key as `enable-banking-prod.pem` in the root directory.
4. Restart the application.
