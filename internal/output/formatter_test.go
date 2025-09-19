package output

import (
	"os"
	"strings"
	"testing"

	"sai/internal/config"
)

func TestNewOutputFormatter(t *testing.T) {
	cfg := &config.Config{
		Output: config.OutputConfig{
			ProviderColor: "blue",
			SuccessColor:  "green",
			ErrorColor:    "red",
			ShowCommands:  true,
			ShowExitCodes: true,
		},
	}

	formatter := NewOutputFormatter(cfg, false, false, false)

	if formatter.config != cfg {
		t.Error("Expected config to be set")
	}

	if formatter.verboseMode {
		t.Error("Expected verbose mode to be false")
	}

	if formatter.quietMode {
		t.Error("Expected quiet mode to be false")
	}

	if formatter.jsonMode {
		t.Error("Expected JSON mode to be false")
	}
}

func TestFormatCommand(t *testing.T) {
	cfg := &config.Config{
		Output: config.OutputConfig{
			ProviderColor: "blue",
			ShowCommands:  true,
		},
	}

	tests := []struct {
		name     string
		jsonMode bool
		command  string
		provider string
	}{
		{
			name:     "normal mode",
			jsonMode: false,
			command:  "apt install nginx",
			provider: "apt",
		},
		{
			name:     "json mode",
			jsonMode: true,
			command:  "brew install nginx",
			provider: "brew",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewOutputFormatter(cfg, false, false, tt.jsonMode)
			result := formatter.FormatCommand(tt.command, tt.provider)

			if tt.jsonMode {
				if result != tt.command {
					t.Errorf("Expected JSON mode to return unformatted command, got %s", result)
				}
			} else {
				if !strings.Contains(result, tt.provider) {
					t.Errorf("Expected result to contain provider name, got %s", result)
				}
				if !strings.Contains(result, tt.command) {
					t.Errorf("Expected result to contain command, got %s", result)
				}
			}
		})
	}
}

func TestFormatProviderName(t *testing.T) {
	cfg := &config.Config{
		Output: config.OutputConfig{
			ProviderColor: "blue",
		},
	}

	tests := []struct {
		name     string
		jsonMode bool
		provider string
		expected string
	}{
		{
			name:     "normal mode",
			jsonMode: false,
			provider: "apt",
			expected: "[apt]", // Will be formatted but we check basic structure
		},
		{
			name:     "json mode",
			jsonMode: true,
			provider: "brew",
			expected: "[brew]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewOutputFormatter(cfg, false, false, tt.jsonMode)
			result := formatter.FormatProviderName(tt.provider)

			if tt.jsonMode {
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			} else {
				// In normal mode, result should contain the provider name
				if !strings.Contains(result, tt.provider) {
					t.Errorf("Expected result to contain provider name, got %s", result)
				}
			}
		})
	}
}

func TestFormatCommandOutput(t *testing.T) {
	cfg := &config.Config{
		Output: config.OutputConfig{
			SuccessColor:  "green",
			ErrorColor:    "red",
			ShowExitCodes: true,
		},
	}

	tests := []struct {
		name      string
		quietMode bool
		jsonMode  bool
		output    string
		exitCode  int
		expectEmpty bool
	}{
		{
			name:      "success in normal mode",
			quietMode: false,
			jsonMode:  false,
			output:    "Installation successful",
			exitCode:  0,
			expectEmpty: false,
		},
		{
			name:      "success in quiet mode",
			quietMode: true,
			jsonMode:  false,
			output:    "Installation successful",
			exitCode:  0,
			expectEmpty: true,
		},
		{
			name:      "failure in quiet mode",
			quietMode: true,
			jsonMode:  false,
			output:    "Installation failed",
			exitCode:  1,
			expectEmpty: false,
		},
		{
			name:      "json mode",
			quietMode: false,
			jsonMode:  true,
			output:    "Installation successful",
			exitCode:  0,
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewOutputFormatter(cfg, false, tt.quietMode, tt.jsonMode)
			result := formatter.FormatCommandOutput(tt.output, tt.exitCode)

			if tt.expectEmpty && result != "" {
				t.Errorf("Expected empty result in quiet mode for success, got %s", result)
			}

			if !tt.expectEmpty && tt.jsonMode {
				if result != tt.output {
					t.Errorf("Expected JSON mode to return unformatted output, got %s", result)
				}
			}

			if !tt.expectEmpty && !tt.jsonMode && !tt.quietMode {
				if !strings.Contains(result, tt.output) {
					t.Errorf("Expected result to contain output, got %s", result)
				}
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	cfg := &config.Config{}
	formatter := NewOutputFormatter(cfg, false, false, true)

	data := map[string]interface{}{
		"action":   "install",
		"software": "nginx",
		"success":  true,
	}

	result := formatter.FormatJSON(data)

	if !strings.Contains(result, "install") {
		t.Error("Expected JSON to contain action")
	}

	if !strings.Contains(result, "nginx") {
		t.Error("Expected JSON to contain software")
	}

	if !strings.Contains(result, "true") {
		t.Error("Expected JSON to contain success status")
	}
}

func TestShowMethods(t *testing.T) {
	cfg := &config.Config{
		Output: config.OutputConfig{
			SuccessColor: "green",
			ErrorColor:   "red",
		},
	}

	tests := []struct {
		name      string
		quietMode bool
		jsonMode  bool
	}{
		{"normal mode", false, false},
		{"quiet mode", true, false},
		{"json mode", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewOutputFormatter(cfg, false, tt.quietMode, tt.jsonMode)

			// These methods should not panic
			formatter.ShowProgress("Testing progress")
			formatter.ShowError(os.ErrNotExist)
			formatter.ShowSuccess("Testing success")
			formatter.ShowWarning("Testing warning")
			formatter.ShowInfo("Testing info")
			formatter.ShowDebug("Testing debug")

			commands := []string{"apt update", "apt install nginx"}
			formatter.ShowCommandPreview(commands, "apt")
		})
	}
}

func TestModeCheckers(t *testing.T) {
	cfg := &config.Config{}

	tests := []struct {
		name        string
		verbose     bool
		quiet       bool
		json        bool
	}{
		{"all false", false, false, false},
		{"verbose only", true, false, false},
		{"quiet only", false, true, false},
		{"json only", false, false, true},
		{"verbose and json", true, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewOutputFormatter(cfg, tt.verbose, tt.quiet, tt.json)

			if formatter.IsVerboseMode() != tt.verbose {
				t.Errorf("Expected verbose mode to be %v", tt.verbose)
			}

			if formatter.IsQuietMode() != tt.quiet {
				t.Errorf("Expected quiet mode to be %v", tt.quiet)
			}

			if formatter.IsJSONMode() != tt.json {
				t.Errorf("Expected JSON mode to be %v", tt.json)
			}
		})
	}
}

func TestColorSupport(t *testing.T) {
	// Save original environment
	originalNoColor := os.Getenv("NO_COLOR")
	originalTerm := os.Getenv("TERM")
	originalCI := os.Getenv("CI")
	originalColorTerm := os.Getenv("COLORTERM")

	// Restore environment after test
	defer func() {
		os.Setenv("NO_COLOR", originalNoColor)
		os.Setenv("TERM", originalTerm)
		os.Setenv("CI", originalCI)
		os.Setenv("COLORTERM", originalColorTerm)
	}()

	tests := []struct {
		name        string
		noColor     string
		term        string
		ci          string
		expected    bool
	}{
		{"normal terminal", "", "xterm-256color", "", true},
		{"no color set", "1", "xterm-256color", "", false},
		{"dumb terminal", "", "dumb", "", false},
		{"no term", "", "", "", false},
		{"ci without colorterm", "", "xterm", "true", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NO_COLOR", tt.noColor)
			os.Setenv("TERM", tt.term)
			os.Setenv("CI", tt.ci)
			os.Setenv("COLORTERM", "") // Clear COLORTERM for consistent testing

			result := isColorSupported()
			if result != tt.expected {
				t.Errorf("Expected color support to be %v, got %v", tt.expected, result)
			}
		})
	}
}