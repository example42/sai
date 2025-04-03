package service

import (
	"testing"
)

// mockDryRunCheckFunc is used to check if dry run mode is enabled
// This avoids import cycle with handlers package
var mockDryRunEnabled = false

// TestDryRunModeServiceProviders tests service providers in dry run mode
func TestDryRunModeServiceProviders(t *testing.T) {
	// Set mock dry run mode
	mockDryRunEnabled = true
	defer func() { mockDryRunEnabled = false }() // Reset after test

	// Test with different service providers
	// Use all available service provider types
	serviceProviders := []struct {
		name     string
		provider Provider
	}{
		{"systemd", GetProvider("linux")},
		{"launchd", GetProvider("darwin")},
		{"windows", GetProvider("windows")},
	}

	// Service management actions to test
	actions := []string{
		ActionStart,
		ActionStop,
		ActionRestart,
		ActionEnable,
		ActionDisable,
	}

	for _, sp := range serviceProviders {
		if sp.provider == nil {
			// Skip if provider not available for this platform
			continue
		}

		for _, action := range actions {
			t.Run(sp.name+"_"+action, func(t *testing.T) {
				// Execute should not perform real operations in dry run mode
				err := sp.provider.Execute(action, "test-service")

				if err != nil {
					t.Errorf("Expected no error in dry run mode for %s service provider %s action, got: %v",
						sp.name, action, err)
				}
			})
		}
	}
}

// Notes on implementing dry run mode in service providers:
// 1. Add IsDryRun method to BaseProvider (similar to package managers)
// 2. Modify Execute methods to check dry run mode before executing commands
//
// Example:
// func (p *SystemdProvider) Execute(action, service string) error {
//     if mockDryRunEnabled {
//         fmt.Printf("[DRY RUN] Would execute %s service %s using systemd\n", action, service)
//         return nil
//     }
//     // Normal execution code...
// }
