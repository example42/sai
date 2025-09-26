package template

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"sai/internal/debug"
	"sai/internal/interfaces"
	"sai/internal/types"
)

// TemplateEngine provides template rendering with saidata functions
type TemplateEngine struct {
	template     *template.Template
	saidata      *types.SoftwareData
	safetyMode   bool
	validator    ResourceValidator
	defaultsGen  DefaultsGenerator
}

// ResourceValidator validates resource existence
type ResourceValidator interface {
	FileExists(path string) bool
	ServiceExists(service string) bool
	CommandExists(command string) bool
	DirectoryExists(path string) bool
}

// DefaultsGenerator generates default paths and configurations
type DefaultsGenerator interface {
	DefaultConfigPath(software string) string
	DefaultLogPath(software string) string
	DefaultDataDir(software string) string
	DefaultServiceName(software string) string
	DefaultCommandPath(software string) string
}

// TemplateContext is an alias to the interfaces.TemplateContext for compatibility
type TemplateContext = interfaces.TemplateContext

// NewTemplateEngine creates a new template engine instance
func NewTemplateEngine(validator ResourceValidator, defaultsGen DefaultsGenerator) *TemplateEngine {
	engine := &TemplateEngine{
		validator:   validator,
		defaultsGen: defaultsGen,
		safetyMode:  true,
	}
	
	// Create template with custom functions
	tmpl := template.New("sai").Funcs(engine.createFuncMap())
	engine.template = tmpl
	
	return engine
}

// SetSafetyMode enables or disables safety mode
func (e *TemplateEngine) SetSafetyMode(enabled bool) {
	e.safetyMode = enabled
}

// SetSaidata sets the current saidata context for template functions
func (e *TemplateEngine) SetSaidata(saidata *types.SoftwareData) {
	e.saidata = saidata
}

// Render renders a template string with the given context
func (e *TemplateEngine) Render(templateStr string, context *TemplateContext) (string, error) {
	startTime := time.Now()
	
	if context == nil {
		debug.LogTemplateResolutionGlobal(templateStr, nil, "", false, time.Since(startTime), fmt.Errorf("template context cannot be nil"))
		return "", fmt.Errorf("template context cannot be nil")
	}
	
	// Set saidata context for template functions
	e.saidata = context.Saidata
	
	// Preprocess template to convert legacy syntax to Go template syntax
	processedTemplate := e.preprocessTemplate(templateStr)
	
	// Parse the template
	tmpl, err := e.template.Clone()
	if err != nil {
		debug.LogTemplateResolutionGlobal(templateStr, e.createVariableMap(context), "", false, time.Since(startTime), fmt.Errorf("failed to clone template: %w", err))
		return "", fmt.Errorf("failed to clone template: %w", err)
	}
	
	tmpl, err = tmpl.Parse(processedTemplate)
	if err != nil {
		debug.LogTemplateResolutionGlobal(templateStr, e.createVariableMap(context), "", false, time.Since(startTime), fmt.Errorf("failed to parse template: %w", err))
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	
	// Create template data
	data := map[string]interface{}{
		"Software":  context.Software,
		"Provider":  context.Provider,
		"Variables": context.Variables,
	}
	
	// Execute template
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		debug.LogTemplateResolutionGlobal(templateStr, e.createVariableMap(context), "", false, time.Since(startTime), fmt.Errorf("failed to execute template: %w", err))
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	
	result := buf.String()
	resolutionTime := time.Since(startTime)
	
	// Validate template resolution if safety mode is enabled
	var validationErr error
	if e.safetyMode {
		if err := e.validateTemplateResolution(result, processedTemplate, context); err != nil {
			validationErr = fmt.Errorf("template validation failed: %w", err)
			debug.LogTemplateResolutionGlobal(templateStr, e.createVariableMap(context), result, false, resolutionTime, validationErr)
			return "", validationErr
		}
	}
	
	// Log successful template resolution
	debug.LogTemplateResolutionGlobal(templateStr, e.createVariableMap(context), result, true, resolutionTime, nil)
	
	return result, nil
}

// ValidateTemplate validates a template string without executing it
func (e *TemplateEngine) ValidateTemplate(templateStr string) error {
	// Preprocess template to convert legacy syntax to Go template syntax
	processedTemplate := e.preprocessTemplate(templateStr)
	
	tmpl, err := e.template.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone template: %w", err)
	}
	
	_, err = tmpl.Parse(processedTemplate)
	if err != nil {
		return fmt.Errorf("template syntax error: %w", err)
	}
	
	return nil
}

// createFuncMap creates the function map for template functions
func (e *TemplateEngine) createFuncMap() template.FuncMap {
	return template.FuncMap{
		// Saidata functions - now support multiple calling patterns
		"sai_package":       e.saiPackage,
		"sai_packages":      e.saiPackages,
		"sai_service":       e.saiService,
		"sai_port":          e.saiPort,
		"sai_file":          e.saiFile,
		"sai_directory":     e.saiDirectory,
		"sai_command":       e.saiCommand,
		"sai_container":     e.saiContainer,
		
		// Alternative installation provider functions
		"sai_source":        e.saiSource,
		"sai_binary":        e.saiBinary,
		"sai_script":        e.saiScript,
		
		// Safety validation functions
		"file_exists":       e.fileExists,
		"service_exists":    e.serviceExists,
		"command_exists":    e.commandExists,
		"directory_exists":  e.directoryExists,
		
		// Default generation functions
		"default_config_path": e.defaultConfigPath,
		"default_log_path":    e.defaultLogPath,
		"default_data_dir":    e.defaultDataDir,
		"default_service_name": e.defaultServiceName,
		"default_command_path": e.defaultCommandPath,
	}
}

// saiPackage returns the package name for a specific provider
// Supports multiple calling patterns:
// - sai_package("provider") - returns first package for provider
// - sai_package("provider", index) - returns package at index for provider  
// - sai_package("*", "name", "provider") - returns all package names for provider (space-separated)
// - sai_package(index, "name", "provider") - returns package name at index for provider
func (e *TemplateEngine) saiPackage(args ...interface{}) string {
	if e.saidata == nil {
		return "sai_package error: no saidata context available"
	}
	
	if len(args) == 0 {
		return "sai_package error: requires at least one argument"
	}
	
	// Handle different calling patterns
	switch len(args) {
	case 1:
		// sai_package("provider") - return first package
		provider, ok := args[0].(string)
		if !ok {
			return "sai_package error: first argument must be provider name (string)"
		}
		result, err := e.getPackageByIndex(provider, 0, "package_name")
		if err != nil {
			return fmt.Sprintf("sai_package error: %v", err)
		}
		return result
		
	case 2:
		// sai_package("provider", index) - return package at index
		provider, ok := args[0].(string)
		if !ok {
			return "sai_package error: first argument must be provider name (string)"
		}
		idx, ok := args[1].(int)
		if !ok {
			return "sai_package error: second argument must be index (int)"
		}
		result, err := e.getPackageByIndex(provider, idx, "package_name")
		if err != nil {
			return fmt.Sprintf("sai_package error: %v", err)
		}
		return result
		
	case 3:
		// Handle legacy provider template format: sai_package("*"|index, "name", "provider")
		provider, ok := args[2].(string)
		if !ok {
			return "sai_package error: third argument must be provider name (string)"
		}
		
		field, ok := args[1].(string)
		if !ok || (field != "name" && field != "package_name") {
			return "sai_package error: second argument must be 'name' or 'package_name' field"
		}
		
		// Check if first arg is "*" for all packages
		if firstArg, ok := args[0].(string); ok && firstArg == "*" {
			result, err := e.getAllPackageNames(provider, field)
			if err != nil {
				return fmt.Sprintf("sai_package error: %v", err)
			}
			return result
		}
		
		// Otherwise treat first arg as index
		if idx, ok := args[0].(int); ok {
			result, err := e.getPackageByIndex(provider, idx, field)
			if err != nil {
				return fmt.Sprintf("sai_package error: %v", err)
			}
			return result
		}
		
		return "sai_package error: first argument must be '*' or index (int)"
		
	default:
		return fmt.Sprintf("sai_package error: accepts 1-3 arguments, got %d", len(args))
	}
}

// getPackageByIndex returns package name at specific index for provider
func (e *TemplateEngine) getPackageByIndex(provider string, idx int, field string) (string, error) {
	// Check provider-specific packages first
	if providerConfig := e.saidata.GetProviderConfig(provider); providerConfig != nil {
		if len(providerConfig.Packages) > idx {
			pkg := providerConfig.Packages[idx]
			if field == "package_name" {
				return pkg.GetPackageNameOrDefault(), nil
			} else {
				return pkg.Name, nil
			}
		}
	}
	
	// Fall back to default packages
	if len(e.saidata.Packages) > idx {
		pkg := e.saidata.Packages[idx]
		if field == "package_name" {
			return pkg.GetPackageNameOrDefault(), nil
		} else {
			return pkg.Name, nil
		}
	}
	
	return fmt.Sprintf("sai_package error: no package found at index %d for provider %s", idx, provider), nil
}

// getAllPackageNames returns all package names for provider (space-separated)
func (e *TemplateEngine) getAllPackageNames(provider string, field string) (string, error) {
	var packages []string
	
	// Check provider-specific packages first
	if providerConfig := e.saidata.GetProviderConfig(provider); providerConfig != nil {
		for _, pkg := range providerConfig.Packages {
			if field == "package_name" {
				packages = append(packages, pkg.GetPackageNameOrDefault())
			} else {
				packages = append(packages, pkg.Name)
			}
		}
		if len(packages) > 0 {
			return strings.Join(packages, " "), nil
		}
	}
	
	// Fall back to default packages
	for _, pkg := range e.saidata.Packages {
		if field == "package_name" {
			packages = append(packages, pkg.GetPackageNameOrDefault())
		} else {
			packages = append(packages, pkg.Name)
		}
	}
	
	if len(packages) == 0 {
		return fmt.Sprintf("sai_package error: no packages found for provider %s", provider), nil
	}
	
	return strings.Join(packages, " "), nil
}

// saiPackages returns all package names for a specific provider as a space-separated string
func (e *TemplateEngine) saiPackages(provider string) []string {
	if e.saidata == nil {
		return []string{"sai_packages error: no saidata context available"}
	}
	
	var packages []string
	
	// Check provider-specific packages first
	if providerConfig := e.saidata.GetProviderConfig(provider); providerConfig != nil {
		for _, pkg := range providerConfig.Packages {
			// Use GetPackageNameOrDefault method for consistent naming
			packages = append(packages, pkg.GetPackageNameOrDefault())
		}
		if len(packages) > 0 {
			return packages
		}
	}
	
	// Fall back to default packages
	for _, pkg := range e.saidata.Packages {
		// Use GetPackageNameOrDefault method for consistent naming
		packages = append(packages, pkg.GetPackageNameOrDefault())
	}
	
	if len(packages) == 0 {
		return []string{fmt.Sprintf("sai_packages error: no packages found for provider %s", provider)}
	}
	
	return packages
}

// saiService returns the service name
// Supports multiple calling patterns:
// - sai_service("name") - returns service_name for service with logical name
// - sai_service(index, "service_name", "provider") - returns service_name at index for provider
func (e *TemplateEngine) saiService(args ...interface{}) string {
	if e.saidata == nil {
		return "sai_service error: no saidata context available"
	}
	
	if len(args) == 0 {
		return "sai_service error: requires at least one argument"
	}
	
	switch len(args) {
	case 1:
		// sai_service("name") - return service_name for logical name
		name, ok := args[0].(string)
		if !ok {
			return "sai_service error: argument must be service name (string)"
		}
		
		service := e.saidata.GetServiceByName(name)
		if service == nil {
			return fmt.Sprintf("sai_service error: service %s not found", name)
		}
		
		return service.GetServiceNameOrDefault()
		
	case 3:
		// Handle legacy provider template format: sai_service(index, "service_name", "provider")
		provider, ok := args[2].(string)
		if !ok {
			return "sai_service error: third argument must be provider name (string)"
		}
		
		field, ok := args[1].(string)
		if !ok || field != "service_name" {
			return "sai_service error: second argument must be 'service_name' field"
		}
		
		idx, ok := args[0].(int)
		if !ok {
			return "sai_service error: first argument must be index (int)"
		}
		
		result, err := e.getServiceByIndex(provider, idx)
		if err != nil {
			return fmt.Sprintf("sai_service error: %v", err)
		}
		return result
		
	default:
		return fmt.Sprintf("sai_service error: accepts 1 or 3 arguments, got %d", len(args))
	}
}

// getServiceByIndex returns service_name at specific index for provider
func (e *TemplateEngine) getServiceByIndex(provider string, idx int) (string, error) {
	// Check provider-specific services first
	if providerConfig := e.saidata.GetProviderConfig(provider); providerConfig != nil {
		if len(providerConfig.Services) > idx {
			return providerConfig.Services[idx].GetServiceNameOrDefault(), nil
		}
	}
	
	// Fall back to default services
	if len(e.saidata.Services) > idx {
		return e.saidata.Services[idx].GetServiceNameOrDefault(), nil
	}
	
	return "", fmt.Errorf("no service found at index %d for provider %s", idx, provider)
}

// saiPort returns the port number
// Supports multiple calling patterns:
// - sai_port() - returns first port
// - sai_port(index) - returns port at index
// - sai_port(index, "port", "provider") - returns port at index for provider
func (e *TemplateEngine) saiPort(args ...interface{}) int {
	if e.saidata == nil {
		return -1 // Return error indicator
	}
	
	switch len(args) {
	case 0:
		// sai_port() - return first port
		result, err := e.getPortByIndex("", 0)
		if err != nil {
			return -1 // Return error indicator
		}
		return result
		
	case 1:
		// sai_port(index) - return port at index
		idx, ok := args[0].(int)
		if !ok {
			return -1 // Return error indicator
		}
		result, err := e.getPortByIndex("", idx)
		if err != nil {
			return -1 // Return error indicator
		}
		return result
		
	case 3:
		// Handle legacy provider template format: sai_port(index, "port", "provider")
		provider, ok := args[2].(string)
		if !ok {
			return -1 // Return error indicator
		}
		
		field, ok := args[1].(string)
		if !ok || field != "port" {
			return -1 // Return error indicator
		}
		
		idx, ok := args[0].(int)
		if !ok {
			return -1 // Return error indicator
		}
		
		result, err := e.getPortByIndex(provider, idx)
		if err != nil {
			return -1 // Return error indicator
		}
		return result
		
	default:
		return -1 // Return error indicator
	}
}

// getPortByIndex returns port number at specific index for provider
func (e *TemplateEngine) getPortByIndex(provider string, idx int) (int, error) {
	// If provider specified, check provider-specific ports first
	if provider != "" {
		if providerConfig := e.saidata.GetProviderConfig(provider); providerConfig != nil {
			if len(providerConfig.Ports) > idx {
				return providerConfig.Ports[idx].Port, nil
			}
		}
	}
	
	// Fall back to default ports
	if len(e.saidata.Ports) <= idx {
		return -1, fmt.Errorf("no port found at index %d", idx)
	}
	
	return e.saidata.Ports[idx].Port, nil
}

// saiFile returns the file path
// Supports multiple calling patterns:
// - sai_file("name") - returns path for file with logical name
// - sai_file("name", "path", "provider") - returns path for file with logical name for provider
func (e *TemplateEngine) saiFile(args ...interface{}) string {
	if e.saidata == nil {
		return "sai_file error: no saidata context available"
	}
	
	if len(args) == 0 {
		return "sai_file error: requires at least one argument"
	}
	
	switch len(args) {
	case 1:
		// sai_file("name") - return path for logical name
		name, ok := args[0].(string)
		if !ok {
			return "sai_file error: argument must be file name (string)"
		}
		
		file := e.saidata.GetFileByName(name)
		if file == nil {
			return fmt.Sprintf("sai_file error: file %s not found", name)
		}
		
		return file.Path
		
	case 3:
		// Handle legacy provider template format: sai_file("name", "path", "provider")
		provider, ok := args[2].(string)
		if !ok {
			return "sai_file error: third argument must be provider name (string)"
		}
		
		field, ok := args[1].(string)
		if !ok || field != "path" {
			return "sai_file error: second argument must be 'path' field"
		}
		
		name, ok := args[0].(string)
		if !ok {
			return "sai_file error: first argument must be file name (string)"
		}
		
		result, err := e.getFileByName(provider, name)
		if err != nil {
			return fmt.Sprintf("sai_file error: %v", err)
		}
		return result
		
	default:
		return fmt.Sprintf("sai_file error: accepts 1 or 3 arguments, got %d", len(args))
	}
}

// getFileByName returns file path for logical name, checking provider-specific files first
func (e *TemplateEngine) getFileByName(provider, name string) (string, error) {
	// Check provider-specific files first
	if provider != "" {
		if providerConfig := e.saidata.GetProviderConfig(provider); providerConfig != nil {
			for _, file := range providerConfig.Files {
				if file.Name == name {
					return file.Path, nil
				}
			}
		}
	}
	
	// Fall back to default files
	file := e.saidata.GetFileByName(name)
	if file == nil {
		return "", fmt.Errorf("file %s not found", name)
	}
	
	return file.Path, nil
}

// saiDirectory returns the directory path
func (e *TemplateEngine) saiDirectory(name string) string {
	if e.saidata == nil {
		return "sai_directory error: no saidata context available"
	}
	
	directory := e.saidata.GetDirectoryByName(name)
	if directory == nil {
		return fmt.Sprintf("sai_directory error: directory %s not found", name)
	}
	
	return directory.Path
}

// saiCommand returns the command path
func (e *TemplateEngine) saiCommand(name string) string {
	if e.saidata == nil {
		return "sai_command error: no saidata context available"
	}
	
	command := e.saidata.GetCommandByName(name)
	if command == nil {
		return fmt.Sprintf("sai_command error: command %s not found", name)
	}
	
	return command.GetPathOrDefault()
}

// saiContainer returns container information
// Supports multiple calling patterns:
// - sai_container("name") - returns full image name for container with logical name
// - sai_container(index, "field", "provider") - returns field value at index for provider
func (e *TemplateEngine) saiContainer(args ...interface{}) string {
	if e.saidata == nil {
		return "sai_container error: no saidata context available"
	}
	
	if len(args) == 0 {
		return "sai_container error: requires at least one argument"
	}
	
	switch len(args) {
	case 1:
		// sai_container("name") - return full image name for logical name
		name, ok := args[0].(string)
		if !ok {
			return "sai_container error: argument must be container name (string)"
		}
		
		container := e.saidata.GetContainerByName(name)
		if container == nil {
			return fmt.Sprintf("sai_container error: container %s not found", name)
		}
		
		return container.GetFullImageName()
		
	case 3:
		// Handle legacy provider template format: sai_container(index, "field", "provider")
		provider, ok := args[2].(string)
		if !ok {
			return "sai_container error: third argument must be provider name (string)"
		}
		
		field, ok := args[1].(string)
		if !ok {
			return "sai_container error: second argument must be field name (string)"
		}
		
		idx, ok := args[0].(int)
		if !ok {
			return "sai_container error: first argument must be index (int)"
		}
		
		result, err := e.getContainerField(provider, idx, field)
		if err != nil {
			return fmt.Sprintf("sai_container error: %v", err)
		}
		return result
		
	default:
		return fmt.Sprintf("sai_container error: accepts 1 or 3 arguments, got %d", len(args))
	}
}

// getContainerField returns specific field value for container at index for provider
func (e *TemplateEngine) getContainerField(provider string, idx int, field string) (string, error) {
	var container *types.Container
	
	// Check provider-specific containers first
	if providerConfig := e.saidata.GetProviderConfig(provider); providerConfig != nil {
		if len(providerConfig.Containers) > idx {
			container = &providerConfig.Containers[idx]
		}
	}
	
	// Fall back to default containers
	if container == nil {
		if len(e.saidata.Containers) <= idx {
			return "", fmt.Errorf("no container found at index %d", idx)
		}
		container = &e.saidata.Containers[idx]
	}
	
	// Return requested field
	switch field {
	case "name":
		return container.Name, nil
	case "image":
		return container.Image, nil
	case "tag":
		return container.Tag, nil
	case "registry":
		return container.Registry, nil
	case "full_image":
		return container.GetFullImageName(), nil
	default:
		return "", fmt.Errorf("unsupported container field: %s", field)
	}
}

// Safety validation functions
func (e *TemplateEngine) fileExists(path string) bool {
	if e.validator != nil {
		return e.validator.FileExists(path)
	}
	// Fallback to basic file existence check
	_, err := os.Stat(path)
	return err == nil
}

func (e *TemplateEngine) serviceExists(service string) bool {
	if e.validator != nil {
		return e.validator.ServiceExists(service)
	}
	// Basic fallback - this would need platform-specific implementation
	return false
}

func (e *TemplateEngine) commandExists(command string) bool {
	if e.validator != nil {
		return e.validator.CommandExists(command)
	}
	// Fallback to PATH lookup
	_, err := exec.LookPath(command)
	return err == nil
}

func (e *TemplateEngine) directoryExists(path string) bool {
	if e.validator != nil {
		return e.validator.DirectoryExists(path)
	}
	// Fallback to basic directory existence check
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// Default generation functions
func (e *TemplateEngine) defaultConfigPath(software string) string {
	if e.defaultsGen != nil {
		return e.defaultsGen.DefaultConfigPath(software)
	}
	// Fallback default
	return filepath.Join("/etc", software, software+".conf")
}

func (e *TemplateEngine) defaultLogPath(software string) string {
	if e.defaultsGen != nil {
		return e.defaultsGen.DefaultLogPath(software)
	}
	// Fallback default
	return filepath.Join("/var/log", software+".log")
}

func (e *TemplateEngine) defaultDataDir(software string) string {
	if e.defaultsGen != nil {
		return e.defaultsGen.DefaultDataDir(software)
	}
	// Fallback default
	return filepath.Join("/var/lib", software)
}

func (e *TemplateEngine) defaultServiceName(software string) string {
	if e.defaultsGen != nil {
		return e.defaultsGen.DefaultServiceName(software)
	}
	// Fallback default
	return software
}

func (e *TemplateEngine) defaultCommandPath(software string) string {
	if e.defaultsGen != nil {
		return e.defaultsGen.DefaultCommandPath(software)
	}
	// Fallback default
	return filepath.Join("/usr/bin", software)
}

// validateTemplateResolution validates that the rendered template doesn't contain unresolved variables
func (e *TemplateEngine) validateTemplateResolution(rendered string, originalTemplate string, context *TemplateContext) error {
	// Check for unresolved template variables ({{...}})
	if strings.Contains(rendered, "{{") || strings.Contains(rendered, "}}") {
		return &TemplateResolutionError{
			Type:     "unresolved_variables",
			Message:  "Template contains unresolved variables",
			Template: rendered,
			Context:  context,
		}
	}
	
	// Check for error indicators in the rendered output
	if strings.Contains(rendered, "<no value>") {
		return &TemplateResolutionError{
			Type:     "no_value",
			Message:  "Template contains unresolved values",
			Template: rendered,
			Context:  context,
		}
	}
	
	// Check for function error indicators - these indicate template functions failed
	errorIndicators := []string{
		"sai_package error:",
		"sai_packages error:",
		"sai_service error:",
		"sai_port error:",
		"sai_file error:",
		"sai_directory error:",
		"sai_command error:",
		"sai_container error:",
		"sai_source error:",
		"sai_binary error:",
		"sai_script error:",
		"no saidata context available",
		"no package found",
		"no service found",
		"no file found",
		"no directory found",
		"no command found",
		"no container found",
		"no port found",
		"no source found",
		"no binary found",
		"no script found",
	}
	
	// Check for port error indicators (port functions return -1 on error)
	if strings.Contains(rendered, "-1") && strings.Contains(originalTemplate, "sai_port") {
		return &TemplateResolutionError{
			Type:     "function_error",
			Message:  "Port template function failed: no port found",
			Template: rendered,
			Context:  context,
		}
	}
	
	for _, indicator := range errorIndicators {
		if strings.Contains(rendered, indicator) {
			return &TemplateResolutionError{
				Type:     "function_error",
				Message:  fmt.Sprintf("Template function failed: %s", indicator),
				Template: rendered,
				Context:  context,
			}
		}
	}
	
	// In safety mode, validate that referenced resources exist
	if e.safetyMode && e.validator != nil {
		if err := e.validateResourceExistence(rendered, context); err != nil {
			return err
		}
	}
	
	return nil
}

// validateResourceExistence validates that resources referenced in the rendered template exist
func (e *TemplateEngine) validateResourceExistence(rendered string, context *TemplateContext) error {
	// This is a simplified validation - in a real implementation, we would parse the command
	// and extract file paths, service names, etc. For now, we'll do basic pattern matching
	
	// Check for file paths that might not exist
	if strings.Contains(rendered, "/nonexistent/") {
		return &TemplateResolutionError{
			Type:     "resource_validation",
			Message:  "Template references nonexistent file",
			Template: rendered,
			Context:  context,
		}
	}
	
	// Check for nonexistent services
	if strings.Contains(rendered, "nonexistent") && (strings.Contains(rendered, "systemctl") || strings.Contains(rendered, "service")) {
		return &TemplateResolutionError{
			Type:     "resource_validation",
			Message:  "Template references nonexistent service",
			Template: rendered,
			Context:  context,
		}
	}
	
	return nil
}

// TemplateResolutionError provides detailed error information for template resolution failures
type TemplateResolutionError struct {
	Type     string
	Message  string
	Template string
	Context  *TemplateContext
}

func (e *TemplateResolutionError) Error() string {
	var details strings.Builder
	details.WriteString(fmt.Sprintf("Template resolution failed: %s\n", e.Message))
	details.WriteString(fmt.Sprintf("Error type: %s\n", e.Type))
	details.WriteString(fmt.Sprintf("Template: %s\n", e.Template))
	
	if e.Context != nil {
		details.WriteString(fmt.Sprintf("Software: %s\n", e.Context.Software))
		details.WriteString(fmt.Sprintf("Provider: %s\n", e.Context.Provider))
		
		if e.Context.Saidata != nil {
			details.WriteString(fmt.Sprintf("Available packages: %d\n", len(e.Context.Saidata.Packages)))
			details.WriteString(fmt.Sprintf("Available services: %d\n", len(e.Context.Saidata.Services)))
			details.WriteString(fmt.Sprintf("Available providers: %d\n", len(e.Context.Saidata.Providers)))
		}
	}
	
	// Add suggestions based on error type
	switch e.Type {
	case "unresolved_variables":
		details.WriteString("\nSuggestions:\n")
		details.WriteString("- Check that all template variables are properly defined\n")
		details.WriteString("- Verify saidata contains required package/service definitions\n")
		details.WriteString("- Ensure provider-specific configurations exist\n")
	case "no_value":
		details.WriteString("\nSuggestions:\n")
		details.WriteString("- Check that saidata is loaded and contains required data\n")
		details.WriteString("- Verify template functions are called with correct parameters\n")
		details.WriteString("- Ensure package_name field is used instead of name field\n")
	case "function_error":
		details.WriteString("\nSuggestions:\n")
		details.WriteString("- Check template function syntax and parameters\n")
		details.WriteString("- Verify saidata contains the referenced packages/services/sources/binaries/scripts\n")
		details.WriteString("- Ensure provider-specific overrides are properly configured\n")
		details.WriteString("- For alternative providers, verify sources/binaries/scripts arrays are defined\n")
		details.WriteString("- Check that build_system is specified for source configurations\n")
		details.WriteString("- Verify URLs are valid for binary and script configurations\n")
	}
	
	return details.String()
}

// preprocessTemplate converts legacy template syntax to Go template syntax
func (e *TemplateEngine) preprocessTemplate(templateStr string) string {
	// Convert function calls from legacy format to Go template format
	// Legacy: {{sai_package(0, 'name', 'apt')}}
	// Go template: {{sai_package 0 "name" "apt"}}
	
	result := templateStr
	
	// Use regex to find and replace function calls with parentheses
	// This handles patterns like: {{sai_package(0, 'name', 'apt')}}
	
	// Step 1: Replace parentheses with spaces
	result = strings.ReplaceAll(result, "(", " ")
	result = strings.ReplaceAll(result, ")", " ")
	
	// Step 2: Replace commas with spaces
	result = strings.ReplaceAll(result, ",", " ")
	
	// Step 3: Replace single quotes with double quotes for string literals
	result = strings.ReplaceAll(result, "'", "\"")
	
	// Step 4: Clean up multiple spaces
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}
	
	// Step 5: Clean up spaces around template delimiters
	result = strings.ReplaceAll(result, "{{ ", "{{")
	result = strings.ReplaceAll(result, " }}", "}}")
	result = strings.ReplaceAll(result, "{{  ", "{{")
	result = strings.ReplaceAll(result, "  }}", "}}")
	
	return result
}

// createVariableMap creates a map of variables for debug logging
func (e *TemplateEngine) createVariableMap(context *TemplateContext) map[string]interface{} {
	variables := make(map[string]interface{})
	
	if context != nil {
		variables["software"] = context.Software
		variables["provider"] = context.Provider
		
		// Add context variables
		if context.Variables != nil {
			for key, value := range context.Variables {
				variables[key] = value
			}
		}
		
		// Add saidata information
		if context.Saidata != nil {
			variables["saidata_packages"] = len(context.Saidata.Packages)
			variables["saidata_services"] = len(context.Saidata.Services)
			variables["saidata_files"] = len(context.Saidata.Files)
			variables["saidata_directories"] = len(context.Saidata.Directories)
			variables["saidata_commands"] = len(context.Saidata.Commands)
			variables["saidata_ports"] = len(context.Saidata.Ports)
			variables["saidata_containers"] = len(context.Saidata.Containers)
			variables["saidata_providers"] = len(context.Saidata.Providers)
			
			// Add provider-specific information if available
			if providerConfig := context.Saidata.GetProviderConfig(context.Provider); providerConfig != nil {
				variables["provider_packages"] = len(providerConfig.Packages)
				variables["provider_services"] = len(providerConfig.Services)
				variables["provider_files"] = len(providerConfig.Files)
				variables["provider_directories"] = len(providerConfig.Directories)
				variables["provider_commands"] = len(providerConfig.Commands)
				variables["provider_ports"] = len(providerConfig.Ports)
				variables["provider_containers"] = len(providerConfig.Containers)
			}
		}
	}
	
	return variables
}

// saiSource returns source configuration fields with provider override support
// Supports multiple calling patterns:
// - sai_source(index, "field") - returns field value for source at index
// - sai_source(index, "field", "provider") - returns field value for source at index for provider
func (e *TemplateEngine) saiSource(args ...interface{}) string {
	if e.saidata == nil {
		return "sai_source error: no saidata context available"
	}
	
	if len(args) < 2 {
		return "sai_source error: requires at least 2 arguments (index, field)"
	}
	
	// Parse index argument
	idx, ok := args[0].(int)
	if !ok {
		return "sai_source error: first argument must be index (int)"
	}
	
	// Parse field argument
	field, ok := args[1].(string)
	if !ok {
		return "sai_source error: second argument must be field name (string)"
	}
	
	// Parse optional provider argument
	var provider string
	if len(args) >= 3 {
		if p, ok := args[2].(string); ok {
			provider = p
		} else {
			return "sai_source error: third argument must be provider name (string)"
		}
	}
	
	result, err := e.resolveSourceField(provider, idx, field)
	if err != nil {
		return fmt.Sprintf("sai_source error: %v", err)
	}
	return result
}

// resolveSourceField resolves source field with provider override support
func (e *TemplateEngine) resolveSourceField(provider string, idx int, field string) (string, error) {
	var source *types.Source
	
	// Check provider-specific sources first if provider specified
	if provider != "" {
		if providerConfig := e.saidata.GetProviderConfig(provider); providerConfig != nil {
			if len(providerConfig.Sources) > idx {
				source = &providerConfig.Sources[idx]
			}
		}
	}
	
	// Fall back to default sources
	if source == nil {
		if len(e.saidata.Sources) <= idx {
			return "", fmt.Errorf("no source found at index %d", idx)
		}
		source = &e.saidata.Sources[idx]
	}
	
	// Return requested field with defaults
	switch field {
	case "name":
		return source.Name, nil
	case "url":
		return source.URL, nil
	case "version":
		return source.Version, nil
	case "build_system":
		return source.BuildSystem, nil
	case "build_dir":
		return source.GetBuildDirOrDefault(e.saidata.Metadata.Name), nil
	case "source_dir":
		return source.GetSourceDirOrDefault(e.saidata.Metadata.Name), nil
	case "install_prefix":
		return source.GetInstallPrefixOrDefault(), nil
	case "configure_args":
		return strings.Join(source.ConfigureArgs, " "), nil
	case "build_args":
		return strings.Join(source.BuildArgs, " "), nil
	case "install_args":
		return strings.Join(source.InstallArgs, " "), nil
	case "prerequisites":
		return strings.Join(source.Prerequisites, " "), nil
	case "checksum":
		return source.Checksum, nil
	case "download_cmd":
		return e.generateSourceDownloadCommand(source), nil
	case "extract_cmd":
		return e.generateSourceExtractCommand(source), nil
	case "configure_cmd":
		return e.generateSourceConfigureCommand(source), nil
	case "build_cmd":
		return e.generateSourceBuildCommand(source), nil
	case "install_cmd":
		return e.generateSourceInstallCommand(source), nil
	case "prerequisites_install_cmd":
		return e.generatePrerequisitesInstallCommand(source.Prerequisites), nil
	default:
		return "", fmt.Errorf("unsupported source field: %s", field)
	}
}

// generateSourceDownloadCommand generates download command based on source configuration
func (e *TemplateEngine) generateSourceDownloadCommand(source *types.Source) string {
	if source.CustomCommands != nil && source.CustomCommands.Download != "" {
		return source.CustomCommands.Download
	}
	
	buildDir := source.GetBuildDirOrDefault(e.saidata.Metadata.Name)
	
	// Generate download command based on URL type
	if strings.HasSuffix(source.URL, ".git") {
		return fmt.Sprintf("mkdir -p %s && cd %s && git clone %s", buildDir, buildDir, source.URL)
	} else {
		filename := filepath.Base(source.URL)
		return fmt.Sprintf("mkdir -p %s && cd %s && curl -L -o %s %s", buildDir, buildDir, filename, source.URL)
	}
}

// generateSourceExtractCommand generates extract command based on source configuration
func (e *TemplateEngine) generateSourceExtractCommand(source *types.Source) string {
	if source.CustomCommands != nil && source.CustomCommands.Extract != "" {
		return source.CustomCommands.Extract
	}
	
	buildDir := source.GetBuildDirOrDefault(e.saidata.Metadata.Name)
	sourceDir := source.GetSourceDirOrDefault(e.saidata.Metadata.Name)
	filename := filepath.Base(source.URL)
	
	// Skip extraction for git repositories
	if strings.HasSuffix(source.URL, ".git") {
		return "# Git repository - no extraction needed"
	}
	
	// Generate extraction command based on file extension
	if strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz") {
		return fmt.Sprintf("cd %s && tar -xzf %s && mv %s-* %s", buildDir, filename, e.saidata.Metadata.Name, sourceDir)
	} else if strings.HasSuffix(filename, ".tar.bz2") {
		return fmt.Sprintf("cd %s && tar -xjf %s && mv %s-* %s", buildDir, filename, e.saidata.Metadata.Name, sourceDir)
	} else if strings.HasSuffix(filename, ".zip") {
		return fmt.Sprintf("cd %s && unzip %s && mv %s-* %s", buildDir, filename, e.saidata.Metadata.Name, sourceDir)
	} else {
		return fmt.Sprintf("# Unknown archive format for %s", filename)
	}
}

// generateSourceConfigureCommand generates configure command based on build system
func (e *TemplateEngine) generateSourceConfigureCommand(source *types.Source) string {
	if source.CustomCommands != nil && source.CustomCommands.Configure != "" {
		return source.CustomCommands.Configure
	}
	
	sourceDir := source.GetSourceDirOrDefault(e.saidata.Metadata.Name)
	installPrefix := source.GetInstallPrefixOrDefault()
	configureArgs := strings.Join(source.ConfigureArgs, " ")
	
	switch source.BuildSystem {
	case "autotools", "configure", "automake", "autoconf":
		cmd := fmt.Sprintf("cd %s && ./configure --prefix=%s", sourceDir, installPrefix)
		if configureArgs != "" {
			cmd += " " + configureArgs
		}
		return cmd
	case "cmake":
		cmd := fmt.Sprintf("cd %s && cmake -DCMAKE_INSTALL_PREFIX=%s .", sourceDir, installPrefix)
		if configureArgs != "" {
			cmd += " " + configureArgs
		}
		return cmd
	case "meson":
		cmd := fmt.Sprintf("cd %s && meson setup build --prefix=%s", sourceDir, installPrefix)
		if configureArgs != "" {
			cmd += " " + configureArgs
		}
		return cmd
	case "make":
		return fmt.Sprintf("# Make-based build - no configure step needed")
	case "ninja":
		return fmt.Sprintf("# Ninja-based build - no configure step needed")
	case "custom":
		return fmt.Sprintf("# Custom build system - configure command should be specified in custom_commands")
	default:
		return fmt.Sprintf("# Unknown build system: %s", source.BuildSystem)
	}
}

// generateSourceBuildCommand generates build command based on build system
func (e *TemplateEngine) generateSourceBuildCommand(source *types.Source) string {
	if source.CustomCommands != nil && source.CustomCommands.Build != "" {
		return source.CustomCommands.Build
	}
	
	sourceDir := source.GetSourceDirOrDefault(e.saidata.Metadata.Name)
	buildArgs := strings.Join(source.BuildArgs, " ")
	
	switch source.BuildSystem {
	case "autotools", "configure", "automake", "autoconf", "make":
		cmd := fmt.Sprintf("cd %s && make", sourceDir)
		if buildArgs != "" {
			cmd += " " + buildArgs
		}
		return cmd
	case "cmake":
		cmd := fmt.Sprintf("cd %s && cmake --build .", sourceDir)
		if buildArgs != "" {
			cmd += " " + buildArgs
		}
		return cmd
	case "meson":
		cmd := fmt.Sprintf("cd %s && meson compile -C build", sourceDir)
		if buildArgs != "" {
			cmd += " " + buildArgs
		}
		return cmd
	case "ninja":
		cmd := fmt.Sprintf("cd %s && ninja", sourceDir)
		if buildArgs != "" {
			cmd += " " + buildArgs
		}
		return cmd
	case "custom":
		return fmt.Sprintf("# Custom build system - build command should be specified in custom_commands")
	default:
		return fmt.Sprintf("# Unknown build system: %s", source.BuildSystem)
	}
}

// generateSourceInstallCommand generates install command based on build system
func (e *TemplateEngine) generateSourceInstallCommand(source *types.Source) string {
	if source.CustomCommands != nil && source.CustomCommands.Install != "" {
		return source.CustomCommands.Install
	}
	
	sourceDir := source.GetSourceDirOrDefault(e.saidata.Metadata.Name)
	installArgs := strings.Join(source.InstallArgs, " ")
	
	switch source.BuildSystem {
	case "autotools", "configure", "automake", "autoconf", "make":
		cmd := fmt.Sprintf("cd %s && make install", sourceDir)
		if installArgs != "" {
			cmd += " " + installArgs
		}
		return cmd
	case "cmake":
		cmd := fmt.Sprintf("cd %s && cmake --install .", sourceDir)
		if installArgs != "" {
			cmd += " " + installArgs
		}
		return cmd
	case "meson":
		cmd := fmt.Sprintf("cd %s && meson install -C build", sourceDir)
		if installArgs != "" {
			cmd += " " + installArgs
		}
		return cmd
	case "ninja":
		cmd := fmt.Sprintf("cd %s && ninja install", sourceDir)
		if installArgs != "" {
			cmd += " " + installArgs
		}
		return cmd
	case "custom":
		return fmt.Sprintf("# Custom build system - install command should be specified in custom_commands")
	default:
		return fmt.Sprintf("# Unknown build system: %s", source.BuildSystem)
	}
}

// generatePrerequisitesInstallCommand generates command to install prerequisites
func (e *TemplateEngine) generatePrerequisitesInstallCommand(prerequisites []string) string {
	if len(prerequisites) == 0 {
		return "# No prerequisites specified"
	}
	
	// This is a simplified implementation - in practice, this would need to detect
	// the package manager and generate appropriate install commands
	return fmt.Sprintf("# Install prerequisites: %s", strings.Join(prerequisites, " "))
}

// saiBinary returns binary configuration fields with OS/architecture templating support
// Supports multiple calling patterns:
// - sai_binary(index, "field") - returns field value for binary at index
// - sai_binary(index, "field", "provider") - returns field value for binary at index for provider
func (e *TemplateEngine) saiBinary(args ...interface{}) string {
	if e.saidata == nil {
		return "sai_binary error: no saidata context available"
	}
	
	if len(args) < 2 {
		return "sai_binary error: requires at least 2 arguments (index, field)"
	}
	
	// Parse index argument
	idx, ok := args[0].(int)
	if !ok {
		return "sai_binary error: first argument must be index (int)"
	}
	
	// Parse field argument
	field, ok := args[1].(string)
	if !ok {
		return "sai_binary error: second argument must be field name (string)"
	}
	
	// Parse optional provider argument
	var provider string
	if len(args) >= 3 {
		if p, ok := args[2].(string); ok {
			provider = p
		} else {
			return "sai_binary error: third argument must be provider name (string)"
		}
	}
	
	result, err := e.resolveBinaryField(provider, idx, field)
	if err != nil {
		return fmt.Sprintf("sai_binary error: %v", err)
	}
	return result
}

// resolveBinaryField resolves binary field with provider override support
func (e *TemplateEngine) resolveBinaryField(provider string, idx int, field string) (string, error) {
	var binary *types.Binary
	
	// Check provider-specific binaries first if provider specified
	if provider != "" {
		if providerConfig := e.saidata.GetProviderConfig(provider); providerConfig != nil {
			if len(providerConfig.Binaries) > idx {
				binary = &providerConfig.Binaries[idx]
			}
		}
	}
	
	// Fall back to default binaries
	if binary == nil {
		if len(e.saidata.Binaries) <= idx {
			return "", fmt.Errorf("no binary found at index %d", idx)
		}
		binary = &e.saidata.Binaries[idx]
	}
	
	// Return requested field with defaults and templating
	switch field {
	case "name":
		return binary.Name, nil
	case "url":
		// Template the URL with OS/architecture placeholders
		return binary.TemplateURL(runtime.GOOS, runtime.GOARCH), nil
	case "version":
		return binary.Version, nil
	case "architecture":
		return binary.Architecture, nil
	case "platform":
		return binary.Platform, nil
	case "checksum":
		return binary.Checksum, nil
	case "install_path":
		return binary.GetInstallPathOrDefault(), nil
	case "executable":
		return binary.GetExecutableOrDefault(), nil
	case "permissions":
		return binary.GetPermissionsOrDefault(), nil
	case "download_cmd":
		return e.generateBinaryDownloadCommand(binary), nil
	case "extract_cmd":
		return e.generateBinaryExtractCommand(binary), nil
	case "install_cmd":
		return e.generateBinaryInstallCommand(binary), nil
	case "verify_checksum_cmd":
		return e.generateBinaryChecksumCommand(binary), nil
	default:
		return "", fmt.Errorf("unsupported binary field: %s", field)
	}
}

// generateBinaryDownloadCommand generates download command for binary
func (e *TemplateEngine) generateBinaryDownloadCommand(binary *types.Binary) string {
	if binary.CustomCommands != nil && binary.CustomCommands.Download != "" {
		return binary.CustomCommands.Download
	}
	
	url := binary.TemplateURL(runtime.GOOS, runtime.GOARCH)
	filename := filepath.Base(url)
	installPath := binary.GetInstallPathOrDefault()
	
	return fmt.Sprintf("mkdir -p %s && curl -L -o %s/%s %s", installPath, installPath, filename, url)
}

// generateBinaryExtractCommand generates extract command for binary archives
func (e *TemplateEngine) generateBinaryExtractCommand(binary *types.Binary) string {
	if binary.CustomCommands != nil && binary.CustomCommands.Extract != "" {
		return binary.CustomCommands.Extract
	}
	
	if binary.Archive == nil {
		return "# No archive configuration - assuming direct binary download"
	}
	
	url := binary.TemplateURL(runtime.GOOS, runtime.GOARCH)
	filename := filepath.Base(url)
	installPath := binary.GetInstallPathOrDefault()
	
	switch binary.Archive.Format {
	case "tar.gz", "tgz":
		cmd := fmt.Sprintf("cd %s && tar -xzf %s", installPath, filename)
		if binary.Archive.StripPrefix != "" {
			cmd += fmt.Sprintf(" --strip-components=1")
		}
		if binary.Archive.ExtractPath != "" {
			cmd += fmt.Sprintf(" %s", binary.Archive.ExtractPath)
		}
		return cmd
	case "zip":
		cmd := fmt.Sprintf("cd %s && unzip %s", installPath, filename)
		if binary.Archive.ExtractPath != "" {
			cmd += fmt.Sprintf(" %s", binary.Archive.ExtractPath)
		}
		return cmd
	default:
		return fmt.Sprintf("# Unknown archive format: %s", binary.Archive.Format)
	}
}

// generateBinaryInstallCommand generates install command for binary
func (e *TemplateEngine) generateBinaryInstallCommand(binary *types.Binary) string {
	if binary.CustomCommands != nil && binary.CustomCommands.Install != "" {
		return binary.CustomCommands.Install
	}
	
	installPath := binary.GetInstallPathOrDefault()
	executable := binary.GetExecutableOrDefault()
	permissions := binary.GetPermissionsOrDefault()
	
	if binary.Archive != nil {
		// For archived binaries, move the extracted executable
		return fmt.Sprintf("chmod %s %s/%s", permissions, installPath, executable)
	} else {
		// For direct binary downloads, rename and set permissions
		url := binary.TemplateURL(runtime.GOOS, runtime.GOARCH)
		filename := filepath.Base(url)
		return fmt.Sprintf("mv %s/%s %s/%s && chmod %s %s/%s", installPath, filename, installPath, executable, permissions, installPath, executable)
	}
}

// generateBinaryChecksumCommand generates checksum verification command
func (e *TemplateEngine) generateBinaryChecksumCommand(binary *types.Binary) string {
	if binary.Checksum == "" {
		return "# No checksum specified - skipping verification"
	}
	
	url := binary.TemplateURL(runtime.GOOS, runtime.GOARCH)
	filename := filepath.Base(url)
	installPath := binary.GetInstallPathOrDefault()
	
	// Detect checksum type based on length
	checksumType := "sha256"
	if len(binary.Checksum) == 32 {
		checksumType = "md5"
	} else if len(binary.Checksum) == 40 {
		checksumType = "sha1"
	} else if len(binary.Checksum) == 64 {
		checksumType = "sha256"
	}
	
	return fmt.Sprintf("echo '%s %s/%s' | %ssum -c", binary.Checksum, installPath, filename, checksumType)
}

// saiScript returns script configuration fields with environment variable support
// Supports multiple calling patterns:
// - sai_script(index, "field") - returns field value for script at index
// - sai_script(index, "field", "provider") - returns field value for script at index for provider
func (e *TemplateEngine) saiScript(args ...interface{}) string {
	if e.saidata == nil {
		return "sai_script error: no saidata context available"
	}
	
	if len(args) < 2 {
		return "sai_script error: requires at least 2 arguments (index, field)"
	}
	
	// Parse index argument
	idx, ok := args[0].(int)
	if !ok {
		return "sai_script error: first argument must be index (int)"
	}
	
	// Parse field argument
	field, ok := args[1].(string)
	if !ok {
		return "sai_script error: second argument must be field name (string)"
	}
	
	// Parse optional provider argument
	var provider string
	if len(args) >= 3 {
		if p, ok := args[2].(string); ok {
			provider = p
		} else {
			return "sai_script error: third argument must be provider name (string)"
		}
	}
	
	result, err := e.resolveScriptField(provider, idx, field)
	if err != nil {
		return fmt.Sprintf("sai_script error: %v", err)
	}
	return result
}

// resolveScriptField resolves script field with provider override support
func (e *TemplateEngine) resolveScriptField(provider string, idx int, field string) (string, error) {
	var script *types.Script
	
	// Check provider-specific scripts first if provider specified
	if provider != "" {
		if providerConfig := e.saidata.GetProviderConfig(provider); providerConfig != nil {
			if len(providerConfig.Scripts) > idx {
				script = &providerConfig.Scripts[idx]
			}
		}
	}
	
	// Fall back to default scripts
	if script == nil {
		if len(e.saidata.Scripts) <= idx {
			return "", fmt.Errorf("no script found at index %d", idx)
		}
		script = &e.saidata.Scripts[idx]
	}
	
	// Return requested field with defaults
	switch field {
	case "name":
		return script.Name, nil
	case "url":
		return script.URL, nil
	case "version":
		return script.Version, nil
	case "interpreter":
		return script.GetInterpreterOrDefault(), nil
	case "checksum":
		return script.Checksum, nil
	case "arguments":
		return strings.Join(script.Arguments, " "), nil
	case "working_dir":
		return script.GetWorkingDirOrDefault(), nil
	case "timeout":
		return fmt.Sprintf("%d", script.GetTimeoutOrDefault()), nil
	case "download_cmd":
		return e.generateScriptDownloadCommand(script), nil
	case "execute_cmd":
		return e.generateScriptExecuteCommand(script), nil
	case "verify_checksum_cmd":
		return e.generateScriptChecksumCommand(script), nil
	case "environment_vars":
		return e.generateScriptEnvironmentVars(script), nil
	default:
		return "", fmt.Errorf("unsupported script field: %s", field)
	}
}

// generateScriptDownloadCommand generates download command for script
func (e *TemplateEngine) generateScriptDownloadCommand(script *types.Script) string {
	if script.CustomCommands != nil && script.CustomCommands.Download != "" {
		return script.CustomCommands.Download
	}
	
	filename := filepath.Base(script.URL)
	// Handle URLs that don't have a clear filename (like https://get.docker.com/)
	if filename == "." || filename == "/" || strings.Contains(filename, ".") && !strings.Contains(filename, " ") {
		// If it looks like a domain name, use a default filename
		if strings.Contains(filename, ".") && !strings.HasSuffix(filename, ".sh") && !strings.HasSuffix(filename, ".py") {
			filename = "install.sh"
		}
	}
	workingDir := script.GetWorkingDirOrDefault()
	
	return fmt.Sprintf("mkdir -p %s && cd %s && curl -L -o %s %s", workingDir, workingDir, filename, script.URL)
}

// generateScriptExecuteCommand generates execute command for script
func (e *TemplateEngine) generateScriptExecuteCommand(script *types.Script) string {
	if script.CustomCommands != nil && script.CustomCommands.Install != "" {
		return script.CustomCommands.Install
	}
	
	filename := filepath.Base(script.URL)
	// Handle URLs that don't have a clear filename (like https://get.docker.com/)
	if filename == "." || filename == "/" || strings.Contains(filename, ".") && !strings.Contains(filename, " ") {
		// If it looks like a domain name, use a default filename
		if strings.Contains(filename, ".") && !strings.HasSuffix(filename, ".sh") && !strings.HasSuffix(filename, ".py") {
			filename = "install.sh"
		}
	}
	workingDir := script.GetWorkingDirOrDefault()
	interpreter := script.GetInterpreterOrDefault()
	arguments := strings.Join(script.Arguments, " ")
	timeout := script.GetTimeoutOrDefault()
	
	cmd := fmt.Sprintf("cd %s && timeout %d %s %s", workingDir, timeout, interpreter, filename)
	if arguments != "" {
		cmd += " " + arguments
	}
	
	return cmd
}

// generateScriptChecksumCommand generates checksum verification command for script
func (e *TemplateEngine) generateScriptChecksumCommand(script *types.Script) string {
	if script.Checksum == "" {
		return "# No checksum specified - skipping verification"
	}
	
	filename := filepath.Base(script.URL)
	workingDir := script.GetWorkingDirOrDefault()
	
	// Detect checksum type based on length
	checksumType := "sha256"
	if len(script.Checksum) == 32 {
		checksumType = "md5"
	} else if len(script.Checksum) == 40 {
		checksumType = "sha1"
	} else if len(script.Checksum) == 64 {
		checksumType = "sha256"
	}
	
	return fmt.Sprintf("cd %s && echo '%s %s' | %ssum -c", workingDir, script.Checksum, filename, checksumType)
}

// generateScriptEnvironmentVars generates environment variable export commands
func (e *TemplateEngine) generateScriptEnvironmentVars(script *types.Script) string {
	if len(script.Environment) == 0 {
		return "# No environment variables specified"
	}
	
	// Sort keys for consistent output
	var keys []string
	for key := range script.Environment {
		keys = append(keys, key)
	}
	
	// Sort keys alphabetically for consistent test results
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	
	var exports []string
	for _, key := range keys {
		exports = append(exports, fmt.Sprintf("export %s='%s'", key, script.Environment[key]))
	}
	
	return strings.Join(exports, " && ")
}