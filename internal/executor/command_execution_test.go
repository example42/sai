package executor

import (
	"context"
	"testing"
	"time"

	"sai/internal/interfaces"
	"sai/internal/template"
	"sai/internal/types"
)

// MockTemplateResourceValidator implements the template engine's ResourceValidator interface
type MockTemplateResourceValidator struct{}

func (m *MockTemplateResourceValidator) FileExists(path string) bool { return true }
func (m *MockTemplateResourceValidator) ServiceExists(service string) bool { return true }
func (m *MockTemplateResourceValidator) CommandExists(command string) bool { return true }
func (m *MockTemplateResourceValidator) DirectoryExists(path string) bool { return true }

// MockDefaultsGenerator implements the DefaultsGenerator interface for testing
type MockDefaultsGenerator struct{}

func (m *MockDefaultsGenerator) GeneratePackageDefaults(software string) []types.Package { return nil }
func (m *MockDefaultsGenerator) GenerateServiceDefaults(software string) []types.Service { return nil }
func (m *MockDefaultsGenerator) GenerateFileDefaults(software string) []types.File { return nil }
func (m *MockDefaultsGenerator) GenerateDirectoryDefaults(software string) []types.Directory { return nil }
func (m *MockDefaultsGenerator) GenerateCommandDefaults(software string) []types.Command { return nil }
func (m *MockDefaultsGenerator) GeneratePortDefaults(software string) []types.Port { return nil }
func (m *MockDefaultsGenerator) ValidatePathExists(path string) bool { return true }
func (m *MockDefaultsGenerator) ValidateServiceExists(service string) bool { return true }
func (m *MockDefaultsGenerator) ValidateCommandExists(command string) bool { return true }
func (m *MockDefaultsGenerator) GenerateDefaults(software string) (*types.SoftwareData, error) { return nil, nil }
func (m *MockDefaultsGenerator) DefaultConfigPath(software string) string { return "/etc/" + software + "/" + software + ".conf" }
func (m *MockDefaultsGenerator) DefaultLogPath(software string) string { return "/var/log/" + software + ".log" }
func (m *MockDefaultsGenerator) DefaultDataDir(software string) string { return "/var/lib/" + software }
func (m *MockDefaultsGenerator) DefaultServiceName(software string) string { return software }
func (m *MockDefaultsGenerator) DefaultCommandPath(software string) string { return "/usr/bin/" + software }

func TestCommandExecutionWithTemplateResolution(t *testing.T) {
	// Create mock implementations
	testLogger := &MockLogger{}
	validator := &MockResourceValidator{}
	templateValidator := &MockTemplateResourceValidator{}
	defaultsGen := &MockDefaultsGenerator{}

	// Create template engine
	templateEngine := template.NewTemplateEngine(templateValidator, defaultsGen)

	// Create command executor
	commandExecutor := NewCommandExecutor(testLogger, validator)

	// Create generic executor
	genericExecutor := NewGenericExecutor(commandExecutor, templateEngine, testLogger, validator)

	// Create test saidata that matches the Apache sample structure
	saidata := &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name:        "apache",
			DisplayName: "Apache HTTP Server",
			Description: "The Apache HTTP Server",
		},
		Packages: []types.Package{
			{
				Name:        "apache2",
				PackageName: "apache2", // Use the new package_name field
				Version:     "2.4.58",
			},
		},
		Services: []types.Service{
			{
				Name:        "apache",
				ServiceName: "apache2",
				Type:        "systemd",
				Enabled:     true,
			},
		},
		Ports: []types.Port{
			{
				Port:     80,
				Protocol: "tcp",
				Service:  "http",
			},
		},
		Containers: []types.Container{
			{
				Name:     "apache-httpd",
				Image:    "httpd",
				Tag:      "2.4",
				Registry: "docker.io",
			},
		},
		Providers: map[string]types.ProviderConfig{
			"apt": {
				Packages: []types.Package{
					{
						Name:        "apache2",
						PackageName: "apache2",
						Version:     "2.4.58-1ubuntu1",
					},
				},
				Services: []types.Service{
					{
						Name:        "apache",
						ServiceName: "apache2",
						Type:        "systemd",
					},
				},
			},
			"docker": {
				Containers: []types.Container{
					{
						Name:  "apache-httpd",
						Image: "httpd",
						Tag:   "2.4-alpine",
					},
				},
			},
		},
	}

	// Create test provider data that matches apt.yaml structure
	providerData := &types.ProviderData{
		Version: "1.0",
		Provider: types.ProviderInfo{
			Name:        "apt",
			DisplayName: "Advanced Package Tool",
			Type:        "package_manager",
			Platforms:   []string{"debian", "ubuntu"},
			Executable:  "apt-get",
		},
		Actions: map[string]types.Action{
			"install": {
				Description: "Install packages via APT",
				Template:    "apt-get install -y {{sai_package \"*\" \"name\" \"apt\"}}",
				Timeout:     600,
			},
			"start": {
				Description: "Start service via systemctl",
				Template:    "systemctl start {{sai_service 0 \"service_name\" \"apt\"}}",
			},
			"status": {
				Description: "Check service status",
				Template:    "systemctl status {{sai_service 0 \"service_name\" \"apt\"}}",
			},
		},
	}

	tests := []struct {
		name           string
		action         string
		expectedResult string
		shouldFail     bool
	}{
		{
			name:           "Install action with package template",
			action:         "install",
			expectedResult: "apt-get install -y apache2",
			shouldFail:     false,
		},
		{
			name:           "Start action with service template",
			action:         "start",
			expectedResult: "systemctl start apache2",
			shouldFail:     false,
		},
		{
			name:           "Status action with service template",
			action:         "status",
			expectedResult: "systemctl status apache2",
			shouldFail:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test template validation
			err := genericExecutor.ValidateAction(providerData, tt.action, "apache", saidata)
			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected validation to fail for action %s, but it passed", tt.action)
				}
				return
			}

			if err != nil {
				t.Errorf("Template validation failed for action %s: %v", tt.action, err)
				return
			}

			// Test dry run to see rendered command
			ctx := context.Background()
			options := interfaces.ExecuteOptions{
				DryRun:  true,
				Verbose: true,
				Timeout: 30 * time.Second,
			}

			result, err := genericExecutor.DryRun(ctx, providerData, tt.action, "apache", saidata, options)
			if err != nil {
				t.Errorf("Dry run failed for action %s: %v", tt.action, err)
				return
			}

			if !result.Success {
				t.Errorf("Dry run was not successful for action %s", tt.action)
				return
			}

			if len(result.Commands) == 0 {
				t.Errorf("No commands generated for action %s", tt.action)
				return
			}

			// Check if the rendered command matches expected result
			renderedCommand := result.Commands[0]
			if renderedCommand != tt.expectedResult {
				t.Errorf("Template rendering mismatch for action %s:\nExpected: %s\nGot: %s", 
					tt.action, tt.expectedResult, renderedCommand)
			}

			t.Logf("Action %s rendered successfully: %s", tt.action, renderedCommand)
		})
	}
}

func TestTemplateEngineWithDockerProvider(t *testing.T) {
	// Create mock implementations
	testLogger := &MockLogger{}
	validator := &MockResourceValidator{}
	templateValidator := &MockTemplateResourceValidator{}
	defaultsGen := &MockDefaultsGenerator{}

	// Create template engine
	templateEngine := template.NewTemplateEngine(templateValidator, defaultsGen)

	// Create command executor
	commandExecutor := NewCommandExecutor(testLogger, validator)

	// Create generic executor
	genericExecutor := NewGenericExecutor(commandExecutor, templateEngine, testLogger, validator)

	// Create test saidata for Docker
	saidata := &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name: "apache",
		},
		Containers: []types.Container{
			{
				Name:     "apache-httpd",
				Image:    "httpd",
				Tag:      "2.4",
				Registry: "docker.io",
			},
		},
		Ports: []types.Port{
			{
				Port:     80,
				Protocol: "tcp",
			},
		},
		Providers: map[string]types.ProviderConfig{
			"docker": {
				Containers: []types.Container{
					{
						Name:  "apache-httpd",
						Image: "httpd",
						Tag:   "2.4-alpine",
					},
				},
			},
		},
	}

	// Create Docker provider data
	providerData := &types.ProviderData{
		Version: "1.0",
		Provider: types.ProviderInfo{
			Name:       "docker",
			Type:       "container",
			Platforms:  []string{"linux", "macos", "windows"},
			Executable: "docker",
		},
		Actions: map[string]types.Action{
			"start": {
				Description: "Start Docker container",
				Template:    "docker start {{sai_container 0 \"name\" \"docker\"}}",
			},
			"logs": {
				Description: "Show Docker container logs",
				Template:    "docker logs --tail 50 {{sai_container 0 \"name\" \"docker\"}}",
			},
		},
	}

	tests := []struct {
		name           string
		action         string
		expectedResult string
	}{
		{
			name:           "Docker start with container name",
			action:         "start",
			expectedResult: "docker start apache-httpd",
		},
		{
			name:           "Docker logs with container name",
			action:         "logs",
			expectedResult: "docker logs --tail 50 apache-httpd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test template validation
			err := genericExecutor.ValidateAction(providerData, tt.action, "apache", saidata)
			if err != nil {
				t.Errorf("Template validation failed for action %s: %v", tt.action, err)
				return
			}

			// Test dry run
			ctx := context.Background()
			options := interfaces.ExecuteOptions{
				DryRun:  true,
				Verbose: true,
				Timeout: 30 * time.Second,
			}

			result, err := genericExecutor.DryRun(ctx, providerData, tt.action, "apache", saidata, options)
			if err != nil {
				t.Errorf("Dry run failed for action %s: %v", tt.action, err)
				return
			}

			if len(result.Commands) == 0 {
				t.Errorf("No commands generated for action %s", tt.action)
				return
			}

			renderedCommand := result.Commands[0]
			if renderedCommand != tt.expectedResult {
				t.Errorf("Template rendering mismatch for action %s:\nExpected: %s\nGot: %s", 
					tt.action, tt.expectedResult, renderedCommand)
			}

			t.Logf("Docker action %s rendered successfully: %s", tt.action, renderedCommand)
		})
	}
}