package errors

import (
	"context"
	"fmt"
	"sync"
	"time"

	"sai/internal/interfaces"
	"sai/internal/types"
)

// DegradationManager handles graceful degradation when providers are unavailable
type DegradationManager struct {
	providerManager interfaces.ProviderManager
	saidataManager  interfaces.SaidataManager
	logger          interfaces.Logger
	
	// Provider health tracking
	providerHealth map[string]*ProviderHealth
	healthMutex    sync.RWMutex
	
	// Degradation policies
	degradationPolicies map[string]*DegradationPolicy
	policyMutex         sync.RWMutex
}

// ProviderHealth tracks the health status of a provider
type ProviderHealth struct {
	Name            string
	Available       bool
	LastCheck       time.Time
	FailureCount    int
	LastFailure     time.Time
	LastSuccess     time.Time
	ConsecutiveFails int
	HealthScore     float64 // 0.0 to 1.0
}

// DegradationPolicy defines how to handle degradation for different scenarios
type DegradationPolicy struct {
	// Fallback strategies
	FallbackProviders []string `yaml:"fallback_providers"`
	UseDefaults       bool     `yaml:"use_defaults"`
	AllowPartial      bool     `yaml:"allow_partial"`
	
	// Thresholds
	MaxFailures       int           `yaml:"max_failures"`
	HealthThreshold   float64       `yaml:"health_threshold"`
	RecoveryTime      time.Duration `yaml:"recovery_time"`
	
	// Actions
	DisableOnFailure  bool `yaml:"disable_on_failure"`
	NotifyOnDegradation bool `yaml:"notify_on_degradation"`
}

// DegradationResult represents the result of a degradation attempt
type DegradationResult struct {
	Success           bool
	Strategy          string
	FallbackProvider  string
	UsedDefaults      bool
	PartialSuccess    bool
	AvailableFeatures []string
	UnavailableFeatures []string
	Warnings          []string
	Error             error
}

// NewDegradationManager creates a new degradation manager
func NewDegradationManager(
	providerManager interfaces.ProviderManager,
	saidataManager interfaces.SaidataManager,
	logger interfaces.Logger,
) *DegradationManager {
	dm := &DegradationManager{
		providerManager:     providerManager,
		saidataManager:      saidataManager,
		logger:              logger,
		providerHealth:      make(map[string]*ProviderHealth),
		degradationPolicies: make(map[string]*DegradationPolicy),
	}
	
	// Set up default degradation policies
	dm.setupDefaultPolicies()
	
	return dm
}

// HandleProviderUnavailable handles scenarios where providers are unavailable
func (dm *DegradationManager) HandleProviderUnavailable(
	ctx context.Context,
	action string,
	software string,
	unavailableProviders []string,
) (*DegradationResult, error) {
	dm.logger.Info("Handling provider unavailability",
		interfaces.LogField{Key: "action", Value: action},
		interfaces.LogField{Key: "software", Value: software},
		interfaces.LogField{Key: "unavailable_providers", Value: unavailableProviders},
	)
	
	result := &DegradationResult{
		Strategy: "provider_fallback",
	}
	
	// Update provider health status
	for _, provider := range unavailableProviders {
		dm.updateProviderHealth(provider, false, fmt.Errorf("provider unavailable"))
	}
	
	// Get degradation policy for this action
	policy := dm.getDegradationPolicy(action)
	
	// Try fallback providers
	if len(policy.FallbackProviders) > 0 {
		fallbackResult, err := dm.tryFallbackProviders(ctx, action, software, policy.FallbackProviders)
		if err == nil && fallbackResult.Success {
			result.Success = true
			result.FallbackProvider = fallbackResult.FallbackProvider
			return result, nil
		}
		result.Warnings = append(result.Warnings, "All fallback providers failed")
	}
	
	// Try using intelligent defaults
	if policy.UseDefaults {
		defaultsResult, err := dm.useIntelligentDefaults(ctx, action, software)
		if err == nil && defaultsResult.Success {
			result.Success = true
			result.UsedDefaults = true
			result.Strategy = "intelligent_defaults"
			result.Warnings = append(result.Warnings, "Using intelligent defaults due to provider unavailability")
			return result, nil
		}
		result.Warnings = append(result.Warnings, "Intelligent defaults failed")
	}
	
	// Try partial functionality
	if policy.AllowPartial {
		partialResult, err := dm.enablePartialFunctionality(ctx, action, software)
		if err == nil && partialResult.PartialSuccess {
			result.Success = true
			result.PartialSuccess = true
			result.Strategy = "partial_functionality"
			result.AvailableFeatures = partialResult.AvailableFeatures
			result.UnavailableFeatures = partialResult.UnavailableFeatures
			result.Warnings = append(result.Warnings, "Operating with reduced functionality")
			return result, nil
		}
		result.Warnings = append(result.Warnings, "Partial functionality not available")
	}
	
	// All degradation strategies failed
	result.Success = false
	result.Error = NewProviderUnavailableError("all", "no degradation strategies succeeded").
		WithSuggestion("Install additional package managers").
		WithSuggestion("Check system requirements").
		WithSuggestion("Try again later")
	
	return result, result.Error
}

// UpdateProviderHealth updates the health status of a provider
func (dm *DegradationManager) UpdateProviderHealth(providerName string, success bool, err error) {
	dm.updateProviderHealth(providerName, success, err)
}

// GetProviderHealth returns the health status of a provider
func (dm *DegradationManager) GetProviderHealth(providerName string) *ProviderHealth {
	dm.healthMutex.RLock()
	defer dm.healthMutex.RUnlock()
	
	if health, exists := dm.providerHealth[providerName]; exists {
		return health
	}
	
	// Return default health for unknown providers
	return &ProviderHealth{
		Name:        providerName,
		Available:   true,
		LastCheck:   time.Now(),
		HealthScore: 1.0,
	}
}

// GetAllProviderHealth returns health status for all tracked providers
func (dm *DegradationManager) GetAllProviderHealth() map[string]*ProviderHealth {
	dm.healthMutex.RLock()
	defer dm.healthMutex.RUnlock()
	
	health := make(map[string]*ProviderHealth)
	for name, providerHealth := range dm.providerHealth {
		// Create a copy to avoid race conditions
		health[name] = &ProviderHealth{
			Name:             providerHealth.Name,
			Available:        providerHealth.Available,
			LastCheck:        providerHealth.LastCheck,
			FailureCount:     providerHealth.FailureCount,
			LastFailure:      providerHealth.LastFailure,
			LastSuccess:      providerHealth.LastSuccess,
			ConsecutiveFails: providerHealth.ConsecutiveFails,
			HealthScore:      providerHealth.HealthScore,
		}
	}
	
	return health
}

// SetDegradationPolicy sets a degradation policy for a specific action
func (dm *DegradationManager) SetDegradationPolicy(action string, policy *DegradationPolicy) {
	dm.policyMutex.Lock()
	defer dm.policyMutex.Unlock()
	
	dm.degradationPolicies[action] = policy
}

// Helper methods

func (dm *DegradationManager) updateProviderHealth(providerName string, success bool, err error) {
	dm.healthMutex.Lock()
	defer dm.healthMutex.Unlock()
	
	health, exists := dm.providerHealth[providerName]
	if !exists {
		health = &ProviderHealth{
			Name:        providerName,
			Available:   true,
			HealthScore: 1.0,
		}
		dm.providerHealth[providerName] = health
	}
	
	health.LastCheck = time.Now()
	
	if success {
		health.Available = true
		health.LastSuccess = time.Now()
		health.ConsecutiveFails = 0
		
		// Improve health score
		health.HealthScore = min(1.0, health.HealthScore+0.1)
		
		dm.logger.Debug("Provider health improved",
			interfaces.LogField{Key: "provider", Value: providerName},
			interfaces.LogField{Key: "health_score", Value: health.HealthScore},
		)
	} else {
		health.FailureCount++
		health.ConsecutiveFails++
		health.LastFailure = time.Now()
		
		// Degrade health score
		health.HealthScore = max(0.0, health.HealthScore-0.2)
		
		// Mark as unavailable if too many consecutive failures
		policy := dm.getDegradationPolicy("default")
		if health.ConsecutiveFails >= policy.MaxFailures {
			health.Available = false
		}
		
		dm.logger.Warn("Provider health degraded",
			interfaces.LogField{Key: "provider", Value: providerName},
			interfaces.LogField{Key: "consecutive_fails", Value: health.ConsecutiveFails},
			interfaces.LogField{Key: "health_score", Value: health.HealthScore},
			interfaces.LogField{Key: "available", Value: health.Available},
		)
		
		if err != nil {
			dm.logger.Debug("Provider failure details",
				interfaces.LogField{Key: "provider", Value: providerName},
				interfaces.LogField{Key: "error", Value: err.Error()},
			)
		}
	}
}

func (dm *DegradationManager) getDegradationPolicy(action string) *DegradationPolicy {
	dm.policyMutex.RLock()
	defer dm.policyMutex.RUnlock()
	
	if policy, exists := dm.degradationPolicies[action]; exists {
		return policy
	}
	
	// Return default policy
	return dm.degradationPolicies["default"]
}

func (dm *DegradationManager) tryFallbackProviders(
	ctx context.Context,
	action string,
	software string,
	fallbackProviders []string,
) (*DegradationResult, error) {
	for _, providerName := range fallbackProviders {
		// Check if this provider is available
		if !dm.providerManager.IsProviderAvailable(providerName) {
			continue
		}
		
		// Check provider health
		health := dm.GetProviderHealth(providerName)
		if !health.Available || health.HealthScore < 0.5 {
			continue
		}
		
		dm.logger.Debug("Trying fallback provider",
			interfaces.LogField{Key: "provider", Value: providerName},
			interfaces.LogField{Key: "action", Value: action},
			interfaces.LogField{Key: "software", Value: software},
		)
		
		// Try to get the provider
		provider, err := dm.providerManager.GetProvider(providerName)
		if err != nil {
			dm.updateProviderHealth(providerName, false, err)
			continue
		}
		
		// Check if provider supports the action
		if _, hasAction := provider.Actions[action]; !hasAction {
			continue
		}
		
		// This provider can be used as fallback
		dm.updateProviderHealth(providerName, true, nil)
		return &DegradationResult{
			Success:          true,
			FallbackProvider: providerName,
		}, nil
	}
	
	return &DegradationResult{Success: false}, fmt.Errorf("no fallback providers available")
}

func (dm *DegradationManager) useIntelligentDefaults(
	ctx context.Context,
	action string,
	software string,
) (*DegradationResult, error) {
	dm.logger.Debug("Attempting to use intelligent defaults",
		interfaces.LogField{Key: "action", Value: action},
		interfaces.LogField{Key: "software", Value: software},
	)
	
	// Generate intelligent defaults
	defaults, err := dm.saidataManager.GenerateDefaults(software)
	if err != nil {
		return &DegradationResult{Success: false}, err
	}
	
	// Check if defaults provide enough information for the action
	if dm.canExecuteWithDefaults(action, defaults) {
		return &DegradationResult{
			Success:      true,
			UsedDefaults: true,
		}, nil
	}
	
	return &DegradationResult{Success: false}, fmt.Errorf("intelligent defaults insufficient for action %s", action)
}

func (dm *DegradationManager) enablePartialFunctionality(
	ctx context.Context,
	action string,
	software string,
) (*DegradationResult, error) {
	dm.logger.Debug("Attempting partial functionality",
		interfaces.LogField{Key: "action", Value: action},
		interfaces.LogField{Key: "software", Value: software},
	)
	
	// Get available providers
	availableProviders := dm.providerManager.GetAvailableProviders()
	
	var availableFeatures []string
	var unavailableFeatures []string
	
	// Determine what features are available with current providers
	featureMap := map[string][]string{
		"install":   {"package_install"},
		"uninstall": {"package_uninstall"},
		"start":     {"service_start"},
		"stop":      {"service_stop"},
		"status":    {"service_status", "package_status"},
		"search":    {"package_search"},
		"info":      {"package_info"},
	}
	
	if features, exists := featureMap[action]; exists {
		for _, feature := range features {
			if dm.isFeatureAvailable(feature, availableProviders) {
				availableFeatures = append(availableFeatures, feature)
			} else {
				unavailableFeatures = append(unavailableFeatures, feature)
			}
		}
	}
	
	// If at least some features are available, consider it partial success
	if len(availableFeatures) > 0 {
		return &DegradationResult{
			Success:             false, // Not full success
			PartialSuccess:      true,
			AvailableFeatures:   availableFeatures,
			UnavailableFeatures: unavailableFeatures,
		}, nil
	}
	
	return &DegradationResult{Success: false}, fmt.Errorf("no features available for partial functionality")
}

func (dm *DegradationManager) canExecuteWithDefaults(action string, defaults *types.SoftwareData) bool {
	// Check if defaults provide enough information for the action
	switch action {
	case "install", "uninstall":
		return len(defaults.Packages) > 0
	case "start", "stop", "restart", "status":
		return len(defaults.Services) > 0
	case "config":
		return len(defaults.Files) > 0
	default:
		return true // Most actions can work with basic defaults
	}
}

func (dm *DegradationManager) isFeatureAvailable(feature string, providers []*types.ProviderData) bool {
	for _, provider := range providers {
		// Check if provider supports the feature
		for _, capability := range provider.Provider.Capabilities {
			if capability == feature {
				return true
			}
		}
	}
	return false
}

func (dm *DegradationManager) setupDefaultPolicies() {
	defaultPolicies := map[string]*DegradationPolicy{
		"default": {
			FallbackProviders:   []string{},
			UseDefaults:         true,
			AllowPartial:        true,
			MaxFailures:         3,
			HealthThreshold:     0.5,
			RecoveryTime:        5 * time.Minute,
			DisableOnFailure:    false,
			NotifyOnDegradation: true,
		},
		"install": {
			FallbackProviders:   []string{"apt", "brew", "dnf", "yum"},
			UseDefaults:         true,
			AllowPartial:        false,
			MaxFailures:         2,
			HealthThreshold:     0.7,
			RecoveryTime:        10 * time.Minute,
			DisableOnFailure:    false,
			NotifyOnDegradation: true,
		},
		"start": {
			FallbackProviders:   []string{},
			UseDefaults:         true,
			AllowPartial:        true,
			MaxFailures:         3,
			HealthThreshold:     0.6,
			RecoveryTime:        2 * time.Minute,
			DisableOnFailure:    false,
			NotifyOnDegradation: false,
		},
		"search": {
			FallbackProviders:   []string{"apt", "brew", "dnf"},
			UseDefaults:         false,
			AllowPartial:        true,
			MaxFailures:         5,
			HealthThreshold:     0.3,
			RecoveryTime:        1 * time.Minute,
			DisableOnFailure:    false,
			NotifyOnDegradation: false,
		},
	}
	
	for action, policy := range defaultPolicies {
		dm.SetDegradationPolicy(action, policy)
	}
	
	dm.logger.Info("Default degradation policies configured",
		interfaces.LogField{Key: "policies_count", Value: len(defaultPolicies)},
	)
}

// Utility functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}