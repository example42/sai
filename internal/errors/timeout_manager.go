package errors

import (
	"context"
	"sync"
	"time"

	"sai/internal/interfaces"
)

// TimeoutManager handles timeout operations with configurable policies
type TimeoutManager struct {
	defaultTimeout time.Duration
	timeoutPolicies map[string]*TimeoutPolicy
	activeOperations map[string]*OperationContext
	mutex           sync.RWMutex
	logger          interfaces.Logger
}

// TimeoutPolicy defines timeout behavior for different operation types
type TimeoutPolicy struct {
	BaseTimeout     time.Duration `yaml:"base_timeout"`
	MaxTimeout      time.Duration `yaml:"max_timeout"`
	ScalingFactor   float64       `yaml:"scaling_factor"`
	RetryMultiplier float64       `yaml:"retry_multiplier"`
	MaxRetries      int           `yaml:"max_retries"`
	BackoffStrategy string        `yaml:"backoff_strategy"` // "linear", "exponential", "fixed"
}

// OperationContext tracks an ongoing operation with timeout
type OperationContext struct {
	ID          string
	Operation   string
	StartTime   time.Time
	Timeout     time.Duration
	Attempts    int
	MaxAttempts int
	Context     context.Context
	Cancel      context.CancelFunc
	Policy      *TimeoutPolicy
}

// TimeoutResult represents the result of a timeout-managed operation
type TimeoutResult struct {
	Success       bool
	TimedOut      bool
	Attempts      int
	TotalDuration time.Duration
	LastError     error
	Operation     string
}

// NewTimeoutManager creates a new timeout manager
func NewTimeoutManager(defaultTimeout time.Duration, logger interfaces.Logger) *TimeoutManager {
	return &TimeoutManager{
		defaultTimeout:   defaultTimeout,
		timeoutPolicies:  make(map[string]*TimeoutPolicy),
		activeOperations: make(map[string]*OperationContext),
		logger:           logger,
	}
}

// SetTimeoutPolicy sets a timeout policy for a specific operation type
func (tm *TimeoutManager) SetTimeoutPolicy(operationType string, policy *TimeoutPolicy) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	tm.timeoutPolicies[operationType] = policy
}

// GetTimeoutPolicy gets the timeout policy for an operation type
func (tm *TimeoutManager) GetTimeoutPolicy(operationType string) *TimeoutPolicy {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	if policy, exists := tm.timeoutPolicies[operationType]; exists {
		return policy
	}
	
	// Return default policy
	return &TimeoutPolicy{
		BaseTimeout:     tm.defaultTimeout,
		MaxTimeout:      tm.defaultTimeout * 3,
		ScalingFactor:   1.0,
		RetryMultiplier: 2.0,
		MaxRetries:      3,
		BackoffStrategy: "exponential",
	}
}

// ExecuteWithTimeout executes an operation with timeout and retry logic
func (tm *TimeoutManager) ExecuteWithTimeout(
	operationType string,
	operationID string,
	operation func(ctx context.Context) error,
) (*TimeoutResult, error) {
	policy := tm.GetTimeoutPolicy(operationType)
	startTime := time.Now()
	
	result := &TimeoutResult{
		Operation: operationType,
	}
	
	tm.logger.Debug("Starting timeout-managed operation",
		interfaces.LogField{Key: "operation_type", Value: operationType},
		interfaces.LogField{Key: "operation_id", Value: operationID},
		interfaces.LogField{Key: "base_timeout", Value: policy.BaseTimeout},
		interfaces.LogField{Key: "max_retries", Value: policy.MaxRetries},
	)
	
	var lastError error
	for attempt := 1; attempt <= policy.MaxRetries; attempt++ {
		result.Attempts = attempt
		
		// Calculate timeout for this attempt
		timeout := tm.calculateTimeout(policy, attempt)
		
		// Create operation context
		opCtx := tm.createOperationContext(operationID, operationType, timeout, attempt, policy.MaxRetries)
		
		tm.logger.Debug("Executing operation attempt",
			interfaces.LogField{Key: "attempt", Value: attempt},
			interfaces.LogField{Key: "timeout", Value: timeout},
			interfaces.LogField{Key: "operation_id", Value: operationID},
		)
		
		// Execute the operation
		err := operation(opCtx.Context)
		
		// Clean up operation context
		tm.cleanupOperation(operationID)
		
		if err == nil {
			// Success!
			result.Success = true
			result.TotalDuration = time.Since(startTime)
			
			tm.logger.Info("Operation completed successfully",
				interfaces.LogField{Key: "operation_type", Value: operationType},
				interfaces.LogField{Key: "operation_id", Value: operationID},
				interfaces.LogField{Key: "attempts", Value: attempt},
				interfaces.LogField{Key: "duration", Value: result.TotalDuration},
			)
			
			return result, nil
		}
		
		lastError = err
		
		// Check if it was a timeout
		if opCtx.Context.Err() == context.DeadlineExceeded {
			result.TimedOut = true
			tm.logger.Warn("Operation attempt timed out",
				interfaces.LogField{Key: "attempt", Value: attempt},
				interfaces.LogField{Key: "timeout", Value: timeout},
				interfaces.LogField{Key: "operation_id", Value: operationID},
			)
		} else {
			tm.logger.Debug("Operation attempt failed",
				interfaces.LogField{Key: "attempt", Value: attempt},
				interfaces.LogField{Key: "error", Value: err.Error()},
				interfaces.LogField{Key: "operation_id", Value: operationID},
			)
		}
		
		// Don't wait after the last attempt
		if attempt < policy.MaxRetries {
			backoffDelay := tm.calculateBackoff(policy, attempt)
			tm.logger.Debug("Waiting before retry",
				interfaces.LogField{Key: "backoff_delay", Value: backoffDelay},
				interfaces.LogField{Key: "next_attempt", Value: attempt + 1},
			)
			time.Sleep(backoffDelay)
		}
	}
	
	// All attempts failed
	result.Success = false
	result.LastError = lastError
	result.TotalDuration = time.Since(startTime)
	
	tm.logger.Error("Operation failed after all attempts", lastError,
		interfaces.LogField{Key: "operation_type", Value: operationType},
		interfaces.LogField{Key: "operation_id", Value: operationID},
		interfaces.LogField{Key: "attempts", Value: result.Attempts},
		interfaces.LogField{Key: "total_duration", Value: result.TotalDuration},
	)
	
	return result, lastError
}

// CancelOperation cancels an active operation
func (tm *TimeoutManager) CancelOperation(operationID string) bool {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	if opCtx, exists := tm.activeOperations[operationID]; exists {
		opCtx.Cancel()
		delete(tm.activeOperations, operationID)
		
		tm.logger.Info("Operation cancelled",
			interfaces.LogField{Key: "operation_id", Value: operationID},
			interfaces.LogField{Key: "operation", Value: opCtx.Operation},
		)
		
		return true
	}
	
	return false
}

// GetActiveOperations returns information about currently active operations
func (tm *TimeoutManager) GetActiveOperations() map[string]*OperationInfo {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	operations := make(map[string]*OperationInfo)
	for id, opCtx := range tm.activeOperations {
		operations[id] = &OperationInfo{
			ID:          id,
			Operation:   opCtx.Operation,
			StartTime:   opCtx.StartTime,
			Timeout:     opCtx.Timeout,
			Attempts:    opCtx.Attempts,
			MaxAttempts: opCtx.MaxAttempts,
			Elapsed:     time.Since(opCtx.StartTime),
		}
	}
	
	return operations
}

// OperationInfo provides information about an active operation
type OperationInfo struct {
	ID          string
	Operation   string
	StartTime   time.Time
	Timeout     time.Duration
	Attempts    int
	MaxAttempts int
	Elapsed     time.Duration
}

// Helper methods

func (tm *TimeoutManager) createOperationContext(
	operationID string,
	operationType string,
	timeout time.Duration,
	attempt int,
	maxAttempts int,
) *OperationContext {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	
	opCtx := &OperationContext{
		ID:          operationID,
		Operation:   operationType,
		StartTime:   time.Now(),
		Timeout:     timeout,
		Attempts:    attempt,
		MaxAttempts: maxAttempts,
		Context:     ctx,
		Cancel:      cancel,
		Policy:      tm.GetTimeoutPolicy(operationType),
	}
	
	tm.mutex.Lock()
	tm.activeOperations[operationID] = opCtx
	tm.mutex.Unlock()
	
	return opCtx
}

func (tm *TimeoutManager) cleanupOperation(operationID string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	
	if opCtx, exists := tm.activeOperations[operationID]; exists {
		opCtx.Cancel()
		delete(tm.activeOperations, operationID)
	}
}

func (tm *TimeoutManager) calculateTimeout(policy *TimeoutPolicy, attempt int) time.Duration {
	baseTimeout := policy.BaseTimeout
	
	// Apply scaling factor for retries
	if attempt > 1 {
		scaledTimeout := time.Duration(float64(baseTimeout) * policy.ScalingFactor * float64(attempt))
		if scaledTimeout > policy.MaxTimeout {
			return policy.MaxTimeout
		}
		return scaledTimeout
	}
	
	return baseTimeout
}

func (tm *TimeoutManager) calculateBackoff(policy *TimeoutPolicy, attempt int) time.Duration {
	baseDelay := time.Second // Base delay of 1 second
	
	switch policy.BackoffStrategy {
	case "linear":
		return time.Duration(int64(baseDelay) * int64(attempt))
	case "exponential":
		multiplier := 1.0
		for i := 1; i < attempt; i++ {
			multiplier *= policy.RetryMultiplier
		}
		return time.Duration(float64(baseDelay) * multiplier)
	case "fixed":
		return baseDelay
	default:
		// Default to exponential
		multiplier := 1.0
		for i := 1; i < attempt; i++ {
			multiplier *= policy.RetryMultiplier
		}
		return time.Duration(float64(baseDelay) * multiplier)
	}
}

// Predefined timeout policies for common operations

// DefaultTimeoutPolicies returns a set of default timeout policies
func DefaultTimeoutPolicies() map[string]*TimeoutPolicy {
	return map[string]*TimeoutPolicy{
		"install": {
			BaseTimeout:     60 * time.Second,
			MaxTimeout:      300 * time.Second,
			ScalingFactor:   1.5,
			RetryMultiplier: 2.0,
			MaxRetries:      3,
			BackoffStrategy: "exponential",
		},
		"uninstall": {
			BaseTimeout:     30 * time.Second,
			MaxTimeout:      120 * time.Second,
			ScalingFactor:   1.2,
			RetryMultiplier: 1.5,
			MaxRetries:      2,
			BackoffStrategy: "linear",
		},
		"start": {
			BaseTimeout:     15 * time.Second,
			MaxTimeout:      60 * time.Second,
			ScalingFactor:   1.3,
			RetryMultiplier: 2.0,
			MaxRetries:      3,
			BackoffStrategy: "exponential",
		},
		"stop": {
			BaseTimeout:     10 * time.Second,
			MaxTimeout:      30 * time.Second,
			ScalingFactor:   1.2,
			RetryMultiplier: 1.5,
			MaxRetries:      2,
			BackoffStrategy: "linear",
		},
		"search": {
			BaseTimeout:     20 * time.Second,
			MaxTimeout:      60 * time.Second,
			ScalingFactor:   1.0,
			RetryMultiplier: 2.0,
			MaxRetries:      2,
			BackoffStrategy: "fixed",
		},
		"repository_sync": {
			BaseTimeout:     120 * time.Second,
			MaxTimeout:      600 * time.Second,
			ScalingFactor:   1.5,
			RetryMultiplier: 2.0,
			MaxRetries:      3,
			BackoffStrategy: "exponential",
		},
	}
}

// SetupDefaultPolicies configures the timeout manager with default policies
func (tm *TimeoutManager) SetupDefaultPolicies() {
	policies := DefaultTimeoutPolicies()
	for operationType, policy := range policies {
		tm.SetTimeoutPolicy(operationType, policy)
	}
	
	tm.logger.Info("Default timeout policies configured",
		interfaces.LogField{Key: "policies_count", Value: len(policies)},
	)
}