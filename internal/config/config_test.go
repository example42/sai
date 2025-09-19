package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetDefaultConfig(t *testing.T) {
	config := getDefaultConfig()

	if config.SaidataRepository == "" {
		t.Error("Expected default saidata repository to be set")
	}

	if config.LogLevel != "info" {
		t.Errorf("Expected default log level to be 'info', got '%s'", config.LogLevel)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout to be 30s, got %v", config.Timeout)
	}

	if !config.Confirmations.Install {
		t.Error("Expected install confirmation to be enabled by default")
	}

	if config.Confirmations.InfoCommands {
		t.Error("Expected info commands confirmation to be disabled by default")
	}

	if config.Output.ProviderColor != "blue" {
		t.Errorf("Expected default provider color to be 'blue', got '%s'", config.Output.ProviderColor)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  getDefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: func() *Config {
				c := getDefaultConfig()
				c.LogLevel = "invalid"
				return c
			}(),
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: func() *Config {
				c := getDefaultConfig()
				c.Timeout = -1 * time.Second
				return c
			}(),
			wantErr: true,
		},
		{
			name: "empty cache dir",
			config: func() *Config {
				c := getDefaultConfig()
				c.CacheDir = ""
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid provider color",
			config: func() *Config {
				c := getDefaultConfig()
				c.Output.ProviderColor = "invalid"
				return c
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplyEnvironmentVariables(t *testing.T) {
	// Save original environment
	originalRepo := os.Getenv("SAI_SAIDATA_REPOSITORY")
	originalProvider := os.Getenv("SAI_DEFAULT_PROVIDER")
	originalLogLevel := os.Getenv("SAI_LOG_LEVEL")

	// Set test environment variables
	os.Setenv("SAI_SAIDATA_REPOSITORY", "https://test.example.com/repo.git")
	os.Setenv("SAI_DEFAULT_PROVIDER", "test-provider")
	os.Setenv("SAI_LOG_LEVEL", "debug")

	// Restore environment after test
	defer func() {
		os.Setenv("SAI_SAIDATA_REPOSITORY", originalRepo)
		os.Setenv("SAI_DEFAULT_PROVIDER", originalProvider)
		os.Setenv("SAI_LOG_LEVEL", originalLogLevel)
	}()

	config := getDefaultConfig()
	config = applyEnvironmentVariables(config)

	if config.SaidataRepository != "https://test.example.com/repo.git" {
		t.Errorf("Expected saidata repository to be overridden by environment variable")
	}

	if config.DefaultProvider != "test-provider" {
		t.Errorf("Expected default provider to be overridden by environment variable")
	}

	if config.LogLevel != "debug" {
		t.Errorf("Expected log level to be overridden by environment variable")
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	configContent := `
saidata_repository: "https://test.example.com/repo.git"
default_provider: "test-provider"
log_level: "debug"
timeout: "60s"
confirmations:
  install: false
  info_commands: true
output:
  provider_color: "red"
  success_color: "yellow"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := getDefaultConfig()
	if err := loadConfigFromFile(config, configPath); err != nil {
		t.Fatalf("Failed to load config from file: %v", err)
	}

	if config.SaidataRepository != "https://test.example.com/repo.git" {
		t.Errorf("Expected saidata repository to be loaded from file")
	}

	if config.DefaultProvider != "test-provider" {
		t.Errorf("Expected default provider to be loaded from file")
	}

	if config.LogLevel != "debug" {
		t.Errorf("Expected log level to be loaded from file")
	}

	if config.Timeout != 60*time.Second {
		t.Errorf("Expected timeout to be loaded from file")
	}

	if config.Confirmations.Install {
		t.Errorf("Expected install confirmation to be disabled from file")
	}

	if !config.Confirmations.InfoCommands {
		t.Errorf("Expected info commands confirmation to be enabled from file")
	}

	if config.Output.ProviderColor != "red" {
		t.Errorf("Expected provider color to be loaded from file")
	}
}

func TestRequiresConfirmation(t *testing.T) {
	config := getDefaultConfig()

	tests := []struct {
		action   string
		expected bool
	}{
		{"install", true},
		{"uninstall", true},
		{"upgrade", true},
		{"start", true},
		{"stop", true},
		{"restart", true},
		{"enable", true},
		{"disable", true},
		{"search", false},
		{"info", false},
		{"version", false},
		{"status", false},
		{"logs", false},
		{"list", false},
		{"stats", false},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			result := config.RequiresConfirmation(tt.action)
			if result != tt.expected {
				t.Errorf("RequiresConfirmation(%s) = %v, expected %v", tt.action, result, tt.expected)
			}
		})
	}
}

func TestIsSystemChangingAction(t *testing.T) {
	config := getDefaultConfig()

	tests := []struct {
		action   string
		expected bool
	}{
		{"install", true},
		{"uninstall", true},
		{"upgrade", true},
		{"start", true},
		{"stop", true},
		{"restart", true},
		{"enable", true},
		{"disable", true},
		{"apply", true},
		{"search", false},
		{"info", false},
		{"version", false},
		{"status", false},
		{"logs", false},
		{"list", false},
		{"stats", false},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			result := config.IsSystemChangingAction(tt.action)
			if result != tt.expected {
				t.Errorf("IsSystemChangingAction(%s) = %v, expected %v", tt.action, result, tt.expected)
			}
		})
	}
}

func TestIsInformationOnlyAction(t *testing.T) {
	config := getDefaultConfig()

	tests := []struct {
		action   string
		expected bool
	}{
		{"search", true},
		{"info", true},
		{"version", true},
		{"status", true},
		{"logs", true},
		{"config", true},
		{"check", true},
		{"cpu", true},
		{"memory", true},
		{"io", true},
		{"list", true},
		{"stats", true},
		{"saidata", true},
		{"install", false},
		{"uninstall", false},
		{"upgrade", false},
		{"start", false},
		{"stop", false},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			result := config.IsInformationOnlyAction(tt.action)
			if result != tt.expected {
				t.Errorf("IsInformationOnlyAction(%s) = %v, expected %v", tt.action, result, tt.expected)
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	config := getDefaultConfig()
	config.DefaultProvider = "test-provider"
	config.LogLevel = "debug"

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "saved-config.yaml")

	if err := SaveConfig(config, configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created")
	}

	// Load the saved config and verify
	loadedConfig := getDefaultConfig()
	if err := loadConfigFromFile(loadedConfig, configPath); err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.DefaultProvider != "test-provider" {
		t.Errorf("Expected saved default provider to be 'test-provider', got '%s'", loadedConfig.DefaultProvider)
	}

	if loadedConfig.LogLevel != "debug" {
		t.Errorf("Expected saved log level to be 'debug', got '%s'", loadedConfig.LogLevel)
	}
}