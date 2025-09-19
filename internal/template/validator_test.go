package template

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSystemResourceValidator_FileExists(t *testing.T) {
	validator := NewSystemResourceValidator()
	
	// Create a temporary file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	
	// File doesn't exist yet
	if validator.FileExists(tmpFile) {
		t.Error("Expected file to not exist")
	}
	
	// Create the file
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// File should exist now
	if !validator.FileExists(tmpFile) {
		t.Error("Expected file to exist")
	}
	
	// Test with empty path
	if validator.FileExists("") {
		t.Error("Expected empty path to return false")
	}
	
	// Test with directory (should return false for FileExists)
	if validator.FileExists(tmpDir) {
		t.Error("Expected directory to return false for FileExists")
	}
}

func TestSystemResourceValidator_DirectoryExists(t *testing.T) {
	validator := NewSystemResourceValidator()
	
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "testdir")
	
	// Directory doesn't exist yet
	if validator.DirectoryExists(testDir) {
		t.Error("Expected directory to not exist")
	}
	
	// Create the directory
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Directory should exist now
	if !validator.DirectoryExists(testDir) {
		t.Error("Expected directory to exist")
	}
	
	// Test with empty path
	if validator.DirectoryExists("") {
		t.Error("Expected empty path to return false")
	}
	
	// Test with file (should return false for DirectoryExists)
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	if validator.DirectoryExists(tmpFile) {
		t.Error("Expected file to return false for DirectoryExists")
	}
}

func TestSystemResourceValidator_CommandExists(t *testing.T) {
	validator := NewSystemResourceValidator()
	
	// Test with empty command
	if validator.CommandExists("") {
		t.Error("Expected empty command to return false")
	}
	
	// Test with a command that should exist on most systems
	var testCommand string
	switch runtime.GOOS {
	case "windows":
		testCommand = "cmd"
	default:
		testCommand = "sh"
	}
	
	if !validator.CommandExists(testCommand) {
		t.Errorf("Expected %s command to exist", testCommand)
	}
	
	// Test with a command that shouldn't exist
	if validator.CommandExists("nonexistent-command-12345") {
		t.Error("Expected nonexistent command to return false")
	}
	
	// Test with absolute path
	tmpDir := t.TempDir()
	var testScript string
	if runtime.GOOS == "windows" {
		testScript = filepath.Join(tmpDir, "test.bat")
		if err := os.WriteFile(testScript, []byte("@echo off\necho test"), 0755); err != nil {
			t.Fatalf("Failed to create test script: %v", err)
		}
	} else {
		testScript = filepath.Join(tmpDir, "test.sh")
		if err := os.WriteFile(testScript, []byte("#!/bin/sh\necho test"), 0755); err != nil {
			t.Fatalf("Failed to create test script: %v", err)
		}
	}
	
	if !validator.CommandExists(testScript) {
		t.Error("Expected absolute path script to exist")
	}
}

func TestSystemResourceValidator_ServiceExists(t *testing.T) {
	validator := NewSystemResourceValidator()
	
	// Test with empty service name
	if validator.ServiceExists("") {
		t.Error("Expected empty service name to return false")
	}
	
	// Note: Service existence testing is platform-specific and may require
	// actual services to be installed. For unit tests, we mainly test
	// that the function doesn't panic and handles edge cases.
	
	// Test with a service that likely doesn't exist
	if validator.ServiceExists("nonexistent-service-12345") {
		t.Error("Expected nonexistent service to return false")
	}
	
	// Platform-specific tests
	switch runtime.GOOS {
	case "linux":
		// On Linux systems, we can test some common services
		// but we can't guarantee they exist, so we just ensure no panic
		_ = validator.ServiceExists("systemd")
	case "darwin":
		// On macOS, test launchd-related functionality
		_ = validator.ServiceExists("com.apple.launchd")
	case "windows":
		// On Windows, test with a common service
		_ = validator.ServiceExists("Spooler")
	}
}

func TestSystemResourceValidator_checkLinuxService(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test on non-Linux platform")
	}
	
	validator := NewSystemResourceValidator()
	
	// Test the Linux-specific service checking logic
	// This is mainly to ensure the function doesn't panic
	result := validator.checkLinuxService("nonexistent-service")
	if result {
		t.Error("Expected nonexistent service to return false")
	}
}

func TestSystemResourceValidator_checkMacOSService(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-macOS platform")
	}
	
	validator := NewSystemResourceValidator()
	
	// Test the macOS-specific service checking logic
	result := validator.checkMacOSService("nonexistent-service")
	if result {
		t.Error("Expected nonexistent service to return false")
	}
}

func TestSystemResourceValidator_checkWindowsService(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}
	
	validator := NewSystemResourceValidator()
	
	// Test the Windows-specific service checking logic
	result := validator.checkWindowsService("nonexistent-service")
	if result {
		t.Error("Expected nonexistent service to return false")
	}
}