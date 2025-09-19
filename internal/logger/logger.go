package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"sai/internal/config"
)

// Logger provides structured logging with different output modes
type Logger struct {
	*logrus.Logger
	config      *config.Config
	verboseMode bool
	quietMode   bool
	jsonMode    bool
}

// NewLogger creates a new logger with the given configuration
func NewLogger(cfg *config.Config, verbose, quiet, jsonOutput bool) *Logger {
	logger := logrus.New()
	
	l := &Logger{
		Logger:      logger,
		config:      cfg,
		verboseMode: verbose,
		quietMode:   quiet,
		jsonMode:    jsonOutput,
	}

	l.setupLogger()
	return l
}

// setupLogger configures the logger based on the configuration and modes
func (l *Logger) setupLogger() {
	// Set log level based on configuration and verbose mode
	level := l.config.LogLevel
	if l.verboseMode {
		level = "debug"
	} else if l.quietMode {
		level = "error"
	}

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	l.SetLevel(logLevel)

	// Set output destination
	if l.quietMode {
		// In quiet mode, only log errors and above to stderr
		l.SetOutput(os.Stderr)
	} else {
		l.SetOutput(os.Stdout)
	}

	// Set formatter based on JSON mode
	if l.jsonMode {
		l.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	} else {
		l.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			DisableColors:   !l.supportsColor(),
		})
	}
}

// LogCommand logs a command execution with structured data
func (l *Logger) LogCommand(command, provider string, exitCode int, duration time.Duration, output string) {
	fields := logrus.Fields{
		"command":   command,
		"provider":  provider,
		"exit_code": exitCode,
		"duration":  duration.String(),
	}

	if l.verboseMode && output != "" {
		fields["output"] = output
	}

	if exitCode == 0 {
		l.WithFields(fields).Info("Command executed successfully")
	} else {
		l.WithFields(fields).Error("Command execution failed")
	}
}

// LogAction logs an action execution with structured data
func (l *Logger) LogAction(action, software, provider string, success bool, duration time.Duration, err error) {
	fields := logrus.Fields{
		"action":   action,
		"software": software,
		"provider": provider,
		"success":  success,
		"duration": duration.String(),
	}

	if err != nil {
		fields["error"] = err.Error()
		l.WithFields(fields).Error("Action failed")
	} else {
		l.WithFields(fields).Info("Action completed")
	}
}

// LogProviderSelection logs provider selection events
func (l *Logger) LogProviderSelection(software string, selectedProvider string, availableProviders []string) {
	fields := logrus.Fields{
		"software":            software,
		"selected_provider":   selectedProvider,
		"available_providers": availableProviders,
	}

	l.WithFields(fields).Debug("Provider selected")
}

// LogConfigLoad logs configuration loading events
func (l *Logger) LogConfigLoad(configPath string, success bool, err error) {
	fields := logrus.Fields{
		"config_path": configPath,
		"success":     success,
	}

	if err != nil {
		fields["error"] = err.Error()
		l.WithFields(fields).Warn("Failed to load configuration")
	} else {
		l.WithFields(fields).Debug("Configuration loaded")
	}
}

// LogRepositoryOperation logs repository operations
func (l *Logger) LogRepositoryOperation(operation, repository string, success bool, err error) {
	fields := logrus.Fields{
		"operation":  operation,
		"repository": repository,
		"success":    success,
	}

	if err != nil {
		fields["error"] = err.Error()
		l.WithFields(fields).Error("Repository operation failed")
	} else {
		l.WithFields(fields).Info("Repository operation completed")
	}
}

// LogValidation logs validation events
func (l *Logger) LogValidation(validationType string, target string, valid bool, issues []string) {
	fields := logrus.Fields{
		"validation_type": validationType,
		"target":          target,
		"valid":           valid,
	}

	if len(issues) > 0 {
		fields["issues"] = issues
	}

	if valid {
		l.WithFields(fields).Debug("Validation passed")
	} else {
		l.WithFields(fields).Warn("Validation failed")
	}
}

// LogUserInteraction logs user interaction events
func (l *Logger) LogUserInteraction(interactionType string, prompt string, response string) {
	fields := logrus.Fields{
		"interaction_type": interactionType,
		"prompt":           prompt,
		"response":         response,
	}

	l.WithFields(fields).Debug("User interaction")
}

// LogPerformance logs performance metrics
func (l *Logger) LogPerformance(operation string, duration time.Duration, metadata map[string]interface{}) {
	fields := logrus.Fields{
		"operation": operation,
		"duration":  duration.String(),
	}

	for key, value := range metadata {
		fields[key] = value
	}

	l.WithFields(fields).Debug("Performance metric")
}

// SetOutput sets the logger output destination
func (l *Logger) SetOutput(output io.Writer) {
	l.Logger.SetOutput(output)
}

// SetLevel sets the logger level
func (l *Logger) SetLevel(level logrus.Level) {
	l.Logger.SetLevel(level)
}

// WithField creates a new logger entry with a field
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

// WithFields creates a new logger entry with multiple fields
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

// IsDebugEnabled returns whether debug logging is enabled
func (l *Logger) IsDebugEnabled() bool {
	return l.GetLevel() >= logrus.DebugLevel
}

// IsVerboseMode returns whether verbose mode is enabled
func (l *Logger) IsVerboseMode() bool {
	return l.verboseMode
}

// IsQuietMode returns whether quiet mode is enabled
func (l *Logger) IsQuietMode() bool {
	return l.quietMode
}

// IsJSONMode returns whether JSON output mode is enabled
func (l *Logger) IsJSONMode() bool {
	return l.jsonMode
}

// supportsColor checks if the terminal supports colors
func (l *Logger) supportsColor() bool {
	// Don't use colors in JSON mode
	if l.jsonMode {
		return false
	}

	// Check if NO_COLOR environment variable is set
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if TERM environment variable indicates color support
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return false
	}

	return true
}

// CreateStructuredLog creates a structured log entry for JSON output
func (l *Logger) CreateStructuredLog(logType string, data map[string]interface{}) string {
	logEntry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"type":      logType,
	}

	for key, value := range data {
		logEntry[key] = value
	}

	jsonData, err := json.MarshalIndent(logEntry, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal log entry: %s"}`, err.Error())
	}

	return string(jsonData)
}