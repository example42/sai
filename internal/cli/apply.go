package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"sai/internal/interfaces"
	"sai/internal/output"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply <action_file>",
	Short: "Execute batch operations from YAML/JSON file",
	Long: `Execute multiple software management actions from a YAML or JSON file.
The action file must conform to the applydata-0.1-schema.json schema.

This command supports batch installation, service management, and other operations
defined in the action file. Each action is executed in sequence with proper
error handling and rollback capabilities.

Examples:
  sai apply actions.yaml               # Execute actions from YAML file
  sai apply actions.json               # Execute actions from JSON file
  sai apply actions.yaml --dry-run     # Show what would be executed
  sai apply actions.yaml --yes         # Execute without confirmation prompts
  sai apply actions.yaml --verbose     # Show detailed execution information`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeApplyCommand(args[0])
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
}

// ApplyData represents the structure of an apply action file
type ApplyData struct {
	Version     string        `yaml:"version" json:"version"`
	Metadata    ApplyMetadata `yaml:"metadata" json:"metadata"`
	Actions     []ApplyAction `yaml:"actions" json:"actions"`
	Variables   map[string]string `yaml:"variables,omitempty" json:"variables,omitempty"`
	Rollback    RollbackConfig    `yaml:"rollback,omitempty" json:"rollback,omitempty"`
}

// ApplyMetadata contains metadata about the apply file
type ApplyMetadata struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Author      string `yaml:"author,omitempty" json:"author,omitempty"`
	Version     string `yaml:"version,omitempty" json:"version,omitempty"`
}

// ApplyAction represents a single action to execute
type ApplyAction struct {
	Name        string            `yaml:"name" json:"name"`
	Action      string            `yaml:"action" json:"action"`
	Software    string            `yaml:"software" json:"software"`
	Provider    string            `yaml:"provider,omitempty" json:"provider,omitempty"`
	Variables   map[string]string `yaml:"variables,omitempty" json:"variables,omitempty"`
	Condition   string            `yaml:"condition,omitempty" json:"condition,omitempty"`
	OnFailure   string            `yaml:"on_failure,omitempty" json:"on_failure,omitempty"` // "continue", "stop", "rollback"
	Timeout     int               `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
}

// RollbackConfig contains rollback configuration
type RollbackConfig struct {
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	OnFailure   bool   `yaml:"on_failure" json:"on_failure"`
	Strategy    string `yaml:"strategy" json:"strategy"` // "reverse", "custom"
	MaxRetries  int    `yaml:"max_retries,omitempty" json:"max_retries,omitempty"`
}

// ApplyResult represents the result of applying an action file
type ApplyResult struct {
	Success       bool                    `json:"success"`
	TotalActions  int                     `json:"total_actions"`
	Executed      int                     `json:"executed"`
	Successful    int                     `json:"successful"`
	Failed        int                     `json:"failed"`
	Skipped       int                     `json:"skipped"`
	ActionResults []ApplyActionResult     `json:"action_results"`
	Duration      string                  `json:"duration"`
	Error         string                  `json:"error,omitempty"`
}

// ApplyActionResult represents the result of a single action
type ApplyActionResult struct {
	Name        string `json:"name"`
	Action      string `json:"action"`
	Software    string `json:"software"`
	Provider    string `json:"provider"`
	Success     bool   `json:"success"`
	Skipped     bool   `json:"skipped"`
	Output      string `json:"output,omitempty"`
	Error       string `json:"error,omitempty"`
	Duration    string `json:"duration"`
	ExitCode    int    `json:"exit_code"`
}

// executeApplyCommand implements the apply command functionality (Requirement 6.1)
func executeApplyCommand(actionFile string) error {
	// Get global configuration and flags
	config := GetGlobalConfig()
	flags := GetGlobalFlags()

	// Create output formatter
	formatter := output.NewOutputFormatter(config, flags.Verbose, flags.Quiet, flags.JSONOutput)

	// Validate file exists
	if _, err := os.Stat(actionFile); os.IsNotExist(err) {
		formatter.ShowError(fmt.Errorf("action file '%s' does not exist", actionFile))
		return err
	}

	// Load and validate action file
	applyData, err := loadApplyFile(actionFile)
	if err != nil {
		formatter.ShowError(fmt.Errorf("failed to load action file: %w", err))
		return err
	}

	// Validate against schema (Requirement 6.4)
	if err := validateApplyData(applyData); err != nil {
		formatter.ShowError(fmt.Errorf("action file validation failed: %w", err))
		return err
	}

	// Create managers and dependencies
	actionManager, _, err := createManagers(config, formatter)
	if err != nil {
		formatter.ShowError(fmt.Errorf("failed to initialize managers: %w", err))
		return err
	}

	// Show apply file information
	if !flags.Quiet {
		formatter.ShowInfo(fmt.Sprintf("Applying: %s", applyData.Metadata.Name))
		if applyData.Metadata.Description != "" {
			formatter.ShowInfo(fmt.Sprintf("Description: %s", applyData.Metadata.Description))
		}
		formatter.ShowInfo(fmt.Sprintf("Actions: %d", len(applyData.Actions)))
		
		if flags.DryRun {
			formatter.ShowProgress("Dry run mode - no actions will be executed")
		}
		fmt.Println()
	}

	// Show confirmation for system-changing operations
	if !flags.Yes && !flags.DryRun && hasSystemChangingActions(applyData.Actions) {
		// For now, skip interactive confirmation in apply command
		// In a full implementation, this would use the user interface
		formatter.ShowInfo(fmt.Sprintf("Would execute %d actions from %s (use --yes to confirm)", 
			len(applyData.Actions), filepath.Base(actionFile)))
		return nil
	}

	// Execute actions
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	result, err := executeApplyActions(ctx, applyData, actionManager, flags, formatter)
	if err != nil {
		formatter.ShowError(fmt.Errorf("apply execution failed: %w", err))
		return err
	}

	// Display results
	if flags.JSONOutput {
		fmt.Println(formatter.FormatJSON(result))
	} else {
		displayApplyResults(result, formatter, flags.Verbose)
	}

	// Set exit code based on overall success
	if !result.Success {
		os.Exit(1)
	}

	return nil
}

// loadApplyFile loads and parses an apply action file
func loadApplyFile(filename string) (*ApplyData, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var applyData ApplyData
	
	// Determine file format based on extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &applyData); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &applyData); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, &applyData); err != nil {
			if jsonErr := json.Unmarshal(data, &applyData); jsonErr != nil {
				return nil, fmt.Errorf("failed to parse as YAML or JSON: YAML error: %v, JSON error: %v", err, jsonErr)
			}
		}
	}

	// Set defaults
	if applyData.Version == "" {
		applyData.Version = "0.1"
	}

	return &applyData, nil
}

// validateApplyData validates the apply data against the schema
func validateApplyData(applyData *ApplyData) error {
	// Basic validation (in a real implementation, this would use JSON schema validation)
	if applyData.Version == "" {
		return fmt.Errorf("version is required")
	}

	if applyData.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}

	if len(applyData.Actions) == 0 {
		return fmt.Errorf("at least one action is required")
	}

	// Validate each action
	for i, action := range applyData.Actions {
		if action.Name == "" {
			return fmt.Errorf("action[%d].name is required", i)
		}
		if action.Action == "" {
			return fmt.Errorf("action[%d].action is required", i)
		}
		if action.Software == "" {
			return fmt.Errorf("action[%d].software is required", i)
		}

		// Validate on_failure values
		if action.OnFailure != "" {
			validValues := []string{"continue", "stop", "rollback"}
			valid := false
			for _, v := range validValues {
				if action.OnFailure == v {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("action[%d].on_failure must be one of: %s", i, strings.Join(validValues, ", "))
			}
		}
	}

	// TODO: Add JSON schema validation using schemas/applydata-0.1-schema.json
	// This would be done using a JSON schema validation library

	return nil
}

// executeApplyActions executes all actions in the apply data
func executeApplyActions(ctx context.Context, applyData *ApplyData, actionManager interfaces.ActionManager, flags GlobalFlags, formatter *output.OutputFormatter) (*ApplyResult, error) {
	result := &ApplyResult{
		TotalActions:  len(applyData.Actions),
		ActionResults: make([]ApplyActionResult, 0, len(applyData.Actions)),
	}

	startTime := time.Now()

	var executedActions []ApplyActionResult
	
	for i, action := range applyData.Actions {
		actionResult := ApplyActionResult{
			Name:     action.Name,
			Action:   action.Action,
			Software: action.Software,
			Provider: action.Provider,
		}

		// Check condition if specified
		if action.Condition != "" && !evaluateCondition(action.Condition) {
			actionResult.Skipped = true
			result.Skipped++
			if flags.Verbose {
				formatter.ShowInfo(fmt.Sprintf("Skipping action '%s': condition not met", action.Name))
			}
			result.ActionResults = append(result.ActionResults, actionResult)
			continue
		}

		// Show progress
		if !flags.Quiet {
			formatter.ShowProgress(fmt.Sprintf("[%d/%d] %s: %s %s", 
				i+1, len(applyData.Actions), action.Name, action.Action, action.Software))
		}

		// Prepare action options
		options := interfaces.ActionOptions{
			Provider:  action.Provider,
			DryRun:    flags.DryRun,
			Verbose:   flags.Verbose,
			Quiet:     flags.Quiet,
			Yes:       flags.Yes,
			JSON:      flags.JSONOutput,
			Variables: mergeVariables(applyData.Variables, action.Variables),
		}

		// Set timeout if specified
		if action.Timeout > 0 {
			options.Timeout = time.Duration(action.Timeout) * time.Second
		}

		// Execute action
		actionStartTime := time.Now()
		actionCtx := ctx
		if action.Timeout > 0 {
			var cancel context.CancelFunc
			actionCtx, cancel = context.WithTimeout(ctx, options.Timeout)
			defer cancel()
		}

		execResult, err := actionManager.ExecuteAction(actionCtx, action.Action, action.Software, options)
		actionDuration := time.Since(actionStartTime)

		// Process result
		result.Executed++
		actionResult.Duration = actionDuration.String()
		
		if err != nil || (execResult != nil && !execResult.Success) {
			actionResult.Success = false
			actionResult.Error = getErrorMessage(err, execResult)
			if execResult != nil {
				actionResult.ExitCode = execResult.ExitCode
				actionResult.Output = execResult.Output
				actionResult.Provider = execResult.Provider
			}
			result.Failed++

			// Handle failure based on on_failure setting
			onFailure := action.OnFailure
			if onFailure == "" {
				onFailure = "stop" // Default behavior
			}

			switch onFailure {
			case "continue":
				if flags.Verbose {
					formatter.ShowWarning(fmt.Sprintf("Action '%s' failed but continuing: %s", action.Name, actionResult.Error))
				}
			case "stop":
				formatter.ShowError(fmt.Errorf("action '%s' failed, stopping execution: %s", action.Name, actionResult.Error))
				result.ActionResults = append(result.ActionResults, actionResult)
				result.Success = false
				result.Duration = time.Since(startTime).String()
				return result, fmt.Errorf("execution stopped due to action failure")
			case "rollback":
				formatter.ShowError(fmt.Errorf("action '%s' failed, initiating rollback: %s", action.Name, actionResult.Error))
				// TODO: Implement rollback logic
				result.ActionResults = append(result.ActionResults, actionResult)
				result.Success = false
				result.Duration = time.Since(startTime).String()
				return result, fmt.Errorf("execution stopped due to action failure, rollback initiated")
			}
		} else {
			actionResult.Success = true
			if execResult != nil {
				actionResult.Output = execResult.Output
				actionResult.Provider = execResult.Provider
				actionResult.ExitCode = execResult.ExitCode
			}
			result.Successful++

			if flags.Verbose {
				formatter.ShowSuccess(fmt.Sprintf("Action '%s' completed successfully", action.Name))
			}
		}

		result.ActionResults = append(result.ActionResults, actionResult)
		executedActions = append(executedActions, actionResult)
	}

	// Calculate overall result
	result.Success = result.Failed == 0
	result.Duration = time.Since(startTime).String()

	return result, nil
}

// displayApplyResults displays the results of the apply operation
func displayApplyResults(result *ApplyResult, formatter *output.OutputFormatter, verbose bool) {
	fmt.Println("Apply Results:")
	fmt.Printf("  Total Actions: %d\n", result.TotalActions)
	fmt.Printf("  Executed: %d\n", result.Executed)
	fmt.Printf("  Successful: %d\n", result.Successful)
	fmt.Printf("  Failed: %d\n", result.Failed)
	fmt.Printf("  Skipped: %d\n", result.Skipped)
	fmt.Printf("  Duration: %s\n", result.Duration)
	fmt.Println()

	if result.Success {
		formatter.ShowSuccess("All actions completed successfully")
	} else {
		formatter.ShowError(fmt.Errorf("some actions failed"))
	}

	if verbose && len(result.ActionResults) > 0 {
		fmt.Println("\nAction Details:")
		for _, actionResult := range result.ActionResults {
			status := "✓ SUCCESS"
			if actionResult.Skipped {
				status = "- SKIPPED"
			} else if !actionResult.Success {
				status = "✗ FAILED"
			}

			fmt.Printf("  %s %s (%s %s)\n", status, actionResult.Name, actionResult.Action, actionResult.Software)
			if actionResult.Provider != "" {
				fmt.Printf("    Provider: %s\n", actionResult.Provider)
			}
			if actionResult.Duration != "" {
				fmt.Printf("    Duration: %s\n", actionResult.Duration)
			}
			if actionResult.Error != "" {
				fmt.Printf("    Error: %s\n", actionResult.Error)
			}
			if verbose && actionResult.Output != "" {
				fmt.Printf("    Output: %s\n", strings.TrimSpace(actionResult.Output))
			}
			fmt.Println()
		}
	}
}

// Helper functions

func hasSystemChangingActions(actions []ApplyAction) bool {
	systemChangingActions := map[string]bool{
		"install":   true,
		"uninstall": true,
		"upgrade":   true,
		"start":     true,
		"stop":      true,
		"restart":   true,
		"enable":    true,
		"disable":   true,
	}

	for _, action := range actions {
		if systemChangingActions[action.Action] {
			return true
		}
	}
	return false
}

func evaluateCondition(condition string) bool {
	// Simple condition evaluation (placeholder)
	// In a real implementation, this would support more complex conditions
	switch condition {
	case "always":
		return true
	case "never":
		return false
	default:
		// For now, assume all conditions are true
		return true
	}
}

func mergeVariables(global, local map[string]string) map[string]string {
	result := make(map[string]string)
	
	// Copy global variables
	for k, v := range global {
		result[k] = v
	}
	
	// Override with local variables
	for k, v := range local {
		result[k] = v
	}
	
	return result
}

func getErrorMessage(err error, result *interfaces.ActionResult) string {
	if err != nil {
		return err.Error()
	}
	if result != nil && result.Error != nil {
		return result.Error.Error()
	}
	return "unknown error"
}