# Binary Installation Guide with SAI

This guide demonstrates how to use SAI's binary provider to download and install pre-compiled binaries, offering fast deployment without compilation overhead.

## Overview

The binary provider enables you to:
- Download pre-compiled binaries with automatic OS/architecture detection
- Install software quickly without build dependencies or compilation time
- Verify binary integrity with checksum validation
- Handle various archive formats automatically (zip, tar.gz, tar.bz2, etc.)
- Install the latest versions often not available in package repositories

## Prerequisites

The binary provider requires basic download utilities:

### Ubuntu/Debian
```bash
sudo apt update
sudo apt install wget curl unzip tar
```

### CentOS/RHEL/Rocky
```bash
sudo yum install wget curl unzip tar
```

### macOS
```bash
# Usually pre-installed, but can install via Homebrew if needed
brew install wget
```

### Windows
```bash
# PowerShell (usually pre-installed)
# Or install via Chocolatey
choco install wget curl 7zip
```

## Basic Binary Installation

### Example 1: Installing Terraform

Terraform provides pre-compiled binaries for all major platforms:

```bash
# Install terraform binary
sai install terraform --provider binary

# Check what will be downloaded (dry run)
sai install terraform --provider binary --dry-run

# Install with verbose output to see download progress
sai install terraform --provider binary --verbose
```

**What happens during installation:**
1. **Platform Detection**: SAI detects your OS and architecture
2. **URL Resolution**: Constructs download URL with OS/arch placeholders
3. **Download**: Downloads the binary/archive from the resolved URL
4. **Verification**: Verifies checksum if provided for security
5. **Extraction**: Extracts binary from archive if needed
6. **Installation**: Places binary in PATH with correct permissions

### Example 2: Installing with Version Specification

```bash
# Install specific version
sai install terraform --provider binary --version 1.5.7

# Install latest version (default)
sai install terraform --provider binary --version latest
```

## Binary Configuration Examples

### Example 3: Simple Binary Download

For a simple binary download without archives:

```yaml
# ~/.sai/custom/kubectl-binary.yaml
version: "0.2"
metadata:
  name: "kubectl"
  description: "Kubernetes command-line tool"

binaries:
  - name: "main"
    url: "https://dl.k8s.io/release/v{{version}}/bin/{{os}}/{{arch}}/kubectl"
    version: "1.28.0"
    executable: "kubectl"
    install_path: "/usr/local/bin"
    permissions: "0755"
    checksum: "sha256:a1b2c3d4e5f6..."
```

### Example 4: Archive-based Binary

For binaries distributed in archives:

```yaml
# Example: Terraform binary in zip archive
binaries:
  - name: "main"
    url: "https://releases.hashicorp.com/terraform/{{version}}/terraform_{{version}}_{{os}}_{{arch}}.zip"
    version: "1.5.7"
    executable: "terraform"
    install_path: "/usr/local/bin"
    archive:
      format: "zip"
      extract_path: "terraform"  # Path within archive
    checksum: "sha256:verified_checksum_here"
```

### Example 5: Complex Archive Structure

For binaries with complex archive structures:

```yaml
# Example: Binary in nested archive structure
binaries:
  - name: "main"
    url: "https://github.com/user/project/releases/download/v{{version}}/project-{{version}}-{{os}}-{{arch}}.tar.gz"
    version: "2.1.0"
    executable: "project"
    archive:
      format: "tar.gz"
      strip_prefix: "project-2.1.0/"  # Remove this prefix during extraction
      extract_path: "bin/project"      # Path to binary within archive
    install_path: "/opt/project/bin"
    permissions: "0755"
```

## OS and Architecture Templating

SAI automatically substitutes OS and architecture placeholders in URLs:

### Supported Placeholders

- `{{os}}`: Operating system (linux, darwin, windows)
- `{{arch}}`: Architecture (amd64, arm64, 386, arm)
- `{{version}}`: Software version
- `{{platform}}`: Alternative to {{os}} for some providers

### Platform Mapping Examples

| Your System | {{os}} | {{arch}} | Example URL |
|-------------|--------|----------|-------------|
| Ubuntu 22.04 x64 | linux | amd64 | `app_1.0.0_linux_amd64.tar.gz` |
| macOS M1 | darwin | arm64 | `app_1.0.0_darwin_arm64.tar.gz` |
| Windows x64 | windows | amd64 | `app_1.0.0_windows_amd64.zip` |
| Raspberry Pi | linux | arm64 | `app_1.0.0_linux_arm64.tar.gz` |

### Custom Platform Mapping

For providers with non-standard naming:

```yaml
binaries:
  - name: "main"
    url: "https://example.com/app-{{version}}-{{custom_os}}-{{custom_arch}}.zip"
    platform_mapping:
      os:
        linux: "Linux"
        darwin: "macOS"
        windows: "Windows"
      arch:
        amd64: "x86_64"
        arm64: "aarch64"
```

## Archive Format Support

SAI supports multiple archive formats with automatic detection:

### Supported Formats

- **ZIP**: `.zip` files (Windows, cross-platform)
- **TAR.GZ**: `.tar.gz`, `.tgz` files (Unix/Linux)
- **TAR.BZ2**: `.tar.bz2`, `.tbz2` files (Unix/Linux)
- **TAR.XZ**: `.tar.xz` files (Unix/Linux)
- **TAR**: `.tar` files (Unix/Linux)
- **None**: Direct binary downloads

### Archive Configuration

```yaml
binaries:
  - archive:
      format: "zip"                    # Override auto-detection
      strip_prefix: "app-1.0.0/"      # Remove prefix during extraction
      extract_path: "bin/app"          # Specific file to extract
```

## Security and Verification

### Checksum Verification

Always include checksums for security:

```yaml
binaries:
  - checksum: "sha256:a1b2c3d4e5f6789..."  # SHA256 checksum
  # or
  - checksum: "md5:1a2b3c4d5e6f..."        # MD5 checksum (less secure)
```

### Automatic Checksum Download

For providers that publish checksum files:

```yaml
binaries:
  - checksum_url: "https://releases.example.com/app/{{version}}/checksums.txt"
  - checksum_pattern: "{{executable}}"  # Pattern to find in checksum file
```

### HTTPS Enforcement

Always use HTTPS URLs for security:

```yaml
binaries:
  - url: "https://secure.example.com/app.zip"  # ✓ Secure
  # Not: "http://example.com/app.zip"          # ✗ Insecure
```

## Managing Binary Installations

### Checking Installation Status

```bash
# Check if terraform is installed via binary provider
sai status terraform --provider binary

# Get version information
sai version terraform --provider binary

# View installation details
sai info terraform --provider binary
```

### Upgrading Binaries

```bash
# Upgrade to latest version
sai upgrade terraform --provider binary

# Upgrade to specific version
sai upgrade terraform --provider binary --version 1.6.0
```

### Uninstalling Binaries

```bash
# Remove binary installation
sai uninstall terraform --provider binary

# Force removal if uninstall fails
sai uninstall terraform --provider binary --force
```

## Advanced Configuration

### Custom Installation Paths

```yaml
binaries:
  - install_path: "/opt/myapp/bin"      # Custom installation directory
  - executable: "myapp"                # Binary name
  - permissions: "0755"                # File permissions
```

### Environment-Specific Configuration

```yaml
# OS-specific overrides
providers:
  binary:
    binaries:
      - name: "main"
        # Windows-specific configuration
        executable: "app.exe"
        install_path: "C:\\Program Files\\MyApp"
        permissions: "0755"
```

### Multiple Binaries

For software packages with multiple binaries:

```yaml
binaries:
  - name: "main"
    url: "https://example.com/app-{{version}}-{{os}}-{{arch}}.tar.gz"
    executable: "app"
    archive:
      extract_path: "bin/app"
  - name: "cli"
    url: "https://example.com/app-cli-{{version}}-{{os}}-{{arch}}.tar.gz"
    executable: "app-cli"
    archive:
      extract_path: "bin/app-cli"
```

## Troubleshooting Binary Installations

### Common Issues and Solutions

#### 1. Download Failures

**Problem**: Binary download fails
```
Error: Failed to download binary from URL
```

**Solutions**:
```bash
# Check URL accessibility
curl -I "https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_linux_amd64.zip"

# Verify network connectivity
ping releases.hashicorp.com

# Check for proxy issues
export https_proxy=http://proxy.company.com:8080
```

#### 2. Checksum Verification Failures

**Problem**: Downloaded binary fails checksum verification
```
Error: Checksum verification failed
```

**Solutions**:
```bash
# Verify checksum manually
sha256sum downloaded_file.zip
echo "expected_checksum downloaded_file.zip" | sha256sum -c

# Update checksum in configuration
# Check provider's official checksums
```

#### 3. Archive Extraction Issues

**Problem**: Cannot extract binary from archive
```
Error: Failed to extract binary from archive
```

**Solutions**:
```bash
# Check archive format
file downloaded_archive.zip

# Test extraction manually
unzip -l downloaded_archive.zip
tar -tzf downloaded_archive.tar.gz

# Update archive configuration
```

#### 4. Permission Issues

**Problem**: Cannot install binary due to permissions
```
Error: Permission denied writing to /usr/local/bin
```

**Solutions**:
```bash
# Use sudo for system installation
sudo sai install terraform --provider binary

# Or install to user directory
sai install terraform --provider binary --config user-config.yaml
```

### Debugging Binary Issues

Enable verbose output for detailed information:

```bash
# See all download and installation steps
sai install terraform --provider binary --verbose

# Dry run to see what would be executed
sai install terraform --provider binary --dry-run

# Check installation logs
tail -f /var/log/sai/terraform-binary.log
```

## Best Practices

### 1. Version Pinning
Specify exact versions for reproducible installations:
```yaml
binaries:
  - version: "1.5.7"  # Specific version
    # Not: "latest" or "stable"
```

### 2. Checksum Verification
Always include checksums for security:
```yaml
binaries:
  - checksum: "sha256:verified_checksum_from_provider"
```

### 3. HTTPS URLs
Use secure download URLs:
```yaml
binaries:
  - url: "https://secure-releases.example.com/..."
```

### 4. Standard Installation Paths
Use conventional installation directories:
```yaml
binaries:
  - install_path: "/usr/local/bin"     # System-wide (Unix)
  - install_path: "/opt/app/bin"       # Application-specific
  - install_path: "~/.local/bin"      # User-specific
```

### 5. Proper Permissions
Set appropriate file permissions:
```yaml
binaries:
  - permissions: "0755"  # Executable by owner, readable by all
```

### 6. Backup Before Upgrades
```bash
# Create backup before upgrading
sudo cp /usr/local/bin/terraform /usr/local/bin/terraform.backup.$(date +%Y%m%d)
sai upgrade terraform --provider binary
```

## Integration Examples

### Docker Integration
```yaml
# Install Docker Compose binary
binaries:
  - name: "docker-compose"
    url: "https://github.com/docker/compose/releases/download/v{{version}}/docker-compose-{{os}}-{{arch}}"
    version: "2.20.0"
    executable: "docker-compose"
    install_path: "/usr/local/bin"
    permissions: "0755"
```

### Kubernetes Tools
```yaml
# Install kubectl, helm, and k9s
binaries:
  - name: "kubectl"
    url: "https://dl.k8s.io/release/v{{version}}/bin/{{os}}/{{arch}}/kubectl"
  - name: "helm"
    url: "https://get.helm.sh/helm-v{{version}}-{{os}}-{{arch}}.tar.gz"
    archive:
      format: "tar.gz"
      strip_prefix: "{{os}}-{{arch}}/"
      extract_path: "helm"
  - name: "k9s"
    url: "https://github.com/derailed/k9s/releases/download/v{{version}}/k9s_{{os}}_{{arch}}.tar.gz"
```

### Development Tools
```yaml
# Install Go, Node.js binaries
binaries:
  - name: "go"
    url: "https://golang.org/dl/go{{version}}.{{os}}-{{arch}}.tar.gz"
    archive:
      format: "tar.gz"
      extract_path: "go/bin/go"
    install_path: "/usr/local/bin"
```

## Performance Optimization

### Parallel Downloads
For multiple binaries:
```bash
# SAI can download multiple binaries in parallel
sai install kubectl helm k9s --provider binary --parallel
```

### Download Caching
Enable download caching:
```yaml
# Global SAI configuration
cache:
  enabled: true
  directory: "/var/cache/sai"
  retention: "30d"
```

### Resume Downloads
For large binaries:
```yaml
binaries:
  - download_options:
      resume: true
      timeout: 1800  # 30 minutes
      retries: 3
```

## Conclusion

Binary installation with SAI provides:
- **Speed**: Fast installation without compilation
- **Simplicity**: Automatic OS/architecture detection
- **Security**: Checksum verification and HTTPS enforcement
- **Flexibility**: Support for various archive formats
- **Integration**: Seamless integration with SAI's management features

The binary provider makes installing pre-compiled software as simple as package managers while providing the flexibility to install the latest versions and custom builds.

For more information, see:
- [SAI Binary Functions Reference](sai_binary_functions.md)
- [Provider Development Guide](PROVIDER_DEVELOPMENT.md)
- [SAI Synopsis](sai_synopsis.md)