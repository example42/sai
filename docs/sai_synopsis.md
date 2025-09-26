# SAI

A lightweight CLI tool for executing software management actions using provider-based configurations.
"Do everything on every software on every OS"

## Specific Software Actions (Based on Provider Capabilities)

SYNOPSIS:
sai <action> <software>

install - Install software packages using available providers
uninstall - Remove/uninstall software packages from the system
upgrade - Update software packages to their latest versions
search - Search for available software packages in repositories
info - Display detailed information about software
version - Show software version

start - Start software services or applications
stop - Stop running software services or applications
restart - Restart software services (stop and start)
enable - Enable automatic startup of services at boot
disable - Disable automatic startup of services at boot
status - Check the current status of software services

logs - Display service logs and output
config - Show configuration files
check - Checks if software is working correctly
cpu - Display CPU usage statistics and performance metrics of the running software
memory - Show memory usage and virtual memory statistics of the running software
io - Display I/O statistics and disk performance metrics of the running software

## General Software Actions

SYNOPSIS:
sai <action>

list - List installed software packages
logs - Display service logs and output
cpu - Display general CPU usage statistics
memory - Show general memory usage 
io - Display general I/O statistics

## Other sai commands

apply <action_file> - Execute multiple software management actions from YAML/JSON file based on schemas/applydata
stats - Display comprehensive statistics about available providers, actions, and system capabilities with detailed breakdowns
saidata - Manages saidata (update saidata repo, show saidata information for a software...)

## Global Options (Available for all commands)

--config/-c <path> - Specify custom sai configuration file path
--provider/-p <name> - Force usage of specific provider instead of auto-detection
--verbose/-v - Enable detailed output and logging information
--dry-run - Show what would be executed without actually running commands
--yes/-y - Automatically confirm all prompts without user interaction
--quiet/-q - Suppress non-essential output for scripting
--json - Output results in JSON format for programmatic consumption

## Environment Autodetection

SAI automatically detects your system environment without requiring manual configuration:

### What SAI Detects
- **Platform**: Hardware/OS family (linux, macos, windows)
- **Operating System**: Specific distribution (ubuntu, debian, centos, rocky, fedora, macos, windows)
- **OS Version**: Major version numbers (22.04, 8, 13.0, etc.)

### Detection Methods
- **Linux**: Analyzes `/etc/os-release`, `/etc/lsb-release`, and distribution files
- **macOS**: Uses `sw_vers` and system version information
- **Windows**: Queries WMI and registry data

### Performance Optimization
- **Intelligent Caching**: Detection results are cached to avoid repeated system queries
- **Fast Execution**: Cached results reduce overhead from seconds to milliseconds
- **Automatic Refresh**: Cache is invalidated when system changes are detected

### Override Selection
Based on detected environment, SAI automatically selects the most specific configuration:
1. Tries OS-specific override: `software/{prefix}/{software}/{os}/{os_version}.yaml`
2. Falls back to base configuration: `software/{prefix}/{software}/default.yaml`
3. Deep merges configurations with OS-specific values taking precedence

## Alternative Installation Methods

SAI supports three alternative installation methods beyond traditional package managers:

### Source Builds
Build software from source code with automatic build system detection:

```bash
# Build nginx from source
sai install nginx --provider source

# Build with custom configuration
sai install nginx --provider source --verbose
```

**Supported Build Systems**: autotools, cmake, make, meson, ninja, custom

### Binary Downloads
Download and install pre-compiled binaries with OS/architecture detection:

```bash
# Download terraform binary
sai install terraform --provider binary

# Install specific version
sai install terraform --provider binary --version 1.5.7
```

**Features**: Automatic OS/arch detection, checksum verification, archive extraction

### Script Installation
Execute installation scripts with safety measures:

```bash
# Run Docker installation script
sai install docker --provider script

# Execute with automatic confirmation
sai install docker --provider script --yes
```

**Security**: Checksum verification, user consent required, rollback support

## Features

- Hierarchical saidata structure support (software/{prefix}/{software}/default.yaml)
- OS-specific overrides support (software/{prefix}/{software}/{os}/{os_version}.yaml)
- Alternative installation methods (source, binary, script) with comprehensive template functions
- Automatic platform, OS, and OS version detection with intelligent caching
- Automatic provider detection and prioritization
- Automatic software repositories management (when defined in saidata)
- Git-based saidata repository management with zip fallback (default: https://github.com/example42/saidata)
- Cross-platform compatibility (Linux, macOS, Windows)
- All actions which change something on the system require, by default, user confirmation
- All action that just show information are unattended

## SaiData Configuration Hierarchy

SAI supports a flexible configuration hierarchy that allows OS-specific customizations:

1. **Base Configuration**: `software/{prefix}/{software}/default.yaml`
   - Contains the default software configuration applicable to all operating systems

2. **OS-Specific Overrides**: `software/{prefix}/{software}/{os}/{os_version}.yaml`
   - OS-specific configurations that override defaults
   - Supported OS types: `ubuntu`, `debian`, `centos`, `rocky`, `fedora`, `macos`, `windows`
   - Version-specific overrides for major versions (e.g., `ubuntu/22.04.yaml`, `centos/8.yaml`)

3. **Configuration Merging**: OS-specific files are deep-merged with defaults
   - OS-specific values take precedence over defaults
   - Allows fine-tuning of packages, services, and provider configurations per OS
   - Enables version-specific compatibility handling

**Examples**:
- `software/ap/apache/default.yaml` - Base Apache configuration
- `software/ap/apache/ubuntu/22.04.yaml` - Ubuntu 22.04-specific Apache settings
- `software/ap/apache/centos/8.yaml` - CentOS 8-specific Apache settings
- `software/ap/apache/macos/13.yaml` - macOS 13-specific Apache configuration

## Environment Autodetection

SAI automatically detects the local environment when executed:

- **Platform Detection**: Identifies hardware/OS family (linux, macos, windows)
- **OS Detection**: Determines specific operating system (ubuntu, debian, centos, rocky, fedora, macos, windows)
- **OS Version Detection**: Detects major version (22.04, 8, 13.0, etc.)
- **Intelligent Caching**: Caches detection results to improve performance on subsequent runs
- **Override Selection**: Uses detected information to automatically select appropriate OS-specific overrides

**Configuration Merging Example**:
When installing Apache on Ubuntu 22.04, SAI will:
1. **Autodetect Environment**: Platform=linux, OS=ubuntu, OS_Version=22.04
2. Load `software/ap/apache/default.yaml` (base configuration)
3. Load `software/ap/apache/ubuntu/22.04.yaml` (Ubuntu 22.04-specific overrides)
4. Deep merge the configurations where Ubuntu-specific values override defaults:
   - Package name remains `apache2` (from default)
   - Service type becomes `systemd` with Ubuntu-specific auto-restart (from override)
   - Adds Ubuntu-specific commands like `a2ensite`, `a2enmod` (from override)
   - Includes Ubuntu-specific directories like `/etc/apache2/sites-available` (from override)
