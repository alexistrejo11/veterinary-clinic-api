.PHONY: help run migrate migrate-down sqlc swagger swagger-fmt build test test-race vet tidy docker-up docker-up-full docker-down docker-build clean

# -----------------------------------------------------------------------------
# Defaults
# -----------------------------------------------------------------------------
GO            ?= go
SQLC          ?= sqlc
SWAG          ?= $(GO) run github.com/swaggo/swag/cmd/swag@v1.16.6
MIGRATE_PATH  ?= database/migrations
SQLC_CONFIG   ?= database/sqlc.yaml
API_PKG       ?= ./cmd/api
MIGRATE_PKG   ?= ./cmd/migrate
BIN_DIR       ?= bin
API_BIN       ?= $(BIN_DIR)/api

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# -----------------------------------------------------------------------------
# App
# -----------------------------------------------------------------------------
run: ## Run the API server (go run ./cmd/api)
	$(GO) run $(API_PKG)

build: ## Build API binary to bin/api
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(API_BIN) $(API_PKG)

test: ## Run all Go tests
	$(GO) test ./...

test-race: ## Run all Go tests with the race detector
	$(GO) test -race ./...

vet: ## Run Go static analysis
	$(GO) vet ./...

tidy: ## Tidy Go module dependencies
	$(GO) mod tidy

# -----------------------------------------------------------------------------
# Database
# -----------------------------------------------------------------------------
migrate: ## Run database migrations up (go run ./cmd/migrate)
	$(GO) run $(MIGRATE_PKG)

migrate-down: ## Roll back the last migration (requires migrate CLI)
	@test -n "$$DATABASE_URL" || (echo "DATABASE_URL is required" >&2; exit 1)
	migrate -path $(MIGRATE_PATH) -database "$$DATABASE_URL" down 1

sqlc: ## Regenerate sqlc Go code from database/queries + migrations
	$(SQLC) generate -f $(SQLC_CONFIG)

# -----------------------------------------------------------------------------
# Docs
# -----------------------------------------------------------------------------
swagger: ## Regenerate Swagger docs from cmd/api annotations
	$(SWAG) init -g cmd/api/main.go -o docs --parseDependency --parseInternal

swagger-fmt: ## Format swag annotations
	$(SWAG) fmt -g cmd/api/main.go

# -----------------------------------------------------------------------------
# Docker
# -----------------------------------------------------------------------------
docker-up: ## Start standalone API (default profile)
	docker compose up --build

docker-up-full: ## Start API + Postgres + Redis (full profile)
	docker compose --profile full up --build

docker-down: ## Stop all compose services
	docker compose --profile full down

docker-build: ## Build the API image only
	docker compose build api

# -----------------------------------------------------------------------------
# Housekeeping
# -----------------------------------------------------------------------------
clean: ## Remove local build artifacts
	rm -rf $(BIN_DIR)
