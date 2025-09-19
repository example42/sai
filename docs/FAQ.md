# Frequently Asked Questions (FAQ)

## General Questions

### What is SAI CLI?

SAI (Software Action Interface) is a lightweight command-line tool that provides a unified interface for managing software across different operating systems and package managers. The core philosophy is "Do everything on every software on every OS" through a single, consistent interface.

### How is SAI different from other package managers?

SAI is not a package manager itself, but rather a **universal interface** that works with existing package managers. Key differences:

- **Universal**: Works with apt, brew, dnf, docker, npm, pip, and many others
- **Cross-platform**: Same commands work on Linux, macOS, and Windows
- **Provider-based**: Extensible architecture supports any command-line tool
- **Intelligent**: Automatically detects the best provider for your system
- **Consistent**: Same interface for packages, services, containers, and specialized tools

### Is SAI safe to use?

Yes, SAI includes multiple safety features:

- **Dry-run mode**: See what would be executed before running commands
- **Confirmation prompts**: System-changing operations require confirmation
- **Validation**: Checks that resources exist before attempting operations
- **Rollback**: Many operations include automatic rollback on failure
- **Sandboxing**: Uses existing system tools, doesn't modify system files directly

## Installation and Setup

### How do I install SAI?

**Quick install (recommended):**

Unix/Linux/macOS:
```bash
curl -fsSL https://raw.githubusercontent.com/sai-cli/sai/main/scripts/install.sh | bash
```

Windows PowerShell:
```powershell
iwr -useb https://raw.githubusercontent.com/sai-cli/sai/main/scripts/install.ps1 | iex
```

**Package managers:**
```bash
# Homebrew (macOS/Linux)
brew install sai-cli/tap/sai

# Scoop (Windows)
scoop bucket add sai-cli https://github.com/sai-cli/scoop-bucket
scoop install sai
```

### Do I need to configure SAI after installation?

No, SAI works out of the box with intelligent defaults. However, you can customize behavior by creating a configuration file at `~/.config/sai/config.yaml`.

### Can I use SAI without internet access?

Yes, SAI can work offline using:
- Cached saidata repository
- Local provider configurations
- System package managers that work offline

However, initial setup and saidata updates require internet access.

## Usage Questions

### How do I know which providers are available?

```bash
# List all providers
sai stats --providers

# Check specific provider
sai stats --provider apt

# See what providers can install specific software
sai search nginx
```

### Why does SAI ask me to choose a provider?

When multiple providers can install the same software, SAI shows you options so you can choose the best one for your needs. For example, nginx can be installed via:
- System package manager (apt, dnf, brew)
- Container (docker)
- Snap package
- Direct binary

Use `--yes` flag to automatically select the highest priority provider.

### How do I install software without confirmation prompts?

```bash
# Skip all confirmations
sai install nginx --yes

# Or configure globally
echo "confirmations: { install: false }" > ~/.config/sai/config.yaml
```

### Can I see what commands SAI will run before executing them?

Yes, use dry-run mode:

```bash
# See commands without executing
sai install nginx --dry-run

# With verbose output
sai install nginx --dry-run --verbose
```

### How do I force a specific provider?

```bash
# Force specific provider
sai install nginx --provider docker

# Set default provider in config
echo "default_provider: docker" > ~/.config/sai/config.yaml
```

## Technical Questions

### What is saidata?

Saidata is a hierarchical configuration system that defines software metadata including:
- Package names for different providers
- Service names and configuration
- File locations and directory structures
- Port numbers and network configuration
- OS-specific overrides

### How does SAI detect my operating system?

SAI automatically detects:
- **Platform**: linux, darwin (macOS), windows
- **Distribution**: ubuntu, centos, fedora, etc.
- **Version**: 22.04, 8, 13.0, etc.
- **Architecture**: amd64, arm64

This information is cached for performance and used to select appropriate configurations.

### What happens if saidata doesn't exist for my software?

SAI uses **intelligent defaults**:
- Package name = software name
- Service name = software name
- Config paths = standard locations (`/etc/{software}/`)
- Log paths = standard locations (`/var/log/{software}.log`)

All defaults are validated for existence before use.

### How does template resolution work?

SAI uses Go's template engine with custom functions:

```yaml
# Template in provider YAML
template: "apt install -y {{sai_package}}"

# Resolves to (for nginx):
# apt install -y nginx
```

Template functions like `{{sai_package}}` automatically resolve values from saidata with OS-specific overrides.

### Can I create custom providers?

Yes! Providers are defined in YAML files. Create a new file in the `providers/` directory:

```yaml
version: "1.0"
provider:
  name: "my-provider"
  type: "package_manager"
  platforms: ["linux"]
  executable: "mypkg"
actions:
  install:
    template: "mypkg install {{sai_package}}"
```

See the [Provider Development Guide](PROVIDER_DEVELOPMENT.md) for details.

## Troubleshooting

### SAI says "command not found" after installation

This usually means SAI isn't in your PATH. Solutions:

1. **Check installation location:**
   ```bash
   find /usr -name "sai" 2>/dev/null
   ```

2. **Add to PATH:**
   ```bash
   export PATH="/usr/local/bin:$PATH"
   echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
   ```

3. **Create symlink:**
   ```bash
   sudo ln -s /path/to/sai /usr/local/bin/sai
   ```

### Why do I get "Provider not found" errors?

This means the underlying tool isn't installed or available. For example:

- `apt` provider requires `apt` command (Debian/Ubuntu)
- `brew` provider requires Homebrew (macOS/Linux)
- `docker` provider requires Docker

Install the required tool or use a different provider.

### Commands are slow to execute

Try these solutions:

1. **Clear caches:**
   ```bash
   rm -rf ~/.cache/sai
   ```

2. **Use specific providers:**
   ```bash
   sai install nginx --provider apt  # Skip provider detection
   ```

3. **Disable unused providers:**
   ```yaml
   # In config file
   disabled_providers: ["slow-provider"]
   ```

### How do I get verbose output for debugging?

```bash
# Maximum verbosity
sai install nginx --verbose --debug

# Save output to file
sai install nginx --verbose 2>&1 | tee debug.log
```

## Advanced Usage

### Can I use SAI in scripts and automation?

Yes, SAI is designed for automation:

```bash
# Non-interactive mode
sai install nginx --yes --quiet

# JSON output for parsing
sai list --json | jq '.installed[].name'

# Exit codes for error handling
if sai check nginx; then
  echo "nginx is healthy"
else
  echo "nginx has issues"
fi
```

### How do I manage multiple services at once?

Use batch operations with YAML files:

```yaml
# services.yaml
version: "0.1"
actions:
  - action: install
    software: nginx
  - action: install
    software: mysql
  - action: start
    software: nginx
  - action: start
    software: mysql
```

```bash
sai apply services.yaml
```

### Can I extend SAI with custom actions?

Yes, through provider YAML files. You can define any action:

```yaml
actions:
  backup:
    description: "Backup software data"
    template: "tar -czf /backup/{{.Software}}.tar.gz {{sai_directory 'data'}}"
  
  monitor:
    description: "Monitor software performance"
    template: "watch -n 1 'ps aux | grep {{.Software}}'"
```

### How do I contribute to SAI?

1. **Report issues**: https://github.com/sai-cli/sai/issues
2. **Suggest features**: https://github.com/sai-cli/sai/discussions
3. **Contribute providers**: Create YAML files and submit PRs
4. **Improve documentation**: Help improve guides and examples
5. **Write code**: Contribute to the Go codebase

## Security and Privacy

### Does SAI collect any data?

No, SAI doesn't collect or transmit any personal data. All operations are local except:
- Downloading saidata repository (optional, can be disabled)
- Downloading software through existing package managers (same as using them directly)

### Is it safe to run SAI with sudo?

SAI only requests elevated privileges when necessary (e.g., installing system packages). You can:
- Use `--dry-run` to see what would be executed
- Configure sudoless operation for specific providers
- Use user-space package managers (brew, snap, etc.)

### How do I verify SAI binaries?

All releases include SHA256 checksums:

```bash
# Download checksum file
curl -L -o checksums.txt https://github.com/sai-cli/sai/releases/latest/download/checksums.txt

# Verify binary
sha256sum -c checksums.txt
```

## Platform-Specific Questions

### Does SAI work on Windows?

Yes, SAI fully supports Windows with providers for:
- **winget**: Windows Package Manager
- **choco**: Chocolatey
- **scoop**: Scoop
- **docker**: Docker Desktop
- **npm**, **pip**, etc.: Language package managers

### What about macOS?

SAI works great on macOS (Intel and Apple Silicon) with providers for:
- **brew**: Homebrew
- **docker**: Docker Desktop
- **npm**, **pip**, **gem**: Language package managers
- **mas**: Mac App Store (future)

### Which Linux distributions are supported?

SAI supports all major Linux distributions:
- **Debian/Ubuntu**: apt provider
- **RHEL/CentOS/Fedora**: dnf/yum providers
- **Arch Linux**: pacman provider
- **openSUSE**: zypper provider
- **Alpine**: apk provider
- **Universal**: snap, flatpak, docker providers

## Performance and Scalability

### How fast is SAI?

SAI is designed for performance:
- **Provider detection**: Cached after first run
- **Template resolution**: Optimized with caching
- **Command execution**: Direct system calls, minimal overhead
- **Parallel operations**: Batch operations can run in parallel

### Can SAI handle large deployments?

Yes, SAI scales well:
- **Batch operations**: Process hundreds of software packages
- **Parallel execution**: Configure concurrent operations
- **Resource management**: Built-in timeout and retry logic
- **Monitoring**: JSON output for integration with monitoring systems

### What are the system requirements?

SAI has minimal requirements:
- **Memory**: ~10MB RAM
- **Disk**: ~20MB for binary + cache
- **CPU**: Any modern architecture (amd64, arm64)
- **OS**: Linux, macOS, or Windows

## Future Plans

### What features are planned?

See our [roadmap](https://github.com/sai-cli/sai/projects) for upcoming features:
- GUI interface
- Web dashboard
- More specialized providers
- Enhanced monitoring
- Cloud integration

### How can I request features?

1. **Check existing requests**: https://github.com/sai-cli/sai/discussions
2. **Create new discussion**: Describe your use case
3. **Vote on features**: Help prioritize development
4. **Contribute**: Implement features yourself

### Is SAI actively maintained?

Yes, SAI is actively developed and maintained. We follow semantic versioning and provide:
- Regular releases
- Security updates
- Bug fixes
- New provider additions
- Documentation improvements

---

## Still Have Questions?

- 📖 [Documentation](../README.md)
- 🐛 [Report Issues](https://github.com/sai-cli/sai/issues)
- 💬 [Ask Questions](https://github.com/sai-cli/sai/discussions)
- 📧 [Email Support](mailto:team@sai-cli.dev)

We're here to help! 🚀