package action

import (
	"context"
	"testing"
	"time"

	"sai/internal/config"
	"sai/internal/interfaces"
	"sai/internal/output"
	"sai/internal/types"
	"sai/internal/ui"
	"sai/internal/validation"
)

// Mock implementations for testing

type mockProviderManager struct {
	providers map[string]*types.ProviderData
}

func (m *mockProviderManager) LoadProviders(providerDir string) error { return nil }
func (m *mockProviderManager) GetProvider(name string) (*types.ProviderData, error) {
	if provider, exists := m.providers[name]; exists {
		return provider, nil
	}
	return nil, nil
}
func (m *mockProviderManager) GetAvailableProviders() []*types.ProviderData {
	var providers []*types.ProviderData
	for _, provider := range m.providers {
		providers = append(providers, provider)
	}
	return providers
}
func (m *mockProviderManager) SelectProvider(software string, action string, preferredProvider string) (*types.ProviderData, error) {
	return m.GetProvider(preferredProvider)
}
func (m *mockProviderManager) IsProviderAvailable(name string) bool {
	_, exists := m.providers[name]
	return exists
}
func (m *mockProviderManager) GetProvidersForAction(action string) []*types.ProviderData {
	var providers []*types.ProviderData
	for _, provider := range m.providers {
		if _, hasAction := provider.Actions[action]; hasAction {
			providers = append(providers, provider)
		}
	}
	return providers
}
func (m *mockProviderManager) ValidateProvider(provider *types.ProviderData) error { return nil }
func (m *mockProviderManager) ReloadProviders() error                              { return nil }

type mockSaidataManager struct {
	saidata map[string]*types.SoftwareData
}

func (m *mockSaidataManager) LoadSoftware(name string) (*types.SoftwareData, error) {
	if data, exists := m.saidata[name]; exists {
		return data, nil
	}
	return nil, nil
}
func (m *mockSaidataManager) GetProviderConfig(software string, provider string) (*types.ProviderConfig, error) {
	return nil, nil
}
func (m *mockSaidataManager) GenerateDefaults(software string) (*types.SoftwareData, error) {
	return &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name: software,
		},
		IsGenerated: true,
	}, nil
}
func (m *mockSaidataManager) UpdateRepository() error                                    { return nil }
func (m *mockSaidataManager) SearchSoftware(query string) ([]*interfaces.SoftwareInfo, error) { return nil, nil }
func (m *mockSaidataManager) ValidateData(data []byte) error                             { return nil }
func (m *mockSaidataManager) ManageRepositoryOperations() error                         { return nil }
func (m *mockSaidataManager) SynchronizeRepository() error                              { return nil }
func (m *mockSaidataManager) GetSoftwareList() ([]string, error)                       { return nil, nil }
func (m *mockSaidataManager) CacheData(software string, data *types.SoftwareData) error { return nil }
func (m *mockSaidataManager) GetCachedData(software string) (*types.SoftwareData, error) { return nil, nil }

type mockExecutor struct{}

func (m *mockExecutor) Execute(ctx context.Context, provider *types.ProviderData, action string, software string, saidata *types.SoftwareData, options interfaces.ExecuteOptions) (*interfaces.ExecutionResult, error) {
	return &interfaces.ExecutionResult{
		Success:  true,
		Output:   "Mock execution successful",
		Commands: []string{"mock-command"},
		ExitCode: 0,
		Duration: time.Millisecond * 100,
	}, nil
}
func (m *mockExecutor) ValidateAction(provider *types.ProviderData, action string, software string, saidata *types.SoftwareData) error {
	return nil
}
func (m *mockExecutor) ValidateResources(saidata *types.SoftwareData, action string) (*interfaces.ResourceValidationResult, error) {
	return &interfaces.ResourceValidationResult{Valid: true, CanProceed: true}, nil
}
func (m *mockExecutor) DryRun(ctx context.Context, provider *types.ProviderData, action string, software string, saidata *types.SoftwareData, options interfaces.ExecuteOptions) (*interfaces.ExecutionResult, error) {
	return &interfaces.ExecutionResult{
		Success:  true,
		Output:   "Mock dry run",
		Commands: []string{"mock-command --dry-run"},
		ExitCode: 0,
		Duration: time.Millisecond * 50,
	}, nil
}
func (m *mockExecutor) CanExecute(provider *types.ProviderData, action string, software string, saidata *types.SoftwareData) bool {
	return true
}
func (m *mockExecutor) RenderTemplate(template string, saidata *types.SoftwareData, provider *types.ProviderData) (string, error) {
	return template, nil
}
func (m *mockExecutor) ExecuteCommand(ctx context.Context, command string, options interfaces.CommandOptions) (*interfaces.CommandResult, error) {
	return &interfaces.CommandResult{
		Command:  command,
		Output:   "Mock command output",
		ExitCode: 0,
		Duration: time.Millisecond * 10,
	}, nil
}
func (m *mockExecutor) ExecuteSteps(ctx context.Context, steps []types.Step, saidata *types.SoftwareData, provider *types.ProviderData, options interfaces.ExecuteOptions) (*interfaces.ExecutionResult, error) {
	return &interfaces.ExecutionResult{
		Success:  true,
		Output:   "Mock steps execution",
		Commands: []string{"step1", "step2"},
		ExitCode: 0,
		Duration: time.Millisecond * 200,
	}, nil
}

func TestActionManager_ExecuteAction(t *testing.T) {
	// Setup test data
	provider := &types.ProviderData{
		Version: "1.0",
		Provider: types.ProviderInfo{
			Name:        "test-provider",
			DisplayName: "Test Provider",
			Type:        "package_manager",
			Platforms:   []string{"linux"},
			Priority:    10,
		},
		Actions: map[string]types.Action{
			"install": {
				Description: "Install software",
				Template:    "test-install {{.Software}}",
			},
		},
	}

	saidata := &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name: "test-software",
		},
		Packages: []types.Package{
			{Name: "test-package"},
		},
	}

	// Setup mocks
	providerManager := &mockProviderManager{
		providers: map[string]*types.ProviderData{
			"test-provider": provider,
		},
	}

	saidataManager := &mockSaidataManager{
		saidata: map[string]*types.SoftwareData{
			"test-software": saidata,
		},
	}

	executor := &mockExecutor{}
	validator := validation.NewResourceValidator()
	cfg := &config.Config{
		Confirmations: config.ConfirmationConfig{
			Install:      false, // Disable confirmation for test
			InfoCommands: false,
		},
		Output: config.OutputConfig{
			ShowCommands: true,
		},
	}

	formatter := output.NewOutputFormatter(cfg, false, false, false)
	ui := ui.NewUserInterface(cfg, formatter)

	// Create action manager
	actionManager := NewActionManager(
		providerManager,
		saidataManager,
		executor,
		validator,
		cfg,
		ui,
		formatter,
	)

	// Test successful action execution
	ctx := context.Background()
	options := interfaces.ActionOptions{
		Provider: "test-provider",
		DryRun:   false,
		Yes:      true, // Skip confirmations
		Timeout:  30 * time.Second,
	}

	result, err := actionManager.ExecuteAction(ctx, "install", "test-software", options)

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if !result.Success {
		t.Errorf("Expected success, got failure: %v", result.Error)
	}

	if result.Action != "install" {
		t.Errorf("Expected action 'install', got: %s", result.Action)
	}

	if result.Software != "test-software" {
		t.Errorf("Expected software 'test-software', got: %s", result.Software)
	}

	if result.Provider != "test-provider" {
		t.Errorf("Expected provider 'test-provider', got: %s", result.Provider)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got: %d", result.ExitCode)
	}
}

func TestActionManager_SafetyChecks(t *testing.T) {
	// Setup test data with missing resources
	provider := &types.ProviderData{
		Version: "1.0",
		Provider: types.ProviderInfo{
			Name:        "test-provider",
			DisplayName: "Test Provider",
			Type:        "package_manager",
			Platforms:   []string{"linux"},
			Priority:    10,
		},
		Actions: map[string]types.Action{
			"install": {
				Description: "Install software",
				Template:    "test-install {{.Software}}",
			},
		},
	}

	saidata := &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name: "test-software",
		},
		Commands: []types.Command{
			{Name: "nonexistent-command", Path: "/nonexistent/path"},
		},
	}

	// Setup mocks
	providerManager := &mockProviderManager{
		providers: map[string]*types.ProviderData{
			"test-provider": provider,
		},
	}

	saidataManager := &mockSaidataManager{
		saidata: map[string]*types.SoftwareData{
			"test-software": saidata,
		},
	}

	executor := &mockExecutor{}
	validator := validation.NewResourceValidator()
	cfg := &config.Config{
		Confirmations: config.ConfirmationConfig{
			Install:      false,
			InfoCommands: false,
		},
		Output: config.OutputConfig{
			ShowCommands: true,
		},
	}

	formatter := output.NewOutputFormatter(cfg, false, false, false)
	ui := ui.NewUserInterface(cfg, formatter)

	// Create action manager
	actionManager := NewActionManager(
		providerManager,
		saidataManager,
		executor,
		validator,
		cfg,
		ui,
		formatter,
	)

	// Test safety checks
	safetyManager := actionManager.safetyManager
	safetyResult, err := safetyManager.CheckActionSafety("install", "test-software", provider, saidata)

	if err != nil {
		t.Errorf("Expected no error from safety check, got: %v", err)
	}

	if safetyResult == nil {
		t.Fatal("Expected safety result, got nil")
	}

	// Should fail due to missing command
	if safetyResult.Safe {
		t.Error("Expected safety check to fail due to missing command")
	}

	// Check that we have safety checks
	if len(safetyResult.Checks) == 0 {
		t.Error("Expected safety checks to be performed")
	}

	// Verify resource existence check failed
	resourceCheckFound := false
	for _, check := range safetyResult.Checks {
		if check.Name == "Resource Existence" && !check.Passed {
			resourceCheckFound = true
			break
		}
	}

	if !resourceCheckFound {
		t.Error("Expected resource existence check to fail")
	}
}

func TestActionManager_ProviderSelection(t *testing.T) {
	// Setup multiple providers
	provider1 := &types.ProviderData{
		Version: "1.0",
		Provider: types.ProviderInfo{
			Name:        "provider1",
			DisplayName: "Provider 1",
			Type:        "package_manager",
			Platforms:   []string{"linux"},
			Priority:    10,
		},
		Actions: map[string]types.Action{
			"install": {Description: "Install software"},
		},
	}

	provider2 := &types.ProviderData{
		Version: "1.0",
		Provider: types.ProviderInfo{
			Name:        "provider2",
			DisplayName: "Provider 2",
			Type:        "package_manager",
			Platforms:   []string{"linux"},
			Priority:    5,
		},
		Actions: map[string]types.Action{
			"install": {Description: "Install software"},
		},
	}

	// Setup mocks
	providerManager := &mockProviderManager{
		providers: map[string]*types.ProviderData{
			"provider1": provider1,
			"provider2": provider2,
		},
	}

	saidataManager := &mockSaidataManager{
		saidata: map[string]*types.SoftwareData{
			"test-software": {
				Version: "0.2",
				Metadata: types.Metadata{Name: "test-software"},
			},
		},
	}

	executor := &mockExecutor{}
	validator := validation.NewResourceValidator()
	cfg := &config.Config{
		ProviderPriority: map[string]int{
			"provider1": 10,
			"provider2": 5,
		},
	}

	formatter := output.NewOutputFormatter(cfg, false, false, false)
	ui := ui.NewUserInterface(cfg, formatter)

	actionManager := NewActionManager(
		providerManager,
		saidataManager,
		executor,
		validator,
		cfg,
		ui,
		formatter,
	)

	// Test provider selection with --yes flag (should select highest priority)
	options, err := actionManager.GetAvailableProviders("test-software", "install")
	if err != nil {
		t.Fatalf("Failed to get available providers: %v", err)
	}

	if len(options) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(options))
	}

	// Test that providers are sorted by priority
	if options[0].Provider.Provider.Name != "provider1" {
		t.Errorf("Expected provider1 to be first (highest priority), got: %s", options[0].Provider.Provider.Name)
	}
}