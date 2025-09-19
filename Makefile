# SAI CLI Build Configuration

# Application name and version
APP_NAME := sai
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build directories
BUILD_DIR := build
DIST_DIR := dist

# Go build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(COMMIT) -s -w"
GCFLAGS := -gcflags="all=-trimpath=$(PWD)"
ASMFLAGS := -asmflags="all=-trimpath=$(PWD)"

# Default target
.PHONY: all
all: clean build

# Build for current platform
.PHONY: build
build:
	@echo "Building $(APP_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/sai

# Build for all platforms
.PHONY: build-all
build-all: clean
	@echo "Building $(APP_NAME) for all platforms..."
	@mkdir -p $(DIST_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 ./cmd/sai
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) -o $(DIST_DIR)/$(APP_NAME)-linux-arm64 ./cmd/sai
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) -o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 ./cmd/sai
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) -o $(DIST_DIR)/$(APP_NAME)-darwin-arm64 ./cmd/sai
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd/sai
	
	# Windows ARM64
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) -o $(DIST_DIR)/$(APP_NAME)-windows-arm64.exe ./cmd/sai

# Development build with race detection
.PHONY: build-dev
build-dev:
	@echo "Building $(APP_NAME) for development..."
	@mkdir -p $(BUILD_DIR)
	go build -race $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-dev ./cmd/sai

# Install to local system
.PHONY: install
install: build
	@echo "Installing $(APP_NAME) to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/$(APP_NAME)
	sudo chmod +x /usr/local/bin/$(APP_NAME)

# Uninstall from local system
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(APP_NAME) from /usr/local/bin..."
	sudo rm -f /usr/local/bin/$(APP_NAME)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(DIST_DIR) coverage.out coverage.html

# Run the application
.PHONY: run
run: build
	./$(BUILD_DIR)/$(APP_NAME)

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build       - Build for current platform"
	@echo "  build-all   - Build for all supported platforms"
	@echo "  build-dev   - Build with race detection for development"
	@echo "  install     - Install to /usr/local/bin"
	@echo "  uninstall   - Remove from /usr/local/bin"
	@echo "  test        - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint        - Run linter"
	@echo "  fmt         - Format code"
	@echo "  tidy        - Tidy dependencies"
	@echo "  clean       - Clean build artifacts"
	@echo "  run         - Build and run the application"
	@echo "  help        - Show this help message"

# Default help target
.DEFAULT_GOAL := help