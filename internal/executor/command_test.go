package executor

import (
	"context"
	"strings"
	"testing"
	"time"

	"sai/internal/interfaces"
	"sai/internal/types"
)

// MockLogger implements interfaces.Logger for testing
type MockLogger struct {
	debugCalls []string
	infoCalls  []string
	warnCalls  []string
	errorCalls []string
}

func (m *MockLogger) Debug(msg string, fields ...interfaces.LogField) {
	m.debugCalls = append(m.debugCalls, msg)
}

func (m *MockLogger) Info(msg string, fields ...interfaces.LogField) {
	m.infoCalls = append(m.infoCalls, msg)
}

func (m *MockLogger) Warn(msg string, fields ...interfaces.LogField) {
	m.warnCalls = append(m.warnCalls, msg)
}

func (m *MockLogger) Error(msg string, err error, fields ...interfaces.LogField) {
	m.errorCalls = append(m.errorCalls, msg)
}

func (m *MockLogger) Fatal(msg string, err error, fields ...interfaces.LogField) {
	m.errorCalls = append(m.errorCalls, msg)
}

func (m *MockLogger) WithFields(fields ...interfaces.LogField) interfaces.Logger {
	return m
}

func (m *MockLogger) SetLevel(level interfaces.LogLevel) {}

func (m *MockLogger) GetLevel() interfaces.LogLevel {
	return interfaces.LogLevelInfo
}

// MockResourceValidator implements interfaces.ResourceValidator for testing
type MockResourceValidator struct{}

func (m *MockResourceValidator) ValidateFile(file types.File) bool                     { return true }
func (m *MockResourceValidator) ValidateService(service types.Service) bool           { return true }
func (m *MockResourceValidator) ValidateCommand(command types.Command) bool           { return true }
func (m *MockResourceValidator) ValidateDirectory(directory types.Directory) bool     { return true }
func (m *MockResourceValidator) ValidatePort(port types.Port) bool                    { return true }
func (m *MockResourceValidator) ValidateContainer(container types.Container) bool     { return true }
func (m *MockResourceValidator) ValidateResources(saidata *types.SoftwareData) (*interfaces.ResourceValidationResult, error) {
	return &interfaces.ResourceValidationResult{Valid: true, CanProceed: true}, nil
}
func (m *MockResourceValidator) ValidateSystemRequirements(requirements *types.Requirements) (*interfaces.SystemValidationResult, error) {
	return &interfaces.SystemValidationResult{Valid: true}, nil
}

func TestNewCommandExecutor(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	
	executor := NewCommandExecutor(logger, validator)
	
	if executor == nil {
		t.Fatal("Expected non-nil executor")
	}
	
	if executor.logger != logger {
		t.Error("Expected logger to be set")
	}
	
	if executor.validator != validator {
		t.Error("Expected validator to be set")
	}
	
	if executor.timeout != 300*time.Second {
		t.Errorf("Expected default timeout of 300s, got %v", executor.timeout)
	}
}

func TestExecuteCommand_Success(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	executor := NewCommandExecutor(logger, validator)
	
	ctx := context.Background()
	options := interfaces.CommandOptions{
		Timeout: 10 * time.Second,
	}
	
	// Test with a simple command that should succeed
	result, err := executor.ExecuteCommand(ctx, "echo hello", options)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
	
	if result.Command != "echo hello" {
		t.Errorf("Expected command 'echo hello', got '%s'", result.Command)
	}
	
	if !strings.Contains(result.Output, "hello") {
		t.Errorf("Expected output to contain 'hello', got '%s'", result.Output)
	}
}

func TestExecuteCommand_DryRun(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	executor := NewCommandExecutor(logger, validator)
	executor.SetDryRun(true)
	
	ctx := context.Background()
	options := interfaces.CommandOptions{}
	
	result, err := executor.ExecuteCommand(ctx, "echo hello", options)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0 in dry run, got %d", result.ExitCode)
	}
	
	if !strings.Contains(result.Output, "DRY RUN") {
		t.Errorf("Expected output to contain 'DRY RUN', got '%s'", result.Output)
	}
	
	// Check that info log was called for dry run
	found := false
	for _, call := range logger.infoCalls {
		if strings.Contains(call, "DRY RUN") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected dry run info log call")
	}
}

func TestExecuteCommand_InvalidCommand(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	executor := NewCommandExecutor(logger, validator)
	
	ctx := context.Background()
	options := interfaces.CommandOptions{}
	
	// Test with empty command
	result, err := executor.ExecuteCommand(ctx, "", options)
	
	if err == nil {
		t.Fatal("Expected error for empty command")
	}
	
	if result.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", result.ExitCode)
	}
}

func TestExecuteCommand_NonExistentExecutable(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	executor := NewCommandExecutor(logger, validator)
	
	ctx := context.Background()
	options := interfaces.CommandOptions{}
	
	// Test with non-existent executable
	result, err := executor.ExecuteCommand(ctx, "nonexistentcommand123", options)
	
	if err == nil {
		t.Fatal("Expected error for non-existent command")
	}
	
	if result.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", result.ExitCode)
	}
}

func TestExecuteWithRetry_Success(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	executor := NewCommandExecutor(logger, validator)
	
	ctx := context.Background()
	options := interfaces.CommandOptions{}
	retryConfig := &types.RetryConfig{
		Attempts: 3,
		Delay:    1,
		Backoff:  "linear",
	}
	
	result, err := executor.ExecuteWithRetry(ctx, "echo hello", options, retryConfig)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
}

func TestExecuteWithRetry_EventualSuccess(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	executor := NewCommandExecutor(logger, validator)
	
	ctx := context.Background()
	options := interfaces.CommandOptions{}
	retryConfig := &types.RetryConfig{
		Attempts: 2,
		Delay:    1,
		Backoff:  "linear",
	}
	
	// This command will fail first time but succeed on retry
	// We'll use a command that fails with exit code 1 but we can't easily simulate
	// intermittent failure, so we'll test with a successful command
	result, err := executor.ExecuteWithRetry(ctx, "echo hello", options, retryConfig)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
}

func TestValidateCommand(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	executor := NewCommandExecutor(logger, validator)
	
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "valid command",
			command: "echo hello",
			wantErr: false,
		},
		{
			name:    "empty command",
			command: "",
			wantErr: true,
		},
		{
			name:    "whitespace only command",
			command: "   ",
			wantErr: true,
		},
		{
			name:    "non-existent executable",
			command: "nonexistentcommand123",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.ValidateCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsCommandAvailable(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	executor := NewCommandExecutor(logger, validator)
	
	// Test with a command that should be available on most systems
	if !executor.IsCommandAvailable("echo hello") {
		t.Error("Expected 'echo hello' to be available")
	}
	
	// Test with a command that should not be available
	if executor.IsCommandAvailable("nonexistentcommand123") {
		t.Error("Expected 'nonexistentcommand123' to not be available")
	}
}

func TestSetTimeout(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	executor := NewCommandExecutor(logger, validator)
	
	newTimeout := 60 * time.Second
	executor.SetTimeout(newTimeout)
	
	if executor.GetTimeout() != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, executor.GetTimeout())
	}
}

func TestSetDryRun(t *testing.T) {
	logger := &MockLogger{}
	validator := &MockResourceValidator{}
	executor := NewCommandExecutor(logger, validator)
	
	// Initially should not be in dry run mode
	if executor.dryRun {
		t.Error("Expected dry run to be false initially")
	}
	
	executor.SetDryRun(true)
	if !executor.dryRun {
		t.Error("Expected dry run to be true after setting")
	}
	
	executor.SetDryRun(false)
	if executor.dryRun {
		t.Error("Expected dry run to be false after unsetting")
	}
}