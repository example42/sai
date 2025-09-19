package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"sai/internal/interfaces"
	"sai/internal/output"
	"sai/internal/ui"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade [software]",
	Short: "Upgrade software packages to latest version",
	Long: `Upgrade software packages to their latest version using the appropriate provider.
The system will detect which provider was used to install the software and use that for upgrading.

Examples:
  sai upgrade nginx                    # Upgrade nginx using detected provider
  sai upgrade nginx --provider apt     # Upgrade nginx using apt provider
  sai upgrade nginx --yes              # Upgrade nginx without confirmation prompts
  sai upgrade nginx --dry-run          # Show what would be executed without upgrading`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeUpgradeCommand(args[0])
	},
}

func executeUpgradeCommand(software string) error {
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

	// Handle provider selection for installed software (Requirement 2.2)
	if flags.Provider == "" && !flags.Yes {
		providerOptions, err := actionManager.GetAvailableProviders(software, "upgrade")
		if err != nil {
			formatter.ShowError(fmt.Errorf("failed to get available providers: %w", err))
			return err
		}

		// Filter to only show providers where software is installed
		var installedOptions []*interfaces.ProviderOption
		for _, option := range providerOptions {
			if option.IsInstalled {
				installedOptions = append(installedOptions, option)
			}
		}

		// If multiple providers have the software installed, show selection
		if len(installedOptions) > 1 {
			uiOptions := make([]*ui.ProviderOption, len(installedOptions))
			for i, option := range installedOptions {
				uiOptions[i] = &ui.ProviderOption{
					Name:        option.Provider.Provider.Name,
					PackageName: option.PackageName,
					Version:     option.Version,
					IsInstalled: option.IsInstalled,
					Description: option.Provider.Provider.Description,
				}
			}

			selectedOption, err := userInterface.ShowProviderSelection(software, uiOptions)
			if err != nil {
				formatter.ShowError(fmt.Errorf("provider selection failed: %w", err))
				return err
			}

			options.Provider = selectedOption.Name
		} else if len(installedOptions) == 1 {
			options.Provider = installedOptions[0].Provider.Provider.Name
		} else {
			formatter.ShowError(fmt.Errorf("software %s does not appear to be installed", software))
			return fmt.Errorf("software not installed")
		}
	}

	// Show progress
	if !flags.Quiet {
		if flags.DryRun {
			formatter.ShowProgress(fmt.Sprintf("Dry run: Upgrading %s...", software))
		} else {
			formatter.ShowProgress(fmt.Sprintf("Upgrading %s...", software))
		}
	}

	// Execute the upgrade action
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	result, err := actionManager.ExecuteAction(ctx, "upgrade", software, options)
	if err != nil {
		formatter.ShowError(fmt.Errorf("upgrade failed: %w", err))
		os.Exit(result.ExitCode)
		return err
	}

	// Handle confirmation for system-changing operations (Requirements 9.1, 9.2)
	if result.RequiredConfirmation && !flags.Yes && !flags.DryRun {
		confirmed, err := userInterface.ConfirmAction("upgrade", software, result.Provider, result.Commands)
		if err != nil {
			formatter.ShowError(fmt.Errorf("confirmation failed: %w", err))
			return err
		}

		if !confirmed {
			formatter.ShowInfo("Upgrade cancelled by user")
			return nil
		}

		// Re-execute with confirmation bypassed
		options.Yes = true
		result, err = actionManager.ExecuteAction(ctx, "upgrade", software, options)
		if err != nil {
			formatter.ShowError(fmt.Errorf("upgrade failed: %w", err))
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
				formatter.ShowSuccess(fmt.Sprintf("Successfully upgraded %s using %s", software, result.Provider))
			}
		} else {
			formatter.ShowError(fmt.Errorf("failed to upgrade %s: %s", software, result.Error))
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
	rootCmd.AddCommand(upgradeCmd)
}