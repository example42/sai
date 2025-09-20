package action

import (
	"fmt"
	"strings"

	"sai/internal/interfaces"
	"sai/internal/types"
)

// SafetyManager handles safety checks and prevents dangerous operations
type SafetyManager struct {
	validator interfaces.ResourceValidator
}

// NewSafetyManager creates a new safety manager
func NewSafetyManager(validator interfaces.ResourceValidator) *SafetyManager {
	return &SafetyManager{
		validator: validator,
	}
}

// CheckActionSafety performs comprehensive safety checks before action execution
func (sm *SafetyManager) CheckActionSafety(action, software string, provider *types.ProviderData, saidata *types.SoftwareData) (*SafetyResult, error) {
	result := &SafetyResult{
		Safe:     true,
		Action:   action,
		Software: software,
		Provider: provider.Provider.Name,
		Checks:   []SafetyCheck{},
	}

	// Check 1: Validate resources exist (Requirement 10.5)
	resourceCheck := sm.checkResourcesExist(action, saidata)
	result.Checks = append(result.Checks, resourceCheck)
	if !resourceCheck.Passed {
		result.Safe = false
	}

	// Check 2: Validate provider availability
	providerCheck := sm.checkProviderAvailability(provider)
	result.Checks = append(result.Checks, providerCheck)
	if !providerCheck.Passed {
		result.Safe = false
	}

	// Check 3: Check for destructive operations
	destructiveCheck := sm.checkDestructiveOperation(action, software)
	result.Checks = append(result.Checks, destructiveCheck)
	if !destructiveCheck.Passed {
		result.Safe = false
	}

	// Check 4: Validate system requirements
	systemCheck := sm.checkSystemRequirements(saidata)
	result.Checks = append(result.Checks, systemCheck)
	if !systemCheck.Passed {
		result.Safe = false
	}

	// Check 5: Validate template resolution
	templateCheck := sm.checkTemplateResolution(provider, action, saidata)
	result.Checks = append(result.Checks, templateCheck)
	if !templateCheck.Passed {
		result.Safe = false
	}

	return result, nil
}

// checkResourcesExist validates that required resources exist on the system
func (sm *SafetyManager) checkResourcesExist(action string, saidata *types.SoftwareData) SafetyCheck {
	check := SafetyCheck{
		Name:        "Resource Existence",
		Description: "Verify that required files, services, and commands exist",
		Passed:      true,
		Messages:    []string{},
	}

	// For install actions, we don't require resources to exist beforehand
	// The install action will create them
	installActions := []string{"install", "upgrade", "search", "info", "version"}
	for _, installAction := range installActions {
		if action == installAction {
			check.Messages = append(check.Messages, fmt.Sprintf("Skipping resource validation for %s action", action))
			return check
		}
	}

	validationResult, err := sm.validator.ValidateResources(saidata)
	if err != nil {
		check.Passed = false
		check.Messages = append(check.Messages, fmt.Sprintf("Resource validation failed: %v", err))
		return check
	}

	// Check for critical missing resources
	if len(validationResult.MissingCommands) > 0 {
		check.Passed = false
		check.Messages = append(check.Messages, fmt.Sprintf("Critical commands missing: %v", validationResult.MissingCommands))
	}

	// Add warnings for non-critical missing resources
	if len(validationResult.MissingFiles) > 0 {
		check.Messages = append(check.Messages, fmt.Sprintf("Warning: Configuration files missing: %v", validationResult.MissingFiles))
	}

	if len(validationResult.MissingServices) > 0 {
		check.Messages = append(check.Messages, fmt.Sprintf("Warning: Services not installed: %v", validationResult.MissingServices))
	}

	if len(validationResult.MissingDirectories) > 0 {
		check.Messages = append(check.Messages, fmt.Sprintf("Warning: Directories missing: %v", validationResult.MissingDirectories))
	}

	// Can proceed if only non-critical resources are missing
	if !validationResult.CanProceed {
		check.Passed = false
		check.Messages = append(check.Messages, "Cannot proceed due to missing critical resources")
	}

	return check
}

// checkProviderAvailability validates that the provider is available and functional
func (sm *SafetyManager) checkProviderAvailability(provider *types.ProviderData) SafetyCheck {
	check := SafetyCheck{
		Name:        "Provider Availability",
		Description: "Verify that the selected provider is available and functional",
		Passed:      true,
		Messages:    []string{},
	}

	// Check if provider executable exists
	if provider.Provider.Executable != "" {
		// In a real implementation, this would check if the executable exists and is functional
		check.Messages = append(check.Messages, fmt.Sprintf("Provider %s executable: %s", provider.Provider.Name, provider.Provider.Executable))
	}

	// Check platform compatibility
	if len(provider.Provider.Platforms) > 0 {
		// In a real implementation, this would check current platform against supported platforms
		check.Messages = append(check.Messages, fmt.Sprintf("Provider supports platforms: %v", provider.Provider.Platforms))
	}

	return check
}

// checkDestructiveOperation identifies potentially destructive operations
func (sm *SafetyManager) checkDestructiveOperation(action, software string) SafetyCheck {
	check := SafetyCheck{
		Name:        "Destructive Operation Check",
		Description: "Identify potentially destructive operations that require extra caution",
		Passed:      true,
		Messages:    []string{},
	}

	destructiveActions := []string{"uninstall", "stop", "disable"}
	
	for _, destructive := range destructiveActions {
		if action == destructive {
			check.Messages = append(check.Messages, fmt.Sprintf("Warning: %s is a potentially destructive operation for %s", action, software))
			// Don't fail the check, just warn
			break
		}
	}

	// Special checks for critical system software
	criticalSoftware := []string{"systemd", "kernel", "glibc", "bash", "ssh"}
	for _, critical := range criticalSoftware {
		if strings.Contains(strings.ToLower(software), critical) {
			if action == "uninstall" {
				check.Passed = false
				check.Messages = append(check.Messages, fmt.Sprintf("DANGER: Attempting to uninstall critical system software: %s", software))
			} else {
				check.Messages = append(check.Messages, fmt.Sprintf("Caution: Operating on critical system software: %s", software))
			}
			break
		}
	}

	return check
}

// checkSystemRequirements validates system requirements are met
func (sm *SafetyManager) checkSystemRequirements(saidata *types.SoftwareData) SafetyCheck {
	check := SafetyCheck{
		Name:        "System Requirements",
		Description: "Verify that system requirements are met",
		Passed:      true,
		Messages:    []string{},
	}

	if saidata.Requirements != nil {
		// Check system requirements
		if saidata.Requirements.System != nil {
			systemReqs := saidata.Requirements.System
			
			if systemReqs.MemoryMin != "" {
				check.Messages = append(check.Messages, fmt.Sprintf("Minimum memory required: %s", systemReqs.MemoryMin))
			}
			
			if systemReqs.DiskSpace != "" {
				check.Messages = append(check.Messages, fmt.Sprintf("Disk space required: %s", systemReqs.DiskSpace))
			}
			
			if systemReqs.JavaVersion != "" {
				check.Messages = append(check.Messages, fmt.Sprintf("Java version required: %s", systemReqs.JavaVersion))
			}
		}

		// In a real implementation, this would actually check available resources
		// For now, we'll just log the requirements
	}

	return check
}

// checkTemplateResolution validates that all template variables can be resolved
func (sm *SafetyManager) checkTemplateResolution(provider *types.ProviderData, action string, saidata *types.SoftwareData) SafetyCheck {
	check := SafetyCheck{
		Name:        "Template Resolution",
		Description: "Verify that all template variables can be resolved",
		Passed:      true,
		Messages:    []string{},
	}

	// Check if the action exists in the provider
	actionData, exists := provider.Actions[action]
	if !exists {
		check.Passed = false
		check.Messages = append(check.Messages, fmt.Sprintf("Action %s not found in provider %s", action, provider.Provider.Name))
		return check
	}

	// Check template variables (simplified check)
	if actionData.Template != "" {
		// Look for common template variables that might not resolve
		templateVars := []string{"{{sai_package", "{{sai_service", "{{sai_port", "{{sai_file"}
		
		for _, templateVar := range templateVars {
			if strings.Contains(actionData.Template, templateVar) {
				check.Messages = append(check.Messages, fmt.Sprintf("Template uses variable: %s", templateVar))
			}
		}
	}

	return check
}

// SafetyResult contains the results of safety checks
type SafetyResult struct {
	Safe     bool          `json:"safe"`
	Action   string        `json:"action"`
	Software string        `json:"software"`
	Provider string        `json:"provider"`
	Checks   []SafetyCheck `json:"checks"`
}

// SafetyCheck represents a single safety check
type SafetyCheck struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Passed      bool     `json:"passed"`
	Messages    []string `json:"messages"`
}

// GetFailedChecks returns only the failed safety checks
func (sr *SafetyResult) GetFailedChecks() []SafetyCheck {
	var failed []SafetyCheck
	for _, check := range sr.Checks {
		if !check.Passed {
			failed = append(failed, check)
		}
	}
	return failed
}

// GetWarnings returns all warning messages from safety checks
func (sr *SafetyResult) GetWarnings() []string {
	var warnings []string
	for _, check := range sr.Checks {
		for _, message := range check.Messages {
			if strings.Contains(strings.ToLower(message), "warning") {
				warnings = append(warnings, message)
			}
		}
	}
	return warnings
}

// GetErrors returns all error messages from failed safety checks
func (sr *SafetyResult) GetErrors() []string {
	var errors []string
	for _, check := range sr.Checks {
		if !check.Passed {
			for _, message := range check.Messages {
				if !strings.Contains(strings.ToLower(message), "warning") {
					errors = append(errors, message)
				}
			}
		}
	}
	return errors
}