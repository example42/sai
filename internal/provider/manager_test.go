package provider

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProviderManager(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test provider files
	createTestProviders(t, tempDir)

	config := &ManagerConfig{
		ProviderDirectory: tempDir,
		SchemaPath:        "../../schemas/providerdata-0.1-schema.json",
		DefaultProvider:   "test1",
		ProviderPriority:  map[string]int{"test1": 100},
		EnableWatching:    false,
	}

	manager, err := NewProviderManager(config)
	require.NoError(t, err)
	assert.NotNil(t, manager)

	// Verify providers were loaded
	providers := manager.GetAvailableProviders()
	assert.NotEmpty(t, providers)

	// Clean up
	err = manager.Close()
	assert.NoError(t, err)
}

func TestProviderManager_GetProvider(t *testing.T) {
	tempDir := t.TempDir()
	createTestProviders(t, tempDir)

	config := &ManagerConfig{
		ProviderDirectory: tempDir,
		SchemaPath:        "../../schemas/providerdata-0.1-schema.json",
		EnableWatching:    false,
	}

	manager, err := NewProviderManager(config)
	require.NoError(t, err)
	defer manager.Close()

	// Test getting existing provider
	provider, err := manager.GetProvider("test1")
	require.NoError(t, err)
	assert.Equal(t, "test1", provider.Provider.Name)

	// Test getting non-existing provider
	_, err = manager.GetProvider("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider nonexistent not found")
}

func TestProviderManager_SelectProvider(t *testing.T) {
	tempDir := t.TempDir()
	createTestProviders(t, tempDir)

	config := &ManagerConfig{
		ProviderDirectory: tempDir,
		SchemaPath:        "../../schemas/providerdata-0.1-schema.json",
		EnableWatching:    false,
	}

	manager, err := NewProviderManager(config)
	require.NoError(t, err)
	defer manager.Close()

	tests := []struct {
		name              string
		software          string
		action            string
		preferredProvider string
		wantError         bool
		expectedProvider  string
	}{
		{
			name:              "select with preferred provider",
			software:          "nginx",
			action:            "install",
			preferredProvider: "test1",
			wantError:         false,
			expectedProvider:  "test1",
		},
		{
			name:              "select with non-existent preferred provider",
			software:          "nginx",
			action:            "install",
			preferredProvider: "nonexistent",
			wantError:         true,
		},
		{
			name:              "automatic selection",
			software:          "nginx",
			action:            "install",
			preferredProvider: "",
			wantError:         false,
		},
		{
			name:              "action not supported",
			software:          "nginx",
			action:            "unsupported-action",
			preferredProvider: "",
			wantError:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := manager.SelectProvider(tt.software, tt.action, tt.preferredProvider)
			
			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.expectedProvider != "" {
					assert.Equal(t, tt.expectedProvider, provider.Provider.Name)
				}
			}
		})
	}
}

func TestProviderManager_GetProvidersForAction(t *testing.T) {
	tempDir := t.TempDir()
	createTestProviders(t, tempDir)

	config := &ManagerConfig{
		ProviderDirectory: tempDir,
		SchemaPath:        "../../schemas/providerdata-0.1-schema.json",
		EnableWatching:    false,
	}

	manager, err := NewProviderManager(config)
	require.NoError(t, err)
	defer manager.Close()

	// Test action supported by multiple providers
	providers := manager.GetProvidersForAction("install")
	assert.NotEmpty(t, providers)

	// Test action not supported by any provider
	providers = manager.GetProvidersForAction("unsupported-action")
	assert.Empty(t, providers)
}

func TestProviderManager_GetProviderSelections(t *testing.T) {
	tempDir := t.TempDir()
	createTestProviders(t, tempDir)

	config := &ManagerConfig{
		ProviderDirectory: tempDir,
		SchemaPath:        "../../schemas/providerdata-0.1-schema.json",
		EnableWatching:    false,
	}

	manager, err := NewProviderManager(config)
	require.NoError(t, err)
	defer manager.Close()

	selections, err := manager.GetProviderSelections("nginx", "install")
	require.NoError(t, err)
	assert.NotEmpty(t, selections)

	// Verify selections are sorted (available first, then by priority)
	for i := 0; i < len(selections)-1; i++ {
		current := selections[i]
		next := selections[i+1]
		
		if current.Available != next.Available {
			assert.True(t, current.Available, "Available providers should come first")
		} else if current.Available && next.Available {
			assert.GreaterOrEqual(t, current.Priority, next.Priority, "Should be sorted by priority")
		}
	}
}

func TestProviderManager_GetMultipleProviderOptions(t *testing.T) {
	tempDir := t.TempDir()
	createTestProviders(t, tempDir)

	config := &ManagerConfig{
		ProviderDirectory: tempDir,
		SchemaPath:        "../../schemas/providerdata-0.1-schema.json",
		EnableWatching:    false,
	}

	manager, err := NewProviderManager(config)
	require.NoError(t, err)
	defer manager.Close()

	options, err := manager.GetMultipleProviderOptions("nginx", "install")
	require.NoError(t, err)
	assert.NotEmpty(t, options)

	// Verify all options are available
	for _, option := range options {
		assert.NotNil(t, option.Provider)
		assert.NotEmpty(t, option.PackageName)
	}
}

func TestProviderManager_SelectProviderWithFallback(t *testing.T) {
	tempDir := t.TempDir()
	createTestProviders(t, tempDir)

	config := &ManagerConfig{
		ProviderDirectory: tempDir,
		SchemaPath:        "../../schemas/providerdata-0.1-schema.json",
		DefaultProvider:   "test2",
		EnableWatching:    false,
	}

	manager, err := NewProviderManager(config)
	require.NoError(t, err)
	defer manager.Close()

	tests := []struct {
		name              string
		preferredProvider string
		wantError         bool
	}{
		{
			name:              "preferred provider works",
			preferredProvider: "test1",
			wantError:         false,
		},
		{
			name:              "fallback to default",
			preferredProvider: "nonexistent",
			wantError:         false, // Should fallback to default
		},
		{
			name:              "automatic selection",
			preferredProvider: "",
			wantError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := manager.SelectProviderWithFallback("nginx", "install", tt.preferredProvider)
			
			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestProviderManager_GetProviderStats(t *testing.T) {
	tempDir := t.TempDir()
	createTestProviders(t, tempDir)

	config := &ManagerConfig{
		ProviderDirectory: tempDir,
		SchemaPath:        "../../schemas/providerdata-0.1-schema.json",
		EnableWatching:    false,
	}

	manager, err := NewProviderManager(config)
	require.NoError(t, err)
	defer manager.Close()

	stats := manager.GetProviderStats()
	assert.NotNil(t, stats)
	assert.Greater(t, stats.TotalProviders, 0)
	assert.NotEmpty(t, stats.ProvidersByType)
	assert.NotEmpty(t, stats.String())
}

func TestProviderManager_IsProviderAvailable(t *testing.T) {
	tempDir := t.TempDir()
	createTestProviders(t, tempDir)

	config := &ManagerConfig{
		ProviderDirectory: tempDir,
		SchemaPath:        "../../schemas/providerdata-0.1-schema.json",
		EnableWatching:    false,
	}

	manager, err := NewProviderManager(config)
	require.NoError(t, err)
	defer manager.Close()

	// Test existing provider
	available := manager.IsProviderAvailable("test1")
	assert.True(t, available) // Should be available since it has no executable requirement

	// Test non-existing provider
	available = manager.IsProviderAvailable("nonexistent")
	assert.False(t, available)
}

func TestProviderManager_ReloadProviders(t *testing.T) {
	tempDir := t.TempDir()
	createTestProviders(t, tempDir)

	config := &ManagerConfig{
		ProviderDirectory: tempDir,
		SchemaPath:        "../../schemas/providerdata-0.1-schema.json",
		EnableWatching:    false,
	}

	manager, err := NewProviderManager(config)
	require.NoError(t, err)
	defer manager.Close()

	// Get initial provider count
	initialStats := manager.GetProviderStats()
	initialCount := initialStats.TotalProviders

	// Add a new provider file
	newProviderYAML := `version: "1.0"
provider:
  name: "test-new"
  type: "package_manager"
  platforms: ["` + runtime.GOOS + `"]
actions:
  install:
    template: "install {{sai_package(0, 'name', 'test-new')}}"`

	newProviderFile := filepath.Join(tempDir, "test-new.yaml")
	err = os.WriteFile(newProviderFile, []byte(newProviderYAML), 0644)
	require.NoError(t, err)

	// Reload providers
	err = manager.ReloadProviders()
	require.NoError(t, err)

	// Verify new provider was loaded
	newStats := manager.GetProviderStats()
	assert.Equal(t, initialCount+1, newStats.TotalProviders)

	// Verify we can get the new provider
	provider, err := manager.GetProvider("test-new")
	require.NoError(t, err)
	assert.Equal(t, "test-new", provider.Provider.Name)
}

func TestProviderStats_String(t *testing.T) {
	stats := &ProviderStats{
		TotalProviders:     5,
		AvailableProviders: 3,
		ProvidersByType: map[string]int{
			"package_manager": 3,
			"container":       2,
		},
		ProvidersByPlatform: map[string]int{
			"linux": 4,
			"macos": 1,
		},
	}

	str := stats.String()
	assert.Contains(t, str, "Total: 5")
	assert.Contains(t, str, "Available: 3")
	assert.Contains(t, str, "Types:")
}

// Helper function to create test provider files
func createTestProviders(t *testing.T, tempDir string) {
	providers := []struct {
		filename string
		content  string
	}{
		{
			"test1.yaml",
			`version: "1.0"
provider:
  name: "test1"
  type: "package_manager"
  platforms: ["` + runtime.GOOS + `"]
  priority: 80
  capabilities: ["install", "uninstall", "start", "stop"]
actions:
  install:
    template: "install {{sai_package(0, 'name', 'test1')}}"
  uninstall:
    template: "uninstall {{sai_package(0, 'name', 'test1')}}"
  start:
    template: "start {{sai_service(0, 'service_name', 'test1')}}"
  stop:
    template: "stop {{sai_service(0, 'service_name', 'test1')}}"`,
		},
		{
			"test2.yaml",
			`version: "1.0"
provider:
  name: "test2"
  type: "container"
  platforms: ["` + runtime.GOOS + `"]
  priority: 60
  capabilities: ["install", "uninstall"]
actions:
  install:
    template: "docker run {{sai_package(0, 'name', 'test2')}}"
  uninstall:
    template: "docker rm {{sai_package(0, 'name', 'test2')}}"`,
		},
		{
			"test3.yaml",
			`version: "1.0"
provider:
  name: "test3"
  type: "package_manager"
  platforms: ["incompatible-os"]
  executable: "nonexistent-command"
  capabilities: ["install"]
actions:
  install:
    template: "install {{sai_package(0, 'name', 'test3')}}"`,
		},
	}

	for _, p := range providers {
		providerFile := filepath.Join(tempDir, p.filename)
		err := os.WriteFile(providerFile, []byte(p.content), 0644)
		require.NoError(t, err)
	}
}