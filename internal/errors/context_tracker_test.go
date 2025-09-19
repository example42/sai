package errors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestErrorContextTracker(t *testing.T) {
	t.Run("NewErrorContextTracker", func(t *testing.T) {
		ect := NewErrorContextTracker(100)

		assert.NotNil(t, ect)
		assert.Equal(t, 100, ect.maxSize)
		assert.NotNil(t, ect.contexts)
	})

	t.Run("NewErrorContextTracker with zero maxSize", func(t *testing.T) {
		ect := NewErrorContextTracker(0)

		assert.Equal(t, 1000, ect.maxSize) // Should use default
	})

	t.Run("TrackError", func(t *testing.T) {
		ect := NewErrorContextTracker(100)
		ctx := context.Background()
		err := NewActionFailedError("install", "nginx", 1, "failed")

		errorCtx := ect.TrackError(ctx, "install", "nginx", "apt", err)

		assert.NotNil(t, errorCtx)
		assert.NotEmpty(t, errorCtx.ID)
		assert.Equal(t, "install", errorCtx.Action)
		assert.Equal(t, "nginx", errorCtx.Software)
		assert.Equal(t, "apt", errorCtx.Provider)
		assert.Equal(t, err, errorCtx.Error)
		assert.Equal(t, ErrorTypeActionFailed, errorCtx.ErrorType)
		assert.NotZero(t, errorCtx.Timestamp)
		assert.NotNil(t, errorCtx.SystemInfo)
		assert.False(t, errorCtx.Recoverable)
		assert.NotEmpty(t, errorCtx.RecoveryHints)
	})

	t.Run("GetErrorContext", func(t *testing.T) {
		ect := NewErrorContextTracker(100)
		ctx := context.Background()
		err := NewActionFailedError("install", "nginx", 1, "failed")

		errorCtx := ect.TrackError(ctx, "install", "nginx", "apt", err)

		retrieved, exists := ect.GetErrorContext(errorCtx.ID)
		assert.True(t, exists)
		assert.Equal(t, errorCtx, retrieved)

		_, exists = ect.GetErrorContext("nonexistent")
		assert.False(t, exists)
	})

	t.Run("GetRecentErrors", func(t *testing.T) {
		ect := NewErrorContextTracker(100)
		ctx := context.Background()

		// Track multiple errors
		err1 := NewActionFailedError("install", "nginx", 1, "failed")
		err2 := NewProviderNotFoundError("apt")
		err3 := NewResourceMissingError("file", "/etc/nginx/nginx.conf")

		ect.TrackError(ctx, "install", "nginx", "apt", err1)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
		ect.TrackError(ctx, "search", "apache", "apt", err2)
		time.Sleep(1 * time.Millisecond)
		ect.TrackError(ctx, "config", "nginx", "apt", err3)

		recent := ect.GetRecentErrors(2)
		assert.Len(t, recent, 2)
		
		// Should be sorted by timestamp (most recent first)
		assert.Equal(t, "config", recent[0].Action)
		assert.Equal(t, "search", recent[1].Action)

		// Test unlimited
		all := ect.GetRecentErrors(0)
		assert.Len(t, all, 3)
	})

	t.Run("GetErrorsByType", func(t *testing.T) {
		ect := NewErrorContextTracker(100)
		ctx := context.Background()

		err1 := NewActionFailedError("install", "nginx", 1, "failed")
		err2 := NewActionFailedError("start", "apache", 1, "failed")
		err3 := NewProviderNotFoundError("apt")

		ect.TrackError(ctx, "install", "nginx", "apt", err1)
		ect.TrackError(ctx, "start", "apache", "apt", err2)
		ect.TrackError(ctx, "search", "mysql", "apt", err3)

		actionFailed := ect.GetErrorsByType(ErrorTypeActionFailed)
		assert.Len(t, actionFailed, 2)

		providerNotFound := ect.GetErrorsByType(ErrorTypeProviderNotFound)
		assert.Len(t, providerNotFound, 1)

		nonExistent := ect.GetErrorsByType(ErrorTypeSystemUnsupported)
		assert.Len(t, nonExistent, 0)
	})

	t.Run("GetErrorsByAction", func(t *testing.T) {
		ect := NewErrorContextTracker(100)
		ctx := context.Background()

		err1 := NewActionFailedError("install", "nginx", 1, "failed")
		err2 := NewActionFailedError("install", "apache", 1, "failed")
		err3 := NewProviderNotFoundError("apt")

		ect.TrackError(ctx, "install", "nginx", "apt", err1)
		ect.TrackError(ctx, "install", "apache", "apt", err2)
		ect.TrackError(ctx, "search", "mysql", "apt", err3)

		installErrors := ect.GetErrorsByAction("install")
		assert.Len(t, installErrors, 2)

		searchErrors := ect.GetErrorsByAction("search")
		assert.Len(t, searchErrors, 1)

		nonExistent := ect.GetErrorsByAction("nonexistent")
		assert.Len(t, nonExistent, 0)
	})

	t.Run("GetErrorStats", func(t *testing.T) {
		ect := NewErrorContextTracker(100)
		ctx := context.Background()

		// Track errors with different types and recoverability
		err1 := NewActionFailedError("install", "nginx", 1, "failed")
		err2 := NewProviderNotFoundError("apt") // Recoverable
		err3 := NewResourceMissingError("file", "/etc/nginx/nginx.conf") // Recoverable

		ect.TrackError(ctx, "install", "nginx", "apt", err1)
		ect.TrackError(ctx, "install", "apache", "apt", err2)
		ect.TrackError(ctx, "config", "nginx", "apt", err3)

		stats := ect.GetErrorStats()

		assert.Equal(t, 3, stats.TotalErrors)
		assert.Equal(t, 1, stats.ErrorsByType[ErrorTypeActionFailed])
		assert.Equal(t, 1, stats.ErrorsByType[ErrorTypeProviderNotFound])
		assert.Equal(t, 1, stats.ErrorsByType[ErrorTypeResourceMissing])
		assert.Equal(t, 2, stats.ErrorsByAction["install"])
		assert.Equal(t, 1, stats.ErrorsByAction["config"])
		assert.Equal(t, 2, stats.RecoverableErrors)
		assert.Equal(t, 3, stats.RecentErrors) // All are recent
	})

	t.Run("ClearErrors", func(t *testing.T) {
		ect := NewErrorContextTracker(100)
		ctx := context.Background()

		err := NewActionFailedError("install", "nginx", 1, "failed")
		ect.TrackError(ctx, "install", "nginx", "apt", err)

		assert.Len(t, ect.contexts, 1)

		ect.ClearErrors()

		assert.Len(t, ect.contexts, 0)
		stats := ect.GetErrorStats()
		assert.Equal(t, 0, stats.TotalErrors)
	})

	t.Run("MaxSize cleanup", func(t *testing.T) {
		ect := NewErrorContextTracker(3) // Small max size
		ctx := context.Background()

		// Track more errors than max size
		for i := 0; i < 5; i++ {
			err := NewActionFailedError("install", "nginx", 1, "failed")
			ect.TrackError(ctx, "install", "nginx", "apt", err)
			time.Sleep(1 * time.Millisecond) // Ensure different timestamps
		}

		// Should have cleaned up to stay under max size
		assert.True(t, len(ect.contexts) <= ect.maxSize)
	})
}

func TestStackTrace(t *testing.T) {
	t.Run("captureStackTrace", func(t *testing.T) {
		frames := captureStackTrace()

		assert.NotEmpty(t, frames)
		
		// Should have captured multiple frames
		assert.True(t, len(frames) > 0)
		
		// Each frame should have required fields
		for _, frame := range frames {
			assert.NotEmpty(t, frame.Function)
			assert.NotEmpty(t, frame.File)
			assert.True(t, frame.Line > 0)
		}
	})

	t.Run("extractPackageName", func(t *testing.T) {
		tests := []struct {
			funcName string
			expected string
		}{
			{
				funcName: "sai/internal/action.(*ActionManager).ExecuteAction",
				expected: "sai/internal/action.(*ActionManager)",
			},
			{
				funcName: "main.main",
				expected: "",
			},
			{
				funcName: "github.com/user/repo/pkg.Function",
				expected: "github.com/user/repo/pkg",
			},
		}

		for _, tt := range tests {
			t.Run(tt.funcName, func(t *testing.T) {
				result := extractPackageName(tt.funcName)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

func TestSystemInfo(t *testing.T) {
	t.Run("captureSystemInfo", func(t *testing.T) {
		info := captureSystemInfo()

		assert.NotNil(t, info)
		assert.NotEmpty(t, info.OS)
		assert.NotEmpty(t, info.Architecture)
		assert.NotEmpty(t, info.Version)
		assert.NotZero(t, info.Timestamp)
		assert.NotNil(t, info.Environment)
	})
}

func TestContextExtraction(t *testing.T) {
	t.Run("extractExecutionPath", func(t *testing.T) {
		// Test with nil context
		steps := extractExecutionPath(nil)
		assert.Empty(t, steps)

		// Test with context containing execution path
		step := ExecutionStep{
			Step:      "test_step",
			Timestamp: time.Now(),
			Success:   true,
		}
		ctx := context.WithValue(context.Background(), ContextKeyExecutionPath, []ExecutionStep{step})
		
		steps = extractExecutionPath(ctx)
		assert.Len(t, steps, 1)
		assert.Equal(t, "test_step", steps[0].Step)
	})

	t.Run("extractVariables", func(t *testing.T) {
		// Test with nil context
		vars := extractVariables(nil)
		assert.Empty(t, vars)

		// Test with context containing variables
		testVars := map[string]interface{}{
			"software": "nginx",
			"provider": "apt",
		}
		ctx := context.WithValue(context.Background(), ContextKeyVariables, testVars)
		
		vars = extractVariables(ctx)
		assert.Len(t, vars, 2)
		assert.Equal(t, "nginx", vars["software"])
		assert.Equal(t, "apt", vars["provider"])
	})

	t.Run("extractCommands", func(t *testing.T) {
		// Test with nil context
		commands := extractCommands(nil)
		assert.Empty(t, commands)

		// Test with context containing commands
		cmd := CommandExecution{
			Command:   "apt install nginx",
			StartTime: time.Now(),
			Success:   true,
		}
		ctx := context.WithValue(context.Background(), ContextKeyCommands, []CommandExecution{cmd})
		
		commands = extractCommands(ctx)
		assert.Len(t, commands, 1)
		assert.Equal(t, "apt install nginx", commands[0].Command)
	})

	t.Run("extractDuration", func(t *testing.T) {
		// Test with nil context
		duration := extractDuration(nil)
		assert.Equal(t, time.Duration(0), duration)

		// Test with context containing start time
		startTime := time.Now().Add(-5 * time.Second)
		ctx := context.WithValue(context.Background(), ContextKeyStartTime, startTime)
		
		duration = extractDuration(ctx)
		assert.True(t, duration > 4*time.Second)
		assert.True(t, duration < 6*time.Second)
	})
}

func TestContextHelpers(t *testing.T) {
	t.Run("WithExecutionStep", func(t *testing.T) {
		ctx := context.Background()
		step := ExecutionStep{
			Step:      "test_step",
			Timestamp: time.Now(),
			Success:   true,
		}

		newCtx := WithExecutionStep(ctx, step)
		
		steps := extractExecutionPath(newCtx)
		assert.Len(t, steps, 1)
		assert.Equal(t, "test_step", steps[0].Step)
	})

	t.Run("WithVariable", func(t *testing.T) {
		ctx := context.Background()

		newCtx := WithVariable(ctx, "software", "nginx")
		newCtx = WithVariable(newCtx, "provider", "apt")
		
		vars := extractVariables(newCtx)
		assert.Len(t, vars, 2)
		assert.Equal(t, "nginx", vars["software"])
		assert.Equal(t, "apt", vars["provider"])
	})

	t.Run("WithCommand", func(t *testing.T) {
		ctx := context.Background()
		cmd := CommandExecution{
			Command:   "apt install nginx",
			StartTime: time.Now(),
			Success:   true,
		}

		newCtx := WithCommand(ctx, cmd)
		
		commands := extractCommands(newCtx)
		assert.Len(t, commands, 1)
		assert.Equal(t, "apt install nginx", commands[0].Command)
	})

	t.Run("WithStartTime", func(t *testing.T) {
		ctx := context.Background()
		startTime := time.Now()

		newCtx := WithStartTime(ctx, startTime)
		
		extractedTime, ok := newCtx.Value(ContextKeyStartTime).(time.Time)
		assert.True(t, ok)
		assert.Equal(t, startTime, extractedTime)
	})
}

func TestGenerateRecoveryHints(t *testing.T) {
	t.Run("SAI error with existing suggestions", func(t *testing.T) {
		err := NewProviderNotFoundError("apt")
		hints := generateRecoveryHints(err)

		assert.NotEmpty(t, hints)
		// Should include original suggestions plus type-specific hints
		assert.Contains(t, hints, "Check available providers with 'sai stats'")
		assert.Contains(t, hints, "Install the required package manager")
	})

	t.Run("Network timeout error", func(t *testing.T) {
		err := NewSAIError(ErrorTypeNetworkTimeout, "network timeout")
		hints := generateRecoveryHints(err)

		assert.Contains(t, hints, "Check internet connectivity")
		assert.Contains(t, hints, "Try again later")
		assert.Contains(t, hints, "Use offline mode if available")
	})

	t.Run("Resource missing error", func(t *testing.T) {
		err := NewResourceMissingError("file", "/etc/nginx/nginx.conf")
		hints := generateRecoveryHints(err)

		assert.Contains(t, hints, "Create the missing resource manually")
		assert.Contains(t, hints, "Check file permissions")
	})

	t.Run("Action timeout error", func(t *testing.T) {
		err := NewActionTimeoutError("install", "nginx", "30s")
		hints := generateRecoveryHints(err)

		assert.Contains(t, hints, "Increase timeout value")
		assert.Contains(t, hints, "Check system performance")
	})

	t.Run("Non-SAI error", func(t *testing.T) {
		err := assert.AnError
		hints := generateRecoveryHints(err)

		assert.Empty(t, hints)
	})
}