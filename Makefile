# ==============================================================================
# GuildBanker — Makefile
# Monorepo: Go API + React Frontend + Infrastructure
# ==============================================================================

APP_NAME        := guildbanker
API_DIR         := api
WEB_DIR         := web/spa
INFRA_DIR       := infra
MIGRATIONS_DIR  := migrations
BINARY          := $(API_DIR)/bin/$(APP_NAME)
DOCKER_COMPOSE  := docker compose

# Make recipes also find asdf when the calling shell has not loaded its config.
export PATH := $(HOME)/.local/bin:$(or $(ASDF_DATA_DIR),$(HOME)/.asdf)/bin:$(or $(ASDF_DATA_DIR),$(HOME)/.asdf)/shims:$(PATH)

# Let Go/asdf select GOROOT instead of inheriting a previous toolchain from the shell.
unexport GOROOT

.DEFAULT_GOAL   := help


# Colors (ANSI)
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RED    := \033[0;31m
CYAN   := \033[0;36m
RESET  := \033[0m

.PHONY: help
help: ## Show available commands
	@echo ""
	@echo "$(CYAN)⚔️  GuildBanker — Available Commands$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-24s$(RESET) %s\n", $$1, $$2}'
	@echo ""

# ==============================================================================
# Infrastructure
# ==============================================================================
.PHONY: infra/up infra/down infra/logs infra/clean

infra/up: ## Start all infrastructure services (Postgres + Keycloak)
	@echo "$(CYAN)▶ Starting infrastructure...$(RESET)"
	cd $(INFRA_DIR) && $(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)✔ Infrastructure is up$(RESET)"

infra/down: ## Stop all infrastructure services
	@echo "$(YELLOW)▶ Stopping infrastructure...$(RESET)"
	cd $(INFRA_DIR) && $(DOCKER_COMPOSE) down
	@echo "$(GREEN)✔ Infrastructure stopped$(RESET)"

infra/logs: ## Tail infrastructure logs (follow mode)
	cd $(INFRA_DIR) && $(DOCKER_COMPOSE) logs -f

infra/clean: ## Stop services and remove volumes (destroys data)
	@echo "$(RED)▶ Destroying infrastructure and volumes...$(RESET)"
	cd $(INFRA_DIR) && $(DOCKER_COMPOSE) down -v --remove-orphans
	@echo "$(GREEN)✔ Infrastructure cleaned$(RESET)"

# ==============================================================================
# Keycloak local realm setup
# ==============================================================================
.PHONY: keycloak/prepare keycloak/validate keycloak/up

keycloak/prepare: ## Combine realm, clients and test user for startup import (requires Python 3)
	python3 $(INFRA_DIR)/keycloak/prepare.py

keycloak/validate: ## Validate the separate Keycloak JSON definitions without writing
	python3 $(INFRA_DIR)/keycloak/prepare.py --check

keycloak/up: keycloak/prepare ## Start Postgres and Keycloak with initial realm import
	cd $(INFRA_DIR) && $(DOCKER_COMPOSE) -f docker-compose.yml -f keycloak/docker-compose.import.yml up -d --wait --wait-timeout 180 postgres keycloak

# ==============================================================================
# API (Go)
# ==============================================================================
.PHONY: api/build api/run api/test api/test/integration api/test/cover api/lint api/fmt api/tidy api/vet

api/build: ## Build the API binary
	@echo "$(CYAN)▶ Building API...$(RESET)"
	cd $(API_DIR) && go build -o bin/$(APP_NAME) ./cmd/server
	@echo "$(GREEN)✔ Binary built at $(BINARY)$(RESET)"

api/run: api/build ## Build and run the API
	@echo "$(CYAN)▶ Running API...$(RESET)"
	./$(BINARY)

api/test: ## Run API unit tests
	@echo "$(CYAN)▶ Running tests...$(RESET)"
	cd $(API_DIR) && go test ./... -race -count=1

api/test/integration: ## Run API integration tests with local fixtures
	cd $(API_DIR) && go test ./... -race -count=1 -tags=integration

api/test/cover: ## Run tests with coverage report
	@echo "$(CYAN)▶ Running tests with coverage...$(RESET)"
	cd $(API_DIR) && go test ./... -race -count=1 -coverprofile=coverage.out
	cd $(API_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✔ Coverage report: $(API_DIR)/coverage.html$(RESET)"

api/lint: ## Run linter (requires golangci-lint)
	@echo "$(CYAN)▶ Linting...$(RESET)"
	cd $(API_DIR) && go tool golangci-lint run ./...

api/fmt: ## Format Go source code
	@echo "$(CYAN)▶ Formatting...$(RESET)"
	cd $(API_DIR) && gofmt -s -w .
	@echo "$(GREEN)✔ Code formatted$(RESET)"

api/tidy: ## Tidy Go modules
	cd $(API_DIR) && go mod tidy

api/vet: ## Run go vet
	cd $(API_DIR) && go vet ./...

# ==============================================================================
# Database Migrations (requires golang-migrate CLI)
# ==============================================================================
.PHONY:migrate/create migrate/force

migrate/create: ## Create a new migration (usage: make migrate/create name=create_users)
	@if [ -z "$(name)" ]; then echo "$(RED)✘ Usage: make migrate/create name=<migration_name>$(RESET)"; exit 1; fi
	cd $(API_DIR) && go tool migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)
	@echo "$(GREEN)✔ Migration created$(RESET)"


migrate/force: ## Force migration version (usage: make migrate/force version=1)
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "Error: DATABASE_URL not defined"; \
		exit 1; \
	fi
	@if [ -z "$(version)" ]; then echo "$(RED)✘ Usage: make migrate/force version=<version>$(RESET)"; exit 1; fi
	cd $(API_DIR) && go tool migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $(version)

# ==============================================================================
# Web (React)
# ==============================================================================
.PHONY: web/install web/dev web/build web/lint web/test web/demo web/types web/e2e web/e2e/prod web/preview

web/install: ## Install frontend dependencies
	@echo "$(CYAN)▶ Installing frontend dependencies...$(RESET)"
	cd $(WEB_DIR) && npm ci

web/dev: ## Start frontend dev server
	@echo "$(CYAN)▶ Starting frontend dev server...$(RESET)"
	cd $(WEB_DIR) && npm run dev

web/build: ## Build frontend for production
	@echo "$(CYAN)▶ Building frontend...$(RESET)"
	cd $(WEB_DIR) && npm run build
	@echo "$(GREEN)✔ Frontend built$(RESET)"

web/lint: ## Lint frontend code
	cd $(WEB_DIR) && npm run lint

web/test: ## Run frontend tests
	cd $(WEB_DIR) && npm test

web/demo: ## Start frontend with explicit demonstration mocks
	cd $(WEB_DIR) && npm run dev:demo

web/types: ## Check generated frontend API types for drift
	cd $(WEB_DIR) && npm run api:types:check

web/e2e: ## Run frontend demonstration E2E tests
	cd $(WEB_DIR) && npm run test:e2e

web/e2e/prod: web/build ## Smoke-test the compiled static frontend
	cd $(WEB_DIR) && npm run test:e2e:prod

web/preview: ## Preview an existing static frontend build locally
	cd $(WEB_DIR) && npm run preview

# ==============================================================================
# All-in-one
# ==============================================================================
.PHONY: setup setup/tools dev check clean

setup/tools: ## Check system prerequisites and install missing asdf, Node.js/npm and Go
	@bash infra/scripts/setup-tools.sh

setup: setup/tools ## Prepare the environment and Keycloak, then stop containers (preserves data)
	@test -f api/config.env || cp api/config.env.example api/config.env
	@test -f infra/.env || cp infra/.env.example infra/.env
	@set -e; \
		trap 'status=$$?; $(MAKE) infra/down || exit $$?; exit $$status' EXIT; \
		trap 'exit 130' INT; trap 'exit 143' TERM; \
		$(MAKE) keycloak/up; \
		$(MAKE) api/tidy web/install
	@echo "$(GREEN)✔ Project setup complete$(RESET)"

dev: ## Start Postgres, Keycloak, API and UI for local development (requires Bash)
	@command -v node >/dev/null 2>&1 && node --version >/dev/null 2>&1 && \
		command -v npm >/dev/null 2>&1 && npm --version >/dev/null 2>&1 || \
		{ echo "Error: Node.js/npm unavailable. Install the Node.js version in .tool-versions (asdf install nodejs) and ensure it is on PATH."; exit 1; }
	$(MAKE) keycloak/up
	$(MAKE) api/build web/install
	@echo "$(CYAN)▶ API: http://localhost:8081 | UI: http://localhost:5173 | Keycloak: http://localhost:8080$(RESET)"
	@bash -c 'set -m; \
		api_pid=; web_pid=; \
		cleanup() { \
			trap "" INT TERM; \
			[ -z "$$api_pid" ] || kill -TERM -- -$$api_pid 2>/dev/null; \
			[ -z "$$web_pid" ] || kill -TERM -- -$$web_pid 2>/dev/null; \
			wait; \
		}; \
		trap cleanup EXIT; \
		trap "exit 130" INT; trap "exit 143" TERM; \
		./$(BINARY) & api_pid=$$!; \
		(cd $(WEB_DIR) && exec npm run dev -- --port 5173 --strictPort) & web_pid=$$!; \
		wait -n "$$api_pid" "$$web_pid"; exit $$?'

check: api/vet api/lint api/test web/lint web/test ## Run all checks (lint + test) for API and Web
	@echo "$(GREEN)✔ All checks passed$(RESET)"

clean: infra/clean ## Full cleanup (infra + build artifacts)
	rm -rf $(API_DIR)/bin $(API_DIR)/coverage.out $(API_DIR)/coverage.html $(API_DIR)/tmp
	rm -rf $(WEB_DIR)/node_modules $(WEB_DIR)/dist $(WEB_DIR)/build
	@echo "$(GREEN)✔ Full cleanup done$(RESET)"
