package debug

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// DebugManager provides comprehensive debug logging and metrics collection
type DebugManager struct {
	enabled       bool
	logger        *logrus.Logger
	startTime     time.Time
	operations    []DebugOperation
	metrics       map[string]*PerformanceMetric
	mutex         sync.RWMutex
	outputFile    *os.File
	sessionID     string
}

// DebugOperation represents a single debug operation with timing and context
type DebugOperation struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Message     string                 `json:"message"`
	StartTime   time.Time              `json:"start_time"`
	Duration    time.Duration          `json:"duration"`
	Success     bool                   `json:"success"`
	Details     map[string]interface{} `json:"details"`
	Timestamp   time.Time              `json:"timestamp"`
	StackTrace  []string               `json:"stack_trace,omitempty"`
	MemoryUsage int64                  `json:"memory_usage"`
	Goroutines  int                    `json:"goroutines"`
}

// PerformanceMetric tracks performance data for operations
type PerformanceMetric struct {
	Operation    string        `json:"operation"`
	Count        int           `json:"count"`
	TotalTime    time.Duration `json:"total_time"`
	AverageTime  time.Duration `json:"average_time"`
	MinTime      time.Duration `json:"min_time"`
	MaxTime      time.Duration `json:"max_time"`
	LastExecuted time.Time     `json:"last_executed"`
	Errors       int           `json:"errors"`
}

// NewDebugManager creates a new debug manager
func NewDebugManager(enabled bool) *DebugManager {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	sessionID := fmt.Sprintf("sai-debug-%d", time.Now().Unix())

	dm := &DebugManager{
		enabled:    enabled,
		logger:     logger,
		startTime:  time.Now(),
		operations: make([]DebugOperation, 0),
		metrics:    make(map[string]*PerformanceMetric),
		sessionID:  sessionID,
	}

	if enabled {
		dm.setupDebugOutput()
		dm.LogSystemInfo()
	}

	return dm
}

// setupDebugOutput configures debug output to file and console
func (dm *DebugManager) setupDebugOutput() {
	// Create debug output file
	debugDir := os.TempDir()
	debugFile := fmt.Sprintf("%s/sai-debug-%s.log", debugDir, dm.sessionID)
	
	file, err := os.OpenFile(debugFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		dm.logger.Warnf("Failed to create debug file %s: %v", debugFile, err)
		return
	}

	dm.outputFile = file
	dm.logger.SetOutput(file)
	
	// Also log to stderr in debug mode
	dm.logger.AddHook(&ConsoleDebugHook{})
	
	dm.logger.Infof("Debug session started: %s", dm.sessionID)
	dm.logger.Infof("Debug output file: %s", debugFile)
}

// IsEnabled returns whether debug mode is enabled
func (dm *DebugManager) IsEnabled() bool {
	return dm.enabled
}

// LogSystemInfo logs comprehensive system information
func (dm *DebugManager) LogSystemInfo() {
	if !dm.enabled {
		return
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	systemInfo := map[string]interface{}{
		"go_version":      runtime.Version(),
		"go_os":           runtime.GOOS,
		"go_arch":         runtime.GOARCH,
		"num_cpu":         runtime.NumCPU(),
		"num_goroutines":  runtime.NumGoroutine(),
		"memory_alloc":    memStats.Alloc,
		"memory_total":    memStats.TotalAlloc,
		"memory_sys":      memStats.Sys,
		"gc_runs":         memStats.NumGC,
		"hostname":        getHostname(),
		"working_dir":     getWorkingDir(),
		"environment":     getRelevantEnvVars(),
	}

	dm.logger.WithFields(logrus.Fields{
		"type":        "system_info",
		"session_id":  dm.sessionID,
		"system_info": systemInfo,
	}).Info("System information collected")
}

// LogProviderDetection logs detailed provider detection information
func (dm *DebugManager) LogProviderDetection(allProviders []string, availableProviders []string, detectionResults map[string]bool, detectionTime time.Duration) {
	if !dm.enabled {
		return
	}

	operation := dm.startOperation("provider_detection", "Provider detection and availability check")
	
	details := map[string]interface{}{
		"all_providers":       allProviders,
		"available_providers": availableProviders,
		"detection_results":   detectionResults,
		"detection_time":      detectionTime.String(),
		"total_providers":     len(allProviders),
		"available_count":     len(availableProviders),
	}

	dm.finishOperation(operation, true, details)
	
	// Log individual provider detection results
	for provider, available := range detectionResults {
		dm.logger.WithFields(logrus.Fields{
			"type":      "provider_check",
			"provider":  provider,
			"available": available,
		}).Debug("Provider availability check")
	}

	dm.updateMetric("provider_detection", detectionTime, true)
}

// LogTemplateResolution logs template resolution with variables and results
func (dm *DebugManager) LogTemplateResolution(template string, variables map[string]interface{}, result string, success bool, resolutionTime time.Duration, err error) {
	if !dm.enabled {
		return
	}

	operation := dm.startOperation("template_resolution", "Template variable resolution")
	
	details := map[string]interface{}{
		"template":        template,
		"variables":       variables,
		"result":          result,
		"resolution_time": resolutionTime.String(),
		"variable_count":  len(variables),
	}

	if err != nil {
		details["error"] = err.Error()
	}

	dm.finishOperation(operation, success, details)
	dm.updateMetric("template_resolution", resolutionTime, success)
}

// LogCommandExecution logs detailed command execution information
func (dm *DebugManager) LogCommandExecution(command string, provider string, args []string, env []string, workingDir string, exitCode int, output string, stderr string, duration time.Duration) {
	if !dm.enabled {
		return
	}

	operation := dm.startOperation("command_execution", "System command execution")
	
	details := map[string]interface{}{
		"command":     command,
		"provider":    provider,
		"args":        args,
		"working_dir": workingDir,
		"exit_code":   exitCode,
		"duration":    duration.String(),
		"output_size": len(output),
		"stderr_size": len(stderr),
	}

	// Include environment variables (filtered for security)
	if len(env) > 0 {
		details["env_count"] = len(env)
		details["env_vars"] = filterSensitiveEnvVars(env)
	}

	// Include output in debug mode (truncated if too long)
	if len(output) > 0 {
		if len(output) > 1000 {
			details["output"] = output[:1000] + "... (truncated)"
		} else {
			details["output"] = output
		}
	}

	if len(stderr) > 0 {
		if len(stderr) > 1000 {
			details["stderr"] = stderr[:1000] + "... (truncated)"
		} else {
			details["stderr"] = stderr
		}
	}

	success := exitCode == 0
	dm.finishOperation(operation, success, details)
	dm.updateMetric("command_execution", duration, success)
}

// LogConfigurationLoading logs configuration loading and decision-making
func (dm *DebugManager) LogConfigurationLoading(configPath string, configFound bool, configData map[string]interface{}, envOverrides map[string]string, loadTime time.Duration, err error) {
	if !dm.enabled {
		return
	}

	operation := dm.startOperation("config_loading", "Configuration loading and processing")
	
	details := map[string]interface{}{
		"config_path":    configPath,
		"config_found":   configFound,
		"load_time":      loadTime.String(),
		"env_overrides":  envOverrides,
		"override_count": len(envOverrides),
	}

	if configFound && configData != nil {
		// Filter sensitive configuration data
		filteredConfig := filterSensitiveConfigData(configData)
		details["config_data"] = filteredConfig
		details["config_keys"] = getConfigKeys(configData)
	}

	if err != nil {
		details["error"] = err.Error()
	}

	success := err == nil
	dm.finishOperation(operation, success, details)
	dm.updateMetric("config_loading", loadTime, success)
}

// LogSaidataLoading logs saidata loading and processing
func (dm *DebugManager) LogSaidataLoading(software string, saidataPath string, osOverride string, mergeResults map[string]interface{}, loadTime time.Duration, success bool, err error) {
	if !dm.enabled {
		return
	}

	operation := dm.startOperation("saidata_loading", "Saidata loading and OS override merging")
	
	details := map[string]interface{}{
		"software":      software,
		"saidata_path":  saidataPath,
		"os_override":   osOverride,
		"load_time":     loadTime.String(),
		"merge_results": mergeResults,
	}

	if err != nil {
		details["error"] = err.Error()
	}

	dm.finishOperation(operation, success, details)
	dm.updateMetric("saidata_loading", loadTime, success)
}

// LogDecisionMaking logs decision-making processes
func (dm *DebugManager) LogDecisionMaking(decisionType string, context map[string]interface{}, options []string, selectedOption string, reasoning string, decisionTime time.Duration) {
	if !dm.enabled {
		return
	}

	operation := dm.startOperation("decision_making", fmt.Sprintf("Decision making: %s", decisionType))
	
	details := map[string]interface{}{
		"decision_type":   decisionType,
		"context":         context,
		"options":         options,
		"selected_option": selectedOption,
		"reasoning":       reasoning,
		"decision_time":   decisionTime.String(),
		"option_count":    len(options),
	}

	dm.finishOperation(operation, true, details)
	dm.updateMetric("decision_making", decisionTime, true)
}

// LogInternalState logs internal application state for troubleshooting
func (dm *DebugManager) LogInternalState(component string, state map[string]interface{}) {
	if !dm.enabled {
		return
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	dm.logger.WithFields(logrus.Fields{
		"type":           "internal_state",
		"component":      component,
		"state":          state,
		"memory_alloc":   memStats.Alloc,
		"num_goroutines": runtime.NumGoroutine(),
		"timestamp":      time.Now(),
	}).Debug("Internal state snapshot")
}

// ShowPerformanceMetrics displays collected performance metrics
func (dm *DebugManager) ShowPerformanceMetrics() {
	if !dm.enabled {
		return
	}

	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	totalDuration := time.Since(dm.startTime)
	
	fmt.Fprintf(os.Stderr, "\n=== SAI Debug Performance Metrics ===\n")
	fmt.Fprintf(os.Stderr, "Session ID: %s\n", dm.sessionID)
	fmt.Fprintf(os.Stderr, "Total Session Duration: %v\n", totalDuration)
	fmt.Fprintf(os.Stderr, "Total Operations: %d\n", len(dm.operations))
	
	if len(dm.metrics) > 0 {
		fmt.Fprintf(os.Stderr, "\nOperation Metrics:\n")
		for operation, metric := range dm.metrics {
			fmt.Fprintf(os.Stderr, "  %s:\n", operation)
			fmt.Fprintf(os.Stderr, "    Count: %d\n", metric.Count)
			fmt.Fprintf(os.Stderr, "    Total Time: %v\n", metric.TotalTime)
			fmt.Fprintf(os.Stderr, "    Average Time: %v\n", metric.AverageTime)
			fmt.Fprintf(os.Stderr, "    Min Time: %v\n", metric.MinTime)
			fmt.Fprintf(os.Stderr, "    Max Time: %v\n", metric.MaxTime)
			fmt.Fprintf(os.Stderr, "    Errors: %d\n", metric.Errors)
			fmt.Fprintf(os.Stderr, "    Last Executed: %v\n", metric.LastExecuted.Format(time.RFC3339))
		}
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	fmt.Fprintf(os.Stderr, "\nMemory Usage:\n")
	fmt.Fprintf(os.Stderr, "  Current Alloc: %d bytes\n", memStats.Alloc)
	fmt.Fprintf(os.Stderr, "  Total Alloc: %d bytes\n", memStats.TotalAlloc)
	fmt.Fprintf(os.Stderr, "  System Memory: %d bytes\n", memStats.Sys)
	fmt.Fprintf(os.Stderr, "  GC Runs: %d\n", memStats.NumGC)
	fmt.Fprintf(os.Stderr, "  Goroutines: %d\n", runtime.NumGoroutine())
	
	if dm.outputFile != nil {
		fmt.Fprintf(os.Stderr, "\nDebug log file: %s\n", dm.outputFile.Name())
	}
	
	fmt.Fprintf(os.Stderr, "=====================================\n\n")
}

// startOperation starts tracking a debug operation
func (dm *DebugManager) startOperation(operationType, message string) *DebugOperation {
	if !dm.enabled {
		return nil
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	operation := &DebugOperation{
		ID:          fmt.Sprintf("%s-%d", operationType, time.Now().UnixNano()),
		Type:        operationType,
		Message:     message,
		StartTime:   time.Now(),
		Timestamp:   time.Now(),
		Details:     make(map[string]interface{}),
		MemoryUsage: int64(memStats.Alloc),
		Goroutines:  runtime.NumGoroutine(),
	}

	// Capture stack trace for debugging
	if dm.shouldCaptureStackTrace(operationType) {
		operation.StackTrace = captureStackTrace()
	}

	dm.logger.WithFields(logrus.Fields{
		"type":         "operation_start",
		"operation_id": operation.ID,
		"operation":    operationType,
		"message":      message,
	}).Debug("Operation started")

	return operation
}

// finishOperation completes tracking a debug operation
func (dm *DebugManager) finishOperation(operation *DebugOperation, success bool, details map[string]interface{}) {
	if !dm.enabled || operation == nil {
		return
	}

	operation.Duration = time.Since(operation.StartTime)
	operation.Success = success
	
	for key, value := range details {
		operation.Details[key] = value
	}

	dm.mutex.Lock()
	dm.operations = append(dm.operations, *operation)
	dm.mutex.Unlock()

	dm.logger.WithFields(logrus.Fields{
		"type":         "operation_complete",
		"operation_id": operation.ID,
		"operation":    operation.Type,
		"duration":     operation.Duration.String(),
		"success":      success,
		"details":      details,
	}).Debug("Operation completed")
}

// updateMetric updates performance metrics for an operation
func (dm *DebugManager) updateMetric(operation string, duration time.Duration, success bool) {
	if !dm.enabled {
		return
	}

	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	metric, exists := dm.metrics[operation]
	if !exists {
		metric = &PerformanceMetric{
			Operation: operation,
			MinTime:   duration,
			MaxTime:   duration,
		}
		dm.metrics[operation] = metric
	}

	metric.Count++
	metric.TotalTime += duration
	metric.AverageTime = metric.TotalTime / time.Duration(metric.Count)
	metric.LastExecuted = time.Now()

	if duration < metric.MinTime {
		metric.MinTime = duration
	}
	if duration > metric.MaxTime {
		metric.MaxTime = duration
	}

	if !success {
		metric.Errors++
	}
}

// shouldCaptureStackTrace determines if stack trace should be captured for operation type
func (dm *DebugManager) shouldCaptureStackTrace(operationType string) bool {
	// Capture stack traces for error-prone operations
	errorProneOps := []string{"command_execution", "template_resolution", "config_loading"}
	for _, op := range errorProneOps {
		if op == operationType {
			return true
		}
	}
	return false
}

// Close closes the debug manager and cleans up resources
func (dm *DebugManager) Close() error {
	if dm.outputFile != nil {
		dm.logger.Info("Debug session ended")
		return dm.outputFile.Close()
	}
	return nil
}

// Helper functions

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func getWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}

func getRelevantEnvVars() map[string]string {
	relevantVars := []string{
		"SAI_CONFIG", "SAI_PROVIDER", "SAI_DEBUG", "SAI_CACHE_DIR",
		"SAI_SAIDATA_REPOSITORY", "SAI_LOG_LEVEL", "SAI_TIMEOUT",
		"PATH", "HOME", "USER", "SHELL", "TERM",
	}
	
	envVars := make(map[string]string)
	for _, varName := range relevantVars {
		if value := os.Getenv(varName); value != "" {
			envVars[varName] = value
		}
	}
	return envVars
}

func filterSensitiveEnvVars(env []string) []string {
	filtered := make([]string, 0)
	sensitivePatterns := []string{"PASSWORD", "SECRET", "TOKEN", "KEY", "CREDENTIAL"}
	
	for _, envVar := range env {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		varName := strings.ToUpper(parts[0])
		isSensitive := false
		
		for _, pattern := range sensitivePatterns {
			if strings.Contains(varName, pattern) {
				isSensitive = true
				break
			}
		}
		
		if isSensitive {
			filtered = append(filtered, parts[0]+"=***REDACTED***")
		} else {
			filtered = append(filtered, envVar)
		}
	}
	
	return filtered
}

func filterSensitiveConfigData(configData map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{})
	sensitiveKeys := []string{"password", "secret", "token", "key", "credential"}
	
	for key, value := range configData {
		lowerKey := strings.ToLower(key)
		isSensitive := false
		
		for _, sensitiveKey := range sensitiveKeys {
			if strings.Contains(lowerKey, sensitiveKey) {
				isSensitive = true
				break
			}
		}
		
		if isSensitive {
			filtered[key] = "***REDACTED***"
		} else {
			filtered[key] = value
		}
	}
	
	return filtered
}

func getConfigKeys(configData map[string]interface{}) []string {
	keys := make([]string, 0, len(configData))
	for key := range configData {
		keys = append(keys, key)
	}
	return keys
}

func captureStackTrace() []string {
	const maxDepth = 10
	stackTrace := make([]string, 0, maxDepth)
	
	for i := 2; i < maxDepth+2; i++ { // Skip captureStackTrace and startOperation
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		
		// Simplify file path
		if idx := strings.LastIndex(file, "/"); idx >= 0 {
			file = file[idx+1:]
		}
		
		stackTrace = append(stackTrace, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
	}
	
	return stackTrace
}

// ConsoleDebugHook sends debug logs to stderr in addition to file
type ConsoleDebugHook struct{}

func (hook *ConsoleDebugHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel}
}

func (hook *ConsoleDebugHook) Fire(entry *logrus.Entry) error {
	// Only show important debug messages on console to avoid spam
	if entry.Level <= logrus.InfoLevel {
		line, err := entry.String()
		if err != nil {
			return err
		}
		fmt.Fprint(os.Stderr, line)
	}
	return nil
}