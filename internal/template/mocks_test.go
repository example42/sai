package template

import (
	"sai/internal/interfaces"
	"sai/internal/types"
)

// MockResourceValidator for testing
type MockResourceValidator struct {
	fileExists      map[string]bool
	serviceExists   map[string]bool
	commandExists   map[string]bool
	directoryExists map[string]bool
}

func NewMockResourceValidator() *MockResourceValidator {
	return &MockResourceValidator{
		fileExists:      make(map[string]bool),
		serviceExists:   make(map[string]bool),
		commandExists:   make(map[string]bool),
		directoryExists: make(map[string]bool),
	}
}

func (m *MockResourceValidator) SetFileExists(path string, exists bool) {
	m.fileExists[path] = exists
}

func (m *MockResourceValidator) SetServiceExists(name string, exists bool) {
	m.serviceExists[name] = exists
}

func (m *MockResourceValidator) SetCommandExists(name string, exists bool) {
	m.commandExists[name] = exists
}

func (m *MockResourceValidator) SetDirectoryExists(path string, exists bool) {
	m.directoryExists[path] = exists
}

// Template package ResourceValidator interface methods
func (m *MockResourceValidator) FileExists(path string) bool {
	if exists, ok := m.fileExists[path]; ok {
		return exists
	}
	return true
}

func (m *MockResourceValidator) ServiceExists(service string) bool {
	if exists, ok := m.serviceExists[service]; ok {
		return exists
	}
	return true
}

func (m *MockResourceValidator) CommandExists(command string) bool {
	if exists, ok := m.commandExists[command]; ok {
		return exists
	}
	return true
}

func (m *MockResourceValidator) DirectoryExists(path string) bool {
	if exists, ok := m.directoryExists[path]; ok {
		return exists
	}
	return true
}

// interfaces.ResourceValidator interface methods (for compatibility)
func (m *MockResourceValidator) ValidateFile(file types.File) bool {
	return m.FileExists(file.Path)
}

func (m *MockResourceValidator) ValidateService(service types.Service) bool {
	return m.ServiceExists(service.ServiceName) || m.ServiceExists(service.Name)
}

func (m *MockResourceValidator) ValidateCommand(command types.Command) bool {
	return m.CommandExists(command.Path) || m.CommandExists(command.Name)
}

func (m *MockResourceValidator) ValidateDirectory(directory types.Directory) bool {
	return m.DirectoryExists(directory.Path)
}

func (m *MockResourceValidator) ValidatePort(port types.Port) bool { return true }
func (m *MockResourceValidator) ValidateContainer(container types.Container) bool { return true }
func (m *MockResourceValidator) ValidateResources(saidata *types.SoftwareData) (*interfaces.ResourceValidationResult, error) {
	return &interfaces.ResourceValidationResult{Valid: true, CanProceed: true}, nil
}
func (m *MockResourceValidator) ValidateSystemRequirements(requirements *types.Requirements) (*interfaces.SystemValidationResult, error) {
	return &interfaces.SystemValidationResult{Valid: true}, nil
}

// MockDefaultsGenerator for testing
type MockDefaultsGenerator struct{}

func NewMockDefaultsGenerator() *MockDefaultsGenerator {
	return &MockDefaultsGenerator{}
}

func (m *MockDefaultsGenerator) GeneratePackageDefaults(software string) []types.Package {
	return []types.Package{{Name: software}}
}

func (m *MockDefaultsGenerator) GenerateServiceDefaults(software string) []types.Service {
	return []types.Service{{Name: software, ServiceName: software}}
}

func (m *MockDefaultsGenerator) GenerateFileDefaults(software string) []types.File {
	return []types.File{{Name: "config", Path: "/etc/" + software + "/" + software + ".conf"}}
}

func (m *MockDefaultsGenerator) GenerateDirectoryDefaults(software string) []types.Directory {
	return []types.Directory{{Name: "data", Path: "/var/lib/" + software}}
}

func (m *MockDefaultsGenerator) GenerateCommandDefaults(software string) []types.Command {
	return []types.Command{{Name: software, Path: "/usr/bin/" + software}}
}

func (m *MockDefaultsGenerator) GeneratePortDefaults(software string) []types.Port {
	return []types.Port{{Port: 8080}}
}

// Template package DefaultsGenerator interface methods
func (m *MockDefaultsGenerator) DefaultConfigPath(software string) string {
	return "/etc/" + software + "/" + software + ".conf"
}

func (m *MockDefaultsGenerator) DefaultLogPath(software string) string {
	return "/var/log/" + software + ".log"
}

func (m *MockDefaultsGenerator) DefaultDataDir(software string) string {
	return "/var/lib/" + software
}

func (m *MockDefaultsGenerator) DefaultServiceName(software string) string {
	return software
}

func (m *MockDefaultsGenerator) DefaultCommandPath(software string) string {
	return "/usr/bin/" + software
}

// interfaces.DefaultsGenerator interface methods (for compatibility)
func (m *MockDefaultsGenerator) ValidatePathExists(path string) bool { return true }
func (m *MockDefaultsGenerator) ValidateServiceExists(service string) bool { return true }
func (m *MockDefaultsGenerator) ValidateCommandExists(command string) bool { return true }
func (m *MockDefaultsGenerator) GenerateDefaults(software string) (*types.SoftwareData, error) {
	return &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{Name: software},
		Packages: m.GeneratePackageDefaults(software),
		Services: m.GenerateServiceDefaults(software),
		Files: m.GenerateFileDefaults(software),
		Directories: m.GenerateDirectoryDefaults(software),
		Commands: m.GenerateCommandDefaults(software),
		Ports: m.GeneratePortDefaults(software),
		IsGenerated: true,
	}, nil
}