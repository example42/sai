package action

import (
	"context"
	"fmt"
	"time"

	"sai/internal/config"
	"sai/internal/interfaces"
	"sai/internal/types"
)

// ActionManager orchestrates software management operations
type ActionManager struct {
	providerManager interfaces.ProviderManager
	saidataManager  interfaces.SaidataManager
	executor        interfaces.GenericExecutor
	validator       interfaces.ResourceValidator
	config          *config.Config
}

// NewActionManager creates a new action manager
func NewActionManager(
	providerManager interfaces.ProviderManager,
	saidataManager interfaces.SaidataManager,
	executor interfaces.GenericExecutor,
	validator interfaces.ResourceValidator,
	config *config.Config,
) *ActionManager {
	return &ActionManager{
		providerManager: providerManager,
		saidataManager:  saidataManager,
		executor:        executor,
		validator:       validator,
		config:          config,
	}
}

// ExecuteAction executes a specific action on software
func (am *ActionManager) ExecuteAction(ctx context.Context, action string, software string, options interfaces.ActionOptions) (*interfaces.ActionResult, error) {
	startTime := time.Now()

	// Resolve software data (saidata or intelligent defaults)
	saidata, err := am.ResolveSoftwareData(software)
	if err != nil {
		return &interfaces.ActionResult{
			Action:   action,
			Software: software,
			Success:  false,
			Error:    fmt.Errorf("failed to resolve software data: %w", err),
			Duration: time.Since(startTime),
			ExitCode: 1,
		}, err
	}

	// Get available providers for this software and action
	providerOptions, err := am.GetAvailableProviders(software, action)
	if err != nil {
		return &interfaces.ActionResult{
			Action:   action,
			Software: software,
			Success:  false,
			Error:    fmt.Errorf("failed to get available providers: %w", err),
			Duration: time.Since(startTime),
			ExitCode: 1,
		}, err
	}

	if len(providerOptions) == 0 {
		return &interfaces.ActionResult{
			Action:   action,
			Software: software,
			Success:  false,
			Error:    fmt.Errorf("no providers available for action %s on software %s", action, software),
			Duration: time.Since(startTime),
			ExitCode: 1,
		}, fmt.Errorf("no providers available")
	}

	// Select provider (preferred, or first available)
	var selectedProvider *types.ProviderData
	if options.Provider != "" {
		// Find the preferred provider in available options
		for _, option := range providerOptions {
			if option.Provider.Provider.Name == options.Provider {
				selectedProvider = option.Provider
				break
			}
		}
		if selectedProvider == nil {
			return &interfaces.ActionResult{
				Action:   action,
				Software: software,
				Success:  false,
				Error:    fmt.Errorf("preferred provider %s not available for action %s", options.Provider, action),
				Duration: time.Since(startTime),
				ExitCode: 1,
			}, fmt.Errorf("preferred provider not available")
		}
	} else {
		// Use the first (highest priority) available provider
		selectedProvider = providerOptions[0].Provider
	}

	// Validate resources exist before execution
	validationResult, err := am.ValidateResourcesExist(saidata, action)
	if err != nil {
		return &interfaces.ActionResult{
			Action:   action,
			Software: software,
			Provider: selectedProvider.Provider.Name,
			Success:  false,
			Error:    fmt.Errorf("resource validation failed: %w", err),
			Duration: time.Since(startTime),
			ExitCode: 1,
		}, err
	}

	if !validationResult.CanProceed {
		return &interfaces.ActionResult{
			Action:   action,
			Software: software,
			Provider: selectedProvider.Provider.Name,
			Success:  false,
			Error:    fmt.Errorf("cannot proceed: missing resources: %v", validationResult.MissingFiles),
			Duration: time.Since(startTime),
			ExitCode: 1,
		}, fmt.Errorf("missing required resources")
	}

	// Execute the action
	executeOptions := interfaces.ExecuteOptions{
		DryRun:    options.DryRun,
		Verbose:   options.Verbose,
		Timeout:   options.Timeout,
		Variables: options.Variables,
	}

	var executionResult *interfaces.ExecutionResult
	if options.DryRun {
		executionResult, err = am.executor.DryRun(ctx, selectedProvider, action, software, saidata, executeOptions)
	} else {
		executionResult, err = am.executor.Execute(ctx, selectedProvider, action, software, saidata, executeOptions)
	}

	// Build action result
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

	return options, nil
}

// RequiresConfirmation checks if an action requires user confirmation
func (am *ActionManager) RequiresConfirmation(action string) bool {
	return am.config.RequiresConfirmation(action)
}

// SearchAcrossProviders searches for software across all providers (Requirement 2.3)
func (am *ActionManager) SearchAcrossProviders(software string) ([]*interfaces.SearchResult, error) {
	providers := am.providerManager.GetAvailableProviders()
	var results []*interfaces.SearchResult

	for _, provider := range providers {
		// Check if provider supports search action
		if _, hasSearch := provider.Actions["search"]; !hasSearch {
			continue
		}

		// TODO: Execute search command and parse results
		// For now, return a placeholder result
		result := &interfaces.SearchResult{
			Software:    software,
			Provider:    provider.Provider.Name,
			PackageName: am.getPackageName(provider, software),
			Version:     "unknown",
			Description: fmt.Sprintf("%s package from %s", software, provider.Provider.DisplayName),
			Available:   true,
		}
		results = append(results, result)
	}

	return results, nil
}

// GetSoftwareInfo gets information about software from all providers (Requirement 2.4)
func (am *ActionManager) GetSoftwareInfo(software string) ([]*interfaces.SoftwareInfo, error) {
	providers := am.providerManager.GetAvailableProviders()
	var results []*interfaces.SoftwareInfo

	for _, provider := range providers {
		// Check if provider supports info action
		if _, hasInfo := provider.Actions["info"]; !hasInfo {
			continue
		}

		// TODO: Execute info command and parse results
		// For now, return a placeholder result
		info := &interfaces.SoftwareInfo{
			Software:    software,
			Provider:    provider.Provider.Name,
			PackageName: am.getPackageName(provider, software),
			Version:     "unknown",
			Description: fmt.Sprintf("%s package information from %s", software, provider.Provider.DisplayName),
			Homepage:    "",
			License:     "unknown",
		}
		results = append(results, info)
	}

	return results, nil
}

// GetSoftwareVersions gets version information with installation status (Requirement 2.5)
func (am *ActionManager) GetSoftwareVersions(software string) ([]*interfaces.VersionInfo, error) {
	providers := am.providerManager.GetAvailableProviders()
	var results []*interfaces.VersionInfo

	for _, provider := range providers {
		// Check if provider supports version action
		if _, hasVersion := provider.Actions["version"]; !hasVersion {
			continue
		}

		// TODO: Execute version command and parse results
		// For now, return a placeholder result
		version := &interfaces.VersionInfo{
			Software:      software,
			Provider:      provider.Provider.Name,
			PackageName:   am.getPackageName(provider, software),
			Version:       "unknown",
			IsInstalled:   am.isPackageInstalled(provider, software),
			LatestVersion: "unknown",
		}
		results = append(results, version)
	}

	return results, nil
}

// ManageRepositorySetup automatically sets up repositories from saidata (Requirement 8.5)
func (am *ActionManager) ManageRepositorySetup(saidata *types.SoftwareData) error {
	// TODO: Implement repository setup based on saidata
	return nil
}

// Helper methods

func (am *ActionManager) getPackageName(provider *types.ProviderData, software string) string {
	// TODO: Get actual package name from saidata
	return software
}

func (am *ActionManager) getProviderVersion(provider *types.ProviderData) string {
	// TODO: Get actual provider version
	return "unknown"
}

func (am *ActionManager) isPackageInstalled(provider *types.ProviderData, software string) bool {
	// TODO: Check if package is actually installed
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