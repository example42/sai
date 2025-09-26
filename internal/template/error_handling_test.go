package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sai-cli/sai/internal/types"
)

func TestTemplateErrorHandling(t *testing.T) {
	t.Run("graceful degradation on missing saidata", func(t *testing.T) {
		// Test with nil saidata
		engine := NewTemplateEngine(nil, "source")
		
		result, err := engine.ExecuteTemplate("{{sai_source(0, 'name')}}")
		assert.Error(t, err, "Should error with nil saidata")
		assert.Empty(t, result, "Result should be empty")
		assert.Contains(t, err.Error(), "saidata is nil", "Error should indicate nil saidata")
	})

	t.Run("graceful degradation on empty saidata", func(t *testing.T) {
		saidata := &types.SoftwareData{}
		engine := NewTemplateEngine(saidata, "source")
		
		result, err := engine.ExecuteTemplate("{{sai_source(0, 'name')}}")
		assert.Error(t, err, "Should error with empty saidata")
		assert.Empty(t, result, "Result should be empty")
		assert.Contains(t, err.Error(), "source index 0 not found", "Error should indicate missing source")
	})

	t.Run("invalid template syntax", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{Name: "main", URL: "https://example.com/source.tar.gz", BuildSystem: "autotools"},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		// Invalid template syntax
		result, err := engine.ExecuteTemplate("{{sai_source(0, 'name'")
		assert.Error(t, err, "Should error with invalid template syntax")
		assert.Empty(t, result, "Result should be empty")
		assert.Contains(t, err.Error(), "template", "Error should mention template")
	})

	t.Run("function with wrong number of arguments", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{Name: "main", URL: "https://example.com/source.tar.gz", BuildSystem: "autotools"},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		// Too few arguments
		result, err := engine.ExecuteTemplate("{{sai_source(0)}}")
		assert.Error(t, err, "Should error with wrong number of arguments")
		assert.Empty(t, result, "Result should be empty")
		
		// Too many arguments
		result, err = engine.ExecuteTemplate("{{sai_source(0, 'name', 'extra')}}")
		assert.Error(t, err, "Should error with too many arguments")
		assert.Empty(t, result, "Result should be empty")
	})

	t.Run("function with wrong argument types", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{Name: "main", URL: "https://example.com/source.tar.gz", BuildSystem: "autotools"},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		// String instead of int for index
		result, err := engine.ExecuteTemplate("{{sai_source('invalid', 'name')}}")
		assert.Error(t, err, "Should error with wrong argument type")
		assert.Empty(t, result, "Result should be empty")
		assert.Contains(t, err.Error(), "invalid index", "Error should mention invalid index")
	})

	t.Run("out of bounds index", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{Name: "main", URL: "https://example.com/source.tar.gz", BuildSystem: "autotools"},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		// Index out of bounds
		result, err := engine.ExecuteTemplate("{{sai_source(99, 'name')}}")
		assert.Error(t, err, "Should error with out of bounds index")
		assert.Empty(t, result, "Result should be empty")
		assert.Contains(t, err.Error(), "source index 99 not found", "Error should mention out of bounds index")
	})

	t.Run("negative index", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{Name: "main", URL: "https://example.com/source.tar.gz", BuildSystem: "autotools"},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		// Negative index
		result, err := engine.ExecuteTemplate("{{sai_source(-1, 'name')}}")
		assert.Error(t, err, "Should error with negative index")
		assert.Empty(t, result, "Result should be empty")
		assert.Contains(t, err.Error(), "invalid index", "Error should mention invalid index")
	})

	t.Run("nonexistent field", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{Name: "main", URL: "https://example.com/source.tar.gz", BuildSystem: "autotools"},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		result, err := engine.ExecuteTemplate("{{sai_source(0, 'nonexistent_field')}}")
		assert.Error(t, err, "Should error with nonexistent field")
		assert.Empty(t, result, "Result should be empty")
		assert.Contains(t, err.Error(), "field 'nonexistent_field' not found", "Error should mention missing field")
	})

	t.Run("nested field access on nil", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{
					Name:        "main",
					URL:         "https://example.com/source.tar.gz",
					BuildSystem: "autotools",
					// CustomCommands is nil
				},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		result, err := engine.ExecuteTemplate("{{sai_source(0, 'custom_commands.download')}}")
		assert.Error(t, err, "Should error when accessing nested field on nil")
		assert.Empty(t, result, "Result should be empty")
		assert.Contains(t, err.Error(), "custom_commands is nil", "Error should mention nil custom_commands")
	})

	t.Run("environment variable not found", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{
					Name:        "main",
					URL:         "https://example.com/source.tar.gz",
					BuildSystem: "autotools",
					Environment: map[string]string{
						"CC": "gcc",
					},
				},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		result, err := engine.ExecuteTemplate("{{sai_source(0, 'environment.NONEXISTENT')}}")
		assert.Error(t, err, "Should error when environment variable not found")
		assert.Empty(t, result, "Result should be empty")
		assert.Contains(t, err.Error(), "environment variable 'NONEXISTENT' not found", "Error should mention missing env var")
	})
}

func TestProviderResolutionErrors(t *testing.T) {
	t.Run("provider config not found", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{Name: "main", URL: "https://example.com/source.tar.gz", BuildSystem: "autotools"},
			},
			// No provider configs
		}
		engine := NewTemplateEngine(saidata, "nonexistent-provider")
		
		// Should fall back to default sources
		result, err := engine.ExecuteTemplate("{{sai_source(0, 'name')}}")
		require.NoError(t, err, "Should fall back to default when provider not found")
		assert.Equal(t, "main", result)
	})

	t.Run("provider config exists but source not found", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{Name: "main", URL: "https://example.com/source.tar.gz", BuildSystem: "autotools"},
			},
			Providers: map[string]types.ProviderConfig{
				"source": {
					// Empty sources in provider config
				},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		// Should fall back to default sources
		result, err := engine.ExecuteTemplate("{{sai_source(0, 'name')}}")
		require.NoError(t, err, "Should fall back to default when provider source not found")
		assert.Equal(t, "main", result)
	})

	t.Run("provider override with missing field falls back to default", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{
					Name:        "main",
					URL:         "https://example.com/source.tar.gz",
					BuildSystem: "autotools",
					BuildDir:    "/tmp/default-build",
				},
			},
			Providers: map[string]types.ProviderConfig{
				"source": {
					Sources: []types.Source{
						{
							Name: "main",
							URL:  "https://example.com/provider-source.tar.gz",
							// BuildDir not specified in provider override
						},
					},
				},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		// URL should come from provider override
		url, err := engine.ExecuteTemplate("{{sai_source(0, 'url')}}")
		require.NoError(t, err)
		assert.Equal(t, "https://example.com/provider-source.tar.gz", url)
		
		// BuildDir should fall back to default
		buildDir, err := engine.ExecuteTemplate("{{sai_source(0, 'build_dir')}}")
		require.NoError(t, err)
		assert.Equal(t, "/tmp/default-build", buildDir)
	})
}

func TestComplexTemplateErrors(t *testing.T) {
	t.Run("template with multiple functions", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{Name: "main", URL: "https://example.com/source.tar.gz", BuildSystem: "autotools"},
			},
			Binaries: []types.Binary{
				{Name: "main", URL: "https://example.com/binary.zip", Executable: "app"},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		// One function succeeds, one fails
		template := "{{sai_source(0, 'name')}} {{sai_binary(99, 'name')}}"
		result, err := engine.ExecuteTemplate(template)
		assert.Error(t, err, "Should error when one function fails")
		assert.Empty(t, result, "Result should be empty")
		assert.Contains(t, err.Error(), "binary index 99 not found", "Error should mention the failing function")
	})

	t.Run("template with conditional logic", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{
					Name:        "main",
					URL:         "https://example.com/source.tar.gz",
					BuildSystem: "autotools",
					Version:     "1.0.0",
				},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		// Template with conditional that might fail
		template := "{{if sai_source(0, 'version')}}Version: {{sai_source(0, 'version')}}{{else}}No version{{end}}"
		result, err := engine.ExecuteTemplate(template)
		require.NoError(t, err, "Conditional template should work")
		assert.Equal(t, "Version: 1.0.0", result)
		
		// Template with conditional that accesses nonexistent field
		template = "{{if sai_source(0, 'nonexistent')}}{{sai_source(0, 'nonexistent')}}{{else}}Default{{end}}"
		result, err = engine.ExecuteTemplate(template)
		assert.Error(t, err, "Should error even in conditional when field doesn't exist")
		assert.Empty(t, result, "Result should be empty")
	})

	t.Run("template with loops", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{Name: "src1", URL: "https://example.com/src1.tar.gz", BuildSystem: "autotools"},
				{Name: "src2", URL: "https://example.com/src2.tar.gz", BuildSystem: "cmake"},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		// Template that tries to loop over sources but uses wrong index
		template := "{{range $i, $src := .Sources}}{{sai_source($i, 'name')}} {{end}}"
		result, err := engine.ExecuteTemplate(template)
		assert.Error(t, err, "Should error when using range variable as function argument")
		assert.Empty(t, result, "Result should be empty")
	})
}

func TestErrorRecovery(t *testing.T) {
	t.Run("engine continues working after error", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{Name: "main", URL: "https://example.com/source.tar.gz", BuildSystem: "autotools"},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		// First template fails
		result, err := engine.ExecuteTemplate("{{sai_source(99, 'name')}}")
		assert.Error(t, err, "First template should fail")
		assert.Empty(t, result, "Result should be empty")
		
		// Second template should work
		result, err = engine.ExecuteTemplate("{{sai_source(0, 'name')}}")
		require.NoError(t, err, "Second template should work after first failed")
		assert.Equal(t, "main", result)
	})

	t.Run("partial template execution", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{Name: "main", URL: "https://example.com/source.tar.gz", BuildSystem: "autotools"},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		// Template that starts with valid content but then fails
		template := "Name: {{sai_source(0, 'name')}} URL: {{sai_source(99, 'url')}}"
		result, err := engine.ExecuteTemplate(template)
		assert.Error(t, err, "Template should fail")
		assert.Empty(t, result, "Result should be empty on error")
		// Template execution should be atomic - either all succeeds or all fails
	})
}

func TestDetailedErrorMessages(t *testing.T) {
	t.Run("error messages contain context", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{Name: "main", URL: "https://example.com/source.tar.gz", BuildSystem: "autotools"},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		_, err := engine.ExecuteTemplate("{{sai_source(5, 'name')}}")
		require.Error(t, err)
		
		errorMsg := err.Error()
		assert.Contains(t, errorMsg, "source", "Error should mention source")
		assert.Contains(t, errorMsg, "index 5", "Error should mention the specific index")
		assert.Contains(t, errorMsg, "not found", "Error should indicate not found")
	})

	t.Run("nested field error messages", func(t *testing.T) {
		saidata := &types.SoftwareData{
			Sources: []types.Source{
				{
					Name:        "main",
					URL:         "https://example.com/source.tar.gz",
					BuildSystem: "autotools",
					Environment: map[string]string{"CC": "gcc"},
				},
			},
		}
		engine := NewTemplateEngine(saidata, "source")
		
		_, err := engine.ExecuteTemplate("{{sai_source(0, 'environment.MISSING_VAR')}}")
		require.Error(t, err)
		
		errorMsg := err.Error()
		assert.Contains(t, errorMsg, "environment variable", "Error should mention environment variable")
		assert.Contains(t, errorMsg, "MISSING_VAR", "Error should mention the specific variable name")
		assert.Contains(t, errorMsg, "not found", "Error should indicate not found")
	})
}