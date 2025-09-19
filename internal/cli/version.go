package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"sai/internal/interfaces"
	"sai/internal/output"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version [software]",
	Short: "Show version information",
	Long: `Show version information for SAI CLI tool or specified software.
When software is specified, shows version information from all providers that provide it,
highlighting whether the package is already installed.

Examples:
  sai version                          # Show SAI CLI version
  sai version nginx                    # Show nginx version info from all providers
  sai version nginx --provider apt     # Show nginx version info from apt only
  sai version nginx --json             # Output version info in JSON format`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// Show SAI version
			fmt.Printf("SAI version %s\n", rootCmd.Version)
			return nil
		}
		return executeVersionCommand(args[0])
	},
}

func executeVersionCommand(software string) error {
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
		formatter.ShowProgress(fmt.Sprintf("Getting version information for %s from all providers...", software))
	}

	// Execute version query across all providers (Requirement 2.5)
	versionResults, err := actionManager.GetSoftwareVersions(software)
	if err != nil {
		formatter.ShowError(fmt.Errorf("version query failed: %w", err))
		return err
	}

	// Filter results by provider if specified
	if flags.Provider != "" {
		var filteredResults []*interfaces.VersionInfo
		for _, result := range versionResults {
			if result.Provider == flags.Provider {
				filteredResults = append(filteredResults, result)
			}
		}
		versionResults = filteredResults
	}

	// Display results
	if flags.JSONOutput {
		fmt.Println(formatter.FormatJSON(map[string]interface{}{
			"software": software,
			"versions": versionResults,
			"count":    len(versionResults),
		}))
	} else {
		if len(versionResults) == 0 {
			formatter.ShowInfo(fmt.Sprintf("No version information found for '%s'", software))
			return nil
		}

		formatter.ShowInfo(fmt.Sprintf("Version information for '%s' from %d provider(s):", software, len(versionResults)))
		fmt.Println()

		// Display results in table format with installation status highlighting (Requirement 2.5)
		headers := []string{"Provider", "Package", "Version", "Latest", "Status"}
		var rows [][]string

		for _, version := range versionResults {
			status := "Not Installed"
			if version.IsInstalled {
				if flags.JSONOutput {
					status = "Installed"
				} else {
					// Highlight installed status in green
					status = "✓ Installed"
				}
			}

			latestVersion := version.LatestVersion
			if latestVersion == "" || latestVersion == "unknown" {
				latestVersion = "-"
			}

			currentVersion := version.Version
			if currentVersion == "" || currentVersion == "unknown" {
				currentVersion = "-"
			}

			row := []string{
				version.Provider,
				version.PackageName,
				currentVersion,
				latestVersion,
				status,
			}
			rows = append(rows, row)
		}

		// Display table
		for i, header := range headers {
			fmt.Printf("%-12s", header)
			if i < len(headers)-1 {
				fmt.Print(" | ")
			}
		}
		fmt.Println()

		for i := 0; i < len(headers); i++ {
			fmt.Print("------------")
			if i < len(headers)-1 {
				fmt.Print("-+-")
			}
		}
		fmt.Println()

		for _, row := range rows {
			for i, cell := range row {
				if i == 0 {
					// Format provider name
					fmt.Printf("%-12s", formatter.FormatProviderName(cell))
				} else if i == 4 && cell == "✓ Installed" && !flags.JSONOutput {
					// Format installed status with green color
					fmt.Printf("\033[32m%-12s\033[0m", cell)
				} else {
					fmt.Printf("%-12s", cell)
				}
				if i < len(row)-1 {
					fmt.Print(" | ")
				}
			}
			fmt.Println()
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(versionCmd)
}