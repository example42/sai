package template

import (
	"fmt"
	"strings"

	"sai/internal/types"
)

// TemplateResolutionValidator validates template resolution and determines action availability
type TemplateResolutionValidator struct {
	engine *TemplateEngine
}

// NewTemplateResolutionValidator creates a new template resolution validator
func NewTemplateResolutionValidator(engine *TemplateEngine) *TemplateResolutionValidator {
	return &TemplateResolutionValidator{
		engine: engine,
	}
}

// ValidationResult represents the result of template validation
type ValidationResult struct {
	Valid              bool
	Resolvable         bool
	UnresolvedVariables []string
	MissingResources   []string
	Errors             []string
	CanExecute         bool
}

// ValidateActionTemplate validates if an action template can be resolved and executed
func (v *TemplateResolutionValidator) ValidateActionTemplate(
	action *types.Action,
	software string,
	provider string,
	saidata *types.SoftwareData,
) *ValidationResult {
	result := &ValidationResult{
		Valid:               true,
		Resolvable:          true,
		UnresolvedVariables: []string{},
		MissingResources:    []string{},
		Errors:              []string{},
		CanExecute:          true,
	}
	
	if action == nil {
		result.Valid = false
		result.Resolvable = false
		result.CanExecute = false
		result.Errors = append(result.Errors, "action is nil")
		return result
	}
	
	// Get the command template to validate
	template := action.GetCommand()
	if template == "" {
		result.Valid = false
		result.Resolvable = false
		result.CanExecute = false
		result.Errors = append(result.Errors, "action has no command template")
		return result
	}
	
	// Create template context
	context := &TemplateContext{
		Software: software,
		Provider: provider,
		Saidata:  saidata,
		Variables: action.Variables,
	}
	
	// First, validate template syntax
	if err := v.engine.ValidateTemplate(template); err != nil {
		result.Valid = false
		result.Resolvable = false
		result.CanExecute = false
		result.Errors = append(result.Errors, fmt.Sprintf("template syntax error: %v", err))
		return result
	}
	
	// Try to render the template with safety mode disabled to check resolvability
	originalSafetyMode := v.engine.safetyMode
	v.engine.SetSafetyMode(false)
	
	rendered, err := v.engine.Render(template, context)
	
	// Restore original safety mode
	v.engine.SetSafetyMode(originalSafetyMode)
	
	if err != nil {
		result.Resolvable = false
		result.CanExecute = false
		result.Errors = append(result.Errors, fmt.Sprintf("template resolution error: %v", err))
		
		// Try to extract specific unresolved variables from the error
		if strings.Contains(err.Error(), "not found") {
			result.UnresolvedVariables = append(result.UnresolvedVariables, v.extractVariableFromError(err.Error()))
		}
		
		return result
	}
	
	// Check for unresolved template variables in the rendered output
	unresolvedVars := v.findUnresolvedVariables(rendered)
	if len(unresolvedVars) > 0 {
		result.Resolvable = false
		result.CanExecute = false
		result.UnresolvedVariables = unresolvedVars
		result.Errors = append(result.Errors, fmt.Sprintf("template contains unresolved variables: %v", unresolvedVars))
		return result
	}
	
	// If safety mode is enabled, validate resource existence
	if v.engine.safetyMode {
		resourceValidation := v.validateResourceExistence(saidata, action)
		if !resourceValidation.CanExecute {
			result.CanExecute = false
			result.MissingResources = resourceValidation.MissingResources
			result.Errors = append(result.Errors, resourceValidation.Errors...)
		}
	}
	
	return result
}

// ValidateProviderActions validates all actions for a provider and returns which ones can be executed
func (v *TemplateResolutionValidator) ValidateProviderActions(
	provider *types.ProviderData,
	software string,
	saidata *types.SoftwareData,
) map[string]*ValidationResult {
	results := make(map[string]*ValidationResult)
	
	for actionName, action := range provider.Actions {
		results[actionName] = v.ValidateActionTemplate(&action, software, provider.Provider.Name, saidata)
	}
	
	return results
}

// GetExecutableActions returns a list of action names that can be executed
func (v *TemplateResolutionValidator) GetExecutableActions(
	provider *types.ProviderData,
	software string,
	saidata *types.SoftwareData,
) []string {
	var executableActions []string
	
	validationResults := v.ValidateProviderActions(provider, software, saidata)
	
	for actionName, result := range validationResults {
		if result.CanExecute {
			executableActions = append(executableActions, actionName)
		}
	}
	
	return executableActions
}

// findUnresolvedVariables finds template variables that weren't resolved
func (v *TemplateResolutionValidator) findUnresolvedVariables(rendered string) []string {
	var unresolved []string
	
	// Look for template syntax that wasn't resolved
	if strings.Contains(rendered, "{{") && strings.Contains(rendered, "}}") {
		// Extract the unresolved variables
		start := 0
		for {
			startIdx := strings.Index(rendered[start:], "{{")
			if startIdx == -1 {
				break
			}
			startIdx += start
			
			endIdx := strings.Index(rendered[startIdx:], "}}")
			if endIdx == -1 {
				break
			}
			endIdx += startIdx + 2
			
			variable := rendered[startIdx:endIdx]
			unresolved = append(unresolved, variable)
			start = endIdx
		}
	}
	
	// Look for <no value> indicators
	if strings.Contains(rendered, "<no value>") {
		unresolved = append(unresolved, "<no value>")
	}
	
	return unresolved
}

// extractVariableFromError extracts variable names from error messages
func (v *TemplateResolutionValidator) extractVariableFromError(errorMsg string) string {
	// Try to extract variable names from common error patterns
	if strings.Contains(errorMsg, "not found") {
		parts := strings.Split(errorMsg, " ")
		for i, part := range parts {
			if part == "not" && i > 0 {
				return parts[i-1]
			}
		}
	}
	return "unknown"
}

// validateResourceExistence validates that resources referenced in saidata actually exist
func (v *TemplateResolutionValidator) validateResourceExistence(
	saidata *types.SoftwareData,
	action *types.Action,
) *ValidationResult {
	result := &ValidationResult{
		Valid:            true,
		Resolvable:       true,
		MissingResources: []string{},
		Errors:           []string{},
		CanExecute:       true,
	}
	
	if saidata == nil {
		return result // No saidata to validate
	}
	
	// Validate files
	for _, file := range saidata.Files {
		if !v.engine.validator.FileExists(file.Path) {
			result.MissingResources = append(result.MissingResources, fmt.Sprintf("file: %s", file.Path))
		}
	}
	
	// Validate services (only for service-related actions)
	if v.isServiceAction(action) {
		for _, service := range saidata.Services {
			if !v.engine.validator.ServiceExists(service.GetServiceNameOrDefault()) {
				result.MissingResources = append(result.MissingResources, fmt.Sprintf("service: %s", service.GetServiceNameOrDefault()))
			}
		}
	}
	
	// Validate commands
	for _, command := range saidata.Commands {
		if !v.engine.validator.CommandExists(command.GetPathOrDefault()) {
			result.MissingResources = append(result.MissingResources, fmt.Sprintf("command: %s", command.GetPathOrDefault()))
		}
	}
	
	// Validate directories
	for _, directory := range saidata.Directories {
		if !v.engine.validator.DirectoryExists(directory.Path) {
			result.MissingResources = append(result.MissingResources, fmt.Sprintf("directory: %s", directory.Path))
		}
	}
	
	// If there are missing resources, the action cannot be executed safely
	if len(result.MissingResources) > 0 {
		result.CanExecute = false
		result.Errors = append(result.Errors, fmt.Sprintf("missing resources prevent safe execution: %v", result.MissingResources))
	}
	
	return result
}

// isServiceAction determines if an action is service-related
func (v *TemplateResolutionValidator) isServiceAction(action *types.Action) bool {
	if action == nil {
		return false
	}
	
	template := strings.ToLower(action.GetCommand())
	
	// Check for common service-related commands
	serviceKeywords := []string{
		"systemctl", "service", "start", "stop", "restart", 
		"enable", "disable", "status", "launchctl",
	}
	
	for _, keyword := range serviceKeywords {
		if strings.Contains(template, keyword) {
			return true
		}
	}
	
	return false
}

// SafetyMode represents different levels of safety validation
type SafetyMode int

const (
	SafetyModeDisabled SafetyMode = iota // No safety checks
	SafetyModeWarning                    // Show warnings but allow execution
	SafetyModeStrict                     // Prevent execution if resources don't exist
)

// SetSafetyMode sets the safety mode for the validator
func (v *TemplateResolutionValidator) SetSafetyMode(mode SafetyMode) {
	switch mode {
	case SafetyModeDisabled:
		v.engine.SetSafetyMode(false)
	case SafetyModeWarning, SafetyModeStrict:
		v.engine.SetSafetyMode(true)
	}
}

// ActionAvailability represents the availability status of an action
type ActionAvailability struct {
	ActionName          string
	Available           bool
	Reason              string
	UnresolvedVariables []string
	MissingResources    []string
}

// GetActionAvailability returns the availability status for all actions
func (v *TemplateResolutionValidator) GetActionAvailability(
	provider *types.ProviderData,
	software string,
	saidata *types.SoftwareData,
) []*ActionAvailability {
	var availability []*ActionAvailability
	
	validationResults := v.ValidateProviderActions(provider, software, saidata)
	
	for actionName, result := range validationResults {
		avail := &ActionAvailability{
			ActionName:          actionName,
			Available:           result.CanExecute,
			UnresolvedVariables: result.UnresolvedVariables,
			MissingResources:    result.MissingResources,
		}
		
		if !result.CanExecute {
			if len(result.UnresolvedVariables) > 0 {
				avail.Reason = fmt.Sprintf("Unresolved template variables: %v", result.UnresolvedVariables)
			} else if len(result.MissingResources) > 0 {
				avail.Reason = fmt.Sprintf("Missing resources: %v", result.MissingResources)
			} else if len(result.Errors) > 0 {
				avail.Reason = strings.Join(result.Errors, "; ")
			} else {
				avail.Reason = "Unknown validation error"
			}
		}
		
		availability = append(availability, avail)
	}
	
	return availability
}