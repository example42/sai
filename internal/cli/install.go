package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"sai/internal/action"
	"sai/internal/config"
	"sai/internal/interfaces"
	"sai/internal/output"
	"sai/internal/provider"
	"sai/internal/saidata"
	"sai/internal/executor"
	"sai/internal/validation"
	"sai/internal/ui"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [software]",
	Short: "Install software packages",
	Long: `Install software packages using the appropriate provider.
The system will detect the best provider automatically or use the one specified with --provider.

When multiple providers are available, SAI will display options and prompt for selection
unless --provider is specified or --yes flag is used (which selects the highest priority provider).

Examples:
  sai install nginx                    # Install nginx using best available provider
  sai install nginx --provider apt     # Install nginx using apt provider
  sai install nginx --yes              # Install nginx without confirmation prompts
  sai install nginx --dry-run          # Show what would be executed without installing`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeInstallCommand(args[0])
	},
}

func executeInstallCommand(software string) error {
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

	// Handle provider selection if multiple providers are available
	if flags.Provider == "" && !flags.Yes {
		providerOptions, err := actionManager.GetAvailableProviders(software, "install")
		if err != nil {
			formatter.ShowError(fmt.Errorf("failed to get available providers: %w", err))
			return err
		}

		// If multiple providers available, show selection (Requirement 1.3)
		if len(providerOptions) > 1 {
			uiOptions := make([]*ui.ProviderOption, len(providerOptions))
			for i, option := range providerOptions {
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
		}
	}

	// Show progress
	if !flags.Quiet {
		if flags.DryRun {
			formatter.ShowProgress(fmt.Sprintf("Dry run: Installing %s...", software))
		} else {
			formatter.ShowProgress(fmt.Sprintf("Installing %s...", software))
		}
	}

	// Execute the install action
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	result, err := actionManager.ExecuteAction(ctx, "install", software, options)
	if err != nil {
		formatter.ShowError(fmt.Errorf("installation failed: %w", err))
		os.Exit(result.ExitCode)
		return err
	}

	// Handle confirmation for system-changing operations (Requirements 9.1, 9.2)
	if result.RequiredConfirmation && !flags.Yes && !flags.DryRun {
		confirmed, err := userInterface.ConfirmAction("install", software, result.Provider, result.Commands)
		if err != nil {
			formatter.ShowError(fmt.Errorf("confirmation failed: %w", err))
			return err
		}

		if !confirmed {
			formatter.ShowInfo("Installation cancelled by user")
			return nil
		}

		// Re-execute with confirmation bypassed
		options.Yes = true
		result, err = actionManager.ExecuteAction(ctx, "install", software, options)
		if err != nil {
			formatter.ShowError(fmt.Errorf("installation failed: %w", err))
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
				formatter.ShowSuccess(fmt.Sprintf("Successfully installed %s using %s", software, result.Provider))
			}
		} else {
			formatter.ShowError(fmt.Errorf("failed to install %s: %s", software, result.Error))
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

// createManagers creates and initializes all required managers
func createManagers(cfg *config.Config, formatter *output.OutputFormatter) (interfaces.ActionManager, *ui.UserInterface, error) {
	// Create provider manager
	providerConfig := &provider.ManagerConfig{
		ProviderDirectory: "providers",
		SchemaPath:        "schemas/providerdata-0.1-schema.json",
		DefaultProvider:   cfg.DefaultProvider,
		ProviderPriority:  cfg.ProviderPriority,
		EnableWatching:    false,
	}

	providerManager, err := provider.NewProviderManager(providerConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create provider manager: %w", err)
	}

	// Create saidata manager
	saidataManager := saidata.NewManager("docs/saidata_samples")

	// Create logger (placeholder - would need actual implementation)
	logger := &MockLogger{}

	// Create validator
	resourceValidator := validation.NewResourceValidator()

	// Create command executor
	commandExecutor := executor.NewCommandExecutor(logger, resourceValidator)

	// Create template engine (placeholder - would need actual implementation)
	templateEngine := &MockTemplateEngine{}

	// Create generic executor
	genericExecutor := executor.NewGenericExecutor(
		commandExecutor,
		templateEngine,
		logger,
		resourceValidator,
	)

	// Create action manager
	actionManager := action.NewActionManager(
		providerManager,
		saidataManager,
		genericExecutor,
		resourceValidator,
		cfg,
	)

	// Create user interface
	userInterface := ui.NewUserInterface(cfg, formatter)

	return actionManager, userInterface, nil
}

func init() {
	rootCmd.AddCommand(installCmd)
}