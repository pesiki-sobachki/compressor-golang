APP_NAME := api
CMD_DIR  := ./cmd/api
BIN_DIR  := ./bin

.PHONY: all deps deps-system deps-go build run test clean

# Default target: build the binary
all: build

COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME := $(shell date +%FT%T%z)

LDFLAGS := -ldflags "-s -w -X main.CommitHash=$(COMMIT_HASH) -X main.BuildTime=$(BUILD_TIME)"
GO_BUILD_FLAGS := -trimpath

########################################
# Dependencies
########################################

# Full dependencies setup: system + Go
deps: deps-system deps-go

# System dependencies for Ubuntu/WSL (libvips + C compiler)
deps-system:
	@echo "==> Installing system dependencies (Ubuntu / Debian)..."
	sudo apt update
	sudo apt install -y build-essential libvips libvips-dev

# Go module dependencies
deps-go:
	@echo "==> Downloading Go modules..."
	go mod tidy
	go mod download

########################################
# Build and run
########################################

build:
	@echo "==> Building binary..."
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)

run: build
	@echo "==> Running $(APP_NAME)..."
	$(BIN_DIR)/$(APP_NAME)

##@ Builds


build-linux:
	@echo "Building Linux binary..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o build/$(BINARY_NAME)-linux $(CMD_API_PATH)
	@echo "Build complete: build/$(BINARY_NAME)-linux"

build-windows: swagger ## Build binary for Windows (AMD64)
	@echo "Building Windows binary..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o build/$(BINARY_NAME)-windows.exe $(CMD_API_PATH)
	@echo "Build complete: build/$(BINARY_NAME)-windows.exe"

test:
	@echo "==> Running tests..."
	go test ./...

clean:
	@echo "==> Cleaning..."
	rm -rf $(BIN_DIR)

########################################
# Audit
########################################

lint: ## Run golangci-lint
	@golangci-lint run ./...

lint-install: ## Install golangci-lint
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
