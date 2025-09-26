# SaiData Schema Reference

## Overview

The SaiData schema (version 0.2) defines the structure for software configuration files in SAI. This document provides comprehensive reference for all fields, with special focus on the alternative installation providers: sources, binaries, and scripts.

## Schema Structure

### Root Level Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `version` | string | ✓ | Schema version (e.g., "0.2") |
| `metadata` | object | ✓ | Software metadata and information |
| `packages` | array | | Default package definitions |
| `services` | array | | Default service definitions |
| `files` | array | | Default file definitions |
| `directories` | array | | Default directory definitions |
| `commands` | array | | Default command definitions |
| `ports` | array | | Default port definitions |
| `containers` | array | | Default container definitions |
| `sources` | array | | **Default source build definitions** |
| `binaries` | array | | **Default binary download definitions** |
| `scripts` | array | | **Default script installation definitions** |
| `providers` | object | | Provider-specific configurations |
| `compatibility` | object | | Compatibility matrix |

## Alternative Installation Providers

### Sources

Source builds allow compiling software from source code with various build systems.

#### Source Object Properties

| Property | Type | Required | Description | Examples |
|----------|------|----------|-------------|----------|
| `name` | string | ✓ | Logical name for the source build | `"main"`, `"stable"`, `"dev"` |
| `url` | string | ✓ | Source download URL (supports templating) | `"https://nginx.org/download/nginx-{{version}}.tar.gz"` |
| `build_system` | string | ✓ | Build system type | `"autotools"`, `"cmake"`, `"make"`, `"meson"`, `"ninja"`, `"custom"` |
| `version` | string | | Version to build | `"1.24.0"`, `"latest"` |
| `build_dir` | string | | Build directory | `"/tmp/sai-build-nginx"` |
| `source_dir` | string | | Source code directory | `"/tmp/sai-build-nginx/nginx-1.24.0"` |
| `install_prefix` | string | | Installation prefix | `"/usr/local"`, `"/opt/software"` |
| `configure_args` | array | | Configure step arguments | `["--with-http_ssl_module", "--enable-shared"]` |
| `build_args` | array | | Build step arguments | `["-j4", "VERBOSE=1"]` |
| `install_args` | array | | Install step arguments | `["DESTDIR=/tmp/staging"]` |
| `prerequisites` | array | | Required packages for building | `["build-essential", "libssl-dev"]` |
| `environment` | object | | Environment variables | `{"CC": "gcc", "CFLAGS": "-O2"}` |
| `checksum` | string | | Expected checksum (format: `algorithm:hash`) | `"sha256:abc123..."` |
| `custom_commands` | object | | Custom command overrides | See [Custom Commands](#custom-commands) |

#### Build System Types

- **autotools**: Traditional `./configure && make && make install`
- **cmake**: CMake-based builds with `cmake . && cmake --build .`
- **make**: Direct Makefile-based builds
- **meson**: Meson build system with `meson setup build`
- **ninja**: Ninja build files
- **custom**: User-defined build commands

### Binaries

Binary downloads allow installing pre-compiled executables with OS/architecture templating.

#### Binary Object Properties

| Property | Type | Required | Description | Examples |
|----------|------|----------|-------------|----------|
| `name` | string | ✓ | Logical name for the binary | `"main"`, `"stable"`, `"lts"` |
| `url` | string | ✓ | Download URL (supports templating) | `"https://releases.example.com/{{version}}/app_{{platform}}_{{architecture}}.zip"` |
| `version` | string | | Version to download | `"1.5.0"`, `"latest"` |
| `architecture` | string | | Target architecture (auto-detected) | `"amd64"`, `"arm64"`, `"386"` |
| `platform` | string | | Target platform (auto-detected) | `"linux"`, `"darwin"`, `"windows"` |
| `checksum` | string | | Expected checksum | `"sha256:fa16d72a..."` |
| `install_path` | string | | Installation directory | `"/usr/local/bin"`, `"/opt/bin"` |
| `executable` | string | | Executable name | `"terraform"`, `"kubectl"` |
| `archive` | object | | Archive extraction config | See [Archive Config](#archive-config) |
| `permissions` | string | | File permissions (octal) | `"0755"`, `"0644"` |
| `custom_commands` | object | | Custom command overrides | See [Custom Commands](#custom-commands) |

#### URL Templating

Binary URLs support the following placeholders:
- `{{version}}`: Software version
- `{{platform}}`: OS platform (`linux`, `darwin`, `windows`)
- `{{architecture}}`: CPU architecture (`amd64`, `arm64`, `386`)

#### Archive Config

| Property | Type | Description | Examples |
|----------|------|-------------|----------|
| `format` | string | Archive format (auto-detected) | `"zip"`, `"tar.gz"`, `"none"` |
| `strip_prefix` | string | Directory prefix to strip | `"terraform_1.5.0_linux_amd64/"` |
| `extract_path` | string | Path within archive to extract | `"bin/"`, `"dist/"` |

### Scripts

Script installations allow executing installation scripts with security measures.

#### Script Object Properties

| Property | Type | Required | Description | Examples |
|----------|------|----------|-------------|----------|
| `name` | string | ✓ | Logical name for the script | `"official"`, `"convenience"`, `"installer"` |
| `url` | string | ✓ | Script download URL (HTTPS recommended) | `"https://get.docker.com"`, `"https://sh.rustup.rs"` |
| `version` | string | | Version identifier | `"latest"`, `"v1.0.0"` |
| `interpreter` | string | | Script interpreter (auto-detected) | `"bash"`, `"sh"`, `"python3"` |
| `checksum` | string | | Expected checksum (required for security) | `"sha256:b5b2b2c5..."` |
| `arguments` | array | | Script arguments | `["--channel", "stable"]`, `["--yes", "--quiet"]` |
| `environment` | object | | Environment variables | `{"CHANNEL": "stable"}` |
| `working_dir` | string | | Working directory | `"/tmp"`, `"~/Downloads"` |
| `timeout` | integer | | Execution timeout (seconds, 1-3600) | `300`, `600` |
| `custom_commands` | object | | Custom command overrides | See [Custom Commands](#custom-commands) |

## Custom Commands

All three installation types support custom command overrides:

### Source Custom Commands

| Command | Description | Example |
|---------|-------------|---------|
| `download` | Custom download command | `"git clone https://github.com/user/repo.git"` |
| `extract` | Custom extract command | `"tar -xzf source.tar.gz"` |
| `configure` | Custom configure command | `"./configure --prefix=/usr/local --enable-ssl"` |
| `build` | Custom build command | `"make -j$(nproc)"` |
| `install` | Custom install command | `"make install"` |
| `uninstall` | Custom uninstall command | `"make uninstall"` |
| `validation` | Installation validation | `"nginx -t"` |
| `version` | Version detection | `"nginx -v 2>&1 \| grep -o 'nginx/[0-9.]*'"` |

### Binary Custom Commands

| Command | Description | Example |
|---------|-------------|---------|
| `download` | Custom download command | `"curl -L -o binary.zip {{url}}"` |
| `extract` | Custom extract command | `"unzip -q binary.zip"` |
| `install` | Custom install command | `"mv binary /usr/local/bin/ && chmod +x /usr/local/bin/binary"` |
| `uninstall` | Custom uninstall command | `"rm -f /usr/local/bin/binary"` |
| `validation` | Installation validation | `"binary --version"` |
| `version` | Version detection | `"binary --version \| cut -d' ' -f2"` |

### Script Custom Commands

| Command | Description | Example |
|---------|-------------|---------|
| `download` | Custom download command | `"curl -fsSL {{url}} -o install.sh"` |
| `install` | Custom install command (overrides script execution) | `"bash install.sh --yes --quiet"` |
| `uninstall` | Custom uninstall command | `"bash uninstall.sh"` |
| `validation` | Installation validation | `"software --version"` |
| `version` | Version detection | `"software --version \| cut -d' ' -f2"` |

## Provider Overrides

All alternative installation types can be overridden in provider-specific configurations:

```yaml
providers:
  source:
    sources:
      - name: "main"
        url: "https://provider-specific-url.com/source.tar.gz"
        build_system: "cmake"
        # ... other overrides
  
  binary:
    binaries:
      - name: "main"
        url: "https://provider-specific-binary.com/{{version}}/app.zip"
        # ... other overrides
  
  script:
    scripts:
      - name: "official"
        url: "https://provider-specific-script.com/install.sh"
        # ... other overrides
```

## Validation Rules

### Required Fields
- **Sources**: `name`, `url`, `build_system`
- **Binaries**: `name`, `url`
- **Scripts**: `name`, `url`

### Checksum Format
All checksums must follow the pattern: `algorithm:hash`
- Supported algorithms: `sha256`, `sha512`, `md5`
- Hash length: 32-128 hexadecimal characters
- Example: `sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7`

### URL Requirements
- **Scripts**: HTTPS URLs strongly recommended for security
- **All types**: Support templating with `{{version}}`, `{{platform}}`, `{{architecture}}`

### Timeout Limits
- **Scripts**: 1-3600 seconds (1 second to 1 hour)
- Default: 300 seconds (5 minutes)

## Examples

### Complete Source Configuration
```yaml
sources:
  - name: "main"
    url: "http://nginx.org/download/nginx-{{version}}.tar.gz"
    version: "1.24.0"
    build_system: "autotools"
    configure_args:
      - "--with-http_ssl_module"
      - "--with-http_v2_module"
    prerequisites:
      - "build-essential"
      - "libssl-dev"
    checksum: "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
    custom_commands:
      validation: "nginx -t && nginx -v"
```

### Complete Binary Configuration
```yaml
binaries:
  - name: "main"
    url: "https://releases.hashicorp.com/terraform/{{version}}/terraform_{{version}}_{{platform}}_{{architecture}}.zip"
    version: "1.5.0"
    checksum: "sha256:fa16d72a078210a54c47dd5bef2f8b9b8a01d94909a51453956b3ec6442ea4c5"
    install_path: "/usr/local/bin"
    executable: "terraform"
    archive:
      format: "zip"
    permissions: "0755"
```

### Complete Script Configuration
```yaml
scripts:
  - name: "convenience"
    url: "https://get.docker.com"
    checksum: "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
    interpreter: "bash"
    arguments: ["--channel", "stable"]
    environment:
      CHANNEL: "stable"
      DOWNLOAD_URL: "https://download.docker.com"
    timeout: 600
```

## Migration Guide

### From Previous Versions

If you have existing saidata files without alternative installation providers:

1. **Add new sections**: Add `sources`, `binaries`, or `scripts` arrays as needed
2. **Provider overrides**: Move provider-specific configurations to the appropriate provider sections
3. **Validate**: Use `ajv validate -s schemas/saidata-0.2-schema.json -d your-file.yaml`

### Best Practices

1. **Security**: Always use HTTPS URLs and checksums for scripts and binaries
2. **Templating**: Use version templating for maintainable configurations
3. **Prerequisites**: List all build dependencies for source builds
4. **Validation**: Include validation commands to verify successful installations
5. **Documentation**: Add clear descriptions and examples in your saidata files