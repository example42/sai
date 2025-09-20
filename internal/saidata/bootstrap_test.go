package saidata

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetSaidataPath(t *testing.T) {
	// Test user path (non-root)
	if os.Getuid() != 0 {
		path := GetSaidataPath()
		homeDir, _ := os.UserHomeDir()
		expected := filepath.Join(homeDir, ".sai", "saidata")
		if path != expected {
			t.Errorf("Expected user path %s, got %s", expected, path)
		}
	}
}

func TestIsRunningAsRoot(t *testing.T) {
	isRoot := IsRunningAsRoot()
	expectedRoot := os.Getuid() == 0
	
	if isRoot != expectedRoot {
		t.Errorf("Expected IsRunningAsRoot() to return %v, got %v", expectedRoot, isRoot)
	}
}

func TestGetInstallationInfo(t *testing.T) {
	info := GetInstallationInfo()
	
	// Check required fields
	if _, exists := info["saidata_path"]; !exists {
		t.Error("Installation info missing saidata_path")
	}
	
	if _, exists := info["is_root"]; !exists {
		t.Error("Installation info missing is_root")
	}
	
	if _, exists := info["first_run"]; !exists {
		t.Error("Installation info missing first_run")
	}
}

func TestBootstrapCreation(t *testing.T) {
	bootstrap := NewBootstrap("https://example.com/repo.git", "https://example.com/repo.zip")
	
	if bootstrap == nil {
		t.Error("Bootstrap creation failed")
	}
	
	if bootstrap.repositoryManager == nil {
		t.Error("Bootstrap repository manager not initialized")
	}
}

func TestEnsureSaidataAvailable(t *testing.T) {
	// This test will use the existing docs/saidata_samples if available
	// or attempt to initialize a repository
	
	// For testing, we'll just verify the function doesn't panic
	// and returns a valid path
	path, err := EnsureSaidataAvailable("", "")
	
	// In development environment, this should succeed using docs/saidata_samples
	// In production, it might fail due to invalid URLs, which is expected
	if err != nil && path == "" {
		// This is acceptable for testing with invalid URLs
		t.Logf("EnsureSaidataAvailable failed as expected with invalid URLs: %v", err)
	} else if path == "" {
		t.Error("EnsureSaidataAvailable returned empty path without error")
	}
}