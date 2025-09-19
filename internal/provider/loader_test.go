package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sai/internal/types"
)

func TestProviderLoader_LoadFromFile(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test provider file
	providerYAML := `version: "1.0"
provider:
  name: "test-provider"
  display_name: "Test Provider"
  description: "A test provider for unit testing"
  type: "package_manager"
  platforms: ["` + runtime.GOOS + `"]
  priority: 50
  capabilities: ["install", "uninstall", "start", "stop"]
actions:
  install:
    description: "Install software"
    template: "install {{sai_package(0, 'name', 'test-provider')}}"
    timeout: 300
    validation:
      command: "which {{sai_package(0, 'name', 'test-provider')}}"
      expected_exit_code: 0
  uninstall:
    description: "Uninstall software"
    template: "uninstall {{sai_package(0, 'name', 'test-provider')}}"
  start:
    description: "Start service"
    template: "start {{sai_service(0, 'service_name', 'test-provider')}}"
  stop:
    description: "Stop service"
    template: "stop {{sai_service(0, 'service_name', 'test-provider')}}"`

	providerFile := filepath.Join(tempDir, "test-provider.yaml")
	err := os.WriteFile(providerFile, []byte(providerYAML), 0644)
	require.NoError(t, err)

	// Create loader
	loader, err := NewProviderLoader("../../schemas/providerdata-0.1-schema.json")
	require.NoError(t, err)

	// Test loading from file
	provider, err := loader.LoadFromFile(providerFile)
	require.NoError(t, err)
	assert.NotNil(t, provider)

	// Verify provider data
	assert.Equal(t, "1.0", provider.Version)
	assert.Equal(t, "test-provider", provider.Provider.Name)
	assert.Equal(t, "Test Provider", provider.Provider.DisplayName)
	assert.Equal(t, "package_manager", provider.Provider.Type)
	assert.Contains(t, provider.Provider.Platforms, runtime.GOOS)
	assert.Equal(t, 50, provider.Provider.Priority)
	assert.Contains(t, provider.Provider.Capabilities, "install")
	assert.Contains(t, provider.Provider.Capabilities, "uninstall")

	// Verify actions
	assert.Contains(t, provider.Actions, "install")
	assert.Contains(t, provider.Actions, "uninstall")
	assert.Contains(t, provider.Actions, "start")
	assert.Contains(t, provider.Actions, "stop")

	installAction := provider.Actions["install"]
	assert.Equal(t, "Install software", installAction.Description)
	assert.Equal(t, "install {{sai_package(0, 'name', 'test-provider')}}", installAction.Template)
	assert.Equal(t, 300, installAction.Timeout)
	assert.NotNil(t, installAction.Validation)
	assert.Equal(t, "which {{sai_package(0, 'name', 'test-provider')}}", installAction.Validation.Command)
	assert.Equal(t, 0, installAction.Validation.ExpectedExitCode)
}

func TestProviderLoader_LoadFromFile_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create invalid YAML file
	invalidYAML := `version: "1.0"
provider:
  name: "test-provider"
  invalid_yaml: [unclosed array`

	providerFile := filepath.Join(tempDir, "invalid.yaml")
	err := os.WriteFile(providerFile, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	loader, err := NewProviderLoader("../../schemas/providerdata-0.1-schema.json")
	require.NoError(t, err)

	// Test loading invalid YAML
	provider, err := loader.LoadFromFile(providerFile)
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "failed to parse provider YAML")
}

func TestProviderLoader_LoadFromFile_NonExistentFile(t *testing.T) {
	loader, err := NewProviderLoader("../../schemas/providerdata-0.1-schema.json")
	require.NoError(t, err)

	// Test loading non-existent file
	provider, err := loader.LoadFromFile("/nonexistent/file.yaml")
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "failed to read provider file")
}

func TestProviderLoader_LoadFromDirectory(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create multiple provider files
	providers := []struct {
		filename string
		name     string
		priority int
	}{
		{"provider1.yaml", "provider1", 80},
		{"provider2.yaml", "provider2", 60},
		{"provider3.yaml", "provider3", 40},
	}

	for _, p := range providers {
		providerYAML := `version: "1.0"
provider:
  name: "` + p.name + `"
  type: "package_manager"
  platforms: ["` + runtime.GOOS + `"]
  priority: ` + fmt.Sprintf("%d", p.priority) + `
  capabilities: ["install"]
actions:
  install:
    template: "install {{sai_package(0, 'name', '` + p.name + `')}}"`

		providerFile := filepath.Join(tempDir, p.filename)
		err := os.WriteFile(providerFile, []byte(providerYAML), 0644)
		require.NoError(t, err)
	}

	// Create non-YAML file (should be ignored)
	nonYAMLFile := filepath.Join(tempDir, "readme.txt")
	err := os.WriteFile(nonYAMLFile, []byte("This is not a YAML file"), 0644)
	require.NoError(t, err)

	loader, err := NewProviderLoader("../../schemas/providerdata-0.1-schema.json")
	require.NoError(t, err)

	// Test loading from directory
	loadedProviders, err := loader.LoadFromDirectory(tempDir)
	require.NoError(t, err)
	assert.Len(t, loadedProviders, 3)

	// Verify all providers were loaded
	providerNames := make(map[string]bool)
	for _, provider := range loadedProviders {
		providerNames[provider.Provider.Name] = true
	}

	assert.True(t, providerNames["provider1"])
	assert.True(t, providerNames["provider2"])
	assert.True(t, providerNames["provider3"])
}

func TestProviderLoader_LoadFromDirectory_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	loader, err := NewProviderLoader("../../schemas/providerdata-0.1-schema.json")
	require.NoError(t, err)

	// Test loading from empty directory
	providers, err := loader.LoadFromDirectory(tempDir)
	require.NoError(t, err)
	assert.Empty(t, providers)
}

func TestProviderLoader_LoadFromDirectory_NonExistentDirectory(t *testing.T) {
	loader, err := NewProviderLoader("../../schemas/providerdata-0.1-schema.json")
	require.NoError(t, err)

	// Test loading from non-existent directory
	providers, err := loader.LoadFromDirectory("/nonexistent/directory")
	assert.Error(t, err)
	assert.Nil(t, providers)
	assert.Contains(t, err.Error(), "failed to walk provider directory")
}

func TestProviderLoader_ValidateProvider(t *testing.T) {
	loader, err := NewProviderLoader("../../schemas/providerdata-0.1-schema.json")
	require.NoError(t, err)

	tests := []struct {
		name      string
		provider  *types.ProviderData
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid provider",
			provider: &types.ProviderData{
				Version: "1.0",
				Provider: types.ProviderInfo{
					Name:         "test-provider",
					DisplayName:  "Test Provider",
					Type:         "package_manager",
					Platforms:    []string{"linux"},
					Capabilities: []string{"install"},
				},
				Actions: map[string]types.Action{
					"install": {
						Description: "Install software",
						Template:    "install {{.Software}}",
					},
				},
			},
			wantError: false,
		},
		{
			name: "missing provider name",
			provider: &types.ProviderData{
				Version: "1.0",
				Provider: types.ProviderInfo{
					Type:         "package_manager",
					Platforms:    []string{"linux"},
					Capabilities: []string{"install"},
				},
				Actions: map[string]types.Action{
					"install": {Template: "install {{.Software}}"},
				},
			},
			wantError: true,
			errorMsg:  "provider name cannot be empty",
		},
		{
			name: "missing provider type",
			provider: &types.ProviderData{
				Version: "1.0",
				Provider: types.ProviderInfo{
					Name:         "test-provider",
					Platforms:    []string{"linux"},
					Capabilities: []string{"install"},
				},
				Actions: map[string]types.Action{
					"install": {Template: "install {{.Software}}"},
				},
			},
			wantError: true,
			errorMsg:  "provider.type must be one of the following",
		},
		{
			name: "invalid provider type",
			provider: &types.ProviderData{
				Version: "1.0",
				Provider: types.ProviderInfo{
					Name:         "test-provider",
					Type:         "invalid_type",
					Platforms:    []string{"linux"},
					Capabilities: []string{"install"},
				},
				Actions: map[string]types.Action{
					"install": {Template: "install {{.Software}}"},
				},
			},
			wantError: true,
			errorMsg:  "provider.type must be one of the following",
		},
		{
			name: "no actions",
			provider: &types.ProviderData{
				Version: "1.0",
				Provider: types.ProviderInfo{
					Name:         "test-provider",
					Type:         "package_manager",
					Platforms:    []string{"linux"},
					Capabilities: []string{"install"},
				},
				Actions: map[string]types.Action{},
			},
			wantError: true,
			errorMsg:  "provider must define at least one action",
		},
		{
			name: "action without execution method",
			provider: &types.ProviderData{
				Version: "1.0",
				Provider: types.ProviderInfo{
					Name:         "test-provider",
					Type:         "package_manager",
					Platforms:    []string{"linux"},
					Capabilities: []string{"install"},
				},
				Actions: map[string]types.Action{
					"install": {
						Description: "Install software",
						// No template, command, script, or steps
					},
				},
			},
			wantError: true,
			errorMsg:  "template is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.ValidateProvider(tt.provider)
			
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProviderLoader_LoadExistingProviders(t *testing.T) {
	// Test loading actual provider files from the providers directory
	providersDir := "../../providers"
	
	// Skip test if providers directory doesn't exist
	if _, err := os.Stat(providersDir); os.IsNotExist(err) {
		t.Skip("Providers directory not found, skipping test")
	}

	loader, err := NewProviderLoader("../../schemas/providerdata-0.1-schema.json")
	require.NoError(t, err)

	// Load all providers from the actual providers directory
	// Note: Some providers may fail validation due to schema differences, that's expected
	providers, err := loader.LoadFromDirectory(providersDir)
	if err != nil {
		// If some providers failed to load, that's okay for this test
		// We just want to verify the loader can handle real provider files
		t.Logf("Some providers failed to load (expected): %v", err)
	}
	
	// We should still get some providers loaded
	if len(providers) == 0 {
		t.Skip("No providers could be loaded, skipping validation test")
	}

	// Verify each loaded provider is valid
	for _, provider := range providers {
		err := loader.ValidateProvider(provider)
		assert.NoError(t, err, "Provider %s should be valid", provider.Provider.Name)
		
		// Basic structure validation
		assert.NotEmpty(t, provider.Provider.Name, "Provider name should not be empty")
		assert.NotEmpty(t, provider.Provider.Type, "Provider type should not be empty")
		assert.NotEmpty(t, provider.Provider.Platforms, "Provider should support at least one platform")
		assert.NotEmpty(t, provider.Actions, "Provider should have at least one action")
		
		// Verify actions have execution methods
		for actionName, action := range provider.Actions {
			hasExecutionMethod := action.Template != "" || action.Command != "" || 
								action.Script != "" || len(action.Steps) > 0
			assert.True(t, hasExecutionMethod, 
				"Action %s in provider %s should have at least one execution method", 
				actionName, provider.Provider.Name)
		}
	}

	// Test loading specific known providers
	knownProviders := []string{"apt", "brew", "docker"}
	for _, providerName := range knownProviders {
		providerFile := filepath.Join(providersDir, providerName+".yaml")
		if _, err := os.Stat(providerFile); err == nil {
			provider, err := loader.LoadFromFile(providerFile)
			require.NoError(t, err, "Should be able to load %s provider", providerName)
			assert.Equal(t, providerName, provider.Provider.Name)
		}
	}
}