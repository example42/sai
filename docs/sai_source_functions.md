# SAI Source Template Function

This document defines the `sai_source` template function used by the source provider for building software from source code.

## Template Function Overview

The source provider uses the `sai_source(index, field)` template function to access source build configuration from saidata files. This function follows the hierarchical resolution order described in the technical documentation.

## Function Signature

```
{{sai_source(index, field)}}
```

- **index**: Source index (usually 0 for the first/main source)
- **field**: The field to retrieve from the source configuration

## Available Fields

### Basic Information Fields

#### `url`
Returns the source code download URL with version templating support.
- **Usage**: `{{sai_source(0, 'url')}}`
- **Resolution Path**: `sources[0].url` → `providers.source.sources[0].url`
- **Example**: `"http://nginx.org/download/nginx-{{version}}.tar.gz"`

#### `version`
Returns the version to build.
- **Usage**: `{{sai_source(0, 'version')}}`
- **Resolution Path**: `sources[0].version` → `providers.source.sources[0].version`
- **Example**: `"1.24.0"`

#### `build_system`
Returns the build system type.
- **Usage**: `{{sai_source(0, 'build_system')}}`
- **Resolution Path**: `sources[0].build_system` → `providers.source.sources[0].build_system`
- **Values**: `autotools`, `cmake`, `make`, `meson`, `ninja`, `custom`

### Directory Fields

#### `build_dir`
Returns the build directory path.
- **Usage**: `{{sai_source(0, 'build_dir')}}`
- **Resolution Path**: `sources[0].build_dir` → `providers.source.sources[0].build_dir`
- **Default**: `/tmp/sai-build-{{metadata.name}}`
- **Example**: `/tmp/sai-build-nginx`

#### `source_dir`
Returns the source directory path (where extracted source code resides).
- **Usage**: `{{sai_source(0, 'source_dir')}}`
- **Resolution Path**: `sources[0].source_dir` → `providers.source.sources[0].source_dir`
- **Default**: `{{sai_source(0, 'build_dir')}}/{{metadata.name}}-{{sai_source(0, 'version')}}`
- **Example**: `/tmp/sai-build-nginx/nginx-1.24.0`

#### `install_prefix`
Returns the installation prefix.
- **Usage**: `{{sai_source(0, 'install_prefix')}}`
- **Resolution Path**: `sources[0].install_prefix` → `providers.source.sources[0].install_prefix`
- **Default**: `/usr/local`

### Command Fields

#### `download_cmd`
Returns the command to download source code.
- **Usage**: `{{sai_source(0, 'download_cmd')}}`
- **Resolution Path**: `sources[0].custom_commands.download` → `providers.source.sources[0].custom_commands.download`
- **Default (autotools/cmake/make)**: `wget -O {{archive_name}} {{sai_source(0, 'url')}}`
- **Example**: `wget -O nginx-1.24.0.tar.gz http://nginx.org/download/nginx-1.24.0.tar.gz`

#### `extract_cmd`
Returns the command to extract source archive.
- **Usage**: `{{sai_source(0, 'extract_cmd')}}`
- **Resolution Path**: `sources[0].custom_commands.extract` → `providers.source.sources[0].custom_commands.extract`
- **Default**: Auto-detected based on file extension (tar.gz, tar.bz2, zip, etc.)
- **Example**: `tar -xzf nginx-1.24.0.tar.gz`

#### `configure_cmd`
Returns the configure command.
- **Usage**: `{{sai_source(0, 'configure_cmd')}}`
- **Resolution Path**: `sources[0].custom_commands.configure` → `providers.source.sources[0].custom_commands.configure`
- **Default (autotools)**: `./configure --prefix={{sai_source(0, 'install_prefix')}} {{configure_args}}`
- **Default (cmake)**: `cmake . -DCMAKE_INSTALL_PREFIX={{sai_source(0, 'install_prefix')}} {{configure_args}}`
- **Example**: `./configure --prefix=/usr/local --with-http_ssl_module --with-http_v2_module`

#### `build_cmd`
Returns the build command.
- **Usage**: `{{sai_source(0, 'build_cmd')}}`
- **Resolution Path**: `sources[0].custom_commands.build` → `providers.source.sources[0].custom_commands.build`
- **Default**: `make -j$(nproc) {{build_args}}`
- **Example**: `make -j$(nproc)`

#### `install_cmd`
Returns the install command.
- **Usage**: `{{sai_source(0, 'install_cmd')}}`
- **Resolution Path**: `sources[0].custom_commands.install` → `providers.source.sources[0].custom_commands.install`
- **Default**: `make install {{install_args}}`
- **Example**: `make install`

#### `uninstall_cmd`
Returns the uninstall command.
- **Usage**: `{{sai_source(0, 'uninstall_cmd')}}`
- **Resolution Path**: `sources[0].custom_commands.uninstall` → `providers.source.sources[0].custom_commands.uninstall`
- **Default**: `make uninstall` (if supported) or custom removal script
- **Example**: `rm -rf /usr/local/sbin/nginx /usr/local/conf/nginx*`

#### `validation_cmd`
Returns the command to validate successful installation.
- **Usage**: `{{sai_source(0, 'validation_cmd')}}`
- **Resolution Path**: `sources[0].custom_commands.validation` → `providers.source.sources[0].custom_commands.validation`
- **Default**: `which {{metadata.name}} && {{metadata.name}} --version`
- **Example**: `/usr/local/sbin/nginx -t`

#### `version_cmd`
Returns the command to get installed version.
- **Usage**: `{{sai_source(0, 'version_cmd')}}`
- **Resolution Path**: `sources[0].custom_commands.version` → `providers.source.sources[0].custom_commands.version`
- **Default**: `{{metadata.name}} --version 2>&1 | head -1`
- **Example**: `/usr/local/sbin/nginx -v 2>&1 | grep -o 'nginx/[0-9.]*'`

### Prerequisites Fields

#### `prerequisites`
Returns space-separated list of prerequisite packages.
- **Usage**: `{{sai_source(0, 'prerequisites')}}`
- **Resolution Path**: `sources[0].prerequisites` → `providers.source.sources[0].prerequisites`
- **Example**: `"build-essential libssl-dev libpcre3-dev zlib1g-dev"`

#### `prerequisites_install_cmd`
Returns the command to install prerequisites.
- **Usage**: `{{sai_source(0, 'prerequisites_install_cmd')}}`
- **Auto-generated based on detected package manager**
- **Ubuntu/Debian**: `apt-get update && apt-get install -y {{sai_source(0, 'prerequisites')}}`
- **CentOS/RHEL**: `yum install -y {{sai_source(0, 'prerequisites')}}`
- **macOS**: `brew install {{sai_source(0, 'prerequisites')}}`

### Utility Fields

#### `manifest_file`
Returns the path to the installation manifest file.
- **Usage**: `{{sai_source(0, 'manifest_file')}}`
- **Default**: `/var/lib/sai/manifests/{{metadata.name}}-source.manifest`
- **Purpose**: Tracks source installation for uninstall operations

#### `checksum`
Returns the expected checksum for source verification.
- **Usage**: `{{sai_source(0, 'checksum')}}`
- **Resolution Path**: `sources[0].checksum` → `providers.source.sources[0].checksum`
- **Format**: `sha256:abc123...` or `md5:def456...`

#### `environment`
Returns environment variables for build process.
- **Usage**: `{{sai_source(0, 'environment')}}`
- **Resolution Path**: `sources[0].environment` → `providers.source.sources[0].environment`
- **Format**: Space-separated KEY=value pairs
- **Example**: `CC=gcc-9 CXX=g++-9 CFLAGS=-O2`

## Template Resolution Examples

### Basic Resolution
```yaml
# saidata file
sources:
  - name: "main"
    url: "https://example.com/app-{{version}}.tar.gz"
    version: "2.1.0"
    build_system: "autotools"
```

Template `{{sai_source(0, 'url')}}` resolves to: `"https://example.com/app-2.1.0.tar.gz"`

### Provider Override Resolution
```yaml
# saidata file
sources:
  - name: "main"
    url: "https://example.com/app-{{version}}.tar.gz"
    version: "2.1.0"

providers:
  source:
    sources:
      - name: "main"
        url: "https://custom-mirror.com/app-{{version}}.tar.gz"
        custom_commands:
          configure: "./configure --prefix=/opt/myapp"
```

Template `{{sai_source(0, 'url')}}` resolves to: `"https://custom-mirror.com/app-2.1.0.tar.gz"`
Template `{{sai_source(0, 'configure_cmd')}}` resolves to: `"./configure --prefix=/opt/myapp"`

### OS-Specific Resolution
```yaml
# software/ng/nginx/ubuntu/22.04.yaml
providers:
  source:
    sources:
      - name: "main"
        prerequisites:
          - "build-essential"
          - "libssl-dev"
          - "libpcre3-dev"
        custom_commands:
          configure: "./configure --prefix=/usr/local --with-http_ssl_module"
```

On Ubuntu 22.04, `{{sai_source(0, 'prerequisites')}}` resolves to: `"build-essential libssl-dev libpcre3-dev"`

## Error Handling

If a template function cannot resolve to a value:
- The action containing that template becomes unavailable
- SAI will not execute commands with unresolved templates
- Users will see an error indicating missing source configuration

## Best Practices

1. **Always provide fallbacks**: Define both base and provider-specific configurations
2. **Use version templating**: Make URLs version-aware with `{{version}}` placeholder
3. **Specify prerequisites**: Include all build dependencies for reliable builds
4. **Custom validation**: Provide meaningful validation commands
5. **Proper cleanup**: Define uninstall commands for complete removal
6. **OS-specific overrides**: Use OS-specific files for platform differences

## Integration with Existing Functions

The `sai_source` function works alongside existing SAI template functions:
- `{{sai_package(index, field, provider)}}` - for prerequisite package names
- `{{sai_service(index, field, provider)}}` - for service management after build
- `{{sai_file(index, field, provider)}}` - for configuration file paths
- `{{sai_directory(index, field, provider)}}` - for directory creation
- `{{sai_container(index, field, provider)}}` - for container configurations

This enables comprehensive software management from source build through service operation.