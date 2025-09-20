package saidata

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewRepositoryManager(t *testing.T) {
	gitURL := "https://example.com/repo.git"
	zipURL := "https://example.com/repo.zip"
	
	rm := NewRepositoryManager(gitURL, zipURL)
	
	if rm == nil {
		t.Error("Repository manager creation failed")
	}
	
	if rm.gitURL != gitURL {
		t.Errorf("Expected git URL %s, got %s", gitURL, rm.gitURL)
	}
	
	if rm.zipFallbackURL != zipURL {
		t.Errorf("Expected zip URL %s, got %s", zipURL, rm.zipFallbackURL)
	}
	
	if rm.localPath == "" {
		t.Error("Local path not set")
	}
}

func TestIsFirstRun(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "sai-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	rm := &RepositoryManager{
		localPath: tmpDir,
	}
	
	// Should be first run for empty directory
	if !rm.IsFirstRun() {
		t.Error("Expected first run for empty directory")
	}
	
	// Create a subdirectory to simulate saidata structure
	softwareDir := filepath.Join(tmpDir, "software")
	if err := os.MkdirAll(softwareDir, 0755); err != nil {
		t.Fatalf("Failed to create software dir: %v", err)
	}
	
	// Should not be first run with valid structure
	if rm.IsFirstRun() {
		t.Error("Expected not first run with valid structure")
	}
}

func TestGetRepositoryStatus(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "sai-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	rm := &RepositoryManager{
		gitURL:         "https://example.com/repo.git",
		zipFallbackURL: "https://example.com/repo.zip",
		localPath:      tmpDir,
	}
	
	status, err := rm.GetRepositoryStatus()
	if err != nil {
		t.Fatalf("Failed to get repository status: %v", err)
	}
	
	if status.LocalPath != tmpDir {
		t.Errorf("Expected local path %s, got %s", tmpDir, status.LocalPath)
	}
	
	// For non-git repositories, it should use the zip fallback URL
	expectedURL := rm.zipFallbackURL // Since it's not a git repo, it should use zip URL
	if status.RemoteURL != expectedURL {
		t.Errorf("Expected remote URL %s, got %s", expectedURL, status.RemoteURL)
	}
	
	// Empty directory should be unhealthy
	if status.IsHealthy {
		t.Error("Expected unhealthy status for empty directory")
	}
}

func TestValidateRepository(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "sai-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	rm := &RepositoryManager{
		localPath: tmpDir,
	}
	
	// Should fail validation for empty directory
	if err := rm.ValidateRepository(); err == nil {
		t.Error("Expected validation to fail for empty directory")
	}
	
	// Create a valid structure
	softwareDir := filepath.Join(tmpDir, "software")
	if err := os.MkdirAll(softwareDir, 0755); err != nil {
		t.Fatalf("Failed to create software dir: %v", err)
	}
	
	// Should pass validation with valid structure
	if err := rm.ValidateRepository(); err != nil {
		t.Errorf("Expected validation to pass with valid structure: %v", err)
	}
}

func TestRepositoryExists(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "sai-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	rm := &RepositoryManager{
		localPath: tmpDir,
	}
	
	// Should exist
	if !rm.repositoryExists() {
		t.Error("Expected repository to exist")
	}
	
	// Remove directory
	os.RemoveAll(tmpDir)
	
	// Should not exist
	if rm.repositoryExists() {
		t.Error("Expected repository to not exist")
	}
}

func TestIsGitRepository(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "sai-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	rm := &RepositoryManager{
		localPath: tmpDir,
	}
	
	// Should not be git repository
	if rm.isGitRepository() {
		t.Error("Expected not to be git repository")
	}
	
	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}
	
	// Should be git repository
	if !rm.isGitRepository() {
		t.Error("Expected to be git repository")
	}
}