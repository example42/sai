package interfaces

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"sai/internal/types"
)

func TestLogLevel(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelDebug, "debug"},
		{LogLevelInfo, "info"},
		{LogLevelWarn, "warn"},
		{LogLevelError, "error"},
		{LogLevelFatal, "fatal"},
		{LogLevel(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDataStructures(t *testing.T) {
	t.Run("ActionOptions", func(t *testing.T) {
		options := ActionOptions{
			Provider: "apt",
			DryRun:   true,
			Verbose:  true,
			Variables: map[string]string{
				"key": "value",
			},
		}
		
		assert.Equal(t, "apt", options.Provider)
		assert.True(t, options.DryRun)
		assert.True(t, options.Verbose)
		assert.Equal(t, "value", options.Variables["key"])
	})

	t.Run("ActionResult", func(t *testing.T) {
		result := ActionResult{
			Action:   "install",
			Software: "nginx",
			Provider: "apt",
			Success:  true,
			ExitCode: 0,
		}
		
		assert.Equal(t, "install", result.Action)
		assert.Equal(t, "nginx", result.Software)
		assert.Equal(t, "apt", result.Provider)
		assert.True(t, result.Success)
		assert.Equal(t, 0, result.ExitCode)
	})

	t.Run("Change", func(t *testing.T) {
		change := Change{
			Type:        "file",
			Resource:    "/etc/nginx/nginx.conf",
			Action:      "created",
			OldValue:    "",
			NewValue:    "config content",
			Reversible:  true,
			RollbackCmd: "rm /etc/nginx/nginx.conf",
		}
		
		assert.Equal(t, "file", change.Type)
		assert.Equal(t, "/etc/nginx/nginx.conf", change.Resource)
		assert.Equal(t, "created", change.Action)
		assert.True(t, change.Reversible)
	})

	t.Run("ProviderOption", func(t *testing.T) {
		option := ProviderOption{
			PackageName: "nginx",
			Version:     "1.18.0",
			IsInstalled: false,
			Priority:    90,
		}
		
		assert.Equal(t, "nginx", option.PackageName)
		assert.Equal(t, "1.18.0", option.Version)
		assert.False(t, option.IsInstalled)
		assert.Equal(t, 90, option.Priority)
	})

	t.Run("Config", func(t *testing.T) {
		config := Config{
			SaidataRepository: "https://github.com/example/saidata.git",
			DefaultProvider:   "apt",
			ProviderPriority: map[string]int{
				"apt":    90,
				"docker": 70,
			},
			LogLevel: "info",
		}
		
		assert.Equal(t, "https://github.com/example/saidata.git", config.SaidataRepository)
		assert.Equal(t, "apt", config.DefaultProvider)
		assert.Equal(t, 90, config.ProviderPriority["apt"])
		assert.Equal(t, "info", config.LogLevel)
	})

	t.Run("ResourceValidationResult", func(t *testing.T) {
		result := ResourceValidationResult{
			Valid:              false,
			MissingFiles:       []string{"/etc/config"},
			MissingDirectories: []string{"/var/lib/app"},
			MissingCommands:    []string{"/usr/bin/app"},
			MissingServices:    []string{"app-service"},
			InvalidPorts:       []int{70000},
			Warnings:           []string{"Port may be in use"},
			CanProceed:         true,
		}
		
		assert.False(t, result.Valid)
		assert.Contains(t, result.MissingFiles, "/etc/config")
		assert.Contains(t, result.MissingDirectories, "/var/lib/app")
		assert.Contains(t, result.MissingCommands, "/usr/bin/app")
		assert.Contains(t, result.MissingServices, "app-service")
		assert.Contains(t, result.InvalidPorts, 70000)
		assert.Contains(t, result.Warnings, "Port may be in use")
		assert.True(t, result.CanProceed)
	})
}

// Test that interfaces can be implemented (compilation test)
func TestInterfaceCompilation(t *testing.T) {
	// This test ensures that the interfaces are properly defined
	// and can be implemented. We don't need actual implementations here,
	// just verification that the interfaces compile correctly.
	
	var _ ProviderManager = (*mockProviderManager)(nil)
	var _ SaidataManager = (*mockSaidataManager)(nil)
	var _ ActionManager = (*mockActionManager)(nil)
	var _ GenericExecutor = (*mockGenericExecutor)(nil)
	var _ DefaultsGenerator = (*mockDefaultsGenerator)(nil)
	var _ ResourceValidator = (*mockResourceValidator)(nil)
	var _ ConfigManager = (*mockConfigManager)(nil)
	var _ Logger = (*mockLogger)(nil)
}

// Mock implementations for compilation testing
type mockProviderManager struct{}
func (m *mockProviderManager) LoadProviders(string) error { return nil }
func (m *mockProviderManager) GetProvider(string) (*types.ProviderData, error) { return nil, nil }
func (m *mockProviderManager) GetAvailableProviders() []*types.ProviderData { return nil }
func (m *mockProviderManager) SelectProvider(string, string, string) (*types.ProviderData, error) { return nil, nil }
func (m *mockProviderManager) IsProviderAvailable(string) bool { return false }
func (m *mockProviderManager) GetProvidersForAction(string) []*types.ProviderData { return nil }
func (m *mockProviderManager) ValidateProvider(*types.ProviderData) error { return nil }
func (m *mockProviderManager) ReloadProviders() error { return nil }

type mockSaidataManager struct{}
func (m *mockSaidataManager) LoadSoftware(string) (*types.SoftwareData, error) { return nil, nil }
func (m *mockSaidataManager) GetProviderConfig(string, string) (*types.ProviderConfig, error) { return nil, nil }
func (m *mockSaidataManager) GenerateDefaults(string) (*types.SoftwareData, error) { return nil, nil }
func (m *mockSaidataManager) UpdateRepository() error { return nil }
func (m *mockSaidataManager) SearchSoftware(string) ([]*SoftwareInfo, error) { return nil, nil }
func (m *mockSaidataManager) ValidateData([]byte) error { return nil }
func (m *mockSaidataManager) ManageRepositoryOperations() error { return nil }
func (m *mockSaidataManager) SynchronizeRepository() error { return nil }
func (m *mockSaidataManager) GetSoftwareList() ([]string, error) { return nil, nil }
func (m *mockSaidataManager) CacheData(string, *types.SoftwareData) error { return nil }
func (m *mockSaidataManager) GetCachedData(string) (*types.SoftwareData, error) { return nil, nil }

type mockActionManager struct{}
func (m *mockActionManager) ExecuteAction(context.Context, string, string, ActionOptions) (*ActionResult, error) { return nil, nil }
func (m *mockActionManager) ValidateAction(string, string) error { return nil }
func (m *mockActionManager) GetAvailableActions(string) ([]string, error) { return nil, nil }
func (m *mockActionManager) GetActionInfo(string) (*ActionInfo, error) { return nil, nil }
func (m *mockActionManager) ResolveSoftwareData(string) (*types.SoftwareData, error) { return nil, nil }
func (m *mockActionManager) ValidateResourcesExist(*types.SoftwareData, string) (*ResourceValidationResult, error) { return nil, nil }
func (m *mockActionManager) GetAvailableProviders(string, string) ([]*ProviderOption, error) { return nil, nil }
func (m *mockActionManager) RequiresConfirmation(string) bool { return false }
func (m *mockActionManager) SearchAcrossProviders(string) ([]*SearchResult, error) { return nil, nil }
func (m *mockActionManager) GetSoftwareInfo(string) ([]*SoftwareInfo, error) { return nil, nil }
func (m *mockActionManager) GetSoftwareVersions(string) ([]*VersionInfo, error) { return nil, nil }
func (m *mockActionManager) ManageRepositorySetup(*types.SoftwareData) error { return nil }

type mockGenericExecutor struct{}
func (m *mockGenericExecutor) Execute(context.Context, *types.ProviderData, string, string, *types.SoftwareData, ExecuteOptions) (*ExecutionResult, error) { return nil, nil }
func (m *mockGenericExecutor) ValidateAction(*types.ProviderData, string, string, *types.SoftwareData) error { return nil }
func (m *mockGenericExecutor) ValidateResources(*types.SoftwareData, string) (*ResourceValidationResult, error) { return nil, nil }
func (m *mockGenericExecutor) DryRun(context.Context, *types.ProviderData, string, string, *types.SoftwareData, ExecuteOptions) (*ExecutionResult, error) { return nil, nil }
func (m *mockGenericExecutor) CanExecute(*types.ProviderData, string, string, *types.SoftwareData) bool { return false }
func (m *mockGenericExecutor) RenderTemplate(string, *types.SoftwareData, *types.ProviderData) (string, error) { return "", nil }
func (m *mockGenericExecutor) ExecuteCommand(context.Context, string, CommandOptions) (*CommandResult, error) { return nil, nil }
func (m *mockGenericExecutor) ExecuteSteps(context.Context, []types.Step, *types.SoftwareData, *types.ProviderData, ExecuteOptions) (*ExecutionResult, error) { return nil, nil }

type mockDefaultsGenerator struct{}
func (m *mockDefaultsGenerator) GeneratePackageDefaults(string) []types.Package { return nil }
func (m *mockDefaultsGenerator) GenerateServiceDefaults(string) []types.Service { return nil }
func (m *mockDefaultsGenerator) GenerateFileDefaults(string) []types.File { return nil }
func (m *mockDefaultsGenerator) GenerateDirectoryDefaults(string) []types.Directory { return nil }
func (m *mockDefaultsGenerator) GenerateCommandDefaults(string) []types.Command { return nil }
func (m *mockDefaultsGenerator) GeneratePortDefaults(string) []types.Port { return nil }
func (m *mockDefaultsGenerator) ValidatePathExists(string) bool { return false }
func (m *mockDefaultsGenerator) ValidateServiceExists(string) bool { return false }
func (m *mockDefaultsGenerator) ValidateCommandExists(string) bool { return false }
func (m *mockDefaultsGenerator) GenerateDefaults(string) (*types.SoftwareData, error) { return nil, nil }

type mockResourceValidator struct{}
func (m *mockResourceValidator) ValidateFile(types.File) bool { return false }
func (m *mockResourceValidator) ValidateService(types.Service) bool { return false }
func (m *mockResourceValidator) ValidateCommand(types.Command) bool { return false }
func (m *mockResourceValidator) ValidateDirectory(types.Directory) bool { return false }
func (m *mockResourceValidator) ValidatePort(types.Port) bool { return false }
func (m *mockResourceValidator) ValidateContainer(types.Container) bool { return false }
func (m *mockResourceValidator) ValidateResources(*types.SoftwareData) (*ResourceValidationResult, error) { return nil, nil }
func (m *mockResourceValidator) ValidateSystemRequirements(*types.Requirements) (*SystemValidationResult, error) { return nil, nil }

type mockConfigManager struct{}
func (m *mockConfigManager) LoadConfig(string) (*Config, error) { return nil, nil }
func (m *mockConfigManager) SaveConfig(*Config, string) error { return nil }
func (m *mockConfigManager) GetDefaultConfig() *Config { return nil }
func (m *mockConfigManager) MergeConfig(*Config, *Config) *Config { return nil }
func (m *mockConfigManager) ValidateConfig(*Config) error { return nil }
func (m *mockConfigManager) GetConfigPaths() []string { return nil }
func (m *mockConfigManager) WatchConfig(string, func(*Config)) error { return nil }

type mockLogger struct{}
func (m *mockLogger) Debug(string, ...LogField) {}
func (m *mockLogger) Info(string, ...LogField) {}
func (m *mockLogger) Warn(string, ...LogField) {}
func (m *mockLogger) Error(string, error, ...LogField) {}
func (m *mockLogger) Fatal(string, error, ...LogField) {}
func (m *mockLogger) WithFields(...LogField) Logger { return m }
func (m *mockLogger) SetLevel(LogLevel) {}
func (m *mockLogger) GetLevel() LogLevel { return LogLevelInfo }