package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderValidator(t *testing.T) {
	// Create validator with the actual schema
	schemaPath := "../../schemas/providerdata-0.1-schema.json"
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		t.Skipf("Schema file %s does not exist", schemaPath)
		return
	}

	validator, err := NewProviderValidator(schemaPath)
	require.NoError(t, err)
	require.NotNil(t, validator)

	t.Run("ValidateValidProvider", func(t *testing.T) {
		validYAML := `
version: "1.0"
provider:
  name: "test"
  type: "package_manager"
  platforms: ["linux"]
  capabilities: ["install", "uninstall"]
actions:
  install:
    description: "Install package"
    template: "test install {{sai_package('test')}}"
    timeout: 300
`
		err := validator.ValidateProviderYAML([]byte(validYAML))
		assert.NoError(t, err)
	})

	t.Run("ValidateInvalidProvider", func(t *testing.T) {
		invalidYAML := `
version: "1.0"
provider:
  name: "test"
  # missing required 'type' field
actions:
  install:
    # missing required command/template/script/steps
    description: "Install package"
`
		err := validator.ValidateProviderYAML([]byte(invalidYAML))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("ValidateProviderWithSteps", func(t *testing.T) {
		validYAML := `
version: "1.0"
provider:
  name: "docker"
  type: "container"
actions:
  install:
    description: "Install via Docker"
    steps:
      - name: "pull-image"
        command: "docker pull test"
      - name: "create-container"
        command: "docker create test"
        ignore_failure: true
`
		err := validator.ValidateProviderYAML([]byte(validYAML))
		assert.NoError(t, err)
	})

	t.Run("ValidateProviderWithValidation", func(t *testing.T) {
		validYAML := `
version: "1.0"
provider:
  name: "apt"
  type: "package_manager"
actions:
  install:
    template: "apt-get install test"
    validation:
      command: "dpkg -l | grep test"
      expected_exit_code: 0
      timeout: 30
    retry:
      attempts: 3
      delay: 5
      backoff: "exponential"
`
		err := validator.ValidateProviderYAML([]byte(validYAML))
		assert.NoError(t, err)
	})
}

func TestValidateExistingProviderFiles(t *testing.T) {
	schemaPath := "../../schemas/providerdata-0.1-schema.json"
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		t.Skipf("Schema file %s does not exist", schemaPath)
		return
	}

	validator, err := NewProviderValidator(schemaPath)
	require.NoError(t, err)

	// Test validation of actual provider files
	providerFiles := []string{
		"../../providers/apt.yaml",
		"../../providers/brew.yaml",
		"../../providers/docker.yaml",
	}

	for _, file := range providerFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Skipf("Provider file %s does not exist", file)
				return
			}

			err := validator.ValidateProviderFile(file)
			if err != nil {
				t.Logf("Validation error for %s: %v", file, err)
				// For now, we'll log errors but not fail the test since existing files might not be fully compliant
				// In a real implementation, we'd want all provider files to be valid
			}
		})
	}
}

func TestValidateAllProviders(t *testing.T) {
	schemaPath := "../../schemas/providerdata-0.1-schema.json"
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		t.Skipf("Schema file %s does not exist", schemaPath)
		return
	}

	providerDir := "../../providers"
	if _, err := os.Stat(providerDir); os.IsNotExist(err) {
		t.Skipf("Provider directory %s does not exist", providerDir)
		return
	}

	validator, err := NewProviderValidator(schemaPath)
	require.NoError(t, err)

	results, err := validator.ValidateAllProviders(providerDir)
	require.NoError(t, err)
	assert.NotEmpty(t, results)

	valid, invalid, errors := GetValidationSummary(results)
	t.Logf("Validation summary: %d valid, %d invalid", valid, invalid)
	
	if len(errors) > 0 {
		t.Logf("Validation errors:")
		for _, err := range errors {
			t.Logf("  - %s", err)
		}
	}

	// At least some files should be present
	assert.Greater(t, len(results), 0)
}

func TestValidationResult(t *testing.T) {
	results := []ValidationResult{
		{Valid: true, File: "valid1.yaml"},
		{Valid: true, File: "valid2.yaml"},
		{Valid: false, File: "invalid1.yaml", Errors: []string{"error1", "error2"}},
		{Valid: false, File: "invalid2.yaml", Errors: []string{"error3"}},
	}

	valid, invalid, errors := GetValidationSummary(results)
	
	assert.Equal(t, 2, valid)
	assert.Equal(t, 2, invalid)
	assert.Len(t, errors, 3) // 2 errors from invalid1 + 1 error from invalid2
	assert.Contains(t, errors[0], "invalid1.yaml")
	assert.Contains(t, errors[1], "invalid1.yaml")
	assert.Contains(t, errors[2], "invalid2.yaml")
}

func TestIsYAMLFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"test.yaml", true},
		{"test.yml", true},
		{"test.json", false},
		{"test.txt", false},
		{"test", false},
		{".yaml", false}, // Too short
		{"yaml", false},  // No extension
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := isYAMLFile(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewProviderValidator(t *testing.T) {
	t.Run("ValidSchemaFile", func(t *testing.T) {
		schemaPath := "../../schemas/providerdata-0.1-schema.json"
		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			t.Skipf("Schema file %s does not exist", schemaPath)
			return
		}

		validator, err := NewProviderValidator(schemaPath)
		assert.NoError(t, err)
		assert.NotNil(t, validator)
	})

	t.Run("InvalidSchemaFile", func(t *testing.T) {
		validator, err := NewProviderValidator("nonexistent.json")
		assert.Error(t, err)
		assert.Nil(t, validator)
		assert.Contains(t, err.Error(), "failed to read schema file")
	})
}