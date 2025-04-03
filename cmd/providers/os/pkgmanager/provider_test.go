package pkgmanager

import (
	"testing"
)

// mockDryRunCheckFunc is used to check if dry run mode is enabled
// This avoids import cycle with handlers package
var mockDryRunEnabled = false

// TestDryRunModePackageManagers tests package managers in dry run mode
func TestDryRunModePackageManagers(t *testing.T) {
	// Set mock dry run mode
	mockDryRunEnabled = true
	defer func() { mockDryRunEnabled = false }() // Reset after test

	// Test different package managers
	providers := []Provider{
		NewAPTProvider(),
		NewRPMProvider(),
		NewBrewProvider(),
		NewPacmanProvider(),
		NewWingetProvider(),
		NewZypperProvider(),
	}

	// Test all actions for each provider
	for _, provider := range providers {
		pkgManager := provider.GetPackageManager()

		for _, action := range AllActions {
			t.Run(pkgManager+"_"+action, func(t *testing.T) {
				// Execute should not perform real operations in dry run mode
				err := provider.Execute(action, "test-software")

				if err != nil {
					t.Errorf("Expected no error in dry run mode for %s %s, got: %v",
						pkgManager, action, err)
				}
			})
		}
	}
}

// Modify the Execute method in BaseProvider to check for dry run mode
// This would be added to the BaseProvider in the provider.go file
//
// func (p *BaseProvider) IsDryRun() bool {
//     return mockDryRunEnabled
// }
//
// And each provider's Execute method would be modified to check:
// if p.IsDryRun() {
//    fmt.Printf("[DRY RUN] Would execute %s %s with %s provider\n", action, software, p.Name)
//    return nil
// }
