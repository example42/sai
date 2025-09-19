package template

import (
	"runtime"
	"strings"
	"testing"
)

func TestSystemDefaultsGenerator_DefaultConfigPath(t *testing.T) {
	validator := NewMockResourceValidator()
	generator := NewSystemDefaultsGenerator(validator)
	
	software := "nginx"
	result := generator.DefaultConfigPath(software)
	
	// Should return a non-empty path
	if result == "" {
		t.Error("Expected non-empty config path")
	}
	
	// Should contain the software name
	if !strings.Contains(result, software) {
		t.Errorf("Expected config path to contain software name %s, got %s", software, result)
	}
	
	// Platform-specific checks
	switch runtime.GOOS {
	case "linux":
		if !strings.HasPrefix(result, "/etc/") {
			t.Errorf("Expected Linux config path to start with /etc/, got %s", result)
		}
	case "darwin":
		if !strings.Contains(result, "/etc/") && !strings.Contains(result, "/usr/local/etc/") && !strings.Contains(result, "/opt/homebrew/etc/") {
			t.Errorf("Expected macOS config path to contain etc directory, got %s", result)
		}
	case "windows":
		// Windows paths should contain the software name in title case
		if !strings.Contains(result, "Nginx") {
			t.Errorf("Expected Windows config path to contain title case software name, got %s", result)
		}
	}
}

func TestSystemDefaultsGenerator_DefaultLogPath(t *testing.T) {
	validator := NewMockResourceValidator()
	generator := NewSystemDefaultsGenerator(validator)
	
	software := "nginx"
	result := generator.DefaultLogPath(software)
	
	// Should return a non-empty path
	if result == "" {
		t.Error("Expected non-empty log path")
	}
	
	// Should contain the software name
	if !strings.Contains(result, software) {
		t.Errorf("Expected log path to contain software name %s, got %s", software, result)
	}
	
	// Should end with .log
	if !strings.HasSuffix(result, ".log") {
		t.Errorf("Expected log path to end with .log, got %s", result)
	}
	
	// Platform-specific checks
	switch runtime.GOOS {
	case "linux":
		if !strings.HasPrefix(result, "/var/log/") {
			t.Errorf("Expected Linux log path to start with /var/log/, got %s", result)
		}
	case "darwin":
		if !strings.Contains(result, "/var/log/") && !strings.Contains(result, "/usr/local/var/log/") && !strings.Contains(result, "/opt/homebrew/var/log/") {
			t.Errorf("Expected macOS log path to contain log directory, got %s", result)
		}
	case "windows":
		// Windows paths should contain logs directory
		if !strings.Contains(result, "logs") && !strings.Contains(result, "Logs") {
			t.Errorf("Expected Windows log path to contain logs directory, got %s", result)
		}
	}
}

func TestSystemDefaultsGenerator_DefaultDataDir(t *testing.T) {
	validator := NewMockResourceValidator()
	generator := NewSystemDefaultsGenerator(validator)
	
	software := "nginx"
	result := generator.DefaultDataDir(software)
	
	// Should return a non-empty path
	if result == "" {
		t.Error("Expected non-empty data directory path")
	}
	
	// Should contain the software name
	if !strings.Contains(result, software) {
		t.Errorf("Expected data dir to contain software name %s, got %s", software, result)
	}
	
	// Platform-specific checks
	switch runtime.GOOS {
	case "linux":
		if !strings.HasPrefix(result, "/var/lib/") && !strings.HasPrefix(result, "/opt/") && !strings.HasPrefix(result, "/usr/share/") && !strings.HasPrefix(result, "/var/") {
			t.Errorf("Expected Linux data dir to start with common data directory, got %s", result)
		}
	case "darwin":
		if !strings.Contains(result, "/usr/local/var/") && !strings.Contains(result, "/usr/local/share/") && !strings.Contains(result, "/opt/homebrew/var/") && !strings.Contains(result, "/var/lib/") {
			t.Errorf("Expected macOS data dir to contain common data directory, got %s", result)
		}
	case "windows":
		// Windows paths should contain the software name in title case
		if !strings.Contains(result, "Nginx") {
			t.Errorf("Expected Windows data dir to contain title case software name, got %s", result)
		}
	}
}

func TestSystemDefaultsGenerator_DefaultServiceName(t *testing.T) {
	validator := NewMockResourceValidator()
	generator := NewSystemDefaultsGenerator(validator)
	
	software := "nginx"
	result := generator.DefaultServiceName(software)
	
	// Should return the software name as-is
	if result != software {
		t.Errorf("Expected service name to be %s, got %s", software, result)
	}
}

func TestSystemDefaultsGenerator_DefaultCommandPath(t *testing.T) {
	validator := NewMockResourceValidator()
	generator := NewSystemDefaultsGenerator(validator)
	
	software := "nginx"
	result := generator.DefaultCommandPath(software)
	
	// Should return a non-empty path
	if result == "" {
		t.Error("Expected non-empty command path")
	}
	
	// Should contain the software name
	if !strings.Contains(result, software) {
		t.Errorf("Expected command path to contain software name %s, got %s", software, result)
	}
	
	// Platform-specific checks
	switch runtime.GOOS {
	case "linux":
		if !strings.HasPrefix(result, "/usr/bin/") && !strings.HasPrefix(result, "/usr/local/bin/") && !strings.HasPrefix(result, "/bin/") && !strings.HasPrefix(result, "/sbin/") && !strings.HasPrefix(result, "/usr/sbin/") {
			t.Errorf("Expected Linux command path to start with common bin directory, got %s", result)
		}
	case "darwin":
		if !strings.Contains(result, "/usr/local/bin/") && !strings.Contains(result, "/opt/homebrew/bin/") && !strings.Contains(result, "/usr/bin/") && !strings.Contains(result, "/bin/") {
			t.Errorf("Expected macOS command path to contain common bin directory, got %s", result)
		}
	case "windows":
		// Windows executables should end with .exe
		if !strings.HasSuffix(result, ".exe") {
			t.Errorf("Expected Windows command path to end with .exe, got %s", result)
		}
	}
}

func TestSystemDefaultsGenerator_FindExistingPath(t *testing.T) {
	validator := NewMockResourceValidator()
	generator := NewSystemDefaultsGenerator(validator)
	
	// Set up mock validator with some existing paths
	validator.SetFileExists("/etc/nginx/nginx.conf", true)
	validator.SetDirectoryExists("/var/lib/nginx", true)
	
	tests := []struct {
		name        string
		candidates  []string
		defaultPath string
		expected    string
	}{
		{
			name: "first candidate exists",
			candidates: []string{
				"/etc/nginx/nginx.conf",
				"/etc/nginx.conf",
			},
			defaultPath: "/etc/nginx.conf",
			expected:    "/etc/nginx/nginx.conf",
		},
		{
			name: "second candidate exists",
			candidates: []string{
				"/etc/nonexistent.conf",
				"/var/lib/nginx",
			},
			defaultPath: "/etc/default.conf",
			expected:    "/var/lib/nginx",
		},
		{
			name: "no candidates exist",
			candidates: []string{
				"/nonexistent1",
				"/nonexistent2",
			},
			defaultPath: "/default/path",
			expected:    "/default/path",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.findExistingPath(tt.candidates, tt.defaultPath)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSystemDefaultsGenerator_LinuxPaths(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test on non-Linux platform")
	}
	
	validator := NewMockResourceValidator()
	generator := NewSystemDefaultsGenerator(validator)
	
	software := "nginx"
	
	// Test Linux-specific path generation
	configPath := generator.linuxConfigPath(software)
	if !strings.HasPrefix(configPath, "/etc/") {
		t.Errorf("Expected Linux config path to start with /etc/, got %s", configPath)
	}
	
	logPath := generator.linuxLogPath(software)
	if !strings.HasPrefix(logPath, "/var/log/") {
		t.Errorf("Expected Linux log path to start with /var/log/, got %s", logPath)
	}
	
	dataDir := generator.linuxDataDir(software)
	if !strings.HasPrefix(dataDir, "/var/lib/") && !strings.HasPrefix(dataDir, "/opt/") && !strings.HasPrefix(dataDir, "/usr/share/") && !strings.HasPrefix(dataDir, "/var/") {
		t.Errorf("Expected Linux data dir to start with common data directory, got %s", dataDir)
	}
	
	commandPath := generator.linuxCommandPath(software)
	if !strings.HasPrefix(commandPath, "/usr/bin/") && !strings.HasPrefix(commandPath, "/usr/local/bin/") && !strings.HasPrefix(commandPath, "/bin/") && !strings.HasPrefix(commandPath, "/sbin/") && !strings.HasPrefix(commandPath, "/usr/sbin/") {
		t.Errorf("Expected Linux command path to start with common bin directory, got %s", commandPath)
	}
}

func TestSystemDefaultsGenerator_MacOSPaths(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-macOS platform")
	}
	
	validator := NewMockResourceValidator()
	generator := NewSystemDefaultsGenerator(validator)
	
	software := "nginx"
	
	// Test macOS-specific path generation
	configPath := generator.macOSConfigPath(software)
	if !strings.Contains(configPath, "/etc/") && !strings.Contains(configPath, "/usr/local/etc/") && !strings.Contains(configPath, "/opt/homebrew/etc/") {
		t.Errorf("Expected macOS config path to contain etc directory, got %s", configPath)
	}
	
	logPath := generator.macOSLogPath(software)
	if !strings.Contains(logPath, "/var/log/") && !strings.Contains(logPath, "/usr/local/var/log/") && !strings.Contains(logPath, "/opt/homebrew/var/log/") {
		t.Errorf("Expected macOS log path to contain log directory, got %s", logPath)
	}
	
	dataDir := generator.macOSDataDir(software)
	if !strings.Contains(dataDir, "/usr/local/var/") && !strings.Contains(dataDir, "/usr/local/share/") && !strings.Contains(dataDir, "/opt/homebrew/var/") && !strings.Contains(dataDir, "/var/lib/") {
		t.Errorf("Expected macOS data dir to contain common data directory, got %s", dataDir)
	}
	
	commandPath := generator.macOSCommandPath(software)
	if !strings.Contains(commandPath, "/usr/local/bin/") && !strings.Contains(commandPath, "/opt/homebrew/bin/") && !strings.Contains(commandPath, "/usr/bin/") && !strings.Contains(commandPath, "/bin/") {
		t.Errorf("Expected macOS command path to contain common bin directory, got %s", commandPath)
	}
}

func TestSystemDefaultsGenerator_WindowsPaths(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}
	
	validator := NewMockResourceValidator()
	generator := NewSystemDefaultsGenerator(validator)
	
	software := "nginx"
	
	// Test Windows-specific path generation
	configPath := generator.windowsConfigPath(software)
	if !strings.Contains(configPath, "Nginx") {
		t.Errorf("Expected Windows config path to contain title case software name, got %s", configPath)
	}
	
	logPath := generator.windowsLogPath(software)
	if !strings.Contains(logPath, "logs") && !strings.Contains(logPath, "Logs") {
		t.Errorf("Expected Windows log path to contain logs directory, got %s", logPath)
	}
	
	dataDir := generator.windowsDataDir(software)
	if !strings.Contains(dataDir, "Nginx") {
		t.Errorf("Expected Windows data dir to contain title case software name, got %s", dataDir)
	}
	
	commandPath := generator.windowsCommandPath(software)
	if !strings.HasSuffix(commandPath, ".exe") {
		t.Errorf("Expected Windows command path to end with .exe, got %s", commandPath)
	}
}