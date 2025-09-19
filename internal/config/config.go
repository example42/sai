package config

import (
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	SaidataRepository string            `yaml:"saidata_repository"`
	DefaultProvider   string            `yaml:"default_provider"`
	ProviderPriority  map[string]int    `yaml:"provider_priority"`
	LogLevel          string            `yaml:"log_level"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	config := &Config{
		SaidataRepository: "https://github.com/sai-data/saidata.git",
		DefaultProvider:   "",
		ProviderPriority:  make(map[string]int),
		LogLevel:          "info",
	}

	if path == "" {
		return config, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return config, err
	}

	return config, nil
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