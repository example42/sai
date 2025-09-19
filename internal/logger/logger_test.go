package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"sai/internal/config"
)

func TestNewLogger(t *testing.T) {
	cfg := &config.Config{
		LogLevel: "info",
	}

	logger := NewLogger(cfg, false, false, false)

	if logger.config != cfg {
		t.Error("Expected config to be set")
	}

	if logger.verboseMode {
		t.Error("Expected verbose mode to be false")
	}

	if logger.quietMode {
		t.Error("Expected quiet mode to be false")
	}

	if logger.jsonMode {
		t.Error("Expected JSON mode to be false")
	}
}

func TestLoggerSetup(t *testing.T) {
	tests := []struct {
		name        string
		configLevel string
		verbose     bool
		quiet       bool
		json        bool
		expectedLevel logrus.Level
	}{
		{
			name:          "default info level",
			configLevel:   "info",
			verbose:       false,
			quiet:         false,
			json:          false,
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "verbose overrides config",
			configLevel:   "warn",
			verbose:       true,
			quiet:         false,
			json:          false,
			expectedLevel: logrus.DebugLevel,
		},
		{
			name:          "quiet overrides config",
			configLevel:   "info",
			verbose:       false,
			quiet:         true,
			json:          false,
			expectedLevel: logrus.ErrorLevel,
		},
		{
			name:          "debug level",
			configLevel:   "debug",
			verbose:       false,
			quiet:         false,
			json:          false,
			expectedLevel: logrus.DebugLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				LogLevel: tt.configLevel,
			}

			logger := NewLogger(cfg, tt.verbose, tt.quiet, tt.json)

			if logger.GetLevel() != tt.expectedLevel {
				t.Errorf("Expected log level %v, got %v", tt.expectedLevel, logger.GetLevel())
			}
		})
	}
}

func TestLoggerFormatters(t *testing.T) {
	cfg := &config.Config{
		LogLevel: "info",
	}

	tests := []struct {
		name     string
		jsonMode bool
	}{
		{"text formatter", false},
		{"json formatter", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(cfg, false, false, tt.jsonMode)

			if tt.jsonMode {
				if _, ok := logger.Formatter.(*logrus.JSONFormatter); !ok {
					t.Error("Expected JSON formatter in JSON mode")
				}
			} else {
				if _, ok := logger.Formatter.(*logrus.TextFormatter); !ok {
					t.Error("Expected text formatter in normal mode")
				}
			}
		})
	}
}

func TestLogCommand(t *testing.T) {
	cfg := &config.Config{
		LogLevel: "debug",
	}

	var buf bytes.Buffer
	logger := NewLogger(cfg, true, false, false)
	logger.SetOutput(&buf)

	// Test successful command
	logger.LogCommand("apt install nginx", "apt", 0, 2*time.Second, "Package installed successfully")

	output := buf.String()
	if !strings.Contains(output, "apt install nginx") {
		t.Error("Expected log to contain command")
	}
	if !strings.Contains(output, "apt") {
		t.Error("Expected log to contain provider")
	}
	if !strings.Contains(output, "Command executed successfully") {
		t.Error("Expected success message")
	}

	// Reset buffer and test failed command
	buf.Reset()
	logger.LogCommand("apt install nonexistent", "apt", 1, 1*time.Second, "Package not found")

	output = buf.String()
	if !strings.Contains(output, "Command execution failed") {
		t.Error("Expected failure message")
	}
}

func TestLogAction(t *testing.T) {
	cfg := &config.Config{
		LogLevel: "debug",
	}

	var buf bytes.Buffer
	logger := NewLogger(cfg, true, false, false)
	logger.SetOutput(&buf)

	// Test successful action
	logger.LogAction("install", "nginx", "apt", true, 5*time.Second, nil)

	output := buf.String()
	if !strings.Contains(output, "install") {
		t.Error("Expected log to contain action")
	}
	if !strings.Contains(output, "nginx") {
		t.Error("Expected log to contain software")
	}
	if !strings.Contains(output, "apt") {
		t.Error("Expected log to contain provider")
	}
	if !strings.Contains(output, "Action completed") {
		t.Error("Expected completion message")
	}

	// Reset buffer and test failed action
	buf.Reset()
	logger.LogAction("install", "nginx", "apt", false, 2*time.Second, os.ErrNotExist)

	output = buf.String()
	if !strings.Contains(output, "Action failed") {
		t.Error("Expected failure message")
	}
	if !strings.Contains(output, "file does not exist") {
		t.Error("Expected error message")
	}
}

func TestLogProviderSelection(t *testing.T) {
	cfg := &config.Config{
		LogLevel: "debug",
	}

	var buf bytes.Buffer
	logger := NewLogger(cfg, true, false, false)
	logger.SetOutput(&buf)

	availableProviders := []string{"apt", "snap", "docker"}
	logger.LogProviderSelection("nginx", "apt", availableProviders)

	output := buf.String()
	if !strings.Contains(output, "nginx") {
		t.Error("Expected log to contain software")
	}
	if !strings.Contains(output, "apt") {
		t.Error("Expected log to contain selected provider")
	}
	if !strings.Contains(output, "Provider selected") {
		t.Error("Expected provider selection message")
	}
}

func TestLogConfigLoad(t *testing.T) {
	cfg := &config.Config{
		LogLevel: "debug",
	}

	var buf bytes.Buffer
	logger := NewLogger(cfg, true, false, false)
	logger.SetOutput(&buf)

	// Test successful config load
	logger.LogConfigLoad("/etc/sai/config.yaml", true, nil)

	output := buf.String()
	if !strings.Contains(output, "/etc/sai/config.yaml") {
		t.Error("Expected log to contain config path")
	}
	if !strings.Contains(output, "Configuration loaded") {
		t.Error("Expected success message")
	}

	// Reset buffer and test failed config load
	buf.Reset()
	logger.LogConfigLoad("/nonexistent/config.yaml", false, os.ErrNotExist)

	output = buf.String()
	if !strings.Contains(output, "Failed to load configuration") {
		t.Error("Expected failure message")
	}
}

func TestLogRepositoryOperation(t *testing.T) {
	cfg := &config.Config{
		LogLevel: "info",
	}

	var buf bytes.Buffer
	logger := NewLogger(cfg, false, false, false)
	logger.SetOutput(&buf)

	// Test successful repository operation
	logger.LogRepositoryOperation("clone", "https://github.com/sai-data/saidata.git", true, nil)

	output := buf.String()
	if !strings.Contains(output, "clone") {
		t.Error("Expected log to contain operation")
	}
	if !strings.Contains(output, "Repository operation completed") {
		t.Error("Expected success message")
	}

	// Reset buffer and test failed repository operation
	buf.Reset()
	logger.LogRepositoryOperation("pull", "https://github.com/sai-data/saidata.git", false, os.ErrPermission)

	output = buf.String()
	if !strings.Contains(output, "Repository operation failed") {
		t.Error("Expected failure message")
	}
}

func TestLogValidation(t *testing.T) {
	cfg := &config.Config{
		LogLevel: "debug",
	}

	var buf bytes.Buffer
	logger := NewLogger(cfg, true, false, false)
	logger.SetOutput(&buf)

	// Test successful validation
	logger.LogValidation("schema", "saidata.yaml", true, nil)

	output := buf.String()
	if !strings.Contains(output, "schema") {
		t.Error("Expected log to contain validation type")
	}
	if !strings.Contains(output, "Validation passed") {
		t.Error("Expected success message")
	}

	// Reset buffer and test failed validation
	buf.Reset()
	issues := []string{"missing required field", "invalid format"}
	logger.LogValidation("schema", "invalid.yaml", false, issues)

	output = buf.String()
	if !strings.Contains(output, "Validation failed") {
		t.Error("Expected failure message")
	}
	if !strings.Contains(output, "missing required field") {
		t.Error("Expected validation issues")
	}
}

func TestLogUserInteraction(t *testing.T) {
	cfg := &config.Config{
		LogLevel: "debug",
	}

	var buf bytes.Buffer
	logger := NewLogger(cfg, true, false, false)
	logger.SetOutput(&buf)

	logger.LogUserInteraction("confirmation", "Install nginx?", "yes")

	output := buf.String()
	if !strings.Contains(output, "confirmation") {
		t.Error("Expected log to contain interaction type")
	}
	if !strings.Contains(output, "Install nginx?") {
		t.Error("Expected log to contain prompt")
	}
	if !strings.Contains(output, "yes") {
		t.Error("Expected log to contain response")
	}
	if !strings.Contains(output, "User interaction") {
		t.Error("Expected interaction message")
	}
}

func TestLogPerformance(t *testing.T) {
	cfg := &config.Config{
		LogLevel: "debug",
	}

	var buf bytes.Buffer
	logger := NewLogger(cfg, true, false, false)
	logger.SetOutput(&buf)

	metadata := map[string]interface{}{
		"provider": "apt",
		"software": "nginx",
	}

	logger.LogPerformance("install", 3*time.Second, metadata)

	output := buf.String()
	if !strings.Contains(output, "install") {
		t.Error("Expected log to contain operation")
	}
	if !strings.Contains(output, "3s") {
		t.Error("Expected log to contain duration")
	}
	if !strings.Contains(output, "Performance metric") {
		t.Error("Expected performance message")
	}
}

func TestModeCheckers(t *testing.T) {
	cfg := &config.Config{
		LogLevel: "debug",
	}

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
			logger := NewLogger(cfg, tt.verbose, tt.quiet, tt.json)

			if logger.IsVerboseMode() != tt.verbose {
				t.Errorf("Expected verbose mode to be %v", tt.verbose)
			}

			if logger.IsQuietMode() != tt.quiet {
				t.Errorf("Expected quiet mode to be %v", tt.quiet)
			}

			if logger.IsJSONMode() != tt.json {
				t.Errorf("Expected JSON mode to be %v", tt.json)
			}

			if tt.verbose && !logger.IsDebugEnabled() {
				t.Error("Expected debug to be enabled in verbose mode")
			}
		})
	}
}

func TestCreateStructuredLog(t *testing.T) {
	cfg := &config.Config{
		LogLevel: "info",
	}

	logger := NewLogger(cfg, false, false, true)

	data := map[string]interface{}{
		"action":   "install",
		"software": "nginx",
		"success":  true,
	}

	result := logger.CreateStructuredLog("action_result", data)

	if !strings.Contains(result, "action_result") {
		t.Error("Expected structured log to contain log type")
	}

	if !strings.Contains(result, "install") {
		t.Error("Expected structured log to contain action")
	}

	if !strings.Contains(result, "nginx") {
		t.Error("Expected structured log to contain software")
	}

	if !strings.Contains(result, "true") {
		t.Error("Expected structured log to contain success status")
	}
}