package provider

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sai/internal/types"
)

func TestNewProviderDetector(t *testing.T) {
	detector, err := NewProviderDetector()
	require.NoError(t, err)
	assert.NotNil(t, detector)

	// Verify basic platform detection
	assert.Equal(t, runtime.GOOS, detector.GetPlatform())
	assert.Equal(t, runtime.GOARCH, detector.GetArchitecture())

	// Verify OS info was detected
	osInfo := detector.GetOSInfo()
	assert.NotNil(t, osInfo)
	assert.Equal(t, runtime.GOOS, osInfo.Platform)
	assert.Equal(t, runtime.GOARCH, osInfo.Architecture)
	assert.NotEmpty(t, osInfo.OS)
	assert.False(t, osInfo.DetectedAt.IsZero())
}

func TestProviderDetector_CheckExecutable(t *testing.T) {
	detector, err := NewProviderDetector()
	require.NoError(t, err)

	// Test with a command that should exist on all systems
	switch runtime.GOOS {
	case "windows":
		assert.True(t, detector.CheckExecutable("cmd"))
	case "darwin", "linux":
		assert.True(t, detector.CheckExecutable("sh"))
	}

	// Test with a command that shouldn't exist
	assert.False(t, detector.CheckExecutable("nonexistent-command-12345"))
}

func TestProviderDetector_CheckCommand(t *testing.T) {
	detector, err := NewProviderDetector()
	require.NoError(t, err)

	// Test with a simple command that should work
	switch runtime.GOOS {
	case "windows":
		assert.True(t, detector.CheckCommand("cmd /c echo test"))
	case "darwin", "linux":
		assert.True(t, detector.CheckCommand("echo test"))
	}

	// Test with a command that should fail
	assert.False(t, detector.CheckCommand("nonexistent-command-12345"))
}

func TestProviderDetector_IsAvailable(t *testing.T) {
	detector, err := NewProviderDetector()
	require.NoError(t, err)

	tests := []struct {
		name      string
		provider  *types.ProviderData
		wantAvail bool
	}{
		{
			name: "provider with existing executable",
			provider: &types.ProviderData{
				Provider: types.ProviderInfo{
					Name:       "test-existing",
					Type:       "package_manager",
					Platforms:  []string{runtime.GOOS},
					Executable: getExistingExecutable(),
				},
			},
			wantAvail: true,
		},
		{
			name: "provider with non-existing executable",
			provider: &types.ProviderData{
				Provider: types.ProviderInfo{
					Name:       "test-nonexisting",
					Type:       "package_manager",
					Platforms:  []string{runtime.GOOS},
					Executable: "nonexistent-command-12345",
				},
			},
			wantAvail: false,
		},
		{
			name: "provider with incompatible platform",
			provider: &types.ProviderData{
				Provider: types.ProviderInfo{
					Name:       "test-incompatible",
					Type:       "package_manager",
					Platforms:  []string{"incompatible-os"},
					Executable: getExistingExecutable(),
				},
			},
			wantAvail: false,
		},
		{
			name: "provider with no platform restrictions",
			provider: &types.ProviderData{
				Provider: types.ProviderInfo{
					Name:      "test-no-restrictions",
					Type:      "package_manager",
					Platforms: []string{}, // No platform restrictions
				},
			},
			wantAvail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			available := detector.IsAvailable(tt.provider)
			assert.Equal(t, tt.wantAvail, available)
		})
	}
}

func TestProviderDetector_GetProviderPriority(t *testing.T) {
	detector, err := NewProviderDetector()
	require.NoError(t, err)

	osInfo := detector.GetOSInfo()

	tests := []struct {
		name     string
		provider *types.ProviderData
		expected int
	}{
		{
			name: "exact OS match",
			provider: &types.ProviderData{
				Provider: types.ProviderInfo{
					Name:      "test-exact",
					Type:      "package_manager",
					Platforms: []string{osInfo.OS},
					Priority:  50,
				},
			},
			expected: 70, // 50 + 20 for exact OS match
		},
		{
			name: "platform match",
			provider: &types.ProviderData{
				Provider: types.ProviderInfo{
					Name:      "test-platform",
					Type:      "package_manager",
					Platforms: []string{osInfo.Platform},
					Priority:  50,
				},
			},
			expected: 60, // 50 + 10 for platform match
		},
		{
			name: "no platform restrictions",
			provider: &types.ProviderData{
				Provider: types.ProviderInfo{
					Name:      "test-no-restrictions",
					Type:      "package_manager",
					Platforms: []string{},
					Priority:  50,
				},
			},
			expected: 50, // Base priority
		},
		{
			name: "incompatible platform",
			provider: &types.ProviderData{
				Provider: types.ProviderInfo{
					Name:      "test-incompatible",
					Type:      "package_manager",
					Platforms: []string{"incompatible-os"},
					Priority:  50,
				},
			},
			expected: 40, // 50 - 10 for platform mismatch
		},
		{
			name: "default priority",
			provider: &types.ProviderData{
				Provider: types.ProviderInfo{
					Name:      "test-default",
					Type:      "package_manager",
					Platforms: []string{osInfo.OS},
					Priority:  0, // Default priority
				},
			},
			expected: 70, // 50 (default) + 20 for exact OS match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority := detector.GetProviderPriority(tt.provider)
			assert.Equal(t, tt.expected, priority)
		})
	}
}

func TestProviderDetector_SupportsAction(t *testing.T) {
	detector, err := NewProviderDetector()
	require.NoError(t, err)

	provider := &types.ProviderData{
		Provider: types.ProviderInfo{
			Name:         "test",
			Type:         "package_manager",
			Capabilities: []string{"install", "uninstall", "search"},
		},
		Actions: map[string]types.Action{
			"install": {
				Template: "install {{sai_package(0, 'name', 'test')}}",
			},
			"start": {
				Template: "start {{sai_service(0, 'service_name', 'test')}}",
			},
		},
	}

	// Test action that exists in actions map
	assert.True(t, detector.SupportsAction(provider, "install"))
	assert.True(t, detector.SupportsAction(provider, "start"))

	// Test action that exists in capabilities but not in actions
	assert.True(t, detector.SupportsAction(provider, "search"))

	// Test action that doesn't exist anywhere
	assert.False(t, detector.SupportsAction(provider, "nonexistent"))
}

func TestProviderDetector_Cache(t *testing.T) {
	detector, err := NewProviderDetector()
	require.NoError(t, err)

	// Set a short cache expiry for testing
	detector.SetCacheExpiry(100 * time.Millisecond)

	provider := &types.ProviderData{
		Provider: types.ProviderInfo{
			Name:       "test-cache",
			Type:       "package_manager",
			Platforms:  []string{runtime.GOOS},
			Executable: "nonexistent-command-12345",
		},
	}

	// First call should perform detection
	available1 := detector.IsAvailable(provider)
	assert.False(t, available1)

	// Second call should use cache
	available2 := detector.IsAvailable(provider)
	assert.False(t, available2)

	// Verify cache was used
	result, exists := detector.GetCachedResult("test-cache")
	assert.True(t, exists)
	assert.False(t, result.Available)
	assert.NotNil(t, result.Error)

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Should no longer be in cache
	_, exists = detector.GetCachedResult("test-cache")
	assert.False(t, exists)

	// Clear cache
	detector.ClearCache()
	_, exists = detector.GetCachedResult("test-cache")
	assert.False(t, exists)
}

func TestProviderDetector_parseOSRelease(t *testing.T) {
	detector, err := NewProviderDetector()
	require.NoError(t, err)

	// Create a temporary os-release file
	tempFile := t.TempDir() + "/os-release"
	content := `NAME="Ubuntu"
VERSION="22.04.1 LTS (Jammy Jellyfish)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 22.04.1 LTS"
VERSION_ID="22.04"
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
VERSION_CODENAME=jammy
UBUNTU_CODENAME=jammy`

	err = os.WriteFile(tempFile, []byte(content), 0644)
	require.NoError(t, err)

	result, err := detector.parseOSRelease(tempFile)
	require.NoError(t, err)

	assert.Equal(t, "Ubuntu", result["NAME"])
	assert.Equal(t, "ubuntu", result["ID"])
	assert.Equal(t, "22.04", result["VERSION_ID"])
	assert.Equal(t, "jammy", result["VERSION_CODENAME"])
}

func TestProviderDetector_extractVersionFromContent(t *testing.T) {
	detector, err := NewProviderDetector()
	require.NoError(t, err)

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "simple version",
			content:  "8.4",
			expected: "8.4",
		},
		{
			name:     "version with text",
			content:  "CentOS Linux release 8.4.2105 (Core)",
			expected: "8.4.2105",
		},
		{
			name:     "ubuntu version",
			content:  "22.04",
			expected: "22.04",
		},
		{
			name:     "no version",
			content:  "some text without version",
			expected: "some",
		},
		{
			name:     "empty content",
			content:  "",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.extractVersionFromContent(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProviderDetector_RefreshOSInfo(t *testing.T) {
	detector, err := NewProviderDetector()
	require.NoError(t, err)

	originalOSInfo := detector.GetOSInfo()
	assert.NotNil(t, originalOSInfo)

	// Refresh OS info
	err = detector.RefreshOSInfo()
	require.NoError(t, err)

	newOSInfo := detector.GetOSInfo()
	assert.NotNil(t, newOSInfo)

	// Should have updated detection time
	assert.True(t, newOSInfo.DetectedAt.After(originalOSInfo.DetectedAt) || 
		newOSInfo.DetectedAt.Equal(originalOSInfo.DetectedAt))
}

// Helper function to get an executable that should exist on the current platform
func getExistingExecutable() string {
	switch runtime.GOOS {
	case "windows":
		return "cmd"
	case "darwin", "linux":
		return "sh"
	default:
		return "sh"
	}
}