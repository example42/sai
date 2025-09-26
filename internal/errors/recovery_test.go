package errors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"sai/internal/interfaces"
	"sai/internal/types"
)

// Mock implementations for testing
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) Execute(ctx context.Context, provider *types.ProviderData, action, software string, saidata *types.SoftwareData, options interfaces.ExecuteOptions) (*interfaces.ExecutionResult, error) {
	args := m.Called(ctx, provider, action, software, saidata, options)
	return args.Get(0).(*interfaces.ExecutionResult), args.Error(1)
}

func (m *MockExecutor) DryRun(ctx context.Context, provider *types.ProviderData, action, software string, saidata *types.SoftwareData, options interfaces.ExecuteOptions) (*interfaces.ExecutionResult, error) {
	args := m.Called(ctx, provider, action, software, saidata, options)
	return args.Get(0).(*interfaces.ExecutionResult), args.Error(1)
}

func (m *MockExecutor) CanExecute(provider *types.ProviderData, action, software string, saidata *types.SoftwareData) bool {
	args := m.Called(provider, action, software, saidata)
	return args.Bool(0)
}

func (m *MockExecutor) ValidateAction(provider *types.ProviderData, action, software string, saidata *types.SoftwareData) error {
	args := m.Called(provider, action, software, saidata)
	return args.Error(0)
}

func (m *MockExecutor) ExecuteCommand(ctx context.Context, command string, options interfaces.CommandOptions) (*interfaces.CommandResult, error) {
	args := m.Called(ctx, command, options)
	return args.Get(0).(*interfaces.CommandResult), args.Error(1)
}

func (m *MockExecutor) ValidateResources(saidata *types.SoftwareData, action string) (*interfaces.ResourceValidationResult, error) {
	args := m.Called(saidata, action)
	return args.Get(0).(*interfaces.ResourceValidationResult), args.Error(1)
}

func (m *MockExecutor) RenderTemplate(template string, saidata *types.SoftwareData, provider *types.ProviderData) (string, error) {
	args := m.Called(template, saidata, provider)
	return args.String(0), args.Error(1)
}

func (m *MockExecutor) ExecuteSteps(ctx context.Context, steps []types.Step, saidata *types.SoftwareData, provider *types.ProviderData, options interfaces.ExecuteOptions) (*interfaces.ExecutionResult, error) {
	args := m.Called(ctx, steps, saidata, provider, options)
	return args.Get(0).(*interfaces.ExecutionResult), args.Error(1)
}

type MockProviderManager struct {
	mock.Mock
}

func (m *MockProviderManager) GetProvidersForAction(action string) []*types.ProviderData {
	args := m.Called(action)
	return args.Get(0).([]*types.ProviderData)
}

func (m *MockProviderManager) IsProviderAvailable(name string) bool {
	args := m.Called(name)
	return args.Bool(0)
}

func (m *MockProviderManager) GetAvailableProviders() []*types.ProviderData {
	args := m.Called()
	return args.Get(0).([]*types.ProviderData)
}

func (m *MockProviderManager) GetAllProviders() []*types.ProviderData {
	args := m.Called()
	return args.Get(0).([]*types.ProviderData)
}

func (m *MockProviderManager) GetProvider(name string) (*types.ProviderData, error) {
	args := m.Called(name)
	return args.Get(0).(*types.ProviderData), args.Error(1)
}

func (m *MockProviderManager) SelectProvider(software, action, preferredProvider string) (*types.ProviderData, error) {
	args := m.Called(software, action, preferredProvider)
	return args.Get(0).(*types.ProviderData), args.Error(1)
}

func (m *MockProviderManager) LoadProviders(providerDir string) error {
	args := m.Called(providerDir)
	return args.Error(0)
}

func (m *MockProviderManager) ValidateProvider(provider *types.ProviderData) error {
	args := m.Called(provider)
	return args.Error(0)
}

func (m *MockProviderManager) ReloadProviders() error {
	args := m.Called()
	return args.Error(0)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(message string, fields ...interfaces.LogField) {
	m.Called(message, fields)
}

func (m *MockLogger) Info(message string, fields ...interfaces.LogField) {
	m.Called(message, fields)
}

func (m *MockLogger) Warn(message string, fields ...interfaces.LogField) {
	m.Called(message, fields)
}

func (m *MockLogger) Error(message string, err error, fields ...interfaces.LogField) {
	m.Called(message, err, fields)
}

func (m *MockLogger) Fatal(message string, err error, fields ...interfaces.LogField) {
	m.Called(message, err, fields)
}

func (m *MockLogger) WithFields(fields ...interfaces.LogField) interfaces.Logger {
	args := m.Called(fields)
	return args.Get(0).(interfaces.Logger)
}

func (m *MockLogger) SetLevel(level interfaces.LogLevel) {
	m.Called(level)
}

func (m *MockLogger) GetLevel() interfaces.LogLevel {
	args := m.Called()
	return args.Get(0).(interfaces.LogLevel)
}

func TestRecoveryManager(t *testing.T) {
	t.Run("NewRecoveryManager", func(t *testing.T) {
		executor := &MockExecutor{}
		providerManager := &MockProviderManager{}
		logger := &MockLogger{}
		config := DefaultRecoveryConfig()

		rm := NewRecoveryManager(executor, providerManager, logger, config)

		assert.NotNil(t, rm)
		assert.Equal(t, executor, rm.executor)
		assert.Equal(t, providerManager, rm.providerManager)
		assert.Equal(t, logger, rm.logger)
		assert.Equal(t, config, rm.config)
	})

	t.Run("NewRecoveryManager with nil config", func(t *testing.T) {
		executor := &MockExecutor{}
		providerManager := &MockProviderManager{}
		logger := &MockLogger{}

		rm := NewRecoveryManager(executor, providerManager, logger, nil)

		assert.NotNil(t, rm)
		assert.NotNil(t, rm.config)
		assert.Equal(t, 3, rm.config.MaxRetries)
	})
}

func TestRecoveryStrategies(t *testing.T) {
	executor := &MockExecutor{}
	providerManager := &MockProviderManager{}
	logger := &MockLogger{}
	config := &RecoveryConfig{
		MaxRetries:        2,
		RetryDelay:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		MaxRetryDelay:     1 * time.Second,
		EnableRollback:    true,
		RollbackTimeout:   5 * time.Second,
	}

	rm := NewRecoveryManager(executor, providerManager, logger, config)

	t.Run("determineRecoveryStrategy", func(t *testing.T) {
		tests := []struct {
			name     string
			error    error
			expected string
		}{
			{
				name:     "timeout error should retry",
				error:    NewActionTimeoutError("install", "nginx", "30s"),
				expected: "retry",
			},
			{
				name:     "provider not found should try alternative",
				error:    NewProviderNotFoundError("apt"),
				expected: "alternative_provider",
			},
			{
				name:     "resource missing should create resources",
				error:    NewResourceMissingError("file", "/etc/nginx/nginx.conf"),
				expected: "resource_creation",
			},
			{
				name:     "saidata not found should degrade gracefully",
				error:    NewSaidataNotFoundError("nginx"),
				expected: "graceful_degradation",
			},
			{
				name:     "action failed with rollback should rollback",
				error:    NewActionFailedError("install", "nginx", 1, "failed").WithContext("rollback_available", true),
				expected: "rollback",
			},
			{
				name:     "action failed without rollback should retry",
				error:    NewActionFailedError("install", "nginx", 1, "failed"),
				expected: "retry",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				strategy := rm.determineRecoveryStrategy(tt.error)
				assert.Equal(t, tt.expected, strategy)
			})
		}
	})

	t.Run("retryWithBackoff success on second attempt", func(t *testing.T) {
		ctx := context.Background()
		provider := &types.ProviderData{
			Provider: types.ProviderInfo{Name: "apt"},
		}
		saidata := &types.SoftwareData{}
		
		recoveryCtx := &RecoveryContext{
			Action:        "install",
			Software:      "nginx",
			Provider:      provider,
			Saidata:       saidata,
			OriginalError: NewActionTimeoutError("install", "nginx", "30s"),
		}

		result := &RecoveryResult{StartTime: time.Now()}

		// First attempt fails, second succeeds
		executor.On("Execute", mock.Anything, provider, "install", "nginx", saidata, mock.Anything).
			Return((*interfaces.ExecutionResult)(nil), NewCommandFailedError("apt install nginx", 1, "error")).Once()
		
		executor.On("Execute", mock.Anything, provider, "install", "nginx", saidata, mock.Anything).
			Return(&interfaces.ExecutionResult{Success: true}, nil).Once()

		logger.On("Info", mock.Anything, mock.Anything).Maybe()
		logger.On("Debug", mock.Anything, mock.Anything).Maybe()

		finalResult, err := rm.retryWithBackoff(ctx, recoveryCtx, result)

		assert.NoError(t, err)
		assert.True(t, finalResult.Success)
		assert.Equal(t, 2, finalResult.AttemptsUsed)
		executor.AssertExpectations(t)
	})

	t.Run("retryWithBackoff all attempts fail", func(t *testing.T) {
		ctx := context.Background()
		provider := &types.ProviderData{
			Provider: types.ProviderInfo{Name: "apt"},
		}
		saidata := &types.SoftwareData{}
		
		recoveryCtx := &RecoveryContext{
			Action:        "install",
			Software:      "nginx",
			Provider:      provider,
			Saidata:       saidata,
			OriginalError: NewActionTimeoutError("install", "nginx", "30s"),
		}

		result := &RecoveryResult{StartTime: time.Now()}

		// All attempts fail
		executor.On("Execute", mock.Anything, provider, "install", "nginx", saidata, mock.Anything).
			Return((*interfaces.ExecutionResult)(nil), NewCommandFailedError("apt install nginx", 1, "error")).Times(2)

		logger.On("Info", mock.Anything, mock.Anything).Maybe()
		logger.On("Debug", mock.Anything, mock.Anything).Maybe()

		finalResult, err := rm.retryWithBackoff(ctx, recoveryCtx, result)

		assert.Error(t, err)
		assert.False(t, finalResult.Success)
		assert.Equal(t, 2, finalResult.AttemptsUsed)
		assert.NotNil(t, finalResult.FinalError)
		executor.AssertExpectations(t)
	})

	t.Run("tryAlternativeProvider success", func(t *testing.T) {
		ctx := context.Background()
		originalProvider := &types.ProviderData{
			Provider: types.ProviderInfo{Name: "apt"},
		}
		alternativeProvider := &types.ProviderData{
			Provider: types.ProviderInfo{Name: "snap"},
		}
		saidata := &types.SoftwareData{}
		
		recoveryCtx := &RecoveryContext{
			Action:        "install",
			Software:      "nginx",
			Provider:      originalProvider,
			Saidata:       saidata,
			OriginalError: NewProviderUnavailableError("apt", "not installed"),
		}

		result := &RecoveryResult{StartTime: time.Now()}

		// Mock provider manager calls
		providerManager.On("GetProvidersForAction", "install").
			Return([]*types.ProviderData{originalProvider, alternativeProvider})
		providerManager.On("IsProviderAvailable", "apt").Return(false).Maybe()
		providerManager.On("IsProviderAvailable", "snap").Return(true).Maybe()

		// Mock executor calls
		executor.On("CanExecute", alternativeProvider, "install", "nginx", saidata).Return(true)
		executor.On("Execute", mock.Anything, alternativeProvider, "install", "nginx", saidata, mock.Anything).
			Return(&interfaces.ExecutionResult{Success: true}, nil)

		logger.On("Info", mock.Anything, mock.Anything).Maybe()
		logger.On("Debug", mock.Anything, mock.Anything).Maybe()

		finalResult, err := rm.tryAlternativeProvider(ctx, recoveryCtx, result)

		assert.NoError(t, err)
		assert.True(t, finalResult.Success)
		assert.Equal(t, 1, finalResult.AttemptsUsed)
		providerManager.AssertExpectations(t)
		executor.AssertExpectations(t)
	})

	t.Run("tryAlternativeProvider no alternatives available", func(t *testing.T) {
		// Create fresh mocks for this test
		freshExecutor := &MockExecutor{}
		freshProviderManager := &MockProviderManager{}
		freshLogger := &MockLogger{}
		freshRM := NewRecoveryManager(freshExecutor, freshProviderManager, freshLogger, config)
		
		ctx := context.Background()
		originalProvider := &types.ProviderData{
			Provider: types.ProviderInfo{Name: "apt"},
		}
		saidata := &types.SoftwareData{}
		
		recoveryCtx := &RecoveryContext{
			Action:        "install",
			Software:      "nginx",
			Provider:      originalProvider,
			Saidata:       saidata,
			OriginalError: NewProviderUnavailableError("apt", "not installed"),
		}

		result := &RecoveryResult{StartTime: time.Now()}

		// Mock provider manager calls - no alternative providers
		freshProviderManager.On("GetProvidersForAction", "install").
			Return([]*types.ProviderData{originalProvider})
		// IsProviderAvailable should not be called since the original provider is filtered out

		freshLogger.On("Info", mock.Anything, mock.Anything).Maybe()

		finalResult, err := freshRM.tryAlternativeProvider(ctx, recoveryCtx, result)

		assert.Error(t, err)
		assert.False(t, finalResult.Success)
		assert.NotNil(t, finalResult.FinalError)
		freshProviderManager.AssertExpectations(t)
	})

	t.Run("gracefulDegradation with saidata not found", func(t *testing.T) {
		ctx := context.Background()
		provider := &types.ProviderData{
			Provider: types.ProviderInfo{Name: "apt"},
		}
		saidata := &types.SoftwareData{}
		
		recoveryCtx := &RecoveryContext{
			Action:        "install",
			Software:      "nginx",
			Provider:      provider,
			Saidata:       saidata,
			OriginalError: NewSaidataNotFoundError("nginx"),
		}

		result := &RecoveryResult{StartTime: time.Now()}

		logger.On("Info", mock.Anything, mock.Anything).Maybe()
		logger.On("Debug", mock.Anything, mock.Anything).Maybe()

		finalResult, err := rm.gracefulDegradation(ctx, recoveryCtx, result)

		assert.NoError(t, err)
		assert.True(t, finalResult.Success)
		assert.Equal(t, 1, finalResult.AttemptsUsed)
	})
}

func TestBuildRecoveryContext(t *testing.T) {
	provider := &types.ProviderData{
		Provider: types.ProviderInfo{Name: "apt"},
	}
	saidata := &types.SoftwareData{}
	err := NewActionFailedError("install", "nginx", 1, "failed")

	ctx := BuildRecoveryContext("install", "nginx", provider, saidata, err)

	assert.Equal(t, "install", ctx.Action)
	assert.Equal(t, "nginx", ctx.Software)
	assert.Equal(t, provider, ctx.Provider)
	assert.Equal(t, saidata, ctx.Saidata)
	assert.Equal(t, err, ctx.OriginalError)
	assert.Equal(t, 0, ctx.AttemptCount)
	assert.NotZero(t, ctx.StartTime)
	assert.NotZero(t, ctx.LastAttemptTime)
	assert.Empty(t, ctx.RollbackCommands)
	assert.Empty(t, ctx.ExecutedCommands)
}

func TestDefaultRecoveryConfig(t *testing.T) {
	config := DefaultRecoveryConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 2*time.Second, config.RetryDelay)
	assert.Equal(t, 2.0, config.BackoffMultiplier)
	assert.Equal(t, 30*time.Second, config.MaxRetryDelay)
	assert.True(t, config.EnableRollback)
	assert.Equal(t, 60*time.Second, config.RollbackTimeout)
	assert.Equal(t, 5, config.CircuitBreakerThreshold)
	assert.Equal(t, 5*time.Minute, config.CircuitBreakerWindow)
}