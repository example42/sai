package action

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"sai/internal/config"
	"sai/internal/errors"
	"sai/internal/interfaces"
	"sai/internal/output"
	"sai/internal/types"
	"sai/internal/ui"
)

// ActionManager orchestrates software management operations
type ActionManager struct {
	providerManager       interfaces.ProviderManager
	saidataManager        interfaces.SaidataManager
	executor              interfaces.GenericExecutor
	validator             interfaces.ResourceValidator
	config                *config.Config
	ui                    *ui.UserInterface
	formatter             *output.OutputFormatter
	safetyManager         *SafetyManager
	confirmationManager   *ConfirmationManager
	recoveryManager       *errors.RecoveryManager
	circuitBreakerManager *errors.CircuitBreakerManager
	errorTracker          *errors.ErrorContextTracker
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
	logger interfaces.Logger,
) *ActionManager {
	safetyManager := NewSafetyManager(validator)
	confirmationManager := NewConfirmationManager(config, userInterface, formatter)
	
	// Initialize error handling and recovery systems
	recoveryConfig := errors.DefaultRecoveryConfig()
	if config.Recovery != nil {
		recoveryConfig = config.Recovery
	}
	
	circuitBreakerConfig := errors.DefaultCircuitBreakerConfig()
	if config.CircuitBreaker != nil {
		circuitBreakerConfig = config.CircuitBreaker
	}
	
	recoveryManager := errors.NewRecoveryManager(executor, providerManager, logger, recoveryConfig)
	circuitBreakerManager := errors.NewCircuitBreakerManager(circuitBreakerConfig)
	errorTracker := errors.NewErrorContextTracker(1000) // Keep last 1000 errors
	
	return &ActionManager{
		providerManager:       providerManager,
		saidataManager:        saidataManager,
		executor:              executor,
		validator:             validator,
		config:                config,
		ui:                    userInterface,
		formatter:             formatter,
		safetyManager:         safetyManager,
		confirmationManager:   confirmationManager,
		recoveryManager:       recoveryManager,
		circuitBreakerManager: circuitBreakerManager,
		errorTracker:          errorTracker,
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

	// Handle automatic execution across all providers for information-only commands (Requirements 15.2, 15.4)
	if selectedProvider == nil && am.confirmationManager.ShouldExecuteAcrossProviders(action) {
		return am.executeAcrossProviders(ctx, action, software, providerOptions, options, saidata, startTime)
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

	// Step 9: Execute the action with circuit breaker protection and error recovery
	var executionResult *interfaces.ExecutionResult
	if options.DryRun {
		am.formatter.ShowInfo("Dry run mode - showing commands that would be executed:")
		executionResult, err = am.executor.DryRun(ctx, selectedProvider, action, software, saidata, executeOptions)
	} else {
		// Execute with circuit breaker protection
		circuitBreakerName := fmt.Sprintf("%s_%s", selectedProvider.Provider.Name, action)
		err = am.circuitBreakerManager.ExecuteWithCircuitBreaker(circuitBreakerName, func() error {
			var execErr error
			executionResult, execErr = am.executor.Execute(ctx, selectedProvider, action, software, saidata, executeOptions)
			return execErr
		})
		
		// If execution failed and error is recoverable, attempt recovery
		if err != nil && errors.IsRecoverable(err) {
			am.formatter.ShowWarning("Action failed, attempting recovery...")
			
			// Track the error for debugging
			errorCtx := am.errorTracker.TrackError(ctx, action, software, selectedProvider.Provider.Name, err)
			am.formatter.ShowDebug(fmt.Sprintf("Error tracked with ID: %s", errorCtx.ID))
			
			// Build recovery context
			recoveryCtx := errors.BuildRecoveryContext(action, software, selectedProvider, saidata, err)
			
			// Attempt recovery
			recoveryResult, _ := am.recoveryManager.AttemptRecovery(ctx, recoveryCtx)
			
			if recoveryResult.Success {
				am.formatter.ShowSuccess(fmt.Sprintf("Recovery successful using strategy: %s", recoveryResult.RecoveryStrategy))
				// Create a successful execution result
				executionResult = &interfaces.ExecutionResult{
					Success:  true,
					Output:   fmt.Sprintf("Recovered from error using %s strategy", recoveryResult.RecoveryStrategy),
					Commands: []string{}, // Would be populated by recovery
					ExitCode: 0,
					Duration: recoveryResult.Duration,
				}
				err = nil // Clear the error since recovery succeeded
			} else {
				am.formatter.ShowError(fmt.Errorf("Recovery failed: %v", recoveryResult.FinalError))
				err = recoveryResult.FinalError
				
				// Track the recovery failure
				am.errorTracker.TrackError(ctx, action, software, selectedProvider.Provider.Name, recoveryResult.FinalError)
			}
		} else if err != nil {
			// Track non-recoverable errors
			am.errorTracker.TrackError(ctx, action, software, selectedProvider.Provider.Name, err)
		}
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
				} else {
					// Log why this provider was rejected
					am.formatter.ShowDebug(fmt.Sprintf("Provider %s rejected: action %s cannot be executed", provider.Provider.Name, action))
				}
			} else {
				// Log saidata resolution failure
				am.formatter.ShowDebug(fmt.Sprintf("Provider %s rejected: failed to resolve saidata for %s: %v", provider.Provider.Name, software, err))
			}
		} else {
			// Log provider availability issue
			am.formatter.ShowDebug(fmt.Sprintf("Provider %s not available on this system", provider.Provider.Name))
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
	var results []*interfaces.SoftwareInfo

	// First, try to get information from saidata
	saidata, err := am.ResolveSoftwareData(software)
	if err == nil && !saidata.IsGenerated {
		// We have actual saidata (not generated defaults), use it as a source
		homepage := ""
		if saidata.Metadata.URLs != nil {
			homepage = saidata.Metadata.URLs.Website
		}
		
		license := saidata.Metadata.License
		if license == "" {
			license = "unknown"
		}

		info := &interfaces.SoftwareInfo{
			Software:     software,
			Provider:     "saidata",
			PackageName:  software,
			Version:      saidata.Metadata.Version,
			Description:  saidata.Metadata.Description,
			Homepage:     homepage,
			License:      license,
			Dependencies: []string{}, // Could extract from requirements if available
		}
		results = append(results, info)
	}

	// Then get information from providers that support info action
	providers := am.providerManager.GetAvailableProviders()
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

		// Get saidata for template resolution (reuse if already loaded)
		if saidata == nil {
			saidata, err = am.ResolveSoftwareData(software)
			if err != nil {
				continue // Skip this provider if we can't resolve saidata
			}
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
	var errors []error
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
			errors = append(errors, fmt.Errorf("failed to resolve saidata for %s with provider %s: %w", software, provider.Provider.Name, err))
			continue // Skip this provider if we can't resolve saidata
		}

		// Check if version action can be executed
		if !am.executor.CanExecute(provider, "version", software, saidata) {
			continue
		}

		// Check installation status first
		isInstalled := am.isPackageInstalled(provider, software)
		
		// Create version info with basic information
		version := &interfaces.VersionInfo{
			Software:      software,
			Provider:      provider.Provider.Name,
			PackageName:   am.getPackageName(provider, software),
			Version:       "Not Installed",
			IsInstalled:   isInstalled,
			LatestVersion: "unknown",
		}

		// Only try to get version if package is installed or if we want to check availability
		executeOptions := interfaces.ExecuteOptions{
			DryRun:  false,
			Verbose: false,
			Timeout: 30 * time.Second,
		}

		executionResult, err := am.executor.Execute(ctx, provider, "version", software, saidata, executeOptions)
		
		if err != nil {
			// Add error but still include the version info to show the provider exists
			errors = append(errors, fmt.Errorf("failed to get version for %s from %s: %w", software, provider.Provider.Name, err))
			version.Version = "Error"
		} else if executionResult.Success {
			// Parse version from output based on provider type
			parsedVersion := am.parseVersionOutput(provider.Provider.Name, executionResult.Output)
			if parsedVersion != "" {
				version.Version = parsedVersion
			} else if isInstalled {
				version.Version = "Installed"
			} else {
				version.Version = "Available"
			}
		} else {
			// Command failed but no error - likely not installed or not available
			if isInstalled {
				version.Version = "Installed (version unknown)"
			} else {
				version.Version = "Not Available"
			}
		}

		results = append(results, version)
	}

	// If we have no results but have errors, return the first error
	if len(results) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("failed to get version information: %w", errors[0])
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
// Updated to support automatic execution for information-only commands (Requirements 15.2, 15.4)
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

	// For information-only commands, don't prompt for provider selection - execute across all providers
	// This implements Requirements 15.2 and 15.4
	if am.confirmationManager.ShouldExecuteAcrossProviders(action) {
		// Return nil to signal that we should execute across all providers
		return nil, nil
	}

	// For system-changing operations, show provider selection with commands
	// Generate command previews for each provider to show in selection UI (Requirement 15.1, 15.3)
	commands := make(map[string][]string)
	
	for _, option := range options {
		// Generate a preview command based on provider and action
		// This is a simplified approach that shows the expected command format
		providerName := option.Provider.Provider.Name
		packageName := option.PackageName
		
		var previewCommand string
		switch action {
		case "install":
			previewCommand = am.generateInstallCommand(providerName, packageName)
		case "uninstall":
			previewCommand = am.generateUninstallCommand(providerName, packageName)
		case "upgrade":
			previewCommand = am.generateUpgradeCommand(providerName, packageName)
		case "start":
			previewCommand = am.generateStartCommand(providerName, packageName)
		case "stop":
			previewCommand = am.generateStopCommand(providerName, packageName)
		case "restart":
			previewCommand = am.generateRestartCommand(providerName, packageName)
		default:
			previewCommand = fmt.Sprintf("%s %s %s", providerName, action, packageName)
		}
		
		if previewCommand != "" {
			commands[providerName] = []string{previewCommand}
		}
	}

	// Multiple providers available - use confirmation manager for selection (Requirement 1.3)
	selectedOption, err := am.confirmationManager.ConfirmProviderSelection(software, options, action, commands)
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

// Helper methods to generate command previews for provider selection (Requirement 15.3)

func (am *ActionManager) generateInstallCommand(provider, packageName string) string {
	switch provider {
	case "apt":
		return fmt.Sprintf("apt install %s", packageName)
	case "brew":
		return fmt.Sprintf("brew install %s", packageName)
	case "yum":
		return fmt.Sprintf("yum install %s", packageName)
	case "dnf":
		return fmt.Sprintf("dnf install %s", packageName)
	case "pacman":
		return fmt.Sprintf("pacman -S %s", packageName)
	case "docker":
		return fmt.Sprintf("docker pull %s", packageName)
	case "npm":
		return fmt.Sprintf("npm install -g %s", packageName)
	case "pip", "pypi":
		return fmt.Sprintf("pip install %s", packageName)
	case "gem":
		return fmt.Sprintf("gem install %s", packageName)
	case "cargo":
		return fmt.Sprintf("cargo install %s", packageName)
	case "go":
		return fmt.Sprintf("go install %s", packageName)
	case "helm":
		return fmt.Sprintf("helm install %s", packageName)
	default:
		return fmt.Sprintf("%s install %s", provider, packageName)
	}
}

func (am *ActionManager) generateUninstallCommand(provider, packageName string) string {
	switch provider {
	case "apt":
		return fmt.Sprintf("apt remove %s", packageName)
	case "brew":
		return fmt.Sprintf("brew uninstall %s", packageName)
	case "yum":
		return fmt.Sprintf("yum remove %s", packageName)
	case "dnf":
		return fmt.Sprintf("dnf remove %s", packageName)
	case "pacman":
		return fmt.Sprintf("pacman -R %s", packageName)
	case "docker":
		return fmt.Sprintf("docker rmi %s", packageName)
	case "npm":
		return fmt.Sprintf("npm uninstall -g %s", packageName)
	case "pip", "pypi":
		return fmt.Sprintf("pip uninstall %s", packageName)
	case "gem":
		return fmt.Sprintf("gem uninstall %s", packageName)
	case "cargo":
		return fmt.Sprintf("cargo uninstall %s", packageName)
	case "helm":
		return fmt.Sprintf("helm uninstall %s", packageName)
	default:
		return fmt.Sprintf("%s uninstall %s", provider, packageName)
	}
}

func (am *ActionManager) generateUpgradeCommand(provider, packageName string) string {
	switch provider {
	case "apt":
		return fmt.Sprintf("apt upgrade %s", packageName)
	case "brew":
		return fmt.Sprintf("brew upgrade %s", packageName)
	case "yum":
		return fmt.Sprintf("yum update %s", packageName)
	case "dnf":
		return fmt.Sprintf("dnf upgrade %s", packageName)
	case "pacman":
		return fmt.Sprintf("pacman -Syu %s", packageName)
	case "docker":
		return fmt.Sprintf("docker pull %s", packageName)
	case "npm":
		return fmt.Sprintf("npm update -g %s", packageName)
	case "pip", "pypi":
		return fmt.Sprintf("pip install --upgrade %s", packageName)
	case "gem":
		return fmt.Sprintf("gem update %s", packageName)
	case "cargo":
		return fmt.Sprintf("cargo install %s", packageName)
	case "helm":
		return fmt.Sprintf("helm upgrade %s", packageName)
	default:
		return fmt.Sprintf("%s upgrade %s", provider, packageName)
	}
}

func (am *ActionManager) generateStartCommand(provider, packageName string) string {
	// Service commands are typically system-level, not provider-specific
	return fmt.Sprintf("systemctl start %s", packageName)
}

func (am *ActionManager) generateStopCommand(provider, packageName string) string {
	return fmt.Sprintf("systemctl stop %s", packageName)
}

func (am *ActionManager) generateRestartCommand(provider, packageName string) string {
	return fmt.Sprintf("systemctl restart %s", packageName)
}

// executeAcrossProviders executes an action across all available providers for information-only commands
// This implements Requirements 15.2 and 15.4 - automatic execution without provider selection prompts
func (am *ActionManager) executeAcrossProviders(ctx context.Context, action, software string, providerOptions []*interfaces.ProviderOption, actionOptions interfaces.ActionOptions, saidata *types.SoftwareData, startTime time.Time) (*interfaces.ActionResult, error) {
	var allResults []*interfaces.ExecutionResult
	var allCommands []string
	var allOutput []string
	var hasErrors bool
	var lastError error

	am.formatter.ShowInfo(fmt.Sprintf("Executing %s for %s across all available providers:", action, software))
	fmt.Println()

	executeOptions := interfaces.ExecuteOptions{
		DryRun:    actionOptions.DryRun,
		Verbose:   actionOptions.Verbose,
		Timeout:   actionOptions.Timeout,
		Variables: actionOptions.Variables,
	}

	for _, option := range providerOptions {
		provider := option.Provider
		
		// Show compact provider header (Requirement 15.5)
		providerHeader := am.formatter.FormatProviderName(provider.Provider.Name)
		fmt.Printf("%s:\n", providerHeader)

		// Execute the action for this provider
		var executionResult *interfaces.ExecutionResult
		var err error

		if actionOptions.DryRun {
			executionResult, err = am.executor.DryRun(ctx, provider, action, software, saidata, executeOptions)
		} else {
			executionResult, err = am.executor.Execute(ctx, provider, action, software, saidata, executeOptions)
		}

		if err != nil {
			hasErrors = true
			lastError = err
			am.formatter.ShowError(fmt.Errorf("  %s failed: %v", provider.Provider.Name, err))
		} else if executionResult != nil {
			allResults = append(allResults, executionResult)
			allCommands = append(allCommands, executionResult.Commands...)
			
			// Show compact output format (Requirements 15.3, 15.5)
			if len(executionResult.Commands) > 0 {
				for _, cmd := range executionResult.Commands {
					fmt.Printf("  Command: %s\n", cmd)
				}
			}
			
			if executionResult.Output != "" && !am.formatter.IsQuietMode() {
				// Show output with proper formatting
				outputLines := strings.Split(strings.TrimSpace(executionResult.Output), "\n")
				for _, line := range outputLines {
					if line != "" {
						fmt.Printf("  %s\n", line)
					}
				}
				allOutput = append(allOutput, executionResult.Output)
			}
			
			// Show exit status
			if executionResult.ExitCode == 0 {
				am.formatter.ShowSuccess("  ✓ Success")
			} else {
				am.formatter.ShowError(fmt.Errorf("  ✗ Failed (exit code: %d)", executionResult.ExitCode))
				hasErrors = true
			}
		}
		
		fmt.Println() // Add spacing between providers
	}

	// Build combined result
	result := &interfaces.ActionResult{
		Action:               action,
		Software:             software,
		Provider:             "multiple", // Indicate multiple providers were used
		Success:              !hasErrors,
		Duration:             time.Since(startTime),
		Commands:             allCommands,
		Output:               strings.Join(allOutput, "\n"),
		ExitCode:             0,
		RequiredConfirmation: false, // Information-only commands don't require confirmation
	}

	if hasErrors {
		result.Success = false
		result.Error = lastError
		result.ExitCode = 1
	}

	return result, lastError
}

// parseVersionOutput parses version information from provider command output
func (am *ActionManager) parseVersionOutput(providerName, output string) string {
	if output == "" {
		return ""
	}
	
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return ""
	}
	
	switch providerName {
	case "apt":
		// APT output format: "package version"
		// Example: "nginx 1.18.0-6ubuntu14.4"
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1] // Return version part
			}
		}
		
	case "brew":
		// Homebrew output format: "package version"
		// Example: "nginx 1.21.6_1"
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1] // Return version part
			}
		}
		
	case "docker":
		// Docker doesn't have traditional versions, but we can show image tag
		// This would be handled differently in a real implementation
		return "container"
		
	case "yum", "dnf":
		// YUM/DNF output format similar to APT
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1]
			}
		}
		
	case "pacman":
		// Pacman output format: "package version-release"
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1]
			}
		}
		
	default:
		// Generic parsing: try to extract version-like strings
		for _, line := range lines {
			// Look for version patterns like "1.2.3", "v1.2.3", etc.
			versionRegex := regexp.MustCompile(`\b(?:v)?(\d+(?:\.\d+)*(?:[-._]\w+)*)\b`)
			matches := versionRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				return matches[1]
			}
		}
	}
	
	// If we can't parse a specific version, return the first line (trimmed)
	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		if len(firstLine) > 50 {
			return firstLine[:47] + "..."
		}
		return firstLine
	}
	
	return ""
}

func (am *ActionManager) getProviderVersion(provider *types.ProviderData) string {
	// TODO: Get actual provider version by executing version command
	return "unknown"
}

func (am *ActionManager) isPackageInstalled(provider *types.ProviderData, software string) bool {
	// Check if provider has a detection command or list action
	ctx := context.Background()
	
	// Try detection command first if available
	if action, hasAction := provider.Actions["version"]; hasAction && action.Detection != "" {
		saidata, err := am.ResolveSoftwareData(software)
		if err != nil {
			return false
		}
		
		// Render detection command template
		detectionCmd, err := am.executor.RenderTemplate(action.Detection, saidata, provider)
		if err != nil {
			return false
		}
		
		// Execute detection command
		result, err := am.executor.ExecuteCommand(ctx, detectionCmd, interfaces.CommandOptions{
			Timeout: 10 * time.Second,
			Verbose: false,
		})
		
		// If command succeeds (exit code 0), package is installed
		return err == nil && result.ExitCode == 0
	}
	
	// Try list action as fallback
	if _, hasListAction := provider.Actions["list"]; hasListAction {
		saidata, err := am.ResolveSoftwareData(software)
		if err != nil {
			return false
		}
		
		if am.executor.CanExecute(provider, "list", software, saidata) {
			executeOptions := interfaces.ExecuteOptions{
				DryRun:  false,
				Verbose: false,
				Timeout: 10 * time.Second,
			}
			
			result, err := am.executor.Execute(ctx, provider, "list", software, saidata, executeOptions)
			return err == nil && result.Success && result.ExitCode == 0
		}
	}
	
	return false
}

func (am *ActionManager) getProviderPriority(provider *types.ProviderData) int {
	if priority, exists := am.config.ProviderPriority[provider.Provider.Name]; exists {
		return priority
	}
	return provider.Provider.Priority
}

// GetErrorStats returns error statistics for debugging and monitoring
func (am *ActionManager) GetErrorStats() *errors.ErrorStats {
	return am.errorTracker.GetErrorStats()
}

// GetRecentErrors returns recent error contexts for troubleshooting
func (am *ActionManager) GetRecentErrors(limit int) []*errors.ErrorContext {
	return am.errorTracker.GetRecentErrors(limit)
}

// GetCircuitBreakerStats returns circuit breaker statistics
func (am *ActionManager) GetCircuitBreakerStats() map[string]*errors.CircuitBreakerStats {
	return am.circuitBreakerManager.GetAllStats()
}

// ResetCircuitBreakers resets all circuit breakers
func (am *ActionManager) ResetCircuitBreakers() {
	am.circuitBreakerManager.ResetAll()
	am.formatter.ShowInfo("All circuit breakers have been reset")
}

// ResetCircuitBreaker resets a specific circuit breaker
func (am *ActionManager) ResetCircuitBreaker(name string) error {
	err := am.circuitBreakerManager.ResetCircuitBreaker(name)
	if err != nil {
		return errors.WrapSAIError(errors.ErrorTypeInternal, "failed to reset circuit breaker", err)
	}
	am.formatter.ShowInfo(fmt.Sprintf("Circuit breaker '%s' has been reset", name))
	return nil
}

// ClearErrorHistory clears the error tracking history
func (am *ActionManager) ClearErrorHistory() {
	am.errorTracker.ClearErrors()
	am.formatter.ShowInfo("Error history has been cleared")
}

// GetErrorContext retrieves detailed error context by ID
func (am *ActionManager) GetErrorContext(errorID string) (*errors.ErrorContext, bool) {
	return am.errorTracker.GetErrorContext(errorID)
}

// ExecuteWithTimeout executes an action with a custom timeout and enhanced error handling
func (am *ActionManager) ExecuteWithTimeout(ctx context.Context, action string, software string, options interfaces.ActionOptions, timeout time.Duration) (*interfaces.ActionResult, error) {
	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Add timeout information to context for error tracking
	timeoutCtx = errors.WithVariable(timeoutCtx, "timeout", timeout.String())
	timeoutCtx = errors.WithStartTime(timeoutCtx, time.Now())
	
	// Execute with the timeout context
	result, err := am.ExecuteAction(timeoutCtx, action, software, options)
	
	// Handle timeout errors specifically
	if err != nil && timeoutCtx.Err() == context.DeadlineExceeded {
		timeoutErr := errors.NewActionTimeoutError(action, software, timeout.String()).
			WithContext("provider", result.Provider).
			WithSuggestion("Increase timeout value").
			WithSuggestion("Check system performance and network connectivity")
		
		am.errorTracker.TrackError(timeoutCtx, action, software, result.Provider, timeoutErr)
		return result, timeoutErr
	}
	
	return result, err
}

// ValidateActionWithRecovery validates an action and provides recovery suggestions
func (am *ActionManager) ValidateActionWithRecovery(action string, software string) (*interfaces.ValidationResult, error) {
	// Perform basic validation
	err := am.ValidateAction(action, software)
	if err == nil {
		return &interfaces.ValidationResult{
			Valid:       true,
			Suggestions: []string{},
		}, nil
	}
	
	// Generate recovery suggestions based on error type
	var suggestions []string
	if saiErr, ok := err.(*errors.SAIError); ok {
		suggestions = append(suggestions, saiErr.Suggestions...)
		
		switch saiErr.Type {
		case errors.ErrorTypeProviderNotFound:
			suggestions = append(suggestions, "Install required package managers")
			suggestions = append(suggestions, "Check provider availability with 'sai stats'")
		case errors.ErrorTypeActionNotSupported:
			suggestions = append(suggestions, "Try a different action")
			suggestions = append(suggestions, "Check available actions with 'sai info "+software+"'")
		}
	}
	
	return &interfaces.ValidationResult{
		Valid:       false,
		Error:       err,
		Suggestions: suggestions,
	}, err
}

// GetProviderManager returns the provider manager for stats and debugging
func (am *ActionManager) GetProviderManager() interfaces.ProviderManager {
	return am.providerManager
}

// Ensure ActionManager implements the interface
var _ interfaces.ActionManager = (*ActionManager)(nil)