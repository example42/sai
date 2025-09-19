package provider

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sai/internal/types"
)

func TestProviderLoader_LoadFromFile(t *testing.T) {
	// Create a temporary provider file
	tempDir := t.TempDir()
	providerFile := filepath.Join(tempDir, "test-provider.yaml")
	
	providerYAML := `version: "1.0"
provider:
  name: "test"
  display_name: "Test Provider"
  description: "A test provider"
  type: "package_manager"
  platforms: ["linux"]
  executable: "test-cmd"
  capabilities: ["install", "uninstall"]
actions:
  install:
    description: "Install packages"
    template: "test-cmd install {{sai_package(0, 'name', 'test')}}"
    timeout: 300
  uninstall:
    description: "Uninstall packages"
    template: "test-cmd remove {{sai_package(0, 'name', 'test')}}"
    timeout: 300`

	err := os.WriteFile(providerFile, []byte(providerYAML), 0644)
	require.NoError(t, err)

	// Create provider loader
	loader, err := NewProviderLoader("../../schemas/providerdata-0.1-schema.json")
	require.NoError(t, err)

	// Load the provider
	provider, err := loader.LoadFromFile(providerFile)
	require.NoError(t, err)
	assert.NotNil(t, provider)

	// Verify provider data
	assert.Equal(t, "1.0", provider.Version)
	assert.Equal(t, "test", provider.Provider.Name)
	assert.Equal(t, "Test Provider", provider.Provider.DisplayName)
	assert.Equal(t, "package_manager", provider.Provider.Type)
	assert.Equal(t, []string{"linux"}, provider.Provider.Platforms)
	assert.Equal(t, "test-cmd", provider.Provider.Executable)
	assert.Equal(t, []string{"install", "uninstall"}, provider.Provider.Capabilities)

	// Verify actions
	assert.Len(t, provider.Actions, 2)
	
	installAction, exists := provider.Actions["install"]
	assert.True(t, exists)
	assert.Equal(t, "Install packages", installAction.Description)
	assert.Equal(t, "test-cmd install {{sai_package(0, 'name', 'test')}}", installAction.Template)
	assert.Equal(t, 300, installAction.Timeout)

	uninstallAction, exists := provider.Actions["uninstall"]
	assert.True(t, exists)
	assert.Equal(t, "Uninstall packages", uninstallAction.Description)
	assert.Equal(t, "test-cmd remove {{sai_package(0, 'name', 'test')}}", uninstallAction.Template)
}

func TestProviderLoader_LoadFromFile_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	providerFile := filepath.Join(tempDir, "invalid.yaml")
	
	invalidYAML := `invalid: yaml: content: [unclosed`
	err := os.WriteFile(providerFile, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	loader, err := NewProviderLoader("../../schemas/providerdata-0.1-schema.json")
	require.NoError(t, err)

	_, err = loader.LoadFromFile(providerFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse provider YAML")
}

func TestProviderLoader_LoadFromFile_SchemaValidation(t *testing.T) {
	tempDir := t.TempDir()
	providerFile := filepath.Join(tempDir, "invalid-schema.yaml")
	
	// Missing required fields
	invalidYAML := `version: "1.0"
provider:
  name: "test"
  # missing type field
actions: {}`

	err := os.WriteFile(providerFile, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	loader, err := NewProviderLoader("../../schemas/providerdata-0.1-schema.json")
	require.NoError(t, err)

	_, err = loader.LoadFromFile(providerFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider validation failed")
}

func TestProviderLoader_LoadFromDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple provider files
	providers := []struct {
		filename string
		content  string
	}{
		{
			"apt.yaml",
			`version: "1.0"
provider:
  name: "apt"
  type: "package_manager"
  platforms: ["debian", "ubuntu"]
  executable: "apt-get"
actions:
  install:
    template: "apt-get install -y {{sai_package(0, 'name', 'apt')}}"`,
		},
		{
			"brew.yaml",
			`version: "1.0"
provider:
  name: "brew"
  type: "package_manager"
  platforms: ["macos"]
  executable: "brew"
actions:
  install:
    template: "brew install {{sai_package(0, 'name', 'brew')}}"`,
		},
	}

	for _, p := range providers {
		providerFile := filepath.Join(tempDir, p.filename)
		err := os.WriteFile(providerFile, []byte(p.content), 0644)
		require.NoError(t, err)
	}

	// Create a non-YAML file that should be ignored
	nonYAMLFile := filepath.Join(tempDir, "readme.txt")
	err := os.WriteFile(nonYAMLFile, []byte("This is not a YAML file"), 0644)
	require.NoError(t, err)

	loader, err := NewProviderLoader("../../schemas/providerdata-0.1-schema.json")
	require.NoError(t, err)

	loadedProviders, err := loader.LoadFromDirectory(tempDir)
	require.NoError(t, err)
	assert.Len(t, loadedProviders, 2)

	// Verify providers were loaded correctly
	providerNames := make([]string, len(loadedProviders))
	for i, provider := range loadedProviders {
		providerNames[i] = provider.Provider.Name
	}
	assert.Contains(t, providerNames, "apt")
	assert.Contains(t, providerNames, "brew")
}

func TestProviderLoader_LoadFromDirectory_WithErrors(t *testing.T) {
	tempDir := t.TempDir()

	// Create one valid and one invalid provider
	validProvider := filepath.Join(tempDir, "valid.yaml")
	validYAML := `version: "1.0"
provider:
  name: "valid"
  type: "package_manager"
actions:
  install:
    template: "install {{sai_package(0, 'name', 'valid')}}"`

	err := os.WriteFile(validProvider, []byte(validYAML), 0644)
	require.NoError(t, err)

	invalidProvider := filepath.Join(tempDir, "invalid.yaml")
	invalidYAML := `invalid: yaml: [unclosed`
	err = os.WriteFile(invalidProvider, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	loader, err := NewProviderLoader("../../schemas/providerdata-0.1-schema.json")
	require.NoError(t, err)

	loadedProviders, err := loader.LoadFromDirectory(tempDir)
	
	// Should return the valid provider but also report errors
	assert.Len(t, loadedProviders, 1)
	assert.Equal(t, "valid", loadedProviders[0].Provider.Name)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "some providers failed to load")
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
					Name: "test",
					Type: "package_manager",
				},
				Actions: map[string]types.Action{
					"install": {
						Template: "install {{sai_package(0, 'name', 'test')}}",
					},
				},
			},
			wantError: false,
		},
		{
			name: "empty provider name",
			provider: &types.ProviderData{
				Version: "1.0",
				Provider: types.ProviderInfo{
					Name: "",
					Type: "package_manager",
				},
				Actions: map[string]types.Action{
					"install": {
						Template: "install {{sai_package(0, 'name', 'test')}}",
					},
				},
			},
			wantError: true,
			errorMsg:  "provider name cannot be empty",
		},
		{
			name: "no actions",
			provider: &types.ProviderData{
				Version: "1.0",
				Provider: types.ProviderInfo{
					Name: "test",
					Type: "package_manager",
				},
				Actions: map[string]types.Action{},
			},
			wantError: true,
			errorMsg:  "provider must define at least one action",
		},
		{
			name: "invalid action",
			provider: &types.ProviderData{
				Version: "1.0",
				Provider: types.ProviderInfo{
					Name: "test",
					Type: "package_manager",
				},
				Actions: map[string]types.Action{
					"install": {
						// No template, command, script, or steps
					},
				},
			},
			wantError: true,
			errorMsg:  "Must validate one and only one schema",
		},
		{
			name: "invalid provider type",
			provider: &types.ProviderData{
				Version: "1.0",
				Provider: types.ProviderInfo{
					Name: "test",
					Type: "invalid_type",
				},
				Actions: map[string]types.Action{
					"install": {
						Template: "install {{sai_package(0, 'name', 'test')}}",
					},
				},
			},
			wantError: true,
			errorMsg:  "provider.type must be one of the following",
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

func TestProviderLoader_WatchDirectory(t *testing.T) {
	tempDir := t.TempDir()
	
	loader, err := NewProviderLoader("../../schemas/providerdata-0.1-schema.json")
	require.NoError(t, err)

	// Channel to receive callback notifications
	callbackChan := make(chan *types.ProviderData, 1)
	
	// Set up watching
	err = loader.WatchDirectory(tempDir, func(provider *types.ProviderData) {
		callbackChan <- provider
	})
	require.NoError(t, err)

	// Create a provider file
	providerFile := filepath.Join(tempDir, "watched.yaml")
	providerYAML := `version: "1.0"
provider:
  name: "watched"
  type: "package_manager"
actions:
  install:
    template: "install {{sai_package(0, 'name', 'watched')}}"`

	err = os.WriteFile(providerFile, []byte(providerYAML), 0644)
	require.NoError(t, err)

	// Wait for the callback with timeout
	select {
	case provider := <-callbackChan:
		assert.Equal(t, "watched", provider.Provider.Name)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for file watch callback")
	}

	// Clean up
	err = loader.StopWatching(tempDir)
	assert.NoError(t, err)
}

func TestProviderLoader_GetSupportedProviderTypes(t *testing.T) {
	loader, err := NewProviderLoader("../../schemas/providerdata-0.1-schema.json")
	require.NoError(t, err)

	types := loader.GetSupportedProviderTypes()
	assert.NotEmpty(t, types)
	
	// Check for some expected types
	assert.Contains(t, types, "package_manager")
	assert.Contains(t, types, "container")
	assert.Contains(t, types, "debug")
	assert.Contains(t, types, "security")
}

func TestNewProviderLoader_InvalidSchema(t *testing.T) {
	_, err := NewProviderLoader("nonexistent-schema.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load provider schema")
}

func TestProviderLoader_LoadRealProviders(t *testing.T) {
	// Test loading actual provider files from the providers directory
	loader, err := NewProviderLoader("../../schemas/providerdata-0.1-schema.json")
	require.NoError(t, err)

	// Test loading apt provider
	aptProvider, err := loader.LoadFromFile("../../providers/apt.yaml")
	if err != nil {
		t.Skipf("Skipping real provider test - apt.yaml not found or invalid: %v", err)
		return
	}

	assert.Equal(t, "apt", aptProvider.Provider.Name)
	assert.Equal(t, "package_manager", aptProvider.Provider.Type)
	assert.Contains(t, aptProvider.Provider.Platforms, "debian")
	assert.Contains(t, aptProvider.Provider.Platforms, "ubuntu")
	assert.Equal(t, "apt-get", aptProvider.Provider.Executable)
	assert.NotEmpty(t, aptProvider.Actions)

	// Test loading brew provider
	brewProvider, err := loader.LoadFromFile("../../providers/brew.yaml")
	if err != nil {
		t.Skipf("Skipping real provider test - brew.yaml not found or invalid: %v", err)
		return
	}

	assert.Equal(t, "brew", brewProvider.Provider.Name)
	assert.Equal(t, "package_manager", brewProvider.Provider.Type)
	assert.Contains(t, brewProvider.Provider.Platforms, "macos")
	assert.Equal(t, "brew", brewProvider.Provider.Executable)
	assert.NotEmpty(t, brewProvider.Actions)
}