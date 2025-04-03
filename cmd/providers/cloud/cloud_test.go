package cloud

import (
	"testing"
)

// TestCloudProviders is a placeholder test for the cloud providers package
func TestCloudProviders(t *testing.T) {
	// Empty test to ensure the package can be tested
}

// TestDryRunModeCloudProviders tests cloud providers in dry run mode
func TestDryRunModeCloudProviders(t *testing.T) {
	// Set dry run mode
	SetDryRun(true)
	defer SetDryRun(false) // Reset after test

	// Test with different cloud providers
	cloudProviders := []Provider{
		NewAWSProvider(),
		NewAzureProvider(),
		NewGCPProvider(),
	}

	// Cloud actions to test
	cloudActions := []string{
		ActionStart,
		ActionStop,
		ActionStatus,
		ActionCreate,
		ActionDelete,
		ActionList,
		ActionDescribe,
	}

	// Test resources for different providers
	testResources := map[string]string{
		"aws":   "ec2/i-1234567890abcdef0",
		"azure": "vm/my-vm",
		"gcp":   "compute/my-instance",
	}

	for _, provider := range cloudProviders {
		platform := provider.GetCloudPlatform()
		testResource := testResources[platform]

		for _, action := range cloudActions {
			t.Run(platform+"_"+action, func(t *testing.T) {
				// Execute should not perform real operations in dry run mode
				err := provider.Execute(action, testResource)

				if err != nil {
					t.Errorf("Expected no error in dry run mode for %s %s, got: %v",
						platform, action, err)
				}
			})
		}
	}
}
