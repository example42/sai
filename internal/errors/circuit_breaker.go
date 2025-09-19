package errors

import (
	"fmt"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	// CircuitBreakerClosed - normal operation, requests are allowed
	CircuitBreakerClosed CircuitBreakerState = iota
	// CircuitBreakerOpen - circuit is open, requests are rejected immediately
	CircuitBreakerOpen
	// CircuitBreakerHalfOpen - testing if the service has recovered
	CircuitBreakerHalfOpen
)

// String returns string representation of circuit breaker state
func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitBreakerClosed:
		return "closed"
	case CircuitBreakerOpen:
		return "open"
	case CircuitBreakerHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig defines circuit breaker configuration
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of failures that will open the circuit
	FailureThreshold int
	// RecoveryTimeout is how long to wait before attempting recovery
	RecoveryTimeout time.Duration
	// SuccessThreshold is the number of successes needed to close the circuit from half-open
	SuccessThreshold int
	// TimeWindow is the time window for counting failures
	TimeWindow time.Duration
}

// DefaultCircuitBreakerConfig returns default circuit breaker configuration
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold: 5,
		RecoveryTimeout:  30 * time.Second,
		SuccessThreshold: 2,
		TimeWindow:       1 * time.Minute,
	}
}

// CircuitBreaker implements the circuit breaker pattern for external dependencies
type CircuitBreaker struct {
	name         string
	config       *CircuitBreakerConfig
	state        CircuitBreakerState
	failures     int
	successes    int
	lastFailTime time.Time
	lastSuccTime time.Time
	mutex        sync.RWMutex
	
	// Failure tracking within time window
	failureHistory []time.Time
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}
	
	return &CircuitBreaker{
		name:           name,
		config:         config,
		state:          CircuitBreakerClosed,
		failures:       0,
		successes:      0,
		failureHistory: make([]time.Time, 0),
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Check if circuit breaker allows execution
	if !cb.AllowRequest() {
		return NewCircuitBreakerOpenError(cb.name, cb.state)
	}

	// Execute the function
	err := fn()
	
	// Record the result
	if err != nil {
		cb.RecordFailure()
		return err
	}
	
	cb.RecordSuccess()
	return nil
}

// AllowRequest checks if a request should be allowed through the circuit breaker
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	now := time.Now()
	
	switch cb.state {
	case CircuitBreakerClosed:
		return true
		
	case CircuitBreakerOpen:
		// Check if recovery timeout has passed
		if now.Sub(cb.lastFailTime) >= cb.config.RecoveryTimeout {
			cb.state = CircuitBreakerHalfOpen
			cb.successes = 0
			return true
		}
		return false
		
	case CircuitBreakerHalfOpen:
		return true
		
	default:
		return false
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.lastSuccTime = time.Now()
	
	switch cb.state {
	case CircuitBreakerClosed:
		// Reset failure count on success
		cb.failures = 0
		cb.clearOldFailures()
		
	case CircuitBreakerHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			// Enough successes to close the circuit
			cb.state = CircuitBreakerClosed
			cb.failures = 0
			cb.successes = 0
			cb.failureHistory = cb.failureHistory[:0]
		}
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	now := time.Now()
	cb.lastFailTime = now
	cb.failureHistory = append(cb.failureHistory, now)
	
	// Clean old failures outside time window
	cb.clearOldFailures()
	
	// Count recent failures
	recentFailures := len(cb.failureHistory)
	
	switch cb.state {
	case CircuitBreakerClosed:
		if recentFailures >= cb.config.FailureThreshold {
			cb.state = CircuitBreakerOpen
		}
		
	case CircuitBreakerHalfOpen:
		// Any failure in half-open state opens the circuit
		cb.state = CircuitBreakerOpen
		cb.successes = 0
	}
}

// clearOldFailures removes failures outside the time window
func (cb *CircuitBreaker) clearOldFailures() {
	now := time.Now()
	cutoff := now.Add(-cb.config.TimeWindow)
	
	// Find the first failure within the time window
	validIndex := 0
	for i, failTime := range cb.failureHistory {
		if failTime.After(cutoff) {
			validIndex = i
			break
		}
		validIndex = len(cb.failureHistory) // All failures are old
	}
	
	// Keep only recent failures
	if validIndex > 0 {
		cb.failureHistory = cb.failureHistory[validIndex:]
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() *CircuitBreakerStats {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	cb.clearOldFailures() // Clean up old failures for accurate count
	
	return &CircuitBreakerStats{
		Name:           cb.name,
		State:          cb.state,
		RecentFailures: len(cb.failureHistory),
		Successes:      cb.successes,
		LastFailTime:   cb.lastFailTime,
		LastSuccTime:   cb.lastSuccTime,
		Config:         cb.config,
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	cb.state = CircuitBreakerClosed
	cb.failures = 0
	cb.successes = 0
	cb.failureHistory = cb.failureHistory[:0]
}

// CircuitBreakerStats holds circuit breaker statistics
type CircuitBreakerStats struct {
	Name           string
	State          CircuitBreakerState
	RecentFailures int
	Successes      int
	LastFailTime   time.Time
	LastSuccTime   time.Time
	Config         *CircuitBreakerConfig
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
	config   *CircuitBreakerConfig
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(config *CircuitBreakerConfig) *CircuitBreakerManager {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}
	
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		config:   config,
	}
}

// GetCircuitBreaker gets or creates a circuit breaker for a given name
func (cbm *CircuitBreakerManager) GetCircuitBreaker(name string) *CircuitBreaker {
	cbm.mutex.RLock()
	if breaker, exists := cbm.breakers[name]; exists {
		cbm.mutex.RUnlock()
		return breaker
	}
	cbm.mutex.RUnlock()
	
	// Create new circuit breaker
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()
	
	// Double-check after acquiring write lock
	if breaker, exists := cbm.breakers[name]; exists {
		return breaker
	}
	
	breaker := NewCircuitBreaker(name, cbm.config)
	cbm.breakers[name] = breaker
	return breaker
}

// ExecuteWithCircuitBreaker executes a function with circuit breaker protection
func (cbm *CircuitBreakerManager) ExecuteWithCircuitBreaker(name string, fn func() error) error {
	breaker := cbm.GetCircuitBreaker(name)
	return breaker.Execute(fn)
}

// GetAllStats returns statistics for all circuit breakers
func (cbm *CircuitBreakerManager) GetAllStats() map[string]*CircuitBreakerStats {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()
	
	stats := make(map[string]*CircuitBreakerStats)
	for name, breaker := range cbm.breakers {
		stats[name] = breaker.GetStats()
	}
	
	return stats
}

// ResetAll resets all circuit breakers
func (cbm *CircuitBreakerManager) ResetAll() {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()
	
	for _, breaker := range cbm.breakers {
		breaker.Reset()
	}
}

// ResetCircuitBreaker resets a specific circuit breaker
func (cbm *CircuitBreakerManager) ResetCircuitBreaker(name string) error {
	cbm.mutex.RLock()
	breaker, exists := cbm.breakers[name]
	cbm.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("circuit breaker %s not found", name)
	}
	
	breaker.Reset()
	return nil
}

// Predefined circuit breaker error
func NewCircuitBreakerOpenError(name string, state CircuitBreakerState) *SAIError {
	return NewSAIError(ErrorTypeNetworkUnavailable, 
		fmt.Sprintf("circuit breaker '%s' is %s", name, state.String())).
		WithContext("circuit_breaker", name).
		WithContext("state", state.String()).
		WithSuggestion("Wait for circuit breaker to recover").
		WithSuggestion("Check external service availability").
		WithSuggestion("Reset circuit breaker if service is known to be healthy")
}

// Common circuit breaker names for SAI operations
const (
	CircuitBreakerRepository = "repository"
	CircuitBreakerProvider   = "provider"
	CircuitBreakerNetwork    = "network"
	CircuitBreakerCommand    = "command"
)