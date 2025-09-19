# SAI CLI Build Configuration

# Application name and version
APP_NAME := sai
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION := $(shell go version | cut -d' ' -f3)

# Build directories
BUILD_DIR := build
DIST_DIR := dist
RELEASE_DIR := release
PACKAGE_DIR := packages

# Platform and architecture combinations
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64
LINUX_PLATFORMS := linux/amd64 linux/arm64
DARWIN_PLATFORMS := darwin/amd64 darwin/arm64
WINDOWS_PLATFORMS := windows/amd64 windows/arm64

# Go build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(COMMIT) -X main.goVersion=$(GO_VERSION) -s -w"
GCFLAGS := -gcflags="all=-trimpath=$(PWD)"
ASMFLAGS := -asmflags="all=-trimpath=$(PWD)"

# Release flags for optimized builds
RELEASE_LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(COMMIT) -X main.goVersion=$(GO_VERSION) -s -w -extldflags '-static'"
CGO_ENABLED := 0

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
	@echo "Building $(APP_NAME) v$(VERSION) for all platforms..."
	@mkdir -p $(DIST_DIR)
	@$(foreach platform,$(PLATFORMS),$(call build_platform,$(platform)))

# Build for specific platform
define build_platform
	$(eval GOOS := $(word 1,$(subst /, ,$(1))))
	$(eval GOARCH := $(word 2,$(subst /, ,$(1))))
	$(eval BINARY_NAME := $(APP_NAME)-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe))
	@echo "Building for $(GOOS)/$(GOARCH)..."
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) $(GCFLAGS) $(ASMFLAGS) -o $(DIST_DIR)/$(BINARY_NAME) ./cmd/sai
endef

# Build release binaries with optimizations
.PHONY: build-release
build-release: clean test lint
	@echo "Building $(APP_NAME) v$(VERSION) release binaries..."
	@mkdir -p $(RELEASE_DIR)
	@$(foreach platform,$(PLATFORMS),$(call build_release_platform,$(platform)))

# Build release for specific platform
define build_release_platform
	$(eval GOOS := $(word 1,$(subst /, ,$(1))))
	$(eval GOARCH := $(word 2,$(subst /, ,$(1))))
	$(eval BINARY_NAME := $(APP_NAME)-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe))
	@echo "Building release for $(GOOS)/$(GOARCH)..."
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(RELEASE_LDFLAGS) $(GCFLAGS) $(ASMFLAGS) -o $(RELEASE_DIR)/$(BINARY_NAME) ./cmd/sai
endef

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

# Create distribution packages
.PHONY: package
package: build-release
	@echo "Creating distribution packages..."
	@mkdir -p $(PACKAGE_DIR)
	@$(foreach platform,$(PLATFORMS),$(call create_package,$(platform)))

# Create package for specific platform
define create_package
	$(eval GOOS := $(word 1,$(subst /, ,$(1))))
	$(eval GOARCH := $(word 2,$(subst /, ,$(1))))
	$(eval BINARY_NAME := $(APP_NAME)-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe))
	$(eval PACKAGE_NAME := $(APP_NAME)-$(VERSION)-$(GOOS)-$(GOARCH))
	@echo "Creating package for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(PACKAGE_DIR)/$(PACKAGE_NAME)
	@cp $(RELEASE_DIR)/$(BINARY_NAME) $(PACKAGE_DIR)/$(PACKAGE_NAME)/$(if $(filter windows,$(GOOS)),$(APP_NAME).exe,$(APP_NAME))
	@cp README.md $(PACKAGE_DIR)/$(PACKAGE_NAME)/
	@cp LICENSE $(PACKAGE_DIR)/$(PACKAGE_NAME)/ 2>/dev/null || echo "LICENSE file not found, skipping..."
	@if [ "$(GOOS)" = "windows" ]; then \
		cd $(PACKAGE_DIR) && zip -r $(PACKAGE_NAME).zip $(PACKAGE_NAME)/; \
	else \
		cd $(PACKAGE_DIR) && tar -czf $(PACKAGE_NAME).tar.gz $(PACKAGE_NAME)/; \
	fi
	@rm -rf $(PACKAGE_DIR)/$(PACKAGE_NAME)
endef

# Generate checksums for packages
.PHONY: checksums
checksums: package
	@echo "Generating checksums..."
	@cd $(PACKAGE_DIR) && find . -name "*.tar.gz" -o -name "*.zip" | xargs shasum -a 256 > checksums.txt
	@echo "Checksums generated in $(PACKAGE_DIR)/checksums.txt"

# Create GitHub release assets
.PHONY: release-assets
release-assets: checksums
	@echo "Release assets created in $(PACKAGE_DIR)/"
	@ls -la $(PACKAGE_DIR)/

# Install development dependencies
.PHONY: deps
deps:
	@echo "Installing development dependencies..."
	@command -v golangci-lint >/dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@command -v goreleaser >/dev/null 2>&1 || go install github.com/goreleaser/goreleaser@latest

# Verify build environment
.PHONY: verify-env
verify-env:
	@echo "Verifying build environment..."
	@echo "Go version: $(GO_VERSION)"
	@echo "Git version: $(shell git --version 2>/dev/null || echo 'Git not found')"
	@echo "Current branch: $(shell git branch --show-current 2>/dev/null || echo 'Not a git repository')"
	@echo "Last commit: $(COMMIT)"
	@echo "Version: $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(DIST_DIR) $(RELEASE_DIR) $(PACKAGE_DIR) coverage.out coverage.html

# Deep clean including Go module cache
.PHONY: clean-all
clean-all: clean
	@echo "Deep cleaning..."
	go clean -modcache
	go clean -cache

# Run the application
.PHONY: run
run: build
	./$(BUILD_DIR)/$(APP_NAME)

# Show help
.PHONY: help
help:
	@echo "SAI CLI Build System"
	@echo "==================="
	@echo ""
	@echo "Development targets:"
	@echo "  build         - Build for current platform"
	@echo "  build-dev     - Build with race detection for development"
	@echo "  run           - Build and run the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  tidy          - Tidy dependencies"
	@echo ""
	@echo "Cross-platform build targets:"
	@echo "  build-all     - Build for all supported platforms"
	@echo "  build-release - Build optimized release binaries"
	@echo ""
	@echo "Release and packaging targets:"
	@echo "  package       - Create distribution packages (tar.gz/zip)"
	@echo "  checksums     - Generate SHA256 checksums for packages"
	@echo "  release-assets - Create complete release assets"
	@echo ""
	@echo "Installation targets:"
	@echo "  install       - Install to /usr/local/bin"
	@echo "  uninstall     - Remove from /usr/local/bin"
	@echo ""
	@echo "Utility targets:"
	@echo "  deps          - Install development dependencies"
	@echo "  verify-env    - Verify build environment"
	@echo "  clean         - Clean build artifacts"
	@echo "  clean-all     - Deep clean including Go caches"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Supported platforms: $(PLATFORMS)"
	@echo "Current version: $(VERSION)"

# Default help target
.DEFAULT_GOAL := help