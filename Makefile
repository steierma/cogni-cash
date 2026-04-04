# =============================================================================
#  Local AI Financial Manager — Makefile
# =============================================================================

# ── Config ────────────────────────────────────────────────────────────────────
ENV_FILE      := backend/.env

-include $(ENV_FILE)

DB_HOST       ?= localhost
DB_PORT       ?= $(DATABASE_PORT)
DB_USER       ?= $(POSTGRES_USER)
DB_NAME       ?= $(POSTGRES_DB)
DB_PASSWORD   ?= $(POSTGRES_PASSWORD)

PSQL          := PGPASSWORD="$(DB_PASSWORD)" psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME)

BACKEND_DIR   := ./backend
FRONTEND_DIR  := ./frontend

TAG           ?= latest
BACKEND_IMAGE := cogni-cash-backend:$(TAG)
FRONTEND_IMAGE:= cogni-cash-frontend:$(TAG)

DEPLOY_HOST   ?= financial-manager
DEPLOY_PATH   ?= /opt/cogni-cash
DEPLOY_SSH    := ssh $(DEPLOY_HOST)
RSYNC_SSH     := ssh

.DEFAULT_GOAL := help

.PHONY: help \
	up down restart logs ps \
	build build-backend build-frontend \
	deploy deploy-sync deploy-up setup-server \
	db-truncate db-reset db-nuke db-migrate db-shell db-reset-password \
	dev backend-run backend-build backend-test \
	gen-testdata \
	frontend-dev frontend-dev-prod frontend-build \
	clean clean-all nuke

# ── Help ──────────────────────────────────────────────────────────────────────
help:
	@echo ""
	@echo "  Local AI Financial Manager"
	@echo ""
	@echo "  Infrastructure"
	@echo "    make up               Start all containers"
	@echo "    make down             Stop all containers"
	@echo "    make restart          Restart all containers"
	@echo "    make logs             Tail logs of all containers"
	@echo "    make ps               Show container status"
	@echo ""
	@echo "  Docker Images"
	@echo "    make build            Build both backend and frontend images"
	@echo "    make build-backend    Build only the backend image"
	@echo "    make build-frontend   Build only the frontend image"
	@echo ""
	@echo "  Deploy (manual, without CI)"
	@echo "    make deploy           Full deploy: build → rsync → pull → up on remote host"
	@echo "    make setup-server     Run the one-time server setup script on DEPLOY_HOST"
	@echo ""
	@echo "  Database"
	@echo "    make db-truncate      Wipe all rows, keep schema"
	@echo "    make db-reset         Drop + recreate DB, re-migrate"
	@echo "    make db-nuke          Drop all tables + re-migrate"
	@echo "    make db-migrate       Run pending SQL migrations"
	@echo "    make db-shell         Open psql shell (uses backend/.env)"
	@echo "    make db-dump-local    Create a compressed backup of your local database"
	@echo "    make db-reset-password USERNAME=<username> PASSWORD=<newpass>  Reset a user password"
	@echo ""
	@echo "  Backend"
	@echo "    make dev-memory       Start backend (In-Memory) + Frontend and open browser"
	@echo "    make dev              Local dev server — in-memory + seed data"
	@echo "    make backend-run      go run ./main.go  (reads backend/.env)"
	@echo "    make backend-build    Compile binary → backend/bin/server"
	@echo "    make backend-test     Run all Go tests"
	@echo "    make gen-testdata     Regenerate anonymised parser fixture files (PDF/CSV/XLS)"
	@echo ""
	@echo "  Frontend"
	@echo "    make frontend-dev          Start Vite dev server"
	@echo "    make frontend-dev-prod     Start Vite dev server proxying to prod backend"
	@echo "    make frontend-build        Production build"
	@echo ""
	@echo "  Cleanup"
	@echo "    make clean            Remove compiled Go binary"
	@echo "    make clean-all        Remove binary + frontend dist + node_modules"
	@echo "    make nuke             ⚠️  Stop containers + wipe volumes + remove images + clean build artifacts"
	@echo ""

# ── Infrastructure ────────────────────────────────────────────────────────────

backend/.env:
	@echo ""
	@echo "  ⚠️  backend/.env not found — creating from backend/.env.example"
	@cp $(BACKEND_DIR)/.env.example $(BACKEND_DIR)/.env
	@echo "  ✏️  Edit backend/.env and set POSTGRES_PASSWORD and OLLAMA_URL, then run make up again."
	@echo ""
	@exit 1

up: backend/.env
	docker compose up -d

down:
	docker compose down

restart:
	docker compose restart

SERVICE_ARG := $(word 2,$(MAKECMDGOALS))
SERVICE := $(if $(SERVICE_ARG),$(SERVICE_ARG),$(SERVICE))

ifneq (,$(filter cogni-cash-backend,$(SERVICE)))
SERVICE := backend
endif
ifneq (,$(filter cogni-cash-frontend,$(SERVICE)))
SERVICE := frontend
endif
ifneq (,$(filter cogni-cash-migrate,$(SERVICE)))
SERVICE := migrate
endif
ifneq (,$(filter cogni-cash-db,$(SERVICE)))
SERVICE := postgres
endif
ifneq (,$(filter cogni-cash-adminer,$(SERVICE)))
SERVICE := adminer
endif

.PHONY: logs backend frontend migrate postgres adminer \
        cogni-cash-backend cogni-cash-frontend cogni-cash-migrate cogni-cash-db cogni-cash-adminer
logs:
	@if [ -n "$(SERVICE)" ]; then \
		echo "Tailing logs for '$(SERVICE)' (docker compose)"; \
		docker compose logs -f $(SERVICE); \
	else \
		echo "Tailing logs for all services"; \
		docker compose logs -f; \
	fi

ps:
	docker compose ps

# ── Docker Image Builds ───────────────────────────────────────────────────────

build: build-backend build-frontend

build-backend:
	@echo ">>> Building backend image: $(BACKEND_IMAGE)"
	docker build -t $(BACKEND_IMAGE) $(BACKEND_DIR)
	@echo ">>> Done: $(BACKEND_IMAGE)"

build-frontend:
	@echo ">>> Building frontend image: $(FRONTEND_IMAGE)"
	docker build --build-arg VITE_ENABLE_SANDBOX=true -t $(FRONTEND_IMAGE) $(FRONTEND_DIR)
	@echo ">>> Done: $(FRONTEND_IMAGE)"

# ── Manual Deploy ─────────────────────────────────────────────────────────────

deploy: build deploy-sync deploy-up
	@echo ">>> Deploy complete. Check status with: make ps"

deploy-sync:
	@echo ">>> Syncing files to $(DEPLOY_HOST):$(DEPLOY_PATH)/"
	rsync -az --delete \
		-e "$(RSYNC_SSH)" \
		--exclude='.git' \
		--exclude='.env' \
		--exclude='docker-compose.override.yml' \
		--exclude='backend/balance/diba_history' \
		--exclude='backend/balance/bulk_import' \
		--exclude='backend/bin' \
		--exclude='frontend/node_modules' \
		--exclude='frontend/dist' \
		--exclude='mobile' \
		--exclude='standalone_mobile' \
		--exclude='tmp' \
		./ \
		$(DEPLOY_HOST):$(DEPLOY_PATH)/

deploy-up:
	@echo ">>> Starting stack on $(DEPLOY_HOST)..."
	$(DEPLOY_SSH) \
		"cd $(DEPLOY_PATH) && \
		 [ -f backend/.env ] || cp backend/.env.example backend/.env && \
		 export IMAGE_SUFFIX=-internal && \
		 docker compose pull && \
		 docker compose up -d --remove-orphans && \
		 docker compose ps"

setup-server:
	@echo ">>> Running server setup on $(DEPLOY_HOST)..."
	ssh root@$(DEPLOY_HOST) "bash -s" < scripts/setup-server.sh

# ── Database ──────────────────────────────────────────────────────────────────

db-truncate:
	@echo ">>> Truncating all tables except users (do this manually if needed) $(DB_HOST)…"
	$(PSQL) -c "TRUNCATE TABLE payslip_bonuses, payslips, settings, reconciliations, transactions, bank_statements, bank_accounts, bank_connections, invoices, categories RESTART IDENTITY CASCADE;"
	@echo ">>> Done."

db-reset:
	@echo ">>> Dropping database $(DB_NAME) on $(DB_HOST)…"
	PGPASSWORD="$(DB_PASSWORD)" psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres \
		-c "DROP DATABASE IF EXISTS $(DB_NAME);" \
		-c "CREATE DATABASE $(DB_NAME) OWNER $(DB_USER);"
	@echo ">>> Running migrations…"
	$(MAKE) db-migrate
	@echo ">>> Done."

db-nuke:
	@echo ">>> Dropping all tables on $(DB_HOST)/$(DB_NAME)…"
	$(PSQL) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@echo ">>> Re-running migrations…"
	$(MAKE) db-migrate
	@echo ">>> Done."

db-migrate:
	@echo ">>> Running migrations against $(DB_HOST)…"
	cd $(BACKEND_DIR) && export $$(grep -v '^#' .env 2>/dev/null | xargs) ; DATABASE_HOST=127.0.0.1 go run ./cmd/migrate
	@echo ">>> Migrations complete."

db-shell:
	$(PSQL)

db-reset-password:
	@[ -n "$(USERNAME)" ] || (echo "ERROR: USERNAME is required.  Usage: make db-reset-password USERNAME=admin PASSWORD=newpass"; exit 1)
	@[ -n "$(PASSWORD)" ] || (echo "ERROR: PASSWORD is required.  Usage: make db-reset-password USERNAME=admin PASSWORD=newpass"; exit 1)
	@if command -v go > /dev/null 2>&1; then \
		echo ">>> Using local go run..."; \
		cd $(BACKEND_DIR) && export $$(grep -v '^#' .env 2>/dev/null | xargs) ; \
		DATABASE_HOST=127.0.0.1 go run ./cmd/resetpw -user "$(USERNAME)" -password "$(PASSWORD)"; \
	else \
		echo ">>> go not found — using docker run..."; \
		export $$(grep -v '^#' $(BACKEND_DIR)/.env 2>/dev/null | xargs) ; \
		docker run --rm \
			--network cogni-cash_app-net \
			-e POSTGRES_USER -e POSTGRES_PASSWORD -e POSTGRES_DB \
			-e DATABASE_HOST=cogni-cash-db \
			-e DATABASE_PORT=5432 \
			--entrypoint /app/cogni-cash-resetpw \
			$(BACKEND_IMAGE) \
			-user "$(USERNAME)" -password "$(PASSWORD)"; \
	fi

# ── Backend ───────────────────────────────────────────────────────────────────
dev-memory:
	@echo ">>> Cleaning up existing processes on :8080 and :5173..."
	@lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@lsof -ti:5173 | xargs kill -9 2>/dev/null || true
	@echo ">>> Starting In-Memory Dev Stack..."
	@( \
		(cd $(BACKEND_DIR) && DB_TYPE=memory DEMO_MODE=true go run ./main.go) & \
		(cd $(FRONTEND_DIR) && npm run dev) & \
		(sleep 5 && \
		  (open http://localhost:5173 || xdg-open http://localhost:5173 || start http://localhost:5173) \
		) & \
		wait \
	)

dev:
	cd $(BACKEND_DIR) && go run ./cmd/dev

backend-run:
	cd $(BACKEND_DIR) && export $$(grep -v '^#' .env | xargs) && \
		PAYSLIP_IMPORT_JSON_PATH="$$PWD/payslips/payslips_import.json" \
		DATABASE_HOST=127.0.0.1 \
		go run ./main.go

backend-build:
	@mkdir -p $(BACKEND_DIR)/bin
	cd $(BACKEND_DIR) && go build -o bin/server ./main.go
	@echo ">>> Binary: $(BACKEND_DIR)/bin/server"

backend-test:
	cd $(BACKEND_DIR) && export $$(grep -v '^#' .env 2>/dev/null | xargs) ; go test ./... -v

gen-testdata:
	@echo ">>> Generating ING PDF fixture..."
	cd /tmp && mkdir -p pdfgen && cp $(CURDIR)/$(BACKEND_DIR)/scripts/testdata/gen_ing_pdf.go /tmp/pdfgen/main.go && \
		cd /tmp/pdfgen && ([ -f go.mod ] || go mod init pdfgen) && \
		go get github.com/jung-kurt/gofpdf@latest && go run main.go
	@echo ">>> Generating ING CSV fixture..."
	cd $(BACKEND_DIR) && go run scripts/testdata/gen_csv.go
	@echo ">>> Generating Amazon Visa XLS fixture..."
	@(python3 -m venv /tmp/xlvenv 2>/dev/null || true) && \
		/tmp/xlvenv/bin/pip install xlwt -q && \
		/tmp/xlvenv/bin/python3 $(BACKEND_DIR)/scripts/testdata/gen_amazon_visa.py
	@echo ">>> All test fixtures regenerated."

# ── Frontend ──────────────────────────────────────────────────────────────────
frontend-dev:
	cd $(FRONTEND_DIR) && npx vite --mode development

frontend-dev-prod:
	cd $(FRONTEND_DIR) && npx vite --mode production

frontend-build:
	cd $(FRONTEND_DIR) && npm run build

# ── Cleanup ───────────────────────────────────────────────────────────────────
clean:
	@rm -rf $(BACKEND_DIR)/bin
	@echo ">>> Removed backend binary."

clean-all: clean
	@rm -rf $(FRONTEND_DIR)/dist $(FRONTEND_DIR)/node_modules
	@echo ">>> Removed frontend dist and node_modules."

nuke: clean-all
	@echo ""
	@echo "  ⚠️  WARNING: This will permanently delete all containers, the"
	@echo "              database volume, and all built Docker images."
	@printf "  Are you sure? [y/N] " && read ans && [ "$${ans}" = "y" ] || (echo "Aborted." && exit 1)
	@echo ""
	@echo ">>> Stopping and removing containers + volumes..."
	docker compose down --volumes --remove-orphans 2>/dev/null || true
	@echo ">>> Removing built Docker images..."
	docker rmi -f \
		cogni-cash-backend:latest \
		cogni-cash-backend:$(TAG) \
		cogni-cash-frontend:latest \
		cogni-cash-frontend:$(TAG) \
		2>/dev/null || true
	@echo ">>> Removing dangling build-cache layers..."
	docker image prune -f 2>/dev/null || true
	@echo ""
	@echo "✅ Everything cleaned. Run 'make up' to start fresh."
	@echo ""

db-dump-local:
	@echo ">>> Dumping local database $(DB_NAME)…"
	@mkdir -p backups
	docker compose exec -T postgres pg_dump -U $(DB_USER) -d $(DB_NAME) -c -F p | gzip > backups/local_db_$$(date +%Y%m%d_%H%M%S).sql.gz
	@echo ">>> Done. Compressed dump saved to backups/ directory."

db-dump-remote:
	@echo ">>> Dumping database from $(DEPLOY_HOST)…"
	@mkdir -p backups
	$(DEPLOY_SSH) \
		"cd $(DEPLOY_PATH) && export \$$(grep -v '^#' backend/.env | xargs) && docker compose exec -T postgres pg_dump -U \$$POSTGRES_USER -d \$$POSTGRES_DB -c -F p" | gzip > backups/prod_db_$$(date +%Y%m%d_%H%M%S).sql.gz
	@echo ">>> Done. Compressed dump saved to backups/ directory."

db-restore-remote:
	@[ -n "$(FILE)" ] || (echo "ERROR: FILE is required. Usage: make db-restore-remote FILE=backups/dump.sql.gz"; exit 1)
	@echo ">>> ⚠️ WARNING: This will overwrite the database on $(DEPLOY_HOST)."
	@printf ">>> Are you sure? [y/N] " && read ans && [ "$${ans}" = "y" ] || (echo "Aborted." && exit 1)
	@echo ">>> Restoring $(FILE) to $(DEPLOY_HOST)…"
	@gunzip -c $(FILE) | $(DEPLOY_SSH) \
		"cd $(DEPLOY_PATH) && export \$$(grep -v '^#' backend/.env | xargs) && docker compose exec -T postgres psql -U \$$POSTGRES_USER -d \$$POSTGRES_DB"
	@echo ">>> Restore complete."

db-seed:
	@echo ">>> Seeding database with dummy data..."
	$(PSQL) -f $(BACKEND_DIR)/balance/dummy-data.sql
	@echo ">>> Dummy data inserted successfully."

db-restore-local:
	@[ -n "$(FILE)" ] || (echo "ERROR: FILE is required. Usage: make db-restore-local FILE=backups/dump.sql.gz"; exit 1)
	@echo ">>> ⚠️ WARNING: This will overwrite your local database."
	@printf ">>> Are you sure? [y/N] " && read ans && [ "$${ans}" = "y" ] || (echo "Aborted." && exit 1)
	@echo ">>> Wiping local schema to prevent foreign key conflicts…"
	@docker compose exec -T postgres psql -U $(DB_USER) -d $(DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@echo ">>> Restoring $(FILE) to local database…"
	@gunzip -c $(FILE) | docker compose exec -T postgres psql -U $(DB_USER) -d $(DB_NAME)
	@echo ">>> Restore complete."