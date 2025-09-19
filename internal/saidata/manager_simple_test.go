package saidata

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaidataManager_Basic(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test saidata structure
	apacheDir := filepath.Join(tempDir, "ap", "apache")
	err := os.MkdirAll(apacheDir, 0755)
	require.NoError(t, err)

	// Create default.yaml
	defaultYAML := `version: "0.2"
metadata:
  name: "apache"
  display_name: "Apache HTTP Server"
  description: "Web server software"
  category: "web-server"
packages:
  - name: "apache2"
    version: "2.4.58"
services:
  - name: "apache"
    service_name: "apache2"
    type: "systemd"
    enabled: true`

	defaultFile := filepath.Join(apacheDir, "default.yaml")
	err = os.WriteFile(defaultFile, []byte(defaultYAML), 0644)
	require.NoError(t, err)

	// Create manager
	manager := NewManager(tempDir)

	// Test loading software
	saidata, err := manager.LoadSoftware("apache")
	require.NoError(t, err)
	assert.NotNil(t, saidata)

	// Verify loaded data
	assert.Equal(t, "0.2", saidata.Version)
	assert.Equal(t, "apache", saidata.Metadata.Name)
	assert.Equal(t, "Apache HTTP Server", saidata.Metadata.DisplayName)
	assert.Equal(t, "web-server", saidata.Metadata.Category)

	// Verify packages
	assert.Len(t, saidata.Packages, 1)
	assert.Equal(t, "apache2", saidata.Packages[0].Name)
	assert.Equal(t, "2.4.58", saidata.Packages[0].Version)

	// Verify services
	assert.Len(t, saidata.Services, 1)
	assert.Equal(t, "apache", saidata.Services[0].Name)
	assert.Equal(t, "apache2", saidata.Services[0].ServiceName)
	assert.Equal(t, "systemd", saidata.Services[0].Type)
}

func TestSaidataManager_NonExistent(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Test loading non-existent software - should generate defaults
	saidata, err := manager.LoadSoftware("nonexistent")
	require.NoError(t, err)
	assert.NotNil(t, saidata)
	
	// Should be generated defaults
	assert.True(t, saidata.IsGenerated)
	assert.Equal(t, "nonexistent", saidata.Metadata.Name)
}

func TestSaidataManager_GenerateDefaults(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir)

	// Test generating defaults for unknown software
	saidata, err := manager.GenerateDefaults("nginx")
	require.NoError(t, err)
	assert.NotNil(t, saidata)

	// Verify generated defaults
	assert.Equal(t, "0.2", saidata.Version)
	assert.Equal(t, "nginx", saidata.Metadata.Name)
	assert.True(t, saidata.IsGenerated)

	// Should have basic defaults
	assert.NotEmpty(t, saidata.Packages)
	assert.Equal(t, "nginx", saidata.Packages[0].Name)

	// Check if services were generated (may be empty depending on implementation)
	if len(saidata.Services) > 0 {
		assert.Equal(t, "nginx", saidata.Services[0].Name)
		assert.Equal(t, "nginx", saidata.Services[0].ServiceName)
	}
}