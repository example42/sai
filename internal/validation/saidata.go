package validation

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/xeipuuv/gojsonschema"
	"sai/internal/interfaces"
	"sai/internal/types"
)

// SaidataValidator validates saidata against the JSON schema
type SaidataValidator struct {
	schemaLoader gojsonschema.JSONLoader
}

// NewSaidataValidator creates a new saidata validator
func NewSaidataValidator(schemaPath string) (*SaidataValidator, error) {
	// Load schema from file
	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	schemaLoader := gojsonschema.NewBytesLoader(schemaData)
	
	return &SaidataValidator{
		schemaLoader: schemaLoader,
	}, nil
}

// ValidateSaidata validates a saidata structure against the schema
func (v *SaidataValidator) ValidateSaidata(saidata *types.SoftwareData) error {
	// Convert saidata to JSON for validation
	jsonData, err := saidata.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert saidata to JSON: %w", err)
	}

	// Create document loader
	documentLoader := gojsonschema.NewBytesLoader(jsonData)

	// Validate
	result, err := gojsonschema.Validate(v.schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if !result.Valid() {
		var errorMessages []string
		for _, desc := range result.Errors() {
			errorMessages = append(errorMessages, desc.String())
		}
		return fmt.Errorf("saidata validation failed: %v", errorMessages)
	}

	return nil
}

// ValidateSaidataYAML validates saidata YAML data against the schema
func (v *SaidataValidator) ValidateSaidataYAML(yamlData []byte) error {
	// First parse the YAML
	saidata, err := types.LoadSoftwareDataFromYAML(yamlData)
	if err != nil {
		return fmt.Errorf("failed to parse saidata YAML: %w", err)
	}

	// Then validate against schema
	return v.ValidateSaidata(saidata)
}

// ValidateSaidataFile validates a saidata file against the schema
func (v *SaidataValidator) ValidateSaidataFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read saidata file: %w", err)
	}

	return v.ValidateSaidataYAML(data)
}

// ValidateAllSaidata validates all saidata files in a directory recursively
func (v *SaidataValidator) ValidateAllSaidata(saidataDir string) ([]ValidationResult, error) {
	var results []ValidationResult

	err := walkSaidataDirectory(saidataDir, func(filePath, relativePath string) {
		result := ValidationResult{
			File: relativePath,
		}

		err := v.ValidateSaidataFile(filePath)
		if err != nil {
			result.Valid = false
			result.Errors = []string{err.Error()}
		} else {
			result.Valid = true
		}

		results = append(results, result)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk saidata directory: %w", err)
	}

	return results, nil
}

// walkSaidataDirectory walks through saidata directory structure recursively
func walkSaidataDirectory(rootDir string, callback func(filePath, relativePath string)) error {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := fmt.Sprintf("%s/%s", rootDir, entry.Name())
		
		if entry.IsDir() {
			// Recursively walk subdirectories
			err := walkSaidataDirectory(fullPath, callback)
			if err != nil {
				return err
			}
		} else if isYAMLFile(entry.Name()) {
			// Process YAML files
			relativePath := entry.Name()
			callback(fullPath, relativePath)
		}
	}

	return nil
}

// ResourceValidator validates that resources exist on the system
type ResourceValidator struct{}

// NewResourceValidator creates a new resource validator
func NewResourceValidator() *ResourceValidator {
	return &ResourceValidator{}
}

// ValidateFile checks if a file exists and is accessible
func (r *ResourceValidator) ValidateFile(file types.File) bool {
	if file.Path == "" {
		return false
	}
	
	info, err := os.Stat(file.Path)
	if err != nil {
		return false
	}
	
	// Check if it's actually a file (not a directory)
	return !info.IsDir()
}

// ValidateDirectory checks if a directory exists and is accessible
func (r *ResourceValidator) ValidateDirectory(directory types.Directory) bool {
	if directory.Path == "" {
		return false
	}
	
	info, err := os.Stat(directory.Path)
	if err != nil {
		return false
	}
	
	// Check if it's actually a directory
	return info.IsDir()
}

// ValidateCommand checks if a command exists and is executable
func (r *ResourceValidator) ValidateCommand(command types.Command) bool {
	path := command.GetPathOrDefault()
	
	// If path is empty, return false
	if path == "" {
		return false
	}
	
	// First try to stat the path directly (for absolute paths)
	if info, err := os.Stat(path); err == nil {
		// Check if it's a file and has execute permissions
		if info.IsDir() {
			return false
		}
		
		// Check execute permissions (basic check)
		mode := info.Mode()
		return mode&0111 != 0 // Check if any execute bit is set
	}
	
	// If direct stat fails, try to find the command in PATH
	_, err := exec.LookPath(path)
	return err == nil
}

// ValidateService checks if a service exists (basic check for systemd)
func (r *ResourceValidator) ValidateService(service types.Service) bool {
	serviceName := service.GetServiceNameOrDefault()
	
	// Check if systemd service file exists
	systemdPaths := []string{
		fmt.Sprintf("/etc/systemd/system/%s.service", serviceName),
		fmt.Sprintf("/lib/systemd/system/%s.service", serviceName),
		fmt.Sprintf("/usr/lib/systemd/system/%s.service", serviceName),
	}
	
	for _, path := range systemdPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	
	return false
}

// ValidatePort checks if a port is open (basic check)
func (r *ResourceValidator) ValidatePort(port types.Port) bool {
	// This is a placeholder - in a real implementation, you'd check if the port is open
	// For now, we'll just validate that the port number is in a valid range
	return port.Port > 0 && port.Port <= 65535
}

// ValidateContainer checks if a container configuration is valid
func (r *ResourceValidator) ValidateContainer(container types.Container) bool {
	// This is a placeholder - in a real implementation, you'd check if the container exists
	// For now, we'll just validate that the container name is not empty
	return container.Name != ""
}

// ValidateSystemRequirements checks system requirements
func (r *ResourceValidator) ValidateSystemRequirements(requirements *types.Requirements) (*interfaces.SystemValidationResult, error) {
	// This is a placeholder implementation
	result := &interfaces.SystemValidationResult{
		Valid:                   true,
		InsufficientMemory:      false,
		InsufficientDisk:        false,
		MissingDependencies:     []string{},
		UnsupportedPlatform:     false,
		Warnings:                []string{},
	}
	return result, nil
}

// ValidateResources validates all resources in saidata
func (r *ResourceValidator) ValidateResources(saidata *types.SoftwareData) (*interfaces.ResourceValidationResult, error) {
	result := &interfaces.ResourceValidationResult{
		Valid: true,
	}
	
	// Validate files
	for i, file := range saidata.Files {
		exists := r.ValidateFile(file)
		saidata.Files[i].Exists = exists
		if !exists {
			result.Valid = false
			result.MissingFiles = append(result.MissingFiles, file.Path)
		}
	}
	
	// Validate directories
	for i, directory := range saidata.Directories {
		exists := r.ValidateDirectory(directory)
		saidata.Directories[i].Exists = exists
		if !exists {
			result.Valid = false
			result.MissingDirectories = append(result.MissingDirectories, directory.Path)
		}
	}
	
	// Validate commands
	for i, command := range saidata.Commands {
		exists := r.ValidateCommand(command)
		saidata.Commands[i].Exists = exists
		if !exists {
			result.Valid = false
			result.MissingCommands = append(result.MissingCommands, command.GetPathOrDefault())
		}
	}
	
	// Validate services
	for i, service := range saidata.Services {
		exists := r.ValidateService(service)
		saidata.Services[i].Exists = exists
		if !exists {
			result.Valid = false
			result.MissingServices = append(result.MissingServices, service.GetServiceNameOrDefault())
		}
	}
	
	// Validate ports
	for i, port := range saidata.Ports {
		valid := r.ValidatePort(port)
		saidata.Ports[i].IsOpen = valid
		if !valid {
			result.Valid = false
			result.InvalidPorts = append(result.InvalidPorts, port.Port)
		}
	}
	
	// Determine if we can proceed despite missing resources
	result.CanProceed = r.canProceedWithMissingResources(result)
	
	// Add warnings for missing optional resources
	if len(result.MissingFiles) > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Some configuration files are missing: %v", result.MissingFiles))
	}
	if len(result.MissingServices) > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Some services are not installed: %v", result.MissingServices))
	}
	
	return result, nil
}

// canProceedWithMissingResources determines if execution can proceed despite missing resources
func (r *ResourceValidator) canProceedWithMissingResources(result *interfaces.ResourceValidationResult) bool {
	// Allow proceeding even with missing resources for flexibility
	// This is a validation check, not a hard requirement for execution
	// The system should be able to handle missing optional resources gracefully
	return true
}



// Ensure ResourceValidator implements the interface
var _ interfaces.ResourceValidator = (*ResourceValidator)(nil)

