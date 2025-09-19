package saidata

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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
	// For now, we'll skip schema validation since we don't have the schema path
	// In a full implementation, this would be passed as a parameter
	return &Manager{
		saidataDir:        saidataDir,
		validator:         nil, // Skip schema validation for now
		cache:             make(map[string]*types.SoftwareData),
		defaultsGenerator: NewDefaultsGenerator(resourceValidator),
		resourceValidator: resourceValidator,
	}
}

// LoadSoftware loads saidata for a specific software with OS-specific overrides
func (m *Manager) LoadSoftware(name string) (*types.SoftwareData, error) {
	// Check cache first
	if cached, exists := m.cache[name]; exists {
		return cached, nil
	}

	// Generate prefix from software name (first 2 characters)
	prefix := generatePrefix(name)
	
	// Load base configuration
	basePath := filepath.Join(m.saidataDir, prefix, name, "default.yaml")
	baseData, err := m.loadSaidataFile(basePath)
	if err != nil {
		// Check if it's a file not found error (including nested path errors)
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file or directory") {
			baseData, err = m.GenerateDefaults(name)
			if err != nil {
				return nil, fmt.Errorf("failed to generate defaults for %s: %w", name, err)
			}
		} else {
			return nil, fmt.Errorf("failed to load base saidata for %s: %w", name, err)
		}
	}

	// Detect current OS and version
	osInfo, err := detectOSInfo()
	if err != nil {
		// If OS detection fails, return base data
		m.cache[name] = baseData
		return baseData, nil
	}

	// Try to load OS-specific override
	overridePath := filepath.Join(m.saidataDir, prefix, name, osInfo.OS, osInfo.Version+".yaml")
	if _, err := os.Stat(overridePath); err == nil {
		overrideData, err := m.loadSaidataFile(overridePath)
		if err != nil {
			// If override fails to load, log warning but continue with base data
			fmt.Printf("Warning: failed to load OS override %s: %v\n", overridePath, err)
		} else {
			// Deep merge override with base data
			baseData = m.mergeSaidata(baseData, overrideData)
		}
	}

	// Cache the result
	m.cache[name] = baseData
	return baseData, nil
}

// loadSaidataFile loads and validates a saidata YAML file
func (m *Manager) loadSaidataFile(filePath string) (*types.SoftwareData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Parse YAML
	saidata, err := types.LoadSoftwareDataFromYAML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML from %s: %w", filePath, err)
	}

	// Validate against schema (skip if no validator)
	if m.validator != nil {
		if err := m.validator.ValidateSaidata(saidata); err != nil {
			return nil, fmt.Errorf("validation failed for %s: %w", filePath, err)
		}
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
			if len(parts) >= 3 {
				softwareName := parts[1] // prefix/software/default.yaml
				
				// Check if software name matches query
				if strings.Contains(strings.ToLower(softwareName), strings.ToLower(query)) {
					// Load basic metadata
					saidata, err := m.loadSaidataFile(path)
					if err != nil {
						return nil // Skip invalid files
					}

					results = append(results, &interfaces.SoftwareInfo{
						Software:     softwareName,
						Provider:     "saidata",
						PackageName:  softwareName,
						Version:      saidata.Metadata.Version,
						Description:  saidata.Metadata.Description,
						Homepage:     "",
						License:      "",
						Dependencies: []string{},
					})
				}
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
	switch runtime.GOOS {
	case "linux":
		return detectLinuxInfo()
	case "darwin":
		return detectMacOSInfo()
	case "windows":
		return detectWindowsInfo()
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// detectLinuxInfo detects Linux distribution and version
func detectLinuxInfo() (*OSInfo, error) {
	// Try /etc/os-release first
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		return parseOSRelease(string(data))
	}

	// Try /etc/lsb-release
	if data, err := os.ReadFile("/etc/lsb-release"); err == nil {
		return parseLSBRelease(string(data))
	}

	// Fallback to generic linux
	return &OSInfo{OS: "linux", Version: "unknown"}, nil
}

// detectMacOSInfo detects macOS version
func detectMacOSInfo() (*OSInfo, error) {
	// Use sw_vers command to get macOS version
	// For now, return a default - this would be implemented with exec.Command
	return &OSInfo{OS: "macos", Version: "13"}, nil
}

// detectWindowsInfo detects Windows version
func detectWindowsInfo() (*OSInfo, error) {
	// For now, return a default - this would be implemented with Windows APIs
	return &OSInfo{OS: "windows", Version: "11"}, nil
}

// parseOSRelease parses /etc/os-release format
func parseOSRelease(content string) (*OSInfo, error) {
	var osID, versionID string
	
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ID=") {
			osID = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			versionID = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
		}
	}

	if osID == "" {
		return nil, fmt.Errorf("could not determine OS ID from os-release")
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

// UpdateRepository updates the saidata repository (placeholder for future implementation)
func (m *Manager) UpdateRepository() error {
	// This would implement Git-based repository updates
	// For now, return nil as this is not implemented yet
	return nil
}

// ManageRepositoryOperations manages saidata repository operations
func (m *Manager) ManageRepositoryOperations() error {
	// This would implement repository management operations
	// For now, return nil as this is not implemented yet
	return nil
}

// SynchronizeRepository synchronizes the saidata repository
func (m *Manager) SynchronizeRepository() error {
	// This would implement repository synchronization
	// For now, return nil as this is not implemented yet
	return nil
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
			if len(parts) >= 3 {
				softwareName := parts[1] // prefix/software/default.yaml
				softwareList = append(softwareList, softwareName)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list software: %w", err)
	}

	return softwareList, nil
}