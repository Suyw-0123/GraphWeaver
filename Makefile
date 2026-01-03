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
# Kubernetes Infrastructure
# ==============================================================================

.PHONY: kind-create
kind-create: ## Create kind cluster with port mappings
	kind create cluster --name graph-weaver-cluster --config config/kind-cluster.yaml || true
	kubectl cluster-info --context kind-graph-weaver-cluster

.PHONY: kind-delete
kind-delete: ## Delete kind cluster
	kind delete cluster --name graph-weaver-cluster

.PHONY: kind-recreate
kind-recreate: kind-delete kind-create ## Recreate kind cluster

.PHONY: k9s
k9s: ## Launch k9s terminal UI
	k9s --context kind-graph-weaver-cluster

# ==============================================================================
# Helm & Database Setup
# ==============================================================================

.PHONY: helm-init
helm-init: ## Add Helm repositories
	@echo "Adding Helm repositories..."
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo add neo4j https://helm.neo4j.com/neo4j
	helm repo add qdrant https://qdrant.github.io/qdrant-helm
	helm repo update
	@echo "✅ Helm repositories added"

.PHONY: db-install
db-install: ## Install all databases (PostgreSQL, Neo4j, Qdrant)
	@echo "Installing databases via Helm..."
	helm upgrade --install postgresql bitnami/postgresql \
		--set auth.postgresPassword=graphweaver123 \
		--set auth.username=graphweaver \
		--set auth.password=graphweaver123 \
		--set auth.database=graphweaver \
		--set primary.persistence.size=2Gi \
		--namespace default
	helm upgrade --install neo4j neo4j/neo4j \
		--set neo4j.name=neo4j \
		--set neo4j.password=neo4j123 \
		--set neo4j.edition=community \
		--set volumes.data.mode=defaultStorageClass \
		--set volumes.data.defaultStorageClass.requests.storage=2Gi \
		--namespace default
	helm upgrade --install qdrant qdrant/qdrant \
		--set persistence.size=2Gi \
		--namespace default
	@echo "✅ Databases installed. Waiting for ready status..."
	kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=postgresql --timeout=300s || true
	kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=neo4j --timeout=300s || true
	kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=qdrant --timeout=300s || true

.PHONY: db-uninstall
db-uninstall: ## Uninstall all databases
	helm uninstall postgresql --namespace default || true
	helm uninstall neo4j --namespace default || true
	helm uninstall qdrant --namespace default || true

.PHONY: db-status
db-status: ## Check database pods status
	@echo "=== Database Pods Status ==="
	kubectl get pods -l app.kubernetes.io/name=postgresql || echo "PostgreSQL not found"
	kubectl get pods -l app.kubernetes.io/name=neo4j || echo "Neo4j not found"
	kubectl get pods -l app.kubernetes.io/name=qdrant || echo "Qdrant not found"
	@echo "\n=== Database Services ==="
	kubectl get svc | grep -E "postgresql|neo4j|qdrant" || echo "No database services found"

.PHONY: db-port-forward
db-port-forward: ## Port forward all databases to localhost
	@echo "Starting port forwarding (Ctrl+C to stop)..."
	@echo "PostgreSQL: localhost:5432"
	@echo "Neo4j: localhost:7474 (HTTP), localhost:7687 (Bolt)"
	@echo "Qdrant: localhost:6333"
	kubectl port-forward svc/postgresql 5432:5432 &
	kubectl port-forward svc/neo4j 7474:7474 7687:7687 &
	kubectl port-forward svc/qdrant 6333:6333 &
	@echo "Port forwarding active. Press Ctrl+C to stop all."

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
# Docker & Skaffold
# ==============================================================================

.PHONY: docker-build
docker-build: ## Build Docker images locally
	docker build -t graphweaver/gateway:latest -f deployments/docker/Dockerfile.gateway .
	docker build -t graphweaver/ingestion:latest -f deployments/docker/Dockerfile.ingestion .
	docker build -t graphweaver/chat:latest -f deployments/docker/Dockerfile.chat .

.PHONY: skaffold-dev
skaffold-dev: ## Run Skaffold in dev mode (auto-rebuild on changes)
	skaffold dev --port-forward

.PHONY: skaffold-run
skaffold-run: ## Deploy to K8s via Skaffold
	skaffold run

# ==============================================================================
# Clean
# ==============================================================================

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache

.PHONY: clean-all
clean-all: clean db-uninstall kind-delete ## Nuclear clean (remove everything)
	@echo "✅ Complete cleanup done"
