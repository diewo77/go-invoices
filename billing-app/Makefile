APP_NAME=billing-app
PKG=./...
PORT?=8080
GOBIN?=$(shell go env GOPATH)/bin
GOLANGCI_LINT?=$(shell command -v golangci-lint 2>/dev/null || echo $(GOBIN)/golangci-lint)

COMPOSE_DEV=docker compose -f docker-compose.yml -f docker-compose.dev.yml
COMPOSE=docker compose

# Default target prints available commands
.DEFAULT_GOAL := help
SHELL := /bin/sh

.PHONY: help run dev test lint tidy fmt build clean \
        docker-build docker-up docker-down docker-logs docker-rebuild \
        docker-dev-up docker-dev-down docker-dev-logs docker-dev-rebuild docker-dev-nocache \
        migrate-up migrate-down migrate-force migrate-new \
	dev-migrate-up dev-migrate-down dev-migrate-force \
	up down logs rebuild nocache migrate start lint-install

help: ## Show this help
	@echo "Make targets:" 
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_.-]+:.*##/ { printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

run: ## Run the server locally
	go run ./cmd/server

dev: ## Run the server locally with reflex (no Docker)
	reflex -r '\.go$$' -s -- sh -c 'go run ./cmd/server'

test: ## Run tests (no cache)
	go test -count=1 $(PKG)

lint: ## Run go vet and golangci-lint
	go vet $(PKG)
	@if [ "$(RUN_LOCAL)" = "1" ]; then \
		$(GOLANGCI_LINT) run $(ARGS) || true; \
	else \
		$(GOLANGCI_LINT) run $(ARGS); \
	fi

lint-install: ## Install golangci-lint locally (pin version with VERSION=vX.Y.Z)
	./scripts/install-golangci-lint.sh $(VERSION)

tidy: ## go mod tidy
	go mod tidy

fmt: ## go fmt all packages
	go fmt $(PKG)

build: ## Build local binary into bin/
	go build -o bin/$(APP_NAME) ./cmd/server

clean: ## Remove build artifacts
	rm -rf bin

# --- Docker (production-like) ---
docker-build:
	$(COMPOSE) build app

docker-up: ## Start prod stack (app + db)
	$(COMPOSE) up -d --build

docker-down: ## Stop prod stack
	$(COMPOSE) down

docker-logs: ## Follow app logs
	$(COMPOSE) logs -f app

docker-rebuild: ## Force rebuild without cache
	$(COMPOSE) build --no-cache app

# --- Docker dev (hot reload) ---
docker-dev-up: ## Start dev stack with watch/reflex
	$(COMPOSE_DEV) up --build

docker-dev-down: ## Stop dev stack
	$(COMPOSE_DEV) down

docker-dev-logs: ## Follow dev app logs
	$(COMPOSE_DEV) logs -f app

docker-dev-rebuild: ## Rebuild dev app image
	$(COMPOSE_DEV) build app

docker-dev-nocache: ## Rebuild dev app image without cache and start stack
	$(COMPOSE_DEV) build --no-cache app
	$(COMPOSE_DEV) up --build

# --- Simple aliases (one-liners) ---
start: up ## Alias for up
up: ## Start dev stack (watch/reflex)
	$(MAKE) docker-dev-up
down: ## Stop dev stack
	$(MAKE) docker-dev-down
logs: ## Follow dev app logs
	$(MAKE) docker-dev-logs
rebuild: ## Rebuild dev app image
	$(MAKE) docker-dev-rebuild
nocache: ## Rebuild dev app image without cache and start stack
	$(MAKE) docker-dev-nocache
migrate: ## Apply all up migrations against docker dev DB
	$(MAKE) dev-migrate-up

# --- Migrations (golang-migrate) ---
# Requires MIGRATIONS=1 in environment when app starts to apply at runtime.
MIGRATE_BIN ?= migrate
DB_DSN ?= $$DATABASE_DSN
DB_DSN_DEV ?= postgres://billing:billing@db:5432/billing?sslmode=disable

migrate-up: ## Apply all up migrations
	@if [ -z "$(DB_DSN)" ]; then echo "DB_DSN (or env DATABASE_DSN) required"; exit 1; fi
	$(MIGRATE_BIN) -path migrations -database "$(DB_DSN)" up

migrate-down: ## Rollback one migration
	@if [ -z "$(DB_DSN)" ]; then echo "DB_DSN (or env DATABASE_DSN) required"; exit 1; fi
	$(MIGRATE_BIN) -path migrations -database "$(DB_DSN)" down 1

migrate-force: ## Force set version (use VERSION=X)
	@if [ -z "$(DB_DSN)" ]; then echo "DB_DSN (or env DATABASE_DSN) required"; exit 1; fi
	@if [ -z "$(VERSION)" ]; then echo "VERSION required"; exit 1; fi
	$(MIGRATE_BIN) -path migrations -database "$(DB_DSN)" force $(VERSION)

migrate-new: ## Create new numbered migration (use NAME=description)
	@if [ -z "$(NAME)" ]; then echo "NAME required"; exit 1; fi
	@next=$$(printf "%04d" $$(( $$(ls migrations | grep -E '^[0-9]{4}_.+\.up\.sql' | wc -l)+1 ))); \
	 up="migrations/$${next}_$(NAME).up.sql"; down="migrations/$${next}_$(NAME).down.sql"; \
	 touch "$$up" "$$down"; \
	 echo "Created $$up $$down";

# --- Docker dev migrations helpers ---

dev-migrate-up: ## Apply all up migrations against docker dev DB
	$(COMPOSE_DEV) run --rm -e DATABASE_DSN="$(DB_DSN_DEV)" app \
		sh -c '$(MIGRATE_BIN) -path migrations -database "$$DATABASE_DSN" up'

dev-migrate-down: ## Rollback one migration on docker dev DB
	$(COMPOSE_DEV) run --rm -e DATABASE_DSN="$(DB_DSN_DEV)" app \
		sh -c '$(MIGRATE_BIN) -path migrations -database "$$DATABASE_DSN" down 1'

dev-migrate-force: ## Force set version on docker dev DB (use VERSION=X)
	@if [ -z "$(VERSION)" ]; then echo "VERSION required"; exit 1; fi
	$(COMPOSE_DEV) run --rm -e DATABASE_DSN="$(DB_DSN_DEV)" app \
		sh -c '$(MIGRATE_BIN) -path migrations -database "$$DATABASE_DSN" force $(VERSION)'
