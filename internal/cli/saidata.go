package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"sai/internal/saidata"
)

var saidataCmd = &cobra.Command{
	Use:   "saidata",
	Short: "Manage saidata repository",
	Long: `Manage the saidata repository that contains software definitions and configurations.

The saidata repository provides metadata for software packages, services, and configurations
across different operating systems and package managers. This command allows you to:

  • Check repository status and health
  • Update the repository with latest definitions
  • Synchronize with remote repository
  • Initialize or reinitialize the repository
  • Clean and reset the local repository

Examples:
  sai saidata status          # Show repository status
  sai saidata update          # Update repository from remote
  sai saidata sync            # Synchronize with remote (alias for update)
  sai saidata init            # Initialize or reinitialize repository
  sai saidata clean           # Remove local repository`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default action is to show status
		return runSaidataStatus(cmd, args)
	},
}

var saidataStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show saidata repository status",
	Long: `Display detailed information about the current saidata repository including:

  • Repository type (git or zip-based)
  • Last update time
  • Local path and remote URL
  • Health status and validation
  • File count and repository size
  • Git commit hash (for git repositories)

This command helps diagnose repository issues and verify the installation.`,
	RunE: runSaidataStatus,
}

var saidataUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update saidata repository",
	Long: `Update the saidata repository with the latest software definitions and configurations.

For git-based repositories, this performs a 'git pull' to fetch the latest changes.
For zip-based repositories, this re-downloads and extracts the latest archive.

The update process:
  1. Validates the current repository
  2. Fetches latest changes from remote
  3. Validates the updated repository
  4. Reports success or failure

Use this command regularly to ensure you have the latest software definitions.`,
	RunE: runSaidataUpdate,
}

var saidataSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize saidata repository (alias for update)",
	Long: `Synchronize the saidata repository with the remote source.

This is an alias for the 'update' command and performs the same operation:
updating the local repository with the latest changes from the remote source.`,
	RunE: runSaidataSync,
}

var saidataInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize or reinitialize saidata repository",
	Long: `Initialize or reinitialize the saidata repository.

This command will:
  1. Remove any existing repository (if present)
  2. Download the repository from the configured source
  3. Validate the downloaded repository
  4. Set up the local directory structure

Use this command to:
  • Set up saidata for the first time
  • Fix a corrupted repository
  • Switch between git and zip-based repositories
  • Reset to a clean state

Warning: This will remove any local modifications to the repository.`,
	RunE: runSaidataInit,
}

var saidataCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove local saidata repository",
	Long: `Remove the local saidata repository completely.

This command will:
  1. Remove the entire local repository directory
  2. Clear any cached data
  3. Reset SAI to first-run state

After running this command, SAI will automatically re-download the repository
on the next operation that requires saidata.

Use this command to:
  • Free up disk space
  • Reset to a clean state
  • Troubleshoot repository issues

Warning: This will remove the entire local repository and any local modifications.`,
	RunE: runSaidataClean,
}

func init() {
	// Add saidata command to root
	rootCmd.AddCommand(saidataCmd)
	
	// Add subcommands
	saidataCmd.AddCommand(saidataStatusCmd)
	saidataCmd.AddCommand(saidataUpdateCmd)
	saidataCmd.AddCommand(saidataSyncCmd)
	saidataCmd.AddCommand(saidataInitCmd)
	saidataCmd.AddCommand(saidataCleanCmd)
}

func runSaidataStatus(cmd *cobra.Command, args []string) error {
	cfg := GetGlobalConfig()
	flags := GetGlobalFlags()
	
	// Create repository manager
	repoManager := saidata.NewRepositoryManager(cfg.Repository.GitURL, cfg.Repository.ZipFallbackURL)
	
	// Get repository status
	status, err := repoManager.GetRepositoryStatus()
	if err != nil {
		return fmt.Errorf("failed to get repository status: %w", err)
	}
	
	// Output in JSON format if requested
	if flags.JSONOutput {
		jsonData, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal status to JSON: %w", err)
		}
		fmt.Println(string(jsonData))
		return nil
	}
	
	// Display status in human-readable format
	fmt.Println("📊 Saidata Repository Status")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Printf("Local Path:    %s\n", status.LocalPath)
	fmt.Printf("Remote URL:    %s\n", status.RemoteURL)
	fmt.Printf("Type:          %s\n", status.Type)
	
	if status.IsHealthy {
		fmt.Printf("Health:        ✅ Healthy\n")
	} else {
		fmt.Printf("Health:        ❌ Unhealthy\n")
	}
	
	if !status.LastUpdate.IsZero() {
		fmt.Printf("Last Update:   %s\n", status.LastUpdate.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("Last Update:   Never\n")
	}
	
	fmt.Printf("Files:         %d\n", status.FileCount)
	fmt.Printf("Size:          %.2f MB\n", float64(status.SizeBytes)/(1024*1024))
	
	if status.CommitHash != "" {
		fmt.Printf("Commit Hash:   %s\n", status.CommitHash[:8])
	}
	
	fmt.Println()
	
	// Show additional information if not healthy
	if !status.IsHealthy {
		fmt.Println("⚠️  Repository Issues Detected")
		fmt.Println("Try running 'sai saidata init' to reinitialize the repository.")
		fmt.Println()
	}
	
	return nil
}

func runSaidataUpdate(cmd *cobra.Command, args []string) error {
	cfg := GetGlobalConfig()
	
	// Create repository manager
	repoManager := saidata.NewRepositoryManager(cfg.Repository.GitURL, cfg.Repository.ZipFallbackURL)
	
	// Update repository
	if err := repoManager.UpdateRepository(); err != nil {
		return fmt.Errorf("failed to update repository: %w", err)
	}
	
	return nil
}

func runSaidataSync(cmd *cobra.Command, args []string) error {
	cfg := GetGlobalConfig()
	
	// Create repository manager
	repoManager := saidata.NewRepositoryManager(cfg.Repository.GitURL, cfg.Repository.ZipFallbackURL)
	
	// Synchronize repository
	if err := repoManager.SynchronizeRepository(); err != nil {
		return fmt.Errorf("failed to synchronize repository: %w", err)
	}
	
	return nil
}

func runSaidataInit(cmd *cobra.Command, args []string) error {
	cfg := GetGlobalConfig()
	flags := GetGlobalFlags()
	
	// Create repository manager
	repoManager := saidata.NewRepositoryManager(cfg.Repository.GitURL, cfg.Repository.ZipFallbackURL)
	
	// Check if repository already exists
	status, err := repoManager.GetRepositoryStatus()
	if err == nil && status.IsHealthy {
		if !flags.Yes {
			fmt.Printf("⚠️  Saidata repository already exists at: %s\n", status.LocalPath)
			fmt.Print("This will remove the existing repository and re-download it. Continue? (y/N): ")
			
			var response string
			fmt.Scanln(&response)
			
			if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
				fmt.Println("Operation cancelled.")
				return nil
			}
		}
	}
	
	// Show welcome message
	if err := repoManager.ShowWelcomeMessage(); err != nil {
		return fmt.Errorf("failed to show welcome message: %w", err)
	}
	
	// Initialize repository
	if err := repoManager.InitializeRepository(); err != nil {
		return fmt.Errorf("failed to initialize repository: %w", err)
	}
	
	return nil
}

func runSaidataClean(cmd *cobra.Command, args []string) error {
	cfg := GetGlobalConfig()
	flags := GetGlobalFlags()
	
	// Create repository manager
	repoManager := saidata.NewRepositoryManager(cfg.Repository.GitURL, cfg.Repository.ZipFallbackURL)
	
	// Get current status
	status, err := repoManager.GetRepositoryStatus()
	if err != nil || !status.IsHealthy {
		fmt.Println("ℹ️  No valid repository found to clean.")
		return nil
	}
	
	// Confirm operation unless --yes flag is used
	if !flags.Yes {
		fmt.Printf("⚠️  This will permanently remove the saidata repository at: %s\n", status.LocalPath)
		fmt.Printf("Repository contains %d files (%.2f MB)\n", status.FileCount, float64(status.SizeBytes)/(1024*1024))
		fmt.Print("Are you sure you want to continue? (y/N): ")
		
		var response string
		fmt.Scanln(&response)
		
		if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
			fmt.Println("Operation cancelled.")
			return nil
		}
	}
	
	// Clean repository
	if err := repoManager.CleanRepository(); err != nil {
		return fmt.Errorf("failed to clean repository: %w", err)
	}
	
	fmt.Println("ℹ️  SAI will automatically re-download the repository on next use.")
	return nil
}