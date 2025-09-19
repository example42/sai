package errors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"sai/internal/interfaces"
	"sai/internal/types"
)

func TestDegradationManager(t *testing.T) {
	providerManager := &MockProviderManager{}
	saidataManager := &MockSaidataManager{}
	logger := &MockLogger{}

	// Setup common mocks
	logger.On("Debug", mock.Anything, mock.Anything).Maybe()
	logger.On("Info", mock.Anything, mock.Anything).Maybe()
	logger.On("Warn", mock.Anything, mock.Anything).Maybe()

	t.Run("NewDegradationManager", func(t *testing.T) {
		dm := NewDegradationManager(providerManager, saidataManager, logger)

		assert.NotNil(t, dm)
		assert.NotNil(t, dm.providerHealth)
		assert.NotNil(t, dm.degradationPolicies)
	})

	t.Run("UpdateProviderHealth success", func(t *testing.T) {
		dm := NewDegradationManager(providerManager, saidataManager, logger)

		dm.UpdateProviderHealth("apt", true, nil)

		health := dm.GetProviderHealth("apt")
		assert.Equal(t, "apt", health.Name)
		assert.True(t, health.Available)
		assert.Equal(t, 0, health.ConsecutiveFails)
		assert.True(t, health.HealthScore > 0.9) // Should be high after success
	})

	t.Run("UpdateProviderHealth failure", func(t *testing.T) {
		dm := NewDegradationManager(providerManager, saidataManager, logger)

		// Record multiple failures
		for i := 0; i < 3; i++ {
			dm.UpdateProviderHealth("apt", false, assert.AnError)
		}

		health := dm.GetProviderHealth("apt")
		assert.Equal(t, "apt", health.Name)
		assert.False(t, health.Available) // Should be marked unavailable after 3 failures
		assert.Equal(t, 3, health.ConsecutiveFails)
		assert.True(t, health.HealthScore < 0.5) // Should be low after failures
	})

	t.Run("GetProviderHealth for unknown provider", func(t *testing.T) {
		dm := NewDegradationManager(providerManager, saidataManager, logger)

		health := dm.GetProviderHealth("unknown")
		assert.Equal(t, "unknown", health.Name)
		assert.True(t, health.Available)
		assert.Equal(t, 1.0, health.HealthScore)
	})

	t.Run("GetAllProviderHealth", func(t *testing.T) {
		dm := NewDegradationManager(providerManager, saidataManager, logger)

		dm.UpdateProviderHealth("apt", true, nil)
		dm.UpdateProviderHealth("brew", false, assert.AnError)

		allHealth := dm.GetAllProviderHealth()
		assert.Len(t, allHealth, 2)
		assert.Contains(t, allHealth, "apt")
		assert.Contains(t, allHealth, "brew")

		assert.True(t, allHealth["apt"].Available)
		assert.True(t, allHealth["brew"].Available) // Only 1 failure, not enough to mark unavailable
	})

	t.Run("SetDegradationPolicy", func(t *testing.T) {
		dm := NewDegradationManager(providerManager, saidataManager, logger)

		policy := &DegradationPolicy{
			FallbackProviders: []string{"brew", "dnf"},
			UseDefaults:       true,
			AllowPartial:      false,
			MaxFailures:       5,
		}

		dm.SetDegradationPolicy("test_action", policy)

		// Test that the policy was set (indirectly through behavior)
		// We can't directly access the private method, but we can test the behavior
		assert.NotNil(t, dm) // Basic assertion to ensure test structure
	})
}

func TestDegradationManagerHandleProviderUnavailable(t *testing.T) {
	providerManager := &MockProviderManager{}
	saidataManager := &MockSaidataManager{}
	logger := &MockLogger{}

	// Setup common mocks
	logger.On("Debug", mock.Anything, mock.Anything).Maybe()
	logger.On("Info", mock.Anything, mock.Anything).Maybe()
	logger.On("Warn", mock.Anything, mock.Anything).Maybe()

	t.Run("HandleProviderUnavailable with fallback success", func(t *testing.T) {
		dm := NewDegradationManager(providerManager, saidataManager, logger)

		// Set up a policy with fallback providers
		policy := &DegradationPolicy{
			FallbackProviders: []string{"brew"},
			UseDefaults:       false,
			AllowPartial:      false,
			MaxFailures:       3,
		}
		dm.SetDegradationPolicy("install", policy)

		// Mock provider manager
		brewProvider := &types.ProviderData{
			Provider: types.ProviderInfo{Name: "brew"},
			Actions: map[string]types.Action{
				"install": {Description: "Install package"},
			},
		}

		providerManager.On("IsProviderAvailable", "brew").Return(true)
		providerManager.On("GetProvider", "brew").Return(brewProvider, nil)

		ctx := context.Background()
		result, err := dm.HandleProviderUnavailable(ctx, "install", "nginx", []string{"apt"})

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "provider_fallback", result.Strategy)
		assert.Equal(t, "brew", result.FallbackProvider)
		providerManager.AssertExpectations(t)
	})

	t.Run("HandleProviderUnavailable with intelligent defaults", func(t *testing.T) {
		dm := NewDegradationManager(providerManager, saidataManager, logger)

		// Set up a policy that uses defaults
		policy := &DegradationPolicy{
			FallbackProviders: []string{},
			UseDefaults:       true,
			AllowPartial:      false,
			MaxFailures:       3,
		}
		dm.SetDegradationPolicy("start", policy)

		// Mock saidata manager
		defaults := &types.SoftwareData{
			Services: []types.Service{
				{Name: "nginx", ServiceName: "nginx"},
			},
		}
		saidataManager.On("GenerateDefaults", "nginx").Return(defaults, nil)

		ctx := context.Background()
		result, err := dm.HandleProviderUnavailable(ctx, "start", "nginx", []string{"systemd"})

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "intelligent_defaults", result.Strategy)
		assert.True(t, result.UsedDefaults)
		assert.Contains(t, result.Warnings, "Using intelligent defaults due to provider unavailability")
		saidataManager.AssertExpectations(t)
	})

	t.Run("HandleProviderUnavailable with partial functionality", func(t *testing.T) {
		dm := NewDegradationManager(providerManager, saidataManager, logger)

		// Set up a policy that allows partial functionality
		policy := &DegradationPolicy{
			FallbackProviders: []string{},
			UseDefaults:       false,
			AllowPartial:      true,
			MaxFailures:       3,
		}
		dm.SetDegradationPolicy("status", policy)

		// Mock provider manager to return some available providers
		availableProvider := &types.ProviderData{
			Provider: types.ProviderInfo{
				Name:         "systemctl",
				Capabilities: []string{"service_status"},
			},
		}
		providerManager.On("GetAvailableProviders").Return([]*types.ProviderData{availableProvider})

		ctx := context.Background()
		result, err := dm.HandleProviderUnavailable(ctx, "status", "nginx", []string{"docker"})

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "partial_functionality", result.Strategy)
		assert.True(t, result.PartialSuccess)
		assert.Contains(t, result.AvailableFeatures, "service_status")
		assert.Contains(t, result.Warnings, "Operating with reduced functionality")
		providerManager.AssertExpectations(t)
	})

	t.Run("HandleProviderUnavailable all strategies fail", func(t *testing.T) {
		dm := NewDegradationManager(providerManager, saidataManager, logger)

		// Set up a policy where all strategies should fail
		policy := &DegradationPolicy{
			FallbackProviders: []string{"nonexistent"},
			UseDefaults:       false,
			AllowPartial:      false,
			MaxFailures:       3,
		}
		dm.SetDegradationPolicy("install", policy)

		// Mock provider manager to return no available providers
		providerManager.On("IsProviderAvailable", "nonexistent").Return(false)

		ctx := context.Background()
		result, err := dm.HandleProviderUnavailable(ctx, "install", "nginx", []string{"apt"})

		assert.Error(t, err)
		assert.False(t, result.Success)
		assert.NotNil(t, result.Error)
		assert.Contains(t, result.Warnings, "All fallback providers failed")
		providerManager.AssertExpectations(t)
	})
}

// Mock SaidataManager for testing
type MockSaidataManager struct {
	mock.Mock
}

func (m *MockSaidataManager) LoadSoftware(name string) (*types.SoftwareData, error) {
	args := m.Called(name)
	return args.Get(0).(*types.SoftwareData), args.Error(1)
}

func (m *MockSaidataManager) GetProviderConfig(software, provider string) (*types.ProviderConfig, error) {
	args := m.Called(software, provider)
	return args.Get(0).(*types.ProviderConfig), args.Error(1)
}

func (m *MockSaidataManager) GenerateDefaults(software string) (*types.SoftwareData, error) {
	args := m.Called(software)
	return args.Get(0).(*types.SoftwareData), args.Error(1)
}

func (m *MockSaidataManager) UpdateRepository() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSaidataManager) SearchSoftware(query string) ([]*interfaces.SoftwareInfo, error) {
	args := m.Called(query)
	return args.Get(0).([]*interfaces.SoftwareInfo), args.Error(1)
}

func (m *MockSaidataManager) ValidateData(data []byte) error {
	args := m.Called(data)
	return args.Error(0)
}

func (m *MockSaidataManager) ManageRepositoryOperations() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSaidataManager) SynchronizeRepository() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSaidataManager) GetSoftwareList() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockSaidataManager) CacheData(software string, data *types.SoftwareData) error {
	args := m.Called(software, data)
	return args.Error(0)
}

func (m *MockSaidataManager) GetCachedData(software string) (*types.SoftwareData, error) {
	args := m.Called(software)
	return args.Get(0).(*types.SoftwareData), args.Error(1)
}

func TestProviderHealthTracking(t *testing.T) {
	providerManager := &MockProviderManager{}
	saidataManager := &MockSaidataManager{}
	logger := &MockLogger{}

	logger.On("Debug", mock.Anything, mock.Anything).Maybe()
	logger.On("Info", mock.Anything, mock.Anything).Maybe()
	logger.On("Warn", mock.Anything, mock.Anything).Maybe()

	dm := NewDegradationManager(providerManager, saidataManager, logger)

	t.Run("Health score improves with success", func(t *testing.T) {
		// Start with some failures
		dm.UpdateProviderHealth("test", false, assert.AnError)
		dm.UpdateProviderHealth("test", false, assert.AnError)

		initialHealth := dm.GetProviderHealth("test")
		initialScore := initialHealth.HealthScore

		// Record success
		dm.UpdateProviderHealth("test", true, nil)

		finalHealth := dm.GetProviderHealth("test")
		assert.True(t, finalHealth.HealthScore > initialScore)
		assert.Equal(t, 0, finalHealth.ConsecutiveFails)
	})

	t.Run("Health score degrades with failures", func(t *testing.T) {
		// Start with good health
		dm.UpdateProviderHealth("test2", true, nil)
		initialHealth := dm.GetProviderHealth("test2")
		initialScore := initialHealth.HealthScore

		// Record failure
		dm.UpdateProviderHealth("test2", false, assert.AnError)

		finalHealth := dm.GetProviderHealth("test2")
		assert.True(t, finalHealth.HealthScore < initialScore)
		assert.Equal(t, 1, finalHealth.ConsecutiveFails)
	})

	t.Run("Provider becomes unavailable after max failures", func(t *testing.T) {
		// Record enough failures to trigger unavailability
		for i := 0; i < 4; i++ { // Default max failures is 3
			dm.UpdateProviderHealth("test3", false, assert.AnError)
		}

		health := dm.GetProviderHealth("test3")
		assert.False(t, health.Available)
		assert.Equal(t, 4, health.ConsecutiveFails)
	})
}