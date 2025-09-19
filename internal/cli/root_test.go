package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sai/internal/config"
)

func TestRootCommand(t *testing.T) {
	// Test that the root command can be created and has the expected properties
	assert.Equal(t, "sai", rootCmd.Use)
	assert.Equal(t, "SAI - Software Action Interface", rootCmd.Short)
	assert.Equal(t, "0.1.0", rootCmd.Version)
	assert.True(t, rootCmd.SilenceUsage)
	assert.True(t, rootCmd.SilenceErrors)
	
	// Test that all expected flags are present
	flags := rootCmd.PersistentFlags()
	assert.NotNil(t, flags.Lookup("config"))
	assert.NotNil(t, flags.Lookup("provider"))
	assert.NotNil(t, flags.Lookup("verbose"))
	assert.NotNil(t, flags.Lookup("dry-run"))
	assert.NotNil(t, flags.Lookup("yes"))
	assert.NotNil(t, flags.Lookup("quiet"))
	assert.NotNil(t, flags.Lookup("json"))
	
	// Test flag shortcuts
	assert.Equal(t, "c", flags.Lookup("config").Shorthand)
	assert.Equal(t, "p", flags.Lookup("provider").Shorthand)
	assert.Equal(t, "v", flags.Lookup("verbose").Shorthand)
	assert.Equal(t, "y", flags.Lookup("yes").Shorthand)
	assert.Equal(t, "q", flags.Lookup("quiet").Shorthand)
}

func TestValidateFlags(t *testing.T) {
	tests := []struct {
		name         string
		providerFlag string
		configFlag   string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "valid provider",
			providerFlag: "apt",
			wantErr:      false,
		},
		{
			name:         "invalid provider",
			providerFlag: "invalid-provider",
			wantErr:      true,
			errMsg:       "invalid provider 'invalid-provider'",
		},
		{
			name:         "empty provider",
			providerFlag: "",
			wantErr:      false,
		},
		{
			name:       "non-existent config file",
			configFlag: "/non/existent/config.yaml",
			wantErr:    true,
			errMsg:     "config file '/non/existent/config.yaml' does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global variables
			provider = tt.providerFlag
			cfgFile = tt.configFlag

			err := ValidateFlags()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetGlobalFlags(t *testing.T) {
	// Set some test values
	cfgFile = "test-config.yaml"
	provider = "apt"
	verbose = true
	dryRun = true
	yes = false
	quiet = false
	jsonOutput = true

	flags := GetGlobalFlags()

	assert.Equal(t, "test-config.yaml", flags.Config)
	assert.Equal(t, "apt", flags.Provider)
	assert.True(t, flags.Verbose)
	assert.True(t, flags.DryRun)
	assert.False(t, flags.Yes)
	assert.False(t, flags.Quiet)
	assert.True(t, flags.JSONOutput)
}

func TestConfigFileDiscovery(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "sai-test-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write test config
	_, err = tmpFile.WriteString(`
saidata_repository: "https://github.com/test/saidata.git"
default_provider: "apt"
log_level: "info"
`)
	require.NoError(t, err)
	tmpFile.Close()

	// Test config loading
	cfgFile = tmpFile.Name()
	err = initializeConfig()
	assert.NoError(t, err)
	assert.NotNil(t, globalConfig)
	assert.Equal(t, "apt", globalConfig.DefaultProvider)
}

func TestApplyFlagOverrides(t *testing.T) {
	// Initialize with default config
	globalConfig = &config.Config{
		DefaultProvider: "brew",
		Confirmations: config.ConfirmationConfig{
			Install:       true,
			Uninstall:     true,
			Upgrade:       true,
			SystemChanges: true,
			ServiceOps:    true,
		},
		Output: config.OutputConfig{
			ShowCommands:  false,
			ShowExitCodes: false,
		},
	}

	// Set flags
	provider = "apt"
	yes = true
	verbose = true

	applyFlagOverrides()

	// Check overrides were applied
	assert.Equal(t, "apt", globalConfig.DefaultProvider)
	assert.False(t, globalConfig.Confirmations.Install)
	assert.False(t, globalConfig.Confirmations.Uninstall)
	assert.False(t, globalConfig.Confirmations.Upgrade)
	assert.False(t, globalConfig.Confirmations.SystemChanges)
	assert.False(t, globalConfig.Confirmations.ServiceOps)
	assert.True(t, globalConfig.Output.ShowCommands)
	assert.True(t, globalConfig.Output.ShowExitCodes)
}

func TestSetupCompletion(t *testing.T) {
	// This test ensures SetupCompletion doesn't panic
	assert.NotPanics(t, func() {
		SetupCompletion()
	})
}