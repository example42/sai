package cli

import (
	"context"
	"fmt"
	"os"

	"sai/internal/interfaces"
	"sai/internal/output"
	"sai/internal/types"
)

// executeServiceCommand executes a service management command
func executeServiceCommand(action string, software string) error {
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

	// Validate that the action is supported
	if err := actionManager.ValidateAction(action, software); err != nil {
		formatter.ShowError(fmt.Errorf("action validation failed: %w", err))
		return err
	}

	// Provider selection is now handled by the Action Manager (Requirements 15.1, 15.3, 15.4)
	// The Action Manager will show commands instead of package details for system-changing operations

	// Show progress for system-changing operations
	if !flags.Quiet && config.IsSystemChangingAction(action) {
		if flags.DryRun {
			formatter.ShowProgress(fmt.Sprintf("Dry run: %s %s service...", getActionVerb(action), software))
		} else {
			formatter.ShowProgress(fmt.Sprintf("%s %s service...", getActionVerb(action), software))
		}
	}

	// Execute the service action
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	result, err := actionManager.ExecuteAction(ctx, action, software, options)
	if err != nil {
		formatter.ShowError(fmt.Errorf("%s failed: %w", action, err))
		if result != nil {
			os.Exit(result.ExitCode)
		}
		os.Exit(1)
		return err
	}

	// Handle confirmation for system-changing operations (Requirements 9.1, 9.2)
	if result.RequiredConfirmation && !flags.Yes && !flags.DryRun {
		confirmed, err := userInterface.ConfirmAction(action, software, result.Provider, result.Commands)
		if err != nil {
			formatter.ShowError(fmt.Errorf("confirmation failed: %w", err))
			return err
		}

		if !confirmed {
			formatter.ShowInfo(fmt.Sprintf("%s cancelled by user", getActionVerb(action)))
			return nil
		}

		// Re-execute with confirmation bypassed
		options.Yes = true
		result, err = actionManager.ExecuteAction(ctx, action, software, options)
		if err != nil {
			formatter.ShowError(fmt.Errorf("%s failed: %w", action, err))
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
				formatter.ShowSuccess(fmt.Sprintf("Dry run completed for %s %s", action, software))
			} else {
				formatter.ShowSuccess(fmt.Sprintf("Successfully %s %s service using %s", 
					getActionPastTense(action), software, result.Provider))
			}
		} else {
			formatter.ShowError(fmt.Errorf("failed to %s %s service: %s", action, software, result.Error))
		}

		// Show command output if verbose or for information-only commands
		if (flags.Verbose || config.IsInformationOnlyAction(action)) && result.Output != "" {
			if !config.IsInformationOnlyAction(action) {
				fmt.Println("\nCommand output:")
			}
			fmt.Println(result.Output)
		}
	}

	// Set exit code based on result (Requirement 10.4)
	if !result.Success {
		os.Exit(result.ExitCode)
	}

	return nil
}

// executeGeneralSystemCommand executes a general system command (without software parameter)
func executeGeneralSystemCommand(action string) error {
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

	// Show progress
	if !flags.Quiet {
		formatter.ShowProgress(fmt.Sprintf("Getting system %s information...", action))
	}

	// Execute the general system action (using empty software name to indicate system-wide)
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	result, err := actionManager.ExecuteAction(ctx, action, "", options)
	if err != nil {
		formatter.ShowError(fmt.Errorf("system %s failed: %w", action, err))
		if result != nil {
			os.Exit(result.ExitCode)
		}
		os.Exit(1)
		return err
	}

	// Display results
	if flags.JSONOutput {
		fmt.Println(formatter.FormatJSON(result))
	} else {
		if result.Success {
			// For information-only commands, show the output directly
			if result.Output != "" {
				fmt.Println(result.Output)
			} else {
				formatter.ShowSuccess(fmt.Sprintf("System %s information retrieved", action))
			}
		} else {
			formatter.ShowError(fmt.Errorf("failed to get system %s information: %s", action, result.Error))
		}
	}

	// Set exit code based on result (Requirement 10.4)
	if !result.Success {
		os.Exit(result.ExitCode)
	}

	return nil
}

// Helper functions

// getActionVerb returns the present tense verb for an action
func getActionVerb(action string) string {
	switch action {
	case "start":
		return "Starting"
	case "stop":
		return "Stopping"
	case "restart":
		return "Restarting"
	case "enable":
		return "Enabling"
	case "disable":
		return "Disabling"
	case "status":
		return "Checking status of"
	case "logs":
		return "Getting logs for"
	case "cpu":
		return "Getting CPU usage for"
	case "memory":
		return "Getting memory usage for"
	case "io":
		return "Getting I/O usage for"
	default:
		return fmt.Sprintf("Executing %s on", action)
	}
}

// getActionPastTense returns the past tense verb for an action
func getActionPastTense(action string) string {
	switch action {
	case "start":
		return "started"
	case "stop":
		return "stopped"
	case "restart":
		return "restarted"
	case "enable":
		return "enabled"
	case "disable":
		return "disabled"
	case "status":
		return "checked status of"
	case "logs":
		return "retrieved logs for"
	case "cpu":
		return "retrieved CPU usage for"
	case "memory":
		return "retrieved memory usage for"
	case "io":
		return "retrieved I/O usage for"
	default:
		return fmt.Sprintf("executed %s on", action)
	}
}

// getServiceName extracts the service name from provider data for display
func getServiceName(provider *types.ProviderData, software string) string {
	// TODO: Extract actual service name from saidata
	// For now, return the software name as the service name
	return software
}

