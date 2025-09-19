package errors

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTimeoutManager(t *testing.T) {
	logger := &MockLogger{}
	logger.On("Debug", mock.Anything, mock.Anything).Maybe()
	logger.On("Info", mock.Anything, mock.Anything).Maybe()
	logger.On("Warn", mock.Anything, mock.Anything).Maybe()
	logger.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe()

	t.Run("NewTimeoutManager", func(t *testing.T) {
		tm := NewTimeoutManager(30*time.Second, logger)

		assert.NotNil(t, tm)
		assert.Equal(t, 30*time.Second, tm.defaultTimeout)
		assert.NotNil(t, tm.timeoutPolicies)
		assert.NotNil(t, tm.activeOperations)
	})

	t.Run("SetTimeoutPolicy and GetTimeoutPolicy", func(t *testing.T) {
		tm := NewTimeoutManager(30*time.Second, logger)

		policy := &TimeoutPolicy{
			BaseTimeout:     60 * time.Second,
			MaxTimeout:      180 * time.Second,
			ScalingFactor:   1.5,
			RetryMultiplier: 2.0,
			MaxRetries:      3,
			BackoffStrategy: "exponential",
		}

		tm.SetTimeoutPolicy("test_operation", policy)

		retrieved := tm.GetTimeoutPolicy("test_operation")
		assert.Equal(t, policy, retrieved)
	})

	t.Run("GetTimeoutPolicy returns default for unknown operation", func(t *testing.T) {
		tm := NewTimeoutManager(30*time.Second, logger)

		policy := tm.GetTimeoutPolicy("unknown_operation")

		assert.NotNil(t, policy)
		assert.Equal(t, 30*time.Second, policy.BaseTimeout)
		assert.Equal(t, 90*time.Second, policy.MaxTimeout) // 3x default
		assert.Equal(t, 3, policy.MaxRetries)
	})

	t.Run("ExecuteWithTimeout success on first attempt", func(t *testing.T) {
		tm := NewTimeoutManager(30*time.Second, logger)

		executed := false
		operation := func(ctx context.Context) error {
			executed = true
			return nil
		}

		result, err := tm.ExecuteWithTimeout("test", "op1", operation)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 1, result.Attempts)
		assert.False(t, result.TimedOut)
		assert.True(t, executed)
		assert.Equal(t, "test", result.Operation)
	})

	t.Run("ExecuteWithTimeout success on retry", func(t *testing.T) {
		tm := NewTimeoutManager(30*time.Second, logger)

		// Set a policy with short delays for testing
		policy := &TimeoutPolicy{
			BaseTimeout:     100 * time.Millisecond,
			MaxTimeout:      500 * time.Millisecond,
			ScalingFactor:   1.0,
			RetryMultiplier: 1.1, // Small multiplier for fast test
			MaxRetries:      3,
			BackoffStrategy: "fixed",
		}
		tm.SetTimeoutPolicy("test", policy)

		attempts := 0
		operation := func(ctx context.Context) error {
			attempts++
			if attempts < 2 {
				return fmt.Errorf("temporary failure")
			}
			return nil
		}

		result, err := tm.ExecuteWithTimeout("test", "op2", operation)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 2, result.Attempts)
		assert.False(t, result.TimedOut)
		assert.Equal(t, 2, attempts)
	})

	t.Run("ExecuteWithTimeout all attempts fail", func(t *testing.T) {
		tm := NewTimeoutManager(30*time.Second, logger)

		// Set a policy with short delays for testing
		policy := &TimeoutPolicy{
			BaseTimeout:     100 * time.Millisecond,
			MaxTimeout:      500 * time.Millisecond,
			ScalingFactor:   1.0,
			RetryMultiplier: 1.1,
			MaxRetries:      2,
			BackoffStrategy: "fixed",
		}
		tm.SetTimeoutPolicy("test", policy)

		testError := fmt.Errorf("persistent failure")
		attempts := 0
		operation := func(ctx context.Context) error {
			attempts++
			return testError
		}

		result, err := tm.ExecuteWithTimeout("test", "op3", operation)

		assert.Error(t, err)
		assert.False(t, result.Success)
		assert.Equal(t, 2, result.Attempts)
		assert.Equal(t, testError, result.LastError)
		assert.Equal(t, 2, attempts)
	})

	t.Run("ExecuteWithTimeout timeout", func(t *testing.T) {
		tm := NewTimeoutManager(30*time.Second, logger)

		// Set a very short timeout
		policy := &TimeoutPolicy{
			BaseTimeout:     5 * time.Millisecond,
			MaxTimeout:      20 * time.Millisecond,
			ScalingFactor:   1.0,
			RetryMultiplier: 1.1,
			MaxRetries:      2,
			BackoffStrategy: "fixed",
		}
		tm.SetTimeoutPolicy("test", policy)

		operation := func(ctx context.Context) error {
			// Sleep longer than timeout and check for cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return nil
			}
		}

		result, err := tm.ExecuteWithTimeout("test", "op4", operation)

		assert.Error(t, err)
		assert.False(t, result.Success)
		assert.True(t, result.TimedOut)
		assert.Equal(t, 2, result.Attempts)
	})

	t.Run("CancelOperation", func(t *testing.T) {
		tm := NewTimeoutManager(30*time.Second, logger)

		// Start a long-running operation
		operationStarted := make(chan bool)
		operationCancelled := make(chan bool)

		go func() {
			operation := func(ctx context.Context) error {
				operationStarted <- true
				select {
				case <-ctx.Done():
					operationCancelled <- true
					return ctx.Err()
				case <-time.After(1 * time.Second):
					return nil
				}
			}
			tm.ExecuteWithTimeout("test", "cancel_test", operation)
		}()

		// Wait for operation to start
		<-operationStarted

		// Cancel the operation
		cancelled := tm.CancelOperation("cancel_test")
		assert.True(t, cancelled)

		// Verify operation was cancelled
		select {
		case <-operationCancelled:
			// Success - operation was cancelled
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Operation was not cancelled")
		}
	})

	t.Run("GetActiveOperations", func(t *testing.T) {
		tm := NewTimeoutManager(30*time.Second, logger)

		// Start a long-running operation
		operationStarted := make(chan bool)

		go func() {
			operation := func(ctx context.Context) error {
				operationStarted <- true
				time.Sleep(100 * time.Millisecond)
				return nil
			}
			tm.ExecuteWithTimeout("test", "active_test", operation)
		}()

		// Wait for operation to start
		<-operationStarted

		// Check active operations
		active := tm.GetActiveOperations()
		assert.Len(t, active, 1)
		assert.Contains(t, active, "active_test")

		opInfo := active["active_test"]
		assert.Equal(t, "active_test", opInfo.ID)
		assert.Equal(t, "test", opInfo.Operation)
		assert.True(t, opInfo.Elapsed > 0)
	})
}

func TestTimeoutPolicyCalculations(t *testing.T) {
	logger := &MockLogger{}
	logger.On("Debug", mock.Anything, mock.Anything).Maybe()
	logger.On("Info", mock.Anything, mock.Anything).Maybe()

	tm := NewTimeoutManager(30*time.Second, logger)

	t.Run("calculateTimeout", func(t *testing.T) {
		policy := &TimeoutPolicy{
			BaseTimeout:   10 * time.Second,
			MaxTimeout:    60 * time.Second,
			ScalingFactor: 1.5,
		}

		// First attempt should use base timeout
		timeout1 := tm.calculateTimeout(policy, 1)
		assert.Equal(t, 10*time.Second, timeout1)

		// Second attempt should scale
		timeout2 := tm.calculateTimeout(policy, 2)
		assert.Equal(t, 30*time.Second, timeout2) // 10 * 1.5 * 2

		// Should not exceed max timeout
		timeout3 := tm.calculateTimeout(policy, 10)
		assert.Equal(t, 60*time.Second, timeout3)
	})

	t.Run("calculateBackoff linear", func(t *testing.T) {
		policy := &TimeoutPolicy{
			BackoffStrategy: "linear",
		}

		backoff1 := tm.calculateBackoff(policy, 1)
		backoff2 := tm.calculateBackoff(policy, 2)
		backoff3 := tm.calculateBackoff(policy, 3)

		assert.Equal(t, 1*time.Second, backoff1)
		assert.Equal(t, 2*time.Second, backoff2)
		assert.Equal(t, 3*time.Second, backoff3)
	})

	t.Run("calculateBackoff exponential", func(t *testing.T) {
		policy := &TimeoutPolicy{
			BackoffStrategy: "exponential",
			RetryMultiplier: 2.0,
		}

		backoff1 := tm.calculateBackoff(policy, 1)
		backoff2 := tm.calculateBackoff(policy, 2)
		backoff3 := tm.calculateBackoff(policy, 3)

		assert.Equal(t, 1*time.Second, backoff1)
		assert.Equal(t, 2*time.Second, backoff2)
		assert.Equal(t, 4*time.Second, backoff3)
	})

	t.Run("calculateBackoff fixed", func(t *testing.T) {
		policy := &TimeoutPolicy{
			BackoffStrategy: "fixed",
		}

		backoff1 := tm.calculateBackoff(policy, 1)
		backoff2 := tm.calculateBackoff(policy, 2)
		backoff3 := tm.calculateBackoff(policy, 3)

		assert.Equal(t, 1*time.Second, backoff1)
		assert.Equal(t, 1*time.Second, backoff2)
		assert.Equal(t, 1*time.Second, backoff3)
	})
}

func TestDefaultTimeoutPolicies(t *testing.T) {
	policies := DefaultTimeoutPolicies()

	assert.NotEmpty(t, policies)
	assert.Contains(t, policies, "install")
	assert.Contains(t, policies, "start")
	assert.Contains(t, policies, "search")

	installPolicy := policies["install"]
	assert.Equal(t, 60*time.Second, installPolicy.BaseTimeout)
	assert.Equal(t, 300*time.Second, installPolicy.MaxTimeout)
	assert.Equal(t, 3, installPolicy.MaxRetries)
	assert.Equal(t, "exponential", installPolicy.BackoffStrategy)
}

func TestSetupDefaultPolicies(t *testing.T) {
	logger := &MockLogger{}
	logger.On("Info", mock.Anything, mock.Anything).Maybe()

	tm := NewTimeoutManager(30*time.Second, logger)
	tm.SetupDefaultPolicies()

	// Verify policies were set
	installPolicy := tm.GetTimeoutPolicy("install")
	assert.Equal(t, 60*time.Second, installPolicy.BaseTimeout)

	searchPolicy := tm.GetTimeoutPolicy("search")
	assert.Equal(t, 20*time.Second, searchPolicy.BaseTimeout)
}