.PHONY: help swag build migrations test lint dev docker docker-down clean setup

SWAG_BIN=~/go/bin/swag
MAIN_FILE=cmd/api/main.go
OUTPUT_DIR=./api/docs
BINARY_NAME=main
BUILD_DIR=./tmp

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Setup development environment
	@./scripts/setup-dev.sh

dev: ## Start development server with hot reload
	@air

swag: ## Generate API documentation
	$(SWAG_BIN) init -g $(MAIN_FILE) --parseDependency --parseInternal --parseVendor -o $(OUTPUT_DIR)

build: ## Build the application
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(MAIN_FILE)

build-linux: ## Build for Linux
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux ./$(MAIN_FILE)

test: ## Run tests
	go test -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	golangci-lint run

fmt: ## Format code
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	go vet ./...

mod: ## Tidy and download dependencies
	go mod tidy
	go mod download

migrations: ## Run database migrations
	go run cmd/migrations/main.go

docker: ## Start all services with Docker Compose
	docker-compose up -d

docker-build: ## Build Docker image
	docker build -t go-api:latest .

docker-down: ## Stop Docker services
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	docker-compose down --remove-orphans --volumes

install-tools: ## Install development tools
	go install github.com/cosmtrek/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2

ci: mod lint test build ## Run CI pipeline locally