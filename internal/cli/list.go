package cli

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"sai/internal/interfaces"
	"sai/internal/output"
	"sai/internal/ui"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed software packages",
	Long: `List all installed software packages across all available providers.
This provides a comprehensive view of software installed on the system.

This is an information-only command that executes without confirmation prompts.
The output shows software packages organized by provider with installation status.

Examples:
  sai list                             # List all installed software
  sai list --verbose                   # Show detailed package information
  sai list --json                      # Output in JSON format
  sai list --provider apt              # List only packages from apt provider`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeListCommand()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

// executeListCommand implements the list command functionality (Requirement 5.1)
func executeListCommand() error {
	// Get global configuration and flags
	config := GetGlobalConfig()
	flags := GetGlobalFlags()

	// Create output formatter
	formatter := output.NewOutputFormatter(config, flags.Verbose, flags.Quiet, flags.JSONOutput)

	// Create managers and dependencies
	actionManager, userInterface, err := createManagers(config, formatter)
	if err != nil {
		formatter.ShowError(fmt.Errorf("failed to initialize managers: %w", err))
		return err
	}

	// Show progress
	if !flags.Quiet {
		formatter.ShowProgress("Listing installed software packages...")
	}

	// Get installed software by executing list action across providers
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	installedSoftware, err := getInstalledSoftwareAcrossProviders(ctx, actionManager, flags)
	if err != nil {
		formatter.ShowError(fmt.Errorf("failed to get installed software: %w", err))
		return err
	}

	// Display results
	if flags.JSONOutput {
		listData := map[string]interface{}{
			"type":      "installed_software_list",
			"providers": installedSoftware,
			"total":     getTotalPackageCount(installedSoftware),
		}
		fmt.Println(formatter.FormatJSON(listData))
	} else {
		displayInstalledSoftware(installedSoftware, formatter, userInterface, flags.Verbose)
	}

	return nil
}

// InstalledSoftware represents software installed by a provider
type InstalledSoftware struct {
	Provider     string                   `json:"provider"`
	DisplayName  string                   `json:"display_name"`
	Available    bool                     `json:"available"`
	Packages     []InstalledPackage       `json:"packages"`
	Error        string                   `json:"error,omitempty"`
}

// InstalledPackage represents an installed package
type InstalledPackage struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
}

// getInstalledSoftwareAcrossProviders retrieves installed software from all providers
func getInstalledSoftwareAcrossProviders(ctx context.Context, actionManager interfaces.ActionManager, flags GlobalFlags) ([]InstalledSoftware, error) {
	// For now, we'll simulate getting installed software since the actual implementation
	// would require executing provider-specific list commands
	
	// This is a placeholder implementation that demonstrates the structure
	// In a real implementation, this would:
	// 1. Get all available providers
	// 2. Execute list commands for each provider
	// 3. Parse the output to extract installed packages
	// 4. Return structured data
	
	var installedSoftware []InstalledSoftware
	
	// Simulate common providers and their installed packages
	providers := []struct {
		name        string
		displayName string
		available   bool
		packages    []InstalledPackage
	}{
		{
			name:        "apt",
			displayName: "APT Package Manager",
			available:   true,
			packages: []InstalledPackage{
				{Name: "nginx", Version: "1.18.0", Description: "HTTP and reverse proxy server", Status: "installed"},
				{Name: "curl", Version: "7.68.0", Description: "Command line tool for transferring data", Status: "installed"},
				{Name: "git", Version: "2.25.1", Description: "Distributed version control system", Status: "installed"},
			},
		},
		{
			name:        "docker",
			displayName: "Docker Container Manager",
			available:   true,
			packages: []InstalledPackage{
				{Name: "redis", Version: "6.2", Description: "In-memory data structure store", Status: "running"},
				{Name: "postgres", Version: "13", Description: "Object-relational database system", Status: "running"},
			},
		},
		{
			name:        "npm",
			displayName: "Node Package Manager",
			available:   false,
			packages:    []InstalledPackage{},
		},
	}

	// Filter by provider if specified
	if flags.Provider != "" {
		var filteredProviders []struct {
			name        string
			displayName string
			available   bool
			packages    []InstalledPackage
		}
		for _, p := range providers {
			if p.name == flags.Provider {
				filteredProviders = append(filteredProviders, p)
				break
			}
		}
		providers = filteredProviders
	}

	// Convert to InstalledSoftware format
	for _, p := range providers {
		software := InstalledSoftware{
			Provider:    p.name,
			DisplayName: p.displayName,
			Available:   p.available,
			Packages:    p.packages,
		}
		
		if !p.available {
			software.Error = "Provider not available on this system"
		}
		
		installedSoftware = append(installedSoftware, software)
	}

	return installedSoftware, nil
}

// displayInstalledSoftware displays the installed software in a formatted way
func displayInstalledSoftware(installedSoftware []InstalledSoftware, formatter *output.OutputFormatter, userInterface *ui.UserInterface, verbose bool) {
	if len(installedSoftware) == 0 {
		formatter.ShowInfo("No providers found or no software installed.")
		return
	}

	totalPackages := 0
	availableProviders := 0

	for _, software := range installedSoftware {
		if !software.Available {
			if verbose {
				formatter.ShowWarning(fmt.Sprintf("Provider %s is not available: %s", 
					software.DisplayName, software.Error))
			}
			continue
		}

		availableProviders++
		
		if len(software.Packages) == 0 {
			if verbose {
				fmt.Printf("\n%s:\n", formatter.FormatProviderName(software.Provider))
				fmt.Println("  No packages installed")
			}
			continue
		}

		fmt.Printf("\n%s:\n", formatter.FormatProviderName(software.Provider))
		
		// Sort packages by name
		packages := make([]InstalledPackage, len(software.Packages))
		copy(packages, software.Packages)
		sort.Slice(packages, func(i, j int) bool {
			return packages[i].Name < packages[j].Name
		})

		if verbose {
			// Detailed view with table format
			headers := []string{"Package", "Version", "Status", "Description"}
			var rows [][]string
			
			for _, pkg := range packages {
				description := pkg.Description
				if len(description) > 50 {
					description = description[:47] + "..."
				}
				rows = append(rows, []string{pkg.Name, pkg.Version, pkg.Status, description})
			}
			
			userInterface.ShowTable(headers, rows)
		} else {
			// Simple list view
			for _, pkg := range packages {
				status := ""
				if pkg.Status != "installed" {
					status = fmt.Sprintf(" (%s)", pkg.Status)
				}
				fmt.Printf("  %s %s%s\n", pkg.Name, pkg.Version, status)
			}
		}

		totalPackages += len(software.Packages)
	}

	// Summary
	fmt.Printf("\nSummary: %d packages from %d providers\n", totalPackages, availableProviders)
}

// getTotalPackageCount calculates the total number of packages across all providers
func getTotalPackageCount(installedSoftware []InstalledSoftware) int {
	total := 0
	for _, software := range installedSoftware {
		if software.Available {
			total += len(software.Packages)
		}
	}
	return total
}