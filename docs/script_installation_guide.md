# Script Installation Guide with SAI

This guide demonstrates how to use SAI's script provider to execute installation scripts safely, providing access to software that offers custom installation scripts while maintaining security and automation.

## Overview

The script provider enables you to:
- Execute installation scripts with comprehensive safety measures
- Automate interactive script installations with predefined responses
- Verify script integrity with checksum validation
- Handle environment variables and script arguments
- Integrate script installations with SAI's unified management interface

## Prerequisites

The script provider requires basic scripting tools:

### Ubuntu/Debian
```bash
sudo apt update
sudo apt install curl wget bash python3 expect
```

### CentOS/RHEL/Rocky
```bash
sudo yum install curl wget bash python3 expect
```

### macOS
```bash
# Usually pre-installed, but can install additional tools via Homebrew
brew install expect
```

### Windows
```bash
# PowerShell (pre-installed)
# Or install additional tools via Chocolatey
choco install curl wget
```

## Basic Script Installation

### Example 1: Installing Docker via Script

Docker provides an official installation script for Linux:

```bash
# Install Docker using their official script
sai install docker --provider script

# Check what will be executed (dry run)
sai install docker --provider script --dry-run

# Install with verbose output to see script execution
sai install docker --provider script --verbose
```

**What happens during installation:**
1. **Script Download**: SAI downloads the installation script from the specified URL
2. **Verification**: Verifies script integrity using checksums if provided
3. **User Consent**: Prompts for user confirmation before script execution
4. **Execution**: Runs the script with specified interpreter and arguments
5. **Validation**: Verifies successful installation using validation commands

### Example 2: Automatic Confirmation

For automated environments, you can skip interactive prompts:

```bash
# Install with automatic confirmation
sai install docker --provider script --yes

# Install with predefined responses for interactive prompts
sai install docker --provider script --auto-confirm
```

## Script Configuration Examples

### Example 3: Simple Script Installation

For a basic script installation:

```yaml
# ~/.sai/custom/docker-script.yaml
version: "0.2"
metadata:
  name: "docker"
  description: "Docker container platform"

scripts:
  - name: "main"
    url: "https://get.docker.com/"
    interpreter: "bash"
    checksum: "sha256:a1b2c3d4e5f6..."
    timeout: 600
    environment:
      DEBIAN_FRONTEND: "noninteractive"
```

### Example 4: Script with Arguments

For scripts requiring specific arguments:

```yaml
# Example: Node.js installation script with version
scripts:
  - name: "main"
    url: "https://nodejs.org/dist/install.sh"
    interpreter: "bash"
    arguments: ["--version", "18.17.0", "--prefix", "/usr/local"]
    timeout: 300
    checksum: "sha256:verified_checksum_here"
```

### Example 5: Interactive Script Automation

For scripts with interactive prompts:

```yaml
# Example: Automated responses for interactive installation
scripts:
  - name: "main"
    url: "https://example.com/interactive-installer.sh"
    interpreter: "bash"
    auto_confirm: true
    confirm_responses: |
      y
      /opt/myapp
      yes
      stable
    timeout: 900
    environment:
      INSTALL_MODE: "automated"
```

## Security Considerations

### Checksum Verification

Always verify script integrity:

```yaml
scripts:
  - name: "main"
    url: "https://get.docker.com/"
    checksum: "sha256:verified_checksum_from_docker"
    # SAI will verify the script before execution
```

### HTTPS Enforcement

Use secure URLs only:

```yaml
scripts:
  - url: "https://get.docker.com/"  # ✓ Secure HTTPS
    # Not: "http://get.docker.com/"  # ✗ Insecure HTTP
```

### User Consent

SAI requires explicit user consent for script execution:

```bash
# User must confirm script execution
sai install docker --provider script
# Output: "This will execute a script from https://get.docker.com/. Continue? [y/N]"

# Or use --yes for automation
sai install docker --provider script --yes
```

### Script Sandboxing

Consider running scripts in isolated environments:

```yaml
scripts:
  - name: "main"
    working_dir: "/tmp/sai-script-sandbox"
    environment:
      HOME: "/tmp/sai-script-sandbox"
      PATH: "/usr/local/bin:/usr/bin:/bin"
```

## Advanced Script Configuration

### Environment Variables

Control script execution environment:

```yaml
scripts:
  - name: "main"
    url: "https://install.example.com/script.sh"
    environment:
      DEBIAN_FRONTEND: "noninteractive"
      INSTALL_PREFIX: "/opt/myapp"
      ENABLE_FEATURE: "true"
      LOG_LEVEL: "info"
```

### Custom Working Directory

Set specific working directory for script execution:

```yaml
scripts:
  - name: "main"
    working_dir: "/tmp/myapp-install"
    # Script will be executed from this directory
```

### Timeout Configuration

Set appropriate timeouts for long-running scripts:

```yaml
scripts:
  - name: "main"
    timeout: 1800  # 30 minutes for complex installations
```

### Custom Commands

Override default script handling:

```yaml
scripts:
  - name: "main"
    custom_commands:
      download: "curl -fsSL {{url}} -o {{script_file}}"
      install: "chmod +x {{script_file}} && {{script_file}} {{arguments}}"
      validation: "which docker && docker --version"
      uninstall: "apt-get remove -y docker-ce docker-ce-cli containerd.io"
```

## Interactive Script Handling

### Automatic Confirmation

For scripts that require yes/no confirmations:

```yaml
scripts:
  - name: "main"
    auto_confirm: true  # Automatically answer 'yes' to prompts
```

### Predefined Responses

For complex interactive scripts:

```yaml
scripts:
  - name: "main"
    confirm_responses: |
      y
      /usr/local
      stable
      yes
      admin@example.com
```

### Expect Scripts

For complex interactive handling:

```yaml
scripts:
  - name: "main"
    expect_script: |
      #!/usr/bin/expect -f
      spawn bash {{script_file}}
      expect "Do you want to continue?" { send "y\r" }
      expect "Installation directory:" { send "/opt/myapp\r" }
      expect "Version to install:" { send "stable\r" }
      expect eof
```

## Managing Script Installations

### Checking Installation Status

```bash
# Check if docker is installed via script
sai status docker --provider script

# Get version information
sai version docker --provider script

# View installation details
sai info docker --provider script
```

### Script Logs

View script execution logs:

```bash
# View installation logs
sai logs docker --provider script

# View specific log file
tail -f /var/log/sai/docker-script.log
```

### Uninstalling Script Installations

```bash
# Remove script-installed software
sai uninstall docker --provider script

# Force removal if uninstall script fails
sai uninstall docker --provider script --force
```

## Common Script Examples

### Docker Installation

```yaml
scripts:
  - name: "docker"
    url: "https://get.docker.com/"
    interpreter: "bash"
    timeout: 600
    environment:
      DEBIAN_FRONTEND: "noninteractive"
    custom_commands:
      validation: "docker --version && systemctl is-active docker"
```

### Node.js via NVM

```yaml
scripts:
  - name: "nvm"
    url: "https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh"
    interpreter: "bash"
    timeout: 300
    environment:
      NVM_DIR: "$HOME/.nvm"
    custom_commands:
      validation: "source ~/.bashrc && nvm --version"
```

### Rust via Rustup

```yaml
scripts:
  - name: "rust"
    url: "https://sh.rustup.rs"
    interpreter: "bash"
    arguments: ["-y"]  # Non-interactive installation
    timeout: 600
    environment:
      RUSTUP_HOME: "$HOME/.rustup"
      CARGO_HOME: "$HOME/.cargo"
```

### Oh My Zsh

```yaml
scripts:
  - name: "oh-my-zsh"
    url: "https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh"
    interpreter: "bash"
    environment:
      RUNZSH: "no"  # Don't start zsh after installation
      CHSH: "no"    # Don't change shell
```

## Troubleshooting Script Installations

### Common Issues and Solutions

#### 1. Script Download Failures

**Problem**: Cannot download installation script
```
Error: Failed to download script from URL
```

**Solutions**:
```bash
# Check URL accessibility
curl -I "https://get.docker.com/"

# Verify network connectivity and proxy settings
export https_proxy=http://proxy.company.com:8080

# Check DNS resolution
nslookup get.docker.com
```

#### 2. Checksum Verification Failures

**Problem**: Script checksum doesn't match
```
Error: Script checksum verification failed
```

**Solutions**:
```bash
# Verify checksum manually
sha256sum downloaded_script.sh
echo "expected_checksum downloaded_script.sh" | sha256sum -c

# Update checksum in configuration
# Check provider's official checksums
```

#### 3. Script Execution Failures

**Problem**: Script fails during execution
```
Error: Script execution failed with exit code 1
```

**Solutions**:
```bash
# Run script manually to debug
bash -x downloaded_script.sh

# Check script logs
tail -f /var/log/sai/docker-script.log

# Verify environment variables
env | grep -E "(DEBIAN_FRONTEND|PATH|HOME)"
```

#### 4. Permission Issues

**Problem**: Script requires elevated privileges
```
Error: Permission denied during script execution
```

**Solutions**:
```bash
# Run with sudo
sudo sai install docker --provider script

# Or configure script to handle permissions
```

#### 5. Interactive Prompt Handling

**Problem**: Script hangs on interactive prompts
```
Script appears to be waiting for input
```

**Solutions**:
```yaml
# Add automatic confirmation
scripts:
  - auto_confirm: true
    confirm_responses: |
      y
      /usr/local
      stable
```

### Debugging Script Issues

Enable detailed logging and debugging:

```bash
# Enable verbose output
sai install docker --provider script --verbose

# Dry run to see what would be executed
sai install docker --provider script --dry-run

# Enable script debugging
export SAI_SCRIPT_DEBUG=1
sai install docker --provider script
```

## Best Practices

### 1. Security First

Always verify script integrity:
```yaml
scripts:
  - checksum: "sha256:verified_checksum_from_official_source"
```

### 2. Use Official Scripts

Prefer official installation scripts:
```yaml
scripts:
  - url: "https://get.docker.com/"  # Official Docker script
    # Not: random GitHub gists or unofficial sources
```

### 3. Explicit User Consent

Never run scripts without user awareness:
```bash
# Always inform users what script will be executed
sai install docker --provider script  # Shows script URL and asks for confirmation
```

### 4. Environment Isolation

Limit environment variable exposure:
```yaml
scripts:
  - environment:
      # Only include necessary variables
      DEBIAN_FRONTEND: "noninteractive"
      # Don't expose sensitive variables like API keys
```

### 5. Timeout Configuration

Set reasonable timeouts:
```yaml
scripts:
  - timeout: 600  # 10 minutes for most installations
    # Adjust based on script complexity
```

### 6. Proper Cleanup

Ensure proper cleanup on failure:
```yaml
scripts:
  - custom_commands:
      uninstall: "apt-get remove -y installed-package && rm -rf /opt/myapp"
```

### 7. Logging and Auditing

Maintain detailed logs:
```yaml
scripts:
  - log_file: "/var/log/sai/{{metadata.name}}-script.log"
```

## Integration with System Services

After script installation, integrate with system management:

### Service Management

```bash
# Start services installed by script
sai start docker --provider script

# Enable automatic startup
sai enable docker --provider script

# Check service status
sai status docker --provider script
```

### Configuration Management

```bash
# View configuration files
sai config docker --provider script

# Edit configuration
sudo sai config docker --provider script --edit
```

## Performance and Optimization

### Parallel Script Execution

For multiple independent scripts:
```bash
# Install multiple tools via scripts
sai install docker nodejs rust --provider script --parallel
```

### Script Caching

Cache downloaded scripts:
```yaml
# Global SAI configuration
cache:
  scripts:
    enabled: true
    directory: "/var/cache/sai/scripts"
    retention: "7d"
```

### Optimized Environments

Use optimized environments for faster execution:
```yaml
scripts:
  - environment:
      DEBIAN_FRONTEND: "noninteractive"
      APT_KEY_DONT_WARN_ON_DANGEROUS_USAGE: "1"
      NEEDRESTART_MODE: "a"  # Automatic restart decisions
```

## Conclusion

Script installation with SAI provides:
- **Flexibility**: Access to software with custom installation methods
- **Security**: Checksum verification and user consent requirements
- **Automation**: Automated handling of interactive scripts
- **Integration**: Seamless integration with SAI's management features
- **Safety**: Rollback capabilities and comprehensive logging

The script provider makes executing installation scripts as safe and manageable as traditional package installations while maintaining the flexibility that custom scripts provide.

For more information, see:
- [SAI Script Functions Reference](sai_script_functions.md)
- [Provider Development Guide](PROVIDER_DEVELOPMENT.md)
- [SAI Synopsis](sai_synopsis.md)