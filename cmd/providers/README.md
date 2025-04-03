# Providers in SAI

This directory contains the various providers that SAI uses to interact with different systems.

## Provider Types

SAI supports these provider types:

- OS providers (`os/`): Package managers and service managers for different operating systems
- Container providers (`container/`): Handles container orchestration tools
- Cloud providers (`cloud/`): Interfaces with cloud service providers

## Dry Run Mode

All providers support a dry run mode that shows what would be executed without actually running the commands. This feature helps users preview potentially destructive actions before executing them.

### How Dry Run Mode Works

1. The `SetDryRun(enabled bool)` function in `cmd/handlers/settings.go` is called with `true` when the `--dry-run` flag is used
2. This function sets the dry run flag in the handlers package and also calls the provider-specific `SetDryRun()` functions
3. Each provider type has its own dry run mode implementation:
   - `pkgmanager.SetDryRun()`
   - `service.SetDryRun()`

### Provider Implementation

Every provider class implements an `IsDryRun()` method that checks if dry run mode is enabled. In the `Execute()` method, providers first check for dry run mode:

```go
// Example from a package manager provider
func (p *Provider) Execute(action, software string) error {
    // Check for dry run mode
    if p.IsDryRun() {
        fmt.Printf("[DRY RUN] Would execute %s %s with %s provider\n",
            action, software, p.GetPackageManager())
        return nil
    }

    // Normal execution code...
}
```

### Testing Dry Run Mode

The package includes comprehensive tests for dry run mode:

```bash
go test ./cmd/providers/os/pkgmanager -v
go test ./cmd/providers/os/service -v
```

These tests verify that all provider actions properly handle dry run mode.
