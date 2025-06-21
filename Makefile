.PHONY: help build test clean deps install example migrate

# Variables
BINARY_NAME=go-active-record
MAIN_FILE=main.go
TEST_DIR=activerecord
EXAMPLE_DIR=examples

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Show help
	@echo "$(GREEN)Available commands:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}'

deps: ## Install dependencies
	@echo "$(GREEN)Installing dependencies...$(NC)"
	go mod tidy
	go mod download

build: deps ## Build project
	@echo "$(GREEN)Building project...$(NC)"
	go build -o $(BINARY_NAME) $(MAIN_FILE)

test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	go test -v ./$(TEST_DIR)/...

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	go test -v -coverprofile=coverage.out ./$(TEST_DIR)/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report saved to coverage.html$(NC)"

install: ## Install package
	@echo "$(GREEN)Installing package...$(NC)"
	go install

example: ## Run example
	@echo "$(GREEN)Running example...$(NC)"
	go run $(MAIN_FILE)

example-validations: ## Run validation example
	@echo "$(GREEN)Running validation example...$(NC)"
	go run $(EXAMPLE_DIR)/validations.go

migrate: ## Run migrations
	@echo "$(GREEN)Running migrations...$(NC)"
	go run $(EXAMPLE_DIR)/migrations.go

clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning artifacts...$(NC)"
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	rm -f coverage.html
	go clean

lint: ## Run linter
	@echo "$(GREEN)Running linter...$(NC)"
	golangci-lint run

fmt: ## Format code
	@echo "$(GREEN)Formatting code...$(NC)"
	go fmt ./...

vet: ## Check code
	@echo "$(GREEN)Checking code...$(NC)"
	go vet ./...

check: fmt vet lint test ## Run all checks

benchmark: ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(NC)"
	go test -bench=. ./$(TEST_DIR)/...

docs: ## Generate documentation
	@echo "$(GREEN)Generating documentation...$(NC)"
	godoc -http=:6060 &
	@echo "$(GREEN)Documentation available at: http://localhost:6060$(NC)"

docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t $(BINARY_NAME) .

docker-run: ## Run Docker container
	@echo "$(GREEN)Running Docker container...$(NC)"
	docker run -it --rm $(BINARY_NAME)

setup-dev: ## Setup development environment
	@echo "$(GREEN)Setting up development environment...$(NC)"
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "$(YELLOW)Installing golangci-lint...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v godoc &> /dev/null; then \
		echo "$(YELLOW)Installing godoc...$(NC)"; \
		go install golang.org/x/tools/cmd/godoc@latest; \
	fi
	@echo "$(GREEN)Development environment is ready!$(NC)"

# Database commands
db-setup: ## Setup database for development
	@echo "$(GREEN)Setting up database...$(NC)"
	@echo "$(YELLOW)Create database and run migrations:$(NC)"
	@echo "  make migrate"

db-reset: ## Reset database
	@echo "$(RED)Resetting database...$(NC)"
	@echo "$(YELLOW)Warning: all data will be deleted!$(NC)"
	@read -p "Continue? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1

# Release commands
release: check ## Prepare release
	@echo "$(GREEN)Preparing release...$(NC)"
	@echo "$(YELLOW)Don't forget to update version in go.mod$(NC)"

version: ## Show version
	@echo "$(GREEN)Version:$(NC)"
	@go version
	@echo "$(GREEN)Go module version:$(NC)"
	@go list -m -json | grep -E '"Path"|"Version"'

# Profiling commands
profile-cpu: ## CPU profiling
	@echo "$(GREEN)CPU profiling...$(NC)"
	go test -cpuprofile=cpu.prof -bench=. ./$(TEST_DIR)/...
	go tool pprof cpu.prof

profile-memory: ## Memory profiling
	@echo "$(GREEN)Memory profiling...$(NC)"
	go test -memprofile=mem.prof -bench=. ./$(TEST_DIR)/...
	go tool pprof mem.prof

# Security commands
security-check: ## Security check
	@echo "$(GREEN)Security check...$(NC)"
	@if command -v gosec &> /dev/null; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)gosec is not installed. Install: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest$(NC)"; \
	fi 