# SAI Binary Template Function

This document defines the `sai_binary` template function used by the binary provider for downloading and installing pre-compiled binaries.

## Template Function Overview

The binary provider uses the `sai_binary(index, field)` template function to access binary download configuration from saidata files. This function follows the hierarchical resolution order described in the technical documentation.

## Function Signature

```
{{sai_binary(index, field)}}
```

- **index**: Binary index (usually 0 for the first/main binary)
- **field**: The field to retrieve from the binary configuration

## Available Fields

### Basic Information Fields

#### `url`
Returns the binary download URL with OS/architecture templating support.
- **Usage**: `{{sai_binary(0, 'url')}}`
- **Resolution Path**: `binaries[0].url` → `providers.binary.binaries[0].url`
- **Templating**: Supports `{{os}}`, `{{arch}}`, `{{version}}` placeholders
- **Example**: `"https://releases.hashicorp.com/terraform/{{version}}/terraform_{{version}}_{{os}}_{{arch}}.zip"`

#### `version`
Returns the version to download.
- **Usage**: `{{sai_binary(0, 'version')}}`
- **Resolution Path**: `binaries[0].version` → `providers.binary.binaries[0].version`
- **Example**: `"1.5.7"`

#### `architecture`
Returns the target architecture.
- **Usage**: `{{sai_binary(0, 'architecture')}}`
- **Resolution Path**: `binaries[0].architecture` → `providers.binary.binaries[0].architecture`
- **Default**: Auto-detected system architecture
- **Values**: `amd64`, `arm64`, `386`, `arm`

#### `platform`
Returns the target platform/OS.
- **Usage**: `{{sai_binary(0, 'platform')}}`
- **Resolution Path**: `binaries[0].platform` → `providers.binary.binaries[0].platform`
- **Default**: Auto-detected OS
- **Values**: `linux`, `darwin`, `windows`

### Installation Fields

#### `install_path`
Returns the installation directory path.
- **Usage**: `{{sai_binary(0, 'install_path')}}`
- **Resolution Path**: `binaries[0].install_path` → `providers.binary.binaries[0].install_path`
- **Default**: `/usr/local/bin` (Unix) or `C:\Program Files\SAI\bin` (Windows)
- **Example**: `/usr/local/bin`

#### `executable`
Returns the executable filename within the archive or download.
- **Usage**: `{{sai_binary(0, 'executable')}}`
- **Resolution Path**: `binaries[0].executable` → `providers.binary.binaries[0].executable`
- **Default**: `{{metadata.name}}` or `{{metadata.name}}.exe` (Windows)
- **Example**: `terraform`

#### `permissions`
Returns the file permissions to set on the binary.
- **Usage**: `{{sai_binary(0, 'permissions')}}`
- **Resolution Path**: `binaries[0].permissions` → `providers.binary.binaries[0].permissions`
- **Default**: `0755`
- **Example**: `0755`

### Archive Configuration Fields

#### `archive_format`
Returns the archive format for extraction.
- **Usage**: `{{sai_binary(0, 'archive_format')}}`
- **Resolution Path**: `binaries[0].archive.format` → `providers.binary.binaries[0].archive.format`
- **Auto-detected from URL extension if not specified**
- **Values**: `zip`, `tar.gz`, `tar.bz2`, `tar.xz`, `none`

#### `strip_prefix`
Returns the prefix to strip during extraction.
- **Usage**: `{{sai_binary(0, 'strip_prefix')}}`
- **Resolution Path**: `binaries[0].archive.strip_prefix` → `providers.binary.binaries[0].archive.strip_prefix`
- **Example**: `terraform_1.5.7_linux_amd64/`

#### `extract_path`
Returns the path within the archive containing the executable.
- **Usage**: `{{sai_binary(0, 'extract_path')}}`
- **Resolution Path**: `binaries[0].archive.extract_path` → `providers.binary.binaries[0].archive.extract_path`
- **Example**: `bin/terraform`

### Command Fields

#### `download_cmd`
Returns the command to download the binary.
- **Usage**: `{{sai_binary(0, 'download_cmd')}}`
- **Resolution Path**: `binaries[0].custom_commands.download` → `providers.binary.binaries[0].custom_commands.download`
- **Default**: `wget -O {{download_file}} {{sai_binary(0, 'url')}}` or `curl -L -o {{download_file}} {{sai_binary(0, 'url')}}`
- **Example**: `wget -O terraform.zip https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_linux_amd64.zip`

#### `extract_cmd`
Returns the command to extract the binary from archive.
- **Usage**: `{{sai_binary(0, 'extract_cmd')}}`
- **Resolution Path**: `binaries[0].custom_commands.extract` → `providers.binary.binaries[0].custom_commands.extract`
- **Default**: Auto-generated based on archive format
- **ZIP**: `unzip -j {{download_file}} {{sai_binary(0, 'extract_path')}} -d {{temp_dir}}`
- **TAR.GZ**: `tar -xzf {{download_file}} --strip-components={{strip_components}} -C {{temp_dir}}`

#### `install_cmd`
Returns the command to install the binary.
- **Usage**: `{{sai_binary(0, 'install_cmd')}}`
- **Resolution Path**: `binaries[0].custom_commands.install` → `providers.binary.binaries[0].custom_commands.install`
- **Default**: `install -m {{sai_binary(0, 'permissions')}} {{temp_dir}}/{{sai_binary(0, 'executable')}} {{sai_binary(0, 'install_path')}}/`
- **Example**: `install -m 0755 /tmp/terraform /usr/local/bin/`

#### `uninstall_cmd`
Returns the command to uninstall the binary.
- **Usage**: `{{sai_binary(0, 'uninstall_cmd')}}`
- **Resolution Path**: `binaries[0].custom_commands.uninstall` → `providers.binary.binaries[0].custom_commands.uninstall`
- **Default**: `rm -f {{sai_binary(0, 'install_path')}}/{{sai_binary(0, 'executable')}}`
- **Example**: `rm -f /usr/local/bin/terraform`

#### `validation_cmd`
Returns the command to validate successful installation.
- **Usage**: `{{sai_binary(0, 'validation_cmd')}}`
- **Resolution Path**: `binaries[0].custom_commands.validation` → `providers.binary.binaries[0].custom_commands.validation`
- **Default**: `which {{sai_binary(0, 'executable')}} && {{sai_binary(0, 'executable')}} --version`
- **Example**: `which terraform && terraform --version`

#### `version_cmd`
Returns the command to get installed version.
- **Usage**: `{{sai_binary(0, 'version_cmd')}}`
- **Resolution Path**: `binaries[0].custom_commands.version` → `providers.binary.binaries[0].custom_commands.version`
- **Default**: `{{sai_binary(0, 'executable')}} --version 2>&1 | head -1`
- **Example**: `terraform --version | grep -o 'Terraform v[0-9.]*'`

### Security Fields

#### `checksum`
Returns the expected checksum for binary verification.
- **Usage**: `{{sai_binary(0, 'checksum')}}`
- **Resolution Path**: `binaries[0].checksum` → `providers.binary.binaries[0].checksum`
- **Format**: `sha256:abc123...` or `md5:def456...`
- **Example**: `sha256:a1b2c3d4e5f6...`

#### `checksum_url`
Returns the URL to download checksum file.
- **Usage**: `{{sai_binary(0, 'checksum_url')}}`
- **Auto-generated**: `{{sai_binary(0, 'url')}}.sha256` if not specified
- **Example**: `https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_SHA256SUMS`

#### `verify_cmd`
Returns the command to verify binary checksum.
- **Usage**: `{{sai_binary(0, 'verify_cmd')}}`
- **Default**: `echo "{{sai_binary(0, 'checksum')}} {{download_file}}" | sha256sum -c -`
- **Example**: `echo "a1b2c3d4... terraform.zip" | sha256sum -c -`

### Utility Fields

#### `download_file`
Returns the local filename for the downloaded binary/archive.
- **Usage**: `{{sai_binary(0, 'download_file')}}`
- **Auto-generated**: Based on URL filename or `{{metadata.name}}-{{version}}.{{extension}}`
- **Example**: `terraform-1.5.7.zip`

#### `temp_dir`
Returns the temporary directory for extraction.
- **Usage**: `{{sai_binary(0, 'temp_dir')}}`
- **Default**: `/tmp/sai-binary-{{metadata.name}}-{{random}}`
- **Example**: `/tmp/sai-binary-terraform-abc123`

#### `final_path`
Returns the full path to the installed binary.
- **Usage**: `{{sai_binary(0, 'final_path')}}`
- **Auto-generated**: `{{sai_binary(0, 'install_path')}}/{{sai_binary(0, 'executable')}}`
- **Example**: `/usr/local/bin/terraform`

## Template Resolution Examples

### Basic Resolution
```yaml
# saidata file
binaries:
  - name: "main"
    url: "https://releases.hashicorp.com/terraform/{{version}}/terraform_{{version}}_{{os}}_{{arch}}.zip"
    version: "1.5.7"
    executable: "terraform"
```

Template `{{sai_binary(0, 'url')}}` on Linux amd64 resolves to: 
`"https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_linux_amd64.zip"`

### Provider Override Resolution
```yaml
# saidata file
binaries:
  - name: "main"
    url: "https://releases.hashicorp.com/terraform/{{version}}/terraform_{{version}}_{{os}}_{{arch}}.zip"
    version: "1.5.7"

providers:
  binary:
    binaries:
      - name: "main"
        url: "https://internal-mirror.com/terraform/{{version}}/terraform-{{os}}-{{arch}}.zip"
        install_path: "/opt/terraform/bin"
        checksum: "sha256:a1b2c3d4e5f6..."
```

Template `{{sai_binary(0, 'url')}}` resolves to: `"https://internal-mirror.com/terraform/1.5.7/terraform-linux-amd64.zip"`
Template `{{sai_binary(0, 'install_path')}}` resolves to: `"/opt/terraform/bin"`

### OS-Specific Resolution
```yaml
# software/te/terraform/windows/11.yaml
providers:
  binary:
    binaries:
      - name: "main"
        executable: "terraform.exe"
        install_path: "C:\\Program Files\\Terraform"
        permissions: "0755"
```

On Windows 11, `{{sai_binary(0, 'executable')}}` resolves to: `"terraform.exe"`

### Archive Configuration
```yaml
# saidata file
binaries:
  - name: "main"
    url: "https://github.com/user/app/releases/download/v{{version}}/app-{{version}}-{{os}}-{{arch}}.tar.gz"
    version: "2.1.0"
    archive:
      format: "tar.gz"
      strip_prefix: "app-2.1.0/"
      extract_path: "bin/app"
```

Template `{{sai_binary(0, 'extract_cmd')}}` resolves to:
`"tar -xzf app-2.1.0-linux-amd64.tar.gz --strip-components=1 -C /tmp/sai-binary-app-xyz123"`

## Error Handling

If a template function cannot resolve to a value:
- The action containing that template becomes unavailable
- SAI will not execute commands with unresolved templates
- Users will see an error indicating missing binary configuration

## Security Considerations

1. **Checksum Verification**: Always provide checksums for binary downloads
2. **HTTPS URLs**: Use secure download URLs to prevent man-in-the-middle attacks
3. **Permission Setting**: Set appropriate file permissions (typically 0755 for executables)
4. **Path Validation**: Ensure installation paths are secure and appropriate
5. **Archive Validation**: Verify archive contents before extraction

## Best Practices

1. **Version Templating**: Use `{{version}}` in URLs for version flexibility
2. **OS/Arch Templating**: Use `{{os}}` and `{{arch}}` for cross-platform support
3. **Checksum Verification**: Always include checksums for security
4. **Proper Permissions**: Set executable permissions appropriately
5. **Clean Installation**: Use standard installation paths
6. **Validation Commands**: Provide meaningful validation to confirm installation
7. **Archive Handling**: Configure extraction properly for different archive formats

## Integration with Existing Functions

The `sai_binary` function works alongside existing SAI template functions:
- `{{sai_package(index, field, provider)}}` - for dependency packages
- `{{sai_service(index, field, provider)}}` - for service management
- `{{sai_file(index, field, provider)}}` - for configuration files
- `{{sai_directory(index, field, provider)}}` - for directory creation
- `{{sai_container(index, field, provider)}}` - for container configurations

This enables comprehensive software management from binary installation through service operation.