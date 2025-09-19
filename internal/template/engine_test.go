package template

import (
	"testing"

	"sai/internal/types"
)

// MockResourceValidator for testing
type MockResourceValidator struct {
	files       map[string]bool
	services    map[string]bool
	commands    map[string]bool
	directories map[string]bool
}

func NewMockResourceValidator() *MockResourceValidator {
	return &MockResourceValidator{
		files:       make(map[string]bool),
		services:    make(map[string]bool),
		commands:    make(map[string]bool),
		directories: make(map[string]bool),
	}
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

// MockDefaultsGenerator for testing
type MockDefaultsGenerator struct{}

func NewMockDefaultsGenerator() *MockDefaultsGenerator {
	return &MockDefaultsGenerator{}
}

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

func TestNewTemplateEngine(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	
	engine := NewTemplateEngine(validator, defaultsGen)
	
	if engine == nil {
		t.Fatal("Expected template engine to be created")
	}
	
	if engine.validator != validator {
		t.Error("Expected validator to be set")
	}
	
	if engine.defaultsGen != defaultsGen {
		t.Error("Expected defaults generator to be set")
	}
	
	if !engine.safetyMode {
		t.Error("Expected safety mode to be enabled by default")
	}
}

func TestTemplateEngine_SetSafetyMode(t *testing.T) {
	engine := NewTemplateEngine(NewMockResourceValidator(), NewMockDefaultsGenerator())
	
	engine.SetSafetyMode(false)
	if engine.safetyMode {
		t.Error("Expected safety mode to be disabled")
	}
	
	engine.SetSafetyMode(true)
	if !engine.safetyMode {
		t.Error("Expected safety mode to be enabled")
	}
}

func TestTemplateEngine_ValidateTemplate(t *testing.T) {
	engine := NewTemplateEngine(NewMockResourceValidator(), NewMockDefaultsGenerator())
	
	tests := []struct {
		name        string
		template    string
		expectError bool
	}{
		{
			name:        "valid template",
			template:    "echo {{.Software}}",
			expectError: false,
		},
		{
			name:        "template with function",
			template:    "systemctl start {{sai_service \"main\"}}",
			expectError: false,
		},
		{
			name:        "invalid template syntax",
			template:    "echo {{.Software",
			expectError: true,
		},
		{
			name:        "empty template",
			template:    "",
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ValidateTemplate(tt.template)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestTemplateEngine_SaiPackage(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)
	
	// Create test saidata
	saidata := &types.SoftwareData{
		Packages: []types.Package{
			{Name: "nginx"},
			{Name: "nginx-common"},
		},
		Providers: map[string]types.ProviderConfig{
			"apt": {
				Packages: []types.Package{
					{Name: "nginx-full"},
				},
			},
		},
	}
	
	context := &TemplateContext{
		Software: "nginx",
		Provider: "apt",
		Saidata:  saidata,
	}
	
	tests := []struct {
		name        string
		template    string
		expected    string
		expectError bool
	}{
		{
			name:        "provider-specific package",
			template:    "{{sai_package \"apt\"}}",
			expected:    "nginx-full",
			expectError: false,
		},
		{
			name:        "default package",
			template:    "{{sai_package \"brew\"}}",
			expected:    "nginx",
			expectError: false,
		},
		{
			name:        "package with index",
			template:    "{{sai_package \"apt\" 0}}",
			expected:    "nginx-full",
			expectError: false,
		},
		{
			name:        "default package with index",
			template:    "{{sai_package \"brew\" 1}}",
			expected:    "nginx-common",
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, context)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %q but got %q", tt.expected, result)
			}
		})
	}
}

func TestTemplateEngine_SaiPackages(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)
	
	saidata := &types.SoftwareData{
		Packages: []types.Package{
			{Name: "nginx"},
			{Name: "nginx-common"},
		},
		Providers: map[string]types.ProviderConfig{
			"apt": {
				Packages: []types.Package{
					{Name: "nginx-full"},
					{Name: "nginx-core"},
				},
			},
		},
	}
	
	context := &TemplateContext{
		Software: "nginx",
		Provider: "apt",
		Saidata:  saidata,
	}
	
	tests := []struct {
		name        string
		template    string
		expected    string
		expectError bool
	}{
		{
			name:        "provider-specific packages",
			template:    "{{range sai_packages \"apt\"}}{{.}} {{end}}",
			expected:    "nginx-full nginx-core ",
			expectError: false,
		},
		{
			name:        "default packages",
			template:    "{{range sai_packages \"brew\"}}{{.}} {{end}}",
			expected:    "nginx nginx-common ",
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, context)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %q but got %q", tt.expected, result)
			}
		})
	}
}

func TestTemplateEngine_SaiService(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)
	
	saidata := &types.SoftwareData{
		Services: []types.Service{
			{Name: "main", ServiceName: "nginx"},
			{Name: "worker"}, // No service name, should use name
		},
	}
	
	context := &TemplateContext{
		Software: "nginx",
		Saidata:  saidata,
	}
	
	tests := []struct {
		name        string
		template    string
		expected    string
		expectError bool
	}{
		{
			name:        "service with explicit name",
			template:    "{{sai_service \"main\"}}",
			expected:    "nginx",
			expectError: false,
		},
		{
			name:        "service with default name",
			template:    "{{sai_service \"worker\"}}",
			expected:    "worker",
			expectError: false,
		},
		{
			name:        "non-existent service",
			template:    "{{sai_service \"nonexistent\"}}",
			expected:    "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, context)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %q but got %q", tt.expected, result)
			}
		})
	}
}

func TestTemplateEngine_SaiPort(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)
	
	saidata := &types.SoftwareData{
		Ports: []types.Port{
			{Port: 80, Protocol: "tcp"},
			{Port: 443, Protocol: "tcp"},
		},
	}
	
	context := &TemplateContext{
		Software: "nginx",
		Saidata:  saidata,
	}
	
	tests := []struct {
		name        string
		template    string
		expected    string
		expectError bool
	}{
		{
			name:        "default port",
			template:    "{{sai_port}}",
			expected:    "80",
			expectError: false,
		},
		{
			name:        "port with index",
			template:    "{{sai_port 1}}",
			expected:    "443",
			expectError: false,
		},
		{
			name:        "invalid port index",
			template:    "{{sai_port 5}}",
			expected:    "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, context)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %q but got %q", tt.expected, result)
			}
		})
	}
}

func TestTemplateEngine_SafetyValidation(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)
	
	// Set up some existing resources
	validator.SetFileExists("/etc/nginx/nginx.conf", true)
	validator.SetServiceExists("nginx", true)
	validator.SetCommandExists("nginx", true)
	validator.SetDirectoryExists("/var/log/nginx", true)
	
	saidata := &types.SoftwareData{
		Services: []types.Service{
			{Name: "main", ServiceName: "nginx"},
		},
		Files: []types.File{
			{Name: "config", Path: "/etc/nginx/nginx.conf"},
		},
		Commands: []types.Command{
			{Name: "nginx", Path: "/usr/bin/nginx"},
		},
		Directories: []types.Directory{
			{Name: "logs", Path: "/var/log/nginx"},
		},
	}
	
	context := &TemplateContext{
		Software: "nginx",
		Saidata:  saidata,
	}
	
	tests := []struct {
		name        string
		template    string
		safetyMode  bool
		expected    string
		expectError bool
	}{
		{
			name:        "file exists check - true",
			template:    "{{file_exists \"/etc/nginx/nginx.conf\"}}",
			safetyMode:  true,
			expected:    "true",
			expectError: false,
		},
		{
			name:        "file exists check - false",
			template:    "{{file_exists \"/nonexistent\"}}",
			safetyMode:  true,
			expected:    "false",
			expectError: false,
		},
		{
			name:        "service exists check - true",
			template:    "{{service_exists \"nginx\"}}",
			safetyMode:  true,
			expected:    "true",
			expectError: false,
		},
		{
			name:        "command exists check - true",
			template:    "{{command_exists \"nginx\"}}",
			safetyMode:  true,
			expected:    "true",
			expectError: false,
		},
		{
			name:        "directory exists check - true",
			template:    "{{directory_exists \"/var/log/nginx\"}}",
			safetyMode:  true,
			expected:    "true",
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine.SetSafetyMode(tt.safetyMode)
			result, err := engine.Render(tt.template, context)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %q but got %q", tt.expected, result)
			}
		})
	}
}

func TestTemplateEngine_DefaultFunctions(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)
	
	context := &TemplateContext{
		Software: "nginx",
	}
	
	tests := []struct {
		name        string
		template    string
		expected    string
		expectError bool
	}{
		{
			name:        "default config path",
			template:    "{{default_config_path \"nginx\"}}",
			expected:    "/etc/nginx/nginx.conf",
			expectError: false,
		},
		{
			name:        "default log path",
			template:    "{{default_log_path \"nginx\"}}",
			expected:    "/var/log/nginx.log",
			expectError: false,
		},
		{
			name:        "default data dir",
			template:    "{{default_data_dir \"nginx\"}}",
			expected:    "/var/lib/nginx",
			expectError: false,
		},
		{
			name:        "default service name",
			template:    "{{default_service_name \"nginx\"}}",
			expected:    "nginx",
			expectError: false,
		},
		{
			name:        "default command path",
			template:    "{{default_command_path \"nginx\"}}",
			expected:    "/usr/bin/nginx",
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, context)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected %q but got %q", tt.expected, result)
			}
		})
	}
}

func TestTemplateEngine_ComplexTemplate(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)
	
	saidata := &types.SoftwareData{
		Packages: []types.Package{
			{Name: "nginx"},
		},
		Services: []types.Service{
			{Name: "main", ServiceName: "nginx"},
		},
		Ports: []types.Port{
			{Port: 80},
		},
	}
	
	context := &TemplateContext{
		Software: "nginx",
		Provider: "apt",
		Saidata:  saidata,
		Variables: map[string]string{
			"action": "install",
		},
	}
	
	template := `apt {{.Variables.action}} {{sai_package "apt"}} && systemctl start {{sai_service "main"}} && echo "Started on port {{sai_port}}"`
	expected := "apt install nginx && systemctl start nginx && echo \"Started on port 80\""
	
	result, err := engine.Render(template, context)
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
	
	if result != expected {
		t.Errorf("Expected %q but got %q", expected, result)
	}
}