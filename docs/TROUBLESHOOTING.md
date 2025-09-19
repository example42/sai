# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with SAI CLI.

## Table of Contents

- [Quick Diagnostics](#quick-diagnostics)
- [Installation Issues](#installation-issues)
- [Provider Issues](#provider-issues)
- [Template and Saidata Issues](#template-and-saidata-issues)
- [Command Execution Issues](#command-execution-issues)
- [Configuration Issues](#configuration-issues)
- [Performance Issues](#performance-issues)
- [Platform-Specific Issues](#platform-specific-issues)
- [Debugging Tools](#debugging-tools)
- [Getting Help](#getting-help)

## Quick Diagnostics

### Basic Health Check

```bash
# Check SAI version and basic functionality
sai version

# Verify configuration
sai stats

# Test with a simple command
sai list --dry-run
```

### Environment Information

```bash
# Check system information
sai stats --verbose

# Verify provider availability
sai stats --providers

# Check configuration file location
sai config --show-path
```

## Installation Issues

### SAI Not Found After Installation

**Symptoms:**
- `command not found: sai` after installation
- SAI installed but not in PATH

**Solutions:**

1. **Check installation location:**
   ```bash
   # Find where SAI was installed
   find /usr -name "sai" 2>/dev/null
   find /usr/local -name "sai" 2>/dev/null
   find ~ -name "sai" 2>/dev/null
   ```

2. **Add to PATH (Unix/Linux/macOS):**
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   export PATH="/usr/local/bin:$PATH"
   
   # Or create symlink
   sudo ln -s /path/to/sai /usr/local/bin/sai
   ```

3. **Windows PATH issues:**
   ```powershell
   # Check current PATH
   $env:PATH -split ';'
   
   # Add to PATH permanently
   [Environment]::SetEnvironmentVariable("PATH", "$env:PATH;C:\path\to\sai", "User")
   ```

### Permission Issues

**Symptoms:**
- `Permission denied` when running SAI
- Cannot install to system directories

**Solutions:**

1. **Fix binary permissions:**
   ```bash
   chmod +x /path/to/sai
   ```

2. **Install to user directory:**
   ```bash
   # Install to user's local bin
   mkdir -p ~/.local/bin
   cp sai ~/.local/bin/
   export PATH="$HOME/.local/bin:$PATH"
   ```

3. **Use sudo for system installation:**
   ```bash
   sudo cp sai /usr/local/bin/
   sudo chmod +x /usr/local/bin/sai
   ```

### Download/Network Issues

**Symptoms:**
- Installation script fails to download
- Cannot fetch latest version

**Solutions:**

1. **Manual download:**
   ```bash
   # Download directly from GitHub
   curl -L -o sai https://github.com/sai-cli/sai/releases/latest/download/sai-linux-amd64
   chmod +x sai
   ```

2. **Use alternative download method:**
   ```bash
   # Use wget instead of curl
   wget https://github.com/sai-cli/sai/releases/latest/download/sai-linux-amd64 -O sai
   ```

3. **Proxy configuration:**
   ```bash
   # Set proxy for downloads
   export https_proxy=http://proxy.example.com:8080
   export http_proxy=http://proxy.example.com:8080
   ```

## Provider Issues

### Provider Not Found

**Symptoms:**
- `Provider 'xyz' not found`
- No providers available for software

**Diagnosis:**
```bash
# List all available providers
sai stats --providers

# Check specific provider
sai stats --provider apt

# Verify provider file exists
ls -la providers/apt.yaml
```

**Solutions:**

1. **Check provider availability:**
   ```bash
   # Verify the underlying tool is installed
   which apt
   which brew
   which docker
   ```

2. **Install missing provider tools:**
   ```bash
   # Install package manager
   # On macOS: install Homebrew
   /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
   
   # On Windows: install Chocolatey
   Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
   ```

3. **Check provider configuration:**
   ```bash
   # Validate provider YAML
   sai validate-provider providers/apt.yaml
   ```

### Provider Selection Issues

**Symptoms:**
- Wrong provider selected automatically
- Cannot force specific provider

**Solutions:**

1. **Force specific provider:**
   ```bash
   sai install nginx --provider docker
   ```

2. **Check provider priority:**
   ```bash
   # View provider priorities
   sai stats --providers --verbose
   ```

3. **Configure provider priority:**
   ```yaml
   # In config file
   provider_priority:
     apt: 100
     brew: 90
     docker: 80
   ```

### Provider Execution Failures

**Symptoms:**
- Provider commands fail
- Timeout errors
- Permission denied from provider

**Diagnosis:**
```bash
# Test provider command manually
sai install nginx --provider apt --dry-run

# Run with verbose output
sai install nginx --provider apt --verbose

# Check provider validation
sai check nginx --provider apt
```

**Solutions:**

1. **Fix permissions:**
   ```bash
   # Run with sudo if required
   sudo sai install nginx --provider apt
   
   # Or configure sudoless operation
   echo "$USER ALL=(ALL) NOPASSWD: /usr/bin/apt" | sudo tee /etc/sudoers.d/sai-apt
   ```

2. **Increase timeout:**
   ```yaml
   # In config file
   timeout: 600s  # 10 minutes
   ```

3. **Update package lists:**
   ```bash
   # Update package manager cache
   sudo apt update
   brew update
   ```

## Template and Saidata Issues

### Template Resolution Failures

**Symptoms:**
- `Template variable not found`
- Actions not available for software
- Empty command templates

**Diagnosis:**
```bash
# Check available saidata
sai info nginx --verbose

# Test template resolution
sai install nginx --dry-run --verbose

# Check saidata repository
sai saidata status
```

**Solutions:**

1. **Update saidata repository:**
   ```bash
   sai saidata update
   ```

2. **Check saidata structure:**
   ```bash
   # Look for software configuration
   find ~/.cache/sai/saidata -name "*nginx*" -type f
   ```

3. **Use provider-specific overrides:**
   ```bash
   # Force provider that has complete saidata
   sai install nginx --provider docker
   ```

### Missing Saidata

**Symptoms:**
- Software not found in saidata
- Using intelligent defaults
- Limited functionality

**Solutions:**

1. **Check if software exists with different name:**
   ```bash
   sai search nginx
   sai search apache
   sai search httpd
   ```

2. **Use provider mappings:**
   ```bash
   # Check provider-specific names
   sai info nginx --provider apt --verbose
   ```

3. **Create custom saidata:**
   ```yaml
   # Create ~/.config/sai/saidata/software/ng/nginx/default.yaml
   version: "0.2"
   metadata:
     name: "nginx"
     description: "Web server"
   packages:
     - name: "nginx"
   services:
     - name: "nginx"
       service_name: "nginx"
   ```

### OS-Specific Override Issues

**Symptoms:**
- Wrong configuration for current OS
- OS detection failures
- Override files not loading

**Diagnosis:**
```bash
# Check OS detection
sai stats --system

# Verify override file exists
ls -la ~/.cache/sai/saidata/software/ap/apache/ubuntu/22.04.yaml
```

**Solutions:**

1. **Clear detection cache:**
   ```bash
   rm -rf ~/.cache/sai/detection
   sai stats --system  # Re-detect
   ```

2. **Manual OS specification:**
   ```bash
   # Set environment variable
   export SAI_OS_OVERRIDE="ubuntu:22.04"
   ```

3. **Check override file format:**
   ```yaml
   # Ensure proper YAML structure
   version: "0.2"
   # Override specific fields only
   packages:
     - name: "apache2"  # Ubuntu-specific package name
   ```

## Command Execution Issues

### Command Not Found

**Symptoms:**
- `command not found` errors during execution
- Provider tools not available

**Solutions:**

1. **Install missing tools:**
   ```bash
   # Install required package managers
   sudo apt install snapd
   sudo dnf install flatpak
   ```

2. **Check PATH:**
   ```bash
   # Verify tools are in PATH
   echo $PATH
   which docker
   which systemctl
   ```

3. **Use full paths in provider configuration:**
   ```yaml
   # In provider YAML
   actions:
     install:
       template: "/usr/bin/apt install -y {{sai_package}}"
   ```

### Permission Denied

**Symptoms:**
- `Permission denied` during command execution
- Cannot access system resources

**Solutions:**

1. **Run with appropriate privileges:**
   ```bash
   sudo sai install nginx
   ```

2. **Configure sudoless operation:**
   ```bash
   # Add to sudoers
   echo "$USER ALL=(ALL) NOPASSWD: /usr/bin/systemctl" | sudo tee /etc/sudoers.d/sai-systemctl
   ```

3. **Use user-space alternatives:**
   ```bash
   # Use user-space package managers
   sai install nginx --provider brew  # User-space on macOS
   sai install nginx --provider snap  # User-space on Linux
   ```

### Timeout Issues

**Symptoms:**
- Commands timeout before completion
- Long-running operations fail

**Solutions:**

1. **Increase timeout:**
   ```yaml
   # In config file
   timeout: 1800s  # 30 minutes
   ```

2. **Use provider-specific timeouts:**
   ```yaml
   # In provider YAML
   actions:
     install:
       timeout: 3600  # 1 hour for this action
   ```

3. **Run in background:**
   ```bash
   # Use screen or tmux for long operations
   screen -S sai-install
   sai install large-software --yes
   ```

## Configuration Issues

### Configuration File Not Found

**Symptoms:**
- Using default configuration
- Custom settings not applied

**Solutions:**

1. **Check configuration file locations:**
   ```bash
   # SAI looks in these locations (in order):
   ls -la ./sai.yaml
   ls -la ~/.config/sai/config.yaml
   ls -la /etc/sai/config.yaml
   ```

2. **Create configuration file:**
   ```bash
   mkdir -p ~/.config/sai
   cat > ~/.config/sai/config.yaml << EOF
   log_level: "info"
   timeout: 300s
   confirmations:
     install: true
     system_changes: true
   EOF
   ```

3. **Specify configuration explicitly:**
   ```bash
   sai install nginx --config /path/to/config.yaml
   ```

### Invalid Configuration

**Symptoms:**
- Configuration parsing errors
- Invalid YAML syntax

**Solutions:**

1. **Validate YAML syntax:**
   ```bash
   # Use online YAML validator or
   python -c "import yaml; yaml.safe_load(open('config.yaml'))"
   ```

2. **Check configuration schema:**
   ```bash
   # Verify against expected structure
   sai config --validate
   ```

3. **Reset to defaults:**
   ```bash
   # Rename problematic config
   mv ~/.config/sai/config.yaml ~/.config/sai/config.yaml.backup
   ```

## Performance Issues

### Slow Command Execution

**Symptoms:**
- Commands take long time to start
- Slow provider detection

**Solutions:**

1. **Clear caches:**
   ```bash
   rm -rf ~/.cache/sai
   ```

2. **Disable unnecessary providers:**
   ```yaml
   # In config file
   disabled_providers:
     - slow-provider
     - unused-provider
   ```

3. **Use specific providers:**
   ```bash
   # Avoid provider detection overhead
   sai install nginx --provider apt
   ```

### High Memory Usage

**Symptoms:**
- SAI uses excessive memory
- System becomes slow during operation

**Solutions:**

1. **Reduce concurrent operations:**
   ```bash
   # Process one item at a time
   sai apply batch.yaml --sequential
   ```

2. **Use streaming output:**
   ```bash
   # Don't buffer large outputs
   sai logs nginx --stream
   ```

## Platform-Specific Issues

### Linux Issues

**Common Problems:**
- Systemd service management
- Package manager conflicts
- Permission issues

**Solutions:**
```bash
# Check systemd status
systemctl --user status
sudo systemctl status

# Fix package manager locks
sudo rm /var/lib/apt/lists/lock
sudo rm /var/cache/apt/archives/lock
sudo rm /var/lib/dpkg/lock*

# Check SELinux/AppArmor
getenforce
sudo aa-status
```

### macOS Issues

**Common Problems:**
- Homebrew not in PATH
- System Integrity Protection (SIP)
- Code signing issues

**Solutions:**
```bash
# Fix Homebrew PATH
echo 'export PATH="/opt/homebrew/bin:$PATH"' >> ~/.zshrc

# Check SIP status
csrutil status

# Allow unsigned binaries (if needed)
sudo spctl --master-disable
```

### Windows Issues

**Common Problems:**
- PowerShell execution policy
- Windows Defender blocking
- PATH issues

**Solutions:**
```powershell
# Fix execution policy
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Add to PATH
$env:PATH += ";C:\path\to\sai"

# Exclude from Windows Defender
Add-MpPreference -ExclusionPath "C:\path\to\sai"
```

## Debugging Tools

### Verbose Output

```bash
# Enable maximum verbosity
sai install nginx --verbose --debug

# Show configuration being used
sai stats --config --verbose

# Display template resolution
sai install nginx --dry-run --verbose
```

### Dry Run Mode

```bash
# See what would be executed
sai install nginx --dry-run

# Test batch operations
sai apply batch.yaml --dry-run

# Verify provider selection
sai start nginx --provider systemd --dry-run
```

### Log Analysis

```bash
# Check SAI logs
tail -f ~/.cache/sai/logs/sai.log

# Enable debug logging
export SAI_LOG_LEVEL=debug
sai install nginx

# Save logs to file
sai install nginx --verbose 2>&1 | tee install.log
```

### System Information

```bash
# Comprehensive system info
sai stats --system --verbose

# Provider availability
sai stats --providers --verbose

# Configuration dump
sai config --dump
```

## Getting Help

### Self-Help Resources

1. **Built-in help:**
   ```bash
   sai --help
   sai install --help
   sai stats --help
   ```

2. **Documentation:**
   - [README](../README.md)
   - [Examples](EXAMPLES.md)
   - [Provider Development](PROVIDER_DEVELOPMENT.md)

3. **Diagnostic commands:**
   ```bash
   sai stats --verbose
   sai config --validate
   sai version --verbose
   ```

### Community Support

1. **GitHub Issues:**
   - Bug reports: https://github.com/sai-cli/sai/issues
   - Feature requests: https://github.com/sai-cli/sai/issues

2. **GitHub Discussions:**
   - Questions: https://github.com/sai-cli/sai/discussions
   - Ideas: https://github.com/sai-cli/sai/discussions

3. **Documentation:**
   - Wiki: https://github.com/sai-cli/sai/wiki
   - Examples: https://github.com/sai-cli/sai/tree/main/docs

### Reporting Issues

When reporting issues, please include:

1. **System information:**
   ```bash
   sai stats --system --verbose
   ```

2. **SAI version:**
   ```bash
   sai version --verbose
   ```

3. **Command that failed:**
   ```bash
   sai install nginx --verbose --dry-run
   ```

4. **Configuration:**
   ```bash
   sai config --dump
   ```

5. **Error logs:**
   ```bash
   tail -50 ~/.cache/sai/logs/sai.log
   ```

### Emergency Recovery

If SAI is completely broken:

1. **Reset all configuration:**
   ```bash
   rm -rf ~/.config/sai
   rm -rf ~/.cache/sai
   ```

2. **Reinstall SAI:**
   ```bash
   curl -fsSL https://raw.githubusercontent.com/sai-cli/sai/main/scripts/install.sh | bash
   ```

3. **Use system package managers directly:**
   ```bash
   # Fallback to native tools
   sudo apt install nginx
   brew install nginx
   ```

---

This troubleshooting guide covers the most common issues. If you encounter problems not covered here, please check the [GitHub Issues](https://github.com/sai-cli/sai/issues) or create a new issue with detailed information about your problem.