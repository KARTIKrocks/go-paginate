.PHONY: test coverage lint fmt vet build clean install-tools example help

# Variables
GO := go
GOTEST := $(GO) test
GOVET := $(GO) vet
GOFMT := $(GO) fmt
COVERAGE_FILE := coverage.out

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

test: ## Run tests
	$(GOTEST) -v -race ./...

coverage: ## Run tests with coverage
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GO) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	golangci-lint run --timeout=5m

fmt: ## Format code
	$(GOFMT) ./...

vet: ## Run go vet
	$(GOVET) ./...

build: ## Build the library
	$(GO) build -v ./...

clean: ## Clean build artifacts and coverage files
	rm -f $(COVERAGE_FILE) coverage.html
	rm -rf dist/ build/
	$(GO) clean

install-tools: ## Install development tools
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed successfully"

example: ## Run the example server
	cd examples/server && $(GO) run main.go

bench: ## Run benchmarks
	$(GO) test -bench=. -benchmem ./...

mod-tidy: ## Tidy go.mod
	$(GO) mod tidy

mod-verify: ## Verify dependencies
	$(GO) mod verify

ci: fmt vet lint test ## Run all CI checks

all: clean fmt vet lint test coverage ## Run all checks and tests

.DEFAULT_GOAL := help
