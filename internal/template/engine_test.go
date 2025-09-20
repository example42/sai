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
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()

	engine := NewTemplateEngine(validator, defaultsGen)
	assert.NotNil(t, engine)
	assert.True(t, engine.safetyMode) // Should be enabled by default
}

func TestTemplateEngine_BasicRendering(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
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
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)

	// Set up saidata with packages using both name and package_name fields
	saidata := &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name: "apache",
		},
		Packages: []types.Package{
			{Name: "apache2", PackageName: "apache2-server", Version: "2.4.58"},
			{Name: "apache2-utils", PackageName: "apache2-utils", Version: "2.4.58"},
		},
		Providers: map[string]types.ProviderConfig{
			"apt": {
				Packages: []types.Package{
					{Name: "apache2", PackageName: "apache2-deb", Version: "2.4.58-1ubuntu1"},
					{Name: "apache2-utils", PackageName: "apache2-utils", Version: "2.4.58-1ubuntu1"},
				},
			},
			"brew": {
				Packages: []types.Package{
					{Name: "httpd", PackageName: "httpd-brew", Version: "2.4.58"},
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
			name:     "sai_package with provider (new format)",
			template: "{{sai_package \"apt\"}}",
			expected: "apache2-deb",
		},
		{
			name:     "sai_package with provider and index (new format)",
			template: "{{sai_package \"apt\" 1}}",
			expected: "apache2-utils",
		},
		{
			name:     "sai_package legacy format - single package (name field)",
			template: "{{sai_package 0 \"name\" \"apt\"}}",
			expected: "apache2",
		},
		{
			name:     "sai_package legacy format - single package (package_name field)",
			template: "{{sai_package 0 \"package_name\" \"apt\"}}",
			expected: "apache2-deb",
		},
		{
			name:     "sai_package legacy format - all packages (package_name field)",
			template: "{{sai_package \"*\" \"package_name\" \"apt\"}}",
			expected: "apache2-deb apache2-utils",
		},
		{
			name:     "sai_package with different provider",
			template: "{{sai_package \"brew\"}}",
			expected: "httpd-brew",
		},
		{
			name:     "sai_packages function returns slice",
			template: "{{range sai_packages \"apt\"}}{{.}} {{end}}",
			expected: "apache2-deb apache2-utils ",
		},
		{
			name:     "fallback to name when package_name not available",
			template: "{{sai_package \"apt\" 1}}",
			expected: "apache2-utils",
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
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
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
			name:     "sai_service with name (new format)",
			template: "{{sai_service \"apache\"}}",
			expected: "apache2",
		},
		{
			name:     "sai_service legacy format",
			template: "{{sai_service 0 \"service_name\" \"apt\"}}",
			expected: "apache2",
		},
		{
			name:     "sai_service with different provider context",
			template: "{{sai_service 0 \"service_name\" \"brew\"}}",
			expected: "httpd",
		},
		{
			name:     "sai_service fallback to name when service_name empty",
			template: "{{sai_service \"apache-ssl\"}}",
			expected: "apache2-ssl",
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
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
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
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
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
	validator := NewMockResourceValidator()
	validator.SetFileExists("/etc/apache2/apache2.conf", true)
	validator.SetFileExists("/var/log/apache2/access.log", false)
	validator.SetServiceExists("apache2", true)
	validator.SetServiceExists("nginx", false)
	validator.SetCommandExists("apache2", true)
	validator.SetCommandExists("nginx", false)
	validator.SetDirectoryExists("/etc/apache2", true)
	validator.SetDirectoryExists("/etc/nginx", false)
	
	defaultsGen := NewMockDefaultsGenerator()
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
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
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
	validator := NewMockResourceValidator()
	validator.SetFileExists("/etc/apache2/apache2.conf", true)
	validator.SetFileExists("/nonexistent/file.conf", false)
	validator.SetServiceExists("apache2", true)
	validator.SetServiceExists("nonexistent", false)
	
	defaultsGen := NewMockDefaultsGenerator()
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
			name:        "nonexistent file should be detected in safety mode",
			template:    "cat {{sai_file \"nonexistent\"}}",
			expectError: true,
			errorMsg:    "nonexistent file",
		},
		{
			name:        "existing service should work",
			template:    "systemctl start {{sai_service \"apache\"}}",
			expectError: false,
		},
		{
			name:        "nonexistent service should be detected in safety mode",
			template:    "systemctl start {{sai_service \"nonexistent\"}}",
			expectError: true,
			errorMsg:    "nonexistent service",
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
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
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
	validator := NewMockResourceValidator()
	validator.SetFileExists("/etc/apache2/apache2.conf", true)
	validator.SetServiceExists("apache2", true)
	
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)

	saidata := &types.SoftwareData{
		Version: "0.2",
		Packages: []types.Package{
			{Name: "apache2", PackageName: "apache2-server", Version: "2.4.58"},
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
					{Name: "apache2", PackageName: "apache2-deb", Version: "2.4.58-1ubuntu1"},
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
apt-get install -y {{sai_package "apt"}} {{.Variables.extra_flags}}

if {{service_exists "{{sai_service \"apache\"}}"}}}; then
    systemctl enable {{sai_service "apache"}}
    systemctl start {{sai_service "apache"}}
fi

echo "Installation complete"`

	result, err := engine.Render(complexTemplate, context)
	require.NoError(t, err)
	
	// Verify the template was rendered correctly
	assert.Contains(t, result, "Install apache using apt")
	assert.Contains(t, result, "apt-get install -y apache2-deb --enable-ssl")
	assert.Contains(t, result, "systemctl enable apache2")
	assert.Contains(t, result, "systemctl start apache2")
	assert.Contains(t, result, "/etc/apache2/apache2.conf")
}

func TestTemplateEngine_ErrorHandling(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)

	// Test with minimal saidata
	saidata := &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name: "test",
		},
	}

	engine.SetSaidata(saidata)

	context := &TemplateContext{
		Software: "test",
		Provider: "apt",
		Saidata:  saidata,
	}

	tests := []struct {
		name        string
		template    string
		expectError bool
		errorType   string
	}{
		{
			name:        "missing package should fail",
			template:    "{{sai_package \"apt\"}}",
			expectError: true,
			errorType:   "no package found",
		},
		{
			name:        "missing service should fail",
			template:    "{{sai_service \"nonexistent\"}}",
			expectError: true,
			errorType:   "service nonexistent not found",
		},
		{
			name:        "invalid function parameters",
			template:    "{{sai_package}}",
			expectError: true,
			errorType:   "requires at least one argument",
		},
		{
			name:        "legacy format with missing data",
			template:    "{{sai_package 0 \"package_name\" \"apt\"}}",
			expectError: true,
			errorType:   "no package found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, context)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorType)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestTemplateEngine_WithExistingSaidataFiles(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)

	// Test with Apache saidata structure (mimicking the actual file)
	saidata := &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name: "apache",
		},
		Packages: []types.Package{
			{Name: "apache2", PackageName: "apache2", Version: "2.4.58"},
		},
		Services: []types.Service{
			{Name: "apache", ServiceName: "apache2", Type: "systemd"},
		},
		Files: []types.File{
			{Name: "config", Path: "/etc/apache2/apache2.conf", Type: "config"},
		},
		Ports: []types.Port{
			{Port: 80, Protocol: "tcp", Service: "http"},
			{Port: 443, Protocol: "tcp", Service: "https"},
		},
		Providers: map[string]types.ProviderConfig{
			"apt": {
				Packages: []types.Package{
					{Name: "apache2", PackageName: "apache2", Version: "2.4.58-1ubuntu1"},
				},
				Services: []types.Service{
					{Name: "apache", ServiceName: "apache2", Type: "systemd"},
				},
			},
			"brew": {
				Packages: []types.Package{
					{Name: "httpd", PackageName: "httpd", Version: "2.4.58"},
				},
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

	// Test templates that match actual provider files
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "apt install template (updated to use package_name)",
			template: "apt-get install -y {{sai_package \"*\" \"package_name\" \"apt\"}}",
			expected: "apt-get install -y apache2",
		},
		{
			name:     "systemctl start template",
			template: "systemctl start {{sai_service 0 \"service_name\" \"apt\"}}",
			expected: "systemctl start apache2",
		},
		{
			name:     "brew install template (updated to use package_name)",
			template: "brew install {{sai_package 0 \"package_name\" \"brew\"}}",
			expected: "brew install httpd",
		},
		{
			name:     "port reference template",
			template: "netstat -ln | grep :{{sai_port 0 \"port\" \"\"}}",
			expected: "netstat -ln | grep :80",
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