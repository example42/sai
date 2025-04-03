package container

import (
	"testing"
)

// TestContainerProviders is a placeholder test for the container providers package
func TestContainerProviders(t *testing.T) {
	// Empty test to ensure the package can be tested
}

// TestDryRunModeContainerProviders tests container providers in dry run mode
func TestDryRunModeContainerProviders(t *testing.T) {
	// Set dry run mode
	SetDryRun(true)
	defer SetDryRun(false) // Reset after test

	// Test with different container providers
	containerProviders := []Provider{
		NewKubectlProvider(),
		NewHelmProvider(),
	}

	// Container actions to test (subset of all actions)
	containerActions := []string{
		ActionStart,
		ActionStop,
		ActionRestart,
		ActionInstall,
		ActionUninstall,
		ActionStatus,
		ActionList,
		ActionLogs,
	}

	// Test resources for different providers
	testResources := map[string]string{
		"kubectl": "deployment/nginx",
		"helm":    "nginx",
	}

	for _, provider := range containerProviders {
		tool := provider.GetContainerTool()
		testResource := testResources[tool]

		for _, action := range containerActions {
			t.Run(tool+"_"+action, func(t *testing.T) {
				// Execute should not perform real operations in dry run mode
				err := provider.Execute(action, testResource)

				if err != nil {
					t.Errorf("Expected no error in dry run mode for %s %s, got: %v",
						tool, action, err)
				}
			})
		}
	}
}
