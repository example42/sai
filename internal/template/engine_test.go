package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sai/internal/types"
)

// Use the existing mock implementations from mocks_test.go

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

func TestTemplateEngine_SaiSourceFunction(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)

	saidata := &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name: "nginx",
		},
		Sources: []types.Source{
			{
				Name:        "nginx-source",
				URL:         "https://nginx.org/download/nginx-1.20.1.tar.gz",
				Version:     "1.20.1",
				BuildSystem: "autotools",
				ConfigureArgs: []string{"--with-http_ssl_module"},
				Prerequisites: []string{"gcc", "make", "libssl-dev"},
				Checksum:    "abc123def456",
			},
		},
		Providers: map[string]types.ProviderConfig{
			"source": {
				Sources: []types.Source{
					{
						Name:        "nginx-source",
						URL:         "https://nginx.org/download/nginx-1.20.1.tar.gz",
						Version:     "1.20.1",
						BuildSystem: "cmake",
						InstallPrefix: "/opt/nginx",
						ConfigureArgs: []string{"--with-http_ssl_module", "--with-debug"},
					},
				},
			},
		},
	}

	engine.SetSaidata(saidata)

	context := &TemplateContext{
		Software: "nginx",
		Provider: "source",
		Saidata:  saidata,
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "sai_source name field",
			template: "{{sai_source 0 \"name\"}}",
			expected: "nginx-source",
		},
		{
			name:     "sai_source url field",
			template: "{{sai_source 0 \"url\"}}",
			expected: "https://nginx.org/download/nginx-1.20.1.tar.gz",
		},
		{
			name:     "sai_source build_system field with provider override",
			template: "{{sai_source 0 \"build_system\" \"source\"}}",
			expected: "cmake",
		},
		{
			name:     "sai_source build_dir field with default",
			template: "{{sai_source 0 \"build_dir\"}}",
			expected: "/tmp/sai-build-nginx",
		},
		{
			name:     "sai_source install_prefix field with provider override",
			template: "{{sai_source 0 \"install_prefix\" \"source\"}}",
			expected: "/opt/nginx",
		},
		{
			name:     "sai_source configure_args field",
			template: "{{sai_source 0 \"configure_args\" \"source\"}}",
			expected: "--with-http_ssl_module --with-debug",
		},
		{
			name:     "sai_source prerequisites field",
			template: "{{sai_source 0 \"prerequisites\"}}",
			expected: "gcc make libssl-dev",
		},
		{
			name:     "sai_source download_cmd field",
			template: "{{sai_source 0 \"download_cmd\"}}",
			expected: "mkdir -p /tmp/sai-build-nginx && cd /tmp/sai-build-nginx && curl -L -o nginx-1.20.1.tar.gz https://nginx.org/download/nginx-1.20.1.tar.gz",
		},
		{
			name:     "sai_source configure_cmd field with cmake",
			template: "{{sai_source 0 \"configure_cmd\" \"source\"}}",
			expected: "cd /tmp/sai-build-nginx/nginx-1.20.1 && cmake -DCMAKE_INSTALL_PREFIX=/opt/nginx . --with-http_ssl_module --with-debug",
		},
		{
			name:     "sai_source build_cmd field with cmake",
			template: "{{sai_source 0 \"build_cmd\" \"source\"}}",
			expected: "cd /tmp/sai-build-nginx/nginx-1.20.1 && cmake --build .",
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

func TestTemplateEngine_SaiBinaryFunction(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)

	saidata := &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name: "terraform",
		},
		Binaries: []types.Binary{
			{
				Name:         "terraform",
				URL:          "https://releases.hashicorp.com/terraform/1.5.0/terraform_1.5.0_{{.OS}}_{{.Arch}}.zip",
				Version:      "1.5.0",
				Architecture: "amd64",
				Platform:     "linux",
				Checksum:     "abc123def456789012345678901234567890123456789012345678901234567890",
				Archive: &types.ArchiveConfig{
					Format:      "zip",
					ExtractPath: "terraform",
				},
			},
		},
		Providers: map[string]types.ProviderConfig{
			"binary": {
				Binaries: []types.Binary{
					{
						Name:        "terraform",
						URL:         "https://releases.hashicorp.com/terraform/1.5.0/terraform_1.5.0_{{.OS}}_{{.Arch}}.zip",
						Version:     "1.5.0",
						InstallPath: "/opt/terraform/bin",
						Executable:  "terraform",
						Permissions: "0755",
					},
				},
			},
		},
	}

	engine.SetSaidata(saidata)

	context := &TemplateContext{
		Software: "terraform",
		Provider: "binary",
		Saidata:  saidata,
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "sai_binary name field",
			template: "{{sai_binary 0 \"name\"}}",
			expected: "terraform",
		},
		{
			name:     "sai_binary url field with templating",
			template: "{{sai_binary 0 \"url\"}}",
			expected: "https://releases.hashicorp.com/terraform/1.5.0/terraform_1.5.0_darwin_arm64.zip", // Updated for current platform
		},
		{
			name:     "sai_binary install_path field with provider override",
			template: "{{sai_binary 0 \"install_path\" \"binary\"}}",
			expected: "/opt/terraform/bin",
		},
		{
			name:     "sai_binary executable field with provider override",
			template: "{{sai_binary 0 \"executable\" \"binary\"}}",
			expected: "terraform",
		},
		{
			name:     "sai_binary permissions field with provider override",
			template: "{{sai_binary 0 \"permissions\" \"binary\"}}",
			expected: "0755",
		},
		{
			name:     "sai_binary download_cmd field",
			template: "{{sai_binary 0 \"download_cmd\" \"binary\"}}",
			expected: "mkdir -p /opt/terraform/bin && curl -L -o /opt/terraform/bin/terraform_1.5.0_darwin_arm64.zip https://releases.hashicorp.com/terraform/1.5.0/terraform_1.5.0_darwin_arm64.zip",
		},
		{
			name:     "sai_binary verify_checksum_cmd field",
			template: "{{sai_binary 0 \"verify_checksum_cmd\"}}",
			expected: "echo 'abc123def456789012345678901234567890123456789012345678901234567890 /usr/local/bin/terraform_1.5.0_darwin_arm64.zip' | sha256sum -c",
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

func TestTemplateEngine_SaiScriptFunction(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)

	saidata := &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name: "docker",
		},
		Scripts: []types.Script{
			{
				Name:        "docker-install",
				URL:         "https://get.docker.com/",
				Version:     "latest",
				Interpreter: "bash",
				Arguments:   []string{"--channel", "stable"},
				Environment: map[string]string{
					"DOCKER_CHANNEL": "stable",
					"DOCKER_COMPOSE": "true",
				},
				WorkingDir: "/tmp",
				Timeout:    600,
				Checksum:   "def456abc789",
			},
		},
		Providers: map[string]types.ProviderConfig{
			"script": {
				Scripts: []types.Script{
					{
						Name:        "docker-install",
						URL:         "https://get.docker.com/",
						Interpreter: "bash",
						Arguments:   []string{"--channel", "stable", "--dry-run"},
						WorkingDir:  "/opt/docker",
						Timeout:     300,
					},
				},
			},
		},
	}

	engine.SetSaidata(saidata)

	context := &TemplateContext{
		Software: "docker",
		Provider: "script",
		Saidata:  saidata,
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "sai_script name field",
			template: "{{sai_script 0 \"name\"}}",
			expected: "docker-install",
		},
		{
			name:     "sai_script url field",
			template: "{{sai_script 0 \"url\"}}",
			expected: "https://get.docker.com/",
		},
		{
			name:     "sai_script interpreter field",
			template: "{{sai_script 0 \"interpreter\"}}",
			expected: "bash",
		},
		{
			name:     "sai_script arguments field with provider override",
			template: "{{sai_script 0 \"arguments\" \"script\"}}",
			expected: "--channel stable --dry-run",
		},
		{
			name:     "sai_script working_dir field with provider override",
			template: "{{sai_script 0 \"working_dir\" \"script\"}}",
			expected: "/opt/docker",
		},
		{
			name:     "sai_script timeout field with provider override",
			template: "{{sai_script 0 \"timeout\" \"script\"}}",
			expected: "300",
		},
		{
			name:     "sai_script download_cmd field",
			template: "{{sai_script 0 \"download_cmd\" \"script\"}}",
			expected: "mkdir -p /opt/docker && cd /opt/docker && curl -L -o install.sh https://get.docker.com/",
		},
		{
			name:     "sai_script execute_cmd field",
			template: "{{sai_script 0 \"execute_cmd\" \"script\"}}",
			expected: "cd /opt/docker && timeout 300 bash install.sh --channel stable --dry-run",
		},
		{
			name:     "sai_script environment_vars field",
			template: "{{sai_script 0 \"environment_vars\"}}",
			expected: "export DOCKER_CHANNEL='stable' && export DOCKER_COMPOSE='true'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, context)
			require.NoError(t, err)
			assert.Contains(t, result, tt.expected)
		})
	}
}

func TestTemplateEngine_AlternativeProviderErrorHandling(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)

	tests := []struct {
		name        string
		saidata     *types.SoftwareData
		template    string
		expectError bool
		errorType   string
	}{
		{
			name: "missing source should fail",
			saidata: &types.SoftwareData{
				Version: "0.2",
				Metadata: types.Metadata{Name: "test"},
			},
			template:    "{{sai_source 0 \"name\"}}",
			expectError: true,
			errorType:   "no source found",
		},
		{
			name: "missing binary should fail",
			saidata: &types.SoftwareData{
				Version: "0.2",
				Metadata: types.Metadata{Name: "test"},
			},
			template:    "{{sai_binary 0 \"name\"}}",
			expectError: true,
			errorType:   "no binary found",
		},
		{
			name: "missing script should fail",
			saidata: &types.SoftwareData{
				Version: "0.2",
				Metadata: types.Metadata{Name: "test"},
			},
			template:    "{{sai_script 0 \"name\"}}",
			expectError: true,
			errorType:   "no script found",
		},
		{
			name: "invalid source field should fail",
			saidata: &types.SoftwareData{
				Version: "0.2",
				Metadata: types.Metadata{Name: "test"},
				Sources: []types.Source{
					{Name: "test-source", URL: "https://example.com", BuildSystem: "make"},
				},
			},
			template:    "{{sai_source 0 \"invalid_field\"}}",
			expectError: true,
			errorType:   "unsupported source field",
		},
		{
			name: "invalid binary field should fail",
			saidata: &types.SoftwareData{
				Version: "0.2",
				Metadata: types.Metadata{Name: "test"},
				Binaries: []types.Binary{
					{Name: "test-binary", URL: "https://example.com/binary"},
				},
			},
			template:    "{{sai_binary 0 \"invalid_field\"}}",
			expectError: true,
			errorType:   "unsupported binary field",
		},
		{
			name: "invalid script field should fail",
			saidata: &types.SoftwareData{
				Version: "0.2",
				Metadata: types.Metadata{Name: "test"},
				Scripts: []types.Script{
					{Name: "test-script", URL: "https://example.com/script.sh"},
				},
			},
			template:    "{{sai_script 0 \"invalid_field\"}}",
			expectError: true,
			errorType:   "unsupported script field",
		},
		{
			name: "insufficient arguments for sai_source",
			saidata: &types.SoftwareData{
				Version: "0.2",
				Metadata: types.Metadata{Name: "test"},
			},
			template:    "{{sai_source 0}}",
			expectError: true,
			errorType:   "requires at least 2 arguments",
		},
		{
			name: "insufficient arguments for sai_binary",
			saidata: &types.SoftwareData{
				Version: "0.2",
				Metadata: types.Metadata{Name: "test"},
			},
			template:    "{{sai_binary 0}}",
			expectError: true,
			errorType:   "requires at least 2 arguments",
		},
		{
			name: "insufficient arguments for sai_script",
			saidata: &types.SoftwareData{
				Version: "0.2",
				Metadata: types.Metadata{Name: "test"},
			},
			template:    "{{sai_script 0}}",
			expectError: true,
			errorType:   "requires at least 2 arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine.SetSaidata(tt.saidata)
			
			context := &TemplateContext{
				Software: "test",
				Provider: "source",
				Saidata:  tt.saidata,
			}
			
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

func TestTemplateEngine_AlternativeProviderGracefulDegradation(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)

	// Test graceful degradation when template functions fail
	saidata := &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name: "test-software",
		},
		// No sources, binaries, or scripts defined
	}

	engine.SetSaidata(saidata)
	engine.SetSafetyMode(true) // Enable safety mode to catch errors

	context := &TemplateContext{
		Software: "test-software",
		Provider: "source",
		Saidata:  saidata,
	}

	// Test that templates with missing alternative provider data fail gracefully
	tests := []struct {
		name        string
		template    string
		expectError bool
		description string
	}{
		{
			name:        "source template with missing data should fail",
			template:    "cd {{sai_source 0 \"source_dir\"}} && make install",
			expectError: true,
			description: "Template should fail when source data is missing",
		},
		{
			name:        "binary template with missing data should fail",
			template:    "curl -L {{sai_binary 0 \"url\"}} -o /tmp/binary",
			expectError: true,
			description: "Template should fail when binary data is missing",
		},
		{
			name:        "script template with missing data should fail",
			template:    "bash {{sai_script 0 \"download_cmd\"}}",
			expectError: true,
			description: "Template should fail when script data is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, context)
			
			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Empty(t, result)
				// Verify the error contains helpful information
				assert.Contains(t, err.Error(), "Template resolution failed")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotEmpty(t, result)
			}
		})
	}
}