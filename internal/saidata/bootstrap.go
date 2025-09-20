package saidata

import (
	"fmt"
	"os"
)

// Bootstrap handles first-time saidata setup
type Bootstrap struct {
	repositoryManager *RepositoryManager
}

// NewBootstrap creates a new bootstrap instance
func NewBootstrap(gitURL, zipFallbackURL string) *Bootstrap {
	return &Bootstrap{
		repositoryManager: NewRepositoryManager(gitURL, zipFallbackURL),
	}
}

// CheckAndInitialize checks if this is the first run and initializes saidata if needed
func (b *Bootstrap) CheckAndInitialize() error {
	if !b.repositoryManager.IsFirstRun() {
		// Repository already exists and is valid
		return nil
	}
	
	// Show welcome message
	if err := b.repositoryManager.ShowWelcomeMessage(); err != nil {
		return fmt.Errorf("failed to show welcome message: %w", err)
	}
	
	// Initialize repository
	if err := b.repositoryManager.InitializeRepository(); err != nil {
		return fmt.Errorf("failed to initialize saidata repository: %w", err)
	}
	
	return nil
}

// GetRepositoryManager returns the repository manager
func (b *Bootstrap) GetRepositoryManager() *RepositoryManager {
	return b.repositoryManager
}

// EnsureSaidataAvailable ensures saidata is available, initializing if necessary
func EnsureSaidataAvailable(gitURL, zipFallbackURL string) (string, error) {
	// For development/testing, check if docs/saidata_samples exists and use it
	if _, err := os.Stat("docs/saidata_samples"); err == nil {
		return "docs/saidata_samples", nil
	}
	
	bootstrap := NewBootstrap(gitURL, zipFallbackURL)
	
	// Check and initialize if needed
	if err := bootstrap.CheckAndInitialize(); err != nil {
		return "", err
	}
	
	// Return the saidata path
	return GetSaidataPath(), nil
}

// IsRunningAsRoot checks if the current process is running as root
func IsRunningAsRoot() bool {
	return os.Getuid() == 0
}

// GetInstallationInfo returns information about the current installation
func GetInstallationInfo() map[string]interface{} {
	info := make(map[string]interface{})
	
	info["saidata_path"] = GetSaidataPath()
	info["is_root"] = IsRunningAsRoot()
	info["first_run"] = NewRepositoryManager("", "").IsFirstRun()
	
	return info
}