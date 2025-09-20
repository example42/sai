package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"sai/internal/interfaces"
	"sai/internal/output"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall [software]",
	Short: "Uninstall software packages",
	Long: `Uninstall software packages using the appropriate provider.
If software is installed using different providers, SAI will provide a list for user selection.

Examples:
  sai uninstall nginx                    # Uninstall nginx using detected provider
  sai uninstall nginx --provider apt     # Uninstall nginx using apt provider
  sai uninstall nginx --yes              # Uninstall nginx without confirmation prompts
  sai uninstall nginx --dry-run          # Show what would be executed without uninstalling`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeUninstallCommand(args[0])
	},
}

func executeUninstallCommand(software string) error {
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

	// Prepare action options
	options := interfaces.ActionOptions{
		Provider:  flags.Provider,
		DryRun:    flags.DryRun,
		Verbose:   flags.Verbose,
		Quiet:     flags.Quiet,
		Yes:       flags.Yes,
		JSON:      flags.JSONOutput,
		Config:    flags.Config,
		Variables: make(map[string]string),
		Timeout:   config.Timeout,
	}

	// Provider selection is now handled by the Action Manager (Requirements 15.1, 15.3, 15.4)
	// The Action Manager will show commands instead of package details for system-changing operations

	// Show progress
	if !flags.Quiet {
		if flags.DryRun {
			formatter.ShowProgress(fmt.Sprintf("Dry run: Uninstalling %s...", software))
		} else {
			formatter.ShowProgress(fmt.Sprintf("Uninstalling %s...", software))
		}
	}

	// Execute the uninstall action
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	result, err := actionManager.ExecuteAction(ctx, "uninstall", software, options)
	if err != nil {
		formatter.ShowError(fmt.Errorf("uninstallation failed: %w", err))
		os.Exit(result.ExitCode)
		return err
	}

	// Handle confirmation for system-changing operations (Requirements 9.1, 9.2)
	if result.RequiredConfirmation && !flags.Yes && !flags.DryRun {
		confirmed, err := userInterface.ConfirmAction("uninstall", software, result.Provider, result.Commands)
		if err != nil {
			formatter.ShowError(fmt.Errorf("confirmation failed: %w", err))
			return err
		}

		if !confirmed {
			formatter.ShowInfo("Uninstallation cancelled by user")
			return nil
		}

		// Re-execute with confirmation bypassed
		options.Yes = true
		result, err = actionManager.ExecuteAction(ctx, "uninstall", software, options)
		if err != nil {
			formatter.ShowError(fmt.Errorf("uninstallation failed: %w", err))
			os.Exit(result.ExitCode)
			return err
		}
	}

	// Display results
	if flags.JSONOutput {
		fmt.Println(formatter.FormatJSON(result))
	} else {
		if result.Success {
			if flags.DryRun {
				formatter.ShowSuccess(fmt.Sprintf("Dry run completed for %s", software))
			} else {
				formatter.ShowSuccess(fmt.Sprintf("Successfully uninstalled %s using %s", software, result.Provider))
			}
		} else {
			formatter.ShowError(fmt.Errorf("failed to uninstall %s: %s", software, result.Error))
		}

		// Show command output if verbose
		if flags.Verbose && result.Output != "" {
			fmt.Println("\nCommand output:")
			fmt.Println(result.Output)
		}
	}

	// Set exit code based on result (Requirement 10.4)
	if !result.Success {
		os.Exit(result.ExitCode)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}