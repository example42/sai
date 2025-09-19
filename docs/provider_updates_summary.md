# Provider Updates Summary

## SAI Function Standardization Implementation

This document summarizes the updates made to SAI providers to use the new standardized template functions.

## Updated Providers

### Core Package Managers
- **apt.yaml** - Debian/Ubuntu package manager
- **brew.yaml** - macOS Homebrew package manager  
- **dnf.yaml** - Fedora/RHEL package manager
- **snap.yaml** - Universal Linux packages
- **npm.yaml** - Node.js package manager

### Container & Orchestration
- **docker.yaml** - Docker container platform
- **helm.yaml** - Kubernetes package manager

### Specialized Tools
- **cat.yaml** - File content viewer
- **lsof.yaml** - List open files
- **strace.yaml** - System call tracer
- **system-admin.yaml** - Comprehensive system administration (new)

## Key Changes Made

### 1. Package Function Updates
**Before:**
```yaml
template: "apt-get install -y {{sai_packages('apt')}}"
detection: "apt-cache show {{sai_package('apt')}} >/dev/null 2>&1"
```

**After:**
```yaml
template: "apt-get install -y {{sai_package('*', 'name', 'apt')}}"
detection: "apt-cache show {{sai_package(0, 'name', 'apt')}} >/dev/null 2>&1"
```

### 2. Service Function Updates
**Before:**
```yaml
template: "systemctl start {{sai_service('service_name')}}"
```

**After:**
```yaml
template: "systemctl start {{sai_service(0, 'service_name', 'apt')}}"
```

### 3. Port Function Updates
**Before:**
```yaml
template: "docker create -p {{sai_port()}}:{{sai_port()}}"
```

**After:**
```yaml
template: "docker create -p {{sai_port(0, 'port', 'docker')}}:{{sai_port(0, 'port', 'docker')}}"
```

### 4. File Function Updates
**Before:**
```yaml
template: "cat '{{sai_file('config','path')}}'"
```

**After:**
```yaml
template: "cat '{{sai_file('config', 'path', 'cat')}}'"
```

### 5. Container Function Updates
**Before:**
```yaml
template: "docker pull {{sai_package() or metadata.name}}:{{sai_package('version') or 'latest'}}"
```

**After:**
```yaml
template: "docker pull {{sai_container(0, 'image', 'docker')}}:{{sai_container(0, 'tag', 'docker')}}"
```

## Benefits Achieved

### 1. Consistency
- All functions now follow the same `sai_{resource}(selector, key)` pattern
- Predictable behavior across all providers
- Easier to learn and maintain

### 2. Flexibility
- **Index access**: `sai_package(0, 'name')` for first package
- **Name access**: `sai_file('config', 'path')` for specific files
- **Wildcard access**: `sai_port('*', 'port')` for all ports

### 3. Shell-Friendly Output
- Wildcard functions return space-separated values
- Perfect for shell commands: `systemctl status {{sai_service('*', 'service_name')}}`
- Enables bulk operations: `tar -czf backup.tar.gz {{sai_file('*', 'path')}}`

### 4. Template Safety
- Empty returns disable invalid actions
- Graceful degradation when data is missing
- No partial command execution

### 5. Enhanced Capabilities
- **Multi-resource operations**: Backup all config files at once
- **Comprehensive monitoring**: Check all ports simultaneously  
- **Bulk management**: Start/stop multiple services
- **Complex workflows**: Chain operations across resource types

## Example Use Cases

### Backup All Configuration Files
```yaml
backup:
  template: "tar -czf backup.tar.gz {{sai_file('*', 'path', 'system-admin')}}"
```

### Monitor All Service Ports
```yaml
monitoring:
  template: "ss -tlnp | grep -E ':({{sai_port('*', 'port', 'system-admin')}})''"
```

### Set Proper File Permissions
```yaml
permissions:
  steps:
    - command: "chown {{sai_file('config', 'owner', 'apt')}}:{{sai_file('config', 'group', 'apt')}} {{sai_file('config', 'path', 'apt')}}"
    - command: "chmod {{sai_file('config', 'mode', 'apt')}} {{sai_file('config', 'path', 'apt')}}"
```

### Comprehensive Health Check
```yaml
health-check:
  template: |
    systemctl is-active {{sai_service(0, 'service_name', 'apt')}} &&
    nc -zv localhost {{sai_port(0, 'port', 'docker')}} &&
    test -f {{sai_file('config', 'path', 'apt')}} &&
    docker ps | grep {{sai_container(0, 'name', 'docker')}}
```

## Migration Guide for Remaining Providers

To update other providers, replace (where `{provider}` is the provider name):

1. `{{sai_packages('provider')}}` → `{{sai_package('*', 'name', '{provider}')}}`
2. `{{sai_package('provider')}}` → `{{sai_package(0, 'name', '{provider}')}}`
3. `{{sai_service('service_name')}}` → `{{sai_service(0, 'service_name', '{provider}')}}`
4. `{{sai_port()}}` → `{{sai_port(0, 'port', '{provider}')}}`
5. `{{sai_file()}}` → `{{sai_file(0, 'path', '{provider}')}}`
6. `{{sai_command('name')}}` → `{{sai_command(0, 'path', '{provider}')}}`

### Provider Parameter Benefits

The third parameter ensures proper template resolution hierarchy:
- **Provider-specific overrides**: Gets values from `saidata.providers.{provider}.{resource}`
- **Fallback to defaults**: Uses `saidata.{resource}` when provider overrides don't exist
- **OS-specific resolution**: Works with OS override files for both provider and default sections

## Next Steps

1. **Complete Migration**: Update remaining providers using the standardized functions
2. **Test Templates**: Validate all template resolutions work correctly
3. **Documentation**: Update provider documentation with new function examples
4. **Schema Updates**: Ensure provider schemas reflect the new function patterns

The standardized SAI functions provide a powerful, consistent foundation for all provider templates while maintaining backward compatibility and enabling new capabilities for complex system management tasks.