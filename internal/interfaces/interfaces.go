package interfaces

import (
	"context"
	"time"

	"sai/internal/types"
)

// ProviderManager handles dynamic provider loading and management
type ProviderManager interface {
	// LoadProviders loads all providers from the specified directory
	LoadProviders(providerDir string) error
	
	// GetProvider returns a provider by name
	GetProvider(name string) (*types.ProviderData, error)
	
	// GetAvailableProviders returns all available providers
	GetAvailableProviders() []*types.ProviderData
	
	// SelectProvider selects the best provider for a software and action
	SelectProvider(software string, action string, preferredProvider string) (*types.ProviderData, error)
	
	// IsProviderAvailable checks if a provider is available on the system
	IsProviderAvailable(name string) bool
	
	// GetProvidersForAction returns providers that support a specific action
	GetProvidersForAction(action string) []*types.ProviderData
	
	// ValidateProvider validates a provider configuration
	ValidateProvider(provider *types.ProviderData) error
	
	// ReloadProviders reloads all providers (useful for development)
	ReloadProviders() error
}

// SaidataManager manages software metadata and configurations
type SaidataManager interface {
	// LoadSoftware loads saidata for a specific software
	LoadSoftware(name string) (*types.SoftwareData, error)
	
	// GetProviderConfig returns provider-specific configuration for software
	GetProviderConfig(software string, provider string) (*types.ProviderConfig, error)
	
	// GenerateDefaults generates intelligent defaults for software without saidata
	GenerateDefaults(software string) (*types.SoftwareData, error)
	
	// UpdateRepository updates the saidata repository
	UpdateRepository() error
	
	// SearchSoftware searches for software in the repository
	SearchSoftware(query string) ([]*SoftwareInfo, error)
	
	// ValidateData validates saidata against schema
	ValidateData(data []byte) error
	
	// ManageRepositoryOperations handles repository management
	ManageRepositoryOperations() error
	
	// SynchronizeRepository synchronizes the local repository
	SynchronizeRepository() error
	
	// GetSoftwareList returns a list of available software
	GetSoftwareList() ([]string, error)
	
	// CacheData caches saidata for performance
	CacheData(software string, data *types.SoftwareData) error
	
	// GetCachedData retrieves cached saidata
	GetCachedData(software string) (*types.SoftwareData, error)
}

// ActionManager orchestrates software management operations
type ActionManager interface {
	// ExecuteAction executes a specific action on software
	ExecuteAction(ctx context.Context, action string, software string, options ActionOptions) (*ActionResult, error)
	
	// ValidateAction validates if an action can be performed
	ValidateAction(action string, software string) error
	
	// GetAvailableActions returns available actions for software
	GetAvailableActions(software string) ([]string, error)
	
	// GetActionInfo returns information about a specific action
	GetActionInfo(action string) (*ActionInfo, error)
	
	// ResolveSoftwareData resolves saidata or generates intelligent defaults
	ResolveSoftwareData(software string) (*types.SoftwareData, error)
	
	// ValidateResourcesExist validates that required resources exist
	ValidateResourcesExist(saidata *types.SoftwareData, action string) (*ResourceValidationResult, error)
	
	// GetAvailableProviders returns providers available for software and action
	GetAvailableProviders(software string, action string) ([]*ProviderOption, error)
	
	// RequiresConfirmation checks if an action requires user confirmation
	RequiresConfirmation(action string) bool
	
	// SearchAcrossProviders searches for software across all providers
	SearchAcrossProviders(software string) ([]*SearchResult, error)
	
	// GetSoftwareInfo gets information about software from all providers
	GetSoftwareInfo(software string) ([]*SoftwareInfo, error)
	
	// GetSoftwareVersions gets version information with installation status
	GetSoftwareVersions(software string) ([]*VersionInfo, error)
	
	// ManageRepositorySetup automatically sets up repositories from saidata
	ManageRepositorySetup(saidata *types.SoftwareData) error
}

// GenericExecutor executes provider actions with safety validation
type GenericExecutor interface {
	// Execute runs a provider action with the given options
	Execute(ctx context.Context, provider *types.ProviderData, action string, software string, saidata *types.SoftwareData, options ExecuteOptions) (*ExecutionResult, error)
	
	// ValidateAction validates that an action can be executed
	ValidateAction(provider *types.ProviderData, action string, software string, saidata *types.SoftwareData) error
	
	// ValidateResources validates that required resources exist
	ValidateResources(saidata *types.SoftwareData, action string) (*ResourceValidationResult, error)
	
	// DryRun shows what would be executed without running commands
	DryRun(ctx context.Context, provider *types.ProviderData, action string, software string, saidata *types.SoftwareData, options ExecuteOptions) (*ExecutionResult, error)
	
	// CanExecute checks if an action can be executed
	CanExecute(provider *types.ProviderData, action string, software string, saidata *types.SoftwareData) bool
	
	// RenderTemplate renders command templates with saidata variables
	RenderTemplate(template string, saidata *types.SoftwareData, provider *types.ProviderData) (string, error)
	
	// ExecuteCommand executes a single command with proper error handling
	ExecuteCommand(ctx context.Context, command string, options CommandOptions) (*CommandResult, error)
	
	// ExecuteSteps executes multiple steps in sequence
	ExecuteSteps(ctx context.Context, steps []types.Step, saidata *types.SoftwareData, provider *types.ProviderData, options ExecuteOptions) (*ExecutionResult, error)
}

// DefaultsGenerator generates intelligent defaults for missing saidata
type DefaultsGenerator interface {
	// GeneratePackageDefaults generates default package configurations
	GeneratePackageDefaults(software string) []types.Package
	
	// GenerateServiceDefaults generates default service configurations
	GenerateServiceDefaults(software string) []types.Service
	
	// GenerateFileDefaults generates default file configurations
	GenerateFileDefaults(software string) []types.File
	
	// GenerateDirectoryDefaults generates default directory configurations
	GenerateDirectoryDefaults(software string) []types.Directory
	
	// GenerateCommandDefaults generates default command configurations
	GenerateCommandDefaults(software string) []types.Command
	
	// GeneratePortDefaults generates default port configurations
	GeneratePortDefaults(software string) []types.Port
	
	// ValidatePathExists checks if a path exists on the system
	ValidatePathExists(path string) bool
	
	// ValidateServiceExists checks if a service exists on the system
	ValidateServiceExists(service string) bool
	
	// ValidateCommandExists checks if a command exists on the system
	ValidateCommandExists(command string) bool
	
	// GenerateDefaults generates complete default saidata
	GenerateDefaults(software string) (*types.SoftwareData, error)
}

// ResourceValidator validates system resources
type ResourceValidator interface {
	// ValidateFile checks if a file exists and is accessible
	ValidateFile(file types.File) bool
	
	// ValidateService checks if a service exists
	ValidateService(service types.Service) bool
	
	// ValidateCommand checks if a command exists and is executable
	ValidateCommand(command types.Command) bool
	
	// ValidateDirectory checks if a directory exists
	ValidateDirectory(directory types.Directory) bool
	
	// ValidatePort checks if a port configuration is valid
	ValidatePort(port types.Port) bool
	
	// ValidateContainer checks if a container configuration is valid
	ValidateContainer(container types.Container) bool
	
	// ValidateResources validates all resources in saidata
	ValidateResources(saidata *types.SoftwareData) (*ResourceValidationResult, error)
	
	// ValidateSystemRequirements checks system requirements
	ValidateSystemRequirements(requirements *types.Requirements) (*SystemValidationResult, error)
}

// ConfigManager manages application configuration
type ConfigManager interface {
	// LoadConfig loads configuration from file
	LoadConfig(configPath string) (*Config, error)
	
	// SaveConfig saves configuration to file
	SaveConfig(config *Config, configPath string) error
	
	// GetDefaultConfig returns default configuration
	GetDefaultConfig() *Config
	
	// MergeConfig merges configurations with precedence
	MergeConfig(base *Config, override *Config) *Config
	
	// ValidateConfig validates configuration
	ValidateConfig(config *Config) error
	
	// GetConfigPaths returns possible configuration file paths
	GetConfigPaths() []string
	
	// WatchConfig watches for configuration changes
	WatchConfig(configPath string, callback func(*Config)) error
}

// TemplateEngine provides template rendering with saidata functions
type TemplateEngine interface {
	// Render renders a template string with the given context
	Render(templateStr string, context *TemplateContext) (string, error)
	
	// ValidateTemplate validates a template string without executing it
	ValidateTemplate(templateStr string) error
	
	// SetSafetyMode enables or disables safety mode
	SetSafetyMode(enabled bool)
	
	// SetSaidata sets the current saidata context for template functions
	SetSaidata(saidata *types.SoftwareData)
}

// TemplateContext holds the context for template rendering
type TemplateContext struct {
	Software  string
	Provider  string
	Saidata   *types.SoftwareData
	Variables map[string]string
}

// Logger provides structured logging
type Logger interface {
	// Debug logs debug messages
	Debug(msg string, fields ...LogField)
	
	// Info logs info messages
	Info(msg string, fields ...LogField)
	
	// Warn logs warning messages
	Warn(msg string, fields ...LogField)
	
	// Error logs error messages
	Error(msg string, err error, fields ...LogField)
	
	// Fatal logs fatal messages and exits
	Fatal(msg string, err error, fields ...LogField)
	
	// WithFields returns a logger with additional fields
	WithFields(fields ...LogField) Logger
	
	// SetLevel sets the logging level
	SetLevel(level LogLevel)
	
	// GetLevel returns the current logging level
	GetLevel() LogLevel
}

// Data structures for interface parameters and results

// ActionOptions contains options for action execution
type ActionOptions struct {
	Provider    string
	DryRun      bool
	Verbose     bool
	Quiet       bool
	Yes         bool
	JSON        bool
	Config      string
	Variables   map[string]string
	Timeout     time.Duration
}

// ExecuteOptions contains options for command execution
type ExecuteOptions struct {
	DryRun    bool
	Verbose   bool
	Timeout   time.Duration
	Variables map[string]string
	WorkDir   string
	Env       map[string]string
}

// CommandOptions contains options for single command execution
type CommandOptions struct {
	Timeout   time.Duration
	WorkDir   string
	Env       map[string]string
	Input     string
	Verbose   bool
}

// ActionResult contains the result of an action execution
type ActionResult struct {
	Action               string
	Software             string
	Provider             string
	Success              bool
	Output               string
	Error                error
	Duration             time.Duration
	Commands             []string
	Changes              []Change
	ExitCode             int
	RequiredConfirmation bool
}

// ExecutionResult contains the result of a command execution
type ExecutionResult struct {
	Success      bool
	Output       string
	Error        error
	ExitCode     int
	Duration     time.Duration
	Commands     []string
	Provider     string
	Changes      []Change
}

// CommandResult contains the result of a single command
type CommandResult struct {
	Command  string
	Output   string
	Error    error
	ExitCode int
	Duration time.Duration
}

// Change represents a system change made during execution
type Change struct {
	Type        string // "file", "service", "package", etc.
	Resource    string
	Action      string // "created", "modified", "deleted", "started", "stopped"
	OldValue    string
	NewValue    string
	Reversible  bool
	RollbackCmd string
}

// ActionInfo contains information about an action
type ActionInfo struct {
	Name         string
	Description  string
	RequiresRoot bool
	Timeout      time.Duration
	Capabilities []string
	Providers    []string
}

// ProviderOption represents a provider option for user selection
type ProviderOption struct {
	Provider    *types.ProviderData
	PackageName string
	Version     string
	IsInstalled bool
	Priority    int
}

// SearchResult represents a search result across providers
type SearchResult struct {
	Software    string
	Provider    string
	PackageName string
	Version     string
	Description string
	Available   bool
}

// SoftwareInfo represents software information from providers
type SoftwareInfo struct {
	Software     string
	Provider     string
	PackageName  string
	Version      string
	Description  string
	Homepage     string
	License      string
	Dependencies []string
}

// VersionInfo represents version information with installation status
type VersionInfo struct {
	Software      string
	Provider      string
	PackageName   string
	Version       string
	IsInstalled   bool
	LatestVersion string
}

// ResourceValidationResult contains resource validation results
type ResourceValidationResult struct {
	Valid              bool
	MissingFiles       []string
	MissingDirectories []string
	MissingCommands    []string
	MissingServices    []string
	InvalidPorts       []int
	Warnings           []string
	CanProceed         bool
}

// SystemValidationResult contains system validation results
type SystemValidationResult struct {
	Valid                bool
	InsufficientMemory   bool
	InsufficientDisk     bool
	MissingDependencies  []string
	UnsupportedPlatform  bool
	Warnings             []string
}

// Config represents application configuration
type Config struct {
	SaidataRepository string                 `yaml:"saidata_repository" json:"saidata_repository"`
	DefaultProvider   string                 `yaml:"default_provider" json:"default_provider"`
	ProviderPriority  map[string]int         `yaml:"provider_priority" json:"provider_priority"`
	Timeout           time.Duration          `yaml:"timeout" json:"timeout"`
	CacheDir          string                 `yaml:"cache_dir" json:"cache_dir"`
	LogLevel          string                 `yaml:"log_level" json:"log_level"`
	Confirmations     ConfirmationConfig     `yaml:"confirmations" json:"confirmations"`
	Output            OutputConfig           `yaml:"output" json:"output"`
	Repository        RepositoryConfig       `yaml:"repository" json:"repository"`
	Providers         map[string]interface{} `yaml:"providers" json:"providers"`
}

// ConfirmationConfig defines confirmation behavior
type ConfirmationConfig struct {
	Install       bool `yaml:"install" json:"install"`
	Uninstall     bool `yaml:"uninstall" json:"uninstall"`
	Upgrade       bool `yaml:"upgrade" json:"upgrade"`
	SystemChanges bool `yaml:"system_changes" json:"system_changes"`
	ServiceOps    bool `yaml:"service_ops" json:"service_ops"`
	InfoCommands  bool `yaml:"info_commands" json:"info_commands"`
}

// OutputConfig defines output formatting
type OutputConfig struct {
	ProviderColor string `yaml:"provider_color" json:"provider_color"`
	CommandStyle  string `yaml:"command_style" json:"command_style"`
	SuccessColor  string `yaml:"success_color" json:"success_color"`
	ErrorColor    string `yaml:"error_color" json:"error_color"`
	ShowCommands  bool   `yaml:"show_commands" json:"show_commands"`
	ShowExitCodes bool   `yaml:"show_exit_codes" json:"show_exit_codes"`
}

// RepositoryConfig defines repository configuration
type RepositoryConfig struct {
	GitURL         string        `yaml:"git_url" json:"git_url"`
	ZipFallbackURL string        `yaml:"zip_fallback_url" json:"zip_fallback_url"`
	LocalPath      string        `yaml:"local_path" json:"local_path"`
	UpdateInterval time.Duration `yaml:"update_interval" json:"update_interval"`
	OfflineMode    bool          `yaml:"offline_mode" json:"offline_mode"`
	AutoSetup      bool          `yaml:"auto_setup" json:"auto_setup"`
}

// LogField represents a structured log field
type LogField struct {
	Key   string
	Value interface{}
}

// LogLevel represents logging levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String returns the string representation of log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	case LogLevelFatal:
		return "fatal"
	default:
		return "unknown"
	}
}