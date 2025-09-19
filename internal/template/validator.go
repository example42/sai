package template

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// SystemResourceValidator provides system-level resource validation
type SystemResourceValidator struct{}

// NewSystemResourceValidator creates a new system resource validator
func NewSystemResourceValidator() *SystemResourceValidator {
	return &SystemResourceValidator{}
}

// FileExists checks if a file exists on the filesystem
func (v *SystemResourceValidator) FileExists(path string) bool {
	if path == "" {
		return false
	}
	
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// ServiceExists checks if a service exists on the system
func (v *SystemResourceValidator) ServiceExists(service string) bool {
	if service == "" {
		return false
	}
	
	switch runtime.GOOS {
	case "linux":
		return v.checkLinuxService(service)
	case "darwin":
		return v.checkMacOSService(service)
	case "windows":
		return v.checkWindowsService(service)
	default:
		return false
	}
}

// CommandExists checks if a command exists in PATH or at a specific location
func (v *SystemResourceValidator) CommandExists(command string) bool {
	if command == "" {
		return false
	}
	
	// If it's an absolute path, check directly
	if strings.HasPrefix(command, "/") || strings.Contains(command, "\\") {
		return v.FileExists(command)
	}
	
	// Otherwise, look in PATH
	_, err := exec.LookPath(command)
	return err == nil
}

// DirectoryExists checks if a directory exists on the filesystem
func (v *SystemResourceValidator) DirectoryExists(path string) bool {
	if path == "" {
		return false
	}
	
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// checkLinuxService checks if a service exists on Linux systems
func (v *SystemResourceValidator) checkLinuxService(service string) bool {
	// Check systemd services
	if v.CommandExists("systemctl") {
		cmd := exec.Command("systemctl", "list-unit-files", "--type=service", "--no-pager", "--no-legend")
		output, err := cmd.Output()
		if err == nil {
			services := string(output)
			return strings.Contains(services, service+".service")
		}
	}
	
	// Check SysV init services
	initPath := "/etc/init.d/" + service
	if v.FileExists(initPath) {
		return true
	}
	
	// Check upstart services
	upstartPath := "/etc/init/" + service + ".conf"
	if v.FileExists(upstartPath) {
		return true
	}
	
	return false
}

// checkMacOSService checks if a service exists on macOS systems
func (v *SystemResourceValidator) checkMacOSService(service string) bool {
	// Check launchd services
	launchdPaths := []string{
		"/System/Library/LaunchDaemons/" + service + ".plist",
		"/Library/LaunchDaemons/" + service + ".plist",
		"/System/Library/LaunchAgents/" + service + ".plist",
		"/Library/LaunchAgents/" + service + ".plist",
	}
	
	for _, path := range launchdPaths {
		if v.FileExists(path) {
			return true
		}
	}
	
	// Check if it's a running service
	if v.CommandExists("launchctl") {
		cmd := exec.Command("launchctl", "list")
		output, err := cmd.Output()
		if err == nil {
			services := string(output)
			return strings.Contains(services, service)
		}
	}
	
	return false
}

// checkWindowsService checks if a service exists on Windows systems
func (v *SystemResourceValidator) checkWindowsService(service string) bool {
	// Use sc.exe to query service
	cmd := exec.Command("sc", "query", service)
	err := cmd.Run()
	return err == nil
}