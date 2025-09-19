package template

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// SystemDefaultsGenerator provides platform-specific default path generation
type SystemDefaultsGenerator struct {
	validator ResourceValidator
}

// NewSystemDefaultsGenerator creates a new system defaults generator
func NewSystemDefaultsGenerator(validator ResourceValidator) *SystemDefaultsGenerator {
	return &SystemDefaultsGenerator{
		validator: validator,
	}
}

// DefaultConfigPath generates a default configuration file path for the software
func (g *SystemDefaultsGenerator) DefaultConfigPath(software string) string {
	switch runtime.GOOS {
	case "linux":
		return g.linuxConfigPath(software)
	case "darwin":
		return g.macOSConfigPath(software)
	case "windows":
		return g.windowsConfigPath(software)
	default:
		return filepath.Join("/etc", software, software+".conf")
	}
}

// DefaultLogPath generates a default log file path for the software
func (g *SystemDefaultsGenerator) DefaultLogPath(software string) string {
	switch runtime.GOOS {
	case "linux":
		return g.linuxLogPath(software)
	case "darwin":
		return g.macOSLogPath(software)
	case "windows":
		return g.windowsLogPath(software)
	default:
		return filepath.Join("/var/log", software+".log")
	}
}

// DefaultDataDir generates a default data directory path for the software
func (g *SystemDefaultsGenerator) DefaultDataDir(software string) string {
	switch runtime.GOOS {
	case "linux":
		return g.linuxDataDir(software)
	case "darwin":
		return g.macOSDataDir(software)
	case "windows":
		return g.windowsDataDir(software)
	default:
		return filepath.Join("/var/lib", software)
	}
}

// DefaultServiceName generates a default service name for the software
func (g *SystemDefaultsGenerator) DefaultServiceName(software string) string {
	// Service names are generally consistent across platforms
	return software
}

// DefaultCommandPath generates a default command path for the software
func (g *SystemDefaultsGenerator) DefaultCommandPath(software string) string {
	switch runtime.GOOS {
	case "linux":
		return g.linuxCommandPath(software)
	case "darwin":
		return g.macOSCommandPath(software)
	case "windows":
		return g.windowsCommandPath(software)
	default:
		return filepath.Join("/usr/bin", software)
	}
}

// Linux-specific default paths
func (g *SystemDefaultsGenerator) linuxConfigPath(software string) string {
	// Try common Linux config paths in order of preference
	candidates := []string{
		filepath.Join("/etc", software, software+".conf"),
		filepath.Join("/etc", software+".conf"),
		filepath.Join("/etc", software, "config"),
		filepath.Join("/etc", software, software+".yaml"),
		filepath.Join("/etc", software, software+".yml"),
	}
	
	return g.findExistingPath(candidates, filepath.Join("/etc", software, software+".conf"))
}

func (g *SystemDefaultsGenerator) linuxLogPath(software string) string {
	candidates := []string{
		filepath.Join("/var/log", software, software+".log"),
		filepath.Join("/var/log", software+".log"),
		filepath.Join("/var/log", software, "access.log"),
		filepath.Join("/var/log", software, "error.log"),
	}
	
	return g.findExistingPath(candidates, filepath.Join("/var/log", software+".log"))
}

func (g *SystemDefaultsGenerator) linuxDataDir(software string) string {
	candidates := []string{
		filepath.Join("/var/lib", software),
		filepath.Join("/opt", software),
		filepath.Join("/usr/share", software),
		filepath.Join("/var", software),
	}
	
	return g.findExistingPath(candidates, filepath.Join("/var/lib", software))
}

func (g *SystemDefaultsGenerator) linuxCommandPath(software string) string {
	candidates := []string{
		filepath.Join("/usr/bin", software),
		filepath.Join("/usr/local/bin", software),
		filepath.Join("/bin", software),
		filepath.Join("/sbin", software),
		filepath.Join("/usr/sbin", software),
	}
	
	return g.findExistingPath(candidates, filepath.Join("/usr/bin", software))
}

// macOS-specific default paths
func (g *SystemDefaultsGenerator) macOSConfigPath(software string) string {
	candidates := []string{
		filepath.Join("/usr/local/etc", software, software+".conf"),
		filepath.Join("/usr/local/etc", software+".conf"),
		filepath.Join("/etc", software, software+".conf"),
		filepath.Join("/etc", software+".conf"),
		filepath.Join("/opt/homebrew/etc", software, software+".conf"),
		filepath.Join("/opt/homebrew/etc", software+".conf"),
	}
	
	return g.findExistingPath(candidates, filepath.Join("/usr/local/etc", software, software+".conf"))
}

func (g *SystemDefaultsGenerator) macOSLogPath(software string) string {
	candidates := []string{
		filepath.Join("/usr/local/var/log", software, software+".log"),
		filepath.Join("/usr/local/var/log", software+".log"),
		filepath.Join("/var/log", software+".log"),
		filepath.Join("/opt/homebrew/var/log", software+".log"),
	}
	
	return g.findExistingPath(candidates, filepath.Join("/usr/local/var/log", software+".log"))
}

func (g *SystemDefaultsGenerator) macOSDataDir(software string) string {
	candidates := []string{
		filepath.Join("/usr/local/var", software),
		filepath.Join("/usr/local/share", software),
		filepath.Join("/opt/homebrew/var", software),
		filepath.Join("/var/lib", software),
	}
	
	return g.findExistingPath(candidates, filepath.Join("/usr/local/var", software))
}

func (g *SystemDefaultsGenerator) macOSCommandPath(software string) string {
	candidates := []string{
		filepath.Join("/usr/local/bin", software),
		filepath.Join("/opt/homebrew/bin", software),
		filepath.Join("/usr/bin", software),
		filepath.Join("/bin", software),
	}
	
	return g.findExistingPath(candidates, filepath.Join("/usr/local/bin", software))
}

// Windows-specific default paths
func (g *SystemDefaultsGenerator) windowsConfigPath(software string) string {
	programData := g.getWindowsProgramData()
	candidates := []string{
		filepath.Join(programData, strings.Title(software), software+".conf"),
		filepath.Join(programData, strings.Title(software), "config", software+".conf"),
		filepath.Join(programData, strings.Title(software), software+".ini"),
		filepath.Join(programData, strings.Title(software), "config.ini"),
	}
	
	return g.findExistingPath(candidates, filepath.Join(programData, strings.Title(software), software+".conf"))
}

func (g *SystemDefaultsGenerator) windowsLogPath(software string) string {
	programData := g.getWindowsProgramData()
	candidates := []string{
		filepath.Join(programData, strings.Title(software), "logs", software+".log"),
		filepath.Join(programData, strings.Title(software), software+".log"),
		filepath.Join("C:", "logs", software+".log"),
	}
	
	return g.findExistingPath(candidates, filepath.Join(programData, strings.Title(software), "logs", software+".log"))
}

func (g *SystemDefaultsGenerator) windowsDataDir(software string) string {
	programData := g.getWindowsProgramData()
	candidates := []string{
		filepath.Join(programData, strings.Title(software)),
		filepath.Join(programData, strings.Title(software), "data"),
		filepath.Join("C:", strings.Title(software)),
	}
	
	return g.findExistingPath(candidates, filepath.Join(programData, strings.Title(software)))
}

func (g *SystemDefaultsGenerator) windowsCommandPath(software string) string {
	candidates := []string{
		filepath.Join("C:", "Program Files", strings.Title(software), software+".exe"),
		filepath.Join("C:", "Program Files (x86)", strings.Title(software), software+".exe"),
		software + ".exe", // Assume it's in PATH
	}
	
	return g.findExistingPath(candidates, software+".exe")
}

// Helper functions
func (g *SystemDefaultsGenerator) findExistingPath(candidates []string, defaultPath string) string {
	if g.validator == nil {
		return defaultPath
	}
	
	for _, candidate := range candidates {
		if g.validator.FileExists(candidate) || g.validator.DirectoryExists(candidate) {
			return candidate
		}
	}
	
	return defaultPath
}

func (g *SystemDefaultsGenerator) getWindowsProgramData() string {
	programData := os.Getenv("PROGRAMDATA")
	if programData == "" {
		programData = "C:\\ProgramData"
	}
	return programData
}