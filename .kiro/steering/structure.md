# SAI Project Structure

## Root Directory Layout

```
├── docs/                    # Documentation and examples
│   ├── sai_synopsis.md     # Main product documentation
│   └── saidata_samples/    # Example saidata configurations
├── providers/              # Provider implementation files
│   ├── specialized/        # Specialized operational providers
│   └── *.yaml             # Standard provider definitions
└── schemas/               # JSON Schema validation files
```

## Key Directories

### `/docs/`
- **sai_synopsis.md**: Complete CLI reference and feature documentation
- **saidata_samples/**: Example software configurations organized by prefix
  - Pattern: `{prefix}/{software}/default.yaml` (e.g., `el/elasticsearch/default.yaml`)
  - Contains real-world examples for major software packages

### `/providers/`
- **Standard Providers**: Package managers and installation methods
  - `apt.yaml`, `brew.yaml`, `dnf.yaml`, `docker.yaml`, etc.
  - Each provider defines actions, capabilities, and platform support
- **specialized/**: Operational and debugging providers
  - `README.md`: Comprehensive guide to specialized providers
  - `CONSOLIDATION_SUMMARY.md`: Provider consolidation documentation
  - Individual provider files for debugging, security, monitoring, etc.

### `/schemas/`
- **saidata-0.2-schema.json**: Software definition schema
- **providerdata-0.1-schema.json**: Provider implementation schema  
- **applydata-0.1-schema.json**: Batch operation schema
- **repository-config-schema.json**: Repository configuration schema

## File Naming Conventions

### Provider Files
- **Standard**: `{provider-name}.yaml` (e.g., `apt.yaml`, `docker.yaml`)
- **Specialized**: `{tool-name}.yaml` in `specialized/` directory

### SaiData Files
- **Base Pattern**: `software/{prefix}/{software}/default.yaml`
- **OS Override Pattern**: `software/{prefix}/{software}/{os}/{os_version}.yaml`
- **Prefix**: First 2 characters of software name
- **Examples**: 
  - `ap/apache/default.yaml` - Base Apache configuration
  - `ap/apache/ubuntu/22.04.yaml` - Ubuntu 22.04-specific overrides
  - `ap/apache/centos/8.yaml` - CentOS 8-specific overrides
  - `ap/apache/macos/13.yaml` - macOS 13-specific Apache configuration

## Configuration Structure

### Provider YAML Structure
```yaml
version: "1.0"
provider:
  name: "provider-name"
  type: "package_manager|container|specialized-type"
  platforms: ["linux", "macos", "windows"]
  capabilities: ["install", "uninstall", "start", "stop", ...]
actions:
  install:
    description: "Action description"
    template: "command template with {{variables}}"
    validation: {...}
    rollback: "rollback command"
```

### SaiData YAML Structure
```yaml
version: "0.2"
metadata:
  name: "software-name"
  description: "Software description"
packages:
  - name: "package-name"
services:
  - name: "service-name"
providers:
  provider-name:
    packages: [...]  # Provider-specific overrides
```

## Configuration Hierarchy

### OS-Specific Overrides
SAI supports OS-specific configuration overrides that extend the base saidata structure:

- **Base Configuration**: `software/{prefix}/{software}/default.yaml`
- **OS Overrides**: `software/{prefix}/{software}/{os}/{os_version}.yaml`
- **Supported OS Types**: `ubuntu`, `debian`, `centos`, `rocky`, `fedora`, `macos`, `windows`
- **Version Targeting**: Major version files like `22.yaml`, `8.yaml`, `11.yaml`

### Configuration Merging Process
1. **Environment Autodetection**: SAI automatically detects platform, OS, and OS version
2. Load base `default.yaml` configuration
3. Load matching OS-specific override file if it exists (based on autodetected environment)
4. Deep merge OS configuration with base configuration
5. OS-specific values take precedence over defaults

### Autodetection Details
- **Automatic**: No manual configuration required - SAI detects the environment on each run
- **Cached**: Detection results are cached to improve performance on subsequent executions
- **Reliable**: Uses multiple detection methods per platform for accuracy
- **Transparent**: Users don't need to specify their OS - SAI handles it automatically

### Override Examples
```
software/ap/apache/
├── default.yaml              # Base Apache configuration
├── ubuntu/
│   ├── 20.04.yaml           # Ubuntu 20.04-specific settings
│   └── 22.04.yaml           # Ubuntu 22.04-specific settings
├── centos/
│   ├── 7.yaml               # CentOS 7-specific settings
│   └── 8.yaml               # CentOS 8-specific settings
├── macos/
│   └── 13.yaml              # macOS 13-specific overrides
└── windows/
    └── 11.yaml              # Windows 11-specific overrides
```

## Development Patterns

### Adding New Providers
1. Create `providers/{name}.yaml` following providerdata schema
2. Define provider metadata, capabilities, and platform support
3. Implement required actions with templates and validation
4. Add to specialized/ directory if operational tool

### Adding New Software
1. Create `docs/saidata_samples/{prefix}/{software}/default.yaml`
2. Follow saidata schema with metadata, packages, services
3. Include provider-specific overrides as needed
4. Add OS-specific override files when needed:
   - `{prefix}/{software}/{os}/{os_version}.yaml` for OS version-specific settings
5. Add compatibility matrix for supported platforms

### Schema Validation
- All YAML files must validate against their respective JSON schemas
- Use schema references for consistent structure
- Include required fields and proper data types