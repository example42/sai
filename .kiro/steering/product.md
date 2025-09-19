# SAI Product Overview

SAI is a lightweight CLI tool for executing software management actions using provider-based configurations. The core philosophy is "Do everything on every software on every OS" through a unified interface.

## Key Concepts

- **Universal Software Management**: Single CLI interface for install, uninstall, upgrade, start, stop, restart, enable, disable, status, logs, and monitoring actions across all software and operating systems
- **Provider-Based Architecture**: Extensible system supporting package managers (apt, brew, dnf), containers (docker, helm), and specialized tools (debugging, security, monitoring)
- **SaiData Repository**: Hierarchical software definitions stored as YAML files following the pattern `software/{prefix}/{software}/default.yaml` with OS-specific overrides support via `software/{prefix}/{software}/{os}/{os_version}.yaml`
- **Intelligent Environment Detection**: Automatic detection of platform, OS, and OS version with caching for optimal performance
- **Cross-Platform Compatibility**: Works on Linux, macOS, and Windows with automatic provider detection and prioritization

## Core Actions

**Software Management**: install, uninstall, upgrade, search, info, version, list
**Service Management**: start, stop, restart, enable, disable, status, logs
**System Monitoring**: cpu, memory, io, check
**Advanced Operations**: apply (batch operations), stats, saidata management

## Specialized Providers

Beyond basic software management, SAI includes specialized providers for:
- **Security**: vulnerability scanning, SBOM generation, auditing
- **Debugging**: GDB integration, performance profiling, system tracing
- **Operations**: backup/restore, network analysis, troubleshooting
- **Monitoring**: resource usage, performance metrics, log analysis

## Target Users

DevOps engineers, system administrators, and developers who need consistent software management across heterogeneous environments.