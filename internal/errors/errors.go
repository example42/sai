package errors

import (
	"fmt"
	"strings"
)

// ErrorType represents different types of errors in the system
type ErrorType string

const (
	// Provider-related errors
	ErrorTypeProviderNotFound     ErrorType = "provider_not_found"
	ErrorTypeProviderUnavailable  ErrorType = "provider_unavailable"
	ErrorTypeProviderInvalid      ErrorType = "provider_invalid"
	ErrorTypeProviderLoadFailed   ErrorType = "provider_load_failed"
	
	// Saidata-related errors
	ErrorTypeSaidataNotFound      ErrorType = "saidata_not_found"
	ErrorTypeSaidataInvalid       ErrorType = "saidata_invalid"
	ErrorTypeSaidataLoadFailed    ErrorType = "saidata_load_failed"
	ErrorTypeSaidataValidation    ErrorType = "saidata_validation"
	
	// Action-related errors
	ErrorTypeActionNotSupported   ErrorType = "action_not_supported"
	ErrorTypeActionFailed         ErrorType = "action_failed"
	ErrorTypeActionTimeout        ErrorType = "action_timeout"
	ErrorTypeActionCancelled      ErrorType = "action_cancelled"
	ErrorTypeActionValidation     ErrorType = "action_validation"
	
	// Command execution errors
	ErrorTypeCommandFailed        ErrorType = "command_failed"
	ErrorTypeCommandTimeout       ErrorType = "command_timeout"
	ErrorTypeCommandNotFound      ErrorType = "command_not_found"
	ErrorTypeCommandPermission    ErrorType = "command_permission"
	
	// Resource validation errors
	ErrorTypeResourceMissing      ErrorType = "resource_missing"
	ErrorTypeResourceInvalid      ErrorType = "resource_invalid"
	ErrorTypeResourcePermission   ErrorType = "resource_permission"
	ErrorTypeResourceValidation   ErrorType = "resource_validation"
	
	// Configuration errors
	ErrorTypeConfigInvalid        ErrorType = "config_invalid"
	ErrorTypeConfigNotFound       ErrorType = "config_not_found"
	ErrorTypeConfigLoadFailed     ErrorType = "config_load_failed"
	
	// Repository errors
	ErrorTypeRepositoryNotFound   ErrorType = "repository_not_found"
	ErrorTypeRepositoryInvalid    ErrorType = "repository_invalid"
	ErrorTypeRepositorySync       ErrorType = "repository_sync"
	ErrorTypeRepositoryAccess     ErrorType = "repository_access"
	
	// Template errors
	ErrorTypeTemplateInvalid      ErrorType = "template_invalid"
	ErrorTypeTemplateRender       ErrorType = "template_render"
	ErrorTypeTemplateVariable     ErrorType = "template_variable"
	
	// System errors
	ErrorTypeSystemRequirement    ErrorType = "system_requirement"
	ErrorTypeSystemPermission     ErrorType = "system_permission"
	ErrorTypeSystemUnsupported    ErrorType = "system_unsupported"
	
	// Network errors
	ErrorTypeNetworkTimeout       ErrorType = "network_timeout"
	ErrorTypeNetworkUnavailable   ErrorType = "network_unavailable"
	ErrorTypeNetworkPermission    ErrorType = "network_permission"
	
	// Generic errors
	ErrorTypeInternal             ErrorType = "internal"
	ErrorTypeUnknown              ErrorType = "unknown"
)

// SAIError represents a structured error in the SAI system
type SAIError struct {
	Type        ErrorType
	Message     string
	Cause       error
	Context     map[string]interface{}
	Suggestions []string
	Recoverable bool
}

// Error implements the error interface
func (e *SAIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause
func (e *SAIError) Unwrap() error {
	return e.Cause
}

// Is checks if the error is of a specific type
func (e *SAIError) Is(target error) bool {
	if saiErr, ok := target.(*SAIError); ok {
		return e.Type == saiErr.Type
	}
	return false
}

// WithContext adds context to the error
func (e *SAIError) WithContext(key string, value interface{}) *SAIError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithSuggestion adds a suggestion to the error
func (e *SAIError) WithSuggestion(suggestion string) *SAIError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// GetDetailedMessage returns a detailed error message with context and suggestions
func (e *SAIError) GetDetailedMessage() string {
	var parts []string
	
	// Main error message
	parts = append(parts, e.Error())
	
	// Add context if available
	if len(e.Context) > 0 {
		var contextParts []string
		for key, value := range e.Context {
			contextParts = append(contextParts, fmt.Sprintf("%s: %v", key, value))
		}
		parts = append(parts, fmt.Sprintf("Context: %s", strings.Join(contextParts, ", ")))
	}
	
	// Add suggestions if available
	if len(e.Suggestions) > 0 {
		parts = append(parts, "Suggestions:")
		for _, suggestion := range e.Suggestions {
			parts = append(parts, fmt.Sprintf("  - %s", suggestion))
		}
	}
	
	return strings.Join(parts, "\n")
}

// NewSAIError creates a new SAI error
func NewSAIError(errorType ErrorType, message string) *SAIError {
	return &SAIError{
		Type:        errorType,
		Message:     message,
		Recoverable: isRecoverable(errorType),
	}
}

// WrapSAIError wraps an existing error with SAI error context
func WrapSAIError(errorType ErrorType, message string, cause error) *SAIError {
	return &SAIError{
		Type:        errorType,
		Message:     message,
		Cause:       cause,
		Recoverable: isRecoverable(errorType),
	}
}

// isRecoverable determines if an error type is recoverable
func isRecoverable(errorType ErrorType) bool {
	switch errorType {
	case ErrorTypeProviderNotFound, ErrorTypeProviderUnavailable:
		return true // Can try different provider
	case ErrorTypeSaidataNotFound:
		return true // Can use defaults
	case ErrorTypeActionTimeout:
		return true // Can retry
	case ErrorTypeNetworkTimeout, ErrorTypeNetworkUnavailable:
		return true // Can retry
	case ErrorTypeResourceMissing:
		return true // Can create or use alternatives
	case ErrorTypeConfigNotFound:
		return true // Can use defaults
	default:
		return false
	}
}

// Predefined error constructors for common scenarios

// Provider errors
func NewProviderNotFoundError(providerName string) *SAIError {
	return NewSAIError(ErrorTypeProviderNotFound, fmt.Sprintf("provider '%s' not found", providerName)).
		WithContext("provider", providerName).
		WithSuggestion("Check available providers with 'sai stats'").
		WithSuggestion("Verify provider name spelling")
}

func NewProviderUnavailableError(providerName string, reason string) *SAIError {
	return NewSAIError(ErrorTypeProviderUnavailable, fmt.Sprintf("provider '%s' is unavailable: %s", providerName, reason)).
		WithContext("provider", providerName).
		WithContext("reason", reason).
		WithSuggestion("Install the required provider executable").
		WithSuggestion("Try a different provider")
}

func NewProviderInvalidError(providerName string, validationError error) *SAIError {
	return WrapSAIError(ErrorTypeProviderInvalid, fmt.Sprintf("provider '%s' configuration is invalid", providerName), validationError).
		WithContext("provider", providerName).
		WithSuggestion("Check provider YAML syntax").
		WithSuggestion("Validate against provider schema")
}

// Saidata errors
func NewSaidataNotFoundError(software string) *SAIError {
	return NewSAIError(ErrorTypeSaidataNotFound, fmt.Sprintf("saidata for '%s' not found", software)).
		WithContext("software", software).
		WithSuggestion("Using intelligent defaults").
		WithSuggestion("Update saidata repository with 'sai saidata update'")
}

func NewSaidataInvalidError(software string, validationError error) *SAIError {
	return WrapSAIError(ErrorTypeSaidataInvalid, fmt.Sprintf("saidata for '%s' is invalid", software), validationError).
		WithContext("software", software).
		WithSuggestion("Check saidata YAML syntax").
		WithSuggestion("Validate against saidata schema")
}

// Action errors
func NewActionNotSupportedError(action string, software string, provider string) *SAIError {
	return NewSAIError(ErrorTypeActionNotSupported, fmt.Sprintf("action '%s' not supported for '%s' by provider '%s'", action, software, provider)).
		WithContext("action", action).
		WithContext("software", software).
		WithContext("provider", provider).
		WithSuggestion("Check available actions with 'sai info " + software + "'").
		WithSuggestion("Try a different provider")
}

func NewActionFailedError(action string, software string, exitCode int, output string) *SAIError {
	return NewSAIError(ErrorTypeActionFailed, fmt.Sprintf("action '%s' failed for '%s' (exit code: %d)", action, software, exitCode)).
		WithContext("action", action).
		WithContext("software", software).
		WithContext("exit_code", exitCode).
		WithContext("output", output).
		WithSuggestion("Check command output for details").
		WithSuggestion("Run with --verbose for more information")
}

func NewActionTimeoutError(action string, software string, timeout string) *SAIError {
	return NewSAIError(ErrorTypeActionTimeout, fmt.Sprintf("action '%s' timed out for '%s' after %s", action, software, timeout)).
		WithContext("action", action).
		WithContext("software", software).
		WithContext("timeout", timeout).
		WithSuggestion("Increase timeout with --timeout flag").
		WithSuggestion("Check system resources and network connectivity")
}

// Command errors
func NewCommandFailedError(command string, exitCode int, stderr string) *SAIError {
	return NewSAIError(ErrorTypeCommandFailed, fmt.Sprintf("command failed: %s (exit code: %d)", command, exitCode)).
		WithContext("command", command).
		WithContext("exit_code", exitCode).
		WithContext("stderr", stderr).
		WithSuggestion("Check command syntax and arguments").
		WithSuggestion("Verify required permissions")
}

func NewCommandNotFoundError(command string) *SAIError {
	return NewSAIError(ErrorTypeCommandNotFound, fmt.Sprintf("command not found: %s", command)).
		WithContext("command", command).
		WithSuggestion("Install the required package").
		WithSuggestion("Check PATH environment variable")
}

func NewCommandPermissionError(command string) *SAIError {
	return NewSAIError(ErrorTypeCommandPermission, fmt.Sprintf("permission denied: %s", command)).
		WithContext("command", command).
		WithSuggestion("Run with appropriate privileges").
		WithSuggestion("Check file permissions")
}

// Resource errors
func NewResourceMissingError(resourceType string, resourcePath string) *SAIError {
	return NewSAIError(ErrorTypeResourceMissing, fmt.Sprintf("%s not found: %s", resourceType, resourcePath)).
		WithContext("resource_type", resourceType).
		WithContext("resource_path", resourcePath).
		WithSuggestion("Create the missing resource").
		WithSuggestion("Check path spelling and permissions")
}

func NewResourceValidationError(details string) *SAIError {
	return NewSAIError(ErrorTypeResourceValidation, fmt.Sprintf("resource validation failed: %s", details)).
		WithContext("details", details).
		WithSuggestion("Check resource availability").
		WithSuggestion("Run with --dry-run to see what would be executed")
}

// Template errors
func NewTemplateRenderError(template string, cause error) *SAIError {
	return WrapSAIError(ErrorTypeTemplateRender, fmt.Sprintf("failed to render template: %s", template), cause).
		WithContext("template", template).
		WithSuggestion("Check template syntax").
		WithSuggestion("Verify all variables are available")
}

func NewTemplateVariableError(variable string, template string) *SAIError {
	return NewSAIError(ErrorTypeTemplateVariable, fmt.Sprintf("template variable '%s' not found in template: %s", variable, template)).
		WithContext("variable", variable).
		WithContext("template", template).
		WithSuggestion("Check saidata for missing variable").
		WithSuggestion("Use intelligent defaults")
}

// System errors
func NewSystemRequirementError(requirement string, current string, minimum string) *SAIError {
	return NewSAIError(ErrorTypeSystemRequirement, fmt.Sprintf("system requirement not met: %s (current: %s, minimum: %s)", requirement, current, minimum)).
		WithContext("requirement", requirement).
		WithContext("current", current).
		WithContext("minimum", minimum).
		WithSuggestion("Upgrade system resources").
		WithSuggestion("Check system requirements")
}

func NewSystemUnsupportedError(platform string, architecture string) *SAIError {
	return NewSAIError(ErrorTypeSystemUnsupported, fmt.Sprintf("unsupported platform: %s/%s", platform, architecture)).
		WithContext("platform", platform).
		WithContext("architecture", architecture).
		WithSuggestion("Check compatibility matrix").
		WithSuggestion("Try a different provider")
}

// Configuration errors
func NewConfigInvalidError(configPath string, cause error) *SAIError {
	return WrapSAIError(ErrorTypeConfigInvalid, fmt.Sprintf("invalid configuration: %s", configPath), cause).
		WithContext("config_path", configPath).
		WithSuggestion("Check configuration syntax").
		WithSuggestion("Validate against configuration schema")
}

func NewConfigNotFoundError(configPaths []string) *SAIError {
	return NewSAIError(ErrorTypeConfigNotFound, "configuration file not found").
		WithContext("searched_paths", configPaths).
		WithSuggestion("Create a configuration file").
		WithSuggestion("Use --config flag to specify path")
}

// Repository errors
func NewRepositorySyncError(repoURL string, cause error) *SAIError {
	return WrapSAIError(ErrorTypeRepositorySync, fmt.Sprintf("failed to sync repository: %s", repoURL), cause).
		WithContext("repository_url", repoURL).
		WithSuggestion("Check network connectivity").
		WithSuggestion("Verify repository URL")
}

func NewRepositoryAccessError(repoPath string, cause error) *SAIError {
	return WrapSAIError(ErrorTypeRepositoryAccess, fmt.Sprintf("cannot access repository: %s", repoPath), cause).
		WithContext("repository_path", repoPath).
		WithSuggestion("Check directory permissions").
		WithSuggestion("Verify repository path exists")
}

// Error checking utilities

// IsRecoverable checks if an error is recoverable
func IsRecoverable(err error) bool {
	if saiErr, ok := err.(*SAIError); ok {
		return saiErr.Recoverable
	}
	return false
}

// GetErrorType returns the error type if it's a SAI error
func GetErrorType(err error) ErrorType {
	if saiErr, ok := err.(*SAIError); ok {
		return saiErr.Type
	}
	return ErrorTypeUnknown
}

// HasErrorType checks if an error is of a specific type
func HasErrorType(err error, errorType ErrorType) bool {
	return GetErrorType(err) == errorType
}

// GetSuggestions returns suggestions from a SAI error
func GetSuggestions(err error) []string {
	if saiErr, ok := err.(*SAIError); ok {
		return saiErr.Suggestions
	}
	return nil
}

// GetContext returns context from a SAI error
func GetContext(err error) map[string]interface{} {
	if saiErr, ok := err.(*SAIError); ok {
		return saiErr.Context
	}
	return nil
}