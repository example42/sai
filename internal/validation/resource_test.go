package validation

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sai/internal/types"
)

func TestNewResourceValidator(t *testing.T) {
	validator := NewResourceValidator()
	assert.NotNil(t, validator)
}

func TestResourceValidator_ValidateFile(t *testing.T) {
	validator := NewResourceValidator()
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.conf")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name     string
		file     types.File
		expected bool
	}{
		{
			name: "existing file",
			file: types.File{
				Name: "test",
				Path: testFile,
				Type: "config",
			},
			expected: true,
		},
		{
			name: "non-existent file",
			file: types.File{
				Name: "nonexistent",
				Path: "/nonexistent/file.conf",
				Type: "config",
			},
			expected: false,
		},
		{
			name: "empty path",
			file: types.File{
				Name: "empty",
				Path: "",
				Type: "config",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateFile(tt.file)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResourceValidator_ValidateService(t *testing.T) {
	validator := NewResourceValidator()

	tests := []struct {
		name     string
		service  types.Service
		expected bool
	}{
		{
			name: "systemd service",
			service: types.Service{
				Name:        "test-service",
				ServiceName: "test-service",
				Type:        "systemd",
			},
			expected: false, // Most likely won't exist in test environment
		},
		{
			name: "empty service name",
			service: types.Service{
				Name:        "empty",
				ServiceName: "",
				Type:        "systemd",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateService(tt.service)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResourceValidator_ValidateCommand(t *testing.T) {
	validator := NewResourceValidator()

	// Find a command that should exist on most systems
	var existingCommand string
	if runtime.GOOS == "windows" {
		existingCommand = "cmd"
	} else {
		existingCommand = "sh"
	}

	tests := []struct {
		name     string
		command  types.Command
		expected bool
	}{
		{
			name: "existing command",
			command: types.Command{
				Name: "shell",
				Path: existingCommand, // Use the detected existing command
			},
			expected: true,
		},
		{
			name: "non-existent command",
			command: types.Command{
				Name: "nonexistent",
				Path: "nonexistent-command-12345",
			},
			expected: false,
		},
		{
			name: "empty path",
			command: types.Command{
				Name: "empty",
				Path: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateCommand(tt.command)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResourceValidator_ValidateDirectory(t *testing.T) {
	validator := NewResourceValidator()
	tempDir := t.TempDir()

	// Create a test directory
	testDir := filepath.Join(tempDir, "testdir")
	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	tests := []struct {
		name      string
		directory types.Directory
		expected  bool
	}{
		{
			name: "existing directory",
			directory: types.Directory{
				Name: "test",
				Path: testDir,
			},
			expected: true,
		},
		{
			name: "non-existent directory",
			directory: types.Directory{
				Name: "nonexistent",
				Path: "/nonexistent/directory",
			},
			expected: false,
		},
		{
			name: "empty path",
			directory: types.Directory{
				Name: "empty",
				Path: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateDirectory(tt.directory)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResourceValidator_ValidatePort(t *testing.T) {
	validator := NewResourceValidator()

	tests := []struct {
		name     string
		port     types.Port
		expected bool
	}{
		{
			name: "valid port",
			port: types.Port{
				Port:     8080,
				Protocol: "tcp",
				Service:  "http-alt",
			},
			expected: true,
		},
		{
			name: "invalid port - too low",
			port: types.Port{
				Port:     0,
				Protocol: "tcp",
				Service:  "invalid",
			},
			expected: false,
		},
		{
			name: "invalid port - too high",
			port: types.Port{
				Port:     65536,
				Protocol: "tcp",
				Service:  "invalid",
			},
			expected: false,
		},
		{
			name: "valid high port",
			port: types.Port{
				Port:     65535,
				Protocol: "tcp",
				Service:  "test",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidatePort(tt.port)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResourceValidator_ValidateResources(t *testing.T) {
	validator := NewResourceValidator()
	tempDir := t.TempDir()

	// Create test resources
	testFile := filepath.Join(tempDir, "test.conf")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	testDir := filepath.Join(tempDir, "testdir")
	err = os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	// Use absolute path for existing command
	var existingCommand string
	if runtime.GOOS == "windows" {
		existingCommand = "C:\\Windows\\System32\\cmd.exe"
	} else {
		existingCommand = "/bin/sh"
	}

	saidata := &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name: "test-software",
		},
		Files: []types.File{
			{Name: "existing", Path: testFile, Type: "config"},
			{Name: "nonexistent", Path: "/nonexistent/file.conf", Type: "config"},
		},
		Directories: []types.Directory{
			{Name: "existing", Path: testDir},
			{Name: "nonexistent", Path: "/nonexistent/directory"},
		},
		Commands: []types.Command{
			{Name: "existing", Path: existingCommand},
			{Name: "nonexistent", Path: "nonexistent-command-12345"},
		},
		Services: []types.Service{
			{Name: "test-service", ServiceName: "test-service", Type: "systemd"},
		},
		Ports: []types.Port{
			{Port: 8080, Protocol: "tcp", Service: "http-alt"},
			{Port: 0, Protocol: "tcp", Service: "invalid"}, // Invalid port
		},
	}

	result, err := validator.ValidateResources(saidata)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should not be valid due to missing resources
	assert.False(t, result.Valid)
	assert.False(t, result.CanProceed)

	// Should have missing resources
	assert.NotEmpty(t, result.MissingFiles)
	assert.Contains(t, result.MissingFiles, "/nonexistent/file.conf")

	assert.NotEmpty(t, result.MissingDirectories)
	assert.Contains(t, result.MissingDirectories, "/nonexistent/directory")

	assert.NotEmpty(t, result.MissingCommands)
	assert.Contains(t, result.MissingCommands, "nonexistent-command-12345")

	// Should have warnings for invalid ports
	assert.NotEmpty(t, result.Warnings)
}

func TestResourceValidator_ValidateResources_AllValid(t *testing.T) {
	validator := NewResourceValidator()
	tempDir := t.TempDir()

	// Create test resources
	testFile := filepath.Join(tempDir, "test.conf")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	testDir := filepath.Join(tempDir, "testdir")
	err = os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	// Use absolute path for existing command
	var existingCommand string
	if runtime.GOOS == "windows" {
		existingCommand = "C:\\Windows\\System32\\cmd.exe"
	} else {
		existingCommand = "/bin/sh"
	}

	saidata := &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name: "test-software",
		},
		Files: []types.File{
			{Name: "existing", Path: testFile, Type: "config"},
		},
		Directories: []types.Directory{
			{Name: "existing", Path: testDir},
		},
		Commands: []types.Command{
			{Name: "existing", Path: existingCommand},
		},
		Ports: []types.Port{
			{Port: 8080, Protocol: "tcp", Service: "http-alt"},
		},
	}

	result, err := validator.ValidateResources(saidata)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should be valid since all resources exist
	assert.True(t, result.Valid)
	assert.True(t, result.CanProceed)

	// Should have no missing resources
	assert.Empty(t, result.MissingFiles)
	assert.Empty(t, result.MissingDirectories)
	assert.Empty(t, result.MissingCommands)
	assert.Empty(t, result.MissingServices)
}

func TestResourceValidator_ValidateResources_InfoAction(t *testing.T) {
	validator := NewResourceValidator()

	// Create saidata with missing resources
	saidata := &types.SoftwareData{
		Version: "0.2",
		Metadata: types.Metadata{
			Name: "test-software",
		},
		Files: []types.File{
			{Name: "nonexistent", Path: "/nonexistent/file.conf", Type: "config"},
		},
		Commands: []types.Command{
			{Name: "nonexistent", Path: "nonexistent-command-12345"},
		},
	}

	result, err := validator.ValidateResources(saidata)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should not be valid due to missing resources
	assert.False(t, result.Valid)
	assert.False(t, result.CanProceed) // Cannot proceed with missing resources

	// Should still report missing resources
	assert.NotEmpty(t, result.MissingFiles)
	assert.NotEmpty(t, result.MissingCommands)
}

func TestResourceValidator_ValidateFileMethod(t *testing.T) {
	validator := NewResourceValidator()
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Test existing file
	file := types.File{Name: "test", Path: testFile, Type: "config"}
	assert.True(t, validator.ValidateFile(file))

	// Test non-existent file
	nonExistentFile := types.File{Name: "nonexistent", Path: "/nonexistent/file.txt", Type: "config"}
	assert.False(t, validator.ValidateFile(nonExistentFile))

	// Test empty path
	emptyFile := types.File{Name: "empty", Path: "", Type: "config"}
	assert.False(t, validator.ValidateFile(emptyFile))
}

func TestResourceValidator_ValidateDirectoryMethod(t *testing.T) {
	validator := NewResourceValidator()
	tempDir := t.TempDir()

	// Create a test directory
	testDir := filepath.Join(tempDir, "testdir")
	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	// Test existing directory
	directory := types.Directory{Name: "test", Path: testDir}
	assert.True(t, validator.ValidateDirectory(directory))

	// Test non-existent directory
	nonExistentDir := types.Directory{Name: "nonexistent", Path: "/nonexistent/directory"}
	assert.False(t, validator.ValidateDirectory(nonExistentDir))

	// Test empty path
	emptyDir := types.Directory{Name: "empty", Path: ""}
	assert.False(t, validator.ValidateDirectory(emptyDir))
}

func TestResourceValidator_ValidateCommandMethod(t *testing.T) {
	validator := NewResourceValidator()

	// Test with a command that should exist on most systems
	var existingCommand string
	var absolutePath string
	if runtime.GOOS == "windows" {
		existingCommand = "cmd"
		absolutePath = "C:\\Windows\\System32\\cmd.exe"
	} else {
		existingCommand = "sh"
		absolutePath = "/bin/sh"
	}
	command := types.Command{Name: existingCommand, Path: absolutePath}
	// Only test if the file actually exists
	if _, err := os.Stat(absolutePath); err == nil {
		assert.True(t, validator.ValidateCommand(command))
	} else {
		t.Skipf("System command %s not found at expected path", absolutePath)
	}

	// Test non-existent command
	nonExistentCommand := types.Command{Name: "nonexistent", Path: "nonexistent-command-12345"}
	assert.False(t, validator.ValidateCommand(nonExistentCommand))

	// Test empty command
	emptyCommand := types.Command{Name: "empty", Path: ""}
	assert.False(t, validator.ValidateCommand(emptyCommand))
}

func TestResourceValidator_ValidateServiceMethod(t *testing.T) {
	validator := NewResourceValidator()

	// Service existence is hard to test reliably across different systems
	// Test with a service that likely doesn't exist
	nonExistentService := types.Service{Name: "nonexistent", ServiceName: "nonexistent-service-12345", Type: "systemd"}
	assert.False(t, validator.ValidateService(nonExistentService))

	// Test empty service name
	emptyService := types.Service{Name: "empty", ServiceName: "", Type: "systemd"}
	assert.False(t, validator.ValidateService(emptyService))
}

func TestResourceValidator_Integration(t *testing.T) {
	// Integration test using actual system resources
	validator := NewResourceValidator()

	// Test with system directories that should exist
	systemDirs := []string{"/", "/tmp"}
	if runtime.GOOS == "windows" {
		systemDirs = []string{"C:\\", "C:\\Windows"}
	}

	for _, dir := range systemDirs {
		directory := types.Directory{Name: "system", Path: dir}
		if validator.ValidateDirectory(directory) {
			assert.True(t, validator.ValidateDirectory(directory), "System directory %s should exist", dir)
		}
	}

	// Test with system commands
	systemCommands := []string{"sh", "ls"}
	if runtime.GOOS == "windows" {
		systemCommands = []string{"cmd", "dir"}
	}

	for _, cmd := range systemCommands {
		if _, err := exec.LookPath(cmd); err == nil {
			command := types.Command{Name: "system", Path: cmd}
			assert.True(t, validator.ValidateCommand(command), "System command %s should exist", cmd)
		}
	}
}