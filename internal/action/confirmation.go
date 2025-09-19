package action

import (
	"fmt"
	"strings"

	"sai/internal/config"
	"sai/internal/interfaces"
	"sai/internal/output"
	"sai/internal/ui"
)

// ConfirmationManager handles user confirmations for different types of operations
type ConfirmationManager struct {
	config    *config.Config
	ui        *ui.UserInterface
	formatter *output.OutputFormatter
}

// NewConfirmationManager creates a new confirmation manager
func NewConfirmationManager(cfg *config.Config, userInterface *ui.UserInterface, formatter *output.OutputFormatter) *ConfirmationManager {
	return &ConfirmationManager{
		config:    cfg,
		ui:        userInterface,
		formatter: formatter,
	}
}

// RequiresConfirmation determines if an action requires user confirmation (Requirements 9.1, 9.2)
func (cm *ConfirmationManager) RequiresConfirmation(action string, options interfaces.ActionOptions) bool {
	// Skip confirmation if --yes flag is provided (Requirement 9.3)
	if options.Yes {
		return false
	}

	// Skip confirmation for information-only commands (Requirement 9.2)
	if cm.config.IsInformationOnlyAction(action) {
		return false
	}

	// Require confirmation for system-changing operations (Requirement 9.1)
	return cm.config.RequiresConfirmation(action)
}

// ConfirmAction prompts the user for confirmation with detailed information
func (cm *ConfirmationManager) ConfirmAction(action, software, provider string, commands []string, safetyResult *SafetyResult) (bool, error) {
	// Show safety warnings first
	if safetyResult != nil {
		cm.showSafetyWarnings(safetyResult)
	}

	// Show command preview
	if len(commands) > 0 && cm.config.Output.ShowCommands {
		cm.formatter.ShowInfo("Commands that will be executed:")
		for _, cmd := range commands {
			fmt.Printf("  %s\n", cm.formatter.FormatCommand(cmd, provider))
		}
		fmt.Println()
	}

	// Handle different confirmation scenarios
	return cm.ui.ConfirmAction(action, software, provider, commands)
}

// ConfirmProviderSelection handles provider selection when multiple providers are available (Requirement 1.3)
func (cm *ConfirmationManager) ConfirmProviderSelection(software string, options []*interfaces.ProviderOption) (*interfaces.ProviderOption, error) {
	if len(options) == 0 {
		return nil, fmt.Errorf("no providers available for %s", software)
	}

	if len(options) == 1 {
		return options[0], nil
	}

	// Convert to UI provider options
	uiOptions := make([]*ui.ProviderOption, len(options))
	for i, option := range options {
		uiOptions[i] = &ui.ProviderOption{
			Name:        option.Provider.Provider.Name,
			PackageName: option.PackageName,
			Version:     option.Version,
			IsInstalled: option.IsInstalled,
			Description: option.Provider.Provider.Description,
		}
	}

	selectedOption, err := cm.ui.ShowProviderSelection(software, uiOptions)
	if err != nil {
		return nil, fmt.Errorf("provider selection failed: %w", err)
	}

	// Find the corresponding provider option
	for _, option := range options {
		if option.Provider.Provider.Name == selectedOption.Name {
			return option, nil
		}
	}

	return nil, fmt.Errorf("selected provider not found")
}

// ConfirmDestructiveAction provides extra confirmation for destructive operations
func (cm *ConfirmationManager) ConfirmDestructiveAction(action, software string, safetyResult *SafetyResult) (bool, error) {
	destructiveActions := []string{"uninstall", "stop", "disable"}
	
	isDestructive := false
	for _, destructive := range destructiveActions {
		if action == destructive {
			isDestructive = true
			break
		}
	}

	if !isDestructive {
		return true, nil
	}

	// Show extra warnings for destructive actions
	cm.formatter.ShowWarning(fmt.Sprintf("This is a destructive operation that will %s %s", action, software))
	
	if safetyResult != nil {
		errors := safetyResult.GetErrors()
		if len(errors) > 0 {
			cm.formatter.ShowError(fmt.Errorf("safety checks failed: %v", errors))
			return false, fmt.Errorf("destructive action blocked by safety checks")
		}
	}

	// Require explicit confirmation
	message := fmt.Sprintf("Are you sure you want to %s %s? This action cannot be easily undone", action, software)
	return cm.ui.PromptForConfirmation(message)
}

// showSafetyWarnings displays safety check results to the user
func (cm *ConfirmationManager) showSafetyWarnings(safetyResult *SafetyResult) {
	if !safetyResult.Safe {
		cm.formatter.ShowError(fmt.Errorf("safety checks failed for %s on %s", safetyResult.Action, safetyResult.Software))
		
		failedChecks := safetyResult.GetFailedChecks()
		for _, check := range failedChecks {
			cm.formatter.ShowError(fmt.Errorf("%s: %s", check.Name, strings.Join(check.Messages, "; ")))
		}
	}

	// Show warnings even if overall safety check passed
	warnings := safetyResult.GetWarnings()
	for _, warning := range warnings {
		cm.formatter.ShowWarning(warning)
	}
}

// buildConfirmationMessage creates an appropriate confirmation message
func (cm *ConfirmationManager) buildConfirmationMessage(action, software, provider string) string {
	switch action {
	case "install":
		return fmt.Sprintf("Install %s using %s?", software, provider)
	case "uninstall":
		return fmt.Sprintf("Uninstall %s using %s? This will remove the software from your system", software, provider)
	case "upgrade":
		return fmt.Sprintf("Upgrade %s using %s?", software, provider)
	case "start":
		return fmt.Sprintf("Start %s service?", software)
	case "stop":
		return fmt.Sprintf("Stop %s service? This will terminate the running service", software)
	case "restart":
		return fmt.Sprintf("Restart %s service?", software)
	case "enable":
		return fmt.Sprintf("Enable %s service to start automatically at boot?", software)
	case "disable":
		return fmt.Sprintf("Disable %s service from starting automatically at boot?", software)
	default:
		return fmt.Sprintf("Execute %s for %s using %s?", action, software, provider)
	}
}

// BypassConfirmation checks if confirmation should be bypassed (Requirement 9.3, 9.4)
func (cm *ConfirmationManager) BypassConfirmation(options interfaces.ActionOptions) bool {
	// Bypass if --yes flag is provided (Requirement 9.3)
	if options.Yes {
		return true
	}

	// Bypass if in quiet mode and it's an information-only command (Requirement 9.4)
	if options.Quiet {
		// This would need the action to determine if it's information-only
		// For now, we'll let the RequiresConfirmation method handle this
		return false
	}

	return false
}

// ValidateUserChoice validates user input for provider selection
func (cm *ConfirmationManager) ValidateUserChoice(choice string, maxOptions int) (int, error) {
	if choice == "" {
		return 0, fmt.Errorf("no selection made")
	}

	// Handle numeric choices
	if len(choice) == 1 && choice[0] >= '1' && choice[0] <= '9' {
		choiceNum := int(choice[0] - '0')
		if choiceNum > 0 && choiceNum <= maxOptions {
			return choiceNum - 1, nil // Convert to 0-based index
		}
	}

	return 0, fmt.Errorf("invalid choice: %s. Please enter a number between 1 and %d", choice, maxOptions)
}

// ShowProviderDetails displays detailed information about provider options
func (cm *ConfirmationManager) ShowProviderDetails(options []*interfaces.ProviderOption, software string) {
	if cm.formatter.IsJSONMode() {
		// In JSON mode, output structured data
		data := map[string]interface{}{
			"type":     "provider_selection",
			"software": software,
			"options":  options,
		}
		fmt.Println(cm.formatter.FormatJSON(data))
		return
	}

	cm.formatter.ShowInfo(fmt.Sprintf("Multiple providers available for %s:", software))
	fmt.Println()

	for i, option := range options {
		status := "Available"
		if option.IsInstalled {
			status = "Installed"
		}

		fmt.Printf("%d. %s\n", i+1, cm.formatter.FormatProviderName(option.Provider.Provider.Name))
		fmt.Printf("   Package: %s\n", option.PackageName)
		if option.Version != "" {
			fmt.Printf("   Version: %s\n", option.Version)
		}
		fmt.Printf("   Status:  %s\n", status)
		if option.Provider.Provider.Description != "" {
			fmt.Printf("   Description: %s\n", option.Provider.Provider.Description)
		}
		fmt.Printf("   Priority: %d\n", option.Priority)
		fmt.Println()
	}
}

// HandleNonInteractiveMode handles confirmations in non-interactive environments
func (cm *ConfirmationManager) HandleNonInteractiveMode(action, software, provider string) error {
	if cm.formatter.IsJSONMode() {
		// Output structured error for non-interactive mode
		errorData := map[string]interface{}{
			"error":    "confirmation_required",
			"action":   action,
			"software": software,
			"provider": provider,
			"message":  "Use --yes flag to skip confirmation prompts in non-interactive mode",
		}
		fmt.Println(cm.formatter.FormatJSON(errorData))
	} else {
		cm.formatter.ShowError(fmt.Errorf("confirmation required for %s %s. Use --yes flag to skip prompts", action, software))
	}
	
	return fmt.Errorf("confirmation required in non-interactive mode")
}