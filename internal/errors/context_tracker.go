package errors

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ErrorContext provides detailed context for debugging and troubleshooting
type ErrorContext struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	Action        string                 `json:"action"`
	Software      string                 `json:"software"`
	Provider      string                 `json:"provider"`
	Error         error                  `json:"error"`
	ErrorType     ErrorType              `json:"error_type"`
	StackTrace    []StackFrame           `json:"stack_trace"`
	SystemInfo    *SystemInfo            `json:"system_info"`
	ExecutionPath []ExecutionStep        `json:"execution_path"`
	Variables     map[string]interface{} `json:"variables"`
	Commands      []CommandExecution     `json:"commands"`
	Duration      time.Duration          `json:"duration"`
	Recoverable   bool                   `json:"recoverable"`
	RecoveryHints []string               `json:"recovery_hints"`
}

// StackFrame represents a single frame in the stack trace
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Package  string `json:"package"`
}

// SystemInfo captures system information at the time of error
type SystemInfo struct {
	OS           string            `json:"os"`
	Architecture string            `json:"architecture"`
	Version      string            `json:"version"`
	Hostname     string            `json:"hostname"`
	User         string            `json:"user"`
	WorkingDir   string            `json:"working_dir"`
	Environment  map[string]string `json:"environment"`
	Timestamp    time.Time         `json:"timestamp"`
}

// ExecutionStep represents a step in the execution path
type ExecutionStep struct {
	Step        string                 `json:"step"`
	Timestamp   time.Time              `json:"timestamp"`
	Duration    time.Duration          `json:"duration"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

// CommandExecution represents a command execution attempt
type CommandExecution struct {
	Command     string            `json:"command"`
	Arguments   []string          `json:"arguments"`
	Environment map[string]string `json:"environment"`
	WorkingDir  string            `json:"working_dir"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     time.Time         `json:"end_time"`
	Duration    time.Duration     `json:"duration"`
	ExitCode    int               `json:"exit_code"`
	Stdout      string            `json:"stdout"`
	Stderr      string            `json:"stderr"`
	Success     bool              `json:"success"`
	Error       string            `json:"error,omitempty"`
}

// ErrorContextTracker tracks error contexts for debugging and troubleshooting
type ErrorContextTracker struct {
	contexts map[string]*ErrorContext
	mutex    sync.RWMutex
	maxSize  int
}

// NewErrorContextTracker creates a new error context tracker
func NewErrorContextTracker(maxSize int) *ErrorContextTracker {
	if maxSize <= 0 {
		maxSize = 1000 // Default maximum contexts to keep
	}
	
	return &ErrorContextTracker{
		contexts: make(map[string]*ErrorContext),
		maxSize:  maxSize,
	}
}

// TrackError tracks an error with full context
func (ect *ErrorContextTracker) TrackError(ctx context.Context, action, software, provider string, err error) *ErrorContext {
	errorCtx := &ErrorContext{
		ID:            generateErrorID(),
		Timestamp:     time.Now(),
		Action:        action,
		Software:      software,
		Provider:      provider,
		Error:         err,
		ErrorType:     GetErrorType(err),
		StackTrace:    captureStackTrace(),
		SystemInfo:    captureSystemInfo(),
		ExecutionPath: extractExecutionPath(ctx),
		Variables:     extractVariables(ctx),
		Commands:      extractCommands(ctx),
		Duration:      extractDuration(ctx),
		Recoverable:   IsRecoverable(err),
		RecoveryHints: generateRecoveryHints(err),
	}

	ect.mutex.Lock()
	defer ect.mutex.Unlock()

	// Add to contexts map
	ect.contexts[errorCtx.ID] = errorCtx

	// Cleanup old contexts if we exceed max size
	if len(ect.contexts) > ect.maxSize {
		ect.cleanupOldContexts()
	}

	return errorCtx
}

// GetErrorContext retrieves an error context by ID
func (ect *ErrorContextTracker) GetErrorContext(id string) (*ErrorContext, bool) {
	ect.mutex.RLock()
	defer ect.mutex.RUnlock()
	
	ctx, exists := ect.contexts[id]
	return ctx, exists
}

// GetRecentErrors returns the most recent error contexts
func (ect *ErrorContextTracker) GetRecentErrors(limit int) []*ErrorContext {
	ect.mutex.RLock()
	defer ect.mutex.RUnlock()
	
	var contexts []*ErrorContext
	for _, ctx := range ect.contexts {
		contexts = append(contexts, ctx)
	}
	
	// Sort by timestamp (most recent first)
	for i := 0; i < len(contexts)-1; i++ {
		for j := i + 1; j < len(contexts); j++ {
			if contexts[i].Timestamp.Before(contexts[j].Timestamp) {
				contexts[i], contexts[j] = contexts[j], contexts[i]
			}
		}
	}
	
	if limit > 0 && len(contexts) > limit {
		contexts = contexts[:limit]
	}
	
	return contexts
}

// GetErrorsByType returns error contexts filtered by error type
func (ect *ErrorContextTracker) GetErrorsByType(errorType ErrorType) []*ErrorContext {
	ect.mutex.RLock()
	defer ect.mutex.RUnlock()
	
	var filtered []*ErrorContext
	for _, ctx := range ect.contexts {
		if ctx.ErrorType == errorType {
			filtered = append(filtered, ctx)
		}
	}
	
	return filtered
}

// GetErrorsByAction returns error contexts filtered by action
func (ect *ErrorContextTracker) GetErrorsByAction(action string) []*ErrorContext {
	ect.mutex.RLock()
	defer ect.mutex.RUnlock()
	
	var filtered []*ErrorContext
	for _, ctx := range ect.contexts {
		if ctx.Action == action {
			filtered = append(filtered, ctx)
		}
	}
	
	return filtered
}

// GetErrorStats returns statistics about tracked errors
func (ect *ErrorContextTracker) GetErrorStats() *ErrorStats {
	ect.mutex.RLock()
	defer ect.mutex.RUnlock()
	
	stats := &ErrorStats{
		TotalErrors:    len(ect.contexts),
		ErrorsByType:   make(map[ErrorType]int),
		ErrorsByAction: make(map[string]int),
		RecoverableErrors: 0,
		RecentErrors:   0,
	}
	
	now := time.Now()
	recentThreshold := now.Add(-1 * time.Hour) // Last hour
	
	for _, ctx := range ect.contexts {
		// Count by type
		stats.ErrorsByType[ctx.ErrorType]++
		
		// Count by action
		stats.ErrorsByAction[ctx.Action]++
		
		// Count recoverable errors
		if ctx.Recoverable {
			stats.RecoverableErrors++
		}
		
		// Count recent errors
		if ctx.Timestamp.After(recentThreshold) {
			stats.RecentErrors++
		}
	}
	
	return stats
}

// ClearErrors clears all tracked error contexts
func (ect *ErrorContextTracker) ClearErrors() {
	ect.mutex.Lock()
	defer ect.mutex.Unlock()
	
	ect.contexts = make(map[string]*ErrorContext)
}

// cleanupOldContexts removes the oldest contexts to maintain size limit
func (ect *ErrorContextTracker) cleanupOldContexts() {
	// Find oldest contexts
	var oldestContexts []*ErrorContext
	for _, ctx := range ect.contexts {
		oldestContexts = append(oldestContexts, ctx)
	}
	
	// Sort by timestamp (oldest first)
	for i := 0; i < len(oldestContexts)-1; i++ {
		for j := i + 1; j < len(oldestContexts); j++ {
			if oldestContexts[i].Timestamp.After(oldestContexts[j].Timestamp) {
				oldestContexts[i], oldestContexts[j] = oldestContexts[j], oldestContexts[i]
			}
		}
	}
	
	// Remove oldest contexts until we're under the limit
	removeCount := len(ect.contexts) - ect.maxSize + 100 // Remove extra to avoid frequent cleanup
	for i := 0; i < removeCount && i < len(oldestContexts); i++ {
		delete(ect.contexts, oldestContexts[i].ID)
	}
}

// ErrorStats provides statistics about tracked errors
type ErrorStats struct {
	TotalErrors       int                `json:"total_errors"`
	ErrorsByType      map[ErrorType]int  `json:"errors_by_type"`
	ErrorsByAction    map[string]int     `json:"errors_by_action"`
	RecoverableErrors int                `json:"recoverable_errors"`
	RecentErrors      int                `json:"recent_errors"`
}

// Helper functions for context extraction

func generateErrorID() string {
	return fmt.Sprintf("err_%d_%d", time.Now().UnixNano(), runtime.NumGoroutine())
}

func captureStackTrace() []StackFrame {
	var frames []StackFrame
	
	// Skip the first few frames (this function and error handling functions)
	skip := 3
	for i := skip; i < skip+10; i++ { // Capture up to 10 frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		
		frame := StackFrame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
			Package:  extractPackageName(fn.Name()),
		}
		
		frames = append(frames, frame)
	}
	
	return frames
}

func extractPackageName(funcName string) string {
	// Extract package name from function name
	// e.g., "sai/internal/action.(*ActionManager).ExecuteAction" -> "sai/internal/action"
	lastSlash := -1
	for i := len(funcName) - 1; i >= 0; i-- {
		if funcName[i] == '/' {
			lastSlash = i
			break
		}
	}
	
	if lastSlash == -1 {
		return ""
	}
	
	lastDot := -1
	for i := len(funcName) - 1; i > lastSlash; i-- {
		if funcName[i] == '.' {
			lastDot = i
			break
		}
	}
	
	if lastDot == -1 {
		return funcName[:lastSlash]
	}
	
	return funcName[:lastDot]
}

func captureSystemInfo() *SystemInfo {
	return &SystemInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		Version:      runtime.Version(),
		Hostname:     getHostname(),
		User:         getCurrentUser(),
		WorkingDir:   getCurrentDir(),
		Environment:  getRelevantEnvVars(),
		Timestamp:    time.Now(),
	}
}

func extractExecutionPath(ctx context.Context) []ExecutionStep {
	// Extract execution path from context if available
	if ctx == nil {
		return []ExecutionStep{}
	}
	
	// This would be populated by the execution framework
	if steps, ok := ctx.Value(ContextKeyExecutionPath).([]ExecutionStep); ok {
		return steps
	}
	
	return []ExecutionStep{}
}

func extractVariables(ctx context.Context) map[string]interface{} {
	variables := make(map[string]interface{})
	
	if ctx == nil {
		return variables
	}
	
	// Extract variables from context if available
	if vars, ok := ctx.Value(ContextKeyVariables).(map[string]interface{}); ok {
		for k, v := range vars {
			variables[k] = v
		}
	}
	
	return variables
}

func extractCommands(ctx context.Context) []CommandExecution {
	if ctx == nil {
		return []CommandExecution{}
	}
	
	// Extract command executions from context if available
	if commands, ok := ctx.Value(ContextKeyCommands).([]CommandExecution); ok {
		return commands
	}
	
	return []CommandExecution{}
}

func extractDuration(ctx context.Context) time.Duration {
	if ctx == nil {
		return 0
	}
	
	// Extract duration from context if available
	if startTime, ok := ctx.Value(ContextKeyStartTime).(time.Time); ok {
		return time.Since(startTime)
	}
	
	return 0
}

func generateRecoveryHints(err error) []string {
	var hints []string
	
	if saiErr, ok := err.(*SAIError); ok {
		// Add existing suggestions
		hints = append(hints, saiErr.Suggestions...)
		
		// Add type-specific recovery hints
		switch saiErr.Type {
		case ErrorTypeProviderNotFound:
			hints = append(hints, "Install the required package manager")
			hints = append(hints, "Check if provider is available on your system")
		case ErrorTypeNetworkTimeout:
			hints = append(hints, "Check internet connectivity")
			hints = append(hints, "Try again later")
			hints = append(hints, "Use offline mode if available")
		case ErrorTypeResourceMissing:
			hints = append(hints, "Create the missing resource manually")
			hints = append(hints, "Check file permissions")
		case ErrorTypeActionTimeout:
			hints = append(hints, "Increase timeout value")
			hints = append(hints, "Check system performance")
		}
	}
	
	return hints
}

// System information helper functions (simplified implementations)

func getHostname() string {
	// Implementation would get actual hostname
	return "localhost"
}

func getCurrentUser() string {
	// Implementation would get actual current user
	return "user"
}

func getCurrentDir() string {
	// Implementation would get actual working directory
	return "/current/dir"
}

func getRelevantEnvVars() map[string]string {
	// Implementation would get relevant environment variables
	return map[string]string{
		"PATH": "/usr/bin:/bin",
		"HOME": "/home/user",
	}
}

// Context keys for storing execution information
type contextKey string

const (
	ContextKeyExecutionPath contextKey = "execution_path"
	ContextKeyVariables     contextKey = "variables"
	ContextKeyCommands      contextKey = "commands"
	ContextKeyStartTime     contextKey = "start_time"
)

// WithExecutionStep adds an execution step to the context
func WithExecutionStep(ctx context.Context, step ExecutionStep) context.Context {
	steps := extractExecutionPath(ctx)
	steps = append(steps, step)
	return context.WithValue(ctx, ContextKeyExecutionPath, steps)
}

// WithVariable adds a variable to the context
func WithVariable(ctx context.Context, key string, value interface{}) context.Context {
	variables := extractVariables(ctx)
	variables[key] = value
	return context.WithValue(ctx, ContextKeyVariables, variables)
}

// WithCommand adds a command execution to the context
func WithCommand(ctx context.Context, cmd CommandExecution) context.Context {
	commands := extractCommands(ctx)
	commands = append(commands, cmd)
	return context.WithValue(ctx, ContextKeyCommands, commands)
}

// WithStartTime adds a start time to the context
func WithStartTime(ctx context.Context, startTime time.Time) context.Context {
	return context.WithValue(ctx, ContextKeyStartTime, startTime)
}