APP_NAME := compressor
CMD_DIR  := ./cmd/api
BIN_DIR  := ./bin

.PHONY: all deps deps-system deps-go build run test clean

# Default target: build the binary
all: build

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

test:
	@echo "==> Running tests..."
	go test ./...

clean:
	@echo "==> Cleaning..."
	rm -rf $(BIN_DIR)
