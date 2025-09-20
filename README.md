# SAI - Software Action Interface

[![CI](https://github.com/example42/sai/workflows/CI/badge.svg)](https://github.com/example42/sai/actions)
[![Release](https://github.com/example42/sai/workflows/Release/badge.svg)](https://github.com/example42/sai/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/example42/sai)](https://goreportcard.com/report/github.com/example42/sai)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

SAI is a lightweight CLI tool for executing software management actions using provider-based configurations. The core philosophy is **"Do everything on every software on every OS"** through a unified interface.

## 🚀 Quick Start

### Installation

#### Quick Install (Recommended)

**Unix/Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/example42/sai/main/scripts/install.sh | bash
```

**Windows PowerShell:**
```powershell
iwr -useb https://raw.githubusercontent.com/example42/sai/main/scripts/install.ps1 | iex
```

#### Package Managers

**Homebrew (macOS/Linux):**
```bash
brew install example42/tap/sai
```

**Scoop (Windows):**
```powershell
scoop bucket add example42 https://github.com/example42/scoop-bucket
scoop install sai
```

#### Manual Installation

1. Download the latest release from [GitHub Releases](https://github.com/example42/sai/releases)
2. Extract the archive
3. Move the binary to a directory in your PATH

### First Steps

```bash
# Install software (auto-detects best provider)
sai install nginx

# Start a service
sai start nginx

# Check service status
sai status nginx

# View logs
sai logs nginx

# Get help
sai --help
```

## 📖 Features

### Universal Software Management
- **Install/Uninstall**: `sai install nginx`, `sai uninstall nginx`
- **Upgrade**: `sai upgrade nginx`
- **Search**: `sai search nginx`
- **Information**: `sai info nginx`, `sai version nginx`
- **List**: `sai list` (all installed software)

### Service Management
- **Control**: `sai start nginx`, `sai stop nginx`, `sai restart nginx`
- **Boot Management**: `sai enable nginx`, `sai disable nginx`
- **Status**: `sai status nginx`
- **Configuration**: `sai config nginx`

### System Monitoring
- **Logs**: `sai logs nginx` or `sai logs` (system logs)
- **Performance**: `sai cpu nginx`, `sai memory nginx`, `sai io nginx`
- **Health**: `sai check nginx`

### Advanced Operations
- **Batch Operations**: `sai apply actions.yaml`
- **System Statistics**: `sai stats`
- **Repository Management**: `sai saidata`

## 🔧 Core Concepts

### Provider-Based Architecture

SAI uses a **dynamic provider system** where all provider logic is defined in YAML files rather than hardcoded implementations:

- **Zero-Code Provider Addition**: Add new providers by creating YAML files
- **Runtime Flexibility**: Modify provider behavior without recompiling
- **Community Contributions**: Non-Go developers can contribute providers
- **Cross-Platform**: Automatic provider detection and prioritization

### Intelligent Defaults

When software-specific configuration (saidata) isn't available, SAI generates intelligent defaults:

- **Package Names**: Uses software name as default
- **Service Names**: Uses software name as default service
- **Configuration Paths**: Generates standard paths like `/etc/{software}/`
- **Safety Validation**: Verifies resources exist before execution

### Hierarchical Configuration

SAI supports OS-specific overrides with automatic environment detection:

```
software/ap/apache/
├── default.yaml              # Base configuration
├── ubuntu/22.04.yaml         # Ubuntu 22.04 specific
├── centos/8.yaml             # CentOS 8 specific
└── macos/13.yaml             # macOS 13 specific
```

## 📋 Usage Examples

### Software Management

```bash
# Install with automatic provider detection
sai install docker

# Install with specific provider
sai install docker --provider apt

# Install with confirmation bypass
sai install docker --yes

# Dry run (show what would be executed)
sai install docker --dry-run

# Search across all providers
sai search docker

# Get detailed information
sai info docker

# Check versions across providers
sai version docker

# Upgrade to latest version
sai upgrade docker

# Uninstall software
sai uninstall docker
```

### Service Management

```bash
# Start service
sai start apache

# Stop service
sai stop apache

# Restart service
sai restart apache

# Enable at boot
sai enable apache

# Disable at boot
sai disable apache

# Check status
sai status apache

# View configuration files
sai config apache

# Check if service is working
sai check apache
```

### Monitoring and Troubleshooting

```bash
# View service logs
sai logs nginx

# View system logs
sai logs

# Monitor CPU usage
sai cpu nginx

# Monitor memory usage
sai memory nginx

# Monitor I/O statistics
sai io nginx

# System-wide monitoring
sai cpu
sai memory
sai io
```

### Batch Operations

Create an `actions.yaml` file:

```yaml
version: "0.1"
actions:
  - action: install
    software: nginx
    provider: apt
  - action: start
    software: nginx
  - action: enable
    software: nginx
```

Execute batch operations:

```bash
sai apply actions.yaml
```

### Global Options

```bash
# Use custom configuration
sai install nginx --config /path/to/config.yaml

# Verbose output
sai install nginx --verbose

# Quiet mode
sai install nginx --quiet

# JSON output
sai list --json

# Force specific provider
sai install nginx --provider docker

# Auto-confirm all prompts
sai install nginx --yes

# Show commands without executing
sai install nginx --dry-run
```

## 🏗️ Building from Source

### Prerequisites

- Go 1.21 or later
- Git

### Build Commands

```bash
# Clone repository
git clone https://github.com/example42/sai.git
cd sai

# Build for current platform
make build

# Build for all platforms
make build-all

# Build optimized release binaries
make build-release

# Create distribution packages
make package

# Generate checksums
make checksums

# Install locally
make install

# Run tests
make test

# Run linter
make lint

# Clean build artifacts
make clean
```

### Development

```bash
# Build with race detection
make build-dev

# Run with verbose output
make run -- --verbose

# Install development dependencies
make deps

# Verify build environment
make verify-env
```

## 📁 Project Structure

```
├── cmd/sai/                   # Main application entry point
├── internal/                  # Private application code
│   ├── cli/                   # CLI command implementations
│   ├── action/                # Action management and orchestration
│   ├── provider/              # Provider loading and management
│   ├── saidata/               # Software data management
│   ├── template/              # Template engine and functions
│   ├── executor/              # Command execution
│   ├── config/                # Configuration management
│   ├── logger/                # Structured logging
│   └── ...                    # Other internal packages
├── docs/                      # Documentation and examples
│   ├── saidata_samples/       # Example software configurations
│   └── *.md                   # Documentation files
├── providers/                 # Provider implementation files
│   ├── specialized/           # Specialized operational providers
│   └── *.yaml                 # Standard provider definitions
├── schemas/                   # JSON Schema validation files
├── scripts/                   # Installation and utility scripts
├── .github/workflows/         # CI/CD workflows
├── Makefile                   # Build configuration
├── Dockerfile                 # Container build configuration
├── .goreleaser.yml           # Release automation configuration
└── go.mod                     # Go module definition
```

## 🔌 Supported Providers

### Package Managers
- **apt** (Debian/Ubuntu)
- **brew** (macOS/Linux)
- **dnf** (Fedora/RHEL)
- **yum** (CentOS/RHEL)
- **pacman** (Arch Linux)
- **zypper** (openSUSE)
- **apk** (Alpine Linux)
- **pkg** (FreeBSD)
- **choco** (Windows)
- **winget** (Windows)
- **scoop** (Windows)

### Container Platforms
- **docker** (Docker containers)
- **helm** (Kubernetes packages)

### Language Package Managers
- **npm** (Node.js)
- **pip** (Python)
- **gem** (Ruby)
- **cargo** (Rust)
- **go** (Go modules)
- **maven** (Java)
- **gradle** (Java/Kotlin)
- **composer** (PHP)
- **nuget** (.NET)

### Specialized Providers
- **Security**: Vulnerability scanning, SBOM generation
- **Debugging**: GDB integration, performance profiling
- **Operations**: Backup/restore, network analysis
- **Monitoring**: Resource usage, performance metrics

## 🌍 Platform Support

SAI works on:
- **Linux** (all major distributions)
- **macOS** (Intel and Apple Silicon)
- **Windows** (x64 and ARM64)

Automatic environment detection includes:
- Platform identification (linux, darwin, windows)
- OS distribution detection (ubuntu, centos, fedora, etc.)
- Version resolution (22.04, 8, 13.0, etc.)
- Architecture detection (amd64, arm64)

## ⚙️ Configuration

### Configuration File

SAI looks for configuration files in:
1. `--config` flag path
2. `./sai.yaml`
3. `~/.config/sai/config.yaml`
4. `/etc/sai/config.yaml`

Example configuration:

```yaml
saidata_repository: "https://github.com/example42/saidata.git"
default_provider: ""
provider_priority:
  apt: 100
  brew: 90
  docker: 80
timeout: 300s
log_level: "info"

confirmations:
  install: true
  uninstall: true
  system_changes: true
  info_commands: false

output:
  provider_color: "blue"
  command_style: "bold"
  success_color: "green"
  error_color: "red"
  show_commands: true
  show_exit_codes: true

repository:
  git_url: "https://github.com/example42/saidata.git"
  local_path: "~/.cache/sai/saidata"
  update_interval: "24h"
  offline_mode: false
```

### Environment Variables

- `SAI_CONFIG`: Configuration file path
- `SAI_PROVIDER`: Default provider
- `SAI_VERBOSE`: Enable verbose output
- `SAI_DRY_RUN`: Enable dry-run mode
- `SAI_YES`: Auto-confirm prompts
- `SAI_QUIET`: Enable quiet mode

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Quick Contribution Steps

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Run `make test lint`
6. Submit a pull request

### Adding Providers

Create a new provider by adding a YAML file to the `providers/` directory:

```yaml
version: "1.0"
provider:
  name: "my-provider"
  type: "package_manager"
  platforms: ["linux", "darwin"]
  executable: "my-package-manager"
actions:
  install:
    description: "Install software package"
    template: "my-package-manager install {{sai_package}}"
    validation:
      command: "my-package-manager list | grep {{sai_package}}"
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Cobra](https://github.com/spf13/cobra) for CLI framework
- [Viper](https://github.com/spf13/viper) for configuration management
- All contributors and community members

## 📞 Support

- 📖 [Documentation](https://github.com/example42/sai/wiki)
- 🐛 [Issue Tracker](https://github.com/example42/sai/issues)
- 💬 [Discussions](https://github.com/example42/sai/discussions)
- 📧 [Email](mailto:sai@example42.com)

---

**SAI CLI** - Do everything on every software on every OS 🚀