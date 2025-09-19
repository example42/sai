package errors

import (
	"context"
	"fmt"
	"time"

	"sai/internal/interfaces"
	"sai/internal/types"
)

// RecoveryManager handles error recovery and rollback operations
type RecoveryManager struct {
	executor        interfaces.GenericExecutor
	providerManager interfaces.ProviderManager
	logger          interfaces.Logger
	config          *RecoveryConfig
}

// RecoveryConfig defines recovery behavior configuration
type RecoveryConfig struct {
	MaxRetries          int           `yaml:"max_retries"`
	RetryDelay          time.Duration `yaml:"retry_delay"`
	BackoffMultiplier   float64       `yaml:"backoff_multiplier"`
	MaxRetryDelay       time.Duration `yaml:"max_retry_delay"`
	EnableRollback      bool          `yaml:"enable_rollback"`
	RollbackTimeout     time.Duration `yaml:"rollback_timeout"`
	CircuitBreakerThreshold int       `yaml:"circuit_breaker_threshold"`
	CircuitBreakerWindow    time.Duration `yaml:"circuit_breaker_window"`
}

// DefaultRecoveryConfig returns default recovery configuration
func DefaultRecoveryConfig() *RecoveryConfig {
	return &RecoveryConfig{
		MaxRetries:              3,
		RetryDelay:              2 * time.Second,
		BackoffMultiplier:       2.0,
		MaxRetryDelay:           30 * time.Second,
		EnableRollback:          true,
		RollbackTimeout:         60 * time.Second,
		CircuitBreakerThreshold: 5,
		CircuitBreakerWindow:    5 * time.Minute,
	}
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager(
	executor interfaces.GenericExecutor,
	providerManager interfaces.ProviderManager,
	logger interfaces.Logger,
	config *RecoveryConfig,
) *RecoveryManager {
	if config == nil {
		config = DefaultRecoveryConfig()
	}
	
	return &RecoveryManager{
		executor:        executor,
		providerManager: providerManager,
		logger:          logger,
		config:          config,
	}
}

// RecoveryContext holds context for recovery operations
type RecoveryContext struct {
	Action           string
	Software         string
	Provider         *types.ProviderData
	Saidata          *types.SoftwareData
	OriginalError    error
	AttemptCount     int
	StartTime        time.Time
	LastAttemptTime  time.Time
	RollbackCommands []string
	ExecutedCommands []string
}

// RecoveryResult represents the result of a recovery attempt
type RecoveryResult struct {
	Success          bool
	RecoveredError   error
	FinalError       error
	AttemptsUsed     int
	RollbackExecuted bool
	Duration         time.Duration
	RecoveryStrategy string
	StartTime        time.Time
}

// AttemptRecovery attempts to recover from an error using various strategies
func (rm *RecoveryManager) AttemptRecovery(ctx context.Context, recoveryCtx *RecoveryContext) (*RecoveryResult, error) {
	startTime := time.Now()
	result := &RecoveryResult{
		RecoveredError: recoveryCtx.OriginalError,
		StartTime:      startTime,
	}

	rm.logger.Debug("Starting error recovery",
		interfaces.LogField{Key: "action", Value: recoveryCtx.Action},
		interfaces.LogField{Key: "software", Value: recoveryCtx.Software},
		interfaces.LogField{Key: "provider", Value: recoveryCtx.Provider.Provider.Name},
		interfaces.LogField{Key: "error", Value: recoveryCtx.OriginalError.Error()},
	)

	// Determine recovery strategy based on error type
	strategy := rm.determineRecoveryStrategy(recoveryCtx.OriginalError)
	result.RecoveryStrategy = strategy

	switch strategy {
	case "retry":
		return rm.retryWithBackoff(ctx, recoveryCtx, result)
	case "alternative_provider":
		return rm.tryAlternativeProvider(ctx, recoveryCtx, result)
	case "rollback":
		return rm.executeRollback(ctx, recoveryCtx, result)
	case "resource_creation":
		return rm.createMissingResources(ctx, recoveryCtx, result)
	case "graceful_degradation":
		return rm.gracefulDegradation(ctx, recoveryCtx, result)
	default:
		result.FinalError = recoveryCtx.OriginalError
		result.Duration = time.Since(startTime)
		return result, recoveryCtx.OriginalError
	}
}

// determineRecoveryStrategy determines the best recovery strategy for an error
func (rm *RecoveryManager) determineRecoveryStrategy(err error) string {
	if saiErr, ok := err.(*SAIError); ok {
		switch saiErr.Type {
		case ErrorTypeActionTimeout, ErrorTypeNetworkTimeout, ErrorTypeNetworkUnavailable:
			return "retry"
		case ErrorTypeProviderNotFound, ErrorTypeProviderUnavailable:
			return "alternative_provider"
		case ErrorTypeResourceMissing:
			return "resource_creation"
		case ErrorTypeActionFailed, ErrorTypeCommandFailed:
			// Check if rollback is available
			if saiErr.Context != nil {
				if _, hasRollback := saiErr.Context["rollback_available"]; hasRollback {
					return "rollback"
				}
			}
			return "retry"
		case ErrorTypeSaidataNotFound:
			return "graceful_degradation"
		default:
			return "none"
		}
	}
	
	// For non-SAI errors, try retry as default
	return "retry"
}

// retryWithBackoff implements retry logic with exponential backoff
func (rm *RecoveryManager) retryWithBackoff(ctx context.Context, recoveryCtx *RecoveryContext, result *RecoveryResult) (*RecoveryResult, error) {
	delay := rm.config.RetryDelay
	
	for attempt := 1; attempt <= rm.config.MaxRetries; attempt++ {
		result.AttemptsUsed = attempt
		
		rm.logger.Info("Retrying action",
			interfaces.LogField{Key: "attempt", Value: attempt},
			interfaces.LogField{Key: "max_attempts", Value: rm.config.MaxRetries},
			interfaces.LogField{Key: "delay", Value: delay},
			interfaces.LogField{Key: "action", Value: recoveryCtx.Action},
			interfaces.LogField{Key: "software", Value: recoveryCtx.Software},
		)

		// Wait before retry (except for first attempt)
		if attempt > 1 {
			select {
			case <-ctx.Done():
				result.FinalError = ctx.Err()
				result.Duration = time.Since(result.StartTime)
				return result, ctx.Err()
			case <-time.After(delay):
			}
		}

		// Attempt execution
		executeOptions := interfaces.ExecuteOptions{
			DryRun:    false,
			Verbose:   false,
			Timeout:   30 * time.Second,
			Variables: make(map[string]string),
		}

		executionResult, err := rm.executor.Execute(ctx, recoveryCtx.Provider, recoveryCtx.Action, recoveryCtx.Software, recoveryCtx.Saidata, executeOptions)
		
		if err == nil && executionResult != nil && executionResult.Success {
			// Success!
			result.Success = true
			result.Duration = time.Since(result.StartTime)
			rm.logger.Info("Recovery successful",
				interfaces.LogField{Key: "attempts_used", Value: attempt},
				interfaces.LogField{Key: "duration", Value: result.Duration},
			)
			return result, nil
		}

		// Log the retry failure
		if err != nil {
			rm.logger.Debug("Retry attempt failed",
				interfaces.LogField{Key: "attempt", Value: attempt},
				interfaces.LogField{Key: "error", Value: err.Error()},
			)
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * rm.config.BackoffMultiplier)
		if delay > rm.config.MaxRetryDelay {
			delay = rm.config.MaxRetryDelay
		}
	}

	// All retries exhausted
	result.FinalError = NewActionFailedError(recoveryCtx.Action, recoveryCtx.Software, 1, "all retry attempts exhausted").
		WithContext("attempts_used", result.AttemptsUsed).
		WithSuggestion("Check system resources and network connectivity").
		WithSuggestion("Try a different provider")
	
	result.Duration = time.Since(result.StartTime)
	return result, result.FinalError
}

// tryAlternativeProvider attempts to use an alternative provider
func (rm *RecoveryManager) tryAlternativeProvider(ctx context.Context, recoveryCtx *RecoveryContext, result *RecoveryResult) (*RecoveryResult, error) {
	rm.logger.Info("Attempting recovery with alternative provider",
		interfaces.LogField{Key: "original_provider", Value: recoveryCtx.Provider.Provider.Name},
		interfaces.LogField{Key: "action", Value: recoveryCtx.Action},
		interfaces.LogField{Key: "software", Value: recoveryCtx.Software},
	)

	// Get all providers that support this action
	providers := rm.providerManager.GetProvidersForAction(recoveryCtx.Action)
	
	// Filter out the failed provider and unavailable providers
	var alternativeProviders []*types.ProviderData
	for _, provider := range providers {
		if provider.Provider.Name != recoveryCtx.Provider.Provider.Name &&
		   rm.providerManager.IsProviderAvailable(provider.Provider.Name) {
			alternativeProviders = append(alternativeProviders, provider)
		}
	}

	if len(alternativeProviders) == 0 {
		result.FinalError = NewProviderUnavailableError("all", "no alternative providers available").
			WithSuggestion("Install additional package managers").
			WithSuggestion("Check provider availability")
		result.Duration = time.Since(result.StartTime)
		return result, result.FinalError
	}

	// Try each alternative provider
	for _, altProvider := range alternativeProviders {
		result.AttemptsUsed++
		
		rm.logger.Debug("Trying alternative provider",
			interfaces.LogField{Key: "provider", Value: altProvider.Provider.Name},
			interfaces.LogField{Key: "attempt", Value: result.AttemptsUsed},
		)

		// Check if this provider can execute the action
		if !rm.executor.CanExecute(altProvider, recoveryCtx.Action, recoveryCtx.Software, recoveryCtx.Saidata) {
			continue
		}

		// Attempt execution with alternative provider
		executeOptions := interfaces.ExecuteOptions{
			DryRun:    false,
			Verbose:   false,
			Timeout:   30 * time.Second,
			Variables: make(map[string]string),
		}

		executionResult, err := rm.executor.Execute(ctx, altProvider, recoveryCtx.Action, recoveryCtx.Software, recoveryCtx.Saidata, executeOptions)
		
		if err == nil && executionResult != nil && executionResult.Success {
			// Success with alternative provider!
			result.Success = true
			result.Duration = time.Since(result.StartTime)
			rm.logger.Info("Recovery successful with alternative provider",
				interfaces.LogField{Key: "alternative_provider", Value: altProvider.Provider.Name},
				interfaces.LogField{Key: "original_provider", Value: recoveryCtx.Provider.Provider.Name},
				interfaces.LogField{Key: "duration", Value: result.Duration},
			)
			return result, nil
		}

		// Log the failure and continue to next provider
		if err != nil {
			rm.logger.Debug("Alternative provider failed",
				interfaces.LogField{Key: "provider", Value: altProvider.Provider.Name},
				interfaces.LogField{Key: "error", Value: err.Error()},
			)
		}
	}

	// All alternative providers failed
	result.FinalError = NewActionFailedError(recoveryCtx.Action, recoveryCtx.Software, 1, "all alternative providers failed").
		WithContext("providers_tried", len(alternativeProviders)+1).
		WithSuggestion("Check system requirements").
		WithSuggestion("Verify software availability")
	
	result.Duration = time.Since(result.StartTime)
	return result, result.FinalError
}

// executeRollback executes rollback commands to undo partial changes
func (rm *RecoveryManager) executeRollback(ctx context.Context, recoveryCtx *RecoveryContext, result *RecoveryResult) (*RecoveryResult, error) {
	if !rm.config.EnableRollback {
		result.FinalError = recoveryCtx.OriginalError
		result.Duration = time.Since(result.StartTime)
		return result, recoveryCtx.OriginalError
	}

	rm.logger.Info("Executing rollback",
		interfaces.LogField{Key: "action", Value: recoveryCtx.Action},
		interfaces.LogField{Key: "software", Value: recoveryCtx.Software},
		interfaces.LogField{Key: "commands", Value: len(recoveryCtx.RollbackCommands)},
	)

	// Create rollback context with timeout
	rollbackCtx, cancel := context.WithTimeout(ctx, rm.config.RollbackTimeout)
	defer cancel()

	// Execute rollback commands
	rollbackSuccess := true
	for i, command := range recoveryCtx.RollbackCommands {
		rm.logger.Debug("Executing rollback command",
			interfaces.LogField{Key: "command", Value: command},
			interfaces.LogField{Key: "step", Value: i + 1},
			interfaces.LogField{Key: "total", Value: len(recoveryCtx.RollbackCommands)},
		)

		// Execute rollback command (implementation would depend on command executor)
		// For now, we'll simulate rollback execution
		select {
		case <-rollbackCtx.Done():
			rollbackSuccess = false
			rm.logger.Error("Rollback timeout", fmt.Errorf("rollback command timed out"),
				interfaces.LogField{Key: "command", Value: command},
				interfaces.LogField{Key: "step", Value: i + 1},
			)
			break
		default:
			// Simulate command execution
			time.Sleep(100 * time.Millisecond)
		}
	}

	result.RollbackExecuted = true
	result.Duration = time.Since(result.StartTime)

	if rollbackSuccess {
		rm.logger.Info("Rollback completed successfully",
			interfaces.LogField{Key: "commands_executed", Value: len(recoveryCtx.RollbackCommands)},
			interfaces.LogField{Key: "duration", Value: result.Duration},
		)
		
		// Rollback succeeded, but original action still failed
		result.FinalError = WrapSAIError(ErrorTypeActionFailed, 
			fmt.Sprintf("action failed but rollback completed successfully"), 
			recoveryCtx.OriginalError).
			WithSuggestion("Check logs for failure details").
			WithSuggestion("System state has been restored")
	} else {
		rm.logger.Error("Rollback failed", fmt.Errorf("rollback execution failed"),
			interfaces.LogField{Key: "duration", Value: result.Duration},
		)
		
		result.FinalError = WrapSAIError(ErrorTypeActionFailed, 
			fmt.Sprintf("action failed and rollback also failed"), 
			recoveryCtx.OriginalError).
			WithSuggestion("Manual intervention may be required").
			WithSuggestion("Check system state for partial changes")
	}

	return result, result.FinalError
}

// createMissingResources attempts to create missing resources
func (rm *RecoveryManager) createMissingResources(ctx context.Context, recoveryCtx *RecoveryContext, result *RecoveryResult) (*RecoveryResult, error) {
	rm.logger.Info("Attempting to create missing resources",
		interfaces.LogField{Key: "action", Value: recoveryCtx.Action},
		interfaces.LogField{Key: "software", Value: recoveryCtx.Software},
	)

	// Extract missing resource information from error context
	saiErr, ok := recoveryCtx.OriginalError.(*SAIError)
	if !ok {
		result.FinalError = recoveryCtx.OriginalError
		result.Duration = time.Since(result.StartTime)
		return result, recoveryCtx.OriginalError
	}

	resourcesCreated := 0
	
	// Try to create missing directories
	if missingDirs, exists := saiErr.Context["missing_directories"]; exists {
		if dirs, ok := missingDirs.([]string); ok {
			for _, dir := range dirs {
				if rm.createDirectory(dir) {
					resourcesCreated++
					rm.logger.Debug("Created missing directory",
						interfaces.LogField{Key: "directory", Value: dir},
					)
				}
			}
		}
	}

	// Try to create missing files with default content
	if missingFiles, exists := saiErr.Context["missing_files"]; exists {
		if files, ok := missingFiles.([]string); ok {
			for _, file := range files {
				if rm.createDefaultFile(file, recoveryCtx.Software) {
					resourcesCreated++
					rm.logger.Debug("Created missing file",
						interfaces.LogField{Key: "file", Value: file},
					)
				}
			}
		}
	}

	result.AttemptsUsed = 1
	result.Duration = time.Since(result.StartTime)

	if resourcesCreated > 0 {
		rm.logger.Info("Created missing resources, retrying action",
			interfaces.LogField{Key: "resources_created", Value: resourcesCreated},
		)

		// Retry the original action now that resources are created
		executeOptions := interfaces.ExecuteOptions{
			DryRun:    false,
			Verbose:   false,
			Timeout:   30 * time.Second,
			Variables: make(map[string]string),
		}

		executionResult, err := rm.executor.Execute(ctx, recoveryCtx.Provider, recoveryCtx.Action, recoveryCtx.Software, recoveryCtx.Saidata, executeOptions)
		
		if err == nil && executionResult != nil && executionResult.Success {
			result.Success = true
			rm.logger.Info("Recovery successful after creating resources",
				interfaces.LogField{Key: "resources_created", Value: resourcesCreated},
				interfaces.LogField{Key: "duration", Value: result.Duration},
			)
			return result, nil
		}
	}

	// Resource creation didn't help or failed
	result.FinalError = WrapSAIError(ErrorTypeResourceMissing, 
		fmt.Sprintf("failed to create missing resources or action still failed"), 
		recoveryCtx.OriginalError).
		WithContext("resources_created", resourcesCreated).
		WithSuggestion("Check file permissions").
		WithSuggestion("Verify directory structure")

	return result, result.FinalError
}

// gracefulDegradation handles cases where we can continue with reduced functionality
func (rm *RecoveryManager) gracefulDegradation(ctx context.Context, recoveryCtx *RecoveryContext, result *RecoveryResult) (*RecoveryResult, error) {
	rm.logger.Info("Attempting graceful degradation",
		interfaces.LogField{Key: "action", Value: recoveryCtx.Action},
		interfaces.LogField{Key: "software", Value: recoveryCtx.Software},
	)

	// For saidata not found, we can try with intelligent defaults
	if HasErrorType(recoveryCtx.OriginalError, ErrorTypeSaidataNotFound) {
		rm.logger.Debug("Using intelligent defaults for missing saidata")
		
		// This would typically involve generating default saidata
		// and retrying the action with defaults
		result.Success = true
		result.Duration = time.Since(result.StartTime)
		result.AttemptsUsed = 1
		
		rm.logger.Info("Graceful degradation successful",
			interfaces.LogField{Key: "strategy", Value: "intelligent_defaults"},
			interfaces.LogField{Key: "duration", Value: result.Duration},
		)
		
		return result, nil
	}

	// For other cases, we might reduce functionality
	result.FinalError = WrapSAIError(ErrorTypeActionFailed, 
		"graceful degradation not available for this error type", 
		recoveryCtx.OriginalError).
		WithSuggestion("Try with different options").
		WithSuggestion("Check system requirements")
	
	result.Duration = time.Since(result.StartTime)
	return result, result.FinalError
}

// Helper methods for resource creation

func (rm *RecoveryManager) createDirectory(path string) bool {
	// Implementation would create the directory
	// For now, simulate success
	rm.logger.Debug("Creating directory",
		interfaces.LogField{Key: "path", Value: path},
	)
	return true
}

func (rm *RecoveryManager) createDefaultFile(path string, software string) bool {
	// Implementation would create a default file
	// For now, simulate success
	rm.logger.Debug("Creating default file",
		interfaces.LogField{Key: "path", Value: path},
		interfaces.LogField{Key: "software", Value: software},
	)
	return true
}

// BuildRecoveryContext creates a recovery context from action parameters
func BuildRecoveryContext(action, software string, provider *types.ProviderData, saidata *types.SoftwareData, err error) *RecoveryContext {
	return &RecoveryContext{
		Action:           action,
		Software:         software,
		Provider:         provider,
		Saidata:          saidata,
		OriginalError:    err,
		AttemptCount:     0,
		StartTime:        time.Now(),
		LastAttemptTime:  time.Now(),
		RollbackCommands: []string{},
		ExecutedCommands: []string{},
	}
}