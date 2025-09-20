package saidata

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"sai/internal/debug"
	"sai/internal/interfaces"
	"sai/internal/types"
	"sai/internal/validation"
)

// Manager handles saidata loading and management
type Manager struct {
	saidataDir        string
	validator         *validation.SaidataValidator
	cache             map[string]*types.SoftwareData
	defaultsGenerator *DefaultsGenerator
	resourceValidator *SystemResourceValidator
}

// NewManager creates a new saidata manager
func NewManager(saidataDir string) *Manager {
	resourceValidator := NewSystemResourceValidator()
	
	// Try to create schema validator
	var validator *validation.SaidataValidator
	schemaPath := "schemas/saidata-0.2-schema.json"
	if v, err := validation.NewSaidataValidator(schemaPath); err == nil {
		validator = v
	} else {
		fmt.Printf("Warning: Could not load schema validator: %v\n", err)
	}
	
	return &Manager{
		saidataDir:        saidataDir,
		validator:         validator,
		cache:             make(map[string]*types.SoftwareData),
		defaultsGenerator: NewDefaultsGenerator(resourceValidator),
		resourceValidator: resourceValidator,
	}
}

// NewManagerWithBootstrap creates a new saidata manager with automatic bootstrap
func NewManagerWithBootstrap(gitURL, zipFallbackURL string) (*Manager, error) {
	// Ensure saidata is available
	saidataDir, err := EnsureSaidataAvailable(gitURL, zipFallbackURL)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure saidata availability: %w", err)
	}
	
	return NewManager(saidataDir), nil
}

// LoadSoftware loads saidata for a specific software with OS-specific overrides
func (m *Manager) LoadSoftware(name string) (*types.SoftwareData, error) {
	startTime := time.Now()
	
	// Check cache first
	if cached, exists := m.cache[name]; exists {
		debug.LogSaidataLoadingGlobal(name, "cache", "", nil, time.Since(startTime), true, nil)
		return cached, nil
	}

	// Generate prefix from software name (first 2 characters)
	prefix := generatePrefix(name)
	
	// Load base configuration following hierarchical pattern: software/{prefix}/{software}/default.yaml
	basePath := filepath.Join(m.saidataDir, "software", prefix, name, "default.yaml")
	baseData, err := m.loadSaidataFile(basePath)
	var saidataPath string = basePath
	var osOverride string = ""
	
	if err != nil {
		// Check if it's a file not found error (including nested path errors)
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file or directory") {
			// Try alternative path without "software" prefix for backward compatibility
			altBasePath := filepath.Join(m.saidataDir, prefix, name, "default.yaml")
			saidataPath = altBasePath
			baseData, err = m.loadSaidataFile(altBasePath)
			if err != nil {
				if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file or directory") {
					// Generate intelligent defaults
					saidataPath = "generated_defaults"
					baseData, err = m.GenerateDefaults(name)
					if err != nil {
						debug.LogSaidataLoadingGlobal(name, saidataPath, osOverride, nil, time.Since(startTime), false, err)
						return nil, fmt.Errorf("failed to generate defaults for software '%s': %w", name, err)
					}
					// Cache and return generated defaults (no OS overrides for generated data)
					m.cache[name] = baseData
					
					mergeResults := map[string]interface{}{
						"source": "generated_defaults",
						"packages": len(baseData.Packages),
						"services": len(baseData.Services),
						"files": len(baseData.Files),
					}
					debug.LogSaidataLoadingGlobal(name, saidataPath, osOverride, mergeResults, time.Since(startTime), true, nil)
					return baseData, nil
				} else {
					debug.LogSaidataLoadingGlobal(name, saidataPath, osOverride, nil, time.Since(startTime), false, err)
					return nil, fmt.Errorf("failed to load base saidata for software '%s' from %s: %w", name, altBasePath, err)
				}
			}
		} else {
			debug.LogSaidataLoadingGlobal(name, saidataPath, osOverride, nil, time.Since(startTime), false, err)
			return nil, fmt.Errorf("failed to load base saidata for software '%s' from %s: %w", name, basePath, err)
		}
	}

	// Detect current OS and version for OS-specific overrides
	osInfo, err := detectOSInfo()
	if err != nil {
		// If OS detection fails, log warning but continue with base data
		fmt.Printf("Warning: OS detection failed, using base saidata only: %v\n", err)
		m.cache[name] = baseData
		return baseData, nil
	}

	// Try to load OS-specific override following pattern: software/{prefix}/{software}/{os}/{os_version}.yaml
	overridePath := filepath.Join(m.saidataDir, "software", prefix, name, osInfo.OS, osInfo.Version+".yaml")
	if _, err := os.Stat(overridePath); err == nil {
		osOverride = fmt.Sprintf("%s/%s", osInfo.OS, osInfo.Version)
		overrideData, err := m.loadSaidataFile(overridePath)
		if err != nil {
			// If override fails to load, log warning but continue with base data
			fmt.Printf("Warning: failed to load OS override from %s: %v\n", overridePath, err)
		} else {
			// Deep merge override with base data
			baseData = m.mergeSaidata(baseData, overrideData)
		}
	} else {
		// Try alternative path without "software" prefix for backward compatibility
		altOverridePath := filepath.Join(m.saidataDir, prefix, name, osInfo.OS, osInfo.Version+".yaml")
		if _, err := os.Stat(altOverridePath); err == nil {
			osOverride = fmt.Sprintf("%s/%s", osInfo.OS, osInfo.Version)
			overrideData, err := m.loadSaidataFile(altOverridePath)
			if err != nil {
				fmt.Printf("Warning: failed to load OS override from %s: %v\n", altOverridePath, err)
			} else {
				// Applying OS override from alternative path
				baseData = m.mergeSaidata(baseData, overrideData)
			}
		} else {
			// Try without version (just OS) - first with "software" prefix
			osOnlyPath := filepath.Join(m.saidataDir, "software", prefix, name, osInfo.OS, "default.yaml")
			if _, err := os.Stat(osOnlyPath); err == nil {
				osOverride = osInfo.OS
				overrideData, err := m.loadSaidataFile(osOnlyPath)
				if err != nil {
					fmt.Printf("Warning: failed to load OS-only override from %s: %v\n", osOnlyPath, err)
				} else {
					// Applying OS-only override
					baseData = m.mergeSaidata(baseData, overrideData)
				}
			} else {
				// Try alternative path without "software" prefix
				altOSOnlyPath := filepath.Join(m.saidataDir, prefix, name, osInfo.OS, "default.yaml")
				if _, err := os.Stat(altOSOnlyPath); err == nil {
					osOverride = osInfo.OS
					overrideData, err := m.loadSaidataFile(altOSOnlyPath)
					if err != nil {
						fmt.Printf("Warning: failed to load OS-only override from %s: %v\n", altOSOnlyPath, err)
					} else {
						// Applying OS-only override from alternative path
						baseData = m.mergeSaidata(baseData, overrideData)
					}
				}
			}
		}
	}

	// Cache the result
	m.cache[name] = baseData
	
	// Log successful saidata loading with merge results
	mergeResults := map[string]interface{}{
		"source": saidataPath,
		"os_override": osOverride,
		"packages": len(baseData.Packages),
		"services": len(baseData.Services),
		"files": len(baseData.Files),
		"directories": len(baseData.Directories),
		"commands": len(baseData.Commands),
		"ports": len(baseData.Ports),
		"containers": len(baseData.Containers),
		"providers": len(baseData.Providers),
	}
	debug.LogSaidataLoadingGlobal(name, saidataPath, osOverride, mergeResults, time.Since(startTime), true, nil)
	
	return baseData, nil
}

// loadSaidataFile loads and validates a saidata YAML file
func (m *Manager) loadSaidataFile(filePath string) (*types.SoftwareData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read saidata file %s: %w", filePath, err)
	}

	// Parse YAML
	saidata, err := types.LoadSoftwareDataFromYAML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse saidata YAML from %s: %w", filePath, err)
	}

	// Validate against schema if validator is available
	if m.validator != nil {
		if err := m.validator.ValidateSaidata(saidata); err != nil {
			return nil, fmt.Errorf("saidata schema validation failed for %s:\n%w\n\nPlease check that the file follows the saidata-0.2-schema.json format", filePath, err)
		}
	} else {
		fmt.Printf("Warning: Schema validation skipped for %s (validator not available)\n", filePath)
	}

	return saidata, nil
}

// mergeSaidata performs deep merge of override data into base data
func (m *Manager) mergeSaidata(base, override *types.SoftwareData) *types.SoftwareData {
	// Create a copy of base to avoid modifying original
	result := *base

	// Merge metadata (override takes precedence)
	if override.Metadata.Name != "" {
		result.Metadata.Name = override.Metadata.Name
	}
	if override.Metadata.DisplayName != "" {
		result.Metadata.DisplayName = override.Metadata.DisplayName
	}
	if override.Metadata.Description != "" {
		result.Metadata.Description = override.Metadata.Description
	}
	if override.Metadata.Version != "" {
		result.Metadata.Version = override.Metadata.Version
	}

	// Merge arrays by replacing or appending
	if len(override.Packages) > 0 {
		result.Packages = mergePackages(result.Packages, override.Packages)
	}
	if len(override.Services) > 0 {
		result.Services = mergeServices(result.Services, override.Services)
	}
	if len(override.Files) > 0 {
		result.Files = mergeFiles(result.Files, override.Files)
	}
	if len(override.Directories) > 0 {
		result.Directories = mergeDirectories(result.Directories, override.Directories)
	}
	if len(override.Commands) > 0 {
		result.Commands = mergeCommands(result.Commands, override.Commands)
	}
	if len(override.Ports) > 0 {
		result.Ports = mergePorts(result.Ports, override.Ports)
	}
	if len(override.Containers) > 0 {
		result.Containers = mergeContainers(result.Containers, override.Containers)
	}

	// Merge provider configurations
	if override.Providers != nil {
		if result.Providers == nil {
			result.Providers = make(map[string]types.ProviderConfig)
		}
		for providerName, providerConfig := range override.Providers {
			result.Providers[providerName] = mergeProviderConfig(result.Providers[providerName], providerConfig)
		}
	}

	// Merge compatibility
	if override.Compatibility != nil {
		if result.Compatibility == nil {
			result.Compatibility = override.Compatibility
		} else {
			result.Compatibility = mergeCompatibility(result.Compatibility, override.Compatibility)
		}
	}

	return &result
}

// GetProviderConfig returns provider-specific configuration with fallback to defaults
func (m *Manager) GetProviderConfig(software string, provider string) (*types.ProviderConfig, error) {
	saidata, err := m.LoadSoftware(software)
	if err != nil {
		return nil, err
	}

	if config, exists := saidata.Providers[provider]; exists {
		return &config, nil
	}

	// Return empty config if no provider-specific configuration exists
	return &types.ProviderConfig{}, nil
}

// SearchSoftware searches for software in the saidata directory
func (m *Manager) SearchSoftware(query string) ([]*interfaces.SoftwareInfo, error) {
	var results []*interfaces.SoftwareInfo

	// Walk through the saidata directory structure
	err := filepath.Walk(m.saidataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors and continue
		}

		// Look for default.yaml files
		if info.Name() == "default.yaml" {
			// Extract software name from path
			relPath, err := filepath.Rel(m.saidataDir, path)
			if err != nil {
				return nil
			}

			parts := strings.Split(relPath, string(filepath.Separator))
			var softwareName string
			
			// Handle both hierarchical patterns:
			// 1. software/{prefix}/{software}/default.yaml (new format)
			// 2. {prefix}/{software}/default.yaml (backward compatibility)
			if len(parts) >= 4 && parts[0] == "software" {
				softwareName = parts[2] // software/prefix/software/default.yaml
			} else if len(parts) >= 3 {
				softwareName = parts[1] // prefix/software/default.yaml
			} else {
				return nil // Skip invalid paths
			}
			
			// Check if software name matches query
			if strings.Contains(strings.ToLower(softwareName), strings.ToLower(query)) {
				// Load basic metadata
				saidata, err := m.loadSaidataFile(path)
				if err != nil {
					fmt.Printf("Warning: Failed to load saidata for %s: %v\n", softwareName, err)
					return nil // Skip invalid files
				}

				homepage := ""
				license := ""
				if saidata.Metadata.URLs != nil {
					homepage = saidata.Metadata.URLs.Website
				}
				if saidata.Metadata.License != "" {
					license = saidata.Metadata.License
				}

				results = append(results, &interfaces.SoftwareInfo{
					Software:     softwareName,
					Provider:     "saidata",
					PackageName:  softwareName,
					Version:      saidata.Metadata.Version,
					Description:  saidata.Metadata.Description,
					Homepage:     homepage,
					License:      license,
					Dependencies: []string{},
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search saidata directory: %w", err)
	}

	return results, nil
}

// ValidateData validates saidata against the schema
func (m *Manager) ValidateData(data []byte) error {
	saidata, err := types.LoadSoftwareDataFromYAML(data)
	if err != nil {
		return err
	}

	if m.validator != nil {
		return m.validator.ValidateSaidata(saidata)
	}
	
	return nil // Skip validation if no validator
}

// SoftwareInfo represents basic software information for search results
type SoftwareInfo struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name,omitempty"`
	Description string   `json:"description,omitempty"`
	Version     string   `json:"version,omitempty"`
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// OSInfo represents detected OS information
type OSInfo struct {
	OS      string
	Version string
}

// generatePrefix generates a 2-character prefix from software name
func generatePrefix(name string) string {
	if len(name) < 2 {
		return name + "x" // Pad with 'x' if name is too short
	}
	return strings.ToLower(name[:2])
}

// detectOSInfo detects the current OS and version
func detectOSInfo() (*OSInfo, error) {
	// Check for environment variable override for testing
	if testOS := os.Getenv("SAI_TEST_OS"); testOS != "" {
		testVersion := os.Getenv("SAI_TEST_OS_VERSION")
		if testVersion == "" {
			testVersion = "unknown"
		}
		// Using test OS override for testing
		return &OSInfo{OS: testOS, Version: testVersion}, nil
	}
	
	switch runtime.GOOS {
	case "linux":
		osInfo, err := detectLinuxInfo()
		if err != nil {
			return nil, fmt.Errorf("failed to detect Linux distribution: %w", err)
		}
		return osInfo, nil
	case "darwin":
		osInfo, err := detectMacOSInfo()
		if err != nil {
			return nil, fmt.Errorf("failed to detect macOS version: %w", err)
		}
		return osInfo, nil
	case "windows":
		osInfo, err := detectWindowsInfo()
		if err != nil {
			return nil, fmt.Errorf("failed to detect Windows version: %w", err)
		}
		return osInfo, nil
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// detectLinuxInfo detects Linux distribution and version
func detectLinuxInfo() (*OSInfo, error) {
	// Try /etc/os-release first (most modern distributions)
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		if osInfo, err := parseOSRelease(string(data)); err == nil {
			return osInfo, nil
		}
		// Failed to parse /etc/os-release, trying other methods
	}

	// Try /etc/lsb-release (Ubuntu/Debian)
	if data, err := os.ReadFile("/etc/lsb-release"); err == nil {
		if osInfo, err := parseLSBRelease(string(data)); err == nil {
			return osInfo, nil
		}
		// Failed to parse /etc/lsb-release, trying other methods
	}

	// Try other distribution-specific files
	distFiles := map[string]string{
		"/etc/redhat-release": "rhel",
		"/etc/centos-release": "centos",
		"/etc/fedora-release": "fedora",
		"/etc/debian_version": "debian",
		"/etc/rocky-release":  "rocky",
		"/etc/almalinux-release": "almalinux",
		"/etc/alpine-release": "alpine",
	}
	
	for file, distro := range distFiles {
		if data, err := os.ReadFile(file); err == nil {
			version := strings.TrimSpace(string(data))
			// Extract version number from release strings
			if strings.Contains(version, "release") {
				parts := strings.Fields(version)
				for _, part := range parts {
					if strings.Contains(part, ".") && len(part) <= 10 {
						version = part
						break
					}
				}
			}
			return &OSInfo{OS: distro, Version: version}, nil
		}
	}

	// Fallback to generic linux
	// No specific Linux distribution detected, using generic linux
	return &OSInfo{OS: "linux", Version: "unknown"}, nil
}

// detectMacOSInfo detects macOS version
func detectMacOSInfo() (*OSInfo, error) {
	// Try to read macOS version from system
	if data, err := os.ReadFile("/System/Library/CoreServices/SystemVersion.plist"); err == nil {
		// Parse plist for version - simplified approach
		content := string(data)
		if strings.Contains(content, "ProductVersion") {
			// Extract version using simple string parsing
			lines := strings.Split(content, "\n")
			for i, line := range lines {
				if strings.Contains(line, "ProductVersion") && i+1 < len(lines) {
					nextLine := strings.TrimSpace(lines[i+1])
					if strings.HasPrefix(nextLine, "<string>") && strings.HasSuffix(nextLine, "</string>") {
						version := strings.TrimPrefix(nextLine, "<string>")
						version = strings.TrimSuffix(version, "</string>")
						// Extract major version (e.g., "13.0.1" -> "13")
						parts := strings.Split(version, ".")
						if len(parts) > 0 {
							return &OSInfo{OS: "macos", Version: parts[0]}, nil
						}
					}
				}
			}
		}
	}
	
	// Fallback to default
	return &OSInfo{OS: "macos", Version: "13"}, nil
}

// detectWindowsInfo detects Windows version
func detectWindowsInfo() (*OSInfo, error) {
	// Try to detect Windows version using registry or system files
	// For now, return a reasonable default
	// In a full implementation, this would use Windows APIs or registry queries
	return &OSInfo{OS: "windows", Version: "11"}, nil
}

// parseOSRelease parses /etc/os-release format
func parseOSRelease(content string) (*OSInfo, error) {
	var osID, versionID, prettyName string
	
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ID=") {
			osID = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			versionID = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
		} else if strings.HasPrefix(line, "PRETTY_NAME=") {
			prettyName = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
		}
	}

	if osID == "" {
		return nil, fmt.Errorf("could not determine OS ID from os-release")
	}

	// If no VERSION_ID, try to extract from PRETTY_NAME
	if versionID == "" && prettyName != "" {
		// Try to extract version from pretty name (e.g., "Ubuntu 22.04.3 LTS")
		parts := strings.Fields(prettyName)
		for _, part := range parts {
			if strings.Contains(part, ".") && len(part) <= 10 {
				// Extract major version (e.g., "22.04.3" -> "22.04")
				versionParts := strings.Split(part, ".")
				if len(versionParts) >= 2 {
					versionID = versionParts[0] + "." + versionParts[1]
				} else {
					versionID = part
				}
				break
			}
		}
	}

	if versionID == "" {
		versionID = "unknown"
	}

	return &OSInfo{OS: osID, Version: versionID}, nil
}

// parseLSBRelease parses /etc/lsb-release format
func parseLSBRelease(content string) (*OSInfo, error) {
	var distID, release string
	
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "DISTRIB_ID=") {
			distID = strings.ToLower(strings.TrimPrefix(line, "DISTRIB_ID="))
		} else if strings.HasPrefix(line, "DISTRIB_RELEASE=") {
			release = strings.TrimPrefix(line, "DISTRIB_RELEASE=")
		}
	}

	if distID == "" {
		return nil, fmt.Errorf("could not determine distribution ID from lsb-release")
	}

	return &OSInfo{OS: distID, Version: release}, nil
}

// GenerateDefaults generates intelligent defaults for missing saidata scenarios
func (m *Manager) GenerateDefaults(software string) (*types.SoftwareData, error) {
	return m.defaultsGenerator.GenerateDefaults(software)
}

// UpdateRepository updates the saidata repository
func (m *Manager) UpdateRepository() error {
	// Create repository manager with default URLs
	repoManager := NewRepositoryManager(
		"https://github.com/example42/saidata.git",
		"https://github.com/example42/saidata/archive/main.zip",
	)
	
	return repoManager.UpdateRepository()
}

// ManageRepositoryOperations manages saidata repository operations
func (m *Manager) ManageRepositoryOperations() error {
	// Create repository manager with default URLs
	repoManager := NewRepositoryManager(
		"https://github.com/example42/saidata.git",
		"https://github.com/example42/saidata/archive/main.zip",
	)
	
	// For now, this just updates the repository
	return repoManager.UpdateRepository()
}

// SynchronizeRepository synchronizes the saidata repository
func (m *Manager) SynchronizeRepository() error {
	// Create repository manager with default URLs
	repoManager := NewRepositoryManager(
		"https://github.com/example42/saidata.git",
		"https://github.com/example42/saidata/archive/main.zip",
	)
	
	return repoManager.SynchronizeRepository()
}

// ValidateResourcesExist validates that resources referenced in saidata actually exist
func (m *Manager) ValidateResourcesExist(saidata *types.SoftwareData, action string) (*ValidationResult, error) {
	return m.resourceValidator.ValidateResources(saidata, action)
}

// GetResourceStatus returns the current status of all resources in saidata
func (m *Manager) GetResourceStatus(saidata *types.SoftwareData) (*ResourceStatus, error) {
	return m.resourceValidator.GetResourceStatus(saidata)
}

// CacheData caches saidata for performance
func (m *Manager) CacheData(software string, data *types.SoftwareData) error {
	m.cache[software] = data
	return nil
}

// GetCachedData retrieves cached saidata
func (m *Manager) GetCachedData(software string) (*types.SoftwareData, error) {
	if cached, exists := m.cache[software]; exists {
		return cached, nil
	}
	return nil, fmt.Errorf("no cached data for software: %s", software)
}

// GetSoftwareList returns a list of available software
func (m *Manager) GetSoftwareList() ([]string, error) {
	var softwareList []string
	
	err := filepath.Walk(m.saidataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Look for default.yaml files
		if info.Name() == "default.yaml" {
			// Extract software name from path
			relPath, err := filepath.Rel(m.saidataDir, path)
			if err != nil {
				return nil
			}

			parts := strings.Split(relPath, string(filepath.Separator))
			var softwareName string
			
			// Handle both hierarchical patterns:
			// 1. software/{prefix}/{software}/default.yaml (new format)
			// 2. {prefix}/{software}/default.yaml (backward compatibility)
			if len(parts) >= 4 && parts[0] == "software" {
				softwareName = parts[2] // software/prefix/software/default.yaml
			} else if len(parts) >= 3 {
				softwareName = parts[1] // prefix/software/default.yaml
			} else {
				return nil // Skip invalid paths
			}
			
			softwareList = append(softwareList, softwareName)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list software: %w", err)
	}

	return softwareList, nil
}