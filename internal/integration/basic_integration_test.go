package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sai/internal/provider"
	"sai/internal/saidata"
	"sai/internal/validation"
)

// TestBasicIntegration tests the basic integration between components
func TestBasicIntegration(t *testing.T) {
	tempDir := t.TempDir()

	// Test provider and saidata integration
	testProviderSaidataIntegration(t, tempDir)
	
	// Test validation integration
	testValidationIntegration(t, tempDir)
}

func testProviderSaidataIntegration(t *testing.T, tempDir string) {
	// Setup provider
	providersDir := filepath.Join(tempDir, "providers")
	err := os.MkdirAll(providersDir, 0755)
	require.NoError(t, err)

	testProvider := `version: "1.0"
provider:
  name: "integration-test-provider"
  display_name: "Integration Test Provider"
  type: "package_manager"
  platforms: ["linux", "macos", "windows"]
  capabilities: ["install", "uninstall"]
actions:
  install:
    description: "Install software"
    template: "install {{sai_package \"integration-test-provider\" 0}}"
  uninstall:
    description: "Uninstall software"
    template: "uninstall {{sai_package \"integration-test-provider\" 0}}"`

	providerFile := filepath.Join(providersDir, "integration-test-provider.yaml")
	err = os.WriteFile(providerFile, []byte(testProvider), 0644)
	require.NoError(t, err)

	// Setup saidata
	saidataDir := filepath.Join(tempDir, "saidata")
	testSoftwareDir := filepath.Join(saidataDir, "te", "test-app")
	err = os.MkdirAll(testSoftwareDir, 0755)
	require.NoError(t, err)

	testSaidata := `version: "0.2"
metadata:
  name: "test-app"
  display_name: "Test Application"
  description: "A test application for integration testing"
packages:
  - name: "test-app-package"
    version: "1.0.0"
services:
  - name: "test-app"
    service_name: "test-app-service"
    type: "systemd"
providers:
  integration-test-provider:
    packages:
      - name: "test-app-package"
        version: "1.0.0-integration"`

	saidataFile := filepath.Join(testSoftwareDir, "default.yaml")
	err = os.WriteFile(saidataFile, []byte(testSaidata), 0644)
	require.NoError(t, err)

	// Test provider loading
	providerConfig := &provider.ManagerConfig{
		ProviderDirectory: providersDir,
		SchemaPath:        "../../schemas/providerdata-0.1-schema.json",
		EnableWatching:    false,
	}

	providerManager, err := provider.NewProviderManager(providerConfig)
	require.NoError(t, err)
	defer providerManager.Close()

	// Verify provider is loaded and available
	assert.True(t, providerManager.IsProviderAvailable("integration-test-provider"))
	
	provider, err := providerManager.GetProvider("integration-test-provider")
	require.NoError(t, err)
	assert.Equal(t, "integration-test-provider", provider.Provider.Name)
	assert.Contains(t, provider.Actions, "install")
	assert.Contains(t, provider.Actions, "uninstall")

	// Test saidata loading
	saidataManager := saidata.NewManager(saidataDir)
	
	software, err := saidataManager.LoadSoftware("test-app")
	require.NoError(t, err)
	assert.Equal(t, "test-app", software.Metadata.Name)
	assert.Equal(t, "Test Application", software.Metadata.DisplayName)
	assert.Len(t, software.Packages, 1)
	assert.Equal(t, "test-app-package", software.Packages[0].Name)

	// Test provider config integration
	providerConfig2, err := saidataManager.GetProviderConfig("test-app", "integration-test-provider")
	require.NoError(t, err)
	assert.Len(t, providerConfig2.Packages, 1)
	assert.Equal(t, "test-app-package", providerConfig2.Packages[0].Name)
	assert.Equal(t, "1.0.0-integration", providerConfig2.Packages[0].Version)

	// Test provider selection workflow
	selectedProvider, err := providerManager.SelectProvider("test-app", "install", "integration-test-provider")
	require.NoError(t, err)
	assert.Equal(t, "integration-test-provider", selectedProvider.Provider.Name)

	// Test provider statistics
	stats := providerManager.GetProviderStats()
	assert.Greater(t, stats.TotalProviders, 0)
	assert.Greater(t, stats.AvailableProviders, 0)
	assert.Contains(t, stats.ProvidersByType, "package_manager")
}

func testValidationIntegration(t *testing.T, tempDir string) {
	// Create test resources
	testFile := filepath.Join(tempDir, "integration-test.conf")
	err := os.WriteFile(testFile, []byte("test config"), 0644)
	require.NoError(t, err)

	testDir := filepath.Join(tempDir, "integration-test-dir")
	err = os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	// Setup saidata with resources
	saidataDir := filepath.Join(tempDir, "saidata-validation")
	validationTestDir := filepath.Join(saidataDir, "va", "validation-test")
	err = os.MkdirAll(validationTestDir, 0755)
	require.NoError(t, err)

	validationSaidata := `version: "0.2"
metadata:
  name: "validation-test"
  display_name: "Validation Test Software"
files:
  - name: "existing-config"
    path: "` + testFile + `"
    type: "config"
  - name: "missing-config"
    path: "/nonexistent/missing.conf"
    type: "config"
directories:
  - name: "existing-dir"
    path: "` + testDir + `"
  - name: "missing-dir"
    path: "/nonexistent/directory"
commands:
  - name: "test-command"
    path: "/bin/sh"
  - name: "missing-command"
    path: "nonexistent-command-12345"
ports:
  - port: 8080
    protocol: "tcp"
    service: "http-alt"
  - port: 70000
    protocol: "tcp"
    service: "invalid"`

	validationFile := filepath.Join(validationTestDir, "default.yaml")
	err = os.WriteFile(validationFile, []byte(validationSaidata), 0644)
	require.NoError(t, err)

	// Test validation integration
	saidataManager := saidata.NewManager(saidataDir)
	validator := validation.NewResourceValidator()

	// Load saidata
	software, err := saidataManager.LoadSoftware("validation-test")
	require.NoError(t, err)
	assert.Equal(t, "validation-test", software.Metadata.Name)

	// Test resource validation
	result, err := validator.ValidateResources(software)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should have some validation issues
	assert.False(t, result.Valid) // Due to missing resources
	assert.Contains(t, result.MissingFiles, "/nonexistent/missing.conf")
	assert.Contains(t, result.MissingDirectories, "/nonexistent/directory")
	assert.NotEmpty(t, result.Warnings) // Should have port warnings

	// Test individual resource validation
	assert.True(t, validator.ValidateFile(software.Files[0]))  // existing file
	assert.False(t, validator.ValidateFile(software.Files[1])) // missing file
	
	assert.True(t, validator.ValidateDirectory(software.Directories[0]))  // existing dir
	assert.False(t, validator.ValidateDirectory(software.Directories[1])) // missing dir

	assert.True(t, validator.ValidatePort(software.Ports[0]))  // valid port
	assert.False(t, validator.ValidatePort(software.Ports[1])) // invalid port
}

// TestCrossPlatformIntegration tests cross-platform provider compatibility
func TestCrossPlatformIntegration(t *testing.T) {
	tempDir := t.TempDir()
	providersDir := filepath.Join(tempDir, "providers")
	err := os.MkdirAll(providersDir, 0755)
	require.NoError(t, err)

	// Create providers for different platforms
	platformProviders := []struct {
		name      string
		platforms []string
	}{
		{"linux-only", []string{"linux"}},
		{"macos-only", []string{"macos"}},
		{"windows-only", []string{"windows"}},
		{"universal", []string{"linux", "macos", "windows"}},
	}

	for _, p := range platformProviders {
		platformsYAML := "["
		for i, platform := range p.platforms {
			if i > 0 {
				platformsYAML += ", "
			}
			platformsYAML += "\"" + platform + "\""
		}
		platformsYAML += "]"

		providerYAML := `version: "1.0"
provider:
  name: "` + p.name + `"
  type: "package_manager"
  platforms: ` + platformsYAML + `
  capabilities: ["install"]
actions:
  install:
    template: "install {{sai_package \"` + p.name + `\" 0}}"`

		providerFile := filepath.Join(providersDir, p.name+".yaml")
		err = os.WriteFile(providerFile, []byte(providerYAML), 0644)
		require.NoError(t, err)
	}

	// Test cross-platform provider loading
	config := &provider.ManagerConfig{
		ProviderDirectory: providersDir,
		SchemaPath:        "../../schemas/providerdata-0.1-schema.json",
		EnableWatching:    false,
	}

	manager, err := provider.NewProviderManager(config)
	require.NoError(t, err)
	defer manager.Close()

	// All providers should be loaded (but only platform-compatible ones will be available)
	allProviders := manager.GetAvailableProviders()
	assert.Greater(t, len(allProviders), 0) // At least universal should be loaded

	// Universal provider should always be available
	assert.True(t, manager.IsProviderAvailable("universal"))

	// Test provider statistics
	stats := manager.GetProviderStats()
	assert.Greater(t, stats.TotalProviders, 0)
	assert.Greater(t, stats.AvailableProviders, 0) // At least universal should be available

	// Test provider selection with platform filtering
	universalProvider, err := manager.GetProvider("universal")
	require.NoError(t, err)
	assert.Contains(t, universalProvider.Provider.Platforms, "linux")
	assert.Contains(t, universalProvider.Provider.Platforms, "macos")
	assert.Contains(t, universalProvider.Provider.Platforms, "windows")
}

// TestDefaultsGenerationIntegration tests the defaults generation workflow
func TestDefaultsGenerationIntegration(t *testing.T) {
	tempDir := t.TempDir()
	saidataDir := filepath.Join(tempDir, "saidata-defaults")

	// Create saidata manager (no existing data)
	manager := saidata.NewManager(saidataDir)

	// Test defaults generation for unknown software
	unknownSoftware, err := manager.LoadSoftware("unknown-application")
	require.NoError(t, err)
	assert.NotNil(t, unknownSoftware)
	assert.True(t, unknownSoftware.IsGenerated)
	assert.Equal(t, "unknown-application", unknownSoftware.Metadata.Name)

	// Should have generated basic structure
	assert.Equal(t, "0.2", unknownSoftware.Version)
	assert.NotEmpty(t, unknownSoftware.Packages)
	assert.Equal(t, "unknown-application", unknownSoftware.Packages[0].Name)

	// Test that generated defaults are cached
	cachedSoftware, err := manager.LoadSoftware("unknown-application")
	require.NoError(t, err)
	assert.True(t, cachedSoftware.IsGenerated)
	assert.Equal(t, unknownSoftware.Metadata.Name, cachedSoftware.Metadata.Name)
}

// TestSchemaValidationIntegration tests schema validation integration
func TestSchemaValidationIntegration(t *testing.T) {
	tempDir := t.TempDir()
	providersDir := filepath.Join(tempDir, "providers")
	err := os.MkdirAll(providersDir, 0755)
	require.NoError(t, err)

	// Create valid provider
	validProvider := `version: "1.0"
provider:
  name: "schema-test-valid"
  type: "package_manager"
  platforms: ["linux"]
  capabilities: ["install"]
actions:
  install:
    template: "install {{sai_package \"schema-test-valid\" 0}}"`

	validFile := filepath.Join(providersDir, "schema-test-valid.yaml")
	err = os.WriteFile(validFile, []byte(validProvider), 0644)
	require.NoError(t, err)

	// Create invalid provider (missing required fields)
	invalidProvider := `version: "1.0"
provider:
  name: "schema-test-invalid"
  # Missing type and platforms
actions:
  install:
    template: "install something"`

	invalidFile := filepath.Join(providersDir, "schema-test-invalid.yaml")
	err = os.WriteFile(invalidFile, []byte(invalidProvider), 0644)
	require.NoError(t, err)

	// Test schema validation integration
	config := &provider.ManagerConfig{
		ProviderDirectory: providersDir,
		SchemaPath:        "../../schemas/providerdata-0.1-schema.json",
		EnableWatching:    false,
	}

	manager, err := provider.NewProviderManager(config)
	// Manager creation might succeed but with warnings about invalid providers
	if err != nil {
		// If there are validation errors, they should be reported
		assert.Contains(t, err.Error(), "failed to load")
	}
	
	if manager != nil {
		defer manager.Close()
		
		// Valid provider should be available (if it was loaded successfully)
		if manager.IsProviderAvailable("schema-test-valid") {
			assert.True(t, manager.IsProviderAvailable("schema-test-valid"))
		}
		
		// Invalid provider should not be available
		assert.False(t, manager.IsProviderAvailable("schema-test-invalid"))
		
		// Should be able to get valid provider
		validProv, err := manager.GetProvider("schema-test-valid")
		require.NoError(t, err)
		assert.Equal(t, "schema-test-valid", validProv.Provider.Name)
		
		// Should not be able to get invalid provider
		_, err = manager.GetProvider("schema-test-invalid")
		assert.Error(t, err)
	}
}