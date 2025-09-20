package saidata

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// RepositoryManager handles saidata repository operations
type RepositoryManager struct {
	gitURL         string
	zipFallbackURL string
	localPath      string
	isRoot         bool
}

// RepositoryStatus represents the current status of the saidata repository
type RepositoryStatus struct {
	Type       string    `json:"type"`        // "git" or "zip"
	LastUpdate time.Time `json:"last_update"`
	LocalPath  string    `json:"local_path"`
	RemoteURL  string    `json:"remote_url"`
	IsHealthy  bool      `json:"is_healthy"`
	Version    string    `json:"version"`
	CommitHash string    `json:"commit_hash,omitempty"` // For git repositories
	FileCount  int       `json:"file_count"`
	SizeBytes  int64     `json:"size_bytes"`
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(gitURL, zipFallbackURL string) *RepositoryManager {
	localPath := GetSaidataPath()
	
	return &RepositoryManager{
		gitURL:         gitURL,
		zipFallbackURL: zipFallbackURL,
		localPath:      localPath,
		isRoot:         os.Getuid() == 0,
	}
}

// GetSaidataPath returns the appropriate saidata directory path
func GetSaidataPath() string {
	// Check if running as root
	if os.Getuid() == 0 {
		// Root installation: /etc/sai/saidata (Linux/macOS) or equivalent on Windows
		if runtime.GOOS == "windows" {
			return filepath.Join("C:", "ProgramData", "sai", "saidata")
		}
		return "/etc/sai/saidata"
	}
	
	// User installation: $HOME/.sai/saidata
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory cannot be determined
		return filepath.Join(".", ".sai", "saidata")
	}
	return filepath.Join(homeDir, ".sai", "saidata")
}

// IsFirstRun checks if this is the first time SAI is being executed
func (rm *RepositoryManager) IsFirstRun() bool {
	// Check if saidata directory exists and has content
	if _, err := os.Stat(rm.localPath); os.IsNotExist(err) {
		return true
	}
	
	// Check if directory is empty or doesn't contain expected structure
	entries, err := os.ReadDir(rm.localPath)
	if err != nil || len(entries) == 0 {
		return true
	}
	
	// Check for expected saidata structure (software directory or prefix directories)
	hasValidStructure := false
	for _, entry := range entries {
		if entry.IsDir() && (entry.Name() == "software" || len(entry.Name()) == 2) {
			hasValidStructure = true
			break
		}
	}
	
	return !hasValidStructure
}

// ShowWelcomeMessage displays a welcome message for first-time users
func (rm *RepositoryManager) ShowWelcomeMessage() error {
	fmt.Println("🎉 Welcome to SAI - Software Action Interface!")
	fmt.Println()
	fmt.Println("SAI provides a unified interface for managing software across different")
	fmt.Println("operating systems and package managers. To get started, SAI needs to")
	fmt.Println("download the saidata repository containing software definitions.")
	fmt.Println()
	
	if rm.isRoot {
		fmt.Printf("📁 Installing saidata to system directory: %s\n", rm.localPath)
	} else {
		fmt.Printf("📁 Installing saidata to user directory: %s\n", rm.localPath)
	}
	
	fmt.Println("🔄 Downloading saidata repository...")
	fmt.Println()
	
	return nil
}

// InitializeRepository sets up the saidata repository for the first time
func (rm *RepositoryManager) InitializeRepository() error {
	// Create the directory structure
	if err := os.MkdirAll(rm.localPath, 0755); err != nil {
		return fmt.Errorf("failed to create saidata directory: %w", err)
	}
	
	// Try Git clone first, fallback to zip download
	if err := rm.gitClone(); err != nil {
		fmt.Printf("⚠️  Git clone failed: %v\n", err)
		fmt.Println("🔄 Falling back to zip download...")
		
		if err := rm.zipDownload(); err != nil {
			return fmt.Errorf("both git clone and zip download failed: %w", err)
		}
	}
	
	// Validate the downloaded repository
	if err := rm.ValidateRepository(); err != nil {
		return fmt.Errorf("repository validation failed: %w", err)
	}
	
	fmt.Println("✅ Saidata repository successfully initialized!")
	fmt.Println()
	fmt.Println("You can now use SAI to manage software. Try:")
	fmt.Println("  sai install nginx")
	fmt.Println("  sai list")
	fmt.Println("  sai stats")
	fmt.Println()
	
	return nil
}

// gitClone attempts to clone the repository using Git
func (rm *RepositoryManager) gitClone() error {
	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found in PATH")
	}
	
	// Remove existing directory if it exists but is empty or invalid
	if _, err := os.Stat(rm.localPath); err == nil {
		if err := os.RemoveAll(rm.localPath); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}
	
	// Clone the repository
	cmd := exec.Command("git", "clone", rm.gitURL, rm.localPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	
	return nil
}

// zipDownload downloads and extracts the repository as a zip file
func (rm *RepositoryManager) zipDownload() error {
	if rm.zipFallbackURL == "" {
		return fmt.Errorf("no zip fallback URL configured")
	}
	
	// Create temporary file for download
	tmpFile, err := os.CreateTemp("", "saidata-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	
	// Download the zip file
	resp, err := http.Get(rm.zipFallbackURL)
	if err != nil {
		return fmt.Errorf("failed to download zip file: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}
	
	// Copy response to temporary file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to save zip file: %w", err)
	}
	
	tmpFile.Close()
	
	// Extract the zip file
	if err := rm.extractZip(tmpFile.Name()); err != nil {
		return fmt.Errorf("failed to extract zip file: %w", err)
	}
	
	return nil
}

// extractZip extracts a zip file to the local path
func (rm *RepositoryManager) extractZip(zipPath string) error {
	// Open zip file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()
	
	// Remove existing directory if it exists
	if _, err := os.Stat(rm.localPath); err == nil {
		if err := os.RemoveAll(rm.localPath); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}
	
	// Create base directory
	if err := os.MkdirAll(rm.localPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Extract files
	for _, file := range reader.File {
		// Skip the root directory (usually named like "saidata-main/")
		pathParts := strings.Split(file.Name, "/")
		if len(pathParts) <= 1 {
			continue
		}
		
		// Remove the first path component (root directory)
		relativePath := strings.Join(pathParts[1:], "/")
		if relativePath == "" {
			continue
		}
		
		destPath := filepath.Join(rm.localPath, relativePath)
		
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, file.FileInfo().Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destPath, err)
			}
			continue
		}
		
		// Create parent directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", destPath, err)
		}
		
		// Extract file
		if err := rm.extractFile(file, destPath); err != nil {
			return fmt.Errorf("failed to extract file %s: %w", destPath, err)
		}
	}
	
	return nil
}

// extractFile extracts a single file from the zip archive
func (rm *RepositoryManager) extractFile(file *zip.File, destPath string) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	
	outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()
	
	_, err = io.Copy(outFile, rc)
	return err
}

// UpdateRepository updates the saidata repository
func (rm *RepositoryManager) UpdateRepository() error {
	if !rm.repositoryExists() {
		return fmt.Errorf("repository not initialized, run 'sai saidata init' first")
	}
	
	// Check if it's a git repository
	if rm.isGitRepository() {
		return rm.gitPull()
	}
	
	// For zip-based repositories, re-download
	fmt.Println("🔄 Updating saidata repository (zip-based)...")
	return rm.zipDownload()
}

// gitPull updates a git-based repository
func (rm *RepositoryManager) gitPull() error {
	fmt.Println("🔄 Updating saidata repository (git-based)...")
	
	cmd := exec.Command("git", "pull")
	cmd.Dir = rm.localPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}
	
	fmt.Println("✅ Repository updated successfully!")
	return nil
}

// GetRepositoryStatus returns the current status of the repository
func (rm *RepositoryManager) GetRepositoryStatus() (*RepositoryStatus, error) {
	status := &RepositoryStatus{
		LocalPath: rm.localPath,
		RemoteURL: rm.gitURL,
	}
	
	// Check if repository exists
	if !rm.repositoryExists() {
		status.IsHealthy = false
		return status, nil
	}
	
	// Determine repository type
	if rm.isGitRepository() {
		status.Type = "git"
		status.RemoteURL = rm.gitURL
		
		// Get commit hash
		if hash, err := rm.getGitCommitHash(); err == nil {
			status.CommitHash = hash
		}
	} else {
		status.Type = "zip"
		status.RemoteURL = rm.zipFallbackURL
	}
	
	// Get last modification time
	if info, err := os.Stat(rm.localPath); err == nil {
		status.LastUpdate = info.ModTime()
	}
	
	// Count files and calculate size
	fileCount, sizeBytes := rm.calculateRepositorySize()
	status.FileCount = fileCount
	status.SizeBytes = sizeBytes
	
	// Validate repository health
	status.IsHealthy = rm.validateRepositoryHealth()
	
	return status, nil
}

// SynchronizeRepository synchronizes the repository (alias for UpdateRepository)
func (rm *RepositoryManager) SynchronizeRepository() error {
	return rm.UpdateRepository()
}

// ValidateRepository validates the repository structure and content
func (rm *RepositoryManager) ValidateRepository() error {
	if !rm.repositoryExists() {
		return fmt.Errorf("repository directory does not exist: %s", rm.localPath)
	}
	
	// Check for expected directory structure
	expectedDirs := []string{"software", "ap", "br", "do", "el", "gr", "je", "ku", "mo", "my", "ng", "pr", "re", "te"}
	hasValidStructure := false
	
	for _, dir := range expectedDirs {
		dirPath := filepath.Join(rm.localPath, dir)
		if _, err := os.Stat(dirPath); err == nil {
			hasValidStructure = true
			break
		}
	}
	
	if !hasValidStructure {
		return fmt.Errorf("repository does not contain expected saidata structure")
	}
	
	return nil
}

// repositoryExists checks if the repository directory exists
func (rm *RepositoryManager) repositoryExists() bool {
	if _, err := os.Stat(rm.localPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// isGitRepository checks if the local path is a git repository
func (rm *RepositoryManager) isGitRepository() bool {
	gitDir := filepath.Join(rm.localPath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return true
	}
	return false
}

// getGitCommitHash returns the current git commit hash
func (rm *RepositoryManager) getGitCommitHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = rm.localPath
	
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	return strings.TrimSpace(string(output)), nil
}

// calculateRepositorySize calculates the number of files and total size
func (rm *RepositoryManager) calculateRepositorySize() (int, int64) {
	var fileCount int
	var totalSize int64
	
	filepath.Walk(rm.localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() {
			fileCount++
			totalSize += info.Size()
		}
		
		return nil
	})
	
	return fileCount, totalSize
}

// validateRepositoryHealth performs basic health checks on the repository
func (rm *RepositoryManager) validateRepositoryHealth() bool {
	// Check if directory is readable
	if _, err := os.ReadDir(rm.localPath); err != nil {
		return false
	}
	
	// Check for at least some YAML files
	yamlCount := 0
	filepath.Walk(rm.localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
			yamlCount++
		}
		
		return nil
	})
	
	return yamlCount > 0
}

// CleanRepository removes the local repository
func (rm *RepositoryManager) CleanRepository() error {
	if !rm.repositoryExists() {
		return fmt.Errorf("repository does not exist: %s", rm.localPath)
	}
	
	if err := os.RemoveAll(rm.localPath); err != nil {
		return fmt.Errorf("failed to remove repository: %w", err)
	}
	
	fmt.Printf("✅ Repository cleaned: %s\n", rm.localPath)
	return nil
}