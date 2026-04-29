.PHONY: help docker up down all logs dev server web migrate migrate-reset build clean

GO       := go
GOFLAGS  :=
SERVER   := ./cmd/server
DB_PORT  ?= 15432

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# --------------- Docker ---------------

docker: ## (use: make docker up/down/all/logs)
	@echo "Usage: make docker <up|down|all|logs>"

up: ## Start postgres + redis containers
	docker compose up -d postgres redis

down: ## Stop all containers
	docker compose down

all: ## Start all services (including server container)
	docker compose up -d --build

logs: ## Tail container logs
	docker compose logs -f

# --------------- Dev ---------------

dev: up server web ## Start everything for local development

server: ## Run Go backend
	DB_PORT=$(DB_PORT) $(GO) run $(GOFLAGS) $(SERVER)

web: ## Run Vite frontend dev server
	cd web && npm run dev

build: ## Build Go binary
	CGO_ENABLED=0 $(GO) build -o bin/sws-server $(SERVER)

# --------------- Database ---------------

migrate: ## Run all up migrations against local DB
	@for f in migrations/*up.sql; do \
		echo "==> $$f"; \
		PGPASSWORD=sws_secret psql -h localhost -p $(DB_PORT) -U sws -d starfall_warsong -f "$$f" 2>&1 || true; \
	done

migrate-reset: ## Drop and recreate the database
	PGPASSWORD=sws_secret psql -h localhost -p $(DB_PORT) -U sws -d postgres \
		-c "DROP DATABASE IF EXISTS starfall_warsong;" \
		-c "CREATE DATABASE starfall_warsong;"
	$(MAKE) migrate

# --------------- Misc ---------------

clean: ## Remove build artifacts
	rm -rf bin/
