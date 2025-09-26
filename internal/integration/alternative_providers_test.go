package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sai-cli/sai/internal/provider"
	"github.com/sai-cli/sai/internal/template"
	"github.com/sai-cli/sai/internal/types"
)

func TestAlternativeProvidersIntegration(t *testing.T) {
	// Skip integration tests in short mode
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("source provider integration", func(t *testing.T) {
		// Load nginx saidata with source configuration
		saidataPath := "../../docs/saidata_samples/ng/nginx/default.yaml"
		if _, err := os.Stat(saidataPath); os.IsNotExist(err) {
			t.Skip("Nginx saidata sample not found")
		}

		data, err := os.ReadFile(saidataPath)
		require.NoError(t, err)

		saidata, err := types.LoadSoftwareDataFromYAML(data)
		require.NoError(t, err)
		require.NotNil(t, saidata)

		// Verify source configuration exists
		require.NotEmpty(t, saidata.Sources)
		source := saidata.Sources[0]
		assert.Equal(t, "main", source.Name)
		assert.NotEmpty(t, source.URL)
		assert.Equal(t, "autotools", source.BuildSystem)

		// Test template resolution
		engine := template.NewTemplateEngine(saidata, "source")
		
		url, err := engine.ExecuteTemplate("{{sai_source(0, 'url')}}")
		require.NoError(t, err)
		assert.Contains(t, url, "nginx")
		assert.Contains(t, url, ".tar.gz")

		buildSystem, err := engine.ExecuteTemplate("{{sai_source(0, 'build_system')}}")
		require.NoError(t, err)
		assert.Equal(t, "autotools", buildSystem)

		// Test provider-specific overrides if they exist
		if providerConfig := saidata.GetProviderConfig("source"); providerConfig != nil && len(providerConfig.Sources) > 0 {
			providerSource := providerConfig.Sources[0]
			if providerSource.BuildDir != "" {
				buildDir, err := engine.ExecuteTemplate("{{sai_source(0, 'build_dir')}}")
				require.NoError(t, err)
				assert.Equal(t, providerSource.BuildDir, buildDir)
			}
		}
	})

	t.Run("binary provider integration", func(t *testing.T) {
		// Load terraform saidata with binary configuration
		saidataPath := "../../docs/saidata_samples/te/terraform/default.yaml"
		if _, err := os.Stat(saidataPath); os.IsNotExist(err) {
			t.Skip("Terraform saidata sample not found")
		}

		data, err := os.ReadFile(saidataPath)
		require.NoError(t, err)

		saidata, err := types.LoadSoftwareDataFromYAML(data)
		require.NoError(t, err)
		require.NotNil(t, saidata)

		// Verify binary configuration exists
		require.NotEmpty(t, saidata.Binaries)
		binary := saidata.Binaries[0]
		assert.Equal(t, "main", binary.Name)
		assert.NotEmpty(t, binary.URL)
		assert.Equal(t, "terraform", binary.Executable)

		// Test template resolution
		engine := template.NewTemplateEngine(saidata, "binary")
		
		url, err := engine.ExecuteTemplate("{{sai_binary(0, 'url')}}")
		require.NoError(t, err)
		assert.Contains(t, url, "terraform")
		assert.Contains(t, url, ".zip")

		executable, err := engine.ExecuteTemplate("{{sai_binary(0, 'executable')}}")
		require.NoError(t, err)
		assert.Equal(t, "terraform", executable)

		// Test archive configuration if present
		if binary.Archive != nil {
			format, err := engine.ExecuteTemplate("{{sai_binary(0, 'archive.format')}}")
			require.NoError(t, err)
			assert.Equal(t, binary.Archive.Format, format)
		}
	})

	t.Run("script provider integration", func(t *testing.T) {
		// Load docker saidata with script configuration
		saidataPath := "../../docs/saidata_samples/do/docker/default.yaml"
		if _, err := os.Stat(saidataPath); os.IsNotExist(err) {
			t.Skip("Docker saidata sample not found")
		}

		data, err := os.ReadFile(saidataPath)
		require.NoError(t, err)

		saidata, err := types.LoadSoftwareDataFromYAML(data)
		require.NoError(t, err)
		require.NotNil(t, saidata)

		// Verify script configuration exists
		require.NotEmpty(t, saidata.Scripts)
		script := saidata.Scripts[0]
		assert.Equal(t, "convenience", script.Name)
		assert.NotEmpty(t, script.URL)
		assert.Equal(t, "bash", script.Interpreter)

		// Test template resolution
		engine := template.NewTemplateEngine(saidata, "script")
		
		url, err := engine.ExecuteTemplate("{{sai_script(0, 'url')}}")
		require.NoError(t, err)
		assert.Contains(t, url, "get.docker.com")

		interpreter, err := engine.ExecuteTemplate("{{sai_script(0, 'interpreter')}}")
		require.NoError(t, err)
		assert.Equal(t, "bash", interpreter)

		// Test environment variables if present
		if len(script.Environment) > 0 {
			for key := range script.Environment {
				envVar, err := engine.ExecuteTemplate("{{sai_script(0, 'environment." + key + "')}}")
				require.NoError(t, err)
				assert.NotEmpty(t, envVar)
				break // Test at least one environment variable
			}
		}
	})
}

func TestProviderDetectionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("alternative provider detection", func(t *testing.T) {
		// Test provider detection for alternative providers
		detector := provider.NewDetector()
		
		// Test source provider detection (requires build tools)
		sourceAvailable := detector.IsProviderAvailable("source")
		t.Logf("Source provider available: %v", sourceAvailable)
		
		// Test binary provider detection (requires download tools)
		binaryAvailable := detector.IsProviderAvailable("binary")
		t.Logf("Binary provider available: %v", binaryAvailable)
		
		// Test script provider detection (requires script interpreters)
		scriptAvailable := detector.IsProviderAvailable("script")
		t.Logf("Script provider available: %v", scriptAvailable)

		// At least one alternative provider should be available on most systems
		assert.True(t, sourceAvailable || binaryAvailable || scriptAvailable,
			"At least one alternative provider should be available")
	})
}

func TestCrossPlatformCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	testCases := []struct {
		name         string
		saidataPath  string
		providerType string
	}{
		{
			name:         "nginx source cross-platform",
			saidataPath:  "../../docs/saidata_samples/ng/nginx/default.yaml",
			providerType: "source",
		},
		{
			name:         "terraform binary cross-platform",
			saidataPath:  "../../docs/saidata_samples/te/terraform/default.yaml",
			providerType: "binary",
		},
		{
			name:         "docker script cross-platform",
			saidataPath:  "../../docs/saidata_samples/do/docker/default.yaml",
			providerType: "script",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := os.Stat(tc.saidataPath); os.IsNotExist(err) {
				t.Skipf("Saidata file %s not found", tc.saidataPath)
			}

			data, err := os.ReadFile(tc.saidataPath)
			require.NoError(t, err)

			saidata, err := types.LoadSoftwareDataFromYAML(data)
			require.NoError(t, err)

			// Test compatibility matrix if present
			if saidata.Compatibility != nil && len(saidata.Compatibility.Matrix) > 0 {
				for _, entry := range saidata.Compatibility.Matrix {
					if entry.Provider == tc.providerType {
						assert.True(t, entry.Supported, 
							"Provider %s should be supported according to compatibility matrix", tc.providerType)
						
						platforms := entry.GetPlatformsAsStrings()
						assert.NotEmpty(t, platforms, 
							"Provider %s should specify supported platforms", tc.providerType)
						
						t.Logf("Provider %s supports platforms: %v", tc.providerType, platforms)
					}
				}
			}

			// Test template resolution works across platforms
			engine := template.NewTemplateEngine(saidata, tc.providerType)
			
			switch tc.providerType {
			case "source":
				if len(saidata.Sources) > 0 {
					url, err := engine.ExecuteTemplate("{{sai_source(0, 'url')}}")
					require.NoError(t, err)
					assert.NotEmpty(t, url)
				}
			case "binary":
				if len(saidata.Binaries) > 0 {
					url, err := engine.ExecuteTemplate("{{sai_binary(0, 'url')}}")
					require.NoError(t, err)
					assert.NotEmpty(t, url)
				}
			case "script":
				if len(saidata.Scripts) > 0 {
					url, err := engine.ExecuteTemplate("{{sai_script(0, 'url')}}")
					require.NoError(t, err)
					assert.NotEmpty(t, url)
				}
			}
		})
	}
}

func TestOSSpecificOverrides(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	testCases := []struct {
		name        string
		basePath    string
		overridePath string
	}{
		{
			name:         "nginx ubuntu override",
			basePath:     "../../docs/saidata_samples/ng/nginx/default.yaml",
			overridePath: "../../docs/saidata_samples/ng/nginx/ubuntu/22.04.yaml",
		},
		{
			name:         "nginx macos override",
			basePath:     "../../docs/saidata_samples/ng/nginx/default.yaml",
			overridePath: "../../docs/saidata_samples/ng/nginx/macos/13.yaml",
		},
		{
			name:         "terraform ubuntu override",
			basePath:     "../../docs/saidata_samples/te/terraform/default.yaml",
			overridePath: "../../docs/saidata_samples/te/terraform/ubuntu/22.04.yaml",
		},
		{
			name:         "terraform macos override",
			basePath:     "../../docs/saidata_samples/te/terraform/default.yaml",
			overridePath: "../../docs/saidata_samples/te/terraform/macos/13.yaml",
		},
		{
			name:         "docker centos override",
			basePath:     "../../docs/saidata_samples/do/docker/default.yaml",
			overridePath: "../../docs/saidata_samples/do/docker/centos/8.yaml",
		},
		{
			name:         "docker windows override",
			basePath:     "../../docs/saidata_samples/do/docker/default.yaml",
			overridePath: "../../docs/saidata_samples/do/docker/windows/11.yaml",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Check if both files exist
			if _, err := os.Stat(tc.basePath); os.IsNotExist(err) {
				t.Skipf("Base saidata file %s not found", tc.basePath)
			}
			if _, err := os.Stat(tc.overridePath); os.IsNotExist(err) {
				t.Skipf("Override saidata file %s not found", tc.overridePath)
			}

			// Load base configuration
			baseData, err := os.ReadFile(tc.basePath)
			require.NoError(t, err)

			baseSaidata, err := types.LoadSoftwareDataFromYAML(baseData)
			require.NoError(t, err)

			// Load override configuration
			overrideData, err := os.ReadFile(tc.overridePath)
			require.NoError(t, err)

			overrideSaidata, err := types.LoadSoftwareDataFromYAML(overrideData)
			require.NoError(t, err)

			// Verify override has some configuration
			hasOverrides := len(overrideSaidata.Sources) > 0 || 
						   len(overrideSaidata.Binaries) > 0 || 
						   len(overrideSaidata.Scripts) > 0 ||
						   len(overrideSaidata.Providers) > 0

			assert.True(t, hasOverrides, "Override file should contain some alternative provider configuration")

			// Test that override configurations are valid
			if len(overrideSaidata.Sources) > 0 {
				for _, source := range overrideSaidata.Sources {
					assert.NotEmpty(t, source.Name, "Source name should not be empty")
					if source.URL != "" {
						assert.NotEmpty(t, source.URL, "Source URL should not be empty if specified")
					}
				}
			}

			if len(overrideSaidata.Binaries) > 0 {
				for _, binary := range overrideSaidata.Binaries {
					assert.NotEmpty(t, binary.Name, "Binary name should not be empty")
					if binary.URL != "" {
						assert.NotEmpty(t, binary.URL, "Binary URL should not be empty if specified")
					}
				}
			}

			if len(overrideSaidata.Scripts) > 0 {
				for _, script := range overrideSaidata.Scripts {
					assert.NotEmpty(t, script.Name, "Script name should not be empty")
					if script.URL != "" {
						assert.NotEmpty(t, script.URL, "Script URL should not be empty if specified")
					}
				}
			}

			t.Logf("Successfully validated OS-specific override: %s", filepath.Base(tc.overridePath))
		})
	}
}