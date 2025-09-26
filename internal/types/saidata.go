package types

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// SoftwareData represents the complete saidata structure for a software package
type SoftwareData struct {
	Version       string                       `yaml:"version" json:"version"`
	Metadata      Metadata                     `yaml:"metadata" json:"metadata"`
	Packages      []Package                    `yaml:"packages,omitempty" json:"packages,omitempty"`
	Services      []Service                    `yaml:"services,omitempty" json:"services,omitempty"`
	Files         []File                       `yaml:"files,omitempty" json:"files,omitempty"`
	Directories   []Directory                  `yaml:"directories,omitempty" json:"directories,omitempty"`
	Commands      []Command                    `yaml:"commands,omitempty" json:"commands,omitempty"`
	Ports         []Port                       `yaml:"ports,omitempty" json:"ports,omitempty"`
	Containers    []Container                  `yaml:"containers,omitempty" json:"containers,omitempty"`
	Sources       []Source                     `yaml:"sources,omitempty" json:"sources,omitempty"`
	Binaries      []Binary                     `yaml:"binaries,omitempty" json:"binaries,omitempty"`
	Scripts       []Script                     `yaml:"scripts,omitempty" json:"scripts,omitempty"`
	Providers     map[string]ProviderConfig    `yaml:"providers,omitempty" json:"providers,omitempty"`
	Compatibility *Compatibility              `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	Requirements  *Requirements                `yaml:"requirements,omitempty" json:"requirements,omitempty"`
	IsGenerated   bool                         `yaml:"-" json:"-"` // Runtime flag for generated defaults
}

// Metadata contains software metadata information
type Metadata struct {
	Name         string            `yaml:"name" json:"name"`
	DisplayName  string            `yaml:"display_name,omitempty" json:"display_name,omitempty"`
	Description  string            `yaml:"description,omitempty" json:"description,omitempty"`
	Version      string            `yaml:"version,omitempty" json:"version,omitempty"`
	Category     string            `yaml:"category,omitempty" json:"category,omitempty"`
	Subcategory  string            `yaml:"subcategory,omitempty" json:"subcategory,omitempty"`
	Tags         []string          `yaml:"tags,omitempty" json:"tags,omitempty"`
	License      string            `yaml:"license,omitempty" json:"license,omitempty"`
	Language     string            `yaml:"language,omitempty" json:"language,omitempty"`
	Maintainer   string            `yaml:"maintainer,omitempty" json:"maintainer,omitempty"`
	URLs         *URLs             `yaml:"urls,omitempty" json:"urls,omitempty"`
	Security     *SecurityMetadata `yaml:"security,omitempty" json:"security,omitempty"`
}

// URLs contains various URLs related to the software
type URLs struct {
	Website       string `yaml:"website,omitempty" json:"website,omitempty"`
	Documentation string `yaml:"documentation,omitempty" json:"documentation,omitempty"`
	Source        string `yaml:"source,omitempty" json:"source,omitempty"`
	Issues        string `yaml:"issues,omitempty" json:"issues,omitempty"`
	Support       string `yaml:"support,omitempty" json:"support,omitempty"`
	Download      string `yaml:"download,omitempty" json:"download,omitempty"`
	Changelog     string `yaml:"changelog,omitempty" json:"changelog,omitempty"`
	License       string `yaml:"license,omitempty" json:"license,omitempty"`
	SBOM          string `yaml:"sbom,omitempty" json:"sbom,omitempty"`
	Icon          string `yaml:"icon,omitempty" json:"icon,omitempty"`
}

// SecurityMetadata contains security-related information
type SecurityMetadata struct {
	CVEExceptions           []string `yaml:"cve_exceptions,omitempty" json:"cve_exceptions,omitempty"`
	SecurityContact         string   `yaml:"security_contact,omitempty" json:"security_contact,omitempty"`
	VulnerabilityDisclosure string   `yaml:"vulnerability_disclosure,omitempty" json:"vulnerability_disclosure,omitempty"`
	SBOMURL                 string   `yaml:"sbom_url,omitempty" json:"sbom_url,omitempty"`
	SigningKey              string   `yaml:"signing_key,omitempty" json:"signing_key,omitempty"`
}

// Package represents a software package
type Package struct {
	Name         string   `yaml:"name" json:"name"`
	PackageName  string   `yaml:"package_name,omitempty" json:"package_name,omitempty"` // New field for consistent naming
	Version      string   `yaml:"version,omitempty" json:"version,omitempty"`
	Alternatives []string `yaml:"alternatives,omitempty" json:"alternatives,omitempty"`
	InstallOptions string `yaml:"install_options,omitempty" json:"install_options,omitempty"`
	Repository   string   `yaml:"repository,omitempty" json:"repository,omitempty"`
	Checksum     string   `yaml:"checksum,omitempty" json:"checksum,omitempty"`
	Signature    string   `yaml:"signature,omitempty" json:"signature,omitempty"`
	DownloadURL  string   `yaml:"download_url,omitempty" json:"download_url,omitempty"`
	// Runtime validation flags
	Exists      bool `yaml:"-" json:"-"`
	IsInstalled bool `yaml:"-" json:"-"`
}

// Service represents a system service
type Service struct {
	Name        string   `yaml:"name" json:"name"`
	ServiceName string   `yaml:"service_name,omitempty" json:"service_name,omitempty"`
	Type        string   `yaml:"type,omitempty" json:"type,omitempty"`
	Enabled     bool     `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	ConfigFiles []string `yaml:"config_files,omitempty" json:"config_files,omitempty"`
	// Runtime validation flags
	Exists   bool `yaml:"-" json:"-"`
	IsActive bool `yaml:"-" json:"-"`
}

// File represents a file resource
type File struct {
	Name   string `yaml:"name" json:"name"`
	Path   string `yaml:"path" json:"path"`
	Type   string `yaml:"type,omitempty" json:"type,omitempty"`
	Owner  string `yaml:"owner,omitempty" json:"owner,omitempty"`
	Group  string `yaml:"group,omitempty" json:"group,omitempty"`
	Mode   string `yaml:"mode,omitempty" json:"mode,omitempty"`
	Backup bool   `yaml:"backup,omitempty" json:"backup,omitempty"`
	// Runtime validation flags
	Exists bool `yaml:"-" json:"-"`
}

// Directory represents a directory resource
type Directory struct {
	Name      string `yaml:"name" json:"name"`
	Path      string `yaml:"path" json:"path"`
	Owner     string `yaml:"owner,omitempty" json:"owner,omitempty"`
	Group     string `yaml:"group,omitempty" json:"group,omitempty"`
	Mode      string `yaml:"mode,omitempty" json:"mode,omitempty"`
	Recursive bool   `yaml:"recursive,omitempty" json:"recursive,omitempty"`
	// Runtime validation flags
	Exists bool `yaml:"-" json:"-"`
}

// Command represents an executable command
type Command struct {
	Name            string   `yaml:"name" json:"name"`
	Path            string   `yaml:"path,omitempty" json:"path,omitempty"`
	Arguments       []string `yaml:"arguments,omitempty" json:"arguments,omitempty"`
	Aliases         []string `yaml:"aliases,omitempty" json:"aliases,omitempty"`
	ShellCompletion bool     `yaml:"shell_completion,omitempty" json:"shell_completion,omitempty"`
	ManPage         string   `yaml:"man_page,omitempty" json:"man_page,omitempty"`
	// Runtime validation flags
	Exists bool `yaml:"-" json:"-"`
}

// Port represents a network port
type Port struct {
	Port        int    `yaml:"port" json:"port"`
	Protocol    string `yaml:"protocol,omitempty" json:"protocol,omitempty"`
	Service     string `yaml:"service,omitempty" json:"service,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	// Runtime validation flags
	IsOpen bool `yaml:"-" json:"-"`
}

// Container represents a container configuration
type Container struct {
	Name        string            `yaml:"name" json:"name"`
	Image       string            `yaml:"image" json:"image"`
	Tag         string            `yaml:"tag,omitempty" json:"tag,omitempty"`
	Registry    string            `yaml:"registry,omitempty" json:"registry,omitempty"`
	Platform    string            `yaml:"platform,omitempty" json:"platform,omitempty"`
	Ports       []string          `yaml:"ports,omitempty" json:"ports,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
	Networks    []string          `yaml:"networks,omitempty" json:"networks,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	// Runtime validation flags
	Exists    bool `yaml:"-" json:"-"`
	IsRunning bool `yaml:"-" json:"-"`
}

// ProviderConfig contains provider-specific configurations
type ProviderConfig struct {
	Prerequisites  []string        `yaml:"prerequisites,omitempty" json:"prerequisites,omitempty"`
	BuildCommands  []string        `yaml:"build_commands,omitempty" json:"build_commands,omitempty"`
	Packages       []Package       `yaml:"packages,omitempty" json:"packages,omitempty"`
	PackageSources []PackageSource `yaml:"package_sources,omitempty" json:"package_sources,omitempty"`
	Repositories   []Repository    `yaml:"repositories,omitempty" json:"repositories,omitempty"`
	Services       []Service       `yaml:"services,omitempty" json:"services,omitempty"`
	Files          []File          `yaml:"files,omitempty" json:"files,omitempty"`
	Directories    []Directory     `yaml:"directories,omitempty" json:"directories,omitempty"`
	Commands       []Command       `yaml:"commands,omitempty" json:"commands,omitempty"`
	Ports          []Port          `yaml:"ports,omitempty" json:"ports,omitempty"`
	Containers     []Container     `yaml:"containers,omitempty" json:"containers,omitempty"`
	Sources        []Source        `yaml:"sources,omitempty" json:"sources,omitempty"`
	Binaries       []Binary        `yaml:"binaries,omitempty" json:"binaries,omitempty"`
	Scripts        []Script        `yaml:"scripts,omitempty" json:"scripts,omitempty"`
}

// PackageSource represents a package source with priority
type PackageSource struct {
	Name        string    `yaml:"name" json:"name"`
	Priority    int       `yaml:"priority,omitempty" json:"priority,omitempty"`
	Recommended bool      `yaml:"recommended,omitempty" json:"recommended,omitempty"`
	Repository  string    `yaml:"repository" json:"repository"`
	Packages    []Package `yaml:"packages" json:"packages"`
	Notes       string    `yaml:"notes,omitempty" json:"notes,omitempty"`
}

// Repository represents a software repository
type Repository struct {
	Name        string      `yaml:"name" json:"name"`
	URL         string      `yaml:"url,omitempty" json:"url,omitempty"`
	Key         string      `yaml:"key,omitempty" json:"key,omitempty"`
	Type        string      `yaml:"type,omitempty" json:"type,omitempty"`
	Components  []string    `yaml:"components,omitempty" json:"components,omitempty"`
	Maintainer  string      `yaml:"maintainer,omitempty" json:"maintainer,omitempty"`
	Priority    int         `yaml:"priority,omitempty" json:"priority,omitempty"`
	Recommended bool        `yaml:"recommended,omitempty" json:"recommended,omitempty"`
	Enabled     bool        `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Notes       string      `yaml:"notes,omitempty" json:"notes,omitempty"`
	Packages    []Package   `yaml:"packages,omitempty" json:"packages,omitempty"`
	Services    []Service   `yaml:"services,omitempty" json:"services,omitempty"`
	Files       []File      `yaml:"files,omitempty" json:"files,omitempty"`
	Directories []Directory `yaml:"directories,omitempty" json:"directories,omitempty"`
	Commands    []Command   `yaml:"commands,omitempty" json:"commands,omitempty"`
	Ports       []Port      `yaml:"ports,omitempty" json:"ports,omitempty"`
	Containers  []Container `yaml:"containers,omitempty" json:"containers,omitempty"`
}

// Compatibility defines platform and version compatibility
type Compatibility struct {
	Matrix   []CompatibilityEntry `yaml:"matrix,omitempty" json:"matrix,omitempty"`
	Versions *VersionCompatibility `yaml:"versions,omitempty" json:"versions,omitempty"`
}

// CompatibilityEntry represents a single compatibility entry
type CompatibilityEntry struct {
	Provider     string      `yaml:"provider" json:"provider"`
	Platform     interface{} `yaml:"platform" json:"platform"` // Can be string or []string
	Architecture interface{} `yaml:"architecture,omitempty" json:"architecture,omitempty"` // Can be string or []string
	OS           interface{} `yaml:"os,omitempty" json:"os,omitempty"` // Can be string or []string
	OSVersion    interface{} `yaml:"os_version,omitempty" json:"os_version,omitempty"` // Can be string or []string
	Supported    bool        `yaml:"supported" json:"supported"`
	Notes        string      `yaml:"notes,omitempty" json:"notes,omitempty"`
	Tested       bool        `yaml:"tested,omitempty" json:"tested,omitempty"`
	Recommended  bool        `yaml:"recommended,omitempty" json:"recommended,omitempty"`
}

// VersionCompatibility defines version compatibility information
type VersionCompatibility struct {
	Latest        string `yaml:"latest,omitempty" json:"latest,omitempty"`
	Minimum       string `yaml:"minimum,omitempty" json:"minimum,omitempty"`
	LatestLTS     string `yaml:"latest_lts,omitempty" json:"latest_lts,omitempty"`
	LatestMinimum string `yaml:"latest_minimum,omitempty" json:"latest_minimum,omitempty"`
}

// Requirements defines system requirements
type Requirements struct {
	System      *SystemRequirements      `yaml:"system,omitempty" json:"system,omitempty"`
	Performance *PerformanceRequirements `yaml:"performance,omitempty" json:"performance,omitempty"`
}

// SystemRequirements defines minimum system requirements
type SystemRequirements struct {
	MemoryMin         string `yaml:"memory_min,omitempty" json:"memory_min,omitempty"`
	MemoryRecommended string `yaml:"memory_recommended,omitempty" json:"memory_recommended,omitempty"`
	DiskSpace         string `yaml:"disk_space,omitempty" json:"disk_space,omitempty"`
	JavaVersion       string `yaml:"java_version,omitempty" json:"java_version,omitempty"`
}

// PerformanceRequirements defines performance-related requirements
type PerformanceRequirements struct {
	HeapSize        string `yaml:"heap_size,omitempty" json:"heap_size,omitempty"`
	FileDescriptors string `yaml:"file_descriptors,omitempty" json:"file_descriptors,omitempty"`
	VirtualMemory   string `yaml:"virtual_memory,omitempty" json:"virtual_memory,omitempty"`
}

// LoadSoftwareDataFromYAML loads saidata from YAML bytes
func LoadSoftwareDataFromYAML(data []byte) (*SoftwareData, error) {
	var saidata SoftwareData
	if err := yaml.Unmarshal(data, &saidata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal saidata YAML: %w", err)
	}
	
	// Set default service names if not specified
	for i, service := range saidata.Services {
		if service.ServiceName == "" {
			saidata.Services[i].ServiceName = service.Name
		}
	}
	
	// Set default command paths if not specified
	for i, command := range saidata.Commands {
		if command.Path == "" {
			saidata.Commands[i].Path = fmt.Sprintf("/usr/bin/%s", command.Name)
		}
	}
	
	// Set default port protocols if not specified
	for i, port := range saidata.Ports {
		if port.Protocol == "" {
			saidata.Ports[i].Protocol = "tcp"
		}
	}
	
	return &saidata, nil
}

// ToJSON converts the saidata to JSON for validation
func (s *SoftwareData) ToJSON() ([]byte, error) {
	// Create a map to properly handle empty values for schema validation
	result := make(map[string]interface{})
	
	// Always include version (required)
	result["version"] = s.Version
	
	// Handle metadata (required)
	metadata := make(map[string]interface{})
	if s.Metadata.Name != "" {
		metadata["name"] = s.Metadata.Name
	}
	if s.Metadata.DisplayName != "" {
		metadata["display_name"] = s.Metadata.DisplayName
	}
	if s.Metadata.Description != "" {
		metadata["description"] = s.Metadata.Description
	}
	if s.Metadata.Version != "" {
		metadata["version"] = s.Metadata.Version
	}
	if s.Metadata.Category != "" {
		metadata["category"] = s.Metadata.Category
	}
	if s.Metadata.Subcategory != "" {
		metadata["subcategory"] = s.Metadata.Subcategory
	}
	if len(s.Metadata.Tags) > 0 {
		metadata["tags"] = s.Metadata.Tags
	}
	if s.Metadata.License != "" {
		metadata["license"] = s.Metadata.License
	}
	if s.Metadata.Language != "" {
		metadata["language"] = s.Metadata.Language
	}
	if s.Metadata.Maintainer != "" {
		metadata["maintainer"] = s.Metadata.Maintainer
	}
	if s.Metadata.URLs != nil {
		metadata["urls"] = s.Metadata.URLs
	}
	if s.Metadata.Security != nil {
		metadata["security"] = s.Metadata.Security
	}
	
	result["metadata"] = metadata
	
	// Add optional arrays only if they have content
	if len(s.Packages) > 0 {
		// Filter out packages with empty names for validation
		var validPackages []interface{}
		for _, pkg := range s.Packages {
			pkgMap := make(map[string]interface{})
			if pkg.Name != "" {
				pkgMap["name"] = pkg.Name
			}
			if pkg.PackageName != "" {
				pkgMap["package_name"] = pkg.PackageName
			}
			if pkg.Version != "" {
				pkgMap["version"] = pkg.Version
			}
			if len(pkg.Alternatives) > 0 {
				pkgMap["alternatives"] = pkg.Alternatives
			}
			if pkg.InstallOptions != "" {
				pkgMap["install_options"] = pkg.InstallOptions
			}
			if pkg.Repository != "" {
				pkgMap["repository"] = pkg.Repository
			}
			if pkg.Checksum != "" {
				pkgMap["checksum"] = pkg.Checksum
			}
			if pkg.Signature != "" {
				pkgMap["signature"] = pkg.Signature
			}
			if pkg.DownloadURL != "" {
				pkgMap["download_url"] = pkg.DownloadURL
			}
			validPackages = append(validPackages, pkgMap)
		}
		result["packages"] = validPackages
	}
	if len(s.Services) > 0 {
		// Filter out services with empty names for validation
		var validServices []interface{}
		for _, svc := range s.Services {
			svcMap := make(map[string]interface{})
			if svc.Name != "" {
				svcMap["name"] = svc.Name
			}
			if svc.ServiceName != "" {
				svcMap["service_name"] = svc.ServiceName
			}
			if svc.Type != "" {
				svcMap["type"] = svc.Type
			}
			if svc.Enabled {
				svcMap["enabled"] = svc.Enabled
			}
			if len(svc.ConfigFiles) > 0 {
				svcMap["config_files"] = svc.ConfigFiles
			}
			validServices = append(validServices, svcMap)
		}
		result["services"] = validServices
	}
	if len(s.Files) > 0 {
		result["files"] = s.Files
	}
	if len(s.Directories) > 0 {
		result["directories"] = s.Directories
	}
	if len(s.Commands) > 0 {
		result["commands"] = s.Commands
	}
	if len(s.Ports) > 0 {
		result["ports"] = s.Ports
	}
	if len(s.Containers) > 0 {
		result["containers"] = s.Containers
	}
	if len(s.Sources) > 0 {
		result["sources"] = s.Sources
	}
	if len(s.Binaries) > 0 {
		result["binaries"] = s.Binaries
	}
	if len(s.Scripts) > 0 {
		result["scripts"] = s.Scripts
	}
	if len(s.Providers) > 0 {
		result["providers"] = s.Providers
	}
	if s.Compatibility != nil {
		result["compatibility"] = s.Compatibility
	}
	if s.Requirements != nil {
		result["requirements"] = s.Requirements
	}
	
	return json.Marshal(result)
}

// GetPackageByName returns a package by name
func (s *SoftwareData) GetPackageByName(name string) *Package {
	for i, pkg := range s.Packages {
		if pkg.Name == name {
			return &s.Packages[i]
		}
	}
	return nil
}

// GetServiceByName returns a service by name
func (s *SoftwareData) GetServiceByName(name string) *Service {
	for i, service := range s.Services {
		if service.Name == name {
			return &s.Services[i]
		}
	}
	return nil
}

// GetFileByName returns a file by name
func (s *SoftwareData) GetFileByName(name string) *File {
	for i, file := range s.Files {
		if file.Name == name {
			return &s.Files[i]
		}
	}
	return nil
}

// GetDirectoryByName returns a directory by name
func (s *SoftwareData) GetDirectoryByName(name string) *Directory {
	for i, dir := range s.Directories {
		if dir.Name == name {
			return &s.Directories[i]
		}
	}
	return nil
}

// GetCommandByName returns a command by name
func (s *SoftwareData) GetCommandByName(name string) *Command {
	for i, cmd := range s.Commands {
		if cmd.Name == name {
			return &s.Commands[i]
		}
	}
	return nil
}

// GetPortByNumber returns a port by port number
func (s *SoftwareData) GetPortByNumber(portNum int) *Port {
	for i, port := range s.Ports {
		if port.Port == portNum {
			return &s.Ports[i]
		}
	}
	return nil
}

// GetContainerByName returns a container by name
func (s *SoftwareData) GetContainerByName(name string) *Container {
	for i, container := range s.Containers {
		if container.Name == name {
			return &s.Containers[i]
		}
	}
	return nil
}

// GetProviderConfig returns provider-specific configuration
func (s *SoftwareData) GetProviderConfig(providerName string) *ProviderConfig {
	if config, exists := s.Providers[providerName]; exists {
		return &config
	}
	return nil
}

// GetSourceByName returns a source by name
func (s *SoftwareData) GetSourceByName(name string) *Source {
	for i, source := range s.Sources {
		if source.Name == name {
			return &s.Sources[i]
		}
	}
	return nil
}

// GetSourceByIndex returns a source by index
func (s *SoftwareData) GetSourceByIndex(index int) *Source {
	if index >= 0 && index < len(s.Sources) {
		return &s.Sources[index]
	}
	return nil
}

// GetBinaryByName returns a binary by name
func (s *SoftwareData) GetBinaryByName(name string) *Binary {
	for i, binary := range s.Binaries {
		if binary.Name == name {
			return &s.Binaries[i]
		}
	}
	return nil
}

// GetBinaryByIndex returns a binary by index
func (s *SoftwareData) GetBinaryByIndex(index int) *Binary {
	if index >= 0 && index < len(s.Binaries) {
		return &s.Binaries[index]
	}
	return nil
}

// GetScriptByName returns a script by name
func (s *SoftwareData) GetScriptByName(name string) *Script {
	for i, script := range s.Scripts {
		if script.Name == name {
			return &s.Scripts[i]
		}
	}
	return nil
}

// GetScriptByIndex returns a script by index
func (s *SoftwareData) GetScriptByIndex(index int) *Script {
	if index >= 0 && index < len(s.Scripts) {
		return &s.Scripts[index]
	}
	return nil
}

// GetSourceByName returns a provider-specific source by name
func (p *ProviderConfig) GetSourceByName(name string) *Source {
	for i, source := range p.Sources {
		if source.Name == name {
			return &p.Sources[i]
		}
	}
	return nil
}

// GetSourceByIndex returns a provider-specific source by index
func (p *ProviderConfig) GetSourceByIndex(index int) *Source {
	if index >= 0 && index < len(p.Sources) {
		return &p.Sources[index]
	}
	return nil
}

// GetBinaryByName returns a provider-specific binary by name
func (p *ProviderConfig) GetBinaryByName(name string) *Binary {
	for i, binary := range p.Binaries {
		if binary.Name == name {
			return &p.Binaries[i]
		}
	}
	return nil
}

// GetBinaryByIndex returns a provider-specific binary by index
func (p *ProviderConfig) GetBinaryByIndex(index int) *Binary {
	if index >= 0 && index < len(p.Binaries) {
		return &p.Binaries[index]
	}
	return nil
}

// GetScriptByName returns a provider-specific script by name
func (p *ProviderConfig) GetScriptByName(name string) *Script {
	for i, script := range p.Scripts {
		if script.Name == name {
			return &p.Scripts[i]
		}
	}
	return nil
}

// GetScriptByIndex returns a provider-specific script by index
func (p *ProviderConfig) GetScriptByIndex(index int) *Script {
	if index >= 0 && index < len(p.Scripts) {
		return &p.Scripts[index]
	}
	return nil
}

// GetPlatformsAsStrings converts platform interface{} to []string
func (c *CompatibilityEntry) GetPlatformsAsStrings() []string {
	return interfaceToStringSlice(c.Platform)
}

// GetArchitecturesAsStrings converts architecture interface{} to []string
func (c *CompatibilityEntry) GetArchitecturesAsStrings() []string {
	return interfaceToStringSlice(c.Architecture)
}

// GetOSAsStrings converts OS interface{} to []string
func (c *CompatibilityEntry) GetOSAsStrings() []string {
	return interfaceToStringSlice(c.OS)
}

// GetOSVersionsAsStrings converts OSVersion interface{} to []string
func (c *CompatibilityEntry) GetOSVersionsAsStrings() []string {
	return interfaceToStringSlice(c.OSVersion)
}

// interfaceToStringSlice converts interface{} (string or []string) to []string
func interfaceToStringSlice(value interface{}) []string {
	if value == nil {
		return nil
	}
	
	switch v := value.(type) {
	case string:
		return []string{v}
	case []string:
		return v
	case []interface{}:
		var result []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return nil
	}
}

// GetFullImageName returns the full container image name with registry and tag
func (c *Container) GetFullImageName() string {
	imageName := c.Image
	if c.Registry != "" {
		imageName = c.Registry + "/" + imageName
	}
	if c.Tag != "" {
		imageName = imageName + ":" + c.Tag
	}
	return imageName
}

// GetServiceNameOrDefault returns the service name or defaults to the logical name
func (s *Service) GetServiceNameOrDefault() string {
	if s.ServiceName != "" {
		return s.ServiceName
	}
	return s.Name
}

// GetPathOrDefault returns the command path or generates a default
func (c *Command) GetPathOrDefault() string {
	if c.Path != "" {
		return c.Path
	}
	return fmt.Sprintf("/usr/bin/%s", c.Name)
}

// GetProtocolOrDefault returns the port protocol or defaults to TCP
func (p *Port) GetProtocolOrDefault() string {
	if p.Protocol != "" {
		return p.Protocol
	}
	return "tcp"
}

// GetPackageNameOrDefault returns the package_name field if available, otherwise falls back to name
func (p *Package) GetPackageNameOrDefault() string {
	if p.PackageName != "" {
		return p.PackageName
	}
	return p.Name
}

// Source represents a source code build configuration
type Source struct {
	Name            string                 `yaml:"name" json:"name"`
	URL             string                 `yaml:"url" json:"url"`
	Version         string                 `yaml:"version,omitempty" json:"version,omitempty"`
	BuildSystem     string                 `yaml:"build_system" json:"build_system"`
	BuildDir        string                 `yaml:"build_dir,omitempty" json:"build_dir,omitempty"`
	SourceDir       string                 `yaml:"source_dir,omitempty" json:"source_dir,omitempty"`
	InstallPrefix   string                 `yaml:"install_prefix,omitempty" json:"install_prefix,omitempty"`
	ConfigureArgs   []string               `yaml:"configure_args,omitempty" json:"configure_args,omitempty"`
	BuildArgs       []string               `yaml:"build_args,omitempty" json:"build_args,omitempty"`
	InstallArgs     []string               `yaml:"install_args,omitempty" json:"install_args,omitempty"`
	Prerequisites   []string               `yaml:"prerequisites,omitempty" json:"prerequisites,omitempty"`
	Environment     map[string]string      `yaml:"environment,omitempty" json:"environment,omitempty"`
	Checksum        string                 `yaml:"checksum,omitempty" json:"checksum,omitempty"`
	CustomCommands  *SourceCustomCommands  `yaml:"custom_commands,omitempty" json:"custom_commands,omitempty"`
}

// SourceCustomCommands defines custom commands for build step overrides
type SourceCustomCommands struct {
	Download   string `yaml:"download,omitempty" json:"download,omitempty"`
	Extract    string `yaml:"extract,omitempty" json:"extract,omitempty"`
	Configure  string `yaml:"configure,omitempty" json:"configure,omitempty"`
	Build      string `yaml:"build,omitempty" json:"build,omitempty"`
	Install    string `yaml:"install,omitempty" json:"install,omitempty"`
	Uninstall  string `yaml:"uninstall,omitempty" json:"uninstall,omitempty"`
	Validation string `yaml:"validation,omitempty" json:"validation,omitempty"`
	Version    string `yaml:"version,omitempty" json:"version,omitempty"`
}

// ValidateBuildSystem validates the build system type
func (s *Source) ValidateBuildSystem() error {
	validBuildSystems := []string{
		"autotools", "cmake", "make", "meson", "ninja", "custom",
		"configure", "automake", "autoconf", "bazel", "gradle", "maven",
	}
	
	for _, valid := range validBuildSystems {
		if s.BuildSystem == valid {
			return nil
		}
	}
	
	return fmt.Errorf("invalid build system '%s', must be one of: %v", s.BuildSystem, validBuildSystems)
}

// ValidateRequiredFields validates that required fields are present
func (s *Source) ValidateRequiredFields() error {
	if s.Name == "" {
		return fmt.Errorf("source name is required")
	}
	if s.URL == "" {
		return fmt.Errorf("source URL is required")
	}
	if s.BuildSystem == "" {
		return fmt.Errorf("build system is required")
	}
	return nil
}

// GenerateDefaults generates default values for build directories and install prefixes
func (s *Source) GenerateDefaults(softwareName string) {
	if s.BuildDir == "" {
		s.BuildDir = fmt.Sprintf("/tmp/sai-build-%s", softwareName)
	}
	if s.SourceDir == "" {
		version := s.Version
		if version == "" {
			version = "latest"
		}
		s.SourceDir = fmt.Sprintf("%s/%s-%s", s.BuildDir, softwareName, version)
	}
	if s.InstallPrefix == "" {
		s.InstallPrefix = "/usr/local"
	}
}

// GetBuildDirOrDefault returns the build directory or generates a default
func (s *Source) GetBuildDirOrDefault(softwareName string) string {
	if s.BuildDir != "" {
		return s.BuildDir
	}
	return fmt.Sprintf("/tmp/sai-build-%s", softwareName)
}

// GetSourceDirOrDefault returns the source directory or generates a default
func (s *Source) GetSourceDirOrDefault(softwareName string) string {
	if s.SourceDir != "" {
		return s.SourceDir
	}
	version := s.Version
	if version == "" {
		version = "latest"
	}
	buildDir := s.GetBuildDirOrDefault(softwareName)
	return fmt.Sprintf("%s/%s-%s", buildDir, softwareName, version)
}

// GetInstallPrefixOrDefault returns the install prefix or defaults to /usr/local
func (s *Source) GetInstallPrefixOrDefault() string {
	if s.InstallPrefix != "" {
		return s.InstallPrefix
	}
	return "/usr/local"
}

// Binary represents a binary download and installation configuration
type Binary struct {
	Name         string                 `yaml:"name" json:"name"`
	URL          string                 `yaml:"url" json:"url"`
	Version      string                 `yaml:"version,omitempty" json:"version,omitempty"`
	Architecture string                 `yaml:"architecture,omitempty" json:"architecture,omitempty"`
	Platform     string                 `yaml:"platform,omitempty" json:"platform,omitempty"`
	Checksum     string                 `yaml:"checksum,omitempty" json:"checksum,omitempty"`
	InstallPath  string                 `yaml:"install_path,omitempty" json:"install_path,omitempty"`
	Executable   string                 `yaml:"executable,omitempty" json:"executable,omitempty"`
	Archive      *ArchiveConfig         `yaml:"archive,omitempty" json:"archive,omitempty"`
	Permissions  string                 `yaml:"permissions,omitempty" json:"permissions,omitempty"`
	CustomCommands *BinaryCustomCommands `yaml:"custom_commands,omitempty" json:"custom_commands,omitempty"`
}

// ArchiveConfig defines configuration for handling compressed downloads
type ArchiveConfig struct {
	Format      string `yaml:"format,omitempty" json:"format,omitempty"`
	StripPrefix string `yaml:"strip_prefix,omitempty" json:"strip_prefix,omitempty"`
	ExtractPath string `yaml:"extract_path,omitempty" json:"extract_path,omitempty"`
}

// BinaryCustomCommands defines custom commands for installation step overrides
type BinaryCustomCommands struct {
	Download   string `yaml:"download,omitempty" json:"download,omitempty"`
	Extract    string `yaml:"extract,omitempty" json:"extract,omitempty"`
	Install    string `yaml:"install,omitempty" json:"install,omitempty"`
	Uninstall  string `yaml:"uninstall,omitempty" json:"uninstall,omitempty"`
	Validation string `yaml:"validation,omitempty" json:"validation,omitempty"`
	Version    string `yaml:"version,omitempty" json:"version,omitempty"`
}

// ValidateRequiredFields validates that required fields are present
func (b *Binary) ValidateRequiredFields() error {
	if b.Name == "" {
		return fmt.Errorf("binary name is required")
	}
	if b.URL == "" {
		return fmt.Errorf("binary URL is required")
	}
	return nil
}

// GenerateDefaults generates default values for binary installation
func (b *Binary) GenerateDefaults() {
	if b.InstallPath == "" {
		b.InstallPath = "/usr/local/bin"
	}
	if b.Permissions == "" {
		b.Permissions = "0755"
	}
	if b.Executable == "" {
		b.Executable = b.Name
	}
}

// GetInstallPathOrDefault returns the install path or defaults to /usr/local/bin
func (b *Binary) GetInstallPathOrDefault() string {
	if b.InstallPath != "" {
		return b.InstallPath
	}
	return "/usr/local/bin"
}

// GetPermissionsOrDefault returns the permissions or defaults to 0755
func (b *Binary) GetPermissionsOrDefault() string {
	if b.Permissions != "" {
		return b.Permissions
	}
	return "0755"
}

// GetExecutableOrDefault returns the executable name or defaults to the binary name
func (b *Binary) GetExecutableOrDefault() string {
	if b.Executable != "" {
		return b.Executable
	}
	return b.Name
}

// TemplateURL replaces OS/architecture placeholders in the URL
func (b *Binary) TemplateURL(osName, arch string) string {
	url := b.URL
	
	// Replace common OS placeholders
	switch osName {
	case "linux":
		url = strings.ReplaceAll(url, "{{.OS}}", "linux")
		url = strings.ReplaceAll(url, "{{.Platform}}", "linux")
	case "darwin":
		url = strings.ReplaceAll(url, "{{.OS}}", "darwin")
		url = strings.ReplaceAll(url, "{{.Platform}}", "macos")
	case "windows":
		url = strings.ReplaceAll(url, "{{.OS}}", "windows")
		url = strings.ReplaceAll(url, "{{.Platform}}", "windows")
	}
	
	// Replace architecture placeholders
	switch arch {
	case "amd64", "x86_64":
		url = strings.ReplaceAll(url, "{{.Arch}}", "amd64")
		url = strings.ReplaceAll(url, "{{.Architecture}}", "x86_64")
	case "arm64", "aarch64":
		url = strings.ReplaceAll(url, "{{.Arch}}", "arm64")
		url = strings.ReplaceAll(url, "{{.Architecture}}", "aarch64")
	case "386", "i386":
		url = strings.ReplaceAll(url, "{{.Arch}}", "386")
		url = strings.ReplaceAll(url, "{{.Architecture}}", "i386")
	}
	
	return url
}

// Script represents a script execution configuration
type Script struct {
	Name         string                 `yaml:"name" json:"name"`
	URL          string                 `yaml:"url" json:"url"`
	Version      string                 `yaml:"version,omitempty" json:"version,omitempty"`
	Interpreter  string                 `yaml:"interpreter,omitempty" json:"interpreter,omitempty"`
	Checksum     string                 `yaml:"checksum,omitempty" json:"checksum,omitempty"`
	Arguments    []string               `yaml:"arguments,omitempty" json:"arguments,omitempty"`
	Environment  map[string]string      `yaml:"environment,omitempty" json:"environment,omitempty"`
	WorkingDir   string                 `yaml:"working_dir,omitempty" json:"working_dir,omitempty"`
	Timeout      int                    `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	CustomCommands *ScriptCustomCommands `yaml:"custom_commands,omitempty" json:"custom_commands,omitempty"`
}

// ScriptCustomCommands defines custom commands for execution step overrides
type ScriptCustomCommands struct {
	Download   string `yaml:"download,omitempty" json:"download,omitempty"`
	Install    string `yaml:"install,omitempty" json:"install,omitempty"`
	Uninstall  string `yaml:"uninstall,omitempty" json:"uninstall,omitempty"`
	Validation string `yaml:"validation,omitempty" json:"validation,omitempty"`
	Version    string `yaml:"version,omitempty" json:"version,omitempty"`
}

// ValidateRequiredFields validates that required fields are present
func (s *Script) ValidateRequiredFields() error {
	if s.Name == "" {
		return fmt.Errorf("script name is required")
	}
	if s.URL == "" {
		return fmt.Errorf("script URL is required")
	}
	return nil
}

// GenerateDefaults generates default values for script execution
func (s *Script) GenerateDefaults() {
	if s.Interpreter == "" {
		// Auto-detect interpreter based on URL extension
		if strings.HasSuffix(s.URL, ".sh") {
			s.Interpreter = "bash"
		} else if strings.HasSuffix(s.URL, ".py") {
			s.Interpreter = "python3"
		} else if strings.HasSuffix(s.URL, ".pl") {
			s.Interpreter = "perl"
		} else if strings.HasSuffix(s.URL, ".rb") {
			s.Interpreter = "ruby"
		} else {
			s.Interpreter = "bash" // Default fallback
		}
	}
	if s.WorkingDir == "" {
		s.WorkingDir = "/tmp"
	}
	if s.Timeout == 0 {
		s.Timeout = 300 // 5 minutes default timeout
	}
}

// GetInterpreterOrDefault returns the interpreter or auto-detects based on URL
func (s *Script) GetInterpreterOrDefault() string {
	if s.Interpreter != "" {
		return s.Interpreter
	}
	
	// Auto-detect based on URL extension
	if strings.HasSuffix(s.URL, ".sh") {
		return "bash"
	} else if strings.HasSuffix(s.URL, ".py") {
		return "python3"
	} else if strings.HasSuffix(s.URL, ".pl") {
		return "perl"
	} else if strings.HasSuffix(s.URL, ".rb") {
		return "ruby"
	}
	
	return "bash" // Default fallback
}

// GetWorkingDirOrDefault returns the working directory or defaults to /tmp
func (s *Script) GetWorkingDirOrDefault() string {
	if s.WorkingDir != "" {
		return s.WorkingDir
	}
	return "/tmp"
}

// GetTimeoutOrDefault returns the timeout or defaults to 300 seconds
func (s *Script) GetTimeoutOrDefault() int {
	if s.Timeout > 0 {
		return s.Timeout
	}
	return 300 // 5 minutes default
}

// ValidateEnvironment validates environment variable names and values
func (s *Script) ValidateEnvironment() error {
	for key, value := range s.Environment {
		if key == "" {
			return fmt.Errorf("environment variable name cannot be empty")
		}
		if strings.Contains(key, "=") {
			return fmt.Errorf("environment variable name '%s' cannot contain '='", key)
		}
		if len(value) > 32768 { // 32KB limit for environment variables
			return fmt.Errorf("environment variable '%s' value exceeds maximum length", key)
		}
	}
	return nil
}