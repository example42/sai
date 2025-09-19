package types

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadProviderFromYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		wantErr  bool
		validate func(t *testing.T, provider *ProviderData)
	}{
		{
			name: "valid basic provider",
			yamlData: `
version: "1.0"
provider:
  name: "test"
  type: "package_manager"
actions:
  install:
    template: "test install {{sai_package('test')}}"
`,
			wantErr: false,
			validate: func(t *testing.T, provider *ProviderData) {
				assert.Equal(t, "1.0", provider.Version)
				assert.Equal(t, "test", provider.Provider.Name)
				assert.Equal(t, "package_manager", provider.Provider.Type)
				assert.Equal(t, 50, provider.Provider.Priority) // Default priority
				assert.Contains(t, provider.Actions, "install")
				assert.Equal(t, "test install {{sai_package('test')}}", provider.Actions["install"].Template)
				assert.Equal(t, 300, provider.Actions["install"].Timeout) // Default timeout
			},
		},
		{
			name: "provider with all fields",
			yamlData: `
version: "1.0"
provider:
  name: "apt"
  display_name: "Advanced Package Tool"
  description: "Package manager for Debian and Ubuntu"
  type: "package_manager"
  platforms: ["debian", "ubuntu"]
  capabilities: ["install", "uninstall", "upgrade"]
  priority: 90
  executable: "apt-get"
actions:
  install:
    description: "Install packages"
    template: "apt-get install -y {{sai_package('apt')}}"
    timeout: 600
    requires_root: true
    validation:
      command: "dpkg -l | grep {{sai_package('apt')}}"
      expected_exit_code: 0
      timeout: 30
    rollback: "apt-get remove -y {{sai_package('apt')}}"
    retry:
      attempts: 3
      delay: 5
      backoff: "exponential"
`,
			wantErr: false,
			validate: func(t *testing.T, provider *ProviderData) {
				assert.Equal(t, "apt", provider.Provider.Name)
				assert.Equal(t, "Advanced Package Tool", provider.Provider.DisplayName)
				assert.Equal(t, 90, provider.Provider.Priority)
				assert.Equal(t, "apt-get", provider.Provider.Executable)
				assert.Equal(t, []string{"debian", "ubuntu"}, provider.Provider.Platforms)
				
				install := provider.Actions["install"]
				assert.Equal(t, "Install packages", install.Description)
				assert.True(t, install.RequiresRoot)
				assert.Equal(t, 600, install.Timeout)
				assert.NotNil(t, install.Validation)
				assert.Equal(t, 30, install.Validation.Timeout)
				assert.NotNil(t, install.Retry)
				assert.Equal(t, 3, install.Retry.Attempts)
				assert.Equal(t, "exponential", install.Retry.Backoff)
			},
		},
		{
			name: "provider with steps",
			yamlData: `
version: "1.0"
provider:
  name: "docker"
  type: "container"
actions:
  install:
    description: "Install via Docker"
    steps:
      - name: "pull-image"
        command: "docker pull {{sai_container('image')}}"
      - name: "create-container"
        command: "docker create --name {{sai_container('name')}}"
        ignore_failure: true
`,
			wantErr: false,
			validate: func(t *testing.T, provider *ProviderData) {
				install := provider.Actions["install"]
				assert.True(t, install.HasSteps())
				assert.Len(t, install.Steps, 2)
				assert.Equal(t, "pull-image", install.Steps[0].Name)
				assert.False(t, install.Steps[0].IgnoreFailure)
				assert.True(t, install.Steps[1].IgnoreFailure)
			},
		},
		{
			name: "invalid YAML",
			yamlData: `
version: "1.0"
provider:
  name: "test"
  type: invalid_type
actions:
  install:
    - invalid structure
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := LoadProviderFromYAML([]byte(tt.yamlData))
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			require.NotNil(t, provider)
			
			if tt.validate != nil {
				tt.validate(t, provider)
			}
		})
	}
}

func TestLoadExistingProviderFiles(t *testing.T) {
	// Test loading actual provider files from the providers directory
	providerFiles := []string{
		"../../providers/apt.yaml",
		"../../providers/brew.yaml", 
		"../../providers/docker.yaml",
	}

	for _, file := range providerFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			// Check if file exists
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Skipf("Provider file %s does not exist", file)
				return
			}

			data, err := os.ReadFile(file)
			require.NoError(t, err)

			provider, err := LoadProviderFromYAML(data)
			require.NoError(t, err)
			require.NotNil(t, provider)

			// Basic validation
			assert.NotEmpty(t, provider.Version)
			assert.NotEmpty(t, provider.Provider.Name)
			assert.NotEmpty(t, provider.Provider.Type)
			assert.NotEmpty(t, provider.Actions)

			// Validate JSON conversion (for schema validation)
			jsonData, err := provider.ToJSON()
			require.NoError(t, err)
			assert.NotEmpty(t, jsonData)

			// Ensure JSON is valid
			var jsonObj map[string]interface{}
			err = json.Unmarshal(jsonData, &jsonObj)
			require.NoError(t, err)
		})
	}
}

func TestActionMethods(t *testing.T) {
	t.Run("GetTimeout", func(t *testing.T) {
		action := Action{Timeout: 120}
		assert.Equal(t, 120*time.Second, action.GetTimeout())

		actionDefault := Action{}
		assert.Equal(t, 300*time.Second, actionDefault.GetTimeout())
	})

	t.Run("HasSteps", func(t *testing.T) {
		actionWithSteps := Action{Steps: []Step{{Command: "test"}}}
		assert.True(t, actionWithSteps.HasSteps())

		actionWithoutSteps := Action{Template: "test"}
		assert.False(t, actionWithoutSteps.HasSteps())
	})

	t.Run("GetCommand", func(t *testing.T) {
		actionTemplate := Action{Template: "template cmd"}
		assert.Equal(t, "template cmd", actionTemplate.GetCommand())

		actionCommand := Action{Command: "static cmd"}
		assert.Equal(t, "static cmd", actionCommand.GetCommand())

		actionScript := Action{Script: "script content"}
		assert.Equal(t, "script content", actionScript.GetCommand())

		actionEmpty := Action{}
		assert.Equal(t, "", actionEmpty.GetCommand())
	})

	t.Run("IsValid", func(t *testing.T) {
		validActions := []Action{
			{Template: "test"},
			{Command: "test"},
			{Script: "test"},
			{Steps: []Step{{Command: "test"}}},
		}

		for _, action := range validActions {
			assert.True(t, action.IsValid())
		}

		invalidAction := Action{}
		assert.False(t, invalidAction.IsValid())
	})
}

func TestPortMapping(t *testing.T) {
	t.Run("GetPortAsInt with integer", func(t *testing.T) {
		mapping := PortMapping{Port: 8080}
		port, err := mapping.GetPortAsInt()
		require.NoError(t, err)
		assert.Equal(t, 8080, port)
	})

	t.Run("GetPortAsInt with string", func(t *testing.T) {
		mapping := PortMapping{Port: "9090"}
		port, err := mapping.GetPortAsInt()
		require.NoError(t, err)
		assert.Equal(t, 9090, port)
	})

	t.Run("GetPortAsInt with invalid string", func(t *testing.T) {
		mapping := PortMapping{Port: "invalid"}
		_, err := mapping.GetPortAsInt()
		assert.Error(t, err)
	})

	t.Run("GetPortAsInt with invalid type", func(t *testing.T) {
		mapping := PortMapping{Port: []string{"invalid"}}
		_, err := mapping.GetPortAsInt()
		assert.Error(t, err)
	})
}

func TestVariableMapping(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"string value", "test", "test"},
		{"int value", 42, "42"},
		{"bool value true", true, "true"},
		{"bool value false", false, "false"},
		{"float value", 3.14, "3.14"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := VariableMapping{Value: tt.value}
			result := mapping.GetValueAsString()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProviderDataToJSON(t *testing.T) {
	provider := &ProviderData{
		Version: "1.0",
		Provider: ProviderInfo{
			Name: "test",
			Type: "package_manager",
		},
		Actions: map[string]Action{
			"install": {
				Template: "test install",
				Timeout:  300,
			},
		},
	}

	jsonData, err := provider.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verify JSON structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err)

	assert.Equal(t, "1.0", result["version"])
	assert.Contains(t, result, "provider")
	assert.Contains(t, result, "actions")
}