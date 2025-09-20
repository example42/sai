package saidata

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManagerWithBootstrap(t *testing.T) {
	// Test with invalid URLs to ensure it doesn't crash
	manager, err := NewManagerWithBootstrap("", "")
	
	// This should either succeed (if docs/saidata_samples exists) or fail gracefully
	if err != nil {
		// Expected failure with invalid URLs
		t.Logf("NewManagerWithBootstrap failed as expected with invalid URLs: %v", err)
		return
	}
	
	if manager == nil {
		t.Error("Manager should not be nil even with invalid URLs")
	}
	
	// Test loading software (should work with docs/saidata_samples in development)
	_, err = manager.LoadSoftware("nginx")
	if err != nil {
		t.Logf("LoadSoftware failed (expected in test environment): %v", err)
	}
}

func TestBootstrapIntegration(t *testing.T) {
	// Create a temporary directory to simulate a clean environment
	tmpDir, err := os.MkdirTemp("", "sai-bootstrap-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create bootstrap with test directory
	bootstrap := NewBootstrap("", "")
	
	// Override the local path for testing
	bootstrap.repositoryManager.localPath = tmpDir
	
	// Should be first run
	if !bootstrap.repositoryManager.IsFirstRun() {
		t.Error("Expected first run for empty directory")
	}
	
	// Create a minimal valid structure
	softwareDir := filepath.Join(tmpDir, "software")
	if err := os.MkdirAll(softwareDir, 0755); err != nil {
		t.Fatalf("Failed to create software dir: %v", err)
	}
	
	// Create a sample YAML file
	apacheDir := filepath.Join(softwareDir, "ap", "apache")
	if err := os.MkdirAll(apacheDir, 0755); err != nil {
		t.Fatalf("Failed to create apache dir: %v", err)
	}
	
	sampleYAML := `version: "0.2"
metadata:
  name: "apache"
  description: "Apache HTTP Server"
packages:
  - name: "apache2"
services:
  - name: "apache2"
    service_name: "apache2"
`
	
	yamlPath := filepath.Join(apacheDir, "default.yaml")
	if err := os.WriteFile(yamlPath, []byte(sampleYAML), 0644); err != nil {
		t.Fatalf("Failed to write sample YAML: %v", err)
	}
	
	// Should not be first run anymore
	if bootstrap.repositoryManager.IsFirstRun() {
		t.Error("Expected not first run with valid structure")
	}
	
	// Validate repository should pass
	if err := bootstrap.repositoryManager.ValidateRepository(); err != nil {
		t.Errorf("Repository validation failed: %v", err)
	}
	
	// Get status should show healthy
	status, err := bootstrap.repositoryManager.GetRepositoryStatus()
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}
	
	if !status.IsHealthy {
		t.Error("Expected healthy status with valid structure")
	}
	
	if status.FileCount == 0 {
		t.Error("Expected non-zero file count")
	}
}

func TestEnsureSaidataAvailableWithDocs(t *testing.T) {
	// Change to project root directory for this test
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)
	
	// Go up two levels to project root
	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}
	
	// Check if docs/saidata_samples exists
	if _, err := os.Stat("docs/saidata_samples"); err != nil {
		t.Skip("Skipping test - docs/saidata_samples not available")
	}
	
	// This should work in development environment
	path, err := EnsureSaidataAvailable("", "")
	if err != nil {
		t.Fatalf("EnsureSaidataAvailable failed: %v", err)
	}
	
	if path == "" {
		t.Error("Expected non-empty path")
	}
	
	if path != "docs/saidata_samples" {
		t.Errorf("Expected path 'docs/saidata_samples', got '%s'", path)
	}
	
	// Path should exist
	if _, err := os.Stat(path); err != nil {
		t.Errorf("Saidata path does not exist: %s", path)
	}
}