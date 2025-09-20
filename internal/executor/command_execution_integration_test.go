package executor

import (
	"context"
	"strings"
	"testing"
	"time"

	"sai/internal/interfaces"
	"sai/internal/template"
	"sai/internal/types"
)

func TestCommandExecutionErrorHandling(t *testing.T) {
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

	// Test case 1: Missing saidata should cause template resolution to fail
	t.Run("Missing saidata context", func(t *testing.T) {
		providerData := &types.ProviderData{
			Version: "1.0",
			Provider: types.ProviderInfo{
				Name:       "apt",
				Type:       "package_manager",
				Executable: "apt-get",
			},
			Actions: map[string]types.Action{
				"install": {
					Description: "Install packages via APT",
					Template:    "apt-get install -y {{sai_package \"*\" \"name\" \"apt\"}}",
				},
			},
		}

		// No saidata provided - should fail template resolution
		err := genericExecutor.ValidateAction(providerData, "install", "apache", nil)
		if err == nil {
			t.Error("Expected validation to fail with missing saidata, but it passed")
		}

		t.Logf("Correctly failed with missing saidata: %v", err)
	})

	// Test case 2: Invalid template function calls should be handled gracefully
	t.Run("Invalid template function", func(t *testing.T) {
		providerData := &types.ProviderData{
			Version: "1.0",
			Provider: types.ProviderInfo{
				Name:       "apt",
				Type:       "package_manager",
				Executable: "apt-get",
			},
			Actions: map[string]types.Action{
				"install": {
					Description: "Install packages via APT",
					Template:    "apt-get install -y {{sai_nonexistent_function}}",
				},
			},
		}

		saidata := &types.SoftwareData{
			Version: "0.2",
			Metadata: types.Metadata{Name: "apache"},
		}

		err := genericExecutor.ValidateAction(providerData, "install", "apache", saidata)
		if err == nil {
			t.Error("Expected validation to fail with invalid template function, but it passed")
		}

		t.Logf("Correctly failed with invalid template function: %v", err)
	})

	// Test case 3: Missing package in saidata should produce error indicator
	t.Run("Missing package in saidata", func(t *testing.T) {
		providerData := &types.ProviderData{
			Version: "1.0",
			Provider: types.ProviderInfo{
				Name:       "apt",
				Type:       "package_manager",
				Executable: "apt-get",
			},
			Actions: map[string]types.Action{
				"install": {
					Description: "Install packages via APT",
					Template:    "apt-get install -y {{sai_package \"*\" \"name\" \"apt\"}}",
				},
			},
		}

		// Saidata without packages - should produce error indicator
		saidata := &types.SoftwareData{
			Version: "0.2",
			Metadata: types.Metadata{Name: "apache"},
			// No packages defined
		}

		err := genericExecutor.ValidateAction(providerData, "install", "apache", saidata)
		if err == nil {
			t.Error("Expected validation to fail with missing packages, but it passed")
		}

		t.Logf("Correctly failed with missing packages: %v", err)
	})

	// Test case 4: Successful template resolution with proper saidata
	t.Run("Successful template resolution", func(t *testing.T) {
		providerData := &types.ProviderData{
			Version: "1.0",
			Provider: types.ProviderInfo{
				Name:       "apt",
				Type:       "package_manager",
				Executable: "apt-get",
			},
			Actions: map[string]types.Action{
				"install": {
					Description: "Install packages via APT",
					Template:    "apt-get install -y {{sai_package \"*\" \"name\" \"apt\"}}",
				},
			},
		}

		// Proper saidata with packages
		saidata := &types.SoftwareData{
			Version: "0.2",
			Metadata: types.Metadata{Name: "apache"},
			Packages: []types.Package{
				{
					Name:        "apache2",
					PackageName: "apache2",
				},
			},
			Providers: map[string]types.ProviderConfig{
				"apt": {
					Packages: []types.Package{
						{
							Name:        "apache2",
							PackageName: "apache2",
						},
					},
				},
			},
		}

		err := genericExecutor.ValidateAction(providerData, "install", "apache", saidata)
		if err != nil {
			t.Errorf("Expected validation to pass with proper saidata, but it failed: %v", err)
		}

		// Test dry run to verify command rendering
		ctx := context.Background()
		options := interfaces.ExecuteOptions{
			DryRun:  true,
			Verbose: true,
			Timeout: 30 * time.Second,
		}

		result, err := genericExecutor.DryRun(ctx, providerData, "install", "apache", saidata, options)
		if err != nil {
			t.Errorf("Dry run failed: %v", err)
			return
		}

		if !result.Success {
			t.Error("Dry run was not successful")
			return
		}

		if len(result.Commands) == 0 {
			t.Error("No commands generated")
			return
		}

		expectedCommand := "apt-get install -y apache2"
		if result.Commands[0] != expectedCommand {
			t.Errorf("Expected command: %s, got: %s", expectedCommand, result.Commands[0])
		}

		t.Logf("Successfully rendered command: %s", result.Commands[0])
	})

	// Test case 5: Test port template function with missing port
	t.Run("Missing port in saidata", func(t *testing.T) {
		providerData := &types.ProviderData{
			Version: "1.0",
			Provider: types.ProviderInfo{
				Name:       "docker",
				Type:       "container",
				Executable: "docker",
			},
			Actions: map[string]types.Action{
				"run": {
					Description: "Run Docker container",
					Template:    "docker run -p {{sai_port 0 \"port\" \"docker\"}}:80 nginx",
				},
			},
		}

		// Saidata without ports - should fail
		saidata := &types.SoftwareData{
			Version: "0.2",
			Metadata: types.Metadata{Name: "nginx"},
			// No ports defined
		}

		// Test dry run to see what gets rendered
		ctx := context.Background()
		options := interfaces.ExecuteOptions{
			DryRun:  true,
			Verbose: true,
			Timeout: 30 * time.Second,
		}

		result, err := genericExecutor.DryRun(ctx, providerData, "run", "nginx", saidata, options)
		if err != nil {
			t.Logf("Correctly failed with missing ports during dry run: %v", err)
		} else if result != nil && len(result.Commands) > 0 {
			t.Logf("Rendered command with missing ports: %s", result.Commands[0])
			if strings.Contains(result.Commands[0], "-1") {
				t.Log("Command contains -1 indicating port resolution failure")
			}
		}

		// Also test validation
		validationErr := genericExecutor.ValidateAction(providerData, "run", "nginx", saidata)
		if validationErr == nil {
			t.Error("Expected validation to fail with missing ports, but it passed")
		} else {
			t.Logf("Correctly failed with missing ports: %v", validationErr)
		}
	})
}

func TestCommandExecutionWithSteps(t *testing.T) {
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

	// Test multi-step action
	t.Run("Multi-step action with template resolution", func(t *testing.T) {
		providerData := &types.ProviderData{
			Version: "1.0",
			Provider: types.ProviderInfo{
				Name:       "apt",
				Type:       "package_manager",
				Executable: "apt-get",
			},
			Actions: map[string]types.Action{
				"install": {
					Description: "Install packages via APT",
					Steps: []types.Step{
						{
							Name:    "update-cache",
							Command: "apt-get update",
						},
						{
							Name:    "install-packages",
							Command: "apt-get install -y {{sai_package \"*\" \"name\" \"apt\"}}",
						},
					},
					Timeout: 600,
				},
			},
		}

		// Proper saidata with packages
		saidata := &types.SoftwareData{
			Version: "0.2",
			Metadata: types.Metadata{Name: "apache"},
			Packages: []types.Package{
				{
					Name:        "apache2",
					PackageName: "apache2",
				},
			},
			Providers: map[string]types.ProviderConfig{
				"apt": {
					Packages: []types.Package{
						{
							Name:        "apache2",
							PackageName: "apache2",
						},
					},
				},
			},
		}

		err := genericExecutor.ValidateAction(providerData, "install", "apache", saidata)
		if err != nil {
			t.Errorf("Expected validation to pass with proper saidata, but it failed: %v", err)
		}

		// Test dry run to verify step rendering
		ctx := context.Background()
		options := interfaces.ExecuteOptions{
			DryRun:  true,
			Verbose: true,
			Timeout: 30 * time.Second,
		}

		result, err := genericExecutor.DryRun(ctx, providerData, "install", "apache", saidata, options)
		if err != nil {
			t.Errorf("Dry run failed: %v", err)
			return
		}

		if !result.Success {
			t.Error("Dry run was not successful")
			return
		}

		if len(result.Commands) != 2 {
			t.Errorf("Expected 2 commands, got %d", len(result.Commands))
			return
		}

		expectedCommands := []string{
			"apt-get update",
			"apt-get install -y apache2",
		}

		for i, expected := range expectedCommands {
			if result.Commands[i] != expected {
				t.Errorf("Step %d: Expected command: %s, got: %s", i+1, expected, result.Commands[i])
			}
		}

		t.Logf("Successfully rendered multi-step commands: %v", result.Commands)
	})
}