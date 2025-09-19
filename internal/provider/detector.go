package provider

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"sai/internal/types"
)

// ProviderDetector handles provider availability detection and platform compatibility
type ProviderDetector struct {
	platform     string
	architecture string
	osInfo       *OSInfo
	cache        map[string]*DetectionResult
	cacheMutex   sync.RWMutex
	cacheExpiry  time.Duration
}

// OSInfo contains detailed operating system information
type OSInfo struct {
	Platform     string // "linux", "darwin", "windows"
	OS           string // "ubuntu", "debian", "centos", "macos", etc.
	Version      string // "22.04", "8", "13.0", etc.
	Architecture string // "amd64", "arm64", etc.
	DetectedAt   time.Time
}

// DetectionResult caches the result of provider detection
type DetectionResult struct {
	Available   bool
	Executable  string
	Version     string
	Error       error
	DetectedAt  time.Time
}

// NewProviderDetector creates a new provider detector with OS detection
func NewProviderDetector() (*ProviderDetector, error) {
	detector := &ProviderDetector{
		platform:     runtime.GOOS,
		architecture: runtime.GOARCH,
		cache:        make(map[string]*DetectionResult),
		cacheExpiry:  5 * time.Minute, // Cache results for 5 minutes
	}

	// Detect OS information
	osInfo, err := detector.detectOSInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to detect OS information: %w", err)
	}
	detector.osInfo = osInfo

	return detector, nil
}

// detectOSInfo detects detailed operating system information
func (pd *ProviderDetector) detectOSInfo() (*OSInfo, error) {
	osInfo := &OSInfo{
		Platform:     pd.platform,
		Architecture: pd.architecture,
		DetectedAt:   time.Now(),
	}

	switch pd.platform {
	case "linux":
		if err := pd.detectLinuxInfo(osInfo); err != nil {
			return nil, err
		}
	case "darwin":
		if err := pd.detectMacOSInfo(osInfo); err != nil {
			return nil, err
		}
	case "windows":
		if err := pd.detectWindowsInfo(osInfo); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported platform: %s", pd.platform)
	}

	return osInfo, nil
}

// detectLinuxInfo detects Linux distribution and version
func (pd *ProviderDetector) detectLinuxInfo(osInfo *OSInfo) error {
	// Try /etc/os-release first (most common)
	if info, err := pd.parseOSRelease("/etc/os-release"); err == nil {
		osInfo.OS = strings.ToLower(info["ID"])
		osInfo.Version = info["VERSION_ID"]
		return nil
	}

	// Try /etc/lsb-release (Ubuntu/Debian)
	if info, err := pd.parseOSRelease("/etc/lsb-release"); err == nil {
		if distrib := info["DISTRIB_ID"]; distrib != "" {
			osInfo.OS = strings.ToLower(distrib)
		}
		if version := info["DISTRIB_RELEASE"]; version != "" {
			osInfo.Version = version
		}
		return nil
	}

	// Try distribution-specific files
	distFiles := map[string]string{
		"/etc/redhat-release": "centos",
		"/etc/debian_version": "debian",
		"/etc/alpine-release": "alpine",
	}

	for file, distro := range distFiles {
		if _, err := os.Stat(file); err == nil {
			osInfo.OS = distro
			if content, err := os.ReadFile(file); err == nil {
				osInfo.Version = pd.extractVersionFromContent(string(content))
			}
			return nil
		}
	}

	// Fallback to generic linux
	osInfo.OS = "linux"
	osInfo.Version = "unknown"
	return nil
}

// detectMacOSInfo detects macOS version information
func (pd *ProviderDetector) detectMacOSInfo(osInfo *OSInfo) error {
	osInfo.OS = "macos"

	// Try sw_vers command
	if cmd := exec.Command("sw_vers", "-productVersion"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			version := strings.TrimSpace(string(output))
			// Extract major version (e.g., "13.0.1" -> "13")
			if parts := strings.Split(version, "."); len(parts) > 0 {
				osInfo.Version = parts[0]
			}
			return nil
		}
	}

	// Fallback to system version
	osInfo.Version = "unknown"
	return nil
}

// detectWindowsInfo detects Windows version information
func (pd *ProviderDetector) detectWindowsInfo(osInfo *OSInfo) error {
	osInfo.OS = "windows"

	// Try wmic command for version detection
	if cmd := exec.Command("wmic", "os", "get", "Version", "/value"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Version=") {
					version := strings.TrimPrefix(line, "Version=")
					version = strings.TrimSpace(version)
					// Extract major version (e.g., "10.0.19041" -> "10")
					if parts := strings.Split(version, "."); len(parts) > 0 {
						osInfo.Version = parts[0]
					}
					return nil
				}
			}
		}
	}

	// Fallback
	osInfo.Version = "unknown"
	return nil
}

// parseOSRelease parses /etc/os-release or /etc/lsb-release files
func (pd *ProviderDetector) parseOSRelease(filename string) (map[string]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	lines := strings.Split(string(content), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove quotes
		value = strings.Trim(value, `"'`)
		
		result[key] = value
	}

	return result, nil
}

// extractVersionFromContent extracts version number from file content
func (pd *ProviderDetector) extractVersionFromContent(content string) string {
	content = strings.TrimSpace(content)
	
	// Look for version patterns like "8.4", "20.04", etc.
	parts := strings.Fields(content)
	for _, part := range parts {
		if strings.Contains(part, ".") && len(part) <= 10 {
			// Simple heuristic for version-like strings
			return part
		}
	}
	
	// Extract first word as fallback
	if len(parts) > 0 {
		return parts[0]
	}
	
	return "unknown"
}

// IsAvailable checks if a provider is available on the current system
func (pd *ProviderDetector) IsAvailable(provider *types.ProviderData) bool {
	// Check cache first
	pd.cacheMutex.RLock()
	if result, exists := pd.cache[provider.Provider.Name]; exists {
		if time.Since(result.DetectedAt) < pd.cacheExpiry {
			pd.cacheMutex.RUnlock()
			return result.Available
		}
	}
	pd.cacheMutex.RUnlock()

	// Perform detection
	result := pd.detectProvider(provider)
	
	// Cache the result
	pd.cacheMutex.Lock()
	pd.cache[provider.Provider.Name] = result
	pd.cacheMutex.Unlock()

	return result.Available
}

// detectProvider performs the actual provider detection
func (pd *ProviderDetector) detectProvider(provider *types.ProviderData) *DetectionResult {
	result := &DetectionResult{
		DetectedAt: time.Now(),
	}

	// Check platform compatibility
	if !pd.isPlatformCompatible(provider) {
		result.Error = fmt.Errorf("provider %s not compatible with platform %s", provider.Provider.Name, pd.platform)
		return result
	}

	// Check executable availability
	if provider.Provider.Executable != "" {
		if pd.CheckExecutable(provider.Provider.Executable) {
			result.Available = true
			result.Executable = provider.Provider.Executable
			
			// Try to get version if possible
			if version := pd.getExecutableVersion(provider.Provider.Executable); version != "" {
				result.Version = version
			}
		} else {
			result.Error = fmt.Errorf("executable %s not found", provider.Provider.Executable)
		}
	} else {
		// If no executable specified, assume available if platform compatible
		result.Available = true
	}

	return result
}

// isPlatformCompatible checks if the provider is compatible with the current platform
func (pd *ProviderDetector) isPlatformCompatible(provider *types.ProviderData) bool {
	if len(provider.Provider.Platforms) == 0 {
		// No platform restrictions
		return true
	}

	// Check against platform (linux, darwin, windows)
	for _, platform := range provider.Provider.Platforms {
		if platform == pd.platform {
			return true
		}
		// Also check against OS name (ubuntu, debian, macos, etc.)
		if platform == pd.osInfo.OS {
			return true
		}
	}

	return false
}

// CheckExecutable checks if an executable is available in PATH
func (pd *ProviderDetector) CheckExecutable(executable string) bool {
	_, err := exec.LookPath(executable)
	return err == nil
}

// CheckCommand checks if a command can be executed successfully
func (pd *ProviderDetector) CheckCommand(command string) bool {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	
	err := cmd.Run()
	return err == nil
}

// getExecutableVersion attempts to get version information from an executable
func (pd *ProviderDetector) getExecutableVersion(executable string) string {
	// Common version flags to try
	versionFlags := []string{"--version", "-version", "-V", "-v"}
	
	for _, flag := range versionFlags {
		cmd := exec.Command(executable, flag)
		cmd.Stderr = nil
		
		if output, err := cmd.Output(); err == nil {
			version := strings.TrimSpace(string(output))
			if version != "" && len(version) < 100 { // Reasonable version string length
				return version
			}
		}
	}
	
	return ""
}

// GetPlatform returns the current platform
func (pd *ProviderDetector) GetPlatform() string {
	return pd.platform
}

// GetArchitecture returns the current architecture
func (pd *ProviderDetector) GetArchitecture() string {
	return pd.architecture
}

// GetOSInfo returns detailed OS information
func (pd *ProviderDetector) GetOSInfo() *OSInfo {
	return pd.osInfo
}

// GetProviderPriority calculates provider priority based on platform compatibility
func (pd *ProviderDetector) GetProviderPriority(provider *types.ProviderData) int {
	basePriority := provider.Provider.Priority
	if basePriority == 0 {
		basePriority = 50 // Default priority
	}

	// Boost priority for exact platform matches
	for _, platform := range provider.Provider.Platforms {
		if platform == pd.osInfo.OS {
			return basePriority + 20 // Exact OS match gets highest boost
		}
		if platform == pd.platform {
			return basePriority + 10 // Platform match gets medium boost
		}
	}

	// No platform restrictions or no match
	if len(provider.Provider.Platforms) == 0 {
		return basePriority // No restrictions, use base priority
	}

	return basePriority - 10 // Platform mismatch gets penalty
}

// SupportsAction checks if a provider supports a specific action
func (pd *ProviderDetector) SupportsAction(provider *types.ProviderData, action string) bool {
	// Check if action exists in provider
	if _, exists := provider.Actions[action]; exists {
		return true
	}

	// Check capabilities list
	for _, capability := range provider.Provider.Capabilities {
		if capability == action {
			return true
		}
	}

	return false
}

// ClearCache clears the detection cache
func (pd *ProviderDetector) ClearCache() {
	pd.cacheMutex.Lock()
	defer pd.cacheMutex.Unlock()
	pd.cache = make(map[string]*DetectionResult)
}

// GetCachedResult returns a cached detection result if available
func (pd *ProviderDetector) GetCachedResult(providerName string) (*DetectionResult, bool) {
	pd.cacheMutex.RLock()
	defer pd.cacheMutex.RUnlock()
	
	result, exists := pd.cache[providerName]
	if !exists {
		return nil, false
	}
	
	// Check if cache is still valid
	if time.Since(result.DetectedAt) >= pd.cacheExpiry {
		return nil, false
	}
	
	return result, true
}

// SetCacheExpiry sets the cache expiry duration
func (pd *ProviderDetector) SetCacheExpiry(duration time.Duration) {
	pd.cacheExpiry = duration
}

// RefreshOSInfo re-detects OS information (useful for testing or dynamic environments)
func (pd *ProviderDetector) RefreshOSInfo() error {
	osInfo, err := pd.detectOSInfo()
	if err != nil {
		return err
	}
	pd.osInfo = osInfo
	
	// Clear cache since OS info changed
	pd.ClearCache()
	
	return nil
}