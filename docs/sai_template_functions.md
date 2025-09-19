# SAI Template Functions Specification

This document defines the standardized SAI template functions used in provider configurations for accessing saidata properties.

## Function Pattern

All SAI functions follow the same standardized pattern:

```
sai_{resource}(selector, key, provider)
```

### Parameters

1. **selector** (string|integer):
   - **Integer (0-based index)**: Access by array position
   - **String (name match)**: Access by name field value
   - **Wildcard (*)**: Access all items

2. **key** (string): The property to extract from the resource object

3. **provider** (string, optional): The provider context for template resolution
   - When specified, looks for provider-specific overrides first
   - When omitted, uses default resolution hierarchy

### Return Values

- **Single match**: Returns the string value of the requested key
- **Wildcard match**: Returns space-separated list of values (shell-friendly)
- **No match**: Returns empty string (disables the action template)

### Template Resolution Hierarchy

All functions follow SAI's hierarchical resolution order. When the `provider` parameter is specified:

1. **OS-specific provider overrides**: `saidata.providers.{provider}.{resource}[match].{key}` (from OS override file)
2. **Default provider overrides**: `saidata.providers.{provider}.{resource}[match].{key}` (from default file)
3. **OS-specific defaults**: `saidata.{resource}[match].{key}` (from OS override file)
4. **Base defaults**: `saidata.{resource}[match].{key}` (from default file)

When the `provider` parameter is omitted, skips steps 1-2 and uses only the default hierarchy.

---

## sai_file(selector, key, provider)

Access file definitions from the `files` array.

### Available Keys
- `name` - Logical file name (e.g., "config", "log", "binary")
- `path` - File system path
- `type` - File type ("config", "binary", "library", "data", "log", "temp", "socket")
- `owner` - File owner
- `group` - File group
- `mode` - File permissions
- `backup` - Backup flag (boolean)

### Examples
```yaml
# Index access
sai_file(0, 'path')                    # "/etc/apache2/apache2.conf"
sai_file(1, 'owner', 'apt')            # "www-data" (from apt provider override)

# Name access
sai_file('config', 'path')             # "/etc/apache2/apache2.conf"
sai_file('log', 'mode', 'docker')      # "0640" (from docker provider override)

# Wildcard access
sai_file('*', 'path')                  # "/etc/apache2/apache2.conf /var/log/apache2/access.log"
sai_file('*', 'name', 'brew')          # "config log error_log" (from brew provider)
```

---

## sai_service(selector, key, provider)

Access service definitions from the `services` array.

### Available Keys
- `name` - Logical service name
- `service_name` - Actual system service name
- `type` - Service type ("systemd", "init", "launchd", "windows_service", "docker", "kubernetes")
- `enabled` - Auto-start flag (boolean)
- `config_files` - Array of configuration file paths

### Examples
```yaml
# Index access
sai_service(0, 'service_name')         # "apache2"
sai_service(0, 'type', 'apt')          # "systemd" (from apt provider)

# Name access
sai_service('apache', 'service_name')  # "apache2"
sai_service('nginx', 'enabled', 'brew') # "true" (from brew provider)

# Wildcard access
sai_service('*', 'service_name')       # "apache2 mysql redis"
sai_service('*', 'name', 'docker')     # "apache mysql redis" (from docker provider)
```

---

## sai_package(selector, key, provider)

Access package definitions from the `packages` array.

### Available Keys
- `name` - Package name
- `version` - Package version
- `alternatives` - Array of alternative package names
- `install_options` - Installation options/flags
- `repository` - Repository name
- `checksum` - Package checksum
- `signature` - Package signature
- `download_url` - Direct download URL

### Examples
```yaml
# Index access
sai_package(0, 'name')                 # "apache2"
sai_package(0, 'version', 'apt')       # "2.4.58" (from apt provider)

# Name access (matches package name)
sai_package('apache2', 'version')      # "2.4.58"
sai_package('nginx', 'repository', 'dnf') # "main" (from dnf provider)

# Wildcard access
sai_package('*', 'name')               # "apache2 mysql-server redis-server"
sai_package('*', 'version', 'brew')    # "2.4.58 8.0.35 7.0.12" (from brew provider)
```

---

## sai_directory(selector, key)

Access directory definitions from the `directories` array.

### Available Keys
- `name` - Logical directory name (e.g., "config", "log", "data", "lib")
- `path` - Directory path
- `owner` - Directory owner
- `group` - Directory group
- `mode` - Directory permissions
- `recursive` - Recursive flag (boolean)

### Examples
```yaml
# Index access
sai_directory(0, 'path')      # "/etc/apache2"
sai_directory(1, 'mode')      # "0755"

# Name access
sai_directory('config', 'path')   # "/etc/apache2"
sai_directory('log', 'owner')     # "www-data"

# Wildcard access
sai_directory('*', 'path')    # "/etc/apache2 /var/log/apache2 /var/www/html"
sai_directory('*', 'name')    # "config log data"
```

---

## sai_command(selector, key)

Access command definitions from the `commands` array.

### Available Keys
- `name` - Command name
- `path` - Executable path
- `arguments` - Array of default arguments
- `aliases` - Array of command aliases
- `shell_completion` - Shell completion flag (boolean)
- `man_page` - Manual page reference

### Examples
```yaml
# Index access
sai_command(0, 'path')        # "/usr/sbin/apache2"
sai_command(0, 'name')        # "apache2"

# Name access
sai_command('apache2', 'path')        # "/usr/sbin/apache2"
sai_command('apache2ctl', 'path')     # "/usr/sbin/apache2ctl"

# Wildcard access
sai_command('*', 'path')      # "/usr/sbin/apache2 /usr/sbin/apache2ctl"
sai_command('*', 'name')      # "apache2 apache2ctl a2ensite"
```

---

## sai_port(selector, key)

Access port definitions from the `ports` array.

### Available Keys
- `port` - Port number (integer)
- `protocol` - Protocol ("tcp", "udp", "sctp")
- `service` - Service name
- `description` - Port description

### Examples
```yaml
# Index access
sai_port(0, 'port')           # "80"
sai_port(1, 'protocol')       # "tcp"

# Port number access (matches port value)
sai_port('80', 'service')     # "http"
sai_port('443', 'service')    # "https"

# Service name access (matches service field)
sai_port('http', 'port')      # "80"
sai_port('https', 'port')     # "443"

# Wildcard access
sai_port('*', 'port')         # "80 443 8080"
sai_port('*', 'service')      # "http https http-alt"
```

### Special Port Function

For backward compatibility, `sai_port()` without parameters returns the first port:

```yaml
sai_port()                    # "80" (equivalent to sai_port(0, 'port'))
```

---

## sai_container(selector, key)

Access container definitions from the `containers` array.

### Available Keys
- `name` - Container name
- `image` - Container image
- `tag` - Image tag
- `registry` - Container registry
- `platform` - Target platform
- `ports` - Array of port mappings
- `volumes` - Array of volume mappings
- `environment` - Environment variables object
- `networks` - Array of networks
- `labels` - Labels object

### Examples
```yaml
# Index access
sai_container(0, 'image')     # "httpd"
sai_container(0, 'tag')       # "2.4"

# Name access
sai_container('apache-httpd', 'image')    # "httpd"
sai_container('nginx', 'registry')        # "docker.io"

# Wildcard access
sai_container('*', 'image')   # "httpd nginx mysql"
sai_container('*', 'name')    # "apache-httpd nginx-server mysql-db"
```

---

## Error Handling

All functions follow consistent error handling:

- **Invalid selector**: Returns empty string (disables action)
- **Invalid key**: Returns empty string (disables action)
- **No matching items**: Returns empty string (disables action)
- **Multiple name matches**: Returns first match
- **Array fields**: For wildcard access on array fields (like `alternatives`, `ports`, `volumes`), returns space-separated concatenation of all array values

---

## Usage in Provider Templates

### Single Value Access
```yaml
actions:
  start:
    template: "systemctl start {{sai_service('apache', 'service_name')}}"
    
  backup_config:
    template: "cp {{sai_file('config', 'path')}} {{sai_file('config', 'path')}}.bak"
    
  check_port:
    template: "netstat -tlnp | grep :{{sai_port('http', 'port')}}"
```

### Bulk Operations
```yaml
actions:
  backup_all_configs:
    template: "tar -czf backup.tar.gz {{sai_file('*', 'path')}}"
    
  stop_all_services:
    template: "systemctl stop {{sai_service('*', 'service_name')}}"
    
  check_all_ports:
    template: "ss -tlnp | grep -E ':({{sai_port('*', 'port') | join:'|'}})')"
```

### Complex Templates
```yaml
actions:
  setup_permissions:
    template: |
      chown {{sai_file('config', 'owner')}}:{{sai_file('config', 'group')}} {{sai_file('config', 'path')}}
      chmod {{sai_file('config', 'mode')}} {{sai_file('config', 'path')}}
      
  container_run:
    template: |
      docker run -d \
        --name {{sai_container('app', 'name')}} \
        -p {{sai_port('http', 'port')}}:{{sai_port('http', 'port')}} \
        {{sai_container('app', 'registry')}}/{{sai_container('app', 'image')}}:{{sai_container('app', 'tag')}}
```

---

## Template Safety

When any SAI function returns an empty string due to missing data, the entire action template becomes invalid and is automatically disabled. This ensures:

- **Graceful degradation**: Software remains manageable through other providers
- **No partial execution**: Commands don't run with missing variables
- **Predictable behavior**: Users see consistent action availability

This template function specification provides a unified, consistent interface for accessing all saidata resources while maintaining SAI's core principles of cross-platform compatibility and graceful degradation.