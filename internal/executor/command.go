package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"sai/internal/debug"
	"sai/internal/interfaces"
	"sai/internal/types"
)

// CommandExecutor implements command execution with safety features
type CommandExecutor struct {
	logger    interfaces.Logger
	validator interfaces.ResourceValidator
	dryRun    bool
	timeout   time.Duration
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor(logger interfaces.Logger, validator interfaces.ResourceValidator) *CommandExecutor {
	return &CommandExecutor{
		logger:    logger,
		validator: validator,
		timeout:   300 * time.Second, // Default 5 minutes
	}
}

// ExecuteCommand executes a single command with proper error handling
func (ce *CommandExecutor) ExecuteCommand(ctx context.Context, command string, options interfaces.CommandOptions) (*interfaces.CommandResult, error) {
	startTime := time.Now()
	
	// Log command execution
	ce.logger.Debug("Executing command", interfaces.LogField{Key: "command", Value: command})
	
	// Validate command before execution
	if err := ce.validateCommand(command); err != nil {
		return &interfaces.CommandResult{
			Command:  command,
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}
	
	// Handle dry-run mode
	if ce.dryRun || options.Timeout == 0 {
		ce.logger.Info("DRY RUN: Would execute command", interfaces.LogField{Key: "command", Value: command})
		return &interfaces.CommandResult{
			Command:  command,
			Output:   fmt.Sprintf("DRY RUN: %s", command),
			ExitCode: 0,
			Duration: time.Since(startTime),
		}, nil
	}
	
	// Set up command context with timeout
	timeout := ce.timeout
	if options.Timeout > 0 {
		timeout = options.Timeout
	}
	
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Parse command and arguments
	parts := strings.Fields(command)
	if len(parts) == 0 {
		err := fmt.Errorf("empty command")
		return &interfaces.CommandResult{
			Command:  command,
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}
	
	// Create command
	cmd := exec.CommandContext(cmdCtx, parts[0], parts[1:]...)
	
	// Set working directory if specified
	if options.WorkDir != "" {
		cmd.Dir = options.WorkDir
	}
	
	// Set environment variables
	if len(options.Env) > 0 {
		env := os.Environ()
		for key, value := range options.Env {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}
	
	// Set input if provided
	if options.Input != "" {
		cmd.Stdin = strings.NewReader(options.Input)
	}
	
	// Final validation before execution
	if err := ce.validateCommandBeforeExecution(cmd); err != nil {
		return &interfaces.CommandResult{
			Command:  command,
			Error:    fmt.Errorf("pre-execution validation failed: %w", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}
	
	// Execute command and capture output
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)
	
	// Get exit code
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			} else {
				exitCode = 1
			}
		} else {
			exitCode = 1
		}
	}
	
	result := &interfaces.CommandResult{
		Command:  command,
		Output:   string(output),
		Error:    err,
		ExitCode: exitCode,
		Duration: duration,
	}
	
	// Log command execution with debug system
	var stderr string
	if err != nil {
		stderr = err.Error()
	}
	
	env := make([]string, 0)
	if cmd.Env != nil {
		env = cmd.Env
	}
	
	debug.LogCommandExecutionGlobal(
		command,
		"", // provider will be set by caller
		parts[1:], // args
		env,
		cmd.Dir, // working directory
		exitCode,
		string(output), // stdout
		stderr,
		duration,
	)
	
	// Log result with comprehensive information
	if err != nil {
		ce.logger.Error("Command execution failed", err, 
			interfaces.LogField{Key: "command", Value: command},
			interfaces.LogField{Key: "exit_code", Value: exitCode},
			interfaces.LogField{Key: "duration", Value: duration},
			interfaces.LogField{Key: "output", Value: string(output)},
			interfaces.LogField{Key: "working_directory", Value: cmd.Dir},
		)
		
		// Log additional context for debugging
		if len(parts) > 0 {
			ce.logger.Debug("Command execution details",
				interfaces.LogField{Key: "executable", Value: parts[0]},
				interfaces.LogField{Key: "arguments", Value: strings.Join(parts[1:], " ")},
				interfaces.LogField{Key: "path_lookup", Value: cmd.Path},
			)
		}
	} else {
		ce.logger.Debug("Command executed successfully",
			interfaces.LogField{Key: "command", Value: command},
			interfaces.LogField{Key: "exit_code", Value: exitCode},
			interfaces.LogField{Key: "duration", Value: duration},
		)
		
		// Log output in verbose mode
		if len(output) > 0 {
			ce.logger.Debug("Command output",
				interfaces.LogField{Key: "command", Value: command},
				interfaces.LogField{Key: "output", Value: string(output)},
			)
		}
	}
	
	return result, nil
}

// ExecuteWithRetry executes a command with retry logic
func (ce *CommandExecutor) ExecuteWithRetry(ctx context.Context, command string, options interfaces.CommandOptions, retryConfig *types.RetryConfig) (*interfaces.CommandResult, error) {
	if retryConfig == nil {
		return ce.ExecuteCommand(ctx, command, options)
	}
	
	var lastResult *interfaces.CommandResult
	var lastErr error
	
	attempts := retryConfig.Attempts
	if attempts <= 0 {
		attempts = 1
	}
	
	for i := 0; i < attempts; i++ {
		if i > 0 {
			// Calculate delay with backoff
			delay := time.Duration(retryConfig.Delay) * time.Second
			if retryConfig.Backoff == "exponential" {
				delay = delay * time.Duration(1<<uint(i-1))
			}
			
			ce.logger.Debug("Retrying command after delay",
				interfaces.LogField{Key: "command", Value: command},
				interfaces.LogField{Key: "attempt", Value: i + 1},
				interfaces.LogField{Key: "delay", Value: delay},
			)
			
			select {
			case <-ctx.Done():
				return lastResult, ctx.Err()
			case <-time.After(delay):
			}
		}
		
		result, err := ce.ExecuteCommand(ctx, command, options)
		lastResult = result
		lastErr = err
		
		// Success - return immediately
		if err == nil && result.ExitCode == 0 {
			return result, nil
		}
		
		// Log retry attempt
		if i < attempts-1 {
			ce.logger.Warn("Command failed, will retry",
				interfaces.LogField{Key: "command", Value: command},
				interfaces.LogField{Key: "attempt", Value: i + 1},
				interfaces.LogField{Key: "error", Value: err},
			)
		}
	}
	
	ce.logger.Error("Command failed after all retry attempts",
		lastErr,
		interfaces.LogField{Key: "command", Value: command},
		interfaces.LogField{Key: "attempts", Value: attempts},
	)
	
	return lastResult, lastErr
}

// ValidateCommand validates that a command can be executed
func (ce *CommandExecutor) ValidateCommand(command string) error {
	return ce.validateCommand(command)
}

// validateCommand performs safety validation on a command
func (ce *CommandExecutor) validateCommand(command string) error {
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("command cannot be empty")
	}
	
	// Parse command to get executable
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("invalid command format")
	}
	
	executable := parts[0]
	
	// Check if executable exists and is executable
	if !ce.isExecutableAvailable(executable) {
		return fmt.Errorf("executable not found or not executable: %s", executable)
	}
	
	// Additional safety checks
	if err := ce.performSafetyChecks(command); err != nil {
		return fmt.Errorf("safety check failed: %w", err)
	}
	
	return nil
}

// isExecutableAvailable checks if an executable is available in PATH or as absolute path
func (ce *CommandExecutor) isExecutableAvailable(executable string) bool {
	// If it's an absolute path, check if it exists and is executable
	if strings.HasPrefix(executable, "/") {
		info, err := os.Stat(executable)
		if err != nil {
			return false
		}
		return info.Mode()&0111 != 0 // Check if executable bit is set
	}
	
	// Check if executable is in PATH
	_, err := exec.LookPath(executable)
	return err == nil
}

// performSafetyChecks performs additional safety validation
func (ce *CommandExecutor) performSafetyChecks(command string) error {
	// Check for potentially dangerous commands
	dangerousPatterns := []string{
		"rm -rf /",
		"dd if=",
		"mkfs",
		"fdisk",
		"format",
		"> /dev/",
	}
	
	lowerCommand := strings.ToLower(command)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerCommand, pattern) {
			ce.logger.Warn("Potentially dangerous command detected",
				interfaces.LogField{Key: "command", Value: command},
				interfaces.LogField{Key: "pattern", Value: pattern},
			)
			// Don't block execution, just warn
		}
	}
	
	return nil
}

// SetDryRun enables or disables dry-run mode
func (ce *CommandExecutor) SetDryRun(dryRun bool) {
	ce.dryRun = dryRun
}

// SetTimeout sets the default timeout for command execution
func (ce *CommandExecutor) SetTimeout(timeout time.Duration) {
	ce.timeout = timeout
}

// GetTimeout returns the current default timeout
func (ce *CommandExecutor) GetTimeout() time.Duration {
	return ce.timeout
}

// IsCommandAvailable checks if a command is available for execution
func (ce *CommandExecutor) IsCommandAvailable(command string) bool {
	return ce.validateCommand(command) == nil
}

// validateCommandBeforeExecution performs final validation before executing a command
func (ce *CommandExecutor) validateCommandBeforeExecution(cmd *exec.Cmd) error {
	if cmd == nil {
		return fmt.Errorf("command is nil")
	}
	
	if cmd.Path == "" {
		return fmt.Errorf("command path is empty")
	}
	
	// Check if the resolved executable exists and is executable
	if !ce.isExecutableAvailable(cmd.Path) {
		// Provide detailed error information
		ce.logger.Error("Executable validation failed", fmt.Errorf("executable not found or not executable"),
			interfaces.LogField{Key: "path", Value: cmd.Path},
			interfaces.LogField{Key: "args", Value: cmd.Args},
		)
		
		// Try to provide helpful suggestions
		if len(cmd.Args) > 0 {
			executable := cmd.Args[0]
			if _, err := exec.LookPath(executable); err != nil {
				return fmt.Errorf("executable '%s' not found in PATH. Please ensure it is installed and available. Resolved path: %s", executable, cmd.Path)
			}
		}
		
		return fmt.Errorf("executable not found or not executable: %s", cmd.Path)
	}
	
	// Log the final command that will be executed
	ce.logger.Debug("Final command validation passed",
		interfaces.LogField{Key: "path", Value: cmd.Path},
		interfaces.LogField{Key: "args", Value: cmd.Args},
		interfaces.LogField{Key: "working_directory", Value: cmd.Dir},
	)
	
	return nil
}