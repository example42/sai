package handlers

import (
	"sai/cmd/providers/cloud"
	"sai/cmd/providers/container"
	"sai/cmd/providers/os/pkgmanager"
	"sai/cmd/providers/os/service"
)

// Global settings for handlers
var (
	dryRunMode bool
)

// SetDryRun sets the dry run mode for all handlers and providers
func SetDryRun(enabled bool) {
	dryRunMode = enabled

	// Also set dry run mode for providers
	pkgmanager.SetDryRun(enabled)
	service.SetDryRun(enabled)
	cloud.SetDryRun(enabled)
	container.SetDryRun(enabled)
}

// IsDryRun returns whether dry run mode is enabled
func IsDryRun() bool {
	return dryRunMode
}
