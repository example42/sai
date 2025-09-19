package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sai/internal/types"
)

// Mock implementations for testing
type mockResourceValidator struct {
	files       map[string]bool
	services    map[string]bool
	commands    map[string]bool
	directories map[string]bool
}

func (m *mockResourceValidator) FileExists(path string) bool {
	return m.files[path]
}

func (m *mockResourceValidator) ServiceExists(service string) bool {
	return m.services[service]
}

func (m *mockResourceValidator) CommandExists(command string) bool {
	return m.commands[command]
}

func (m *mockResourceValidator) DirectoryExists(path string) bool {
	return m.directories[path]
}

type mockDefaultsGenerator struct{}

func (m *mockDefaultsGenerator) DefaultConfigPath(software string) string {
	return "/etc/" + software + "/" + software + ".conf"
}

func (m *mockDefaultsGenerator) DefaultLogPath(software string) string {
	return "/var/log/" + software + ".log"
}

func (m *mockDefaultsGenerator) DefaultDataDir(software string) string {
	return "/var/lib/" + software
}

func (m *mockDefaultsGenerator) DefaultServiceName(software string) string {
	return software
}

func (m *mockDefaultsGenerator) DefaultCommandPath(software string) string {
	return "/usr/bin/" + software
}

func TestNewTemplateEngine(t *testing.T) {
	validator := &mockResourceValidator{
		files:       make(map[string]bool),
		services:    make(map[string]bool),
		commands:    make(map[string]bool),
		directories: make(map[string]bool),
	}
	defaultsGen := &mockDefaultsGenerator{}

	engine := NewTemplateEngine(validator, defaultsGen)
	assert.NotNil(t, engine)
	assert.True(t, engine.safetyMode) // Should be enabled by default
}

func TestTemplateEngine_BasicRendering(t *testing.T) {
	validator := &mockResourceValidator{
		files:       make(map[string]bool),
		services:    make(map[string]bool),
		commands:    make(map[string]bool),
		directories: make(map[string]bool),
	}
	defaultsGen := &mockDefaultsGenerator{}
	engine := NewTemplateEngine(validator, defaultsGen)

	// Test basic template rendering
	context := &TemplateContext{
		Software: "nginx",
		Provider: "apt",
		Variables: map[string]string{
			"version": "1.20.1",
		},
	}

	result, err := engine.Render("install {{.Software}} version {{.Variables.version}}", context)
	require.NoError(t, err)
	assert.Equal(t, "install nginx version 1.20.1", result)
}

func TestTemplateEngine_SaiPackageFunction(t *testing.T) {
	validator := &mockResourceValidator{
		files:       make(map[string]bool),
		services:    make(map[string]bool),
		commands:    make(map[string]bool),
		directories: make(map[string]bool),
	}
	defaultsGen := &mockDefaultsGenerator{}
	engine := NewTemplateEngine(validator, defaultsGen)

	// Set up saidata with packages
	saidata := &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name: "apache",
		},
		Packages: []types.Package{
			{Name: "apache2", Version: "2.4.58"},
			{Name: "apache2-utils", Version: "2.4.58"},
		},
		Providers: map[string]types.ProviderConfig{
			"apt": {
				Packages: []types.Package{
					{Name: "apache2", Version: "2.4.58-1ubuntu1"},
					{Name: "apache2-utils", Version: "2.4.58-1ubuntu1"},
				},
			},
			"brew": {
				Packages: []types.Package{
					{Name: "httpd", Version: "2.4.58"},
				},
			},
		},
	}

	engine.SetSaidata(saidata)

	context := &TemplateContext{
		Software: "apache",
		Provider: "apt",
		Saidata:  saidata,
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "sai_package with provider and index",
			template: "{{sai_package \"apt\" 0}}",
			expected: "apache2",
		},
		{
			name:     "sai_package with provider wildcard",
			template: "{{sai_package \"apt\" \"*\"}}",
			expected: "apache2 apache2-utils",
		},
		{
			name:     "sai_package with different provider",
			template: "{{sai_package \"brew\" 0}}",
			expected: "httpd",
		},
		{
			name:     "sai_packages function",
			template: "{{sai_packages \"apt\"}}",
			expected: "apache2 apache2-utils",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, context)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTemplateEngine_SaiServiceFunction(t *testing.T) {
	validator := &mockResourceValidator{
		files:       make(map[string]bool),
		services:    make(map[string]bool),
		commands:    make(map[string]bool),
		directories: make(map[string]bool),
	}
	defaultsGen := &mockDefaultsGenerator{}
	engine := NewTemplateEngine(validator, defaultsGen)

	saidata := &types.SoftwareData{
		Version: "0.2",
		Services: []types.Service{
			{Name: "apache", ServiceName: "apache2", Type: "systemd"},
			{Name: "apache-ssl", ServiceName: "apache2-ssl", Type: "systemd"},
		},
		Providers: map[string]types.ProviderConfig{
			"apt": {
				Services: []types.Service{
					{Name: "apache", ServiceName: "apache2", Type: "systemd"},
				},
			},
			"brew": {
				Services: []types.Service{
					{Name: "apache", ServiceName: "httpd", Type: "launchd"},
				},
			},
		},
	}

	engine.SetSaidata(saidata)

	context := &TemplateContext{
		Software: "apache",
		Provider: "apt",
		Saidata:  saidata,
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "sai_service with name",
			template: "{{sai_service \"apache\"}}",
			expected: "apache2",
		},
		{
			name:     "sai_service with different provider",
			template: "{{sai_service \"apache\"}}",
			expected: "apache2", // Should use apt provider from context
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, context)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTemplateEngine_SaiFileFunction(t *testing.T) {
	validator := &mockResourceValidator{
		files:       make(map[string]bool),
		services:    make(map[string]bool),
		commands:    make(map[string]bool),
		directories: make(map[string]bool),
	}
	defaultsGen := &mockDefaultsGenerator{}
	engine := NewTemplateEngine(validator, defaultsGen)

	saidata := &types.SoftwareData{
		Version: "0.2",
		Files: []types.File{
			{Name: "config", Path: "/etc/apache2/apache2.conf", Type: "config"},
			{Name: "log", Path: "/var/log/apache2/access.log", Type: "log"},
		},
		Providers: map[string]types.ProviderConfig{
			"apt": {
				Files: []types.File{
					{Name: "config", Path: "/etc/apache2/apache2.conf", Type: "config"},
				},
			},
			"brew": {
				Files: []types.File{
					{Name: "config", Path: "/opt/homebrew/etc/httpd/httpd.conf", Type: "config"},
				},
			},
		},
	}

	engine.SetSaidata(saidata)

	context := &TemplateContext{
		Software: "apache",
		Provider: "apt",
		Saidata:  saidata,
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "sai_file with name",
			template: "{{sai_file \"config\"}}",
			expected: "/etc/apache2/apache2.conf",
		},
		{
			name:     "sai_file with log",
			template: "{{sai_file \"log\"}}",
			expected: "/var/log/apache2/access.log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, context)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTemplateEngine_SaiPortFunction(t *testing.T) {
	validator := &mockResourceValidator{
		files:       make(map[string]bool),
		services:    make(map[string]bool),
		commands:    make(map[string]bool),
		directories: make(map[string]bool),
	}
	defaultsGen := &mockDefaultsGenerator{}
	engine := NewTemplateEngine(validator, defaultsGen)

	saidata := &types.SoftwareData{
		Version: "0.2",
		Ports: []types.Port{
			{Port: 80, Protocol: "tcp", Service: "http"},
			{Port: 443, Protocol: "tcp", Service: "https"},
		},
	}

	engine.SetSaidata(saidata)

	context := &TemplateContext{
		Software: "apache",
		Provider: "apt",
		Saidata:  saidata,
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "sai_port default (first port)",
			template: "{{sai_port}}",
			expected: "80",
		},
		{
			name:     "sai_port with index",
			template: "{{sai_port 1}}",
			expected: "443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, context)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTemplateEngine_ValidationFunctions(t *testing.T) {
	validator := &mockResourceValidator{
		files: map[string]bool{
			"/etc/apache2/apache2.conf": true,
			"/var/log/apache2/access.log": false,
		},
		services: map[string]bool{
			"apache2": true,
			"nginx":   false,
		},
		commands: map[string]bool{
			"apache2": true,
			"nginx":   false,
		},
		directories: map[string]bool{
			"/etc/apache2": true,
			"/etc/nginx":   false,
		},
	}
	defaultsGen := &mockDefaultsGenerator{}
	engine := NewTemplateEngine(validator, defaultsGen)

	context := &TemplateContext{
		Software: "apache",
		Provider: "apt",
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "file_exists true",
			template: "{{file_exists \"/etc/apache2/apache2.conf\"}}",
			expected: "true",
		},
		{
			name:     "file_exists false",
			template: "{{file_exists \"/var/log/apache2/access.log\"}}",
			expected: "false",
		},
		{
			name:     "service_exists true",
			template: "{{service_exists \"apache2\"}}",
			expected: "true",
		},
		{
			name:     "service_exists false",
			template: "{{service_exists \"nginx\"}}",
			expected: "false",
		},
		{
			name:     "command_exists true",
			template: "{{command_exists \"apache2\"}}",
			expected: "true",
		},
		{
			name:     "command_exists false",
			template: "{{command_exists \"nginx\"}}",
			expected: "false",
		},
		{
			name:     "directory_exists true",
			template: "{{directory_exists \"/etc/apache2\"}}",
			expected: "true",
		},
		{
			name:     "directory_exists false",
			template: "{{directory_exists \"/etc/nginx\"}}",
			expected: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, context)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTemplateEngine_DefaultGenerationFunctions(t *testing.T) {
	validator := &mockResourceValidator{
		files:       make(map[string]bool),
		services:    make(map[string]bool),
		commands:    make(map[string]bool),
		directories: make(map[string]bool),
	}
	defaultsGen := &mockDefaultsGenerator{}
	engine := NewTemplateEngine(validator, defaultsGen)

	context := &TemplateContext{
		Software: "nginx",
		Provider: "apt",
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "default_config_path",
			template: "{{default_config_path \"nginx\"}}",
			expected: "/etc/nginx/nginx.conf",
		},
		{
			name:     "default_log_path",
			template: "{{default_log_path \"nginx\"}}",
			expected: "/var/log/nginx.log",
		},
		{
			name:     "default_data_dir",
			template: "{{default_data_dir \"nginx\"}}",
			expected: "/var/lib/nginx",
		},
		{
			name:     "default_service_name",
			template: "{{default_service_name \"nginx\"}}",
			expected: "nginx",
		},
		{
			name:     "default_command_path",
			template: "{{default_command_path \"nginx\"}}",
			expected: "/usr/bin/nginx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, context)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTemplateEngine_SafetyMode(t *testing.T) {
	validator := &mockResourceValidator{
		files: map[string]bool{
			"/etc/apache2/apache2.conf": true,
			"/nonexistent/file.conf":    false,
		},
		services: map[string]bool{
			"apache2":     true,
			"nonexistent": false,
		},
		commands:    make(map[string]bool),
		directories: make(map[string]bool),
	}
	defaultsGen := &mockDefaultsGenerator{}
	engine := NewTemplateEngine(validator, defaultsGen)

	saidata := &types.SoftwareData{
		Version: "0.2",
		Files: []types.File{
			{Name: "config", Path: "/etc/apache2/apache2.conf", Type: "config"},
			{Name: "nonexistent", Path: "/nonexistent/file.conf", Type: "config"},
		},
		Services: []types.Service{
			{Name: "apache", ServiceName: "apache2", Type: "systemd"},
			{Name: "nonexistent", ServiceName: "nonexistent", Type: "systemd"},
		},
	}

	engine.SetSaidata(saidata)
	engine.SetSafetyMode(true)

	context := &TemplateContext{
		Software: "apache",
		Provider: "apt",
		Saidata:  saidata,
	}

	tests := []struct {
		name        string
		template    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "existing file should work",
			template:    "cat {{sai_file \"config\"}}",
			expectError: false,
		},
		{
			name:        "nonexistent file should fail in safety mode",
			template:    "cat {{sai_file \"nonexistent\"}}",
			expectError: true,
			errorMsg:    "file does not exist",
		},
		{
			name:        "existing service should work",
			template:    "systemctl start {{sai_service \"apache\"}}",
			expectError: false,
		},
		{
			name:        "nonexistent service should fail in safety mode",
			template:    "systemctl start {{sai_service \"nonexistent\"}}",
			expectError: true,
			errorMsg:    "service does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, context)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestTemplateEngine_ValidateTemplate(t *testing.T) {
	validator := &mockResourceValidator{
		files:       make(map[string]bool),
		services:    make(map[string]bool),
		commands:    make(map[string]bool),
		directories: make(map[string]bool),
	}
	defaultsGen := &mockDefaultsGenerator{}
	engine := NewTemplateEngine(validator, defaultsGen)

	tests := []struct {
		name        string
		template    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid template",
			template:    "install {{.Software}}",
			expectError: false,
		},
		{
			name:        "valid template with sai functions",
			template:    "install {{sai_package \"apt\" 0}}",
			expectError: false,
		},
		{
			name:        "invalid template syntax",
			template:    "install {{.Software",
			expectError: true,
			errorMsg:    "template syntax error",
		},
		{
			name:        "template with unknown function",
			template:    "install {{unknown_function}}",
			expectError: true,
			errorMsg:    "function \"unknown_function\" not defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ValidateTemplate(tt.template)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTemplateEngine_ComplexTemplate(t *testing.T) {
	validator := &mockResourceValidator{
		files: map[string]bool{
			"/etc/apache2/apache2.conf": true,
		},
		services: map[string]bool{
			"apache2": true,
		},
		commands:    make(map[string]bool),
		directories: make(map[string]bool),
	}
	defaultsGen := &mockDefaultsGenerator{}
	engine := NewTemplateEngine(validator, defaultsGen)

	saidata := &types.SoftwareData{
		Version: "0.2",
		Packages: []types.Package{
			{Name: "apache2", Version: "2.4.58"},
		},
		Services: []types.Service{
			{Name: "apache", ServiceName: "apache2", Type: "systemd"},
		},
		Files: []types.File{
			{Name: "config", Path: "/etc/apache2/apache2.conf", Type: "config"},
		},
		Providers: map[string]types.ProviderConfig{
			"apt": {
				Packages: []types.Package{
					{Name: "apache2", Version: "2.4.58-1ubuntu1"},
				},
			},
		},
	}

	engine.SetSaidata(saidata)

	context := &TemplateContext{
		Software: "apache",
		Provider: "apt",
		Saidata:  saidata,
		Variables: map[string]string{
			"extra_flags": "--enable-ssl",
		},
	}

	// Complex template with multiple functions and conditionals
	complexTemplate := `#!/bin/bash
# Install {{.Software}} using {{.Provider}}
if {{file_exists "{{sai_file \"config\"}}"}}; then
    echo "Backing up existing config"
    cp {{sai_file "config"}} {{sai_file "config"}}.backup
fi

apt-get update
apt-get install -y {{sai_package "apt" 0}} {{.Variables.extra_flags}}

if {{service_exists "{{sai_service \"apache\"}}"}}}; then
    systemctl enable {{sai_service "apache"}}
    systemctl start {{sai_service "apache"}}
fi

echo "Installation complete"`

	result, err := engine.Render(complexTemplate, context)
	require.NoError(t, err)
	
	// Verify the template was rendered correctly
	assert.Contains(t, result, "Install apache using apt")
	assert.Contains(t, result, "apt-get install -y apache2 --enable-ssl")
	assert.Contains(t, result, "systemctl enable apache2")
	assert.Contains(t, result, "systemctl start apache2")
	assert.Contains(t, result, "/etc/apache2/apache2.conf")
}