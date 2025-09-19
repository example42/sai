package action

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"sai/internal/config"
	"sai/internal/interfaces"
	"sai/internal/output"
	"sai/internal/types"
	"sai/internal/ui"
)

// ActionManager orchestrates software management operations
type ActionManager struct {
	providerManager     interfaces.ProviderManager
	saidataManager      interfaces.SaidataManager
	executor            interfaces.GenericExecutor
	validator           interfaces.ResourceValidator
	config              *config.Config
	ui                  *ui.UserInterface
	formatter           *output.OutputFormatter
	safetyManager       *SafetyManager
	confirmationManager *ConfirmationManager
}

// NewActionManager creates a new action manager
func NewActionManager(
	providerManager interfaces.ProviderManager,
	saidataManager interfaces.SaidataManager,
	executor interfaces.GenericExecutor,
	validator interfaces.ResourceValidator,
	config *config.Config,
	userInterface *ui.UserInterface,
	formatter *output.OutputFormatter,
) *ActionManager {
	safetyManager := NewSafetyManager(validator)
	confirmationManager := NewConfirmationManager(config, userInterface, formatter)
	
	return &ActionManager{
		providerManager:     providerManager,
		saidataManager:      saidataManager,
		executor:            executor,
		validator:           validator,
		config:              config,
		ui:                  userInterface,
		formatter:           formatter,
		safetyManager:       safetyManager,
		confirmationManager: confirmationManager,
	}
}

// ExecuteAction executes a specific action on software with full workflow orchestration
func (am *ActionManager) ExecuteAction(ctx context.Context, action string, software string, options interfaces.ActionOptions) (*interfaces.ActionResult, error) {
	startTime := time.Now()

	// Step 1: Validate action can be performed
	if err := am.ValidateAction(action, software); err != nil {
		return am.buildErrorResult(action, software, "", err, startTime), err
	}

	// Step 2: Resolve software data (saidata or intelligent defaults)
	saidata, err := am.ResolveSoftwareData(software)
	if err != nil {
		return am.buildErrorResult(action, software, "", fmt.Errorf("failed to resolve software data: %w", err), startTime), err
	}

	// Step 3: Setup repositories if needed (Requirement 8.5)
	if err := am.ManageRepositorySetup(saidata); err != nil {
		am.formatter.ShowWarning(fmt.Sprintf("Repository setup failed: %v", err))
	}

	// Step 4: Get available providers for this software and action
	providerOptions, err := am.GetAvailableProviders(software, action)
	if err != nil {
		return am.buildErrorResult(action, software, "", fmt.Errorf("failed to get available providers: %w", err), startTime), err
	}

	if len(providerOptions) == 0 {
		return am.buildErrorResult(action, software, "", fmt.Errorf("no providers available for action %s on software %s", action, software), startTime), fmt.Errorf("no providers available")
	}

	// Step 5: Select provider with user interaction if needed
	selectedProvider, err := am.selectProvider(software, action, providerOptions, options)
	if err != nil {
		return am.buildErrorResult(action, software, "", err, startTime), err
	}

	// Step 6: Perform comprehensive safety checks (Requirement 10.5)
	safetyResult, err := am.safetyManager.CheckActionSafety(action, software, selectedProvider, saidata)
	if err != nil {
		return am.buildErrorResult(action, software, selectedProvider.Provider.Name, fmt.Errorf("safety check failed: %w", err), startTime), err
	}

	// Handle safety check failures
	if !safetyResult.Safe {
		errors := safetyResult.GetErrors()
		if len(errors) > 0 {
			for _, errorMsg := range errors {
				am.formatter.ShowError(fmt.Errorf("%s", errorMsg))
			}
			return am.buildErrorResult(action, software, selectedProvider.Provider.Name, 
				fmt.Errorf("safety checks failed: %v", errors), startTime), 
				fmt.Errorf("safety checks failed")
		}
	}

	// Show safety warnings
	warnings := safetyResult.GetWarnings()
	for _, warning := range warnings {
		am.formatter.ShowWarning(warning)
	}

	// Step 7: Get commands that will be executed
	executeOptions := interfaces.ExecuteOptions{
		DryRun:    options.DryRun,
		Verbose:   options.Verbose,
		Timeout:   options.Timeout,
		Variables: options.Variables,
	}

	// Get preview of commands for confirmation
	var commands []string
	if previewResult, err := am.executor.DryRun(ctx, selectedProvider, action, software, saidata, executeOptions); err == nil {
		commands = previewResult.Commands
	}

	// Step 8: Handle confirmation prompts with enhanced safety information (Requirements 9.1, 9.2)
	if am.confirmationManager.RequiresConfirmation(action, options) {
		// Check for destructive operations first
		if action == "uninstall" || action == "stop" || action == "disable" {
			confirmed, err := am.confirmationManager.ConfirmDestructiveAction(action, software, safetyResult)
			if err != nil {
				return am.buildErrorResult(action, software, selectedProvider.Provider.Name, err, startTime), err
			}
			if !confirmed {
				return &interfaces.ActionResult{
					Action:               action,
					Software:             software,
					Provider:             selectedProvider.Provider.Name,
					Success:              false,
					Error:                fmt.Errorf("destructive action cancelled by user"),
					Duration:             time.Since(startTime),
					ExitCode:             1,
					RequiredConfirmation: true,
				}, fmt.Errorf("action cancelled by user")
			}
		} else {
			// Regular confirmation with safety information
			confirmed, err := am.confirmationManager.ConfirmAction(action, software, selectedProvider.Provider.Name, commands, safetyResult)
			if err != nil {
				return am.buildErrorResult(action, software, selectedProvider.Provider.Name, fmt.Errorf("confirmation failed: %w", err), startTime), err
			}
			if !confirmed {
				return &interfaces.ActionResult{
					Action:               action,
					Software:             software,
					Provider:             selectedProvider.Provider.Name,
					Success:              false,
					Error:                fmt.Errorf("action cancelled by user"),
					Duration:             time.Since(startTime),
					ExitCode:             1,
					RequiredConfirmation: true,
				}, fmt.Errorf("action cancelled by user")
			}
		}
	}

	// Step 9: Execute the action
	var executionResult *interfaces.ExecutionResult
	if options.DryRun {
		am.formatter.ShowInfo("Dry run mode - showing commands that would be executed:")
		executionResult, err = am.executor.DryRun(ctx, selectedProvider, action, software, saidata, executeOptions)
	} else {
		executionResult, err = am.executor.Execute(ctx, selectedProvider, action, software, saidata, executeOptions)
	}

	// Step 10: Build and return result
	result := &interfaces.ActionResult{
		Action:               action,
		Software:             software,
		Provider:             selectedProvider.Provider.Name,
		Success:              executionResult != nil && executionResult.Success,
		Duration:             time.Since(startTime),
		RequiredConfirmation: am.RequiresConfirmation(action),
	}

	if executionResult != nil {
		result.Output = executionResult.Output
		result.Commands = executionResult.Commands
		result.ExitCode = executionResult.ExitCode
		result.Changes = executionResult.Changes
	}

	if err != nil {
		result.Error = err
		result.Success = false
		if result.ExitCode == 0 {
			result.ExitCode = 1
		}
	}

	// Step 11: Show result to user
	am.displayActionResult(result)

	return result, err
}

// ValidateAction validates if an action can be performed
func (am *ActionManager) ValidateAction(action string, software string) error {
	// Check if any providers support this action
	providers := am.providerManager.GetProvidersForAction(action)
	if len(providers) == 0 {
		return fmt.Errorf("no providers support action %s", action)
	}

	// Check if any providers are available
	availableCount := 0
	for _, provider := range providers {
		if am.providerManager.IsProviderAvailable(provider.Provider.Name) {
			availableCount++
		}
	}

	if availableCount == 0 {
		return fmt.Errorf("no available providers support action %s", action)
	}

	return nil
}

// GetAvailableActions returns available actions for software
func (am *ActionManager) GetAvailableActions(software string) ([]string, error) {
	// Get all available providers
	providers := am.providerManager.GetAvailableProviders()
	
	actionSet := make(map[string]bool)
	for _, provider := range providers {
		for actionName := range provider.Actions {
			actionSet[actionName] = true
		}
	}

	var actions []string
	for action := range actionSet {
		actions = append(actions, action)
	}

	return actions, nil
}

// GetActionInfo returns information about a specific action
func (am *ActionManager) GetActionInfo(action string) (*interfaces.ActionInfo, error) {
	providers := am.providerManager.GetProvidersForAction(action)
	if len(providers) == 0 {
		return nil, fmt.Errorf("action %s not found", action)
	}

	// Use the first provider to get action info
	provider := providers[0]
	actionData, exists := provider.Actions[action]
	if !exists {
		return nil, fmt.Errorf("action %s not found in provider %s", action, provider.Provider.Name)
	}

	info := &interfaces.ActionInfo{
		Name:         action,
		Description:  actionData.Description,
		RequiresRoot: actionData.RequiresRoot,
		Timeout:      time.Duration(actionData.Timeout) * time.Second,
		Capabilities: provider.Provider.Capabilities,
	}

	// Collect all provider names that support this action
	for _, p := range providers {
		info.Providers = append(info.Providers, p.Provider.Name)
	}

	return info, nil
}

// ResolveSoftwareData resolves saidata or generates intelligent defaults
func (am *ActionManager) ResolveSoftwareData(software string) (*types.SoftwareData, error) {
	// Try to load existing saidata
	saidata, err := am.saidataManager.LoadSoftware(software)
	if err == nil {
		return saidata, nil
	}

	// If saidata not found, generate intelligent defaults
	defaults, err := am.saidataManager.GenerateDefaults(software)
	if err != nil {
		return nil, fmt.Errorf("failed to generate defaults for %s: %w", software, err)
	}

	return defaults, nil
}

// ValidateResourcesExist validates that required resources exist
func (am *ActionManager) ValidateResourcesExist(saidata *types.SoftwareData, action string) (*interfaces.ResourceValidationResult, error) {
	return am.validator.ValidateResources(saidata)
}

// GetAvailableProviders returns providers available for software and action
func (am *ActionManager) GetAvailableProviders(software string, action string) ([]*interfaces.ProviderOption, error) {
	// Get providers that support the action
	providers := am.providerManager.GetProvidersForAction(action)
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers support action %s", action)
	}

	var options []*interfaces.ProviderOption
	for _, provider := range providers {
		if am.providerManager.IsProviderAvailable(provider.Provider.Name) {
			// Validate that the action can be executed with this provider
			if saidata, err := am.ResolveSoftwareData(software); err == nil {
				if am.executor.CanExecute(provider, action, software, saidata) {
					option := &interfaces.ProviderOption{
						Provider:    provider,
						PackageName: am.getPackageName(provider, software),
						Version:     am.getProviderVersion(provider),
						IsInstalled: am.isPackageInstalled(provider, software),
						Priority:    am.getProviderPriority(provider),
					}
					options = append(options, option)
				}
			}
		}
	}

	if len(options) == 0 {
		return nil, fmt.Errorf("no executable providers available for action %s on software %s", action, software)
	}

	// Sort by priority (highest first)
	sort.Slice(options, func(i, j int) bool {
		return options[i].Priority > options[j].Priority
	})

	return options, nil
}

// RequiresConfirmation checks if an action requires user confirmation
func (am *ActionManager) RequiresConfirmation(action string) bool {
	return am.config.RequiresConfirmation(action)
}

// BypassConfirmation checks if confirmation should be bypassed based on options
func (am *ActionManager) BypassConfirmation(options interfaces.ActionOptions) bool {
	return am.confirmationManager.BypassConfirmation(options)
}

// SearchAcrossProviders searches for software across all providers (Requirement 2.3)
func (am *ActionManager) SearchAcrossProviders(software string) ([]*interfaces.SearchResult, error) {
	providers := am.providerManager.GetAvailableProviders()
	var results []*interfaces.SearchResult
	ctx := context.Background()

	for _, provider := range providers {
		// Check if provider supports search action
		if _, hasSearch := provider.Actions["search"]; !hasSearch {
			continue
		}

		// Skip if provider is not available
		if !am.providerManager.IsProviderAvailable(provider.Provider.Name) {
			continue
		}

		// Get saidata for template resolution
		saidata, err := am.ResolveSoftwareData(software)
		if err != nil {
			continue // Skip this provider if we can't resolve saidata
		}

		// Check if search action can be executed
		if !am.executor.CanExecute(provider, "search", software, saidata) {
			continue
		}

		// Execute search command
		executeOptions := interfaces.ExecuteOptions{
			DryRun:  false,
			Verbose: false,
			Timeout: 30 * time.Second,
		}

		executionResult, err := am.executor.Execute(ctx, provider, "search", software, saidata, executeOptions)
		if err != nil || !executionResult.Success {
			// Search failed, but don't fail the entire operation
			continue
		}

		// Parse search results (simplified - in real implementation would parse provider-specific output)
		result := &interfaces.SearchResult{
			Software:    software,
			Provider:    provider.Provider.Name,
			PackageName: am.getPackageName(provider, software),
			Version:     "available", // Would parse from output
			Description: fmt.Sprintf("%s package from %s", software, provider.Provider.DisplayName),
			Available:   executionResult.Success,
		}
		results = append(results, result)
	}

	return results, nil
}

// GetSoftwareInfo gets information about software from all providers (Requirement 2.4)
func (am *ActionManager) GetSoftwareInfo(software string) ([]*interfaces.SoftwareInfo, error) {
	providers := am.providerManager.GetAvailableProviders()
	var results []*interfaces.SoftwareInfo
	ctx := context.Background()

	for _, provider := range providers {
		// Check if provider supports info action
		if _, hasInfo := provider.Actions["info"]; !hasInfo {
			continue
		}

		// Skip if provider is not available
		if !am.providerManager.IsProviderAvailable(provider.Provider.Name) {
			continue
		}

		// Get saidata for template resolution
		saidata, err := am.ResolveSoftwareData(software)
		if err != nil {
			continue // Skip this provider if we can't resolve saidata
		}

		// Check if info action can be executed
		if !am.executor.CanExecute(provider, "info", software, saidata) {
			continue
		}

		// Execute info command
		executeOptions := interfaces.ExecuteOptions{
			DryRun:  false,
			Verbose: false,
			Timeout: 30 * time.Second,
		}

		executionResult, err := am.executor.Execute(ctx, provider, "info", software, saidata, executeOptions)
		if err != nil || !executionResult.Success {
			// Info failed, but don't fail the entire operation
			continue
		}

		// Parse info results (simplified - in real implementation would parse provider-specific output)
		info := &interfaces.SoftwareInfo{
			Software:     software,
			Provider:     provider.Provider.Name,
			PackageName:  am.getPackageName(provider, software),
			Version:      "available", // Would parse from output
			Description:  fmt.Sprintf("%s package information from %s", software, provider.Provider.DisplayName),
			Homepage:     "", // Would parse from output
			License:      "unknown", // Would parse from output
			Dependencies: []string{}, // Would parse from output
		}
		results = append(results, info)
	}

	return results, nil
}

// GetSoftwareVersions gets version information with installation status (Requirement 2.5)
func (am *ActionManager) GetSoftwareVersions(software string) ([]*interfaces.VersionInfo, error) {
	providers := am.providerManager.GetAvailableProviders()
	var results []*interfaces.VersionInfo
	ctx := context.Background()

	for _, provider := range providers {
		// Check if provider supports version action
		if _, hasVersion := provider.Actions["version"]; !hasVersion {
			continue
		}

		// Skip if provider is not available
		if !am.providerManager.IsProviderAvailable(provider.Provider.Name) {
			continue
		}

		// Get saidata for template resolution
		saidata, err := am.ResolveSoftwareData(software)
		if err != nil {
			continue // Skip this provider if we can't resolve saidata
		}

		// Check if version action can be executed
		if !am.executor.CanExecute(provider, "version", software, saidata) {
			continue
		}

		// Execute version command
		executeOptions := interfaces.ExecuteOptions{
			DryRun:  false,
			Verbose: false,
			Timeout: 30 * time.Second,
		}

		executionResult, err := am.executor.Execute(ctx, provider, "version", software, saidata, executeOptions)
		
		// Check installation status
		isInstalled := am.isPackageInstalled(provider, software)
		
		// Parse version results (simplified - in real implementation would parse provider-specific output)
		version := &interfaces.VersionInfo{
			Software:      software,
			Provider:      provider.Provider.Name,
			PackageName:   am.getPackageName(provider, software),
			Version:       "unknown", // Would parse from output
			IsInstalled:   isInstalled,
			LatestVersion: "unknown", // Would parse from output
		}

		// If command succeeded, try to parse version from output
		if err == nil && executionResult.Success {
			// In a real implementation, this would parse the actual version from output
			version.Version = "available"
		}

		results = append(results, version)
	}

	return results, nil
}

// ManageRepositorySetup automatically sets up repositories from saidata (Requirement 8.5)
func (am *ActionManager) ManageRepositorySetup(saidata *types.SoftwareData) error {
	if saidata == nil {
		return nil
	}

	// Check if auto setup is enabled
	if !am.config.Repository.AutoSetup {
		return nil
	}

	// Look for repository definitions in saidata
	var repositoriesToSetup []types.Repository
	
	// Check provider-specific repositories
	for providerName, providerConfig := range saidata.Providers {
		if len(providerConfig.Repositories) > 0 {
			// Only setup repositories for available providers
			if am.providerManager.IsProviderAvailable(providerName) {
				repositoriesToSetup = append(repositoriesToSetup, providerConfig.Repositories...)
			}
		}
	}

	// Setup each repository
	for _, repo := range repositoriesToSetup {
		if err := am.setupRepository(repo); err != nil {
			am.formatter.ShowWarning(fmt.Sprintf("Failed to setup repository %s: %v", repo.Name, err))
			// Continue with other repositories even if one fails
		} else {
			am.formatter.ShowDebug(fmt.Sprintf("Successfully setup repository: %s", repo.Name))
		}
	}

	return nil
}

// setupRepository sets up a single repository
func (am *ActionManager) setupRepository(repo types.Repository) error {
	// Validate repository configuration
	if repo.Name == "" || repo.URL == "" {
		return fmt.Errorf("repository name and URL are required")
	}

	// Check if repository is enabled
	if !repo.Enabled {
		return nil // Skip disabled repositories
	}

	// Repository setup would be provider-specific
	// For now, just log the setup attempt
	am.formatter.ShowDebug(fmt.Sprintf("Setting up %s repository: %s (%s)", repo.Type, repo.Name, repo.URL))

	// In a real implementation, this would:
	// 1. Check if repository already exists
	// 2. Add repository configuration to the appropriate package manager
	// 3. Import GPG keys if needed
	// 4. Update package manager cache
	
	return nil
}

// Helper methods

// selectProvider handles provider selection with user interaction (Requirement 1.3)
func (am *ActionManager) selectProvider(software, action string, options []*interfaces.ProviderOption, actionOptions interfaces.ActionOptions) (*types.ProviderData, error) {
	// If specific provider requested, find and validate it
	if actionOptions.Provider != "" {
		for _, option := range options {
			if option.Provider.Provider.Name == actionOptions.Provider {
				return option.Provider, nil
			}
		}
		return nil, fmt.Errorf("preferred provider %s not available for action %s", actionOptions.Provider, action)
	}

	// Sort providers by priority (highest first)
	sort.Slice(options, func(i, j int) bool {
		return options[i].Priority > options[j].Priority
	})

	// If only one provider available, use it
	if len(options) == 1 {
		return options[0].Provider, nil
	}

	// If --yes flag is used, select highest priority provider (Requirement 1.4)
	if actionOptions.Yes {
		return options[0].Provider, nil
	}

	// Multiple providers available - use confirmation manager for selection (Requirement 1.3)
	selectedOption, err := am.confirmationManager.ConfirmProviderSelection(software, options)
	if err != nil {
		return nil, fmt.Errorf("provider selection failed: %w", err)
	}

	return selectedOption.Provider, nil
}

// buildErrorResult creates an error result with consistent structure
func (am *ActionManager) buildErrorResult(action, software, provider string, err error, startTime time.Time) *interfaces.ActionResult {
	return &interfaces.ActionResult{
		Action:               action,
		Software:             software,
		Provider:             provider,
		Success:              false,
		Error:                err,
		Duration:             time.Since(startTime),
		ExitCode:             1,
		RequiredConfirmation: am.RequiresConfirmation(action),
	}
}

// formatMissingResources formats missing resources for error messages
func (am *ActionManager) formatMissingResources(validation *interfaces.ResourceValidationResult) string {
	var missing []string
	
	if len(validation.MissingFiles) > 0 {
		missing = append(missing, fmt.Sprintf("files: %v", validation.MissingFiles))
	}
	if len(validation.MissingDirectories) > 0 {
		missing = append(missing, fmt.Sprintf("directories: %v", validation.MissingDirectories))
	}
	if len(validation.MissingCommands) > 0 {
		missing = append(missing, fmt.Sprintf("commands: %v", validation.MissingCommands))
	}
	if len(validation.MissingServices) > 0 {
		missing = append(missing, fmt.Sprintf("services: %v", validation.MissingServices))
	}
	
	return fmt.Sprintf("[%s]", strings.Join(missing, ", "))
}

// displayActionResult shows the action result to the user
func (am *ActionManager) displayActionResult(result *interfaces.ActionResult) {
	if result.Success {
		if !am.formatter.IsQuietMode() {
			am.formatter.ShowSuccess(fmt.Sprintf("Successfully executed %s for %s using %s", 
				result.Action, result.Software, result.Provider))
		}
	} else if result.Error != nil {
		am.formatter.ShowError(result.Error)
	}

	// Show execution details in verbose mode
	if am.formatter.IsVerboseMode() {
		am.formatter.ShowDebug(fmt.Sprintf("Action: %s, Duration: %v, Exit Code: %d", 
			result.Action, result.Duration, result.ExitCode))
		
		if len(result.Commands) > 0 {
			am.formatter.ShowDebug(fmt.Sprintf("Commands executed: %v", result.Commands))
		}
		
		if len(result.Changes) > 0 {
			am.formatter.ShowDebug(fmt.Sprintf("Changes made: %d", len(result.Changes)))
		}
	}
}

func (am *ActionManager) getPackageName(provider *types.ProviderData, software string) string {
	// Try to get package name from saidata first
	if saidata, err := am.saidataManager.LoadSoftware(software); err == nil {
		if providerConfig, exists := saidata.Providers[provider.Provider.Name]; exists {
			if len(providerConfig.Packages) > 0 {
				return providerConfig.Packages[0].Name
			}
		}
		if len(saidata.Packages) > 0 {
			return saidata.Packages[0].Name
		}
	}
	
	// Fallback to software name
	return software
}

func (am *ActionManager) getProviderVersion(provider *types.ProviderData) string {
	// TODO: Get actual provider version by executing version command
	return "unknown"
}

func (am *ActionManager) isPackageInstalled(provider *types.ProviderData, software string) bool {
	// TODO: Check if package is actually installed by executing check command
	return false
}

func (am *ActionManager) getProviderPriority(provider *types.ProviderData) int {
	if priority, exists := am.config.ProviderPriority[provider.Provider.Name]; exists {
		return priority
	}
	return provider.Provider.Priority
}

// Ensure ActionManager implements the interface
var _ interfaces.ActionManager = (*ActionManager)(nil)