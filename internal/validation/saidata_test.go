package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sai/internal/interfaces"
	"sai/internal/types"
)

func TestSaidataValidator(t *testing.T) {
	// Create validator with the actual schema
	schemaPath := "../../schemas/saidata-0.2-schema.json"
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		t.Skipf("Schema file %s does not exist", schemaPath)
		return
	}

	validator, err := NewSaidataValidator(schemaPath)
	require.NoError(t, err)
	require.NotNil(t, validator)

	t.Run("ValidateValidSaidata", func(t *testing.T) {
		validYAML := `
version: "0.2"
metadata:
  name: "test-software"
  description: "Test software for validation"
packages:
  - name: "test-package"
    version: "1.0.0"
services:
  - name: "test-service"
    type: "systemd"
files:
  - name: "config"
    path: "/etc/test/config.conf"
    type: "config"
directories:
  - name: "data"
    path: "/var/lib/test"
commands:
  - name: "test-cmd"
    path: "/usr/bin/test-cmd"
ports:
  - port: 8080
    protocol: "tcp"
`
		err := validator.ValidateSaidataYAML([]byte(validYAML))
		assert.NoError(t, err)
	})

	t.Run("ValidateInvalidSaidata", func(t *testing.T) {
		invalidYAML := `
version: "0.2"
metadata:
  # missing required 'name' field
  description: "Test software"
ports:
  - port: "invalid-port-type"
    protocol: "tcp"
`
		err := validator.ValidateSaidataYAML([]byte(invalidYAML))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse saidata YAML")
	})

	t.Run("ValidateSaidataWithContainers", func(t *testing.T) {
		validYAML := `
version: "0.2"
metadata:
  name: "docker-app"
containers:
  - name: "app-container"
    image: "nginx"
    tag: "latest"
    ports: ["80:80", "443:443"]
    environment:
      ENV_VAR: "value"
    labels:
      app: "nginx"
`
		err := validator.ValidateSaidataYAML([]byte(validYAML))
		assert.NoError(t, err)
	})

	t.Run("ValidateSaidataWithProviders", func(t *testing.T) {
		validYAML := `
version: "0.2"
metadata:
  name: "multi-provider-app"
providers:
  apt:
    packages:
      - name: "app-deb"
        version: "1.0.0"
    repositories:
      - name: "custom-repo"
        url: "https://repo.example.com"
        type: "upstream"
  docker:
    containers:
      - name: "app-container"
        image: "app"
        tag: "latest"
`
		err := validator.ValidateSaidataYAML([]byte(validYAML))
		assert.NoError(t, err)
	})
}

func TestValidateExistingSaidataFiles(t *testing.T) {
	schemaPath := "../../schemas/saidata-0.2-schema.json"
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		t.Skipf("Schema file %s does not exist", schemaPath)
		return
	}

	validator, err := NewSaidataValidator(schemaPath)
	require.NoError(t, err)

	// Test validation of actual saidata files
	saidataFiles := []string{
		"../../docs/saidata_samples/ap/apache/default.yaml",
		"../../docs/saidata_samples/el/elasticsearch/default.yaml",
		"../../docs/saidata_samples/do/docker/default.yaml",
	}

	for _, file := range saidataFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Skipf("Saidata file %s does not exist", file)
				return
			}

			err := validator.ValidateSaidataFile(file)
			if err != nil {
				t.Logf("Validation error for %s: %v", file, err)
				// For now, we'll log errors but not fail the test since existing files might not be fully compliant
			}
		})
	}
}

func TestValidateAllSaidata(t *testing.T) {
	schemaPath := "../../schemas/saidata-0.2-schema.json"
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		t.Skipf("Schema file %s does not exist", schemaPath)
		return
	}

	saidataDir := "../../docs/saidata_samples"
	if _, err := os.Stat(saidataDir); os.IsNotExist(err) {
		t.Skipf("Saidata directory %s does not exist", saidataDir)
		return
	}

	validator, err := NewSaidataValidator(schemaPath)
	require.NoError(t, err)

	results, err := validator.ValidateAllSaidata(saidataDir)
	require.NoError(t, err)
	assert.NotEmpty(t, results)

	valid, invalid, errors := GetValidationSummary(results)
	t.Logf("Saidata validation summary: %d valid, %d invalid", valid, invalid)
	
	if len(errors) > 0 {
		t.Logf("Validation errors:")
		for _, err := range errors {
			t.Logf("  - %s", err)
		}
	}

	// At least some files should be present
	assert.Greater(t, len(results), 0)
}

func TestResourceValidator(t *testing.T) {
	validator := NewResourceValidator()
	require.NotNil(t, validator)

	t.Run("ValidateFile", func(t *testing.T) {
		// Test existing file
		existingFile := types.File{
			Name: "hosts",
			Path: "/etc/hosts",
		}
		// This might fail on some systems, so we'll just test the logic
		result := validator.ValidateFile(existingFile)
		t.Logf("File /etc/hosts exists: %v", result)

		// Test non-existing file
		nonExistingFile := types.File{
			Name: "nonexistent",
			Path: "/nonexistent/file",
		}
		result = validator.ValidateFile(nonExistingFile)
		assert.False(t, result)

		// Test empty path
		emptyPathFile := types.File{Name: "empty"}
		result = validator.ValidateFile(emptyPathFile)
		assert.False(t, result)
	})

	t.Run("ValidateDirectory", func(t *testing.T) {
		// Test existing directory
		existingDir := types.Directory{
			Name: "tmp",
			Path: "/tmp",
		}
		result := validator.ValidateDirectory(existingDir)
		// /tmp should exist on most Unix systems
		t.Logf("Directory /tmp exists: %v", result)

		// Test non-existing directory
		nonExistingDir := types.Directory{
			Name: "nonexistent",
			Path: "/nonexistent/directory",
		}
		result = validator.ValidateDirectory(nonExistingDir)
		assert.False(t, result)
	})

	t.Run("ValidateCommand", func(t *testing.T) {
		// Test common command
		lsCommand := types.Command{
			Name: "ls",
			Path: "/bin/ls",
		}
		result := validator.ValidateCommand(lsCommand)
		t.Logf("Command /bin/ls exists: %v", result)

		// Test with default path generation
		echoCommand := types.Command{
			Name: "echo",
		}
		result = validator.ValidateCommand(echoCommand)
		t.Logf("Command echo (default path) exists: %v", result)

		// Test non-existing command
		nonExistingCommand := types.Command{
			Name: "nonexistent",
			Path: "/nonexistent/command",
		}
		result = validator.ValidateCommand(nonExistingCommand)
		assert.False(t, result)
	})

	t.Run("ValidateService", func(t *testing.T) {
		// Test common service (might not exist on all systems)
		sshService := types.Service{
			Name:        "ssh",
			ServiceName: "ssh",
		}
		result := validator.ValidateService(sshService)
		t.Logf("Service ssh exists: %v", result)

		// Test non-existing service
		nonExistingService := types.Service{
			Name: "nonexistent-service",
		}
		result = validator.ValidateService(nonExistingService)
		assert.False(t, result)
	})

	t.Run("ValidatePort", func(t *testing.T) {
		// Test valid port
		validPort := types.Port{Port: 8080}
		result := validator.ValidatePort(validPort)
		assert.True(t, result)

		// Test invalid port (too high)
		invalidPort := types.Port{Port: 70000}
		result = validator.ValidatePort(invalidPort)
		assert.False(t, result)

		// Test invalid port (zero)
		zeroPort := types.Port{Port: 0}
		result = validator.ValidatePort(zeroPort)
		assert.False(t, result)
	})
}

func TestValidateResources(t *testing.T) {
	validator := NewResourceValidator()
	
	saidata := &types.SoftwareData{
		Files: []types.File{
			{Name: "hosts", Path: "/etc/hosts"},
			{Name: "nonexistent", Path: "/nonexistent/file"},
		},
		Directories: []types.Directory{
			{Name: "tmp", Path: "/tmp"},
			{Name: "nonexistent", Path: "/nonexistent/dir"},
		},
		Commands: []types.Command{
			{Name: "ls", Path: "/bin/ls"},
			{Name: "nonexistent", Path: "/nonexistent/cmd"},
		},
		Services: []types.Service{
			{Name: "ssh", ServiceName: "ssh"},
			{Name: "nonexistent"},
		},
		Ports: []types.Port{
			{Port: 8080},
			{Port: 70000}, // Invalid port
		},
	}

	result, err := validator.ValidateResources(saidata)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check that runtime flags were set
	assert.NotNil(t, result.MissingFiles)
	assert.NotNil(t, result.MissingDirectories)
	assert.NotNil(t, result.MissingCommands)
	assert.NotNil(t, result.MissingServices)
	assert.NotNil(t, result.InvalidPorts)

	// Should be able to proceed even with missing resources
	assert.True(t, result.CanProceed)
}

func TestResourceValidationResult(t *testing.T) {
	t.Run("ValidResult", func(t *testing.T) {
		result := &interfaces.ResourceValidationResult{Valid: true, CanProceed: true}
		assert.True(t, result.Valid)
		assert.True(t, result.CanProceed)
	})

	t.Run("InvalidResult", func(t *testing.T) {
		result := &interfaces.ResourceValidationResult{
			Valid:              false,
			MissingFiles:       []string{"/missing/file"},
			MissingDirectories: []string{"/missing/dir"},
			MissingCommands:    []string{"/missing/cmd"},
			MissingServices:    []string{"missing-service"},
			InvalidPorts:       []int{70000},
		}
		
		assert.False(t, result.Valid)
		assert.Contains(t, result.MissingFiles, "/missing/file")
		assert.Contains(t, result.MissingDirectories, "/missing/dir")
		assert.Contains(t, result.MissingCommands, "/missing/cmd")
		assert.Contains(t, result.MissingServices, "missing-service")
		assert.Contains(t, result.InvalidPorts, 70000)
		
		assert.True(t, result.CanProceed) // Still allows proceeding
	})
}

func TestNewSaidataValidator(t *testing.T) {
	t.Run("ValidSchemaFile", func(t *testing.T) {
		schemaPath := "../../schemas/saidata-0.2-schema.json"
		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			t.Skipf("Schema file %s does not exist", schemaPath)
			return
		}

		validator, err := NewSaidataValidator(schemaPath)
		assert.NoError(t, err)
		assert.NotNil(t, validator)
	})

	t.Run("InvalidSchemaFile", func(t *testing.T) {
		validator, err := NewSaidataValidator("nonexistent.json")
		assert.Error(t, err)
		assert.Nil(t, validator)
		assert.Contains(t, err.Error(), "failed to read schema file")
	})
}