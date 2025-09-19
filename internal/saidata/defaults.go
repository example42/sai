package saidata

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"sai/internal/types"
)

// DefaultsGenerator generates intelligent defaults for missing saidata
type DefaultsGenerator struct {
	validator ResourceValidator
}

// ResourceValidator validates resource existence
type ResourceValidator interface {
	ValidateFile(path string) bool
	ValidateService(serviceName string) bool
	ValidateCommand(command string) bool
	ValidateDirectory(path string) bool
	ValidatePort(port int) bool
}

// NewDefaultsGenerator creates a new defaults generator
func NewDefaultsGenerator(validator ResourceValidator) *DefaultsGenerator {
	return &DefaultsGenerator{
		validator: validator,
	}
}

// GenerateDefaults generates intelligent defaults for a software when saidata is missing
func (g *DefaultsGenerator) GenerateDefaults(software string) (*types.SoftwareData, error) {
	saidata := &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name:        software,
			DisplayName: strings.Title(software),
			Description: fmt.Sprintf("Generated defaults for %s", software),
		},
		IsGenerated: true,
	}

	// Generate platform-specific defaults
	switch runtime.GOOS {
	case "linux":
		g.generateLinuxDefaults(saidata, software)
	case "darwin":
		g.generateMacOSDefaults(saidata, software)
	case "windows":
		g.generateWindowsDefaults(saidata, software)
	}

	// Validate generated resources and filter out non-existent ones
	g.validateAndFilterResources(saidata)

	return saidata, nil
}

// generateLinuxDefaults generates Linux-specific defaults
func (g *DefaultsGenerator) generateLinuxDefaults(saidata *types.SoftwareData, software string) {
	// Generate package defaults
	saidata.Packages = g.GeneratePackageDefaults(software)
	
	// Generate service defaults
	saidata.Services = g.GenerateServiceDefaults(software)
	
	// Generate file defaults
	saidata.Files = g.GenerateFileDefaults(software)
	
	// Generate directory defaults
	saidata.Directories = g.GenerateDirectoryDefaults(software)
	
	// Generate command defaults
	saidata.Commands = g.GenerateCommandDefaults(software)
	
	// Generate port defaults
	saidata.Ports = g.GeneratePortDefaults(software)
}

// generateMacOSDefaults generates macOS-specific defaults
func (g *DefaultsGenerator) generateMacOSDefaults(saidata *types.SoftwareData, software string) {
	// Package defaults for macOS (Homebrew style)
	saidata.Packages = []types.Package{
		{
			Name: software,
		},
	}
	
	// Service defaults for macOS (launchd)
	saidata.Services = []types.Service{
		{
			Name:        software,
			ServiceName: software,
			Type:        "launchd",
		},
	}
	
	// File defaults for macOS
	saidata.Files = []types.File{
		{
			Name: "config",
			Path: fmt.Sprintf("/opt/homebrew/etc/%s/%s.conf", software, software),
			Type: "config",
		},
		{
			Name: "log",
			Path: fmt.Sprintf("/opt/homebrew/var/log/%s.log", software),
			Type: "log",
		},
	}
	
	// Directory defaults for macOS
	saidata.Directories = []types.Directory{
		{
			Name: "config",
			Path: fmt.Sprintf("/opt/homebrew/etc/%s", software),
		},
		{
			Name: "data",
			Path: fmt.Sprintf("/opt/homebrew/var/lib/%s", software),
		},
	}
	
	// Command defaults for macOS
	saidata.Commands = []types.Command{
		{
			Name: software,
			Path: fmt.Sprintf("/opt/homebrew/bin/%s", software),
		},
	}
	
	// Port defaults
	saidata.Ports = g.GeneratePortDefaults(software)
}

// generateWindowsDefaults generates Windows-specific defaults
func (g *DefaultsGenerator) generateWindowsDefaults(saidata *types.SoftwareData, software string) {
	// Package defaults for Windows
	saidata.Packages = []types.Package{
		{
			Name: software,
		},
	}
	
	// Service defaults for Windows
	saidata.Services = []types.Service{
		{
			Name:        software,
			ServiceName: software,
			Type:        "windows_service",
		},
	}
	
	// File defaults for Windows
	saidata.Files = []types.File{
		{
			Name: "config",
			Path: fmt.Sprintf("C:\\Program Files\\%s\\%s.conf", strings.Title(software), software),
			Type: "config",
		},
		{
			Name: "log",
			Path: fmt.Sprintf("C:\\ProgramData\\%s\\%s.log", strings.Title(software), software),
			Type: "log",
		},
	}
	
	// Directory defaults for Windows
	saidata.Directories = []types.Directory{
		{
			Name: "config",
			Path: fmt.Sprintf("C:\\Program Files\\%s", strings.Title(software)),
		},
		{
			Name: "data",
			Path: fmt.Sprintf("C:\\ProgramData\\%s", strings.Title(software)),
		},
	}
	
	// Command defaults for Windows
	saidata.Commands = []types.Command{
		{
			Name: software,
			Path: fmt.Sprintf("C:\\Program Files\\%s\\%s.exe", strings.Title(software), software),
		},
	}
	
	// Port defaults
	saidata.Ports = g.GeneratePortDefaults(software)
}

// GeneratePackageDefaults generates default package definitions
func (g *DefaultsGenerator) GeneratePackageDefaults(software string) []types.Package {
	return []types.Package{
		{
			Name: software,
		},
	}
}

// GenerateServiceDefaults generates default service definitions
func (g *DefaultsGenerator) GenerateServiceDefaults(software string) []types.Service {
	serviceType := "systemd"
	if runtime.GOOS == "darwin" {
		serviceType = "launchd"
	} else if runtime.GOOS == "windows" {
		serviceType = "windows_service"
	}
	
	return []types.Service{
		{
			Name:        software,
			ServiceName: software,
			Type:        serviceType,
		},
	}
}

// GenerateFileDefaults generates default file definitions
func (g *DefaultsGenerator) GenerateFileDefaults(software string) []types.File {
	var files []types.File
	
	switch runtime.GOOS {
	case "linux":
		files = []types.File{
			{
				Name: "config",
				Path: fmt.Sprintf("/etc/%s/%s.conf", software, software),
				Type: "config",
			},
			{
				Name: "alt_config",
				Path: fmt.Sprintf("/etc/%s.conf", software),
				Type: "config",
			},
			{
				Name: "binary",
				Path: fmt.Sprintf("/usr/bin/%s", software),
				Type: "binary",
			},
			{
				Name: "alt_binary",
				Path: fmt.Sprintf("/usr/sbin/%s", software),
				Type: "binary",
			},
			{
				Name: "log",
				Path: fmt.Sprintf("/var/log/%s.log", software),
				Type: "log",
			},
			{
				Name: "alt_log",
				Path: fmt.Sprintf("/var/log/%s/%s.log", software, software),
				Type: "log",
			},
		}
	case "darwin":
		files = []types.File{
			{
				Name: "config",
				Path: fmt.Sprintf("/opt/homebrew/etc/%s/%s.conf", software, software),
				Type: "config",
			},
			{
				Name: "binary",
				Path: fmt.Sprintf("/opt/homebrew/bin/%s", software),
				Type: "binary",
			},
			{
				Name: "log",
				Path: fmt.Sprintf("/opt/homebrew/var/log/%s.log", software),
				Type: "log",
			},
		}
	case "windows":
		files = []types.File{
			{
				Name: "config",
				Path: fmt.Sprintf("C:\\Program Files\\%s\\%s.conf", strings.Title(software), software),
				Type: "config",
			},
			{
				Name: "binary",
				Path: fmt.Sprintf("C:\\Program Files\\%s\\%s.exe", strings.Title(software), software),
				Type: "binary",
			},
			{
				Name: "log",
				Path: fmt.Sprintf("C:\\ProgramData\\%s\\%s.log", strings.Title(software), software),
				Type: "log",
			},
		}
	}
	
	return files
}

// GenerateDirectoryDefaults generates default directory definitions
func (g *DefaultsGenerator) GenerateDirectoryDefaults(software string) []types.Directory {
	var directories []types.Directory
	
	switch runtime.GOOS {
	case "linux":
		directories = []types.Directory{
			{
				Name: "config",
				Path: fmt.Sprintf("/etc/%s", software),
			},
			{
				Name: "data",
				Path: fmt.Sprintf("/var/lib/%s", software),
			},
			{
				Name: "log",
				Path: fmt.Sprintf("/var/log/%s", software),
			},
			{
				Name: "cache",
				Path: fmt.Sprintf("/var/cache/%s", software),
			},
			{
				Name: "run",
				Path: fmt.Sprintf("/var/run/%s", software),
			},
		}
	case "darwin":
		directories = []types.Directory{
			{
				Name: "config",
				Path: fmt.Sprintf("/opt/homebrew/etc/%s", software),
			},
			{
				Name: "data",
				Path: fmt.Sprintf("/opt/homebrew/var/lib/%s", software),
			},
			{
				Name: "log",
				Path: fmt.Sprintf("/opt/homebrew/var/log/%s", software),
			},
		}
	case "windows":
		directories = []types.Directory{
			{
				Name: "config",
				Path: fmt.Sprintf("C:\\Program Files\\%s", strings.Title(software)),
			},
			{
				Name: "data",
				Path: fmt.Sprintf("C:\\ProgramData\\%s", strings.Title(software)),
			},
		}
	}
	
	return directories
}

// GenerateCommandDefaults generates default command definitions
func (g *DefaultsGenerator) GenerateCommandDefaults(software string) []types.Command {
	var commands []types.Command
	
	switch runtime.GOOS {
	case "linux":
		commands = []types.Command{
			{
				Name: software,
				Path: fmt.Sprintf("/usr/bin/%s", software),
			},
			{
				Name: fmt.Sprintf("%s-alt", software),
				Path: fmt.Sprintf("/usr/sbin/%s", software),
			},
			{
				Name: fmt.Sprintf("%s-local", software),
				Path: fmt.Sprintf("/usr/local/bin/%s", software),
			},
		}
	case "darwin":
		commands = []types.Command{
			{
				Name: software,
				Path: fmt.Sprintf("/opt/homebrew/bin/%s", software),
			},
			{
				Name: fmt.Sprintf("%s-system", software),
				Path: fmt.Sprintf("/usr/bin/%s", software),
			},
		}
	case "windows":
		commands = []types.Command{
			{
				Name: software,
				Path: fmt.Sprintf("C:\\Program Files\\%s\\%s.exe", strings.Title(software), software),
			},
		}
	}
	
	return commands
}

// GeneratePortDefaults generates default port definitions based on well-known ports
func (g *DefaultsGenerator) GeneratePortDefaults(software string) []types.Port {
	// Well-known port mappings
	wellKnownPorts := map[string][]types.Port{
		"apache":        {{Port: 80, Protocol: "tcp", Service: "http"}, {Port: 443, Protocol: "tcp", Service: "https"}},
		"nginx":         {{Port: 80, Protocol: "tcp", Service: "http"}, {Port: 443, Protocol: "tcp", Service: "https"}},
		"mysql":         {{Port: 3306, Protocol: "tcp", Service: "mysql"}},
		"postgresql":    {{Port: 5432, Protocol: "tcp", Service: "postgresql"}},
		"redis":         {{Port: 6379, Protocol: "tcp", Service: "redis"}},
		"mongodb":       {{Port: 27017, Protocol: "tcp", Service: "mongodb"}},
		"elasticsearch": {{Port: 9200, Protocol: "tcp", Service: "elasticsearch"}, {Port: 9300, Protocol: "tcp", Service: "elasticsearch-cluster"}},
		"jenkins":       {{Port: 8080, Protocol: "tcp", Service: "jenkins"}},
		"grafana":       {{Port: 3000, Protocol: "tcp", Service: "grafana"}},
		"prometheus":    {{Port: 9090, Protocol: "tcp", Service: "prometheus"}},
		"docker":        {{Port: 2376, Protocol: "tcp", Service: "docker"}},
		"kubernetes":    {{Port: 6443, Protocol: "tcp", Service: "kubernetes-api"}},
		"ssh":           {{Port: 22, Protocol: "tcp", Service: "ssh"}},
		"ftp":           {{Port: 21, Protocol: "tcp", Service: "ftp"}},
		"smtp":          {{Port: 25, Protocol: "tcp", Service: "smtp"}},
		"dns":           {{Port: 53, Protocol: "udp", Service: "dns"}},
		"dhcp":          {{Port: 67, Protocol: "udp", Service: "dhcp"}},
		"ntp":           {{Port: 123, Protocol: "udp", Service: "ntp"}},
		"snmp":          {{Port: 161, Protocol: "udp", Service: "snmp"}},
	}
	
	// Check for exact match first
	if ports, exists := wellKnownPorts[strings.ToLower(software)]; exists {
		return ports
	}
	
	// Check for partial matches
	softwareLower := strings.ToLower(software)
	for name, ports := range wellKnownPorts {
		if strings.Contains(softwareLower, name) || strings.Contains(name, softwareLower) {
			return ports
		}
	}
	
	// Return empty slice if no well-known ports found
	return []types.Port{}
}

// validateAndFilterResources validates generated resources and removes non-existent ones
func (g *DefaultsGenerator) validateAndFilterResources(saidata *types.SoftwareData) {
	if g.validator == nil {
		return // Skip validation if no validator provided
	}
	
	// Filter files
	var validFiles []types.File
	for _, file := range saidata.Files {
		if g.validator.ValidateFile(file.Path) {
			file.Exists = true
			validFiles = append(validFiles, file)
		}
	}
	saidata.Files = validFiles
	
	// Filter services
	var validServices []types.Service
	for _, service := range saidata.Services {
		if g.validator.ValidateService(service.GetServiceNameOrDefault()) {
			service.Exists = true
			validServices = append(validServices, service)
		}
	}
	saidata.Services = validServices
	
	// Filter commands
	var validCommands []types.Command
	for _, command := range saidata.Commands {
		if g.validator.ValidateCommand(command.GetPathOrDefault()) {
			command.Exists = true
			validCommands = append(validCommands, command)
		}
	}
	saidata.Commands = validCommands
	
	// Filter directories
	var validDirectories []types.Directory
	for _, directory := range saidata.Directories {
		if g.validator.ValidateDirectory(directory.Path) {
			directory.Exists = true
			validDirectories = append(validDirectories, directory)
		}
	}
	saidata.Directories = validDirectories
	
	// Filter ports (check if they're open/in use)
	var validPorts []types.Port
	for _, port := range saidata.Ports {
		if g.validator.ValidatePort(port.Port) {
			port.IsOpen = true
			validPorts = append(validPorts, port)
		}
	}
	saidata.Ports = validPorts
}

// ValidatePathExists checks if a file or directory path exists
func (g *DefaultsGenerator) ValidatePathExists(path string) bool {
	if g.validator != nil {
		return g.validator.ValidateFile(path) || g.validator.ValidateDirectory(path)
	}
	
	// Fallback to basic file system check
	_, err := os.Stat(path)
	return err == nil
}

// ValidateServiceExists checks if a service exists on the system
func (g *DefaultsGenerator) ValidateServiceExists(service string) bool {
	if g.validator != nil {
		return g.validator.ValidateService(service)
	}
	
	// Fallback implementation would check systemctl, launchctl, etc.
	return false
}

// ValidateCommandExists checks if a command exists and is executable
func (g *DefaultsGenerator) ValidateCommandExists(command string) bool {
	if g.validator != nil {
		return g.validator.ValidateCommand(command)
	}
	
	// Fallback to basic executable check
	if filepath.IsAbs(command) {
		info, err := os.Stat(command)
		if err != nil {
			return false
		}
		return info.Mode()&0111 != 0 // Check if executable
	}
	
	// Check PATH for relative commands
	_, err := exec.LookPath(command)
	return err == nil
}

// GetDefaultConfigPath generates a default configuration file path
func GetDefaultConfigPath(software string) string {
	switch runtime.GOOS {
	case "linux":
		return fmt.Sprintf("/etc/%s/%s.conf", software, software)
	case "darwin":
		return fmt.Sprintf("/opt/homebrew/etc/%s/%s.conf", software, software)
	case "windows":
		return fmt.Sprintf("C:\\Program Files\\%s\\%s.conf", strings.Title(software), software)
	default:
		return fmt.Sprintf("./%s.conf", software)
	}
}

// GetDefaultLogPath generates a default log file path
func GetDefaultLogPath(software string) string {
	switch runtime.GOOS {
	case "linux":
		return fmt.Sprintf("/var/log/%s.log", software)
	case "darwin":
		return fmt.Sprintf("/opt/homebrew/var/log/%s.log", software)
	case "windows":
		return fmt.Sprintf("C:\\ProgramData\\%s\\%s.log", strings.Title(software), software)
	default:
		return fmt.Sprintf("./%s.log", software)
	}
}

// GetDefaultDataDir generates a default data directory path
func GetDefaultDataDir(software string) string {
	switch runtime.GOOS {
	case "linux":
		return fmt.Sprintf("/var/lib/%s", software)
	case "darwin":
		return fmt.Sprintf("/opt/homebrew/var/lib/%s", software)
	case "windows":
		return fmt.Sprintf("C:\\ProgramData\\%s", strings.Title(software))
	default:
		return fmt.Sprintf("./%s-data", software)
	}
}

// GetDefaultServiceName generates a default service name
func GetDefaultServiceName(software string) string {
	return software
}

// GetDefaultCommandPath generates a default command path
func GetDefaultCommandPath(software string) string {
	switch runtime.GOOS {
	case "linux":
		return fmt.Sprintf("/usr/bin/%s", software)
	case "darwin":
		return fmt.Sprintf("/opt/homebrew/bin/%s", software)
	case "windows":
		return fmt.Sprintf("C:\\Program Files\\%s\\%s.exe", strings.Title(software), software)
	default:
		return software
	}
}