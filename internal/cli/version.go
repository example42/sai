package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"sai/internal/interfaces"
	"sai/internal/output"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version <software>",
	Short: "Show software version information",
	Long: `Show version information for specified software from all available providers.
This command shows version information from all providers that provide the software,
highlighting whether the package is already installed.

Note: To show SAI CLI version, use 'sai --version' instead.

Examples:
  sai version nginx                    # Show nginx version info from all providers
  sai version nginx --provider apt     # Show nginx version info from apt only
  sai version nginx --json             # Output version info in JSON format`,
	Args: cobra.ExactArgs(1), // Require exactly one argument (software name)
	RunE: func(cmd *cobra.Command, args []string) error {
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
			formatter.ShowError(fmt.Errorf("no version information found for '%s' - no providers support this software or version action", software))
			return fmt.Errorf("no version information available")
		}

		formatter.ShowInfo(fmt.Sprintf("Version information for '%s' from %d provider(s):", software, len(versionResults)))
		fmt.Println()

		// Display results in table format with installation status highlighting (Requirement 2.5)
		headers := []string{"Provider", "Package", "Version", "Status"}
		var rows [][]string

		for _, version := range versionResults {
			// Determine status with color coding
			var status string
			if version.IsInstalled {
				status = "✓ Installed"
			} else if version.Version == "Available" {
				status = "Available"
			} else if version.Version == "Not Available" {
				status = "Not Available"
			} else if version.Version == "Error" {
				status = "⚠ Error"
			} else {
				status = "Not Installed"
			}

			// Handle version display
			currentVersion := version.Version
			if currentVersion == "" || currentVersion == "unknown" {
				currentVersion = "-"
			}

			row := []string{
				version.Provider,
				version.PackageName,
				currentVersion,
				status,
			}
			rows = append(rows, row)
		}

		// Display table with proper formatting
		colWidths := []int{12, 20, 25, 15}
		
		// Print headers
		for i, header := range headers {
			fmt.Printf("%-*s", colWidths[i], header)
			if i < len(headers)-1 {
				fmt.Print(" | ")
			}
		}
		fmt.Println()

		// Print separator
		for i := 0; i < len(headers); i++ {
			fmt.Print(strings.Repeat("-", colWidths[i]))
			if i < len(headers)-1 {
				fmt.Print("-+-")
			}
		}
		fmt.Println()

		// Print rows with color formatting
		for _, row := range rows {
			for i, cell := range row {
				if i == 0 {
					// Format provider name with color
					fmt.Printf("%-*s", colWidths[i], formatter.FormatProviderName(cell))
				} else if i == 3 {
					// Format status with appropriate color
					switch {
					case strings.Contains(cell, "✓ Installed"):
						fmt.Printf("\033[32m%-*s\033[0m", colWidths[i], cell) // Green
					case strings.Contains(cell, "Available"):
						fmt.Printf("\033[33m%-*s\033[0m", colWidths[i], cell) // Yellow
					case strings.Contains(cell, "⚠ Error"):
						fmt.Printf("\033[31m%-*s\033[0m", colWidths[i], cell) // Red
					case strings.Contains(cell, "Not Available"):
						fmt.Printf("\033[90m%-*s\033[0m", colWidths[i], cell) // Gray
					default:
						fmt.Printf("%-*s", colWidths[i], cell)
					}
				} else {
					// Truncate long text if needed
					displayCell := cell
					if len(cell) > colWidths[i]-1 {
						displayCell = cell[:colWidths[i]-4] + "..."
					}
					fmt.Printf("%-*s", colWidths[i], displayCell)
				}
				if i < len(row)-1 {
					fmt.Print(" | ")
				}
			}
			fmt.Println()
		}
		
		// Show summary
		installedCount := 0
		availableCount := 0
		errorCount := 0
		
		for _, version := range versionResults {
			if version.IsInstalled {
				installedCount++
			} else if version.Version == "Available" {
				availableCount++
			} else if version.Version == "Error" {
				errorCount++
			}
		}
		
		fmt.Println()
		if installedCount > 0 {
			fmt.Printf("✓ Installed in %d provider(s)", installedCount)
			if availableCount > 0 || errorCount > 0 {
				fmt.Print(", ")
			}
		}
		if availableCount > 0 {
			fmt.Printf("Available in %d provider(s)", availableCount)
			if errorCount > 0 {
				fmt.Print(", ")
			}
		}
		if errorCount > 0 {
			fmt.Printf("⚠ %d provider(s) had errors", errorCount)
		}
		fmt.Println()
	}

	return nil
}

func init() {
	rootCmd.AddCommand(versionCmd)
}