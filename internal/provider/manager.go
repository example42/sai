package provider

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"sai/internal/interfaces"
	"sai/internal/types"
)

// ProviderManager implements the provider management interface
type ProviderManager struct {
	loader    *ProviderLoader
	detector  *ProviderDetector
	providers map[string]*types.ProviderData
	mutex     sync.RWMutex
	config    *ManagerConfig
}

// ManagerConfig contains configuration for the provider manager
type ManagerConfig struct {
	ProviderDirectory string
	SchemaPath        string
	DefaultProvider   string
	ProviderPriority  map[string]int
	EnableWatching    bool
}

// ProviderSelection represents a provider option for user selection
type ProviderSelection struct {
	Provider    *types.ProviderData
	PackageName string
	Version     string
	IsInstalled bool
	Priority    int
	Available   bool
	Reason      string // Why this provider was selected/rejected
}

// NewProviderManager creates a new provider manager
func NewProviderManager(config *ManagerConfig) (*ProviderManager, error) {
	// Create loader
	loader, err := NewProviderLoader(config.SchemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider loader: %w", err)
	}

	// Create detector
	detector, err := NewProviderDetector()
	if err != nil {
		return nil, fmt.Errorf("failed to create provider detector: %w", err)
	}

	manager := &ProviderManager{
		loader:    loader,
		detector:  detector,
		providers: make(map[string]*types.ProviderData),
		config:    config,
	}

	// Load providers from directory
	if err := manager.LoadProviders(config.ProviderDirectory); err != nil {
		return nil, fmt.Errorf("failed to load providers: %w", err)
	}

	// Set up file watching if enabled
	if config.EnableWatching {
		err = manager.setupWatching()
		if err != nil {
			// Log warning but don't fail - watching is optional
			fmt.Printf("Warning: failed to set up provider watching: %v\n", err)
		}
	}

	return manager, nil
}

// LoadProviders loads all providers from the specified directory
func (pm *ProviderManager) LoadProviders(providerDir string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Load providers from directory
	loadedProviders, err := pm.loader.LoadFromDirectory(providerDir)
	if err != nil {
		// Check if this is a partial failure (some providers loaded)
		if len(loadedProviders) > 0 {
			fmt.Printf("Warning: some providers failed to load: %v\n", err)
		} else {
			return fmt.Errorf("failed to load providers from %s: %w", providerDir, err)
		}
	}

	// Store providers in map
	pm.providers = make(map[string]*types.ProviderData)
	for _, provider := range loadedProviders {
		pm.providers[provider.Provider.Name] = provider
	}

	return nil
}

// GetProvider returns a provider by name
func (pm *ProviderManager) GetProvider(name string) (*types.ProviderData, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	provider, exists := pm.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// GetAvailableProviders returns all available providers
func (pm *ProviderManager) GetAvailableProviders() []*types.ProviderData {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var available []*types.ProviderData
	for _, provider := range pm.providers {
		if pm.detector.IsAvailable(provider) {
			available = append(available, provider)
		}
	}

	// Sort by priority (highest first)
	sort.Slice(available, func(i, j int) bool {
		priorityI := pm.getEffectivePriority(available[i])
		priorityJ := pm.getEffectivePriority(available[j])
		return priorityI > priorityJ
	})

	return available
}

// SelectProvider selects the best provider for a software and action
func (pm *ProviderManager) SelectProvider(software string, action string, preferredProvider string) (*types.ProviderData, error) {
	// If a preferred provider is specified, try to use it
	if preferredProvider != "" {
		provider, err := pm.GetProvider(preferredProvider)
		if err != nil {
			return nil, fmt.Errorf("preferred provider %s not found: %w", preferredProvider, err)
		}

		if !pm.detector.IsAvailable(provider) {
			return nil, fmt.Errorf("preferred provider %s is not available on this system", preferredProvider)
		}

		if !pm.detector.SupportsAction(provider, action) {
			return nil, fmt.Errorf("preferred provider %s does not support action %s", preferredProvider, action)
		}

		return provider, nil
	}

	// Get all providers that support the action
	candidates := pm.GetProvidersForAction(action)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no providers found that support action %s", action)
	}

	// Filter to only available providers
	var availableCandidates []*types.ProviderData
	for _, provider := range candidates {
		if pm.detector.IsAvailable(provider) {
			availableCandidates = append(availableCandidates, provider)
		}
	}

	if len(availableCandidates) == 0 {
		return nil, fmt.Errorf("no available providers found that support action %s", action)
	}

	// Sort by effective priority
	sort.Slice(availableCandidates, func(i, j int) bool {
		priorityI := pm.getEffectivePriority(availableCandidates[i])
		priorityJ := pm.getEffectivePriority(availableCandidates[j])
		return priorityI > priorityJ
	})

	// Return the highest priority provider
	return availableCandidates[0], nil
}

// IsProviderAvailable checks if a provider is available on the system
func (pm *ProviderManager) IsProviderAvailable(name string) bool {
	provider, err := pm.GetProvider(name)
	if err != nil {
		return false
	}

	return pm.detector.IsAvailable(provider)
}

// GetProvidersForAction returns providers that support a specific action
func (pm *ProviderManager) GetProvidersForAction(action string) []*types.ProviderData {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var supportingProviders []*types.ProviderData
	for _, provider := range pm.providers {
		if pm.detector.SupportsAction(provider, action) {
			supportingProviders = append(supportingProviders, provider)
		}
	}

	return supportingProviders
}

// ValidateProvider validates a provider configuration
func (pm *ProviderManager) ValidateProvider(provider *types.ProviderData) error {
	return pm.loader.ValidateProvider(provider)
}

// ReloadProviders reloads all providers (useful for development)
func (pm *ProviderManager) ReloadProviders() error {
	return pm.LoadProviders(pm.config.ProviderDirectory)
}

// GetProviderSelections returns provider options for user selection (Requirement 1.3)
func (pm *ProviderManager) GetProviderSelections(software string, action string) ([]*ProviderSelection, error) {
	candidates := pm.GetProvidersForAction(action)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no providers found that support action %s", action)
	}

	var selections []*ProviderSelection
	for _, provider := range candidates {
		selection := &ProviderSelection{
			Provider:    provider,
			Priority:    pm.getEffectivePriority(provider),
			Available:   pm.detector.IsAvailable(provider),
			PackageName: pm.getPackageName(provider, software),
			Version:     pm.getProviderVersion(provider),
			IsInstalled: false, // TODO: Implement installation detection
		}

		// Set reason for availability/unavailability
		if selection.Available {
			selection.Reason = "Available"
		} else {
			if result, exists := pm.detector.GetCachedResult(provider.Provider.Name); exists && result.Error != nil {
				selection.Reason = result.Error.Error()
			} else {
				selection.Reason = "Not available"
			}
		}

		selections = append(selections, selection)
	}

	// Sort by priority (available providers first, then by priority)
	sort.Slice(selections, func(i, j int) bool {
		if selections[i].Available != selections[j].Available {
			return selections[i].Available // Available providers first
		}
		return selections[i].Priority > selections[j].Priority
	})

	return selections, nil
}

// GetMultipleProviderOptions returns multiple provider options when more than one is available
func (pm *ProviderManager) GetMultipleProviderOptions(software string, action string) ([]*interfaces.ProviderOption, error) {
	selections, err := pm.GetProviderSelections(software, action)
	if err != nil {
		return nil, err
	}

	// Filter to only available providers
	var availableSelections []*ProviderSelection
	for _, selection := range selections {
		if selection.Available {
			availableSelections = append(availableSelections, selection)
		}
	}

	if len(availableSelections) == 0 {
		return nil, fmt.Errorf("no available providers found for action %s", action)
	}

	// Convert to interface format
	var options []*interfaces.ProviderOption
	for _, selection := range availableSelections {
		option := &interfaces.ProviderOption{
			Provider:    selection.Provider,
			PackageName: selection.PackageName,
			Version:     selection.Version,
			IsInstalled: selection.IsInstalled,
			Priority:    selection.Priority,
		}
		options = append(options, option)
	}

	return options, nil
}

// SelectProviderWithFallback selects a provider with automatic fallback
func (pm *ProviderManager) SelectProviderWithFallback(software string, action string, preferredProvider string) (*types.ProviderData, error) {
	// Try preferred provider first
	if preferredProvider != "" {
		provider, err := pm.SelectProvider(software, action, preferredProvider)
		if err == nil {
			return provider, nil
		}
		// Log warning but continue with fallback
		fmt.Printf("Warning: preferred provider %s failed: %v, trying fallback\n", preferredProvider, err)
	}

	// Try default provider from config
	if pm.config.DefaultProvider != "" && pm.config.DefaultProvider != preferredProvider {
		provider, err := pm.SelectProvider(software, action, pm.config.DefaultProvider)
		if err == nil {
			return provider, nil
		}
	}

	// Fallback to automatic selection
	return pm.SelectProvider(software, action, "")
}

// GetProviderStats returns statistics about loaded providers
func (pm *ProviderManager) GetProviderStats() *ProviderStats {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	stats := &ProviderStats{
		TotalProviders:     len(pm.providers),
		AvailableProviders: 0,
		ProvidersByType:    make(map[string]int),
		ProvidersByPlatform: make(map[string]int),
	}

	for _, provider := range pm.providers {
		// Count by type
		stats.ProvidersByType[provider.Provider.Type]++

		// Count by platform
		for _, platform := range provider.Provider.Platforms {
			stats.ProvidersByPlatform[platform]++
		}

		// Count available providers
		if pm.detector.IsAvailable(provider) {
			stats.AvailableProviders++
		}
	}

	return stats
}

// getEffectivePriority calculates the effective priority for a provider
func (pm *ProviderManager) getEffectivePriority(provider *types.ProviderData) int {
	// Start with detector priority (includes platform compatibility boost)
	priority := pm.detector.GetProviderPriority(provider)

	// Apply config-based priority override
	if configPriority, exists := pm.config.ProviderPriority[provider.Provider.Name]; exists {
		priority = configPriority
	}

	return priority
}

// getPackageName attempts to get the package name for a software from provider
func (pm *ProviderManager) getPackageName(provider *types.ProviderData, software string) string {
	// TODO: This would need saidata integration to get actual package names
	// For now, return the software name as a placeholder
	return software
}

// getProviderVersion gets version information for a provider
func (pm *ProviderManager) getProviderVersion(provider *types.ProviderData) string {
	if result, exists := pm.detector.GetCachedResult(provider.Provider.Name); exists {
		return result.Version
	}
	return "unknown"
}

// setupWatching sets up file watching for provider changes
func (pm *ProviderManager) setupWatching() error {
	return pm.loader.WatchDirectory(pm.config.ProviderDirectory, func(provider *types.ProviderData) {
		pm.mutex.Lock()
		defer pm.mutex.Unlock()
		
		// Update the provider in our map
		pm.providers[provider.Provider.Name] = provider
		fmt.Printf("Provider %s reloaded\n", provider.Provider.Name)
	})
}

// Close cleans up resources used by the provider manager
func (pm *ProviderManager) Close() error {
	return pm.loader.StopAllWatching()
}

// ProviderStats contains statistics about loaded providers
type ProviderStats struct {
	TotalProviders      int
	AvailableProviders  int
	ProvidersByType     map[string]int
	ProvidersByPlatform map[string]int
}

// String returns a string representation of provider stats
func (ps *ProviderStats) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Total: %d", ps.TotalProviders))
	parts = append(parts, fmt.Sprintf("Available: %d", ps.AvailableProviders))
	
	if len(ps.ProvidersByType) > 0 {
		var types []string
		for pType, count := range ps.ProvidersByType {
			types = append(types, fmt.Sprintf("%s: %d", pType, count))
		}
		parts = append(parts, fmt.Sprintf("Types: {%s}", strings.Join(types, ", ")))
	}
	
	return strings.Join(parts, ", ")
}

// Ensure ProviderManager implements the interface
var _ interfaces.ProviderManager = (*ProviderManager)(nil)