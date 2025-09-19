package executor

import (
	"context"
	"testing"

	"sai/internal/interfaces"
	"sai/internal/types"
)

// MockTemplateEngine implements interfaces.TemplateEngine for testing
type MockTemplateEngine struct {
	renderFunc    func(string, *interfaces.TemplateContext) (string, error)
	validateFunc  func(string) error
	safetyMode    bool
}

func (m *MockTemplateEngine) Render(templateStr string, context *interfaces.TemplateContext) (string, error) {
	if m.renderFunc != nil {
		return m.renderFunc(templateStr, context)
	}
	// Default behavior - return template as-is for simple cases
	return templateStr, nil
}

func (m *MockTemplateEngine) ValidateTemplate(templateStr string) error {
	if m.validateFunc != nil {
		return m.validateFunc(templateStr)
	}
	return nil
}

func (m *MockTemplateEngine) SetSafetyMode(enabled bool) {
	m.safetyMode = enabled
}

func (m *MockTemplateEngine) SetSaidata(saidata *types.SoftwareData) {
	// Mock implementation
}

func TestNewGenericExecutor(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	commandExecutor := NewCommandExecutor(logger, validator)
	templateEngine := &MockTemplateEngine{}
	
	executor := NewGenericExecutor(commandExecutor, templateEngine, logger, validator)
	
	if executor == nil {
		t.Fatal("Expected non-nil executor")
	}
	
	if executor.commandExecutor != commandExecutor {
		t.Error("Expected command executor to be set")
	}
	
	if executor.templateEngine != templateEngine {
		t.Error("Expected template engine to be set")
	}
	
	if executor.logger != logger {
		t.Error("Expected logger to be set")
	}
	
	if executor.validator != validator {
		t.Error("Expected validator to be set")
	}
}

func TestValidateAction(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	commandExecutor := NewCommandExecutor(logger, validator)
	templateEngine := &MockTemplateEngine{}
	executor := NewGenericExecutor(commandExecutor, templateEngine, logger, validator)
	
	// Create test provider with valid action
	provider := &types.ProviderData{
		Provider: types.ProviderInfo{
			Name: "test-provider",
		},
		Actions: map[string]types.Action{
			"install": {
				Command: "echo install",
			},
		},
	}
	
	// Test valid action
	err := executor.ValidateAction(provider, "install", "test-software", nil)
	if err != nil {
		t.Errorf("Expected no error for valid action, got %v", err)
	}
	
	// Test non-existent action
	err = executor.ValidateAction(provider, "nonexistent", "test-software", nil)
	if err == nil {
		t.Error("Expected error for non-existent action")
	}
	
	// Test action with no execution method
	provider.Actions["invalid"] = types.Action{}
	err = executor.ValidateAction(provider, "invalid", "test-software", nil)
	if err == nil {
		t.Error("Expected error for action with no execution method")
	}
}

func TestCanExecute(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	commandExecutor := NewCommandExecutor(logger, validator)
	templateEngine := &MockTemplateEngine{}
	executor := NewGenericExecutor(commandExecutor, templateEngine, logger, validator)
	
	provider := &types.ProviderData{
		Provider: types.ProviderInfo{
			Name: "test-provider",
		},
		Actions: map[string]types.Action{
			"install": {
				Command: "echo install",
			},
		},
	}
	
	// Test valid action
	canExecute := executor.CanExecute(provider, "install", "test-software", nil)
	if !canExecute {
		t.Error("Expected action to be executable")
	}
	
	// Test invalid action
	canExecute = executor.CanExecute(provider, "nonexistent", "test-software", nil)
	if canExecute {
		t.Error("Expected action to not be executable")
	}
}

func TestDryRun(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	commandExecutor := NewCommandExecutor(logger, validator)
	templateEngine := &MockTemplateEngine{
		renderFunc: func(template string, context *interfaces.TemplateContext) (string, error) {
			return "echo test-command", nil
		},
	}
	executor := NewGenericExecutor(commandExecutor, templateEngine, logger, validator)
	
	provider := &types.ProviderData{
		Provider: types.ProviderInfo{
			Name: "test-provider",
		},
		Actions: map[string]types.Action{
			"install": {
				Command: "echo install {{.Software}}",
			},
		},
	}
	
	ctx := context.Background()
	options := interfaces.ExecuteOptions{
		DryRun: true,
	}
	
	result, err := executor.DryRun(ctx, provider, "install", "test-software", nil, options)
	
	if err != nil {
		t.Fatalf("Expected no error in dry run, got %v", err)
	}
	
	if !result.Success {
		t.Error("Expected dry run to succeed")
	}
	
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
	
	if len(result.Commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(result.Commands))
	}
	
	if result.Commands[0] != "echo test-command" {
		t.Errorf("Expected 'echo test-command', got '%s'", result.Commands[0])
	}
}

func TestDryRun_MultipleSteps(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	commandExecutor := NewCommandExecutor(logger, validator)
	templateEngine := &MockTemplateEngine{
		renderFunc: func(template string, context *interfaces.TemplateContext) (string, error) {
			return template, nil // Return template as-is for simplicity
		},
	}
	executor := NewGenericExecutor(commandExecutor, templateEngine, logger, validator)
	
	provider := &types.ProviderData{
		Provider: types.ProviderInfo{
			Name: "test-provider",
		},
		Actions: map[string]types.Action{
			"install": {
				Steps: []types.Step{
					{Name: "step1", Command: "echo step1"},
					{Name: "step2", Command: "echo step2"},
				},
			},
		},
	}
	
	ctx := context.Background()
	options := interfaces.ExecuteOptions{
		DryRun: true,
	}
	
	result, err := executor.DryRun(ctx, provider, "install", "test-software", nil, options)
	
	if err != nil {
		t.Fatalf("Expected no error in dry run, got %v", err)
	}
	
	if !result.Success {
		t.Error("Expected dry run to succeed")
	}
	
	if len(result.Commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(result.Commands))
	}
	
	if result.Commands[0] != "echo step1" {
		t.Errorf("Expected 'echo step1', got '%s'", result.Commands[0])
	}
	
	if result.Commands[1] != "echo step2" {
		t.Errorf("Expected 'echo step2', got '%s'", result.Commands[1])
	}
}

func TestExecute_SingleCommand(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	commandExecutor := NewCommandExecutor(logger, validator)
	templateEngine := &MockTemplateEngine{
		renderFunc: func(template string, context *interfaces.TemplateContext) (string, error) {
			return "echo hello", nil
		},
	}
	executor := NewGenericExecutor(commandExecutor, templateEngine, logger, validator)
	
	provider := &types.ProviderData{
		Provider: types.ProviderInfo{
			Name: "test-provider",
		},
		Actions: map[string]types.Action{
			"install": {
				Command: "echo install {{.Software}}",
			},
		},
	}
	
	ctx := context.Background()
	options := interfaces.ExecuteOptions{}
	
	result, err := executor.Execute(ctx, provider, "install", "test-software", nil, options)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !result.Success {
		t.Error("Expected execution to succeed")
	}
	
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
	
	if result.Provider != "test-provider" {
		t.Errorf("Expected provider 'test-provider', got '%s'", result.Provider)
	}
}

func TestExecute_NonExistentAction(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	commandExecutor := NewCommandExecutor(logger, validator)
	templateEngine := &MockTemplateEngine{}
	executor := NewGenericExecutor(commandExecutor, templateEngine, logger, validator)
	
	provider := &types.ProviderData{
		Provider: types.ProviderInfo{
			Name: "test-provider",
		},
		Actions: map[string]types.Action{},
	}
	
	ctx := context.Background()
	options := interfaces.ExecuteOptions{}
	
	result, err := executor.Execute(ctx, provider, "nonexistent", "test-software", nil, options)
	
	if err == nil {
		t.Fatal("Expected error for non-existent action")
	}
	
	if result.Success {
		t.Error("Expected execution to fail")
	}
	
	if result.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", result.ExitCode)
	}
}

func TestExecuteSteps(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	commandExecutor := NewCommandExecutor(logger, validator)
	templateEngine := &MockTemplateEngine{
		renderFunc: func(template string, context *interfaces.TemplateContext) (string, error) {
			return template, nil // Return template as-is
		},
	}
	executor := NewGenericExecutor(commandExecutor, templateEngine, logger, validator)
	
	steps := []types.Step{
		{Name: "step1", Command: "echo step1"},
		{Name: "step2", Command: "echo step2"},
	}
	
	provider := &types.ProviderData{
		Provider: types.ProviderInfo{
			Name: "test-provider",
		},
	}
	
	ctx := context.Background()
	options := interfaces.ExecuteOptions{}
	
	result, err := executor.ExecuteSteps(ctx, steps, nil, provider, options)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !result.Success {
		t.Error("Expected execution to succeed")
	}
	
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
	
	if len(result.Commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(result.Commands))
	}
}

func TestExecuteSteps_IgnoreFailure(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	commandExecutor := NewCommandExecutor(logger, validator)
	templateEngine := &MockTemplateEngine{
		renderFunc: func(template string, context *interfaces.TemplateContext) (string, error) {
			if template == "failing-command" {
				return "nonexistentcommand123", nil // This will fail
			}
			return template, nil
		},
	}
	executor := NewGenericExecutor(commandExecutor, templateEngine, logger, validator)
	
	steps := []types.Step{
		{Name: "step1", Command: "echo step1"},
		{Name: "failing-step", Command: "failing-command", IgnoreFailure: true},
		{Name: "step3", Command: "echo step3"},
	}
	
	provider := &types.ProviderData{
		Provider: types.ProviderInfo{
			Name: "test-provider",
		},
	}
	
	ctx := context.Background()
	options := interfaces.ExecuteOptions{}
	
	result, err := executor.ExecuteSteps(ctx, steps, nil, provider, options)
	
	if err != nil {
		t.Fatalf("Expected no error when ignoring failures, got %v", err)
	}
	
	if !result.Success {
		t.Error("Expected execution to succeed when ignoring failures")
	}
	
	// Should have executed 3 commands (including the failing one)
	if len(result.Commands) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(result.Commands))
	}
}

func TestRenderTemplate(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	commandExecutor := NewCommandExecutor(logger, validator)
	templateEngine := &MockTemplateEngine{
		renderFunc: func(template string, context *interfaces.TemplateContext) (string, error) {
			return "rendered: " + template, nil
		},
	}
	executor := NewGenericExecutor(commandExecutor, templateEngine, logger, validator)
	
	provider := &types.ProviderData{
		Provider: types.ProviderInfo{
			Name: "test-provider",
		},
	}
	
	result, err := executor.RenderTemplate("test template", nil, provider)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result != "rendered: test template" {
		t.Errorf("Expected 'rendered: test template', got '%s'", result)
	}
}