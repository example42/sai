package saidata

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"sai/internal/types"
)

// ValidationResult contains detailed validation results
type ValidationResult struct {
	Valid           bool     `json:"valid"`
	MissingFiles    []string `json:"missing_files,omitempty"`
	MissingServices []string `json:"missing_services,omitempty"`
	MissingCommands []string `json:"missing_commands,omitempty"`
	MissingDirs     []string `json:"missing_directories,omitempty"`
	ClosedPorts     []int    `json:"closed_ports,omitempty"`
	Warnings        []string `json:"warnings,omitempty"`
	CanProceed      bool     `json:"can_proceed"`
	Details         string   `json:"details,omitempty"`
}

// SystemResourceValidator validates system resources
type SystemResourceValidator struct {
	timeout time.Duration
}

// NewSystemResourceValidator creates a new system resource validator
func NewSystemResourceValidator() *SystemResourceValidator {
	return &SystemResourceValidator{
		timeout: 5 * time.Second,
	}
}

// ValidateResources validates all resources in saidata and returns detailed results
func (v *SystemResourceValidator) ValidateResources(saidata *types.SoftwareData, action string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:      true,
		CanProceed: true,
	}

	// Validate files
	for _, file := range saidata.Files {
		if !v.ValidateFile(file.Path) {
			result.MissingFiles = append(result.MissingFiles, file.Path)
			result.Valid = false
		}
	}

	// Validate services
	for _, service := range saidata.Services {
		serviceName := service.GetServiceNameOrDefault()
		if !v.ValidateService(serviceName) {
			result.MissingServices = append(result.MissingServices, serviceName)
			result.Valid = false
		}
	}

	// Validate commands
	for _, command := range saidata.Commands {
		commandPath := command.GetPathOrDefault()
		if !v.ValidateCommand(commandPath) {
			result.MissingCommands = append(result.MissingCommands, commandPath)
			result.Valid = false
		}
	}

	// Validate directories
	for _, directory := range saidata.Directories {
		if !v.ValidateDirectory(directory.Path) {
			result.MissingDirs = append(result.MissingDirs, directory.Path)
			result.Valid = false
		}
	}

	// Validate ports (check if they're open/listening)
	for _, port := range saidata.Ports {
		if !v.ValidatePort(port.Port) {
			result.ClosedPorts = append(result.ClosedPorts, port.Port)
			// Ports being closed is not necessarily an error for all actions
			if action == "start" || action == "status" {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Port %d is not listening", port.Port))
			}
		}
	}

	// Determine if we can proceed based on action type
	result.CanProceed = v.canProceedWithAction(action, result)
	
	// Generate details message
	result.Details = v.generateDetailsMessage(result)

	return result, nil
}

// ValidateFile checks if a file exists and is accessible
func (v *SystemResourceValidator) ValidateFile(path string) bool {
	if path == "" {
		return false
	}

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check if it's actually a file (not a directory)
	return !info.IsDir()
}

// ValidateService checks if a service exists on the system
func (v *SystemResourceValidator) ValidateService(serviceName string) bool {
	if serviceName == "" {
		return false
	}

	switch runtime.GOOS {
	case "linux":
		return v.validateLinuxService(serviceName)
	case "darwin":
		return v.validateMacOSService(serviceName)
	case "windows":
		return v.validateWindowsService(serviceName)
	default:
		return false
	}
}

// ValidateCommand checks if a command exists and is executable
func (v *SystemResourceValidator) ValidateCommand(command string) bool {
	if command == "" {
		return false
	}

	// If it's an absolute path, check directly
	if filepath.IsAbs(command) {
		return v.validateExecutablePath(command)
	}

	// Check in PATH
	_, err := exec.LookPath(command)
	return err == nil
}

// ValidateDirectory checks if a directory exists and is accessible
func (v *SystemResourceValidator) ValidateDirectory(path string) bool {
	if path == "" {
		return false
	}

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// ValidatePort checks if a port is open/listening
func (v *SystemResourceValidator) ValidatePort(port int) bool {
	if port <= 0 || port > 65535 {
		return false
	}

	// Try to connect to the port on localhost
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), v.timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	
	return true
}

// validateLinuxService checks if a systemd service exists
func (v *SystemResourceValidator) validateLinuxService(serviceName string) bool {
	// Try systemctl first
	cmd := exec.Command("systemctl", "list-unit-files", serviceName+".service")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), serviceName+".service") {
		return true
	}

	// Try service command as fallback
	cmd = exec.Command("service", serviceName, "status")
	err = cmd.Run()
	return err == nil
}

// validateMacOSService checks if a launchd service exists
func (v *SystemResourceValidator) validateMacOSService(serviceName string) bool {
	// Check launchctl
	cmd := exec.Command("launchctl", "list")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), serviceName)
}

// validateWindowsService checks if a Windows service exists
func (v *SystemResourceValidator) validateWindowsService(serviceName string) bool {
	// Use sc query command
	cmd := exec.Command("sc", "query", serviceName)
	err := cmd.Run()
	return err == nil
}

// validateExecutablePath checks if a file is executable
func (v *SystemResourceValidator) validateExecutablePath(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return false
	}

	// Check if it's executable
	if runtime.GOOS == "windows" {
		// On Windows, check file extension
		ext := strings.ToLower(filepath.Ext(path))
		return ext == ".exe" || ext == ".bat" || ext == ".cmd" || ext == ".com"
	} else {
		// On Unix-like systems, check execute permission
		return info.Mode()&0111 != 0
	}
}

// canProceedWithAction determines if an action can proceed based on validation results
func (v *SystemResourceValidator) canProceedWithAction(action string, result *ValidationResult) bool {
	// Information-only actions can always proceed
	infoActions := []string{"search", "info", "version", "status", "logs", "config", "check", "cpu", "memory", "io", "list", "stats"}
	for _, infoAction := range infoActions {
		if action == infoAction {
			return true
		}
	}

	// System-changing actions require more validation
	systemActions := []string{"install", "uninstall", "upgrade", "start", "stop", "restart", "enable", "disable"}
	for _, systemAction := range systemActions {
		if action == systemAction {
			return v.canProceedWithSystemAction(systemAction, result)
		}
	}

	// Default to allowing action if we don't recognize it
	return true
}

// canProceedWithSystemAction determines if system-changing actions can proceed
func (v *SystemResourceValidator) canProceedWithSystemAction(action string, result *ValidationResult) bool {
	switch action {
	case "install":
		// Install can proceed even if software doesn't exist yet
		return true
	case "uninstall", "upgrade":
		// These require the software to be present
		return len(result.MissingCommands) == 0 || len(result.MissingServices) == 0
	case "start", "restart", "enable":
		// Service management requires service to exist
		return len(result.MissingServices) == 0
	case "stop", "disable":
		// Stop/disable can proceed even if service isn't running
		return len(result.MissingServices) == 0
	default:
		return true
	}
}

// generateDetailsMessage creates a human-readable details message
func (v *SystemResourceValidator) generateDetailsMessage(result *ValidationResult) string {
	if result.Valid {
		return "All resources validated successfully"
	}

	var details []string

	if len(result.MissingFiles) > 0 {
		details = append(details, fmt.Sprintf("Missing files: %s", strings.Join(result.MissingFiles, ", ")))
	}

	if len(result.MissingServices) > 0 {
		details = append(details, fmt.Sprintf("Missing services: %s", strings.Join(result.MissingServices, ", ")))
	}

	if len(result.MissingCommands) > 0 {
		details = append(details, fmt.Sprintf("Missing commands: %s", strings.Join(result.MissingCommands, ", ")))
	}

	if len(result.MissingDirs) > 0 {
		details = append(details, fmt.Sprintf("Missing directories: %s", strings.Join(result.MissingDirs, ", ")))
	}

	if len(result.ClosedPorts) > 0 {
		portStrs := make([]string, len(result.ClosedPorts))
		for i, port := range result.ClosedPorts {
			portStrs[i] = strconv.Itoa(port)
		}
		details = append(details, fmt.Sprintf("Closed ports: %s", strings.Join(portStrs, ", ")))
	}

	if len(result.Warnings) > 0 {
		details = append(details, fmt.Sprintf("Warnings: %s", strings.Join(result.Warnings, "; ")))
	}

	return strings.Join(details, "; ")
}

// ValidateTemplateResolution validates that template variables can be resolved
func (v *SystemResourceValidator) ValidateTemplateResolution(template string, saidata *types.SoftwareData) error {
	// This is a placeholder for template validation logic
	// In a full implementation, this would parse the template and check
	// that all sai_* functions can resolve to actual values
	
	if saidata == nil {
		return fmt.Errorf("saidata is nil, cannot resolve template variables")
	}

	// Check for common template patterns that might fail
	if strings.Contains(template, "{{sai_package") && len(saidata.Packages) == 0 {
		return fmt.Errorf("template references sai_package but no packages defined")
	}

	if strings.Contains(template, "{{sai_service") && len(saidata.Services) == 0 {
		return fmt.Errorf("template references sai_service but no services defined")
	}

	if strings.Contains(template, "{{sai_port") && len(saidata.Ports) == 0 {
		return fmt.Errorf("template references sai_port but no ports defined")
	}

	if strings.Contains(template, "{{sai_file") && len(saidata.Files) == 0 {
		return fmt.Errorf("template references sai_file but no files defined")
	}

	return nil
}

// GetResourceStatus returns the current status of resources
func (v *SystemResourceValidator) GetResourceStatus(saidata *types.SoftwareData) (*ResourceStatus, error) {
	status := &ResourceStatus{
		Files:       make(map[string]FileStatus),
		Services:    make(map[string]ServiceStatus),
		Commands:    make(map[string]CommandStatus),
		Directories: make(map[string]DirectoryStatus),
		Ports:       make(map[int]PortStatus),
	}

	// Check file status
	for _, file := range saidata.Files {
		fileStatus := FileStatus{
			Path:   file.Path,
			Exists: v.ValidateFile(file.Path),
		}
		if fileStatus.Exists {
			if info, err := os.Stat(file.Path); err == nil {
				fileStatus.Size = info.Size()
				fileStatus.ModTime = info.ModTime()
				fileStatus.Mode = info.Mode()
			}
		}
		status.Files[file.Name] = fileStatus
	}

	// Check service status
	for _, service := range saidata.Services {
		serviceName := service.GetServiceNameOrDefault()
		serviceStatus := ServiceStatus{
			Name:   serviceName,
			Exists: v.ValidateService(serviceName),
		}
		if serviceStatus.Exists {
			serviceStatus.IsActive = v.isServiceActive(serviceName)
			serviceStatus.IsEnabled = v.isServiceEnabled(serviceName)
		}
		status.Services[service.Name] = serviceStatus
	}

	// Check command status
	for _, command := range saidata.Commands {
		commandPath := command.GetPathOrDefault()
		commandStatus := CommandStatus{
			Path:   commandPath,
			Exists: v.ValidateCommand(commandPath),
		}
		if commandStatus.Exists && filepath.IsAbs(commandPath) {
			if info, err := os.Stat(commandPath); err == nil {
				commandStatus.Size = info.Size()
				commandStatus.Mode = info.Mode()
			}
		}
		status.Commands[command.Name] = commandStatus
	}

	// Check directory status
	for _, directory := range saidata.Directories {
		dirStatus := DirectoryStatus{
			Path:   directory.Path,
			Exists: v.ValidateDirectory(directory.Path),
		}
		if dirStatus.Exists {
			if info, err := os.Stat(directory.Path); err == nil {
				dirStatus.Mode = info.Mode()
			}
		}
		status.Directories[directory.Name] = dirStatus
	}

	// Check port status
	for _, port := range saidata.Ports {
		portStatus := PortStatus{
			Port:     port.Port,
			Protocol: port.GetProtocolOrDefault(),
			IsOpen:   v.ValidatePort(port.Port),
		}
		status.Ports[port.Port] = portStatus
	}

	return status, nil
}

// isServiceActive checks if a service is currently active/running
func (v *SystemResourceValidator) isServiceActive(serviceName string) bool {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("systemctl", "is-active", serviceName)
		output, err := cmd.Output()
		return err == nil && strings.TrimSpace(string(output)) == "active"
	case "darwin":
		// For macOS, this would check launchctl
		return false // Placeholder
	case "windows":
		// For Windows, this would check service status
		return false // Placeholder
	default:
		return false
	}
}

// isServiceEnabled checks if a service is enabled to start at boot
func (v *SystemResourceValidator) isServiceEnabled(serviceName string) bool {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("systemctl", "is-enabled", serviceName)
		output, err := cmd.Output()
		return err == nil && strings.TrimSpace(string(output)) == "enabled"
	case "darwin":
		// For macOS, this would check launchctl
		return false // Placeholder
	case "windows":
		// For Windows, this would check service startup type
		return false // Placeholder
	default:
		return false
	}
}

// ResourceStatus contains the current status of all resources
type ResourceStatus struct {
	Files       map[string]FileStatus      `json:"files"`
	Services    map[string]ServiceStatus   `json:"services"`
	Commands    map[string]CommandStatus   `json:"commands"`
	Directories map[string]DirectoryStatus `json:"directories"`
	Ports       map[int]PortStatus         `json:"ports"`
}

// FileStatus represents the status of a file resource
type FileStatus struct {
	Path    string      `json:"path"`
	Exists  bool        `json:"exists"`
	Size    int64       `json:"size,omitempty"`
	ModTime time.Time   `json:"mod_time,omitempty"`
	Mode    os.FileMode `json:"mode,omitempty"`
}

// ServiceStatus represents the status of a service resource
type ServiceStatus struct {
	Name      string `json:"name"`
	Exists    bool   `json:"exists"`
	IsActive  bool   `json:"is_active"`
	IsEnabled bool   `json:"is_enabled"`
}

// CommandStatus represents the status of a command resource
type CommandStatus struct {
	Path   string      `json:"path"`
	Exists bool        `json:"exists"`
	Size   int64       `json:"size,omitempty"`
	Mode   os.FileMode `json:"mode,omitempty"`
}

// DirectoryStatus represents the status of a directory resource
type DirectoryStatus struct {
	Path   string      `json:"path"`
	Exists bool        `json:"exists"`
	Mode   os.FileMode `json:"mode,omitempty"`
}

// PortStatus represents the status of a port resource
type PortStatus struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	IsOpen   bool   `json:"is_open"`
}