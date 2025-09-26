# Best Practices for Alternative Installation Providers

This document outlines best practices for configuring and using SAI's alternative installation providers (source, binary, script) to ensure secure, reliable, and maintainable software installations.

## Table of Contents

- [General Principles](#general-principles)
- [Source Provider Best Practices](#source-provider-best-practices)
- [Binary Provider Best Practices](#binary-provider-best-practices)
- [Script Provider Best Practices](#script-provider-best-practices)
- [Security Best Practices](#security-best-practices)
- [Configuration Management](#configuration-management)
- [Testing and Validation](#testing-and-validation)
- [Troubleshooting Guidelines](#troubleshooting-guidelines)

## General Principles

### 1. Security First

Always prioritize security in your configurations:

```yaml
# ✓ Good: Always include checksums
sources:
  - checksum: "sha256:verified_checksum_here"

binaries:
  - checksum: "sha256:verified_checksum_here"

scripts:
  - checksum: "sha256:verified_checksum_here"

# ✗ Bad: No checksum verification
sources:
  - url: "https://example.com/software.tar.gz"
    # Missing checksum - security risk
```

### 2. Version Pinning

Always specify exact versions for reproducible installations:

```yaml
# ✓ Good: Specific version
sources:
  - version: "2.1.0"

# ✗ Bad: Vague version specifiers
sources:
  - version: "latest"  # Unpredictable
  - version: "stable"  # Changes over time
```

### 3. Use HTTPS URLs

Always use secure URLs for downloads:

```yaml
# ✓ Good: HTTPS URLs
sources:
  - url: "https://secure.example.com/software-{{version}}.tar.gz"

# ✗ Bad: HTTP URLs
sources:
  - url: "http://example.com/software-{{version}}.tar.gz"  # Insecure
```

### 4. Comprehensive Documentation

Document your configurations thoroughly:

```yaml
version: "0.2"
metadata:
  name: "nginx"
  description: "High-performance web server and reverse proxy"
  documentation: "https://nginx.org/en/docs/"
  
sources:
  - name: "main"
    # Purpose: Build nginx with custom SSL and HTTP/2 support
    # Rationale: Package manager version lacks required modules
    url: "http://nginx.org/download/nginx-{{version}}.tar.gz"
    version: "1.24.0"
    build_system: "autotools"
```

## Source Provider Best Practices

### 1. Prerequisite Management

Always specify complete prerequisites:

```yaml
# ✓ Good: Complete prerequisite list
sources:
  - prerequisites:
      - "build-essential"      # Compiler toolchain
      - "libssl-dev"          # SSL development headers
      - "libpcre3-dev"        # PCRE development headers
      - "zlib1g-dev"          # Compression library headers
      - "pkg-config"          # Build configuration tool

# ✗ Bad: Incomplete prerequisites
sources:
  - prerequisites:
      - "gcc"  # Missing many required dependencies
```

### 2. Build System Configuration

Configure build systems appropriately:

```yaml
# ✓ Good: Proper autotools configuration
sources:
  - build_system: "autotools"
    configure_args:
      - "--prefix=/usr/local"
      - "--with-http_ssl_module"
      - "--with-http_v2_module"
      - "--with-http_realip_module"
      - "--with-pcre"
      - "--enable-shared"

# ✓ Good: Proper CMake configuration
sources:
  - build_system: "cmake"
    configure_args:
      - "-DCMAKE_BUILD_TYPE=Release"
      - "-DCMAKE_INSTALL_PREFIX=/usr/local"
      - "-DENABLE_SSL=ON"
      - "-DENABLE_TESTS=OFF"
```

### 3. Build Environment

Set appropriate build environment:

```yaml
# ✓ Good: Optimized build environment
sources:
  - environment:
      CC: "gcc-9"
      CXX: "g++-9"
      CFLAGS: "-O2 -march=native"
      CXXFLAGS: "-O2 -march=native"
      MAKEFLAGS: "-j$(nproc)"

# ✗ Bad: No environment optimization
sources:
  - environment: {}  # Missing optimization flags
```

### 4. Installation Paths

Use standard installation paths:

```yaml
# ✓ Good: Standard paths
sources:
  - install_prefix: "/usr/local"        # System-wide installation
  # or
  - install_prefix: "/opt/{{metadata.name}}"  # Application-specific

# ✗ Bad: Non-standard paths
sources:
  - install_prefix: "/random/path"      # Unpredictable location
```

### 5. Validation Commands

Provide meaningful validation:

```yaml
# ✓ Good: Comprehensive validation
sources:
  - custom_commands:
      validation: |
        which nginx && 
        nginx -t && 
        nginx -V 2>&1 | grep -q "http_ssl_module"

# ✗ Bad: Minimal validation
sources:
  - custom_commands:
      validation: "which nginx"  # Doesn't verify functionality
```

## Binary Provider Best Practices

### 1. URL Templating

Use proper OS/architecture templating:

```yaml
# ✓ Good: Comprehensive templating
binaries:
  - url: "https://releases.hashicorp.com/terraform/{{version}}/terraform_{{version}}_{{os}}_{{arch}}.zip"
    # Supports: linux_amd64, darwin_amd64, windows_amd64, etc.

# ✗ Bad: Hardcoded platform
binaries:
  - url: "https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_linux_amd64.zip"
    # Only works on Linux amd64
```

### 2. Archive Configuration

Configure archive extraction properly:

```yaml
# ✓ Good: Proper archive handling
binaries:
  - archive:
      format: "zip"                    # Explicit format
      extract_path: "terraform"       # Specific file to extract
      strip_prefix: ""                # No prefix to strip

# ✓ Good: Complex archive structure
binaries:
  - archive:
      format: "tar.gz"
      strip_prefix: "app-{{version}}/" # Remove version directory
      extract_path: "bin/app"          # Extract specific binary
```

### 3. Installation Configuration

Set proper installation parameters:

```yaml
# ✓ Good: Complete installation config
binaries:
  - executable: "terraform"
    install_path: "/usr/local/bin"
    permissions: "0755"
    
# ✗ Bad: Missing configuration
binaries:
  - executable: "terraform"
    # Missing install_path and permissions
```

### 4. Checksum Management

Implement proper checksum verification:

```yaml
# ✓ Good: SHA256 checksum
binaries:
  - checksum: "sha256:a1b2c3d4e5f6789abcdef..."

# ✓ Good: Automatic checksum download
binaries:
  - checksum_url: "https://releases.hashicorp.com/terraform/{{version}}/terraform_{{version}}_SHA256SUMS"
    checksum_pattern: "terraform_{{version}}_{{os}}_{{arch}}.zip"

# ✗ Bad: MD5 checksum (less secure)
binaries:
  - checksum: "md5:1a2b3c4d5e6f..."
```

## Script Provider Best Practices

### 1. Script Security

Implement comprehensive security measures:

```yaml
# ✓ Good: Secure script configuration
scripts:
  - url: "https://get.docker.com/"      # Official HTTPS URL
    checksum: "sha256:verified_hash"    # Integrity verification
    interpreter: "bash"                 # Explicit interpreter
    timeout: 600                        # Reasonable timeout

# ✗ Bad: Insecure configuration
scripts:
  - url: "http://random-site.com/install.sh"  # HTTP, unofficial
    # Missing checksum and timeout
```

### 2. Environment Control

Carefully manage script environment:

```yaml
# ✓ Good: Controlled environment
scripts:
  - environment:
      DEBIAN_FRONTEND: "noninteractive"    # Prevent interactive prompts
      PATH: "/usr/local/bin:/usr/bin:/bin" # Controlled PATH
      HOME: "/tmp/sai-script-home"         # Isolated home directory
      LANG: "C.UTF-8"                      # Consistent locale

# ✗ Bad: Uncontrolled environment
scripts:
  - environment:
      # Inherits all environment variables - potential security risk
```

### 3. Interactive Handling

Properly handle interactive scripts:

```yaml
# ✓ Good: Automated responses
scripts:
  - auto_confirm: true
    confirm_responses: |
      y
      /usr/local
      stable
      admin@example.com
    timeout: 900

# ✓ Good: Expect script for complex interactions
scripts:
  - expect_script: |
      #!/usr/bin/expect -f
      spawn bash {{script_file}} {{arguments}}
      expect "Continue?" { send "y\r" }
      expect "Directory:" { send "/opt/app\r" }
      expect eof
```

### 4. Rollback Procedures

Implement proper cleanup:

```yaml
# ✓ Good: Comprehensive rollback
scripts:
  - custom_commands:
      uninstall: |
        systemctl stop docker
        apt-get remove -y docker-ce docker-ce-cli containerd.io
        rm -rf /var/lib/docker
        userdel docker

# ✗ Bad: No rollback procedure
scripts:
  - custom_commands: {}  # No cleanup defined
```

## Security Best Practices

### 1. Checksum Verification

Always verify integrity:

```yaml
# ✓ Best: SHA256 checksums for all providers
sources:
  - checksum: "sha256:source_checksum_here"
binaries:
  - checksum: "sha256:binary_checksum_here"
scripts:
  - checksum: "sha256:script_checksum_here"
```

### 2. HTTPS Enforcement

Use secure protocols:

```yaml
# ✓ Good: HTTPS URLs
- url: "https://secure.example.com/file"

# ✗ Bad: HTTP URLs
- url: "http://example.com/file"  # Vulnerable to MITM attacks
```

### 3. User Consent

Require explicit user approval:

```bash
# ✓ Good: User must confirm
sai install docker --provider script
# Prompts: "Execute script from https://get.docker.com/? [y/N]"

# ✓ Good: Automated with explicit flag
sai install docker --provider script --yes
```

### 4. Minimal Privileges

Use least privilege principle:

```yaml
# ✓ Good: Specific permissions
binaries:
  - permissions: "0755"  # Executable, not writable by others

# ✗ Bad: Excessive permissions
binaries:
  - permissions: "0777"  # World-writable - security risk
```

### 5. Sandboxing

Consider isolation:

```yaml
# ✓ Good: Isolated working directory
scripts:
  - working_dir: "/tmp/sai-script-{{metadata.name}}"
    environment:
      HOME: "/tmp/sai-script-{{metadata.name}}"
```

## Configuration Management

### 1. OS-Specific Overrides

Use OS-specific configurations appropriately:

```yaml
# software/do/docker/default.yaml
scripts:
  - name: "main"
    url: "https://get.docker.com/"
    interpreter: "bash"

# software/do/docker/ubuntu/22.04.yaml
providers:
  script:
    scripts:
      - name: "main"
        environment:
          DEBIAN_FRONTEND: "noninteractive"
          APT_KEY_DONT_WARN_ON_DANGEROUS_USAGE: "1"

# software/do/docker/centos/8.yaml
providers:
  script:
    scripts:
      - name: "main"
        environment:
          YUM_OPTS: "-y"
```

### 2. Provider Fallbacks

Configure multiple installation methods:

```yaml
# Provide multiple installation options
packages:
  - name: "nginx"  # Package manager fallback

sources:
  - name: "main"   # Source build option
    url: "http://nginx.org/download/nginx-{{version}}.tar.gz"

binaries:
  - name: "main"   # Binary download option
    url: "https://nginx.org/packages/binaries/nginx-{{version}}-{{os}}-{{arch}}.tar.gz"
```

### 3. Version Management

Handle versions consistently:

```yaml
# ✓ Good: Consistent version handling
metadata:
  version: "1.24.0"  # Default version

sources:
  - version: "{{metadata.version}}"  # Use metadata version

binaries:
  - version: "{{metadata.version}}"  # Consistent across providers
```

## Testing and Validation

### 1. Comprehensive Testing

Test all installation methods:

```bash
# Test each provider
sai install nginx --provider source --dry-run
sai install nginx --provider binary --dry-run
sai install nginx --provider script --dry-run

# Test on different platforms
sai install nginx --provider source --verbose  # Ubuntu
sai install nginx --provider source --verbose  # CentOS
sai install nginx --provider source --verbose  # macOS
```

### 2. Validation Commands

Implement thorough validation:

```yaml
# ✓ Good: Multi-step validation
sources:
  - custom_commands:
      validation: |
        # Check binary exists
        which nginx || exit 1
        # Check configuration syntax
        nginx -t || exit 1
        # Check required modules
        nginx -V 2>&1 | grep -q "http_ssl_module" || exit 1
        # Check version
        nginx -v 2>&1 | grep -q "{{version}}" || exit 1
```

### 3. Rollback Testing

Test rollback procedures:

```bash
# Test installation and rollback
sai install nginx --provider source
sai uninstall nginx --provider source
# Verify complete removal
```

## Troubleshooting Guidelines

### 1. Enable Verbose Logging

Use verbose output for debugging:

```bash
# Enable detailed logging
sai install nginx --provider source --verbose

# Use dry-run to see commands
sai install nginx --provider source --dry-run
```

### 2. Check Prerequisites

Verify all prerequisites are met:

```bash
# Check build tools
gcc --version
make --version
cmake --version

# Check libraries
pkg-config --exists openssl
pkg-config --exists libpcre
```

### 3. Validate URLs

Test URLs manually:

```bash
# Test source URL
curl -I "http://nginx.org/download/nginx-1.24.0.tar.gz"

# Test binary URL
curl -I "https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_linux_amd64.zip"

# Test script URL
curl -I "https://get.docker.com/"
```

### 4. Check Permissions

Verify file and directory permissions:

```bash
# Check installation directory
ls -la /usr/local/bin/

# Check working directory
ls -la /tmp/sai-build-nginx/
```

### 5. Review Logs

Examine detailed logs:

```bash
# View installation logs
tail -f /var/log/sai/nginx-source.log
tail -f /var/log/sai/terraform-binary.log
tail -f /var/log/sai/docker-script.log
```

## Performance Optimization

### 1. Parallel Builds

Optimize build performance:

```yaml
# ✓ Good: Parallel compilation
sources:
  - build_args: ["-j$(nproc)"]
    environment:
      MAKEFLAGS: "-j$(nproc)"
```

### 2. Build Caching

Enable build caching:

```yaml
# Enable ccache for faster rebuilds
sources:
  - environment:
      CC: "ccache gcc"
      CXX: "ccache g++"
```

### 3. Download Optimization

Optimize downloads:

```yaml
# Use fastest mirrors
binaries:
  - url: "https://cdn.example.com/fast-mirror/{{file}}"
    
# Enable resume for large files
scripts:
  - download_options:
      resume: true
      timeout: 1800
```

## Conclusion

Following these best practices ensures:

- **Security**: Proper verification and secure protocols
- **Reliability**: Consistent and reproducible installations
- **Maintainability**: Clear documentation and configuration
- **Performance**: Optimized build and download processes
- **Troubleshooting**: Comprehensive logging and validation

Alternative installation providers become as reliable and secure as traditional package managers when these practices are followed consistently.

For more information, see:
- [Source Build Tutorial](source_build_tutorial.md)
- [Binary Installation Guide](binary_installation_guide.md)
- [Script Installation Guide](script_installation_guide.md)
- [Provider Development Guide](PROVIDER_DEVELOPMENT.md)