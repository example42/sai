# Provider Development Guide

This guide explains how to create and extend providers for SAI CLI. Providers are the core mechanism that enables SAI to work with different package managers, container platforms, and specialized tools across various operating systems.

## Table of Contents

- [Overview](#overview)
- [Provider Architecture](#provider-architecture)
- [Creating a New Provider](#creating-a-new-provider)
- [Provider Configuration](#provider-configuration)
- [Template System](#template-system)
- [Validation and Safety](#validation-and-safety)
- [Testing Providers](#testing-providers)
- [Best Practices](#best-practices)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

## Overview

SAI uses a **dynamic provider system** where all provider logic is defined in YAML files rather than hardcoded Go implementations. This approach provides several key benefits:

- **Zero-Code Provider Addition**: New providers can be added by simply creating a YAML file
- **Runtime Flexibility**: Provider behavior can be modified without recompiling SAI
- **Community Contributions**: Non-Go developers can contribute providers using familiar YAML syntax
- **Rapid Prototyping**: New provider concepts can be tested quickly without code changes
- **Maintenance Simplification**: Provider updates don't require application releases

## Provider Architecture

### Provider Types

SAI supports several types of providers:

1. **Package Managers**: `apt`, `brew`, `dnf`, `yum`, `pacman`, etc.
2. **Container Platforms**: `docker`, `helm`, `podman`
3. **Language Package Managers**: `npm`, `pip`, `gem`, `cargo`, `go`
4. **Alternative Installation Methods**: `source`, `binary`, `script`
5. **Specialized Tools**: `debug`, `security`, `monitoring`, `backup`
6. **Cloud Platforms**: `aws`, `gcp`, `azure`
7. **Custom Tools**: Any command-line tool that can be automated

### Provider Structure

Each provider is defined in a YAML file following the `providerdata-0.1-schema.json` schema:

```yaml
version: "1.0"
provider:
  name: "provider-name"
  display_name: "Human Readable Name"
  description: "Brief description of what this provider does"
  type: "package_manager"  # or container, specialized, etc.
  platforms: ["linux", "darwin", "windows"]
  capabilities: ["install", "uninstall", "start", "stop", "status"]
  priority: 100
  executable: "command-name"  # Used for availability detection

actions:
  install:
    description: "Install a software package"
    template: "command install {{sai_package}}"
    requires_root: false
    timeout: 300
    validation:
      command: "command list | grep {{sai_package}}"
    rollback: "command uninstall {{sai_package}}"

mappings:
  # Optional: map generic software names to provider-specific names
  nginx: "nginx-full"
  apache: "apache2"
```

## Creating a New Provider

### Step 1: Choose Provider Location

- **Standard providers**: Place in `providers/` directory
- **Specialized providers**: Place in `providers/specialized/` directory

### Step 2: Create Provider File

Create a new YAML file named after your provider (e.g., `my-provider.yaml`):

```yaml
version: "1.0"
provider:
  name: "my-provider"
  display_name: "My Package Manager"
  description: "Custom package manager for my organization"
  type: "package_manager"
  platforms: ["linux"]
  capabilities: ["install", "uninstall", "upgrade", "search", "info", "list"]
  priority: 50
  executable: "mypkg"

actions:
  install:
    description: "Install a software package"
    template: "mypkg install {{sai_package}}"
    requires_root: true
    timeout: 600
    validation:
      command: "mypkg list --installed | grep -q '^{{sai_package}}$'"
    rollback: "mypkg remove {{sai_package}}"
    
  uninstall:
    description: "Remove a software package"
    template: "mypkg remove {{sai_package}}"
    requires_root: true
    timeout: 300
    validation:
      command: "! mypkg list --installed | grep -q '^{{sai_package}}$'"
    
  upgrade:
    description: "Upgrade a software package"
    template: "mypkg upgrade {{sai_package}}"
    requires_root: true
    timeout: 600
    
  search:
    description: "Search for software packages"
    template: "mypkg search {{.Software}}"
    requires_root: false
    timeout: 30
    
  info:
    description: "Show package information"
    template: "mypkg info {{sai_package}}"
    requires_root: false
    timeout: 30
    
  list:
    description: "List installed packages"
    template: "mypkg list --installed"
    requires_root: false
    timeout: 60

mappings:
  # Map generic names to provider-specific package names
  nginx: "nginx-server"
  apache: "httpd"
  mysql: "mysql-server"
```

### Step 3: Validate Provider

Use SAI's built-in validation to check your provider:

```bash
# Validate provider syntax
sai stats --provider my-provider

# Test provider availability
sai info nginx --provider my-provider --dry-run
```

## Provider Configuration

### Provider Metadata

```yaml
provider:
  name: "unique-provider-name"           # Must be unique across all providers
  display_name: "Human Readable Name"   # Shown in UI
  description: "Brief description"       # Help text
  type: "package_manager"               # Provider category
  platforms: ["linux", "darwin"]       # Supported platforms
  capabilities: ["install", "start"]   # Supported actions
  priority: 100                         # Selection priority (higher = preferred)
  executable: "command-name"            # Command used for availability detection
```

### Platform Support

Specify which platforms your provider supports:

```yaml
platforms: ["linux", "darwin", "windows"]  # All platforms
platforms: ["linux"]                       # Linux only
platforms: ["darwin"]                      # macOS only
platforms: ["windows"]                     # Windows only
```

### Capabilities

List all actions your provider supports:

```yaml
capabilities: [
  "install", "uninstall", "upgrade",     # Software management
  "start", "stop", "restart",            # Service management
  "enable", "disable", "status",         # Service control
  "search", "info", "version",           # Information
  "logs", "config", "check",             # Troubleshooting
  "cpu", "memory", "io"                  # Monitoring
]
```

## Template System

SAI uses Go's `text/template` engine with custom functions for dynamic command generation.

### Template Variables

- `{{.Software}}`: The software name passed to the command
- `{{.Action}}`: The current action being performed
- `{{.Provider}}`: The current provider name

### SAI Template Functions

These functions automatically resolve values from saidata:

```yaml
# Package functions
{{sai_package}}              # Get default package name
{{sai_package "apt"}}        # Get package name for specific provider
{{sai_packages}}             # Get all package names as space-separated string
{{sai_packages "apt"}}       # Get all packages for specific provider

# Service functions
{{sai_service}}              # Get default service name
{{sai_service "nginx"}}      # Get specific service name

# File and directory functions
{{sai_file "config"}}        # Get file path by name
{{sai_directory "data"}}     # Get directory path by name

# Port functions
{{sai_port}}                 # Get default port
{{sai_port 0}}               # Get first port
{{sai_port 1}}               # Get second port

# Command functions
{{sai_command "start"}}      # Get command path by name

# Validation functions
{{file_exists "/path/to/file"}}        # Check if file exists
{{service_exists "nginx"}}             # Check if service exists
{{command_exists "nginx"}}             # Check if command exists
{{directory_exists "/path/to/dir"}}    # Check if directory exists

# Default generation functions
{{default_config_path .Software}}     # Generate default config path
{{default_log_path .Software}}        # Generate default log path
{{default_data_dir .Software}}        # Generate default data directory
```

### Template Examples

```yaml
actions:
  install:
    # Simple template
    template: "apt install -y {{sai_package}}"
    
  start:
    # Using service name
    template: "systemctl start {{sai_service}}"
    
  config:
    # Using file paths
    template: "cat {{sai_file 'config'}}"
    
  logs:
    # Using default paths with fallback
    template: "tail -f {{sai_file 'log'}} || tail -f {{default_log_path .Software}}"
    
  install_with_validation:
    # Complex template with validation
    template: |
      if {{file_exists (sai_file "config")}}; then
        echo "Config exists, installing..."
        apt install -y {{sai_package}}
      else
        echo "Creating default config..."
        mkdir -p {{default_config_path .Software | dirname}}
        apt install -y {{sai_package}}
      fi
```

### Multi-Step Actions

For complex operations, use the `steps` field instead of `template`:

```yaml
actions:
  install:
    description: "Install with repository setup"
    steps:
      - name: "Add repository"
        command: "curl -fsSL https://example.com/key | apt-key add -"
        requires_root: true
      - name: "Update package list"
        command: "apt update"
        requires_root: true
      - name: "Install package"
        command: "apt install -y {{sai_package}}"
        requires_root: true
    timeout: 600
    rollback: "apt remove -y {{sai_package}}"
```

## Validation and Safety

### Command Validation

Add validation commands to verify successful execution:

```yaml
actions:
  install:
    template: "apt install -y {{sai_package}}"
    validation:
      command: "dpkg -l | grep -q '^ii.*{{sai_package}}'"
      timeout: 30
    rollback: "apt remove -y {{sai_package}}"
```

### Safety Checks

Use template functions to add safety checks:

```yaml
actions:
  start:
    template: |
      if {{service_exists (sai_service)}}; then
        systemctl start {{sai_service}}
      else
        echo "Service {{sai_service}} not found"
        exit 1
      fi
```

### Rollback Actions

Define rollback commands for destructive operations:

```yaml
actions:
  install:
    template: "apt install -y {{sai_package}}"
    rollback: "apt remove -y {{sai_package}}"
    
  enable:
    template: "systemctl enable {{sai_service}}"
    rollback: "systemctl disable {{sai_service}}"
```

## Testing Providers

### Dry Run Testing

Test your provider without executing commands:

```bash
# Test specific action
sai install nginx --provider my-provider --dry-run

# Test with verbose output
sai install nginx --provider my-provider --dry-run --verbose
```

### Validation Testing

Test provider validation:

```bash
# Check provider syntax
sai stats --provider my-provider

# Validate against schema
sai validate-provider providers/my-provider.yaml
```

### Integration Testing

Create integration tests for your provider:

```bash
# Test installation flow
sai install test-package --provider my-provider --yes
sai status test-package --provider my-provider
sai uninstall test-package --provider my-provider --yes
```

## Best Practices

### 1. Provider Naming

- Use lowercase, hyphenated names: `my-provider`
- Make names descriptive and unique
- Avoid generic names like `installer` or `manager`

### 2. Platform Compatibility

- Test on all supported platforms
- Use platform-specific commands when necessary
- Handle platform differences gracefully

### 3. Error Handling

- Provide clear error messages
- Use appropriate exit codes
- Include rollback procedures for destructive actions

### 4. Template Design

- Keep templates simple and readable
- Use validation functions to prevent errors
- Provide fallbacks for missing saidata

### 5. Documentation

- Include clear descriptions for all actions
- Document any special requirements
- Provide usage examples

### 6. Security

- Minimize required privileges
- Validate inputs to prevent injection
- Use secure download methods

## Alternative Installation Providers

SAI includes three specialized providers for alternative installation methods that go beyond traditional package managers:

### Source Provider

The source provider enables building software from source code with support for multiple build systems:

```yaml
version: "1.0"
provider:
  name: "source"
  display_name: "Source Build Provider"
  type: "source"
  platforms: ["linux", "darwin"]
  capabilities: ["install", "uninstall", "upgrade", "version", "info"]
  executable: "make"  # Basic requirement for most builds

actions:
  install:
    description: "Build and install from source"
    steps:
      - name: "Install prerequisites"
        command: "{{sai_source(0, 'prerequisites_install_cmd')}}"
        requires_root: true
      - name: "Download source"
        command: "{{sai_source(0, 'download_cmd')}}"
      - name: "Extract source"
        command: "{{sai_source(0, 'extract_cmd')}}"
      - name: "Configure build"
        command: "cd {{sai_source(0, 'source_dir')}} && {{sai_source(0, 'configure_cmd')}}"
      - name: "Build software"
        command: "cd {{sai_source(0, 'source_dir')}} && {{sai_source(0, 'build_cmd')}}"
      - name: "Install software"
        command: "cd {{sai_source(0, 'source_dir')}} && {{sai_source(0, 'install_cmd')}}"
        requires_root: true
    timeout: 1800
    validation:
      command: "{{sai_source(0, 'validation_cmd')}}"
    rollback: "{{sai_source(0, 'uninstall_cmd')}}"
```

**Key Template Functions:**
- `{{sai_source(index, field)}}` - Access source configuration
- Supports autotools, cmake, make, meson, ninja build systems
- Automatic prerequisite detection and installation
- Build directory and path management

### Binary Provider

The binary provider handles downloading and installing pre-compiled binaries:

```yaml
version: "1.0"
provider:
  name: "binary"
  display_name: "Binary Download Provider"
  type: "binary"
  platforms: ["linux", "darwin", "windows"]
  capabilities: ["install", "uninstall", "upgrade", "version", "info"]
  executable: "wget"  # Or curl for downloads

actions:
  install:
    description: "Download and install binary"
    steps:
      - name: "Download binary"
        command: "{{sai_binary(0, 'download_cmd')}}"
      - name: "Verify checksum"
        command: "{{sai_binary(0, 'verify_cmd')}}"
      - name: "Extract archive"
        command: "{{sai_binary(0, 'extract_cmd')}}"
      - name: "Install binary"
        command: "{{sai_binary(0, 'install_cmd')}}"
        requires_root: true
    timeout: 600
    validation:
      command: "{{sai_binary(0, 'validation_cmd')}}"
    rollback: "{{sai_binary(0, 'uninstall_cmd')}}"
```

**Key Template Functions:**
- `{{sai_binary(index, field)}}` - Access binary configuration
- OS/architecture templating in URLs
- Automatic archive extraction (zip, tar.gz, etc.)
- Checksum verification for security

### Script Provider

The script provider executes installation scripts with safety measures:

```yaml
version: "1.0"
provider:
  name: "script"
  display_name: "Script Installation Provider"
  type: "script"
  platforms: ["linux", "darwin", "windows"]
  capabilities: ["install", "uninstall", "version", "info"]
  executable: "bash"  # Or sh, python, etc.

actions:
  install:
    description: "Execute installation script"
    steps:
      - name: "Download script"
        command: "{{sai_script(0, 'download_cmd')}}"
      - name: "Verify script"
        command: "{{sai_script(0, 'verify_cmd')}}"
      - name: "Execute script"
        command: "{{sai_script(0, 'install_cmd')}}"
        requires_root: true
    timeout: 900
    validation:
      command: "{{sai_script(0, 'validation_cmd')}}"
    rollback: "{{sai_script(0, 'uninstall_cmd')}}"
```

**Key Template Functions:**
- `{{sai_script(index, field)}}` - Access script configuration
- Environment variable management
- Interactive prompt handling
- Security verification with checksums

### Alternative Provider SaiData Configuration

These providers require specific saidata configurations:

```yaml
# Example: nginx with source build
version: "0.2"
metadata:
  name: "nginx"
  description: "High-performance web server"

# Source build configuration
sources:
  - name: "main"
    url: "http://nginx.org/download/nginx-{{version}}.tar.gz"
    version: "1.24.0"
    build_system: "autotools"
    prerequisites: ["build-essential", "libssl-dev", "libpcre3-dev"]
    configure_args: ["--with-http_ssl_module", "--with-http_v2_module"]

# Binary download configuration  
binaries:
  - name: "main"
    url: "https://github.com/nginx/nginx/releases/download/v{{version}}/nginx-{{version}}-{{os}}-{{arch}}.tar.gz"
    version: "1.24.0"
    executable: "nginx"
    checksum: "sha256:abc123..."

# Script installation configuration
scripts:
  - name: "main"
    url: "https://nginx.org/packages/install.sh"
    interpreter: "bash"
    arguments: "--version {{version}}"
    checksum: "sha256:def456..."
```

### Security Considerations for Alternative Providers

1. **Source Builds**: Verify source integrity, use trusted mirrors
2. **Binary Downloads**: Always verify checksums, use HTTPS URLs
3. **Script Execution**: Require user consent, verify script signatures
4. **Sandboxing**: Consider containerized builds for isolation
5. **Rollback**: Implement proper cleanup and rollback procedures

## Examples

### Package Manager Provider

```yaml
version: "1.0"
provider:
  name: "custom-pkg"
  display_name: "Custom Package Manager"
  type: "package_manager"
  platforms: ["linux"]
  executable: "custompkg"
  priority: 75

actions:
  install:
    description: "Install package"
    template: "custompkg install {{sai_package}}"
    requires_root: true
    validation:
      command: "custompkg list | grep -q '^{{sai_package}}$'"
    rollback: "custompkg remove {{sai_package}}"
    
  search:
    description: "Search packages"
    template: "custompkg search {{.Software}}"
    requires_root: false
```

### Container Provider

```yaml
version: "1.0"
provider:
  name: "podman"
  display_name: "Podman Container Engine"
  type: "container"
  platforms: ["linux", "darwin"]
  executable: "podman"

actions:
  install:
    description: "Pull container image"
    template: "podman pull {{sai_package}}"
    
  start:
    description: "Start container"
    template: "podman run -d --name {{.Software}} {{sai_package}}"
    
  stop:
    description: "Stop container"
    template: "podman stop {{.Software}}"
    
  status:
    description: "Check container status"
    template: "podman ps -a --filter name={{.Software}}"
```

### Specialized Tool Provider

```yaml
version: "1.0"
provider:
  name: "security-scanner"
  display_name: "Security Scanner"
  type: "security"
  platforms: ["linux", "darwin"]
  executable: "scanner"

actions:
  check:
    description: "Run security scan"
    template: "scanner scan --target {{sai_file 'binary'}} --format json"
    timeout: 300
    
  info:
    description: "Show security information"
    template: "scanner info {{.Software}}"
```

## Troubleshooting

### Common Issues

1. **Provider Not Found**
   - Check file location and naming
   - Verify YAML syntax
   - Ensure executable is in PATH

2. **Template Errors**
   - Validate template syntax
   - Check saidata availability
   - Test with dry-run mode

3. **Validation Failures**
   - Check validation commands
   - Verify expected output format
   - Test validation independently

4. **Platform Issues**
   - Verify platform support
   - Test on target platforms
   - Check platform-specific commands

### Debugging Tips

```bash
# Enable verbose output
sai install nginx --provider my-provider --verbose

# Use dry-run to see generated commands
sai install nginx --provider my-provider --dry-run

# Check provider availability
sai stats --provider my-provider

# Validate provider file
sai validate-provider providers/my-provider.yaml
```

### Getting Help

- Check existing providers for examples
- Review the schema documentation
- Ask questions in GitHub Discussions
- Submit issues for bugs or feature requests

## Contributing Providers

1. Fork the SAI repository
2. Create your provider file
3. Test thoroughly on supported platforms
4. Add documentation and examples
5. Submit a pull request

Your contributions help make SAI more useful for everyone!

---

For more information, see the [main documentation](../README.md) or visit the [SAI CLI website](https://example42.com).