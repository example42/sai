# SAI Script Template Function

This document defines the `sai_script` template function used by the script provider for executing installation scripts.

## Template Function Overview

The script provider uses the `sai_script(index, field)` template function to access script execution configuration from saidata files. This function follows the hierarchical resolution order described in the technical documentation.

## Function Signature

```
{{sai_script(index, field)}}
```

- **index**: Script index (usually 0 for the first/main script)
- **field**: The field to retrieve from the script configuration

## Available Fields

### Basic Information Fields

#### `url`
Returns the script download URL.
- **Usage**: `{{sai_script(0, 'url')}}`
- **Resolution Path**: `scripts[0].url` → `providers.script.scripts[0].url`
- **Security**: Must be HTTPS for security
- **Example**: `"https://get.docker.com/"`

#### `version`
Returns the script version or software version to install.
- **Usage**: `{{sai_script(0, 'version')}}`
- **Resolution Path**: `scripts[0].version` → `providers.script.scripts[0].version`
- **Example**: `"latest"` or `"20.10.21"`

#### `interpreter`
Returns the script interpreter to use.
- **Usage**: `{{sai_script(0, 'interpreter')}}`
- **Resolution Path**: `scripts[0].interpreter` → `providers.script.scripts[0].interpreter`
- **Default**: Auto-detected from script shebang or file extension
- **Values**: `bash`, `sh`, `python`, `python3`, `powershell`
- **Example**: `bash`

### Execution Configuration Fields

#### `arguments`
Returns space-separated script arguments.
- **Usage**: `{{sai_script(0, 'arguments')}}`
- **Resolution Path**: `scripts[0].arguments` → `providers.script.scripts[0].arguments`
- **Format**: Space-separated string of arguments
- **Example**: `"--version stable --user myuser"`

#### `working_dir`
Returns the working directory for script execution.
- **Usage**: `{{sai_script(0, 'working_dir')}}`
- **Resolution Path**: `scripts[0].working_dir` → `providers.script.scripts[0].working_dir`
- **Default**: `/tmp/sai-script-{{metadata.name}}`
- **Example**: `/tmp/sai-script-docker`

#### `timeout`
Returns the execution timeout in seconds.
- **Usage**: `{{sai_script(0, 'timeout')}}`
- **Resolution Path**: `scripts[0].timeout` → `providers.script.scripts[0].timeout`
- **Default**: `300` (5 minutes)
- **Example**: `600`

### Environment Configuration Fields

#### `environment`
Returns environment variables for script execution.
- **Usage**: `{{sai_script(0, 'environment')}}`
- **Resolution Path**: `scripts[0].environment` → `providers.script.scripts[0].environment`
- **Format**: Space-separated KEY=value pairs
- **Example**: `"DEBIAN_FRONTEND=noninteractive DOCKER_VERSION=20.10.21"`

#### `environment_file`
Returns path to environment file to source before execution.
- **Usage**: `{{sai_script(0, 'environment_file')}}`
- **Resolution Path**: `scripts[0].environment_file` → `providers.script.scripts[0].environment_file`
- **Example**: `/etc/sai/docker-env`

### Command Fields

#### `download_cmd`
Returns the command to download the script.
- **Usage**: `{{sai_script(0, 'download_cmd')}}`
- **Resolution Path**: `scripts[0].custom_commands.download` → `providers.script.scripts[0].custom_commands.download`
- **Default**: `curl -fsSL {{sai_script(0, 'url')}} -o {{script_file}}`
- **Example**: `curl -fsSL https://get.docker.com/ -o /tmp/docker-install.sh`

#### `install_cmd`
Returns the command to execute the installation script.
- **Usage**: `{{sai_script(0, 'install_cmd')}}`
- **Resolution Path**: `scripts[0].custom_commands.install` → `providers.script.scripts[0].custom_commands.install`
- **Default**: `{{sai_script(0, 'interpreter')}} {{script_file}} {{sai_script(0, 'arguments')}}`
- **Example**: `bash /tmp/docker-install.sh --version stable`

#### `uninstall_cmd`
Returns the command to uninstall or cleanup.
- **Usage**: `{{sai_script(0, 'uninstall_cmd')}}`
- **Resolution Path**: `scripts[0].custom_commands.uninstall` → `providers.script.scripts[0].custom_commands.uninstall`
- **Example**: `apt-get remove -y docker-ce docker-ce-cli containerd.io`

#### `validation_cmd`
Returns the command to validate successful installation.
- **Usage**: `{{sai_script(0, 'validation_cmd')}}`
- **Resolution Path**: `scripts[0].custom_commands.validation` → `providers.script.scripts[0].custom_commands.validation`
- **Default**: `which {{metadata.name}} && {{metadata.name}} --version`
- **Example**: `docker --version && systemctl is-active docker`

#### `version_cmd`
Returns the command to get installed version.
- **Usage**: `{{sai_script(0, 'version_cmd')}}`
- **Resolution Path**: `scripts[0].custom_commands.version` → `providers.script.scripts[0].custom_commands.version`
- **Default**: `{{metadata.name}} --version 2>&1 | head -1`
- **Example**: `docker --version | grep -o 'Docker version [0-9.]*'`

### Security Fields

#### `checksum`
Returns the expected checksum for script verification.
- **Usage**: `{{sai_script(0, 'checksum')}}`
- **Resolution Path**: `scripts[0].checksum` → `providers.script.scripts[0].checksum`
- **Format**: `sha256:abc123...` or `md5:def456...`
- **Example**: `sha256:a1b2c3d4e5f6...`

#### `verify_cmd`
Returns the command to verify script checksum.
- **Usage**: `{{sai_script(0, 'verify_cmd')}}`
- **Default**: `echo "{{sai_script(0, 'checksum')}} {{script_file}}" | sha256sum -c -`
- **Example**: `echo "a1b2c3d4... /tmp/docker-install.sh" | sha256sum -c -`

#### `signature_url`
Returns the URL to download script signature for verification.
- **Usage**: `{{sai_script(0, 'signature_url')}}`
- **Resolution Path**: `scripts[0].signature_url` → `providers.script.scripts[0].signature_url`
- **Example**: `"https://get.docker.com/gpg"`

### Utility Fields

#### `script_file`
Returns the local path to the downloaded script.
- **Usage**: `{{sai_script(0, 'script_file')}}`
- **Auto-generated**: `{{sai_script(0, 'working_dir')}}/{{metadata.name}}-install.sh`
- **Example**: `/tmp/sai-script-docker/docker-install.sh`

#### `log_file`
Returns the path to the execution log file.
- **Usage**: `{{sai_script(0, 'log_file')}}`
- **Default**: `/var/log/sai/{{metadata.name}}-script.log`
- **Example**: `/var/log/sai/docker-script.log`

#### `pid_file`
Returns the path to store the script process ID.
- **Usage**: `{{sai_script(0, 'pid_file')}}`
- **Default**: `/var/run/sai/{{metadata.name}}-script.pid`
- **Example**: `/var/run/sai/docker-script.pid`

### Interactive Handling Fields

#### `auto_confirm`
Returns whether to automatically confirm interactive prompts.
- **Usage**: `{{sai_script(0, 'auto_confirm')}}`
- **Resolution Path**: `scripts[0].auto_confirm` → `providers.script.scripts[0].auto_confirm`
- **Default**: `false` (requires explicit user consent)
- **Values**: `true`, `false`

#### `confirm_responses`
Returns predefined responses for interactive prompts.
- **Usage**: `{{sai_script(0, 'confirm_responses')}}`
- **Resolution Path**: `scripts[0].confirm_responses` → `providers.script.scripts[0].confirm_responses`
- **Format**: Newline-separated responses
- **Example**: `"y\nyes\n/usr/local\n"`

#### `expect_script`
Returns an expect script for handling complex interactions.
- **Usage**: `{{sai_script(0, 'expect_script')}}`
- **Resolution Path**: `scripts[0].expect_script` → `providers.script.scripts[0].expect_script`
- **Example**: Path to expect script file for complex installations

## Template Resolution Examples

### Basic Resolution
```yaml
# saidata file
scripts:
  - name: "main"
    url: "https://get.docker.com/"
    interpreter: "bash"
    arguments: "--version stable"
```

Template `{{sai_script(0, 'install_cmd')}}` resolves to: 
`"bash /tmp/sai-script-docker/docker-install.sh --version stable"`

### Provider Override Resolution
```yaml
# saidata file
scripts:
  - name: "main"
    url: "https://get.docker.com/"
    interpreter: "bash"

providers:
  script:
    scripts:
      - name: "main"
        url: "https://internal-mirror.com/docker-install.sh"
        checksum: "sha256:a1b2c3d4e5f6..."
        environment: "DEBIAN_FRONTEND=noninteractive"
        timeout: 600
```

Template `{{sai_script(0, 'url')}}` resolves to: `"https://internal-mirror.com/docker-install.sh"`
Template `{{sai_script(0, 'timeout')}}` resolves to: `"600"`

### OS-Specific Resolution
```yaml
# software/do/docker/ubuntu/22.04.yaml
providers:
  script:
    scripts:
      - name: "main"
        arguments: "--version stable --channel stable"
        environment: "DEBIAN_FRONTEND=noninteractive APT_KEY_DONT_WARN_ON_DANGEROUS_USAGE=1"
        custom_commands:
          validation: "docker --version && systemctl is-active docker"
```

On Ubuntu 22.04, `{{sai_script(0, 'environment')}}` resolves to: 
`"DEBIAN_FRONTEND=noninteractive APT_KEY_DONT_WARN_ON_DANGEROUS_USAGE=1"`

### Interactive Script Handling
```yaml
# saidata file
scripts:
  - name: "main"
    url: "https://example.com/interactive-installer.sh"
    auto_confirm: true
    confirm_responses: |
      y
      /opt/myapp
      yes
    timeout: 900
```

Template `{{sai_script(0, 'confirm_responses')}}` resolves to predefined responses for automation.

## Error Handling

If a template function cannot resolve to a value:
- The action containing that template becomes unavailable
- SAI will not execute commands with unresolved templates
- Users will see an error indicating missing script configuration

## Security Considerations

1. **HTTPS Only**: Script URLs must use HTTPS to prevent tampering
2. **Checksum Verification**: Always verify script integrity before execution
3. **Signature Verification**: Use GPG signatures when available
4. **Sandboxing**: Consider running scripts in isolated environments
5. **User Consent**: Require explicit user approval for script execution
6. **Audit Logging**: Log all script executions for security auditing
7. **Environment Isolation**: Limit environment variable exposure
8. **Timeout Enforcement**: Prevent runaway script execution

## Best Practices

1. **Security First**: Always verify script integrity with checksums
2. **Explicit Consent**: Never run scripts without user awareness
3. **Timeout Configuration**: Set reasonable execution timeouts
4. **Environment Control**: Carefully manage environment variables
5. **Error Handling**: Provide meaningful error messages and rollback
6. **Logging**: Maintain detailed execution logs
7. **Interactive Handling**: Automate interactive prompts safely
8. **Version Pinning**: Use specific script versions when possible

## Interactive Script Automation

SAI provides several mechanisms for handling interactive scripts:

### Automatic Confirmation
```yaml
scripts:
  - name: "main"
    auto_confirm: true  # Automatically answer 'yes' to prompts
```

### Predefined Responses
```yaml
scripts:
  - name: "main"
    confirm_responses: |
      y
      /usr/local
      stable
```

### Expect Scripts
```yaml
scripts:
  - name: "main"
    expect_script: "/etc/sai/scripts/docker-expect.exp"
```

## Integration with Existing Functions

The `sai_script` function works alongside existing SAI template functions:
- `{{sai_package(index, field, provider)}}` - for dependency packages
- `{{sai_service(index, field, provider)}}` - for service management after installation
- `{{sai_file(index, field, provider)}}` - for configuration files
- `{{sai_directory(index, field, provider)}}` - for directory creation
- `{{sai_container(index, field, provider)}}` - for container configurations

This enables comprehensive software management from script installation through service operation.