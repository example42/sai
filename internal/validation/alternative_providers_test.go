package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sai-cli/sai/internal/types"
)

func TestSourceValidation(t *testing.T) {
	t.Run("valid source configuration", func(t *testing.T) {
		source := types.Source{
			Name:        "main",
			URL:         "https://example.com/source.tar.gz",
			BuildSystem: "autotools",
			Version:     "1.0.0",
			BuildDir:    "/tmp/build",
			SourceDir:   "/tmp/build/source",
			InstallPrefix: "/usr/local",
			ConfigureArgs: []string{"--enable-ssl"},
			Prerequisites: []string{"build-essential"},
			Environment: map[string]string{"CC": "gcc"},
			Checksum:    "sha256:abcd1234",
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateSource(&source)
		assert.Empty(t, errors, "Valid source should not have validation errors")
	})

	t.Run("missing required fields", func(t *testing.T) {
		source := types.Source{
			// Missing name and URL
			BuildSystem: "autotools",
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateSource(&source)
		assert.NotEmpty(t, errors, "Source with missing required fields should have validation errors")
		
		// Check for specific error messages
		errorMessages := make([]string, len(errors))
		for i, err := range errors {
			errorMessages[i] = err.Error()
		}
		
		assert.Contains(t, errorMessages, "source name is required")
		assert.Contains(t, errorMessages, "source URL is required")
	})

	t.Run("invalid build system", func(t *testing.T) {
		source := types.Source{
			Name:        "main",
			URL:         "https://example.com/source.tar.gz",
			BuildSystem: "invalid-build-system",
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateSource(&source)
		assert.NotEmpty(t, errors, "Source with invalid build system should have validation errors")
		
		found := false
		for _, err := range errors {
			if err.Error() == "invalid build system: invalid-build-system" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain build system validation error")
	})

	t.Run("invalid URL format", func(t *testing.T) {
		source := types.Source{
			Name:        "main",
			URL:         "not-a-valid-url",
			BuildSystem: "autotools",
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateSource(&source)
		assert.NotEmpty(t, errors, "Source with invalid URL should have validation errors")
		
		found := false
		for _, err := range errors {
			if err.Error() == "invalid URL format: not-a-valid-url" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain URL validation error")
	})

	t.Run("invalid checksum format", func(t *testing.T) {
		source := types.Source{
			Name:        "main",
			URL:         "https://example.com/source.tar.gz",
			BuildSystem: "autotools",
			Checksum:    "invalid-checksum",
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateSource(&source)
		assert.NotEmpty(t, errors, "Source with invalid checksum should have validation errors")
		
		found := false
		for _, err := range errors {
			if err.Error() == "invalid checksum format: invalid-checksum" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain checksum validation error")
	})
}

func TestBinaryValidation(t *testing.T) {
	t.Run("valid binary configuration", func(t *testing.T) {
		binary := types.Binary{
			Name:         "main",
			URL:          "https://example.com/binary.zip",
			Version:      "1.0.0",
			Architecture: "amd64",
			Platform:     "linux",
			Checksum:     "sha256:abcd1234",
			InstallPath:  "/usr/local/bin",
			Executable:   "app",
			Permissions:  "0755",
			Archive: &types.ArchiveConfig{
				Format:      "zip",
				ExtractPath: "/tmp/extract",
			},
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateBinary(&binary)
		assert.Empty(t, errors, "Valid binary should not have validation errors")
	})

	t.Run("missing required fields", func(t *testing.T) {
		binary := types.Binary{
			// Missing name, URL, and executable
			Version: "1.0.0",
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateBinary(&binary)
		assert.NotEmpty(t, errors, "Binary with missing required fields should have validation errors")
		
		errorMessages := make([]string, len(errors))
		for i, err := range errors {
			errorMessages[i] = err.Error()
		}
		
		assert.Contains(t, errorMessages, "binary name is required")
		assert.Contains(t, errorMessages, "binary URL is required")
		assert.Contains(t, errorMessages, "binary executable is required")
	})

	t.Run("invalid URL format", func(t *testing.T) {
		binary := types.Binary{
			Name:       "main",
			URL:        "not-a-valid-url",
			Executable: "app",
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateBinary(&binary)
		assert.NotEmpty(t, errors, "Binary with invalid URL should have validation errors")
		
		found := false
		for _, err := range errors {
			if err.Error() == "invalid URL format: not-a-valid-url" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain URL validation error")
	})

	t.Run("invalid permissions format", func(t *testing.T) {
		binary := types.Binary{
			Name:        "main",
			URL:         "https://example.com/binary",
			Executable:  "app",
			Permissions: "invalid-permissions",
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateBinary(&binary)
		assert.NotEmpty(t, errors, "Binary with invalid permissions should have validation errors")
		
		found := false
		for _, err := range errors {
			if err.Error() == "invalid permissions format: invalid-permissions" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain permissions validation error")
	})

	t.Run("invalid archive format", func(t *testing.T) {
		binary := types.Binary{
			Name:       "main",
			URL:        "https://example.com/binary.zip",
			Executable: "app",
			Archive: &types.ArchiveConfig{
				Format: "invalid-format",
			},
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateBinary(&binary)
		assert.NotEmpty(t, errors, "Binary with invalid archive format should have validation errors")
		
		found := false
		for _, err := range errors {
			if err.Error() == "invalid archive format: invalid-format" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain archive format validation error")
	})
}

func TestScriptValidation(t *testing.T) {
	t.Run("valid script configuration", func(t *testing.T) {
		script := types.Script{
			Name:        "install",
			URL:         "https://example.com/install.sh",
			Version:     "1.0.0",
			Interpreter: "bash",
			Checksum:    "sha256:abcd1234",
			Arguments:   []string{"--verbose"},
			Environment: map[string]string{"DEBUG": "1"},
			WorkingDir:  "/tmp",
			Timeout:     300,
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateScript(&script)
		assert.Empty(t, errors, "Valid script should not have validation errors")
	})

	t.Run("missing required fields", func(t *testing.T) {
		script := types.Script{
			// Missing name and URL
			Version: "1.0.0",
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateScript(&script)
		assert.NotEmpty(t, errors, "Script with missing required fields should have validation errors")
		
		errorMessages := make([]string, len(errors))
		for i, err := range errors {
			errorMessages[i] = err.Error()
		}
		
		assert.Contains(t, errorMessages, "script name is required")
		assert.Contains(t, errorMessages, "script URL is required")
	})

	t.Run("invalid URL format", func(t *testing.T) {
		script := types.Script{
			Name: "install",
			URL:  "not-a-valid-url",
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateScript(&script)
		assert.NotEmpty(t, errors, "Script with invalid URL should have validation errors")
		
		found := false
		for _, err := range errors {
			if err.Error() == "invalid URL format: not-a-valid-url" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain URL validation error")
	})

	t.Run("invalid interpreter", func(t *testing.T) {
		script := types.Script{
			Name:        "install",
			URL:         "https://example.com/install.sh",
			Interpreter: "invalid-interpreter",
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateScript(&script)
		assert.NotEmpty(t, errors, "Script with invalid interpreter should have validation errors")
		
		found := false
		for _, err := range errors {
			if err.Error() == "invalid interpreter: invalid-interpreter" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain interpreter validation error")
	})

	t.Run("invalid timeout", func(t *testing.T) {
		script := types.Script{
			Name:    "install",
			URL:     "https://example.com/install.sh",
			Timeout: -1, // Negative timeout
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateScript(&script)
		assert.NotEmpty(t, errors, "Script with invalid timeout should have validation errors")
		
		found := false
		for _, err := range errors {
			if err.Error() == "timeout must be positive: -1" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain timeout validation error")
	})
}

func TestSaidataValidationWithAlternativeProviders(t *testing.T) {
	t.Run("valid saidata with alternative providers", func(t *testing.T) {
		yamlData := `
version: "0.2"
metadata:
  name: "test-app"
  description: "Test application"
sources:
  - name: "main"
    url: "https://example.com/source.tar.gz"
    build_system: "autotools"
binaries:
  - name: "main"
    url: "https://example.com/binary.zip"
    executable: "app"
scripts:
  - name: "install"
    url: "https://example.com/install.sh"
    interpreter: "bash"
`

		saidata, err := types.LoadSoftwareDataFromYAML([]byte(yamlData))
		require.NoError(t, err)

		validator := NewSaidataValidator()
		errors := validator.ValidateSaidata(saidata)
		assert.Empty(t, errors, "Valid saidata with alternative providers should not have validation errors")
	})

	t.Run("invalid saidata with alternative providers", func(t *testing.T) {
		yamlData := `
version: "0.2"
metadata:
  name: "test-app"
sources:
  - name: ""  # Empty name
    url: "not-a-url"  # Invalid URL
    build_system: "invalid"  # Invalid build system
binaries:
  - name: "main"
    url: "https://example.com/binary.zip"
    # Missing executable
scripts:
  - name: "install"
    # Missing URL
    interpreter: "invalid-interpreter"  # Invalid interpreter
`

		saidata, err := types.LoadSoftwareDataFromYAML([]byte(yamlData))
		require.NoError(t, err)

		validator := NewSaidataValidator()
		errors := validator.ValidateSaidata(saidata)
		assert.NotEmpty(t, errors, "Invalid saidata should have validation errors")
		
		// Should have multiple validation errors
		assert.GreaterOrEqual(t, len(errors), 5, "Should have multiple validation errors")
	})
}

func TestTemplateResolutionFailureScenarios(t *testing.T) {
	t.Run("template resolution with missing source", func(t *testing.T) {
		saidata := &types.SoftwareData{
			// No sources defined
		}

		engine := template.NewTemplateEngine(saidata, "source")
		
		result, err := engine.ExecuteTemplate("{{sai_source(0, 'name')}}")
		assert.Error(t, err, "Template resolution should fail when source doesn't exist")
		assert.Empty(t, result, "Result should be empty on error")
		assert.Contains(t, err.Error(), "source index 0 not found", "Error should indicate missing source")
	})

	t.Run("template resolution with missing binary", func(t *testing.T) {
		saidata := &types.SoftwareData{
			// No binaries defined
		}

		engine := template.NewTemplateEngine(saidata, "binary")
		
		result, err := engine.ExecuteTemplate("{{sai_binary(0, 'name')}}")
		assert.Error(t, err, "Template resolution should fail when binary doesn't exist")
		assert.Empty(t, result, "Result should be empty on error")
		assert.Contains(t, err.Error(), "binary index 0 not found", "Error should indicate missing binary")
	})

	t.Run("template resolution with missing script", func(t *testing.T) {
		saidata := &types.SoftwareData{
			// No scripts defined
		}

		engine := template.NewTemplateEngine(saidata, "script")
		
		result, err := engine.ExecuteTemplate("{{sai_script(0, 'name')}}")
		assert.Error(t, err, "Template resolution should fail when script doesn't exist")
		assert.Empty(t, result, "Result should be empty on error")
		assert.Contains(t, err.Error(), "script index 0 not found", "Error should indicate missing script")
	})

	t.Run("template resolution with invalid field", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{Name: "main", URL: "https://example.com/source.tar.gz", BuildSystem: "autotools"},
			},
		}

		engine := template.NewTemplateEngine(saidata, "source")
		
		result, err := engine.ExecuteTemplate("{{sai_source(0, 'nonexistent_field')}}")
		assert.Error(t, err, "Template resolution should fail for nonexistent field")
		assert.Empty(t, result, "Result should be empty on error")
		assert.Contains(t, err.Error(), "field 'nonexistent_field' not found", "Error should indicate missing field")
	})

	t.Run("graceful degradation when provider actions fail", func(t *testing.T) {
		// Test that when template resolution fails, the provider action is disabled
		saidata := &types.SoftwareData{
			Metadata: types.Metadata{Name: "test-app"},
			// No alternative provider configurations
		}

		engine := template.NewTemplateEngine(saidata, "source")
		
		// This should fail gracefully and not crash
		result, err := engine.ExecuteTemplate("{{sai_source(0, 'url')}}")
		assert.Error(t, err, "Should fail when no sources are defined")
		assert.Empty(t, result, "Result should be empty")
		
		// The error should be informative
		assert.Contains(t, err.Error(), "source", "Error should mention source")
	})
}

func TestSecurityValidation(t *testing.T) {
	t.Run("script execution security validation", func(t *testing.T) {
		script := types.Script{
			Name:        "malicious",
			URL:         "http://malicious.com/script.sh", // HTTP instead of HTTPS
			Interpreter: "bash",
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateScript(&script)
		assert.NotEmpty(t, errors, "Script with HTTP URL should have security validation errors")
		
		found := false
		for _, err := range errors {
			if err.Error() == "insecure URL: HTTP URLs are not allowed for scripts" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain security validation error for HTTP URL")
	})

	t.Run("binary download security validation", func(t *testing.T) {
		binary := types.Binary{
			Name:       "app",
			URL:        "http://example.com/binary", // HTTP instead of HTTPS
			Executable: "app",
			// Missing checksum
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateBinary(&binary)
		assert.NotEmpty(t, errors, "Binary with HTTP URL and no checksum should have security validation errors")
		
		httpError := false
		checksumError := false
		for _, err := range errors {
			if err.Error() == "insecure URL: HTTP URLs are not recommended for binaries" {
				httpError = true
			}
			if err.Error() == "checksum is recommended for binary downloads" {
				checksumError = true
			}
		}
		assert.True(t, httpError, "Should contain HTTP URL warning")
		assert.True(t, checksumError, "Should contain checksum warning")
	})

	t.Run("source download security validation", func(t *testing.T) {
		source := types.Source{
			Name:        "app",
			URL:         "http://example.com/source.tar.gz", // HTTP instead of HTTPS
			BuildSystem: "autotools",
			// Missing checksum
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateSource(&source)
		assert.NotEmpty(t, errors, "Source with HTTP URL and no checksum should have security validation errors")
		
		httpError := false
		checksumError := false
		for _, err := range errors {
			if err.Error() == "insecure URL: HTTP URLs are not recommended for sources" {
				httpError = true
			}
			if err.Error() == "checksum is recommended for source downloads" {
				checksumError = true
			}
		}
		assert.True(t, httpError, "Should contain HTTP URL warning")
		assert.True(t, checksumError, "Should contain checksum warning")
	})
}

func TestRollbackValidation(t *testing.T) {
	t.Run("source rollback validation", func(t *testing.T) {
		source := types.Source{
			Name:        "main",
			URL:         "https://example.com/source.tar.gz",
			BuildSystem: "autotools",
			CustomCommands: &types.SourceCustomCommands{
				Install: "make install",
				// Missing uninstall command
			},
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateSource(&source)
		assert.NotEmpty(t, errors, "Source without uninstall command should have validation warnings")
		
		found := false
		for _, err := range errors {
			if err.Error() == "uninstall command is recommended for rollback capability" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain rollback validation warning")
	})

	t.Run("binary rollback validation", func(t *testing.T) {
		binary := types.Binary{
			Name:       "main",
			URL:        "https://example.com/binary.zip",
			Executable: "app",
			CustomCommands: &types.BinaryCustomCommands{
				Install: "cp binary /usr/local/bin/",
				// Missing uninstall command
			},
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateBinary(&binary)
		assert.NotEmpty(t, errors, "Binary without uninstall command should have validation warnings")
		
		found := false
		for _, err := range errors {
			if err.Error() == "uninstall command is recommended for rollback capability" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain rollback validation warning")
	})

	t.Run("script rollback validation", func(t *testing.T) {
		script := types.Script{
			Name:        "install",
			URL:         "https://example.com/install.sh",
			Interpreter: "bash",
			CustomCommands: &types.ScriptCustomCommands{
				Install: "bash install.sh",
				// Missing uninstall command
			},
		}

		validator := NewSaidataValidator()
		errors := validator.ValidateScript(&script)
		assert.NotEmpty(t, errors, "Script without uninstall command should have validation warnings")
		
		found := false
		for _, err := range errors {
			if err.Error() == "uninstall command is recommended for rollback capability" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain rollback validation warning")
	})
}