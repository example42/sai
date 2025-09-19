package template

import (
	"testing"

	"sai/internal/types"
)

func TestTemplateResolutionValidator_ValidateActionTemplate(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)
	resolutionValidator := NewTemplateResolutionValidator(engine)
	
	// Set up some existing resources
	validator.SetFileExists("/etc/nginx/nginx.conf", true)
	validator.SetServiceExists("nginx", true)
	validator.SetCommandExists("nginx", true)
	validator.SetCommandExists("/usr/bin/nginx", true)
	validator.SetDirectoryExists("/var/log/nginx", true)
	
	saidata := &types.SoftwareData{
		Packages: []types.Package{
			{Name: "nginx"},
		},
		Services: []types.Service{
			{Name: "main", ServiceName: "nginx"},
		},
		Files: []types.File{
			{Name: "config", Path: "/etc/nginx/nginx.conf"},
		},
		Commands: []types.Command{
			{Name: "nginx", Path: "/usr/bin/nginx"},
		},
	}
	
	tests := []struct {
		name           string
		action         *types.Action
		software       string
		provider       string
		expectValid    bool
		expectExecute  bool
		expectResolved bool
	}{
		{
			name: "valid action with resolvable template",
			action: &types.Action{
				Template: "apt install {{sai_package \"apt\"}}",
			},
			software:       "nginx",
			provider:       "apt",
			expectValid:    true,
			expectExecute:  true,
			expectResolved: true,
		},
		{
			name: "action with unresolvable sai function",
			action: &types.Action{
				Template: "systemctl start {{sai_service \"nonexistent\"}}",
			},
			software:       "nginx",
			provider:       "apt",
			expectValid:    true,
			expectExecute:  false,
			expectResolved: false,
		},
		{
			name: "action with invalid template syntax",
			action: &types.Action{
				Template: "apt install {{sai_package \"apt\"",
			},
			software:       "nginx",
			provider:       "apt",
			expectValid:    false,
			expectExecute:  false,
			expectResolved: false,
		},
		{
			name: "nil action",
			action:         nil,
			software:       "nginx",
			provider:       "apt",
			expectValid:    false,
			expectExecute:  false,
			expectResolved: false,
		},
		{
			name: "action with no command",
			action: &types.Action{
				Description: "Test action",
			},
			software:       "nginx",
			provider:       "apt",
			expectValid:    false,
			expectExecute:  false,
			expectResolved: false,
		},
		{
			name: "action with service template and existing service",
			action: &types.Action{
				Template: "systemctl start {{sai_service \"main\"}}",
			},
			software:       "nginx",
			provider:       "systemd",
			expectValid:    true,
			expectExecute:  true,
			expectResolved: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For basic template resolution tests, disable safety mode to focus on template logic
			if tt.name == "valid action with resolvable template" || tt.name == "action with service template and existing service" {
				resolutionValidator.SetSafetyMode(SafetyModeDisabled)
			} else {
				resolutionValidator.SetSafetyMode(SafetyModeStrict)
			}
			
			result := resolutionValidator.ValidateActionTemplate(tt.action, tt.software, tt.provider, saidata)
			
			if result.Valid != tt.expectValid {
				t.Errorf("Expected Valid=%v, got %v", tt.expectValid, result.Valid)
			}
			
			if result.CanExecute != tt.expectExecute {
				t.Errorf("Expected CanExecute=%v, got %v", tt.expectExecute, result.CanExecute)
			}
			
			if result.Resolvable != tt.expectResolved {
				t.Errorf("Expected Resolvable=%v, got %v", tt.expectResolved, result.Resolvable)
			}
			
			if !tt.expectValid && len(result.Errors) == 0 {
				t.Error("Expected errors for invalid action")
			}
			
			if !tt.expectResolved && len(result.UnresolvedVariables) == 0 && len(result.Errors) == 0 {
				t.Error("Expected unresolved variables or errors for unresolvable template")
			}
		})
	}
}

func TestTemplateResolutionValidator_ValidateProviderActions(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)
	resolutionValidator := NewTemplateResolutionValidator(engine)
	
	// Set up existing resources
	validator.SetServiceExists("nginx", true)
	
	provider := &types.ProviderData{
		Provider: types.ProviderInfo{
			Name: "apt",
		},
		Actions: map[string]types.Action{
			"install": {
				Template: "apt install {{sai_package \"apt\"}}",
			},
			"start": {
				Template: "systemctl start {{sai_service \"main\"}}",
			},
			"invalid": {
				Template: "systemctl start {{sai_service \"nonexistent\"}}",
			},
		},
	}
	
	saidata := &types.SoftwareData{
		Packages: []types.Package{
			{Name: "nginx"},
		},
		Services: []types.Service{
			{Name: "main", ServiceName: "nginx"},
		},
	}
	
	results := resolutionValidator.ValidateProviderActions(provider, "nginx", saidata)
	
	if len(results) != 3 {
		t.Errorf("Expected 3 validation results, got %d", len(results))
	}
	
	// Check install action (should be executable)
	if installResult, exists := results["install"]; exists {
		if !installResult.CanExecute {
			t.Error("Expected install action to be executable")
		}
	} else {
		t.Error("Expected install action result")
	}
	
	// Check start action (should be executable)
	if startResult, exists := results["start"]; exists {
		if !startResult.CanExecute {
			t.Error("Expected start action to be executable")
		}
	} else {
		t.Error("Expected start action result")
	}
	
	// Check invalid action (should not be executable)
	if invalidResult, exists := results["invalid"]; exists {
		if invalidResult.CanExecute {
			t.Error("Expected invalid action to not be executable")
		}
	} else {
		t.Error("Expected invalid action result")
	}
}

func TestTemplateResolutionValidator_GetExecutableActions(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)
	resolutionValidator := NewTemplateResolutionValidator(engine)
	
	// Set up existing resources
	validator.SetServiceExists("nginx", true)
	
	provider := &types.ProviderData{
		Provider: types.ProviderInfo{
			Name: "apt",
		},
		Actions: map[string]types.Action{
			"install": {
				Template: "apt install {{sai_package \"apt\"}}",
			},
			"start": {
				Template: "systemctl start {{sai_service \"main\"}}",
			},
			"invalid": {
				Template: "systemctl start {{sai_service \"nonexistent\"}}",
			},
		},
	}
	
	saidata := &types.SoftwareData{
		Packages: []types.Package{
			{Name: "nginx"},
		},
		Services: []types.Service{
			{Name: "main", ServiceName: "nginx"},
		},
	}
	
	executableActions := resolutionValidator.GetExecutableActions(provider, "nginx", saidata)
	
	// Should have 2 executable actions (install and start)
	if len(executableActions) != 2 {
		t.Errorf("Expected 2 executable actions, got %d: %v", len(executableActions), executableActions)
	}
	
	// Check that the correct actions are executable
	expectedActions := map[string]bool{
		"install": false,
		"start":   false,
	}
	
	for _, action := range executableActions {
		if _, exists := expectedActions[action]; exists {
			expectedActions[action] = true
		} else {
			t.Errorf("Unexpected executable action: %s", action)
		}
	}
	
	for action, found := range expectedActions {
		if !found {
			t.Errorf("Expected action %s to be executable", action)
		}
	}
}

func TestTemplateResolutionValidator_FindUnresolvedVariables(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)
	resolutionValidator := NewTemplateResolutionValidator(engine)
	
	tests := []struct {
		name     string
		rendered string
		expected []string
	}{
		{
			name:     "no unresolved variables",
			rendered: "apt install nginx",
			expected: []string{},
		},
		{
			name:     "single unresolved variable",
			rendered: "apt install {{.Package}}",
			expected: []string{"{{.Package}}"},
		},
		{
			name:     "multiple unresolved variables",
			rendered: "{{.Command}} {{.Action}} {{.Package}}",
			expected: []string{"{{.Command}}", "{{.Action}}", "{{.Package}}"},
		},
		{
			name:     "no value indicator",
			rendered: "apt install <no value>",
			expected: []string{"<no value>"},
		},
		{
			name:     "mixed unresolved",
			rendered: "{{.Command}} install <no value>",
			expected: []string{"{{.Command}}", "<no value>"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolutionValidator.findUnresolvedVariables(tt.rendered)
			
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d unresolved variables, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			
			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected unresolved variable %s at index %d, got %v", expected, i, result)
				}
			}
		})
	}
}

func TestTemplateResolutionValidator_ValidateResourceExistence(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)
	resolutionValidator := NewTemplateResolutionValidator(engine)
	
	// Set up some existing resources
	validator.SetFileExists("/etc/nginx/nginx.conf", true)
	validator.SetServiceExists("nginx", true)
	validator.SetCommandExists("/usr/bin/nginx", true)
	validator.SetDirectoryExists("/var/log/nginx", true)
	
	saidata := &types.SoftwareData{
		Files: []types.File{
			{Name: "config", Path: "/etc/nginx/nginx.conf"},
			{Name: "missing", Path: "/nonexistent/file"},
		},
		Services: []types.Service{
			{Name: "main", ServiceName: "nginx"},
			{Name: "missing", ServiceName: "nonexistent"},
		},
		Commands: []types.Command{
			{Name: "nginx", Path: "/usr/bin/nginx"},
			{Name: "missing", Path: "/nonexistent/command"},
		},
		Directories: []types.Directory{
			{Name: "logs", Path: "/var/log/nginx"},
			{Name: "missing", Path: "/nonexistent/directory"},
		},
	}
	
	// Test with service action
	serviceAction := &types.Action{
		Template: "systemctl start nginx",
	}
	
	result := resolutionValidator.validateResourceExistence(saidata, serviceAction)
	
	// Should find missing resources
	if result.CanExecute {
		t.Error("Expected action to not be executable due to missing resources")
	}
	
	if len(result.MissingResources) == 0 {
		t.Error("Expected to find missing resources")
	}
	
	// Should find missing file, service, command, and directory
	expectedMissing := 4
	if len(result.MissingResources) != expectedMissing {
		t.Errorf("Expected %d missing resources, got %d: %v", expectedMissing, len(result.MissingResources), result.MissingResources)
	}
}

func TestTemplateResolutionValidator_IsServiceAction(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)
	resolutionValidator := NewTemplateResolutionValidator(engine)
	
	tests := []struct {
		name     string
		action   *types.Action
		expected bool
	}{
		{
			name: "systemctl action",
			action: &types.Action{
				Template: "systemctl start nginx",
			},
			expected: true,
		},
		{
			name: "service action",
			action: &types.Action{
				Template: "service nginx start",
			},
			expected: true,
		},
		{
			name: "launchctl action",
			action: &types.Action{
				Template: "launchctl start nginx",
			},
			expected: true,
		},
		{
			name: "install action",
			action: &types.Action{
				Template: "apt install nginx",
			},
			expected: false,
		},
		{
			name: "nil action",
			action: nil,
			expected: false,
		},
		{
			name: "action with no command",
			action: &types.Action{
				Description: "Test action",
			},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolutionValidator.isServiceAction(tt.action)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTemplateResolutionValidator_GetActionAvailability(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)
	resolutionValidator := NewTemplateResolutionValidator(engine)
	
	// Set up existing resources
	validator.SetServiceExists("nginx", true)
	
	provider := &types.ProviderData{
		Provider: types.ProviderInfo{
			Name: "apt",
		},
		Actions: map[string]types.Action{
			"install": {
				Template: "apt install {{sai_package \"apt\"}}",
			},
			"invalid": {
				Template: "systemctl start {{sai_service \"nonexistent\"}}",
			},
		},
	}
	
	saidata := &types.SoftwareData{
		Packages: []types.Package{
			{Name: "nginx"},
		},
		Services: []types.Service{
			{Name: "main", ServiceName: "nginx"},
		},
	}
	
	availability := resolutionValidator.GetActionAvailability(provider, "nginx", saidata)
	
	if len(availability) != 2 {
		t.Errorf("Expected 2 availability results, got %d", len(availability))
	}
	
	// Check availability results
	availabilityMap := make(map[string]*ActionAvailability)
	for _, avail := range availability {
		availabilityMap[avail.ActionName] = avail
	}
	
	// Install action should be available
	if installAvail, exists := availabilityMap["install"]; exists {
		if !installAvail.Available {
			t.Errorf("Expected install action to be available, reason: %s", installAvail.Reason)
		}
	} else {
		t.Error("Expected install action availability result")
	}
	
	// Invalid action should not be available
	if invalidAvail, exists := availabilityMap["invalid"]; exists {
		if invalidAvail.Available {
			t.Error("Expected invalid action to not be available")
		}
		if invalidAvail.Reason == "" {
			t.Error("Expected reason for unavailable action")
		}
	} else {
		t.Error("Expected invalid action availability result")
	}
}

func TestTemplateResolutionValidator_SetSafetyMode(t *testing.T) {
	validator := NewMockResourceValidator()
	defaultsGen := NewMockDefaultsGenerator()
	engine := NewTemplateEngine(validator, defaultsGen)
	resolutionValidator := NewTemplateResolutionValidator(engine)
	
	// Test disabled mode
	resolutionValidator.SetSafetyMode(SafetyModeDisabled)
	if engine.safetyMode {
		t.Error("Expected safety mode to be disabled")
	}
	
	// Test warning mode
	resolutionValidator.SetSafetyMode(SafetyModeWarning)
	if !engine.safetyMode {
		t.Error("Expected safety mode to be enabled for warning mode")
	}
	
	// Test strict mode
	resolutionValidator.SetSafetyMode(SafetyModeStrict)
	if !engine.safetyMode {
		t.Error("Expected safety mode to be enabled for strict mode")
	}
}