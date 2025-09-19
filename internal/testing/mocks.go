package testing

import (
	"context"
	"strings"
	"time"

	"sai/internal/errors"
	"sai/internal/interfaces"
	"sai/internal/types"
)

// MockProviderManager implements interfaces.ProviderManager for testing
type MockProviderManager struct {
	providers        map[string]*types.ProviderData
	availableProviders []*types.ProviderData
	loadError        error
	selectError      error
}

func NewMockProviderManager() *MockProviderManager {
	return &MockProviderManager{
		providers:        make(map[string]*types.ProviderData),
		availableProviders: make([]*types.ProviderData, 0),
	}
}

func (m *MockProviderManager) AddProvider(provider *types.ProviderData) {
	m.providers[provider.Provider.Name] = provider
	m.availableProviders = append(m.availableProviders, provider)
}

func (m *MockProviderManager) SetLoadError(err error) {
	m.loadError = err
}

func (m *MockProviderManager) SetSelectError(err error) {
	m.selectError = err
}

func (m *MockProviderManager) LoadProviders(providerDir string) error {
	return m.loadError
}

func (m *MockProviderManager) GetProvider(name string) (*types.ProviderData, error) {
	if m.loadError != nil {
		return nil, m.loadError
	}
	if provider, exists := m.providers[name]; exists {
		return provider, nil
	}
	return nil, errors.NewProviderNotFoundError(name)
}

func (m *MockProviderManager) GetAvailableProviders() []*types.ProviderData {
	return m.availableProviders
}

func (m *MockProviderManager) SelectProvider(software string, action string, preferredProvider string) (*types.ProviderData, error) {
	if m.selectError != nil {
		return nil, m.selectError
	}
	if preferredProvider != "" {
		return m.GetProvider(preferredProvider)
	}
	// Return first available provider
	if len(m.availableProviders) > 0 {
		return m.availableProviders[0], nil
	}
	return nil, errors.NewSAIError(errors.ErrorTypeProviderNotFound, "no providers available")
}

func (m *MockProviderManager) IsProviderAvailable(name string) bool {
	_, exists := m.providers[name]
	return exists
}

func (m *MockProviderManager) GetProvidersForAction(action string) []*types.ProviderData {
	var providers []*types.ProviderData
	for _, provider := range m.providers {
		if _, hasAction := provider.Actions[action]; hasAction {
			providers = append(providers, provider)
		}
	}
	return providers
}

func (m *MockProviderManager) ValidateProvider(provider *types.ProviderData) error {
	return nil
}

func (m *MockProviderManager) ReloadProviders() error {
	return m.loadError
}

// MockSaidataManager implements interfaces.SaidataManager for testing
type MockSaidataManager struct {
	saidata       map[string]*types.SoftwareData
	loadError     error
	generateError error
}

func NewMockSaidataManager() *MockSaidataManager {
	return &MockSaidataManager{
		saidata: make(map[string]*types.SoftwareData),
	}
}

func (m *MockSaidataManager) AddSaidata(name string, data *types.SoftwareData) {
	m.saidata[name] = data
}

func (m *MockSaidataManager) SetLoadError(err error) {
	m.loadError = err
}

func (m *MockSaidataManager) SetGenerateError(err error) {
	m.generateError = err
}

func (m *MockSaidataManager) LoadSoftware(name string) (*types.SoftwareData, error) {
	if m.loadError != nil {
		return nil, m.loadError
	}
	if data, exists := m.saidata[name]; exists {
		return data, nil
	}
	return nil, errors.NewSaidataNotFoundError(name)
}

func (m *MockSaidataManager) GetProviderConfig(software string, provider string) (*types.ProviderConfig, error) {
	data, err := m.LoadSoftware(software)
	if err != nil {
		return nil, err
	}
	if config, exists := data.Providers[provider]; exists {
		return &config, nil
	}
	return nil, errors.NewSAIError(errors.ErrorTypeProviderNotFound, "provider config not found")
}

func (m *MockSaidataManager) GenerateDefaults(software string) (*types.SoftwareData, error) {
	if m.generateError != nil {
		return nil, m.generateError
	}
	return &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name:        software,
			DisplayName: software + " (Generated)",
		},
		Packages: []types.Package{
			{Name: software, Version: "latest"},
		},
		Services: []types.Service{
			{Name: software, ServiceName: software, Type: "systemd"},
		},
		Commands: []types.Command{
			{Name: software, Path: "/usr/bin/" + software},
		},
		IsGenerated: true,
	}, nil
}

func (m *MockSaidataManager) UpdateRepository() error {
	return nil
}

func (m *MockSaidataManager) SearchSoftware(query string) ([]*interfaces.SoftwareInfo, error) {
	var results []*interfaces.SoftwareInfo
	for name, data := range m.saidata {
		if contains(name, query) || contains(data.Metadata.DisplayName, query) || contains(data.Metadata.Category, query) {
			results = append(results, &interfaces.SoftwareInfo{
				Software:    name,
				Provider:    "mock",
				PackageName: name,
				Version:     "1.0.0",
				Description: data.Metadata.Description,
			})
		}
	}
	return results, nil
}

func (m *MockSaidataManager) ValidateData(data []byte) error {
	return nil
}

func (m *MockSaidataManager) ManageRepositoryOperations() error {
	return nil
}

func (m *MockSaidataManager) SynchronizeRepository() error {
	return nil
}

func (m *MockSaidataManager) GetSoftwareList() ([]string, error) {
	var list []string
	for name := range m.saidata {
		list = append(list, name)
	}
	return list, nil
}

func (m *MockSaidataManager) CacheData(software string, data *types.SoftwareData) error {
	m.saidata[software] = data
	return nil
}

func (m *MockSaidataManager) GetCachedData(software string) (*types.SoftwareData, error) {
	return m.LoadSoftware(software)
}

// MockExecutor implements interfaces.GenericExecutor for testing
type MockExecutor struct {
	executeResult *interfaces.ExecutionResult
	executeError  error
	validateError error
	canExecute    bool
}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		executeResult: &interfaces.ExecutionResult{
			Success:  true,
			Output:   "Mock execution successful",
			Commands: []string{"mock-command"},
			ExitCode: 0,
			Duration: time.Millisecond * 100,
		},
		canExecute: true,
	}
}

func (m *MockExecutor) SetExecuteResult(result *interfaces.ExecutionResult) {
	m.executeResult = result
}

func (m *MockExecutor) SetExecuteError(err error) {
	m.executeError = err
}

func (m *MockExecutor) SetValidateError(err error) {
	m.validateError = err
}

func (m *MockExecutor) SetCanExecute(canExecute bool) {
	m.canExecute = canExecute
}

func (m *MockExecutor) Execute(ctx context.Context, provider *types.ProviderData, action string, software string, saidata *types.SoftwareData, options interfaces.ExecuteOptions) (*interfaces.ExecutionResult, error) {
	if m.executeError != nil {
		return &interfaces.ExecutionResult{
			Success:  false,
			Error:    m.executeError,
			ExitCode: 1,
		}, m.executeError
	}
	return m.executeResult, nil
}

func (m *MockExecutor) ValidateAction(provider *types.ProviderData, action string, software string, saidata *types.SoftwareData) error {
	return m.validateError
}

func (m *MockExecutor) ValidateResources(saidata *types.SoftwareData, action string) (*interfaces.ResourceValidationResult, error) {
	return &interfaces.ResourceValidationResult{
		Valid:      true,
		CanProceed: true,
	}, nil
}

func (m *MockExecutor) DryRun(ctx context.Context, provider *types.ProviderData, action string, software string, saidata *types.SoftwareData, options interfaces.ExecuteOptions) (*interfaces.ExecutionResult, error) {
	if m.executeError != nil {
		return &interfaces.ExecutionResult{
			Success:  false,
			Error:    m.executeError,
			ExitCode: 1,
		}, m.executeError
	}
	result := *m.executeResult
	result.Output = "Mock dry run: " + result.Output
	return &result, nil
}

func (m *MockExecutor) CanExecute(provider *types.ProviderData, action string, software string, saidata *types.SoftwareData) bool {
	return m.canExecute
}

func (m *MockExecutor) RenderTemplate(template string, saidata *types.SoftwareData, provider *types.ProviderData) (string, error) {
	return template, nil
}

func (m *MockExecutor) ExecuteCommand(ctx context.Context, command string, options interfaces.CommandOptions) (*interfaces.CommandResult, error) {
	return &interfaces.CommandResult{
		Command:  command,
		Output:   "Mock command output",
		ExitCode: 0,
		Duration: time.Millisecond * 10,
	}, nil
}

func (m *MockExecutor) ExecuteSteps(ctx context.Context, steps []types.Step, saidata *types.SoftwareData, provider *types.ProviderData, options interfaces.ExecuteOptions) (*interfaces.ExecutionResult, error) {
	commands := make([]string, len(steps))
	for i, step := range steps {
		commands[i] = step.Command
	}
	return &interfaces.ExecutionResult{
		Success:  true,
		Output:   "Mock steps execution",
		Commands: commands,
		ExitCode: 0,
		Duration: time.Millisecond * 200,
	}, nil
}

// MockResourceValidator implements interfaces.ResourceValidator for testing
type MockResourceValidator struct {
	files       map[string]bool
	services    map[string]bool
	commands    map[string]bool
	directories map[string]bool
	ports       map[int]bool
}

func NewMockResourceValidator() *MockResourceValidator {
	return &MockResourceValidator{
		files:       make(map[string]bool),
		services:    make(map[string]bool),
		commands:    make(map[string]bool),
		directories: make(map[string]bool),
		ports:       make(map[int]bool),
	}
}

func (m *MockResourceValidator) SetFileExists(path string, exists bool) {
	m.files[path] = exists
}

func (m *MockResourceValidator) SetServiceExists(service string, exists bool) {
	m.services[service] = exists
}

func (m *MockResourceValidator) SetCommandExists(command string, exists bool) {
	m.commands[command] = exists
}

func (m *MockResourceValidator) SetDirectoryExists(path string, exists bool) {
	m.directories[path] = exists
}

func (m *MockResourceValidator) SetPortAvailable(port int, available bool) {
	m.ports[port] = available
}

func (m *MockResourceValidator) ValidateFile(file types.File) bool {
	return m.files[file.Path]
}

func (m *MockResourceValidator) ValidateService(service types.Service) bool {
	return m.services[service.ServiceName]
}

func (m *MockResourceValidator) ValidateCommand(command types.Command) bool {
	return m.commands[command.Path]
}

func (m *MockResourceValidator) ValidateDirectory(directory types.Directory) bool {
	return m.directories[directory.Path]
}

func (m *MockResourceValidator) ValidatePort(port types.Port) bool {
	if port.Port <= 0 || port.Port > 65535 {
		return false
	}
	if available, exists := m.ports[port.Port]; exists {
		return available
	}
	return true // Default to available
}

func (m *MockResourceValidator) ValidateResources(saidata *types.SoftwareData, action string) (*interfaces.ResourceValidationResult, error) {
	result := &interfaces.ResourceValidationResult{
		Valid:      true,
		CanProceed: true,
	}

	// Check files
	for _, file := range saidata.Files {
		if !m.ValidateFile(file) {
			result.Valid = false
			result.MissingFiles = append(result.MissingFiles, file.Path)
		}
	}

	// Check services
	for _, service := range saidata.Services {
		if !m.ValidateService(service) {
			result.Valid = false
			result.MissingServices = append(result.MissingServices, service.ServiceName)
		}
	}

	// Check commands
	for _, command := range saidata.Commands {
		if !m.ValidateCommand(command) {
			result.Valid = false
			result.MissingCommands = append(result.MissingCommands, command.Path)
		}
	}

	// Check directories
	for _, directory := range saidata.Directories {
		if !m.ValidateDirectory(directory) {
			result.Valid = false
			result.MissingDirectories = append(result.MissingDirectories, directory.Path)
		}
	}

	// Check ports
	for _, port := range saidata.Ports {
		if !m.ValidatePort(port) {
			result.Warnings = append(result.Warnings, "Port "+string(rune(port.Port))+" is not available")
		}
	}

	// Info actions can proceed even with missing resources
	if action == "info" || action == "search" || action == "version" || action == "status" {
		result.CanProceed = true
	} else {
		result.CanProceed = result.Valid
	}

	return result, nil
}

func (m *MockResourceValidator) FileExists(path string) bool {
	return m.files[path]
}

func (m *MockResourceValidator) ServiceExists(service string) bool {
	return m.services[service]
}

func (m *MockResourceValidator) CommandExists(command string) bool {
	return m.commands[command]
}

func (m *MockResourceValidator) DirectoryExists(path string) bool {
	return m.directories[path]
}

// MockLogger implements interfaces.Logger for testing
type MockLogger struct {
	logs []LogEntry
}

type LogEntry struct {
	Level   string
	Message string
	Fields  map[string]interface{}
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		logs: make([]LogEntry, 0),
	}
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "debug", Message: msg, Fields: fieldsToMap(fields...)})
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "info", Message: msg, Fields: fieldsToMap(fields...)})
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "warn", Message: msg, Fields: fieldsToMap(fields...)})
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.logs = append(m.logs, LogEntry{Level: "error", Message: msg, Fields: fieldsToMap(fields...)})
}

func (m *MockLogger) GetLogs() []LogEntry {
	return m.logs
}

func (m *MockLogger) ClearLogs() {
	m.logs = make([]LogEntry, 0)
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr ||
			 strings.Contains(s, substr))))
}

func fieldsToMap(fields ...interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for i := 0; i < len(fields)-1; i += 2 {
		if key, ok := fields[i].(string); ok {
			result[key] = fields[i+1]
		}
	}
	return result
}

// Helper function to create test provider data
func CreateTestProvider(name, providerType string, platforms []string, actions map[string]types.Action) *types.ProviderData {
	return &types.ProviderData{
		Version: "1.0",
		Provider: types.ProviderInfo{
			Name:        name,
			DisplayName: name + " Provider",
			Type:        providerType,
			Platforms:   platforms,
			Priority:    50,
		},
		Actions: actions,
	}
}

// Helper function to create test saidata
func CreateTestSaidata(name string, packages []types.Package, services []types.Service) *types.SoftwareData {
	return &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name:        name,
			DisplayName: name + " Software",
			Description: "Test software for " + name,
		},
		Packages: packages,
		Services: services,
	}
}