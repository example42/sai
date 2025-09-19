# SAI Technical Stack

## Architecture

- **Configuration Format**: YAML-based configuration files for all data structures
- **Schema Validation**: JSON Schema validation for saidata (v0.2) and providerdata (v0.1) formats
- **Template Engine**: Variable substitution using `{{variable}}` syntax with sai-specific functions
- **Environment Autodetection**: Automatic detection of platform, OS, and OS version with intelligent caching
- **Cross-Platform**: Native support for Linux, macOS, and Windows

## Data Schemas

- **saidata-0.2-schema.json**: Defines software metadata, packages, services, files, directories, commands, ports, containers, and provider-specific overrides
- **providerdata-0.1-schema.json**: Defines provider implementations with actions, mappings, validation, and rollback capabilities
- **applydata-0.1-schema.json**: Batch operation definitions
- **repository-config-schema.json**: Repository configuration structure

## Provider Types

**Standard Providers**: package_manager, container, binary, source, cloud, custom
**Specialized Providers**: debug, trace, profile, security, sbom, troubleshoot, network, audit, backup, filesystem, system, monitoring, io, memory, monitor, process

## Template Variables

SAI uses a custom templating system with functions like:
- `{{sai_package('provider')}}` - Get package name for specific provider
- `{{sai_packages('provider')}}` - Get all packages for provider
- `{{sai_service('service_name')}}` - Get service name
- `{{sai_port()}}` - Get default port
- `{{container_name}}` - Generated container names


### Template Resolution Order
Template functions follow a hierarchical resolution order with OS-specific overrides:
1. **OS-specific provider overrides**: `saidata.providers.{provider}.{key}` from OS override file
2. **Default provider overrides**: `saidata.providers.{provider}.{key}` from default file
3. **OS-specific defaults**: `saidata.{key}` from OS override file
4. **Base defaults**: `saidata.{key}` from default file

**Example**: `{{sai_package('apt')}}` on Ubuntu 22.04 will look for:
1. `software/ap/apache/ubuntu/22.04.yaml` → `providers.apt.packages[0].name`
2. `software/ap/apache/default.yaml` → `providers.apt.packages[0].name`
3. `software/ap/apache/ubuntu/22.04.yaml` → `packages[0].name`
4. `software/ap/apache/default.yaml` → `packages[0].name`

### Template Variable Resolution Behavior

**Critical Rule**: If no value is found for a saidata variable or function during template resolution, the command template that references it becomes invalid and the corresponding action will not be available for that software.

**Behavior Details**:
- **Missing Variables**: When `{{sai_service('service_name')}}`, `{{sai_package('provider')}}`, `{{sai_port()}}`, or other sai functions cannot resolve to a value, the entire action is disabled
- **Action Availability**: Only actions with fully resolvable templates are made available to users
- **Graceful Degradation**: Software can still be managed through other providers/actions that have complete saidata definitions
- **No Partial Execution**: Commands with unresolved variables will not execute with empty or null values

**Examples**:
- If `sai_service('service_name')` returns no value, service management actions (start, stop, restart, enable, disable, status, logs) become unavailable
- If `sai_package('docker')` returns no value, Docker-based installation actions are disabled
- If `sai_port()` returns no value, port-specific monitoring or network actions are not available

**Best Practices**:
- Ensure complete saidata definitions for all intended use cases
- Use provider-specific overrides to handle edge cases
- Test template resolution across different OS environments
- Provide fallback providers when possible for broader compatibility

## Configuration Patterns

- **Hierarchical Structure**: `software/{prefix}/{software}/default.yaml`
- **OS-Specific Overrides**: `software/{prefix}/{software}/{os}/{os_version}.yaml`
- **Automatic Override Selection**: SAI autodetects local environment and selects appropriate overrides
- **Configuration Merging**: Deep merge of OS-specific files with defaults, OS values take precedence
- **Provider Overrides**: Provider-specific configurations can override defaults
- **Compatibility Matrix**: Platform/architecture/version compatibility definitions
- **Repository Management**: Automatic repository setup and management

## Environment Detection

SAI performs automatic environment detection on each execution:

### Detection Process
1. **Platform Identification**: Detects underlying platform (linux, macos, windows)
2. **OS Distribution Detection**: Identifies specific OS (ubuntu, debian, centos, rocky, fedora, etc.)
3. **Version Resolution**: Determines OS major version (22.04, 8, 13.0, etc.)
4. **Caching Strategy**: Stores detection results to optimize performance on subsequent runs

### Detection Methods
- **Linux**: Parses `/etc/os-release`, `/etc/lsb-release`, and distribution-specific files
- **macOS**: Uses `sw_vers` command and system version files
- **Windows**: Leverages WMI queries and registry information

### Cache Management
- **Cache Location**: Stores detection results in user-specific cache directory
- **Cache Invalidation**: Automatically refreshes when system changes are detected
- **Performance Optimization**: Reduces detection overhead from seconds to milliseconds

## Common Commands

```bash
# Basic software management
sai install <software> [--provider <name>]
sai uninstall <software> [--provider <name>]
sai upgrade <software>

# Service management
sai start <software>
sai stop <software>
sai status <software>

# Batch operations
sai apply <action_file>

# System information
sai stats
sai list

# Global options
--config/-c <path>    # Custom config file
--provider/-p <name>  # Force specific provider
--verbose/-v          # Detailed output
--dry-run            # Show commands without executing
--yes/-y             # Auto-confirm prompts
--quiet/-q           # Suppress output
--json               # JSON output format
```

## Development Guidelines

- All YAML files must validate against their respective JSON schemas
- Provider actions support templates, commands, scripts, or multi-step execution
- Include validation commands and rollback procedures for destructive actions
- Use consistent variable naming and template patterns
- Support timeout, retry, and error handling configurations
- **Template Completeness**: Ensure all saidata variables and functions used in provider templates can be resolved, as actions with unresolvable templates will be automatically disabled
- **Graceful Degradation**: Design saidata configurations to support multiple providers so software remains manageable even if some actions are unavailable due to missing template variables