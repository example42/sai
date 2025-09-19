package template

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

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

// TemplateContext holds the context for template rendering
type TemplateContext struct {
	Software string
	Provider string
	Saidata  *types.SoftwareData
	Variables map[string]string
}

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
	if context == nil {
		return "", fmt.Errorf("template context cannot be nil")
	}
	
	// Set saidata context for template functions
	e.saidata = context.Saidata
	
	// Parse the template
	tmpl, err := e.template.Clone()
	if err != nil {
		return "", fmt.Errorf("failed to clone template: %w", err)
	}
	
	tmpl, err = tmpl.Parse(templateStr)
	if err != nil {
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
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	
	result := buf.String()
	
	// Validate template resolution if safety mode is enabled
	if e.safetyMode {
		if err := e.validateTemplateResolution(result, context); err != nil {
			return "", fmt.Errorf("template validation failed: %w", err)
		}
	}
	
	return result, nil
}

// ValidateTemplate validates a template string without executing it
func (e *TemplateEngine) ValidateTemplate(templateStr string) error {
	tmpl, err := e.template.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone template: %w", err)
	}
	
	_, err = tmpl.Parse(templateStr)
	if err != nil {
		return fmt.Errorf("template syntax error: %w", err)
	}
	
	return nil
}

// createFuncMap creates the function map for template functions
func (e *TemplateEngine) createFuncMap() template.FuncMap {
	return template.FuncMap{
		// Saidata functions
		"sai_package":       e.saiPackage,
		"sai_packages":      e.saiPackages,
		"sai_service":       e.saiService,
		"sai_port":          e.saiPort,
		"sai_file":          e.saiFile,
		"sai_directory":     e.saiDirectory,
		"sai_command":       e.saiCommand,
		"sai_container":     e.saiContainer,
		
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
func (e *TemplateEngine) saiPackage(provider string, index ...int) (string, error) {
	if e.saidata == nil {
		return "", fmt.Errorf("no saidata context available")
	}
	
	idx := 0
	if len(index) > 0 {
		idx = index[0]
	}
	
	// Check provider-specific packages first
	if providerConfig := e.saidata.GetProviderConfig(provider); providerConfig != nil {
		if len(providerConfig.Packages) > idx {
			return providerConfig.Packages[idx].Name, nil
		}
	}
	
	// Fall back to default packages
	if len(e.saidata.Packages) > idx {
		return e.saidata.Packages[idx].Name, nil
	}
	
	return "", fmt.Errorf("no package found at index %d for provider %s", idx, provider)
}

// saiPackages returns all package names for a specific provider
func (e *TemplateEngine) saiPackages(provider string) ([]string, error) {
	if e.saidata == nil {
		return nil, fmt.Errorf("no saidata context available")
	}
	
	var packages []string
	
	// Check provider-specific packages first
	if providerConfig := e.saidata.GetProviderConfig(provider); providerConfig != nil {
		for _, pkg := range providerConfig.Packages {
			packages = append(packages, pkg.Name)
		}
		if len(packages) > 0 {
			return packages, nil
		}
	}
	
	// Fall back to default packages
	for _, pkg := range e.saidata.Packages {
		packages = append(packages, pkg.Name)
	}
	
	if len(packages) == 0 {
		return nil, fmt.Errorf("no packages found for provider %s", provider)
	}
	
	return packages, nil
}

// saiService returns the service name
func (e *TemplateEngine) saiService(name string) (string, error) {
	if e.saidata == nil {
		return "", fmt.Errorf("no saidata context available")
	}
	
	service := e.saidata.GetServiceByName(name)
	if service == nil {
		return "", fmt.Errorf("service %s not found", name)
	}
	
	return service.GetServiceNameOrDefault(), nil
}

// saiPort returns the port number
func (e *TemplateEngine) saiPort(index ...int) (int, error) {
	if e.saidata == nil {
		return 0, fmt.Errorf("no saidata context available")
	}
	
	idx := 0
	if len(index) > 0 {
		idx = index[0]
	}
	
	if len(e.saidata.Ports) <= idx {
		return 0, fmt.Errorf("no port found at index %d", idx)
	}
	
	return e.saidata.Ports[idx].Port, nil
}

// saiFile returns the file path
func (e *TemplateEngine) saiFile(name string) (string, error) {
	if e.saidata == nil {
		return "", fmt.Errorf("no saidata context available")
	}
	
	file := e.saidata.GetFileByName(name)
	if file == nil {
		return "", fmt.Errorf("file %s not found", name)
	}
	
	return file.Path, nil
}

// saiDirectory returns the directory path
func (e *TemplateEngine) saiDirectory(name string) (string, error) {
	if e.saidata == nil {
		return "", fmt.Errorf("no saidata context available")
	}
	
	directory := e.saidata.GetDirectoryByName(name)
	if directory == nil {
		return "", fmt.Errorf("directory %s not found", name)
	}
	
	return directory.Path, nil
}

// saiCommand returns the command path
func (e *TemplateEngine) saiCommand(name string) (string, error) {
	if e.saidata == nil {
		return "", fmt.Errorf("no saidata context available")
	}
	
	command := e.saidata.GetCommandByName(name)
	if command == nil {
		return "", fmt.Errorf("command %s not found", name)
	}
	
	return command.GetPathOrDefault(), nil
}

// saiContainer returns the container name
func (e *TemplateEngine) saiContainer(name string) (string, error) {
	if e.saidata == nil {
		return "", fmt.Errorf("no saidata context available")
	}
	
	container := e.saidata.GetContainerByName(name)
	if container == nil {
		return "", fmt.Errorf("container %s not found", name)
	}
	
	return container.GetFullImageName(), nil
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
func (e *TemplateEngine) validateTemplateResolution(rendered string, context *TemplateContext) error {
	// Check for unresolved template variables ({{...}})
	if strings.Contains(rendered, "{{") || strings.Contains(rendered, "}}") {
		return fmt.Errorf("template contains unresolved variables: %s", rendered)
	}
	
	// Check for error indicators in the rendered output
	if strings.Contains(rendered, "<no value>") {
		return fmt.Errorf("template contains unresolved values: %s", rendered)
	}
	
	return nil
}