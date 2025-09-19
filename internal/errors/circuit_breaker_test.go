package errors

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker(t *testing.T) {
	t.Run("NewCircuitBreaker", func(t *testing.T) {
		config := DefaultCircuitBreakerConfig()
		cb := NewCircuitBreaker("test", config)

		assert.Equal(t, "test", cb.name)
		assert.Equal(t, config, cb.config)
		assert.Equal(t, CircuitBreakerClosed, cb.state)
		assert.Equal(t, 0, cb.failures)
		assert.Equal(t, 0, cb.successes)
	})

	t.Run("NewCircuitBreaker with nil config", func(t *testing.T) {
		cb := NewCircuitBreaker("test", nil)

		assert.NotNil(t, cb.config)
		assert.Equal(t, 5, cb.config.FailureThreshold)
	})

	t.Run("Execute success", func(t *testing.T) {
		cb := NewCircuitBreaker("test", DefaultCircuitBreakerConfig())
		
		executed := false
		err := cb.Execute(func() error {
			executed = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())
	})

	t.Run("Execute failure", func(t *testing.T) {
		cb := NewCircuitBreaker("test", DefaultCircuitBreakerConfig())
		
		testErr := fmt.Errorf("test error")
		err := cb.Execute(func() error {
			return testErr
		})

		assert.Equal(t, testErr, err)
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())
	})

	t.Run("Circuit opens after threshold failures", func(t *testing.T) {
		config := &CircuitBreakerConfig{
			FailureThreshold: 3,
			RecoveryTimeout:  1 * time.Second,
			SuccessThreshold: 2,
			TimeWindow:       1 * time.Minute,
		}
		cb := NewCircuitBreaker("test", config)

		testErr := fmt.Errorf("test error")

		// Record failures up to threshold
		for i := 0; i < 3; i++ {
			err := cb.Execute(func() error {
				return testErr
			})
			assert.Equal(t, testErr, err)
		}

		// Circuit should now be open
		assert.Equal(t, CircuitBreakerOpen, cb.GetState())

		// Next request should be rejected immediately
		err := cb.Execute(func() error {
			t.Fatal("Function should not be executed when circuit is open")
			return nil
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circuit breaker")
		assert.Contains(t, err.Error(), "open")
	})

	t.Run("Circuit transitions to half-open after recovery timeout", func(t *testing.T) {
		config := &CircuitBreakerConfig{
			FailureThreshold: 2,
			RecoveryTimeout:  100 * time.Millisecond,
			SuccessThreshold: 1,
			TimeWindow:       1 * time.Minute,
		}
		cb := NewCircuitBreaker("test", config)

		testErr := fmt.Errorf("test error")

		// Open the circuit
		for i := 0; i < 2; i++ {
			cb.Execute(func() error { return testErr })
		}
		assert.Equal(t, CircuitBreakerOpen, cb.GetState())

		// Wait for recovery timeout
		time.Sleep(150 * time.Millisecond)

		// Next request should be allowed (half-open state)
		executed := false
		err := cb.Execute(func() error {
			executed = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)
		assert.Equal(t, CircuitBreakerClosed, cb.GetState()) // Should close after success
	})

	t.Run("Circuit closes after success threshold in half-open state", func(t *testing.T) {
		config := &CircuitBreakerConfig{
			FailureThreshold: 2,
			RecoveryTimeout:  100 * time.Millisecond,
			SuccessThreshold: 2,
			TimeWindow:       1 * time.Minute,
		}
		cb := NewCircuitBreaker("test", config)

		testErr := fmt.Errorf("test error")

		// Open the circuit
		for i := 0; i < 2; i++ {
			cb.Execute(func() error { return testErr })
		}
		assert.Equal(t, CircuitBreakerOpen, cb.GetState())

		// Wait for recovery timeout
		time.Sleep(150 * time.Millisecond)

		// First success - should be half-open
		cb.Execute(func() error { return nil })
		assert.Equal(t, CircuitBreakerHalfOpen, cb.GetState())

		// Second success - should close
		cb.Execute(func() error { return nil })
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())
	})

	t.Run("Circuit reopens on failure in half-open state", func(t *testing.T) {
		config := &CircuitBreakerConfig{
			FailureThreshold: 2,
			RecoveryTimeout:  100 * time.Millisecond,
			SuccessThreshold: 2,
			TimeWindow:       1 * time.Minute,
		}
		cb := NewCircuitBreaker("test", config)

		testErr := fmt.Errorf("test error")

		// Open the circuit
		for i := 0; i < 2; i++ {
			cb.Execute(func() error { return testErr })
		}
		assert.Equal(t, CircuitBreakerOpen, cb.GetState())

		// Wait for recovery timeout
		time.Sleep(150 * time.Millisecond)

		// Failure in half-open state should reopen circuit
		cb.Execute(func() error { return testErr })
		assert.Equal(t, CircuitBreakerOpen, cb.GetState())
	})

	t.Run("Time window failure counting", func(t *testing.T) {
		config := &CircuitBreakerConfig{
			FailureThreshold: 3,
			RecoveryTimeout:  1 * time.Second,
			SuccessThreshold: 1,
			TimeWindow:       200 * time.Millisecond,
		}
		cb := NewCircuitBreaker("test", config)

		testErr := fmt.Errorf("test error")

		// Record 2 failures
		for i := 0; i < 2; i++ {
			cb.Execute(func() error { return testErr })
		}
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())

		// Wait for time window to pass
		time.Sleep(250 * time.Millisecond)

		// Record 2 more failures (old ones should be cleared)
		for i := 0; i < 2; i++ {
			cb.Execute(func() error { return testErr })
		}
		
		// Should still be closed since old failures were cleared
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())

		// One more failure should open it
		cb.Execute(func() error { return testErr })
		assert.Equal(t, CircuitBreakerOpen, cb.GetState())
	})

	t.Run("GetStats", func(t *testing.T) {
		config := DefaultCircuitBreakerConfig()
		cb := NewCircuitBreaker("test", config)

		// Record some failures and successes
		testErr := fmt.Errorf("test error")
		cb.Execute(func() error { return testErr })
		cb.Execute(func() error { return nil })

		stats := cb.GetStats()

		assert.Equal(t, "test", stats.Name)
		assert.Equal(t, CircuitBreakerClosed, stats.State)
		assert.Equal(t, 1, stats.RecentFailures)
		assert.Equal(t, 0, stats.Successes) // Only counted in half-open state
		assert.Equal(t, config, stats.Config)
		assert.False(t, stats.LastFailTime.IsZero())
		assert.False(t, stats.LastSuccTime.IsZero())
	})

	t.Run("Reset", func(t *testing.T) {
		config := &CircuitBreakerConfig{
			FailureThreshold: 2,
			RecoveryTimeout:  1 * time.Second,
			SuccessThreshold: 1,
			TimeWindow:       1 * time.Minute,
		}
		cb := NewCircuitBreaker("test", config)

		testErr := fmt.Errorf("test error")

		// Open the circuit
		for i := 0; i < 2; i++ {
			cb.Execute(func() error { return testErr })
		}
		assert.Equal(t, CircuitBreakerOpen, cb.GetState())

		// Reset should close the circuit
		cb.Reset()
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())

		stats := cb.GetStats()
		assert.Equal(t, 0, stats.RecentFailures)
		assert.Equal(t, 0, stats.Successes)
	})
}

func TestCircuitBreakerManager(t *testing.T) {
	t.Run("NewCircuitBreakerManager", func(t *testing.T) {
		config := DefaultCircuitBreakerConfig()
		cbm := NewCircuitBreakerManager(config)

		assert.NotNil(t, cbm)
		assert.Equal(t, config, cbm.config)
		assert.NotNil(t, cbm.breakers)
	})

	t.Run("NewCircuitBreakerManager with nil config", func(t *testing.T) {
		cbm := NewCircuitBreakerManager(nil)

		assert.NotNil(t, cbm.config)
		assert.Equal(t, 5, cbm.config.FailureThreshold)
	})

	t.Run("GetCircuitBreaker creates new breaker", func(t *testing.T) {
		cbm := NewCircuitBreakerManager(DefaultCircuitBreakerConfig())

		cb := cbm.GetCircuitBreaker("test")

		assert.NotNil(t, cb)
		assert.Equal(t, "test", cb.name)
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())
	})

	t.Run("GetCircuitBreaker returns existing breaker", func(t *testing.T) {
		cbm := NewCircuitBreakerManager(DefaultCircuitBreakerConfig())

		cb1 := cbm.GetCircuitBreaker("test")
		cb2 := cbm.GetCircuitBreaker("test")

		assert.Equal(t, cb1, cb2)
	})

	t.Run("ExecuteWithCircuitBreaker", func(t *testing.T) {
		cbm := NewCircuitBreakerManager(DefaultCircuitBreakerConfig())

		executed := false
		err := cbm.ExecuteWithCircuitBreaker("test", func() error {
			executed = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("GetAllStats", func(t *testing.T) {
		cbm := NewCircuitBreakerManager(DefaultCircuitBreakerConfig())

		// Create some circuit breakers
		cbm.GetCircuitBreaker("test1")
		cbm.GetCircuitBreaker("test2")

		stats := cbm.GetAllStats()

		assert.Len(t, stats, 2)
		assert.Contains(t, stats, "test1")
		assert.Contains(t, stats, "test2")
	})

	t.Run("ResetAll", func(t *testing.T) {
		config := &CircuitBreakerConfig{
			FailureThreshold: 1,
			RecoveryTimeout:  1 * time.Second,
			SuccessThreshold: 1,
			TimeWindow:       1 * time.Minute,
		}
		cbm := NewCircuitBreakerManager(config)

		// Create and open some circuit breakers
		testErr := fmt.Errorf("test error")
		cbm.ExecuteWithCircuitBreaker("test1", func() error { return testErr })
		cbm.ExecuteWithCircuitBreaker("test2", func() error { return testErr })

		// Verify they are open
		cb1 := cbm.GetCircuitBreaker("test1")
		cb2 := cbm.GetCircuitBreaker("test2")
		assert.Equal(t, CircuitBreakerOpen, cb1.GetState())
		assert.Equal(t, CircuitBreakerOpen, cb2.GetState())

		// Reset all
		cbm.ResetAll()

		// Verify they are closed
		assert.Equal(t, CircuitBreakerClosed, cb1.GetState())
		assert.Equal(t, CircuitBreakerClosed, cb2.GetState())
	})

	t.Run("ResetCircuitBreaker", func(t *testing.T) {
		config := &CircuitBreakerConfig{
			FailureThreshold: 1,
			RecoveryTimeout:  1 * time.Second,
			SuccessThreshold: 1,
			TimeWindow:       1 * time.Minute,
		}
		cbm := NewCircuitBreakerManager(config)

		// Create and open a circuit breaker
		testErr := fmt.Errorf("test error")
		cbm.ExecuteWithCircuitBreaker("test", func() error { return testErr })

		cb := cbm.GetCircuitBreaker("test")
		assert.Equal(t, CircuitBreakerOpen, cb.GetState())

		// Reset specific circuit breaker
		err := cbm.ResetCircuitBreaker("test")
		assert.NoError(t, err)
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())
	})

	t.Run("ResetCircuitBreaker not found", func(t *testing.T) {
		cbm := NewCircuitBreakerManager(DefaultCircuitBreakerConfig())

		err := cbm.ResetCircuitBreaker("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestCircuitBreakerState(t *testing.T) {
	tests := []struct {
		state    CircuitBreakerState
		expected string
	}{
		{CircuitBreakerClosed, "closed"},
		{CircuitBreakerOpen, "open"},
		{CircuitBreakerHalfOpen, "half-open"},
		{CircuitBreakerState(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.String())
		})
	}
}

func TestDefaultCircuitBreakerConfig(t *testing.T) {
	config := DefaultCircuitBreakerConfig()

	assert.Equal(t, 5, config.FailureThreshold)
	assert.Equal(t, 30*time.Second, config.RecoveryTimeout)
	assert.Equal(t, 2, config.SuccessThreshold)
	assert.Equal(t, 1*time.Minute, config.TimeWindow)
}

func TestNewCircuitBreakerOpenError(t *testing.T) {
	err := NewCircuitBreakerOpenError("test", CircuitBreakerOpen)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test")
	assert.Contains(t, err.Error(), "open")
}

func TestCircuitBreakerConstants(t *testing.T) {
	assert.Equal(t, "repository", CircuitBreakerRepository)
	assert.Equal(t, "provider", CircuitBreakerProvider)
	assert.Equal(t, "network", CircuitBreakerNetwork)
	assert.Equal(t, "command", CircuitBreakerCommand)
}