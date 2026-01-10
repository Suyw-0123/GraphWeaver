# GraphWeaver Makefile
# Engineering Best Practices: Automation & Reproducibility

.PHONY: help
help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ==============================================================================
# Development Environment
# ==============================================================================

.PHONY: dev-tools
dev-tools: ## Install Go development tools
	@echo "Installing Go development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install go.uber.org/mock/mockgen@latest
	@echo "✅ Development tools installed"

.PHONY: deps
deps: ## Download Go module dependencies
	go mod download
	go mod tidy
	go mod verify

# ==============================================================================
# Code Quality
# ==============================================================================

.PHONY: fmt
fmt: ## Format code with goimports
	goimports -w -local github.com/suyw-0123/graphweaver .
	gofmt -s -w .

.PHONY: lint
lint: ## Run linter
	golangci-lint run --timeout=5m ./...

.PHONY: test
test: ## Run all tests
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
test-coverage: test ## Run tests and show coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ==============================================================================
# Database Migration
# ==============================================================================

.PHONY: migrate-create
migrate-create: ## Create new migration (usage: make migrate-create NAME=init_schema)
	@if [ -z "$(NAME)" ]; then echo "Error: NAME is required. Usage: make migrate-create NAME=your_migration_name"; exit 1; fi
	migrate create -ext sql -dir migrations/postgres -seq $(NAME)

.PHONY: migrate-up
migrate-up: ## Run database migrations up
	migrate -path migrations/postgres -database "postgres://graphweaver:graphweaver123@localhost:5432/graphweaver?sslmode=disable" up

.PHONY: migrate-down
migrate-down: ## Rollback last migration
	migrate -path migrations/postgres -database "postgres://graphweaver:graphweaver123@localhost:5432/graphweaver?sslmode=disable" down 1

.PHONY: migrate-force
migrate-force: ## Force migration version (usage: make migrate-force VERSION=1)
	@if [ -z "$(VERSION)" ]; then echo "Error: VERSION is required"; exit 1; fi
	migrate -path migrations/postgres -database "postgres://graphweaver:graphweaver123@localhost:5432/graphweaver?sslmode=disable" force $(VERSION)

# ==============================================================================
# Build & Run
# ==============================================================================

.PHONY: build
build: ## Build all services
	@echo "Building services..."
	go build -o bin/ingestion ./cmd/ingestion
	go build -o bin/chat ./cmd/chat
	go build -o bin/gateway ./cmd/gateway
	@echo "✅ Build complete: bin/"

.PHONY: run-gateway
run-gateway: ## Run API Gateway locally
	go run cmd/gateway/main.go

.PHONY: run-ingestion
run-ingestion: ## Run Ingestion Service locally
	go run cmd/ingestion/main.go

.PHONY: run-chat
run-chat: ## Run Chat Service locally
	go run cmd/chat/main.go

# ==============================================================================
# Docker Build
# ==============================================================================

.PHONY: docker-build
docker-build: ## Build Docker images locally
	docker build -t graphweaver-server:latest .

# ==============================================================================
# Clean
# ==============================================================================

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache

.PHONY: clean-all
clean-all: clean ## Complete cleanup
	@echo "✅ Complete cleanup done"

