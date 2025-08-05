# Makefile for ClamAV API Go
# This Makefile provides standard Go development tasks and Docker build options
# including exportable tarballs for x86_64 infrastructure deployment.

# Project variables
PROJECT_NAME := clamav-api-go
BINARY_NAME := clamav-api
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)

# Docker variables
DOCKER_IMAGE := $(PROJECT_NAME)
DOCKER_TAG := $(VERSION)
DOCKER_REGISTRY := registry.smartservices.tech
DOCKER_FULL_IMAGE := $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
DOCKER_LATEST_IMAGE := $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest
DOCKER_PLATFORM := linux/amd64

# Build directories
BUILD_DIR := ./bin
DIST_DIR := ./dist
DOCKER_BUILD_DIR := ./build

# Go variables
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GO_VERSION := $(shell go version | cut -d' ' -f3)

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

.PHONY: help
help: ## Show this help message
	@echo "$(BLUE)ClamAV API Go - Makefile Help$(NC)"
	@echo "$(YELLOW)Version: $(VERSION)$(NC)"
	@echo "$(YELLOW)Go Version: $(GO_VERSION)$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "$(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Development Tasks

.PHONY: clean
clean: ## Clean build artifacts and caches
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -rf $(DOCKER_BUILD_DIR)
	go clean -cache -modcache -testcache
	docker system prune -f --filter label=project=$(PROJECT_NAME) 2>/dev/null || true

.PHONY: deps
deps: ## Download and tidy dependencies
	@echo "$(YELLOW)Downloading dependencies...$(NC)"
	go mod download
	go mod tidy
	go mod verify

.PHONY: tools
tools: ## Install development tools
	@echo "$(YELLOW)Installing development tools...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/air-verse/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest

## Build Tasks

.PHONY: build
build: clean deps ## Build the application binary
	@echo "$(YELLOW)Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)...$(NC)"
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) \
		.
	@echo "$(GREEN)Built $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

.PHONY: build-linux
build-linux: ## Build Linux x86_64 binary
	@echo "$(YELLOW)Building $(BINARY_NAME) for linux/amd64...$(NC)"
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 \
		.
	@echo "$(GREEN)Built $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64$(NC)"

.PHONY: build-all
build-all: ## Build binaries for multiple platforms
	@echo "$(YELLOW)Building for multiple platforms...$(NC)"
	mkdir -p $(BUILD_DIR)
	# Linux AMD64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	# Linux ARM64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	# macOS AMD64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	# macOS ARM64
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
		-ldflags "$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@echo "$(GREEN)Built all platform binaries$(NC)"

## Test Tasks

.PHONY: test
test: ## Run unit tests
	@echo "$(YELLOW)Running unit tests...$(NC)"
	go test -v -race -timeout=30s ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests with Venom
	@echo "$(YELLOW)Running E2E tests...$(NC)"
	@if command -v venom >/dev/null 2>&1; then \
		cd e2e && venom run venom.e2e.yaml; \
	else \
		echo "$(RED)Venom not installed. Install with: go install github.com/ovh/venom@latest$(NC)"; \
		exit 1; \
	fi

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(YELLOW)Running benchmarks...$(NC)"
	go test -bench=. -benchmem -run=^$$ ./...

## Quality Tasks

.PHONY: fmt
fmt: ## Format Go code
	@echo "$(YELLOW)Formatting code...$(NC)"
	go fmt ./...
	gofmt -s -w .

.PHONY: vet
vet: ## Run go vet
	@echo "$(YELLOW)Running go vet...$(NC)"
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	@echo "$(YELLOW)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(RED)golangci-lint not installed. Run 'make tools' first$(NC)"; \
		exit 1; \
	fi

.PHONY: lint-fix
lint-fix: ## Run golangci-lint with auto-fix
	@echo "$(YELLOW)Running linter with auto-fix...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix ./...; \
	else \
		echo "$(RED)golangci-lint not installed. Run 'make tools' first$(NC)"; \
		exit 1; \
	fi

.PHONY: check
check: fmt vet lint test ## Run all quality checks

## Security Tasks

.PHONY: security
security: ## Run security checks with gosec
	@echo "$(YELLOW)Running security checks...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)Installing gosec...$(NC)"; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

.PHONY: vuln-check
vuln-check: ## Check for vulnerabilities in dependencies
	@echo "$(YELLOW)Checking for vulnerabilities...$(NC)"
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "$(YELLOW)Installing govulncheck...$(NC)"; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
		govulncheck ./...; \
	fi

## Docker Tasks

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(YELLOW)Building Docker image...$(NC)"
	docker build \
		--platform $(DOCKER_PLATFORM) \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--label project=$(PROJECT_NAME) \
		--label version=$(VERSION) \
		--label commit=$(COMMIT) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):latest \
		.
	@echo "$(GREEN)Built Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)$(NC)"

.PHONY: docker-build-multiarch
docker-build-multiarch: ## Build multi-architecture Docker image
	@echo "$(YELLOW)Building multi-architecture Docker image...$(NC)"
	docker buildx create --use --name $(PROJECT_NAME)-builder 2>/dev/null || true
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--label project=$(PROJECT_NAME) \
		--label version=$(VERSION) \
		--label commit=$(COMMIT) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):latest \
		--push \
		.

.PHONY: docker-export
docker-export: docker-build ## Export Docker image as tarball for x86_64
	@echo "$(YELLOW)Exporting Docker image as tarball...$(NC)"
	mkdir -p $(DIST_DIR)
	docker save $(DOCKER_IMAGE):$(DOCKER_TAG) | gzip > $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)-linux-amd64.tar.gz
	@echo "$(GREEN)Exported Docker image: $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)-linux-amd64.tar.gz$(NC)"
	@echo "$(BLUE)Image size: $$(du -h $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)-linux-amd64.tar.gz | cut -f1)$(NC)"

.PHONY: docker-export-registry
docker-export-registry: ## Export Docker image with registry tag as tarball
	@echo "$(YELLOW)Building and exporting registry-tagged Docker image...$(NC)"
	mkdir -p $(DIST_DIR)
	docker build \
		--platform $(DOCKER_PLATFORM) \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--label project=$(PROJECT_NAME) \
		--label version=$(VERSION) \
		--label commit=$(COMMIT) \
		-t $(DOCKER_FULL_IMAGE) \
		-t $(DOCKER_LATEST_IMAGE) \
		.
	docker save $(DOCKER_FULL_IMAGE) | gzip > $(DIST_DIR)/$(PROJECT_NAME)-registry-$(VERSION)-linux-amd64.tar.gz
	@echo "$(GREEN)Exported registry image: $(DIST_DIR)/$(PROJECT_NAME)-registry-$(VERSION)-linux-amd64.tar.gz$(NC)"
	@echo "$(BLUE)Image size: $$(du -h $(DIST_DIR)/$(PROJECT_NAME)-registry-$(VERSION)-linux-amd64.tar.gz | cut -f1)$(NC)"

.PHONY: docker-load
docker-load: ## Load Docker image from tarball
	@echo "$(YELLOW)Loading Docker image from tarball...$(NC)"
	@if [ -f "$(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)-linux-amd64.tar.gz" ]; then \
		gunzip -c $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)-linux-amd64.tar.gz | docker load; \
		echo "$(GREEN)Loaded Docker image from tarball$(NC)"; \
	else \
		echo "$(RED)Tarball not found: $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)-linux-amd64.tar.gz$(NC)"; \
		echo "$(YELLOW)Run 'make docker-export' first$(NC)"; \
		exit 1; \
	fi

.PHONY: docker-push
docker-push: docker-build ## Push Docker image to registry
	@echo "$(YELLOW)Pushing Docker image to registry...$(NC)"
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_FULL_IMAGE)
	docker tag $(DOCKER_IMAGE):latest $(DOCKER_LATEST_IMAGE)
	docker push $(DOCKER_FULL_IMAGE)
	docker push $(DOCKER_LATEST_IMAGE)
	@echo "$(GREEN)Pushed images to registry$(NC)"

.PHONY: docker-run
docker-run: ## Run Docker container locally
	@echo "$(YELLOW)Running Docker container...$(NC)"
	docker run --rm -it \
		-p 8080:8080 \
		-e LOGGER_LOG_LEVEL=debug \
		-e LOGGER_FORMAT=console \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-compose-up
docker-compose-up: ## Start services with docker-compose
	@echo "$(YELLOW)Starting services with docker-compose...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)Services started. ClamAV API available at http://localhost:8080$(NC)"

.PHONY: docker-compose-down
docker-compose-down: ## Stop services with docker-compose
	@echo "$(YELLOW)Stopping services with docker-compose...$(NC)"
	docker-compose down

## Development Tasks

.PHONY: dev
dev: ## Start development server with hot reload
	@echo "$(YELLOW)Starting development server with hot reload...$(NC)"
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "$(RED)Air not installed. Run 'make tools' first$(NC)"; \
		exit 1; \
	fi

.PHONY: run
run: build ## Build and run the application
	@echo "$(YELLOW)Running $(BINARY_NAME)...$(NC)"
	$(BUILD_DIR)/$(BINARY_NAME)

## Release Tasks

.PHONY: release-artifacts
release-artifacts: clean build-all docker-export docker-export-registry ## Build all release artifacts
	@echo "$(YELLOW)Creating release artifacts...$(NC)"
	mkdir -p $(DIST_DIR)
	# Create checksums
	cd $(BUILD_DIR) && sha256sum * > ../$(DIST_DIR)/checksums.txt
	cd $(DIST_DIR) && sha256sum *.tar.gz >> checksums.txt
	@echo "$(GREEN)Release artifacts created in $(DIST_DIR)/$(NC)"
	@echo "$(BLUE)Artifacts:$(NC)"
	@ls -la $(DIST_DIR)/
	@echo "$(BLUE)Binaries:$(NC)"
	@ls -la $(BUILD_DIR)/

.PHONY: info
info: ## Show build information
	@echo "$(BLUE)ClamAV API Go - Build Information$(NC)"
	@echo "$(YELLOW)Project:$(NC)     $(PROJECT_NAME)"
	@echo "$(YELLOW)Version:$(NC)     $(VERSION)"
	@echo "$(YELLOW)Commit:$(NC)      $(COMMIT)"
	@echo "$(YELLOW)Build Time:$(NC)  $(BUILD_TIME)"
	@echo "$(YELLOW)Go Version:$(NC)  $(GO_VERSION)"
	@echo "$(YELLOW)Platform:$(NC)    $(GOOS)/$(GOARCH)"
	@echo "$(YELLOW)Docker Image:$(NC) $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo "$(YELLOW)Registry:$(NC)    $(DOCKER_FULL_IMAGE)"

# Default target
.DEFAULT_GOAL := help
