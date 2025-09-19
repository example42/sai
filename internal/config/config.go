package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"sai/internal/errors"
)

// Config represents the application configuration
type Config struct {
	SaidataRepository string                        `yaml:"saidata_repository"`
	DefaultProvider   string                        `yaml:"default_provider"`
	ProviderPriority  map[string]int                `yaml:"provider_priority"`
	Timeout           time.Duration                 `yaml:"timeout"`
	CacheDir          string                        `yaml:"cache_dir"`
	LogLevel          string                        `yaml:"log_level"`
	Confirmations     ConfirmationConfig            `yaml:"confirmations"`
	Output            OutputConfig                  `yaml:"output"`
	Repository        RepositoryConfig              `yaml:"repository"`
	Recovery          *errors.RecoveryConfig        `yaml:"recovery,omitempty"`
	CircuitBreaker    *errors.CircuitBreakerConfig  `yaml:"circuit_breaker,omitempty"`
}

// RepositoryConfig handles Git-based management with zip fallback (Requirement 8.4)
type RepositoryConfig struct {
	GitURL          string        `yaml:"git_url"`
	ZipFallbackURL  string        `yaml:"zip_fallback_url"`
	LocalPath       string        `yaml:"local_path"`
	UpdateInterval  time.Duration `yaml:"update_interval"`
	OfflineMode     bool          `yaml:"offline_mode"`
	AutoSetup       bool          `yaml:"auto_setup"`
}

// ConfirmationConfig controls confirmation prompts (Requirements 9.1, 9.2, 9.3, 9.4)
type ConfirmationConfig struct {
	Install       bool `yaml:"install"`       // System-changing operations require confirmation
	Uninstall     bool `yaml:"uninstall"`     // System-changing operations require confirmation
	Upgrade       bool `yaml:"upgrade"`       // System-changing operations require confirmation
	SystemChanges bool `yaml:"system_changes"` // System-changing operations require confirmation
	ServiceOps    bool `yaml:"service_ops"`   // Service start/stop/restart/enable/disable require confirmation
	InfoCommands  bool `yaml:"info_commands"` // Info commands execute without confirmation (default: false)
}

// OutputConfig controls output formatting (Requirements 7.2, 7.5, 7.6, 10.1, 10.2, 10.3)
type OutputConfig struct {
	ProviderColor    string `yaml:"provider_color"`
	CommandStyle     string `yaml:"command_style"`
	SuccessColor     string `yaml:"success_color"`
	ErrorColor       string `yaml:"error_color"`
	ShowCommands     bool   `yaml:"show_commands"`
	ShowExitCodes    bool   `yaml:"show_exit_codes"`
}

// LoadConfig loads configuration with file discovery, environment variables, and validation
func LoadConfig(configPath string) (*Config, error) {
	config := getDefaultConfig()

	// Discover configuration file if not explicitly provided
	if configPath == "" {
		var err error
		configPath, err = discoverConfigFile()
		if err != nil {
			// No config file found, use defaults
			return applyEnvironmentVariables(config), nil
		}
	}

	// Load configuration from file
	if configPath != "" {
		if err := loadConfigFromFile(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
		}
	}

	// Apply environment variable overrides
	config = applyEnvironmentVariables(config)

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// getDefaultConfig returns a configuration with default values
func getDefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".sai", "cache")
	
	return &Config{
		SaidataRepository: "https://github.com/sai-data/saidata.git",
		DefaultProvider:   "",
		ProviderPriority:  make(map[string]int),
		Timeout:           30 * time.Second,
		CacheDir:          cacheDir,
		LogLevel:          "info",
		Recovery:          errors.DefaultRecoveryConfig(),
		CircuitBreaker:    errors.DefaultCircuitBreakerConfig(),
		Confirmations: ConfirmationConfig{
			Install:       true,  // Require confirmation for system-changing operations
			Uninstall:     true,  // Require confirmation for system-changing operations
			Upgrade:       true,  // Require confirmation for system-changing operations
			SystemChanges: true,  // Require confirmation for system-changing operations
			ServiceOps:    true,  // Require confirmation for service operations
			InfoCommands:  false, // Info commands execute without confirmation
		},
		Output: OutputConfig{
			ProviderColor: "blue",
			CommandStyle:  "bold",
			SuccessColor:  "green",
			ErrorColor:    "red",
			ShowCommands:  true,
			ShowExitCodes: true,
		},
		Repository: RepositoryConfig{
			GitURL:         "https://github.com/sai-data/saidata.git",
			ZipFallbackURL: "https://github.com/sai-data/saidata/archive/main.zip",
			LocalPath:      filepath.Join(cacheDir, "saidata"),
			UpdateInterval: 24 * time.Hour,
			OfflineMode:    false,
			AutoSetup:      true,
		},
	}
}

// discoverConfigFile searches for configuration files in standard locations
func discoverConfigFile() (string, error) {
	// Configuration file search order (user home, system, current directory)
	searchPaths := []string{
		"./sai.yaml",           // Current directory
		"./sai.yml",            // Current directory (alternative extension)
		"./.sai/config.yaml",   // Current directory .sai folder
		"./.sai/config.yml",    // Current directory .sai folder (alternative extension)
	}

	// Add user home directory paths
	if homeDir, err := os.UserHomeDir(); err == nil {
		homePaths := []string{
			filepath.Join(homeDir, ".sai", "config.yaml"),
			filepath.Join(homeDir, ".sai", "config.yml"),
			filepath.Join(homeDir, ".config", "sai", "config.yaml"),
			filepath.Join(homeDir, ".config", "sai", "config.yml"),
		}
		searchPaths = append(searchPaths, homePaths...)
	}

	// Add system-wide paths
	systemPaths := []string{
		"/etc/sai/config.yaml",
		"/etc/sai/config.yml",
		"/usr/local/etc/sai/config.yaml",
		"/usr/local/etc/sai/config.yml",
	}
	searchPaths = append(searchPaths, systemPaths...)

	// Search for the first existing configuration file
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no configuration file found in standard locations")
}

// loadConfigFromFile loads configuration from a YAML file
func loadConfigFromFile(config *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return nil
}

// applyEnvironmentVariables applies environment variable overrides
func applyEnvironmentVariables(config *Config) *Config {
	// SAI_SAIDATA_REPOSITORY
	if repo := os.Getenv("SAI_SAIDATA_REPOSITORY"); repo != "" {
		config.SaidataRepository = repo
	}

	// SAI_DEFAULT_PROVIDER
	if provider := os.Getenv("SAI_DEFAULT_PROVIDER"); provider != "" {
		config.DefaultProvider = provider
	}

	// SAI_LOG_LEVEL
	if level := os.Getenv("SAI_LOG_LEVEL"); level != "" {
		config.LogLevel = level
	}

	// SAI_CACHE_DIR
	if cacheDir := os.Getenv("SAI_CACHE_DIR"); cacheDir != "" {
		config.CacheDir = cacheDir
	}

	// SAI_TIMEOUT
	if timeout := os.Getenv("SAI_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil {
			config.Timeout = duration
		}
	}

	// SAI_OFFLINE_MODE
	if offline := os.Getenv("SAI_OFFLINE_MODE"); offline != "" {
		config.Repository.OfflineMode = strings.ToLower(offline) == "true"
	}

	// SAI_AUTO_SETUP
	if autoSetup := os.Getenv("SAI_AUTO_SETUP"); autoSetup != "" {
		config.Repository.AutoSetup = strings.ToLower(autoSetup) == "true"
	}

	return config
}

// validateConfig validates the configuration values
func validateConfig(config *Config) error {
	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, config.LogLevel) {
		return fmt.Errorf("invalid log level '%s', must be one of: %s", 
			config.LogLevel, strings.Join(validLogLevels, ", "))
	}

	// Validate timeout
	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got: %v", config.Timeout)
	}

	// Validate cache directory
	if config.CacheDir == "" {
		return fmt.Errorf("cache directory cannot be empty")
	}

	// Validate repository configuration
	if config.Repository.GitURL == "" && config.Repository.ZipFallbackURL == "" {
		return fmt.Errorf("either git_url or zip_fallback_url must be specified")
	}

	if config.Repository.LocalPath == "" {
		return fmt.Errorf("repository local_path cannot be empty")
	}

	if config.Repository.UpdateInterval <= 0 {
		return fmt.Errorf("repository update_interval must be positive, got: %v", config.Repository.UpdateInterval)
	}

	// Validate output colors
	validColors := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}
	if !contains(validColors, config.Output.ProviderColor) {
		return fmt.Errorf("invalid provider color '%s', must be one of: %s", 
			config.Output.ProviderColor, strings.Join(validColors, ", "))
	}

	if !contains(validColors, config.Output.SuccessColor) {
		return fmt.Errorf("invalid success color '%s', must be one of: %s", 
			config.Output.SuccessColor, strings.Join(validColors, ", "))
	}

	if !contains(validColors, config.Output.ErrorColor) {
		return fmt.Errorf("invalid error color '%s', must be one of: %s", 
			config.Output.ErrorColor, strings.Join(validColors, ", "))
	}

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// SetupLogging configures the logging system
func SetupLogging(level string) {
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	
	logrus.SetLevel(logLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

// GetConfigPaths returns all possible configuration file paths in search order
func GetConfigPaths() []string {
	paths := []string{
		"./sai.yaml",
		"./sai.yml",
		"./.sai/config.yaml",
		"./.sai/config.yml",
	}

	if homeDir, err := os.UserHomeDir(); err == nil {
		homePaths := []string{
			filepath.Join(homeDir, ".sai", "config.yaml"),
			filepath.Join(homeDir, ".sai", "config.yml"),
			filepath.Join(homeDir, ".config", "sai", "config.yaml"),
			filepath.Join(homeDir, ".config", "sai", "config.yml"),
		}
		paths = append(paths, homePaths...)
	}

	systemPaths := []string{
		"/etc/sai/config.yaml",
		"/etc/sai/config.yml",
		"/usr/local/etc/sai/config.yaml",
		"/usr/local/etc/sai/config.yml",
	}
	paths = append(paths, systemPaths...)

	return paths
}

// SaveConfig saves the configuration to a YAML file
func SaveConfig(config *Config, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// RequiresConfirmation checks if an action requires user confirmation
func (c *Config) RequiresConfirmation(action string) bool {
	switch action {
	case "install":
		return c.Confirmations.Install
	case "uninstall":
		return c.Confirmations.Uninstall
	case "upgrade":
		return c.Confirmations.Upgrade
	case "start", "stop", "restart", "enable", "disable":
		return c.Confirmations.ServiceOps
	case "search", "info", "version", "status", "logs", "config", "check", "cpu", "memory", "io", "list", "stats":
		return c.Confirmations.InfoCommands
	default:
		return c.Confirmations.SystemChanges
	}
}

// IsSystemChangingAction determines if an action changes system state
func (c *Config) IsSystemChangingAction(action string) bool {
	systemChangingActions := []string{
		"install", "uninstall", "upgrade",
		"start", "stop", "restart", "enable", "disable",
		"apply",
	}
	
	for _, sysAction := range systemChangingActions {
		if action == sysAction {
			return true
		}
	}
	return false
}

// IsInformationOnlyAction determines if an action only displays information
func (c *Config) IsInformationOnlyAction(action string) bool {
	infoOnlyActions := []string{
		"search", "info", "version", "status",
		"logs", "config", "check", "cpu", "memory", "io",
		"list", "stats", "saidata",
	}
	
	for _, infoAction := range infoOnlyActions {
		if action == infoAction {
			return true
		}
	}
	return false
}