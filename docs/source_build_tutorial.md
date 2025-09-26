# Building Software from Source with SAI

This tutorial demonstrates how to use SAI's source provider to build and install software from source code, providing flexibility and customization beyond traditional package managers.

## Overview

The source provider enables you to:
- Build software from source code with automatic build system detection
- Customize build configurations and installation paths
- Install the latest versions not available in package repositories
- Apply custom patches or modifications during the build process

## Prerequisites

Before building from source, ensure you have the necessary build tools installed:

### Ubuntu/Debian
```bash
sudo apt update
sudo apt install build-essential git wget curl
```

### CentOS/RHEL/Rocky
```bash
sudo yum groupinstall "Development Tools"
sudo yum install git wget curl
```

### macOS
```bash
# Install Xcode Command Line Tools
xcode-select --install

# Install Homebrew (if not already installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

## Basic Source Installation

### Example 1: Building Nginx from Source

Let's build nginx from source with custom SSL and HTTP/2 support:

```bash
# Install nginx from source
sai install nginx --provider source

# Check what will be executed (dry run)
sai install nginx --provider source --dry-run

# Install with verbose output to see build progress
sai install nginx --provider source --verbose
```

**What happens during installation:**
1. **Prerequisites Check**: SAI installs required build tools and dependencies
2. **Source Download**: Downloads nginx source code from official repository
3. **Configuration**: Runs `./configure` with appropriate flags for your system
4. **Compilation**: Builds nginx using `make` with optimal parallel jobs
5. **Installation**: Installs nginx to the configured prefix (typically `/usr/local`)

### Example 2: Building with Custom Configuration

For more control over the build process, you can customize the saidata configuration:

```yaml
# ~/.sai/custom/nginx-source.yaml
version: "0.2"
metadata:
  name: "nginx"
  description: "Custom nginx build"

sources:
  - name: "main"
    url: "http://nginx.org/download/nginx-{{version}}.tar.gz"
    version: "1.24.0"
    build_system: "autotools"
    install_prefix: "/opt/nginx"
    configure_args:
      - "--with-http_ssl_module"
      - "--with-http_v2_module"
      - "--with-http_realip_module"
      - "--with-http_gzip_static_module"
      - "--with-pcre"
      - "--with-file-aio"
    prerequisites:
      - "build-essential"
      - "libssl-dev"
      - "libpcre3-dev"
      - "zlib1g-dev"
```

Then install using your custom configuration:

```bash
sai install nginx --provider source --config ~/.sai/custom/nginx-source.yaml
```

## Advanced Source Builds

### Example 3: Building CMake-based Software

For software using CMake build system:

```yaml
# Example: Building a CMake project
sources:
  - name: "main"
    url: "https://github.com/project/software/archive/v{{version}}.tar.gz"
    version: "2.1.0"
    build_system: "cmake"
    install_prefix: "/usr/local"
    configure_args:
      - "-DCMAKE_BUILD_TYPE=Release"
      - "-DENABLE_TESTS=OFF"
      - "-DENABLE_SSL=ON"
    build_args:
      - "--parallel"
      - "$(nproc)"
```

### Example 4: Custom Build Commands

For projects with unique build requirements:

```yaml
sources:
  - name: "main"
    url: "https://example.com/software-{{version}}.tar.gz"
    version: "1.0.0"
    build_system: "custom"
    custom_commands:
      configure: "./setup.sh --prefix={{install_prefix}}"
      build: "make -j$(nproc) CFLAGS=-O3"
      install: "make install PREFIX={{install_prefix}}"
      validation: "{{install_prefix}}/bin/software --version"
```

## Build System Support

SAI automatically detects and supports multiple build systems:

### Autotools (./configure && make)
```yaml
build_system: "autotools"
configure_args: ["--prefix=/usr/local", "--enable-feature"]
```

### CMake
```yaml
build_system: "cmake"
configure_args: ["-DCMAKE_BUILD_TYPE=Release", "-DENABLE_FEATURE=ON"]
```

### Make (Direct)
```yaml
build_system: "make"
build_args: ["CFLAGS=-O2", "PREFIX=/usr/local"]
```

### Meson
```yaml
build_system: "meson"
configure_args: ["--prefix=/usr/local", "-Dfeature=enabled"]
```

### Ninja
```yaml
build_system: "ninja"
build_args: ["-j$(nproc)"]
```

## Managing Source Installations

### Checking Installation Status
```bash
# Check if nginx is installed via source
sai status nginx --provider source

# Get version information
sai version nginx --provider source

# View installation details
sai info nginx --provider source
```

### Upgrading Source Builds
```bash
# Upgrade to latest version
sai upgrade nginx --provider source

# Upgrade to specific version
sai upgrade nginx --provider source --version 1.25.0
```

### Uninstalling Source Builds
```bash
# Remove source-built software
sai uninstall nginx --provider source

# Force removal if uninstall fails
sai uninstall nginx --provider source --force
```

## Troubleshooting Source Builds

### Common Issues and Solutions

#### 1. Missing Dependencies
**Problem**: Build fails due to missing libraries or headers
```
configure: error: SSL modules require the OpenSSL library
```

**Solution**: Install development packages
```bash
# Ubuntu/Debian
sudo apt install libssl-dev libpcre3-dev zlib1g-dev

# CentOS/RHEL
sudo yum install openssl-devel pcre-devel zlib-devel
```

#### 2. Build Directory Conflicts
**Problem**: Previous build artifacts interfere with new builds

**Solution**: Clean build directory
```bash
# SAI automatically cleans build directories, but you can force cleanup
sai uninstall nginx --provider source
rm -rf /tmp/sai-build-nginx
```

#### 3. Permission Issues
**Problem**: Installation fails due to insufficient permissions

**Solution**: Ensure proper permissions or use sudo
```bash
# Install to user directory
sai install nginx --provider source --config custom-config.yaml

# Or run with sudo for system installation
sudo sai install nginx --provider source
```

#### 4. Build Timeout
**Problem**: Large projects exceed default timeout

**Solution**: Increase timeout in configuration
```yaml
sources:
  - name: "main"
    # ... other config
    timeout: 3600  # 1 hour timeout
```

### Debugging Build Issues

Enable verbose output to see detailed build information:
```bash
# See all build commands and output
sai install nginx --provider source --verbose

# Dry run to see what would be executed
sai install nginx --provider source --dry-run

# Check build logs
tail -f /var/log/sai/nginx-source.log
```

## Best Practices

### 1. Version Pinning
Always specify exact versions for reproducible builds:
```yaml
sources:
  - version: "1.24.0"  # Specific version
    # Not: "latest" or "stable"
```

### 2. Checksum Verification
Include checksums for security:
```yaml
sources:
  - checksum: "sha256:a1b2c3d4e5f6..."
```

### 3. Custom Installation Paths
Use dedicated prefixes for source builds:
```yaml
sources:
  - install_prefix: "/opt/nginx"  # Separate from package manager installs
```

### 4. Environment Variables
Set build environment for consistency:
```yaml
sources:
  - environment:
      CC: "gcc-9"
      CXX: "g++-9"
      CFLAGS: "-O2 -march=native"
```

### 5. Backup Before Upgrades
```bash
# Create backup before upgrading
sudo cp -r /opt/nginx /opt/nginx.backup.$(date +%Y%m%d)
sai upgrade nginx --provider source
```

## Integration with System Services

After building from source, integrate with system service management:

### Create Systemd Service (Linux)
```bash
# SAI can manage services after source installation
sai enable nginx --provider source
sai start nginx --provider source
sai status nginx --provider source
```

### macOS LaunchDaemon
```bash
# On macOS, SAI integrates with launchctl
sai enable nginx --provider source
sai start nginx --provider source
```

## Performance Optimization

### Parallel Builds
SAI automatically uses optimal parallel job counts:
```yaml
sources:
  - build_args: ["-j$(nproc)"]  # Use all available CPU cores
```

### Build Caching
Enable ccache for faster rebuilds:
```bash
# Install ccache
sudo apt install ccache  # Ubuntu/Debian
sudo yum install ccache   # CentOS/RHEL

# SAI will automatically use ccache if available
export CC="ccache gcc"
export CXX="ccache g++"
```

### Optimized Builds
```yaml
sources:
  - environment:
      CFLAGS: "-O3 -march=native -flto"
      CXXFLAGS: "-O3 -march=native -flto"
```

## Security Considerations

### 1. Source Verification
Always verify source integrity:
```yaml
sources:
  - checksum: "sha256:verified_checksum_here"
```

### 2. Trusted Sources
Use official repositories and mirrors:
```yaml
sources:
  - url: "https://nginx.org/download/nginx-{{version}}.tar.gz"  # Official
    # Not: random GitHub forks or unofficial mirrors
```

### 3. Build Isolation
Consider using containers for build isolation:
```bash
# Build in container for security
docker run --rm -v $(pwd):/workspace ubuntu:22.04 bash -c "
  cd /workspace && 
  sai install nginx --provider source
"
```

## Conclusion

Building software from source with SAI provides:
- **Flexibility**: Custom configurations and latest versions
- **Control**: Full control over build process and dependencies
- **Integration**: Seamless integration with SAI's service management
- **Automation**: Automated prerequisite installation and build process
- **Safety**: Rollback capabilities and validation checks

The source provider makes building from source as simple as using package managers while maintaining the flexibility and control that source builds provide.

For more information, see:
- [SAI Source Functions Reference](sai_source_functions.md)
- [Provider Development Guide](PROVIDER_DEVELOPMENT.md)
- [SAI Synopsis](sai_synopsis.md)