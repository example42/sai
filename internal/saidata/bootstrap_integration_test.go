package saidata

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompleteBootstrapWorkflow(t *testing.T) {
	// Create a temporary directory to simulate a clean environment
	tmpDir, err := os.MkdirTemp("", "sai-complete-bootstrap-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create a mock saidata repository structure
	repoDir := filepath.Join(tmpDir, "mock-repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatalf("Failed to create repo dir: %v", err)
	}
	
	// Create software directory structure
	softwareDir := filepath.Join(repoDir, "software")
	apacheDir := filepath.Join(softwareDir, "ap", "apache")
	nginxDir := filepath.Join(softwareDir, "ng", "nginx")
	
	if err := os.MkdirAll(apacheDir, 0755); err != nil {
		t.Fatalf("Failed to create apache dir: %v", err)
	}
	if err := os.MkdirAll(nginxDir, 0755); err != nil {
		t.Fatalf("Failed to create nginx dir: %v", err)
	}
	
	// Create sample saidata files
	apacheYAML := `version: "0.2"
metadata:
  name: "apache"
  description: "Apache HTTP Server"
  version: "2.4.58"
packages:
  - name: "apache2"
    package_name: "apache2"
services:
  - name: "apache2"
    service_name: "apache2"
files:
  - name: "config"
    path: "/etc/apache2/apache2.conf"
`
	
	nginxYAML := `version: "0.2"
metadata:
  name: "nginx"
  description: "Nginx HTTP Server"
  version: "1.24.0"
packages:
  - name: "nginx"
    package_name: "nginx"
services:
  - name: "nginx"
    service_name: "nginx"
files:
  - name: "config"
    path: "/etc/nginx/nginx.conf"
`
	
	// Write the YAML files
	if err := os.WriteFile(filepath.Join(apacheDir, "default.yaml"), []byte(apacheYAML), 0644); err != nil {
		t.Fatalf("Failed to write apache YAML: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nginxDir, "default.yaml"), []byte(nginxYAML), 0644); err != nil {
		t.Fatalf("Failed to write nginx YAML: %v", err)
	}
	
	// Test 1: Repository Manager Creation
	rm := &RepositoryManager{
		localPath: repoDir,
	}
	
	// Should not be first run with valid structure
	if rm.IsFirstRun() {
		t.Error("Expected not first run with valid structure")
	}
	
	// Test 2: Repository Validation
	if err := rm.ValidateRepository(); err != nil {
		t.Errorf("Repository validation failed: %v", err)
	}
	
	// Test 3: Repository Status
	status, err := rm.GetRepositoryStatus()
	if err != nil {
		t.Fatalf("Failed to get repository status: %v", err)
	}
	
	if !status.IsHealthy {
		t.Error("Expected healthy repository status")
	}
	
	if status.FileCount < 2 {
		t.Errorf("Expected at least 2 files, got %d", status.FileCount)
	}
	
	if status.SizeBytes == 0 {
		t.Error("Expected non-zero repository size")
	}
	
	// Test 4: Saidata Manager Integration
	manager := NewManager(repoDir)
	
	// Load apache software
	apacheData, err := manager.LoadSoftware("apache")
	if err != nil {
		t.Fatalf("Failed to load apache software: %v", err)
	}
	
	if apacheData.Metadata.Name != "apache" {
		t.Errorf("Expected apache name, got %s", apacheData.Metadata.Name)
	}
	
	if apacheData.Metadata.Version != "2.4.58" {
		t.Errorf("Expected version 2.4.58, got %s", apacheData.Metadata.Version)
	}
	
	if len(apacheData.Packages) == 0 {
		t.Error("Expected at least one package")
	}
	
	if len(apacheData.Services) == 0 {
		t.Error("Expected at least one service")
	}
	
	// Load nginx software
	nginxData, err := manager.LoadSoftware("nginx")
	if err != nil {
		t.Fatalf("Failed to load nginx software: %v", err)
	}
	
	if nginxData.Metadata.Name != "nginx" {
		t.Errorf("Expected nginx name, got %s", nginxData.Metadata.Name)
	}
	
	// Test 5: Search functionality
	searchResults, err := manager.SearchSoftware("apache")
	if err != nil {
		t.Fatalf("Failed to search software: %v", err)
	}
	
	if len(searchResults) == 0 {
		t.Error("Expected search results for apache")
	}
	
	found := false
	for _, result := range searchResults {
		if result.Software == "apache" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected to find apache in search results")
	}
	
	// Test 6: Software list functionality
	softwareList, err := manager.GetSoftwareList()
	if err != nil {
		t.Fatalf("Failed to get software list: %v", err)
	}
	
	if len(softwareList) < 2 {
		t.Errorf("Expected at least 2 software items, got %d", len(softwareList))
	}
	
	// Verify both apache and nginx are in the list
	hasApache := false
	hasNginx := false
	for _, software := range softwareList {
		if software == "apache" {
			hasApache = true
		}
		if software == "nginx" {
			hasNginx = true
		}
	}
	
	if !hasApache {
		t.Error("Expected apache in software list")
	}
	if !hasNginx {
		t.Error("Expected nginx in software list")
	}
	
	t.Logf("Bootstrap workflow test completed successfully")
	t.Logf("Repository status: %d files, %.2f KB", status.FileCount, float64(status.SizeBytes)/1024)
	t.Logf("Software found: %v", softwareList)
}