package ui

import (
	"strings"
	"testing"

	"sai/internal/config"
	"sai/internal/output"
)

func TestNewUserInterface(t *testing.T) {
	cfg := &config.Config{}
	formatter := output.NewOutputFormatter(cfg, false, false, false)
	ui := NewUserInterface(cfg, formatter)

	if ui.config != cfg {
		t.Error("Expected config to be set")
	}

	if ui.formatter != formatter {
		t.Error("Expected formatter to be set")
	}

	if ui.reader == nil {
		t.Error("Expected reader to be initialized")
	}
}

func TestShowProviderSelectionSingleOption(t *testing.T) {
	cfg := &config.Config{}
	formatter := output.NewOutputFormatter(cfg, false, false, false)
	ui := NewUserInterface(cfg, formatter)

	options := []*ProviderOption{
		{
			Name:        "apt",
			PackageName: "nginx",
			Version:     "1.18.0",
			IsInstalled: false,
			Description: "APT package manager",
		},
	}

	result, err := ui.ShowProviderSelection("nginx", options)
	if err != nil {
		t.Fatalf("Expected no error for single option, got %v", err)
	}

	if result != options[0] {
		t.Error("Expected to return the single option")
	}
}

func TestShowProviderSelectionNoOptions(t *testing.T) {
	cfg := &config.Config{}
	formatter := output.NewOutputFormatter(cfg, false, false, false)
	ui := NewUserInterface(cfg, formatter)

	options := []*ProviderOption{}

	_, err := ui.ShowProviderSelection("nginx", options)
	if err == nil {
		t.Error("Expected error for no options")
	}

	if !strings.Contains(err.Error(), "no providers available") {
		t.Errorf("Expected 'no providers available' error, got %v", err)
	}
}

func TestShowProviderSelectionJSONMode(t *testing.T) {
	cfg := &config.Config{}
	formatter := output.NewOutputFormatter(cfg, false, false, true) // JSON mode
	ui := NewUserInterface(cfg, formatter)

	options := []*ProviderOption{
		{
			Name:        "apt",
			PackageName: "nginx",
			Version:     "1.18.0",
			IsInstalled: false,
		},
		{
			Name:        "snap",
			PackageName: "nginx",
			Version:     "1.19.0",
			IsInstalled: true,
		},
	}

	_, err := ui.ShowProviderSelection("nginx", options)
	if err == nil {
		t.Error("Expected error in JSON mode for multiple options")
	}

	if !strings.Contains(err.Error(), "non-interactive mode") {
		t.Errorf("Expected 'non-interactive mode' error, got %v", err)
	}
}

func TestConfirmActionNoConfirmationRequired(t *testing.T) {
	cfg := &config.Config{
		Confirmations: config.ConfirmationConfig{
			InfoCommands: false, // Info commands don't require confirmation
		},
	}
	formatter := output.NewOutputFormatter(cfg, false, false, false)
	ui := NewUserInterface(cfg, formatter)

	// Mock the config method
	cfg.Confirmations.InfoCommands = false

	commands := []string{"nginx -t"}
	confirmed, err := ui.ConfirmAction("info", "nginx", "apt", commands)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !confirmed {
		t.Error("Expected confirmation to be true for info commands")
	}
}

func TestConfirmActionJSONMode(t *testing.T) {
	cfg := &config.Config{
		Confirmations: config.ConfirmationConfig{
			Install: true, // Install requires confirmation
		},
	}
	formatter := output.NewOutputFormatter(cfg, false, false, true) // JSON mode
	ui := NewUserInterface(cfg, formatter)

	commands := []string{"apt install nginx"}
	_, err := ui.ConfirmAction("install", "nginx", "apt", commands)
	if err == nil {
		t.Error("Expected error in JSON mode for confirmation")
	}

	if !strings.Contains(err.Error(), "non-interactive mode") {
		t.Errorf("Expected 'non-interactive mode' error, got %v", err)
	}
}

func TestPromptForInputJSONMode(t *testing.T) {
	cfg := &config.Config{}
	formatter := output.NewOutputFormatter(cfg, false, false, true) // JSON mode
	ui := NewUserInterface(cfg, formatter)

	_, err := ui.PromptForInput("Enter value: ")
	if err == nil {
		t.Error("Expected error in JSON mode for input prompt")
	}

	if !strings.Contains(err.Error(), "not supported in JSON mode") {
		t.Errorf("Expected 'not supported in JSON mode' error, got %v", err)
	}
}

func TestPromptForConfirmationJSONMode(t *testing.T) {
	cfg := &config.Config{}
	formatter := output.NewOutputFormatter(cfg, false, false, true) // JSON mode
	ui := NewUserInterface(cfg, formatter)

	_, err := ui.PromptForConfirmation("Continue?")
	if err == nil {
		t.Error("Expected error in JSON mode for confirmation prompt")
	}

	if !strings.Contains(err.Error(), "not supported in JSON mode") {
		t.Errorf("Expected 'not supported in JSON mode' error, got %v", err)
	}
}

func TestModeCheckers(t *testing.T) {
	cfg := &config.Config{}

	tests := []struct {
		name    string
		verbose bool
		quiet   bool
		json    bool
	}{
		{"all false", false, false, false},
		{"verbose only", true, false, false},
		{"quiet only", false, true, false},
		{"json only", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := output.NewOutputFormatter(cfg, tt.verbose, tt.quiet, tt.json)
			ui := NewUserInterface(cfg, formatter)

			if ui.IsVerboseMode() != tt.verbose {
				t.Errorf("Expected verbose mode to be %v", tt.verbose)
			}

			if ui.IsQuietMode() != tt.quiet {
				t.Errorf("Expected quiet mode to be %v", tt.quiet)
			}
		})
	}
}

func TestShowTable(t *testing.T) {
	cfg := &config.Config{}

	tests := []struct {
		name     string
		jsonMode bool
		headers  []string
		rows     [][]string
	}{
		{
			name:     "normal mode",
			jsonMode: false,
			headers:  []string{"Software", "Provider", "Status"},
			rows: [][]string{
				{"nginx", "apt", "installed"},
				{"docker", "snap", "available"},
			},
		},
		{
			name:     "json mode",
			jsonMode: true,
			headers:  []string{"Software", "Provider", "Status"},
			rows: [][]string{
				{"nginx", "apt", "installed"},
			},
		},
		{
			name:     "empty rows",
			jsonMode: false,
			headers:  []string{"Software", "Provider", "Status"},
			rows:     [][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := output.NewOutputFormatter(cfg, false, false, tt.jsonMode)
			ui := NewUserInterface(cfg, formatter)

			// This should not panic
			ui.ShowTable(tt.headers, tt.rows)
		})
	}
}

func TestProviderOption(t *testing.T) {
	option := &ProviderOption{
		Name:        "apt",
		PackageName: "nginx",
		Version:     "1.18.0",
		IsInstalled: true,
		Description: "APT package manager",
	}

	if option.Name != "apt" {
		t.Errorf("Expected name to be 'apt', got %s", option.Name)
	}

	if option.PackageName != "nginx" {
		t.Errorf("Expected package name to be 'nginx', got %s", option.PackageName)
	}

	if option.Version != "1.18.0" {
		t.Errorf("Expected version to be '1.18.0', got %s", option.Version)
	}

	if !option.IsInstalled {
		t.Error("Expected IsInstalled to be true")
	}

	if option.Description != "APT package manager" {
		t.Errorf("Expected description to be 'APT package manager', got %s", option.Description)
	}
}