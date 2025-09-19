package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSAIError(t *testing.T) {
	t.Run("NewSAIError", func(t *testing.T) {
		err := NewSAIError(ErrorTypeProviderNotFound, "test provider not found")
		
		assert.Equal(t, ErrorTypeProviderNotFound, err.Type)
		assert.Equal(t, "test provider not found", err.Message)
		assert.Nil(t, err.Cause)
		assert.True(t, err.Recoverable) // Provider not found is recoverable
		assert.Equal(t, "provider_not_found: test provider not found", err.Error())
	})

	t.Run("WrapSAIError", func(t *testing.T) {
		originalErr := fmt.Errorf("original error")
		err := WrapSAIError(ErrorTypeCommandFailed, "command execution failed", originalErr)
		
		assert.Equal(t, ErrorTypeCommandFailed, err.Type)
		assert.Equal(t, "command execution failed", err.Message)
		assert.Equal(t, originalErr, err.Cause)
		assert.False(t, err.Recoverable) // Command failed is not recoverable
		assert.Contains(t, err.Error(), "caused by: original error")
	})

	t.Run("WithContext", func(t *testing.T) {
		err := NewSAIError(ErrorTypeActionFailed, "action failed").
			WithContext("software", "nginx").
			WithContext("provider", "apt")
		
		assert.Equal(t, "nginx", err.Context["software"])
		assert.Equal(t, "apt", err.Context["provider"])
	})

	t.Run("WithSuggestion", func(t *testing.T) {
		err := NewSAIError(ErrorTypeProviderNotFound, "provider not found").
			WithSuggestion("Check available providers").
			WithSuggestion("Verify provider name")
		
		assert.Len(t, err.Suggestions, 2)
		assert.Contains(t, err.Suggestions, "Check available providers")
		assert.Contains(t, err.Suggestions, "Verify provider name")
	})

	t.Run("GetDetailedMessage", func(t *testing.T) {
		err := NewSAIError(ErrorTypeActionFailed, "action failed").
			WithContext("software", "nginx").
			WithContext("exit_code", 1).
			WithSuggestion("Check logs for details").
			WithSuggestion("Run with --verbose")
		
		detailed := err.GetDetailedMessage()
		assert.Contains(t, detailed, "action_failed: action failed")
		assert.Contains(t, detailed, "Context:")
		assert.Contains(t, detailed, "software: nginx")
		assert.Contains(t, detailed, "exit_code: 1")
		assert.Contains(t, detailed, "Suggestions:")
		assert.Contains(t, detailed, "Check logs for details")
		assert.Contains(t, detailed, "Run with --verbose")
	})

	t.Run("Is", func(t *testing.T) {
		err1 := NewSAIError(ErrorTypeProviderNotFound, "provider not found")
		err2 := NewSAIError(ErrorTypeProviderNotFound, "different message")
		err3 := NewSAIError(ErrorTypeActionFailed, "action failed")
		
		assert.True(t, err1.Is(err2))
		assert.False(t, err1.Is(err3))
	})

	t.Run("Unwrap", func(t *testing.T) {
		originalErr := fmt.Errorf("original error")
		err := WrapSAIError(ErrorTypeCommandFailed, "wrapped error", originalErr)
		
		assert.Equal(t, originalErr, err.Unwrap())
	})
}

func TestPredefinedErrors(t *testing.T) {
	t.Run("NewProviderNotFoundError", func(t *testing.T) {
		err := NewProviderNotFoundError("apt")
		
		assert.Equal(t, ErrorTypeProviderNotFound, err.Type)
		assert.Contains(t, err.Message, "apt")
		assert.Equal(t, "apt", err.Context["provider"])
		assert.NotEmpty(t, err.Suggestions)
		assert.True(t, err.Recoverable)
	})

	t.Run("NewProviderUnavailableError", func(t *testing.T) {
		err := NewProviderUnavailableError("docker", "docker command not found")
		
		assert.Equal(t, ErrorTypeProviderUnavailable, err.Type)
		assert.Contains(t, err.Message, "docker")
		assert.Contains(t, err.Message, "docker command not found")
		assert.Equal(t, "docker", err.Context["provider"])
		assert.Equal(t, "docker command not found", err.Context["reason"])
		assert.True(t, err.Recoverable)
	})

	t.Run("NewSaidataNotFoundError", func(t *testing.T) {
		err := NewSaidataNotFoundError("nginx")
		
		assert.Equal(t, ErrorTypeSaidataNotFound, err.Type)
		assert.Contains(t, err.Message, "nginx")
		assert.Equal(t, "nginx", err.Context["software"])
		assert.True(t, err.Recoverable)
	})

	t.Run("NewActionNotSupportedError", func(t *testing.T) {
		err := NewActionNotSupportedError("start", "nginx", "apt")
		
		assert.Equal(t, ErrorTypeActionNotSupported, err.Type)
		assert.Contains(t, err.Message, "start")
		assert.Contains(t, err.Message, "nginx")
		assert.Contains(t, err.Message, "apt")
		assert.Equal(t, "start", err.Context["action"])
		assert.Equal(t, "nginx", err.Context["software"])
		assert.Equal(t, "apt", err.Context["provider"])
	})

	t.Run("NewActionFailedError", func(t *testing.T) {
		err := NewActionFailedError("install", "nginx", 1, "Package not found")
		
		assert.Equal(t, ErrorTypeActionFailed, err.Type)
		assert.Contains(t, err.Message, "install")
		assert.Contains(t, err.Message, "nginx")
		assert.Equal(t, "install", err.Context["action"])
		assert.Equal(t, "nginx", err.Context["software"])
		assert.Equal(t, 1, err.Context["exit_code"])
		assert.Equal(t, "Package not found", err.Context["output"])
	})

	t.Run("NewCommandFailedError", func(t *testing.T) {
		err := NewCommandFailedError("apt-get install nginx", 100, "Package not found")
		
		assert.Equal(t, ErrorTypeCommandFailed, err.Type)
		assert.Contains(t, err.Message, "apt-get install nginx")
		assert.Equal(t, "apt-get install nginx", err.Context["command"])
		assert.Equal(t, 100, err.Context["exit_code"])
		assert.Equal(t, "Package not found", err.Context["stderr"])
	})

	t.Run("NewResourceMissingError", func(t *testing.T) {
		err := NewResourceMissingError("file", "/etc/nginx/nginx.conf")
		
		assert.Equal(t, ErrorTypeResourceMissing, err.Type)
		assert.Contains(t, err.Message, "file")
		assert.Contains(t, err.Message, "/etc/nginx/nginx.conf")
		assert.Equal(t, "file", err.Context["resource_type"])
		assert.Equal(t, "/etc/nginx/nginx.conf", err.Context["resource_path"])
		assert.True(t, err.Recoverable)
	})

	t.Run("NewSystemRequirementError", func(t *testing.T) {
		err := NewSystemRequirementError("memory", "1GB", "2GB")
		
		assert.Equal(t, ErrorTypeSystemRequirement, err.Type)
		assert.Contains(t, err.Message, "memory")
		assert.Equal(t, "memory", err.Context["requirement"])
		assert.Equal(t, "1GB", err.Context["current"])
		assert.Equal(t, "2GB", err.Context["minimum"])
	})
}

func TestErrorUtilities(t *testing.T) {
	t.Run("IsRecoverable", func(t *testing.T) {
		recoverableErr := NewProviderNotFoundError("apt")
		nonRecoverableErr := NewCommandFailedError("test", 1, "error")
		regularErr := fmt.Errorf("regular error")
		
		assert.True(t, IsRecoverable(recoverableErr))
		assert.False(t, IsRecoverable(nonRecoverableErr))
		assert.False(t, IsRecoverable(regularErr))
	})

	t.Run("GetErrorType", func(t *testing.T) {
		saiErr := NewProviderNotFoundError("apt")
		regularErr := fmt.Errorf("regular error")
		
		assert.Equal(t, ErrorTypeProviderNotFound, GetErrorType(saiErr))
		assert.Equal(t, ErrorTypeUnknown, GetErrorType(regularErr))
	})

	t.Run("HasErrorType", func(t *testing.T) {
		err := NewProviderNotFoundError("apt")
		
		assert.True(t, HasErrorType(err, ErrorTypeProviderNotFound))
		assert.False(t, HasErrorType(err, ErrorTypeActionFailed))
	})

	t.Run("GetSuggestions", func(t *testing.T) {
		err := NewProviderNotFoundError("apt")
		regularErr := fmt.Errorf("regular error")
		
		suggestions := GetSuggestions(err)
		assert.NotEmpty(t, suggestions)
		
		noSuggestions := GetSuggestions(regularErr)
		assert.Nil(t, noSuggestions)
	})

	t.Run("GetContext", func(t *testing.T) {
		err := NewProviderNotFoundError("apt")
		regularErr := fmt.Errorf("regular error")
		
		context := GetContext(err)
		assert.NotNil(t, context)
		assert.Equal(t, "apt", context["provider"])
		
		noContext := GetContext(regularErr)
		assert.Nil(t, noContext)
	})
}

func TestErrorRecoverability(t *testing.T) {
	tests := []struct {
		errorType   ErrorType
		recoverable bool
	}{
		{ErrorTypeProviderNotFound, true},
		{ErrorTypeProviderUnavailable, true},
		{ErrorTypeSaidataNotFound, true},
		{ErrorTypeActionTimeout, true},
		{ErrorTypeNetworkTimeout, true},
		{ErrorTypeResourceMissing, true},
		{ErrorTypeConfigNotFound, true},
		{ErrorTypeCommandFailed, false},
		{ErrorTypeActionFailed, false},
		{ErrorTypeSystemUnsupported, false},
		{ErrorTypeInternal, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.errorType), func(t *testing.T) {
			result := isRecoverable(tt.errorType)
			assert.Equal(t, tt.recoverable, result)
		})
	}
}

func TestErrorTypeString(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		expected  string
	}{
		{ErrorTypeProviderNotFound, "provider_not_found"},
		{ErrorTypeActionFailed, "action_failed"},
		{ErrorTypeCommandTimeout, "command_timeout"},
		{ErrorTypeResourceMissing, "resource_missing"},
		{ErrorTypeSystemUnsupported, "system_unsupported"},
	}

	for _, tt := range tests {
		t.Run(string(tt.errorType), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.errorType))
		})
	}
}

func TestComplexErrorScenarios(t *testing.T) {
	t.Run("ChainedErrors", func(t *testing.T) {
		originalErr := fmt.Errorf("network connection failed")
		wrappedErr := WrapSAIError(ErrorTypeRepositorySync, "failed to sync repository", originalErr)
		finalErr := WrapSAIError(ErrorTypeSaidataLoadFailed, "failed to load saidata", wrappedErr)
		
		assert.Equal(t, ErrorTypeSaidataLoadFailed, finalErr.Type)
		assert.Equal(t, wrappedErr, finalErr.Cause)
		assert.Equal(t, originalErr, finalErr.Unwrap().(*SAIError).Unwrap())
	})

	t.Run("ErrorWithMultipleContext", func(t *testing.T) {
		err := NewActionFailedError("install", "nginx", 1, "Package not found").
			WithContext("provider", "apt").
			WithContext("repository", "main").
			WithContext("architecture", "amd64").
			WithSuggestion("Update package cache").
			WithSuggestion("Check repository configuration").
			WithSuggestion("Try different provider")
		
		assert.Len(t, err.Context, 7) // 4 from constructor + 3 added
		assert.Len(t, err.Suggestions, 5) // 2 from constructor + 3 added
		
		detailed := err.GetDetailedMessage()
		assert.Contains(t, detailed, "provider: apt")
		assert.Contains(t, detailed, "repository: main")
		assert.Contains(t, detailed, "architecture: amd64")
		assert.Contains(t, detailed, "Update package cache")
		assert.Contains(t, detailed, "Try different provider")
	})
}