package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"sai/internal/interfaces"
	"sai/internal/output"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info [software]",
	Short: "Display detailed information about software",
	Long: `Display detailed information about software from all available providers.
This command executes without requiring further user confirmation as it only displays information.

The info command will query all available providers and display:
- Package name and version
- Description
- Homepage URL
- License information
- Dependencies (if available)

Examples:
  sai info nginx                       # Get info about nginx from all providers
  sai info nginx --provider apt        # Get info about nginx only from apt
  sai info nginx --json                # Output info in JSON format`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeInfoCommand(args[0])
	},
}

func executeInfoCommand(software string) error {
	// Get global configuration and flags
	config := GetGlobalConfig()
	flags := GetGlobalFlags()

	// Create output formatter
	formatter := output.NewOutputFormatter(config, flags.Verbose, flags.Quiet, flags.JSONOutput)

	// Create managers and dependencies
	actionManager, _, err := createManagers(config, formatter)
	if err != nil {
		formatter.ShowError(fmt.Errorf("failed to initialize managers: %w", err))
		return err
	}

	// Show progress
	if !flags.Quiet {
		formatter.ShowProgress(fmt.Sprintf("Getting information for %s from all providers...", software))
	}

	// Execute info across all providers (Requirement 2.4)
	infoResults, err := actionManager.GetSoftwareInfo(software)
	if err != nil {
		formatter.ShowError(fmt.Errorf("info query failed: %w", err))
		return err
	}

	// Filter results by provider if specified
	if flags.Provider != "" {
		var filteredResults []*interfaces.SoftwareInfo
		for _, result := range infoResults {
			if result.Provider == flags.Provider {
				filteredResults = append(filteredResults, result)
			}
		}
		infoResults = filteredResults
	}

	// Display results
	if flags.JSONOutput {
		fmt.Println(formatter.FormatJSON(map[string]interface{}{
			"software": software,
			"info":     infoResults,
			"count":    len(infoResults),
		}))
	} else {
		if len(infoResults) == 0 {
			formatter.ShowInfo(fmt.Sprintf("No information found for '%s'", software))
			return nil
		}

		formatter.ShowInfo(fmt.Sprintf("Information for '%s' from %d provider(s):", software, len(infoResults)))
		fmt.Println()

		for i, info := range infoResults {
			if i > 0 {
				fmt.Println("---")
			}

			fmt.Printf("Provider: %s\n", formatter.FormatProviderName(info.Provider))
			fmt.Printf("Package:  %s\n", info.PackageName)
			
			if info.Version != "" && info.Version != "unknown" {
				fmt.Printf("Version:  %s\n", info.Version)
			}
			
			if info.Description != "" {
				fmt.Printf("Description: %s\n", info.Description)
			}
			
			if info.Homepage != "" {
				fmt.Printf("Homepage: %s\n", info.Homepage)
			}
			
			if info.License != "" && info.License != "unknown" {
				fmt.Printf("License:  %s\n", info.License)
			}
			
			if len(info.Dependencies) > 0 {
				fmt.Printf("Dependencies: %v\n", info.Dependencies)
			}
			
			fmt.Println()
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(infoCmd)
}