package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ApplyConfig represents the structure of the YAML configuration file
type ApplyConfig struct {
	SAI struct {
		Status  []string `yaml:"status"`
		Check   []string `yaml:"check"`
		Install []string `yaml:"install"`
		Start   []string `yaml:"start"`
		Enable  []string `yaml:"enable"`
		// Add more actions as needed
	} `yaml:"sai"`
}

// ApplyHandler handles the apply action
type ApplyHandler struct {
	BaseHandler
}

// NewApplyHandler creates a new ApplyHandler
func NewApplyHandler() *ApplyHandler {
	return &ApplyHandler{
		BaseHandler: BaseHandler{
			Action: "apply",
		},
	}
}

// Handle executes the apply action
func (h *ApplyHandler) Handle(configFile, provider string) {
	if configFile == "" {
		fmt.Println("Error: No configuration file specified")
		return
	}

	// Read and parse the YAML file
	config, err := h.readConfig(configFile)
	if err != nil {
		fmt.Printf("Error reading configuration file: %v\n", err)
		return
	}

	// Execute each action in sequence
	h.executeActions(config, provider)
}

// readConfig reads and parses the YAML configuration file
func (h *ApplyHandler) readConfig(configFile string) (*ApplyConfig, error) {
	// Get absolute path if relative
	absPath, err := filepath.Abs(configFile)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path: %v", err)
	}

	// Read the file
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Parse YAML
	var config ApplyConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %v", err)
	}

	return &config, nil
}

// executeActions executes the actions defined in the configuration
func (h *ApplyHandler) executeActions(config *ApplyConfig, provider string) {
	// Execute status checks first
	for _, software := range config.SAI.Status {
		fmt.Printf("Checking status of %s...\n", software)
		NewStatusHandler().Handle(software, provider)
	}

	// Execute install actions
	for _, software := range config.SAI.Install {
		fmt.Printf("Installing %s...\n", software)
		NewInstallHandler().Handle(software, provider)
	}

	// Execute start actions
	for _, software := range config.SAI.Start {
		fmt.Printf("Starting %s...\n", software)
		NewStartHandler().Handle(software, provider)
	}

	// Execute enable actions
	for _, software := range config.SAI.Enable {
		fmt.Printf("Enabling %s...\n", software)
		NewEnableHandler().Handle(software, provider)
	}

	// Execute check actions
	for _, software := range config.SAI.Check {
		fmt.Printf("Checking %s...\n", software)
		NewCheckHandler().Handle(software, provider)
	}
}
