package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"sai/internal/interfaces"
	"sai/internal/output"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [software]",
	Short: "Search for available packages in repositories",
	Long: `Search for available packages in repositories across all available providers.
This command executes without requiring further user confirmation as it only displays information.

The search will query all available providers and display results showing:
- Provider name
- Package name
- Available version
- Description (if available)

Examples:
  sai search nginx                     # Search for nginx across all providers
  sai search nginx --provider apt      # Search for nginx only in apt repositories
  sai search nginx --json              # Output search results in JSON format`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeSearchCommand(args[0])
	},
}

func executeSearchCommand(software string) error {
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
		formatter.ShowProgress(fmt.Sprintf("Searching for %s across all providers...", software))
	}

	// Execute search across all providers (Requirement 2.3)
	searchResults, err := actionManager.SearchAcrossProviders(software)
	if err != nil {
		formatter.ShowError(fmt.Errorf("search failed: %w", err))
		return err
	}

	// Filter results by provider if specified
	if flags.Provider != "" {
		var filteredResults []*interfaces.SearchResult
		for _, result := range searchResults {
			if result.Provider == flags.Provider {
				filteredResults = append(filteredResults, result)
			}
		}
		searchResults = filteredResults
	}

	// Display results
	if flags.JSONOutput {
		fmt.Println(formatter.FormatJSON(map[string]interface{}{
			"software": software,
			"results":  searchResults,
			"count":    len(searchResults),
		}))
	} else {
		if len(searchResults) == 0 {
			formatter.ShowInfo(fmt.Sprintf("No packages found for '%s'", software))
			return nil
		}

		formatter.ShowInfo(fmt.Sprintf("Found %d package(s) for '%s':", len(searchResults), software))
		fmt.Println()

		// Display results in table format
		headers := []string{"Provider", "Package", "Version", "Available", "Description"}
		var rows [][]string

		for _, result := range searchResults {
			availability := "Yes"
			if !result.Available {
				availability = "No"
			}

			row := []string{
				formatter.FormatProviderName(result.Provider),
				result.PackageName,
				result.Version,
				availability,
				result.Description,
			}
			rows = append(rows, row)
		}

		// Use a simple table display for now
		for i, header := range headers {
			fmt.Printf("%-15s", header)
			if i < len(headers)-1 {
				fmt.Print(" | ")
			}
		}
		fmt.Println()

		for i := 0; i < len(headers); i++ {
			fmt.Print("---------------")
			if i < len(headers)-1 {
				fmt.Print("-+-")
			}
		}
		fmt.Println()

		for _, row := range rows {
			for i, cell := range row {
				fmt.Printf("%-15s", cell)
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
	rootCmd.AddCommand(searchCmd)
}