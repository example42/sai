#!/bin/bash
# SAI CLI Installation Script for Unix-like systems (Linux, macOS)

set -e

# Configuration
REPO_OWNER="example42"
REPO_NAME="sai"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="sai"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect platform and architecture
detect_platform() {
    local os
    local arch
    
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        *)          log_error "Unsupported operating system: $(uname -s)"; exit 1 ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        *)              log_error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac
    
    echo "${os}-${arch}"
}

# Get latest release version from GitHub
get_latest_version() {
    local version
    version=$(curl -s "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        log_error "Failed to get latest version from GitHub"
        exit 1
    fi
    
    echo "$version"
}

# Download and install SAI
install_sai() {
    local platform="$1"
    local version="$2"
    local download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${version}/${BINARY_NAME}-${version}-${platform}.tar.gz"
    local temp_dir
    
    log_info "Installing SAI CLI version ${version} for ${platform}..."
    
    # Create temporary directory
    temp_dir=$(mktemp -d)
    trap "rm -rf ${temp_dir}" EXIT
    
    # Download release
    log_info "Downloading from ${download_url}..."
    if ! curl -L -o "${temp_dir}/${BINARY_NAME}.tar.gz" "$download_url"; then
        log_error "Failed to download SAI CLI"
        exit 1
    fi
    
    # Extract archive
    log_info "Extracting archive..."
    if ! tar -xzf "${temp_dir}/${BINARY_NAME}.tar.gz" -C "$temp_dir"; then
        log_error "Failed to extract archive"
        exit 1
    fi
    
    # Find the binary in the extracted directory
    local binary_path
    binary_path=$(find "$temp_dir" -name "$BINARY_NAME" -type f | head -1)
    
    if [ ! -f "$binary_path" ]; then
        log_error "Binary not found in archive"
        exit 1
    fi
    
    # Install binary
    log_info "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
    if [ -w "$INSTALL_DIR" ]; then
        cp "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo cp "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    log_success "SAI CLI installed successfully!"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local installed_version
        installed_version=$("$BINARY_NAME" version --short 2>/dev/null || echo "unknown")
        log_success "SAI CLI is installed and available in PATH"
        log_info "Installed version: ${installed_version}"
        log_info "Run 'sai --help' to get started"
    else
        log_warning "SAI CLI installed but not found in PATH"
        log_info "You may need to add ${INSTALL_DIR} to your PATH"
        log_info "Or run directly: ${INSTALL_DIR}/${BINARY_NAME}"
    fi
}

# Main installation flow
main() {
    log_info "SAI CLI Installation Script"
    log_info "=========================="
    
    # Check if already installed
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local current_version
        current_version=$("$BINARY_NAME" version --short 2>/dev/null || echo "unknown")
        log_warning "SAI CLI is already installed (version: ${current_version})"
        
        read -p "Do you want to reinstall? [y/N]: " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Installation cancelled"
            exit 0
        fi
    fi
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    log_info "Detected platform: ${platform}"
    
    # Get latest version
    local version
    if [ -n "${SAI_VERSION:-}" ]; then
        version="$SAI_VERSION"
        log_info "Using specified version: ${version}"
    else
        log_info "Fetching latest release version..."
        version=$(get_latest_version)
        log_info "Latest version: ${version}"
    fi
    
    # Install SAI
    install_sai "$platform" "$version"
    
    # Verify installation
    verify_installation
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "SAI CLI Installation Script"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --version, -v  Show script version"
        echo ""
        echo "Environment variables:"
        echo "  SAI_VERSION    Install specific version (default: latest)"
        echo ""
        echo "Examples:"
        echo "  $0                    # Install latest version"
        echo "  SAI_VERSION=v1.0.0 $0 # Install specific version"
        exit 0
        ;;
    --version|-v)
        echo "SAI CLI Installation Script v1.0.0"
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac