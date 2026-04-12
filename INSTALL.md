Here is the fully updated and structurally corrected `INSTALL.md`.

I have added the **v2.0.0 upgrade instructions** to the top, updated the startup sequence explanation to reflect the new
`001_initial_schema.sql` file, added the new squash command to the database management section, and fixed a structural
error where "Option B" was duplicated in the middle of your original file.

```markdown
# Installation Guide

## Requirements

| Scenario                        | What you need                             |
|---------------------------------|-------------------------------------------|
| **Docker deploy** (recommended) | Docker Engine 24+ with the Compose plugin |
| **Local dev — full stack** | Go 1.26+, Node.js 22+, PostgreSQL, Ollama |

---

## ⚠️ Upgrading to v2.0.0 (From v1.5.x or older)

In version 2.0.0, we consolidated the historical database schema (migrations 001 through 017) into a single, clean
`001_initial_schema.sql` file.

If you are upgrading an existing instance from v1.5.0 or older, your database's internal tracking will conflict with
this new file structure. **Before running `make up` or deploying v2.0.0**, you MUST sync your migration history.

**Upgrade Steps for Existing Users:**

1. **Take a backup** (optional but highly recommended):
   ```bash
   make db-dump-local
   ```

2. **Run the migration history sync:**
   ```bash
   make db-squash-upgrade
   ```
3. **Start the application normally:**
   ```bash
   make up
   ```

*(Note: If you are installing CogniCash for the very first time, you can completely ignore this step.)*

---

## Option A — Docker (recommended for any server)

The only thing required on the host is Docker. No Go, no Node.js, no manual database setup.

> **How the database is provisioned automatically**
>
> On the very first `docker compose up`, the `postgres:16-alpine` image reads `POSTGRES_USER`, `POSTGRES_PASSWORD`, and
`POSTGRES_DB` from your `.env` and automatically creates the user, sets the password, and creates the database — no
> manual `CREATE USER` or `CREATE DATABASE` needed.
>
> The `migrate` container starts immediately after postgres is healthy and applies all versioned SQL files in
`backend/migrations/` in order. By the time the backend starts, the schema is fully ready.
>
> The full startup sequence is:
> ```text
> postgres  (auto-creates user + db)
>    ↓  healthy
> migrate   (applies 001_initial_schema.sql, then exits 0)
>    ↓  completed successfully
> backend   (connects, seeds admin user, serves API on :8080)
>    ↓  healthy
> frontend  (Nginx SPA + /api proxy on :3000)
> ```

### 1. Clone the repository

```bash
git clone [https://github.com/steierma/cogni-cash.git](https://github.com/steierma/cogni-cash.git)
cd cogni-cash
```

### 2. Create the environment file

```bash
cp backend/.env.example backend/.env
# Edit backend/.env — set the values below at minimum
```

`backend/.env` is the **single source of truth** for all credentials — loaded by `docker compose` (via
`env_file: backend/.env`) and by `make backend-run` / `make db-migrate`.

```dotenv
JWT_SECRET=<random-hex-string>             # openssl rand -hex 32
ADMIN_USERNAME=admin                       # login username
ADMIN_PASSWORD=<your-strong-password>      # login password
POSTGRES_PASSWORD=<your-strong-password>   # database password
OLLAMA_URL=[http://192.168.](http://192.168.)x.x:11434        # Use your host IP, NOT localhost
DOMAIN_NAME=localhost                      # Change if using public domain
```

> **Note on OLLAMA_URL**: When running inside Docker, `localhost` refers to the container itself. To reach an Ollama
> instance running on the host machine, you must use the host's actual network IP (e.g., `192.168.1.50`) or
`host.docker.internal` (on Docker for Mac/Windows).

`DATABASE_HOST` must stay `postgres` (the Docker service name) — do not change it.

### 3. Build the images

If you are deploying for the first time or `docker compose pull` fails, you should build the images locally:

1. **Enable local builds:** Local builds are supported by default if you use the provided
   `docker-compose.override.yml` (already in the repo). Ensure you don't have an empty one overriding things.
2. **Build and start:**
   ```bash
   make build
   make up
   ```

### 4. Special Files (Optional)

If you plan to use **Enable Banking**, you must place your private key file in the root directory:

- **Filename:** `enable-banking-prod.pem`
- **Why:** The `docker-compose.yml` expects this file to exist for mounting into the backend container. If you don't use
  it, you can create an empty file: `touch enable-banking-prod.pem`.

| Image                 | Base                        | Size   |
|-----------------------|-----------------------------|--------|
| `cogni-cash-backend`  | `distroless/static:nonroot` | ~10 MB |
| `cogni-cash-frontend` | `nginx:1.27-alpine`         | ~21 MB |

### 5. Start the stack

```bash
make up
```

#### Using a specific version

By default, `make up` uses the `:latest` image. If you want to run a specific release version (e.g., v1.4.0):

```bash
TAG=v1.4.0 make up
```

### 6. Verify

```bash
make ps                            # all containers should show "healthy"
curl http://localhost:8080/health  # → {"status":"ok"}
```

Open `http://localhost:3000` (or your `DOMAIN_NAME`) in your browser to access the frontend.

### Useful Commands

```bash
make logs                # tail logs for all containers
make down                # stop everything (data volume is preserved)
make restart             # restart all containers
docker compose down -v   # stop + wipe the postgres volume (full reset)
```

---

## Local Network Usage (Bypassing SSL)

By default, the bundled Caddy reverse proxy tries to enable HTTPS. If you are accessing CogniCash via a local IP (e.g.,
`http://192.168.1.50`) on your home network, Caddy may cause `ERR_SSL_PROTOCOL_ERROR`.

To allow plain HTTP access over your local network:

1. Open `caddy/Caddyfile`.
2. Change the first line from `{$DOMAIN_NAME:localhost} {` to `:80 {`.
3. Restart the stack: `make restart`.

This will serve the application on port 80 without attempting to negotiate SSL certificates.

---

## Troubleshooting

### Bind Mount is a Directory instead of a File

If you see an error like `is a directory` or `permission denied` regarding a `.pem` or `.env` file mounted in Docker:

- **Cause:** Docker creates a directory if the source file is missing on the host during startup (common if the file is
  not tracked by Git or was deleted).
- **Fix:** 1. Stop the stack: `make down`.
    2. Remove the accidental directory on the host: `rm -rf enable-banking-prod.pem` (replace with your filename).
    3. Ensure the actual file exists on the host.
    4. Start the stack: `make up`.

---

## Option B — Deploy to a remote Linux server

### 1. Clone the repository on the server

```bash
git clone [https://github.com/steierma/cogni-cash.git](https://github.com/steierma/cogni-cash.git)
cd cogni-cash
```

### 2. Run the setup script (once only)

Run it **from inside the cloned repo** as root. It detects the repo directory and writes `backend/.env` directly there
with a randomly generated password.

```bash
sudo bash scripts/setup-server.sh
```

What it does:

- Installs Docker Engine + Compose plugin
- Creates a `deploy` user with Docker access
- Generates an SSH key pair for CI/CD
- **Writes `backend/.env` with a random 32-char PostgreSQL password**

You can override the Ollama URL before running:

```bash
sudo OLLAMA_URL=http://YOUR_OLLAMA_HOST:11434 bash scripts/setup-server.sh
```

Re-running the script is safe — it skips steps that are already done and **never overwrites an existing `backend/.env`
**.

### 3. Add Forgejo secrets (If using CI/CD)

Go to your Forgejo repository → **Settings → Secrets → Actions** and add:

| Secret           | Example value                                  |
|------------------|------------------------------------------------|
| `DEPLOY_HOST`    | `YOUR_SERVER_IP`                               |
| `DEPLOY_USER`    | `deploy`                                       |
| `DEPLOY_PATH`    | `/opt/cogni-cash`                              |
| `DEPLOY_SSH_KEY` | *(the private key printed by setup-server.sh)* |

### 4. Push to `main` — CI/CD runs automatically

```bash
git push origin main
```

The pipeline runs:

```text
[test]   go test ./...
[build]  docker build backend + frontend
[deploy] rsync files → copy images → docker load → docker compose up -d
```

### Manual deploy (without CI)

```bash
make deploy                                       # uses defaults in Makefile
make deploy DEPLOY_HOST=x.x.x.x                   # override host
```

---

## Option C — Local development (no containers)

### With a real database + Ollama

You must have a PostgreSQL instance running locally or accessible via your network.

```bash
# 1. Run migrations
make db-migrate

# 2. Start backend
make backend-run

# 3. Start frontend
make frontend-dev
```

> **`make backend-run` automatically sets `DATABASE_HOST=localhost` and resolves `PAYSLIP_IMPORT_JSON_PATH` to the
local `backend/payslips/` directory**, so the cron worker works the same way as in Docker without any extra
> configuration.

### Payslip JSON bulk import (Docker & local)

Drop a `payslips_import.json` file into `backend/payslips/` at any time:

```bash
cp my_payslips.json backend/payslips/payslips_import.json
```

The background worker (polling every `PAYSLIP_IMPORT_INTERVAL`, default `1m` in `.env`) will:

1. Read and parse the JSON manifest.
2. Skip entries whose `original_file_name` already exists in the database.
3. Persist all new entries.
4. **Delete the individual PDF file** (same directory as the JSON) for each successfully imported entry.
5. **Keep `payslips_import.json`** permanently — it acts as a manifest and can be extended with new entries at any time.
6. Log a warning (but continue) if a PDF file is missing from disk — the record is still imported.

See `backend/payslips/payslips_import.json` for a ready-to-use sample with three realistic entries.

---

## Environment variables

All variables live in `backend/.env`. Docker Compose reads it automatically via the `env_file` directive.

| Variable                   | Default                          | Description                                                                                                                                            |
|----------------------------|----------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| `SERVER_ADDR`              | `:8080`                          | Port the backend listens on                                                                                                                            |
| `JWT_SECRET`               | *(generated by setup-server.sh)* | Secret key used to sign JWTs — **never use the default in production**                                                                                 |
| `ADMIN_USERNAME`           | `admin`                          | Username for the initial admin login                                                                                                                   |
| `ADMIN_PASSWORD`           | *(generated by setup-server.sh)* | Password for the initial admin login — **this is your initial 'registration'.** Log in with these credentials to create more users. Seeded on startup. |
| `POSTGRES_DB`              | `invoice_db`                     | Database name                                                                                                                                          |
| `POSTGRES_USER`            | `invoice_manager`                | Database user                                                                                                                                          |
| `POSTGRES_PASSWORD`        | *(generated by setup-server.sh)* | Database password — **never hardcoded**                                                                                                                |
| `DATABASE_HOST`            | `postgres`                       | DB hostname — `postgres` inside Docker, IP for local dev                                                                                               |
| `DATABASE_PORT`            | `5432`                           | DB port                                                                                                                                                |
| `OLLAMA_URL`               | `http://localhost:11434`         | Ollama API base URL                                                                                                                                    |
| `IMPORT_DIR`               | *(empty)*                        | Directory watched for auto-import. Leave empty to disable                                                                                              |
| `IMPORT_INTERVAL`          | `1h`                             | Re-scan interval (Go duration: `30m`, `1h`, …)                                                                                                         |
| `PAYSLIP_IMPORT_JSON_PATH` | *(empty)*                        | Absolute path to a `payslips_import.json` manifest. Worker imports all entries, deletes each imported PDF, and keeps the JSON. Leave empty to disable. |
| `PAYSLIP_IMPORT_INTERVAL`  | `1h`                             | How often the worker checks for the JSON file (Go duration: `1m`, `1h`, …)                                                                             |
| `SMTP_HOST`                | *(empty)*                        | SMTP server hostname for email notifications.                                                                                                          |
| `SMTP_PORT`                | `587`                            | SMTP server port.                                                                                                                                      |
| `SMTP_USER`                | *(empty)*                        | SMTP authentication username.                                                                                                                          |
| `SMTP_PASSWORD`            | *(empty)*                        | SMTP authentication password.                                                                                                                          |
| `SMTP_FROM_EMAIL`          | `noreply@cognicash.local`        | Sender address for outgoing emails.                                                                                                                    |

---

## Live Banking Setup (Enable Banking)
To enable real-time bank synchronization, the administrator must register an application with Enable Banking.
1.  Register at [enablebanking.com](https://enablebanking.com).
2.  Provide your `ENABLE_BANKING_APP_ID` in `backend/.env`.
3.  Place your private RSA key in the root directory as `enable-banking-prod.pem`.
4.  **HTTPS Mandatory:** Ensure your instance is accessible via **HTTPS**. Enable Banking requires an encrypted redirect URL for all production bank links. For local-only setups, you can use a **`hosts` file** override (e.g. `cognicash.local`) on your computer as long as your browser can resolve the name and access your instance via HTTPS.
5.  Restart the application to enable the "Bank Connections" feature for all users.

---

## Database management

```bash
make db-squash-upgrade  # ⚠️ Safely transition migration history for v2.0.0 upgrade
make db-migrate         # apply pending migrations
make db-truncate        # wipe all rows, keep schema
make db-nuke            # drop + recreate all tables, re-migrate
make db-reset           # drop + recreate the entire database, re-migrate
make db-shell           # open a psql shell
```

---

## Running tests

The domain tests utilize isolated mocks, so no database or Ollama connection is needed to verify core logic.

```bash
make backend-test
```

```