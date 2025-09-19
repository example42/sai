# SAI Template Functions Quick Reference

## Function Overview

| Function | Resource | Purpose |
|----------|----------|---------|
| `sai_file(selector, key, provider)` | files | Access file definitions (paths, permissions, etc.) |
| `sai_service(selector, key, provider)` | services | Access service definitions (names, types, etc.) |
| `sai_package(selector, key, provider)` | packages | Access package definitions (names, versions, etc.) |
| `sai_directory(selector, key, provider)` | directories | Access directory definitions (paths, permissions, etc.) |
| `sai_command(selector, key, provider)` | commands | Access command definitions (paths, arguments, etc.) |
| `sai_port(selector, key, provider)` | ports | Access port definitions (numbers, protocols, etc.) |
| `sai_container(selector, key, provider)` | containers | Access container definitions (images, tags, etc.) |

## Selector Types

| Selector | Description | Example |
|----------|-------------|---------|
| `0, 1, 2...` | Array index (0-based) | `sai_file(0, 'path')` |
| `"name"` | Match by name field | `sai_file('config', 'path')` |
| `"*"` | All items (space-separated) | `sai_file('*', 'path')` |

## Common Keys by Resource

### Files
- `name`, `path`, `type`, `owner`, `group`, `mode`, `backup`

### Services  
- `name`, `service_name`, `type`, `enabled`, `config_files`

### Packages
- `name`, `version`, `alternatives`, `install_options`, `repository`, `checksum`, `signature`, `download_url`

### Directories
- `name`, `path`, `owner`, `group`, `mode`, `recursive`

### Commands
- `name`, `path`, `arguments`, `aliases`, `shell_completion`, `man_page`

### Ports
- `port`, `protocol`, `service`, `description`

### Containers
- `name`, `image`, `tag`, `registry`, `platform`, `ports`, `volumes`, `environment`, `networks`, `labels`

## Quick Examples

```yaml
# File operations
{{sai_file('config', 'path')}}                    # /etc/apache2/apache2.conf
{{sai_file('config', 'path', 'apt')}}             # /etc/apache2/apache2.conf (from apt provider)
{{sai_file('*', 'path')}}                         # All file paths

# Service management
{{sai_service('apache', 'service_name')}}         # apache2
{{sai_service('apache', 'service_name', 'brew')}} # httpd (from brew provider)
{{sai_service('*', 'service_name')}}              # All service names

# Package info
{{sai_package('apache2', 'version')}}             # 2.4.58
{{sai_package('apache2', 'version', 'dnf')}}      # 2.4.58-1.el8 (from dnf provider)
{{sai_package('*', 'name')}}                      # All package names

# Directory paths
{{sai_directory('config', 'path')}}               # /etc/apache2
{{sai_directory('config', 'path', 'brew')}}       # /opt/homebrew/etc/httpd (from brew)
{{sai_directory('*', 'path')}}                    # All directory paths

# Command paths
{{sai_command('apache2', 'path')}}                # /usr/sbin/apache2
{{sai_command('httpd', 'path', 'brew')}}          # /opt/homebrew/bin/httpd (from brew)
{{sai_command('*', 'path')}}                      # All command paths

# Port numbers
{{sai_port('http', 'port')}}                      # 80
{{sai_port('http', 'port', 'docker')}}            # 8080 (from docker provider)
{{sai_port('*', 'port')}}                         # All port numbers

# Container images
{{sai_container('app', 'image')}}                 # httpd
{{sai_container('app', 'image', 'docker')}}       # httpd:2.4-alpine (from docker provider)
{{sai_container('*', 'image')}}                   # All container images
```

## Special Cases

### Backward Compatibility
```yaml
{{sai_port()}}                           # First port number (legacy)
```

### Port Matching
```yaml
{{sai_port('80', 'service')}}            # Match by port number
{{sai_port('http', 'port')}}             # Match by service name
```

### Array Fields
For wildcard access on array fields, returns space-separated values:
```yaml
{{sai_package('*', 'alternatives')}}     # "httpd apache2-bin apache2-utils"
```

## Template Resolution Order

1. OS-specific provider overrides: `providers.{provider}.{resource}[match].{key}`
2. Default provider overrides: `providers.{provider}.{resource}[match].{key}`
3. OS-specific defaults: `{resource}[match].{key}` (from OS override file)
4. Base defaults: `{resource}[match].{key}` (from default file)

## Error Behavior

- Missing selector/key → Empty string → Action disabled
- No matches → Empty string → Action disabled  
- Multiple name matches → First match returned
- Invalid data → Empty string → Action disabled

This ensures graceful degradation and template safety across all SAI functions.