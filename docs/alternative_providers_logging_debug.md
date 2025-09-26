# Alternative Providers Logging, Debugging, and Dry-Run Support

## Overview

The alternative installation providers (source, binary, script) have comprehensive support for logging, debugging, and dry-run functionality that integrates seamlessly with SAI's existing infrastructure.

## Dry-Run Support

### Implementation
- **Global Flag**: `--dry-run` flag is available for all commands
- **Provider Integration**: All alternative providers respect the dry-run mode
- **Command Preview**: Shows exactly what commands would be executed without running them
- **Template Resolution**: Templates are resolved and displayed in dry-run mode

### Examples
```bash
# Show what would be executed for source build
sai install nginx --provider source --dry-run

# Show what would be executed for binary download  
sai install terraform --provider binary --dry-run

# Show what would be executed for script installation
sai install docker --provider script --dry-run
```

### Provider-Specific Dry-Run Output
Each provider shows detailed step-by-step commands:

**Source Provider:**
```
DRY RUN: Would execute command: mkdir -p /tmp/sai-build-nginx
DRY RUN: Would execute command: cd /tmp/sai-build-nginx && curl -L -o nginx-1.24.0.tar.gz http://nginx.org/download/nginx-1.24.0.tar.gz
DRY RUN: Would execute command: cd /tmp/sai-build-nginx && tar -xzf nginx-1.24.0.tar.gz
DRY RUN: Would execute command: cd /tmp/sai-build-nginx/nginx-1.24.0 && ./configure --prefix=/usr/local --with-http_ssl_module
DRY RUN: Would execute command: cd /tmp/sai-build-nginx/nginx-1.24.0 && make -j$(nproc)
DRY RUN: Would execute command: cd /tmp/sai-build-nginx/nginx-1.24.0 && make install
```

**Binary Provider:**
```
DRY RUN: Would execute command: mkdir -p /tmp/sai-binary-terraform
DRY RUN: Would execute command: cd /tmp/sai-binary-terraform && curl -L -o terraform.zip https://releases.hashicorp.com/terraform/1.5.0/terraform_1.5.0_linux_amd64.zip
DRY RUN: Would execute command: cd /tmp/sai-binary-terraform && unzip terraform.zip
DRY RUN: Would execute command: cp /tmp/sai-binary-terraform/terraform /usr/local/bin/
DRY RUN: Would execute command: chmod 0755 /usr/local/bin/terraform
```

**Script Provider:**
```
DRY RUN: Would execute command: mkdir -p /tmp/sai-script-docker
DRY RUN: Would execute command: cd /tmp/sai-script-docker && curl -fsSL https://get.docker.com -o install.sh
DRY RUN: Would execute command: chmod +x /tmp/sai-script-docker/install.sh
DRY RUN: Would execute command: cd /tmp && timeout 600 bash /tmp/sai-script-docker/install.sh --channel stable
```

## Logging Support

### Detailed Logging
- **Verbose Mode**: `--verbose` flag enables detailed logging for all operations
- **Step-by-Step Logging**: Each provider action step is logged with timing information
- **Error Logging**: Comprehensive error messages with context and suggestions
- **Template Resolution Logging**: Debug output for template function resolution

### Log Levels
```bash
# Standard output (default)
sai install nginx --provider source

# Verbose logging with detailed information
sai install nginx --provider source --verbose

# Quiet mode (minimal output)
sai install nginx --provider source --quiet
```

### Template Function Error Logging
The template functions provide detailed error messages:

```
sai_source error: no saidata context available
sai_binary error: requires at least 2 arguments (index, field)
sai_script error: first argument must be index (int)
sai_source error: source not found at index 1
sai_binary error: field 'invalid_field' not supported
```

## Debugging Support

### Debug Mode
- **Debug Flag**: Enable with `SAI_DEBUG=true` environment variable
- **Template Resolution**: Debug output shows how templates are resolved
- **Provider Detection**: Debug information about provider availability
- **Performance Metrics**: Timing information for all operations

### Debug Output Examples

**Provider Detection Debug:**
```
[DEBUG] Detecting provider source availability...
[DEBUG] Provider source detection result: available=true, executable=make
[DEBUG] Provider source version: GNU Make 4.3

[DEBUG] Detecting provider binary availability...
[DEBUG] Provider binary detection result: available=true, executable=curl
[DEBUG] Provider binary version: curl 7.81.0

[DEBUG] Detecting provider script availability...
[DEBUG] Provider script detection result: available=true, executable=bash
[DEBUG] Provider script version: GNU bash, version 5.1.16
```

**Template Resolution Debug:**
```
[DEBUG] Resolving template: {{sai_source(0, 'url')}}
[DEBUG] Template function: sai_source, args: [0, url]
[DEBUG] Source resolution: provider=source, index=0, field=url
[DEBUG] Resolved value: http://nginx.org/download/nginx-1.24.0.tar.gz

[DEBUG] Resolving template: {{sai_binary(0, 'download_cmd')}}
[DEBUG] Template function: sai_binary, args: [0, download_cmd]
[DEBUG] Binary resolution: provider=binary, index=0, field=download_cmd
[DEBUG] Resolved value: curl -L -o terraform.zip https://releases.hashicorp.com/terraform/1.5.0/terraform_1.5.0_linux_amd64.zip
```

## Progress Indicators

### Long-Running Operations
All alternative providers include progress indicators for time-consuming operations:

**Source Builds:**
- Download progress
- Build progress (with timeout: 3600 seconds)
- Installation progress

**Binary Downloads:**
- Download progress (with timeout: 600 seconds)
- Extraction progress
- Installation progress

**Script Execution:**
- Download progress
- Script execution progress (with configurable timeout, default: 1800 seconds)

### Progress Display
```bash
Installing nginx...
[1/6] Creating build directory...
[2/6] Downloading source code...
[3/6] Extracting archive...
[4/6] Configuring build...
[5/6] Building software...
[6/6] Installing files...
✓ nginx installed successfully
```

## Timeout Configuration

### Provider-Level Timeouts
Each provider has appropriate timeout values:

- **Source Provider**: 3600 seconds (1 hour) for complex builds
- **Binary Provider**: 600 seconds (10 minutes) for downloads
- **Script Provider**: 1800 seconds (30 minutes) default, configurable per script

### Per-Script Timeout Configuration
Scripts can specify custom timeouts in saidata:

```yaml
scripts:
  - name: "installer"
    url: "https://example.com/install.sh"
    timeout: 900  # 15 minutes for this specific script
```

### Timeout Handling
- **Graceful Termination**: Operations are terminated gracefully on timeout
- **Cleanup**: Temporary files and partial installations are cleaned up
- **Error Reporting**: Clear timeout error messages with suggestions

## Error Handling and Recovery

### Rollback Support
All providers include comprehensive rollback procedures:

**Source Provider:**
- Removes partially built files
- Cleans up build directories
- Restores previous installation on upgrade failure

**Binary Provider:**
- Removes partially downloaded files
- Restores previous binary on upgrade failure
- Cleans up temporary directories

**Script Provider:**
- Executes rollback scripts if provided
- Restores system state when possible
- Cleans up temporary files

### Error Context
Error messages include:
- **Operation Context**: What was being attempted
- **Failure Reason**: Specific cause of failure
- **Suggestions**: Recommended next steps
- **Rollback Status**: Whether cleanup was successful

## Integration with Existing SAI Features

### Consistent Interface
Alternative providers integrate seamlessly with:
- **Standard Actions**: install, uninstall, upgrade, version, info
- **Global Flags**: --provider, --verbose, --dry-run, --yes, --quiet
- **Output Formats**: JSON output support for all operations
- **Service Management**: Integration with service start/stop/status
- **File Management**: Integration with file and directory handling

### Provider Selection
```bash
# Automatic provider selection with debug info
SAI_DEBUG=true sai install nginx --verbose

# Force specific provider with dry-run
sai install nginx --provider source --dry-run --verbose

# Multiple provider comparison
sai info nginx --verbose  # Shows info from all available providers
```

## Best Practices

### Development and Testing
1. **Always Use Dry-Run First**: Test configurations with `--dry-run`
2. **Enable Verbose Logging**: Use `--verbose` for troubleshooting
3. **Check Provider Availability**: Use debug mode to verify provider detection
4. **Validate Templates**: Ensure template functions resolve correctly

### Production Deployment
1. **Set Appropriate Timeouts**: Configure timeouts based on expected operation duration
2. **Monitor Long Operations**: Use progress indicators to track build/download progress
3. **Enable Logging**: Keep detailed logs for audit and troubleshooting
4. **Test Rollback Procedures**: Verify rollback works in failure scenarios

### Troubleshooting
1. **Template Issues**: Use debug mode to see template resolution
2. **Provider Detection**: Check executable availability and platform compatibility
3. **Timeout Problems**: Adjust timeout values for slow networks or large builds
4. **Permission Issues**: Verify installation paths and file permissions

## Configuration Examples

### Comprehensive Logging Configuration
```yaml
# Enable all logging and debugging features
environment:
  SAI_DEBUG: "true"
  SAI_LOG_LEVEL: "debug"
  SAI_PROGRESS: "true"

# Provider-specific timeout overrides
providers:
  source:
    sources:
      - name: "main"
        timeout: 7200  # 2 hours for very large builds
        
  script:
    scripts:
      - name: "installer"
        timeout: 300   # 5 minutes for quick scripts
```

This comprehensive logging, debugging, and dry-run support ensures that alternative providers provide the same level of operational visibility and safety as traditional package managers.