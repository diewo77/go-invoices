.PHONY: help run dev test build clean \
        docker-build docker-up docker-down docker-logs docker-rebuild \
        docker-dev-up docker-dev-down docker-dev-logs docker-dev-rebuild docker-dev-nocache

APP_NAME=go-invoices
PORT?=8080
DATABASE_DSN?=postgres://invoices:invoices123@localhost:5432/invoices?sslmode=disable

.DEFAULT_GOAL := help

help: ## Show this help
	@echo "Make targets:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_.-]+:.*##/ { printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

run: ## Run the server locally
	go run ./go-invoices/cmd/server

dev: ## Run with hot reload using reflex
	reflex -r '\.go$$' -s -- sh -c 'go run ./go-invoices/cmd/server'

build: ## Build the server binary
	go build -o bin/server ./go-invoices/cmd/server

test: ## Run all tests
	go test ./go-invoices/... ./go-gate/...

clean: ## Clean build artifacts
	rm -rf bin/

# Docker production targets
docker-build: ## Build production Docker image
	docker build -t $(APP_NAME):latest .

docker-up: ## Start production stack
	docker compose up -d

docker-down: ## Stop production stack
	docker compose down

docker-logs: ## View production logs
	docker compose logs -f

docker-rebuild: ## Rebuild and restart production stack
	docker compose up -d --build

# Docker dev targets
docker-dev-up: ## Start dev stack with live reload
	docker compose -f docker-compose.dev.yml up 

docker-dev-down: ## Stop dev stack
	docker compose -f docker-compose.dev.yml down

docker-dev-logs: ## View dev logs
	docker compose -f docker-compose.dev.yml logs -f

docker-dev-rebuild: ## Rebuild and restart dev stack
	docker compose -f docker-compose.dev.yml up -d --build

docker-dev-nocache: ## Rebuild dev stack without cache
	docker compose -f docker-compose.dev.yml build --no-cache
	docker compose -f docker-compose.dev.yml up -d

# Database migration targets
migrate: ## Run DB migrations locally (uses DATABASE_DSN -> exported to DATABASE_URL)
	@echo "Running migrations against $(DATABASE_DSN)"
	DATABASE_URL=$(DATABASE_DSN) go run ./go-invoices/cmd/server -migrate-only

seed: ## Run DB seed locally
	@echo "Seeding database $(DATABASE_DSN)"
	DATABASE_URL=$(DATABASE_DSN) go run ./go-invoices/cmd/server -seed-only

docker-migrate: ## Run migrations inside dev container (uses go run)
	@echo "Running migrations inside container"
	docker compose -f docker-compose.dev.yml run --rm app sh -c "cd /app/go-invoices && go run ./cmd/server -migrate-only"

docker-seed: ## Run seeding inside dev container (uses go run)
	@echo "Seeding inside container"
	docker compose -f docker-compose.dev.yml run --rm app sh -c "cd /app/go-invoices && go run ./cmd/server -seed-only"
