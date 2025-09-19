package executor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sai/internal/interfaces"
	"sai/internal/types"
)

// GenericExecutor implements provider action execution with template rendering and validation
type GenericExecutor struct {
	commandExecutor *CommandExecutor
	templateEngine  interfaces.TemplateEngine
	logger          interfaces.Logger
	validator       interfaces.ResourceValidator
}

// NewGenericExecutor creates a new generic executor
func NewGenericExecutor(
	commandExecutor *CommandExecutor,
	templateEngine interfaces.TemplateEngine,
	logger interfaces.Logger,
	validator interfaces.ResourceValidator,
) *GenericExecutor {
	return &GenericExecutor{
		commandExecutor: commandExecutor,
		templateEngine:  templateEngine,
		logger:          logger,
		validator:       validator,
	}
}

// Execute runs a provider action with the given options
func (ge *GenericExecutor) Execute(
	ctx context.Context,
	provider *types.ProviderData,
	action string,
	software string,
	saidata *types.SoftwareData,
	options interfaces.ExecuteOptions,
) (*interfaces.ExecutionResult, error) {
	startTime := time.Now()
	
	ge.logger.Info("Executing provider action",
		interfaces.LogField{Key: "provider", Value: provider.Provider.Name},
		interfaces.LogField{Key: "action", Value: action},
		interfaces.LogField{Key: "software", Value: software},
	)
	
	// Validate action exists
	providerAction, exists := provider.Actions[action]
	if !exists {
		return &interfaces.ExecutionResult{
			Success:  false,
			Error:    fmt.Errorf("action %s not found in provider %s", action, provider.Provider.Name),
			ExitCode: 1,
			Duration: time.Since(startTime),
			Provider: provider.Provider.Name,
		}, fmt.Errorf("action %s not found", action)
	}
	
	// Validate action can be executed
	if err := ge.ValidateAction(provider, action, software, saidata); err != nil {
		return &interfaces.ExecutionResult{
			Success:  false,
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
			Provider: provider.Provider.Name,
		}, err
	}
	
	// Handle dry-run mode
	if options.DryRun {
		return ge.DryRun(ctx, provider, action, software, saidata, options)
	}
	
	// Execute the action
	var result *interfaces.ExecutionResult
	var err error
	
	if providerAction.HasSteps() {
		result, err = ge.ExecuteSteps(ctx, providerAction.Steps, saidata, provider, options)
	} else {
		result, err = ge.executeSingleAction(ctx, &providerAction, software, saidata, provider, options)
	}
	
	if result != nil {
		result.Duration = time.Since(startTime)
		result.Provider = provider.Provider.Name
	}
	
	// Handle rollback on failure
	if err != nil && providerAction.Rollback != "" {
		ge.logger.Warn("Action failed, attempting rollback",
			interfaces.LogField{Key: "action", Value: action},
			interfaces.LogField{Key: "error", Value: err},
		)
		
		if rollbackErr := ge.executeRollback(ctx, providerAction.Rollback, software, saidata, provider, options); rollbackErr != nil {
			ge.logger.Error("Rollback failed", rollbackErr,
				interfaces.LogField{Key: "action", Value: action},
			)
		} else {
			ge.logger.Info("Rollback completed successfully",
				interfaces.LogField{Key: "action", Value: action},
			)
		}
	}
	
	return result, err
}

// ValidateAction validates that an action can be executed
func (ge *GenericExecutor) ValidateAction(
	provider *types.ProviderData,
	action string,
	software string,
	saidata *types.SoftwareData,
) error {
	// Check if action exists
	providerAction, exists := provider.Actions[action]
	if !exists {
		return fmt.Errorf("action %s not supported by provider %s", action, provider.Provider.Name)
	}
	
	// Validate action has execution method
	if !providerAction.IsValid() {
		return fmt.Errorf("action %s has no valid execution method", action)
	}
	
	// Validate template if present
	if providerAction.Template != "" {
		if err := ge.templateEngine.ValidateTemplate(providerAction.Template); err != nil {
			return fmt.Errorf("invalid template for action %s: %w", action, err)
		}
	}
	
	// Validate resources if saidata is available
	if saidata != nil {
		if validationResult, err := ge.ValidateResources(saidata, action); err != nil {
			return fmt.Errorf("resource validation failed: %w", err)
		} else if !validationResult.CanProceed {
			return fmt.Errorf("cannot proceed with action %s: missing required resources", action)
		}
	}
	
	return nil
}

// ValidateResources validates that required resources exist
func (ge *GenericExecutor) ValidateResources(
	saidata *types.SoftwareData,
	action string,
) (*interfaces.ResourceValidationResult, error) {
	if ge.validator == nil {
		return &interfaces.ResourceValidationResult{
			Valid:      true,
			CanProceed: true,
		}, nil
	}
	
	return ge.validator.ValidateResources(saidata)
}

// DryRun shows what would be executed without running commands
func (ge *GenericExecutor) DryRun(
	ctx context.Context,
	provider *types.ProviderData,
	action string,
	software string,
	saidata *types.SoftwareData,
	options interfaces.ExecuteOptions,
) (*interfaces.ExecutionResult, error) {
	startTime := time.Now()
	
	ge.logger.Info("DRY RUN: Showing what would be executed",
		interfaces.LogField{Key: "provider", Value: provider.Provider.Name},
		interfaces.LogField{Key: "action", Value: action},
		interfaces.LogField{Key: "software", Value: software},
	)
	
	providerAction := provider.Actions[action]
	var commands []string
	var output strings.Builder
	
	if providerAction.HasSteps() {
		// Render each step
		for i, step := range providerAction.Steps {
			rendered, err := ge.renderCommand(step.Command, software, saidata, provider, options)
			if err != nil {
				return &interfaces.ExecutionResult{
					Success:  false,
					Error:    fmt.Errorf("failed to render step %d: %w", i+1, err),
					ExitCode: 1,
					Duration: time.Since(startTime),
					Provider: provider.Provider.Name,
				}, err
			}
			commands = append(commands, rendered)
			output.WriteString(fmt.Sprintf("Step %d: %s\n", i+1, rendered))
		}
	} else {
		// Render single command
		command := providerAction.GetCommand()
		rendered, err := ge.renderCommand(command, software, saidata, provider, options)
		if err != nil {
			return &interfaces.ExecutionResult{
				Success:  false,
				Error:    fmt.Errorf("failed to render command: %w", err),
				ExitCode: 1,
				Duration: time.Since(startTime),
				Provider: provider.Provider.Name,
			}, err
		}
		commands = append(commands, rendered)
		output.WriteString(fmt.Sprintf("Command: %s\n", rendered))
	}
	
	return &interfaces.ExecutionResult{
		Success:  true,
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
		Commands: commands,
		Provider: provider.Provider.Name,
	}, nil
}

// CanExecute checks if an action can be executed
func (ge *GenericExecutor) CanExecute(
	provider *types.ProviderData,
	action string,
	software string,
	saidata *types.SoftwareData,
) bool {
	return ge.ValidateAction(provider, action, software, saidata) == nil
}

// RenderTemplate renders command templates with saidata variables
func (ge *GenericExecutor) RenderTemplate(
	templateStr string,
	saidata *types.SoftwareData,
	provider *types.ProviderData,
) (string, error) {
	context := &interfaces.TemplateContext{
		Software: "", // Will be set by caller if needed
		Provider: provider.Provider.Name,
		Saidata:  saidata,
	}
	
	return ge.templateEngine.Render(templateStr, context)
}

// ExecuteCommand executes a single command with proper error handling
func (ge *GenericExecutor) ExecuteCommand(
	ctx context.Context,
	command string,
	options interfaces.CommandOptions,
) (*interfaces.CommandResult, error) {
	return ge.commandExecutor.ExecuteCommand(ctx, command, options)
}

// ExecuteSteps executes multiple steps in sequence
func (ge *GenericExecutor) ExecuteSteps(
	ctx context.Context,
	steps []types.Step,
	saidata *types.SoftwareData,
	provider *types.ProviderData,
	options interfaces.ExecuteOptions,
) (*interfaces.ExecutionResult, error) {
	startTime := time.Now()
	var allOutput strings.Builder
	var allCommands []string
	var changes []interfaces.Change
	
	for i, step := range steps {
		ge.logger.Debug("Executing step",
			interfaces.LogField{Key: "step", Value: i + 1},
			interfaces.LogField{Key: "name", Value: step.Name},
		)
		
		// Check step condition if present
		if step.Condition != "" {
			shouldExecute, err := ge.evaluateCondition(step.Condition, saidata, provider)
			if err != nil {
				ge.logger.Warn("Failed to evaluate step condition",
					interfaces.LogField{Key: "step", Value: i + 1},
					interfaces.LogField{Key: "condition", Value: step.Condition},
					interfaces.LogField{Key: "error", Value: err},
				)
			}
			if !shouldExecute {
				ge.logger.Debug("Skipping step due to condition",
					interfaces.LogField{Key: "step", Value: i + 1},
				)
				continue
			}
		}
		
		// Render step command
		rendered, err := ge.renderCommand(step.Command, "", saidata, provider, options)
		if err != nil {
			if step.IgnoreFailure {
				ge.logger.Warn("Step command rendering failed, ignoring",
					interfaces.LogField{Key: "step", Value: i + 1},
					interfaces.LogField{Key: "error", Value: err},
				)
				continue
			}
			return &interfaces.ExecutionResult{
				Success:  false,
				Output:   allOutput.String(),
				Error:    fmt.Errorf("failed to render step %d command: %w", i+1, err),
				ExitCode: 1,
				Duration: time.Since(startTime),
				Commands: allCommands,
				Provider: provider.Provider.Name,
				Changes:  changes,
			}, err
		}
		
		allCommands = append(allCommands, rendered)
		
		// Execute step command
		stepTimeout := options.Timeout
		if step.Timeout > 0 {
			stepTimeout = time.Duration(step.Timeout) * time.Second
		}
		
		cmdOptions := interfaces.CommandOptions{
			Timeout: stepTimeout,
			WorkDir: options.WorkDir,
			Env:     options.Env,
			Verbose: options.Verbose,
		}
		
		result, err := ge.commandExecutor.ExecuteCommand(ctx, rendered, cmdOptions)
		if result != nil {
			allOutput.WriteString(result.Output)
			allOutput.WriteString("\n")
		}
		
		if err != nil || (result != nil && result.ExitCode != 0) {
			if step.IgnoreFailure {
				ge.logger.Warn("Step failed, ignoring",
					interfaces.LogField{Key: "step", Value: i + 1},
					interfaces.LogField{Key: "error", Value: err},
				)
				continue
			}
			
			return &interfaces.ExecutionResult{
				Success:  false,
				Output:   allOutput.String(),
				Error:    fmt.Errorf("step %d failed: %w", i+1, err),
				ExitCode: result.ExitCode,
				Duration: time.Since(startTime),
				Commands: allCommands,
				Provider: provider.Provider.Name,
				Changes:  changes,
			}, err
		}
		
		ge.logger.Debug("Step completed successfully",
			interfaces.LogField{Key: "step", Value: i + 1},
		)
	}
	
	return &interfaces.ExecutionResult{
		Success:  true,
		Output:   allOutput.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
		Commands: allCommands,
		Provider: provider.Provider.Name,
		Changes:  changes,
	}, nil
}

// executeSingleAction executes a single action (non-step based)
func (ge *GenericExecutor) executeSingleAction(
	ctx context.Context,
	action *types.Action,
	software string,
	saidata *types.SoftwareData,
	provider *types.ProviderData,
	options interfaces.ExecuteOptions,
) (*interfaces.ExecutionResult, error) {
	startTime := time.Now()
	
	// Get command to execute
	command := action.GetCommand()
	if command == "" {
		return &interfaces.ExecutionResult{
			Success:  false,
			Error:    fmt.Errorf("no command found for action"),
			ExitCode: 1,
			Duration: time.Since(startTime),
			Provider: provider.Provider.Name,
		}, fmt.Errorf("no command found for action")
	}
	
	// Render command template
	rendered, err := ge.renderCommand(command, software, saidata, provider, options)
	if err != nil {
		return &interfaces.ExecutionResult{
			Success:  false,
			Error:    fmt.Errorf("failed to render command: %w", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
			Provider: provider.Provider.Name,
		}, err
	}
	
	// Set up command options
	cmdOptions := interfaces.CommandOptions{
		Timeout: action.GetTimeout(),
		WorkDir: options.WorkDir,
		Env:     options.Env,
		Verbose: options.Verbose,
	}
	
	// Execute with retry if configured
	var result *interfaces.CommandResult
	if action.Retry != nil {
		result, err = ge.commandExecutor.ExecuteWithRetry(ctx, rendered, cmdOptions, action.Retry)
	} else {
		result, err = ge.commandExecutor.ExecuteCommand(ctx, rendered, cmdOptions)
	}
	
	// Validate result if validation is configured
	if err == nil && action.Validation != nil {
		if validationErr := ge.validateActionResult(result, action.Validation); validationErr != nil {
			err = fmt.Errorf("action validation failed: %w", validationErr)
		}
	}
	
	executionResult := &interfaces.ExecutionResult{
		Success:  err == nil && result.ExitCode == 0,
		Output:   result.Output,
		Error:    err,
		ExitCode: result.ExitCode,
		Duration: time.Since(startTime),
		Commands: []string{rendered},
		Provider: provider.Provider.Name,
	}
	
	return executionResult, err
}

// renderCommand renders a command template with the current context
func (ge *GenericExecutor) renderCommand(
	command string,
	software string,
	saidata *types.SoftwareData,
	provider *types.ProviderData,
	options interfaces.ExecuteOptions,
) (string, error) {
	context := &interfaces.TemplateContext{
		Software:  software,
		Provider:  provider.Provider.Name,
		Saidata:   saidata,
		Variables: options.Variables,
	}
	
	return ge.templateEngine.Render(command, context)
}

// evaluateCondition evaluates a step condition
func (ge *GenericExecutor) evaluateCondition(
	condition string,
	saidata *types.SoftwareData,
	provider *types.ProviderData,
) (bool, error) {
	// For now, implement basic condition evaluation
	// This could be extended to support more complex expressions
	
	// Render the condition as a template
	context := &interfaces.TemplateContext{
		Provider: provider.Provider.Name,
		Saidata:  saidata,
	}
	
	rendered, err := ge.templateEngine.Render(condition, context)
	if err != nil {
		return false, err
	}
	
	// Simple boolean evaluation
	switch strings.ToLower(strings.TrimSpace(rendered)) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid condition result: %s", rendered)
	}
}

// validateActionResult validates the result of an action execution
func (ge *GenericExecutor) validateActionResult(
	result *interfaces.CommandResult,
	validation *types.Validation,
) error {
	// Check exit code
	if validation.ExpectedExitCode != 0 && result.ExitCode != validation.ExpectedExitCode {
		return fmt.Errorf("expected exit code %d, got %d", validation.ExpectedExitCode, result.ExitCode)
	}
	
	// Check output pattern
	if validation.ExpectedOutput != "" {
		if !strings.Contains(result.Output, validation.ExpectedOutput) {
			return fmt.Errorf("expected output to contain '%s'", validation.ExpectedOutput)
		}
	}
	
	return nil
}

// executeRollback executes a rollback command
func (ge *GenericExecutor) executeRollback(
	ctx context.Context,
	rollbackCommand string,
	software string,
	saidata *types.SoftwareData,
	provider *types.ProviderData,
	options interfaces.ExecuteOptions,
) error {
	rendered, err := ge.renderCommand(rollbackCommand, software, saidata, provider, options)
	if err != nil {
		return fmt.Errorf("failed to render rollback command: %w", err)
	}
	
	cmdOptions := interfaces.CommandOptions{
		Timeout: 60 * time.Second, // Default rollback timeout
		WorkDir: options.WorkDir,
		Env:     options.Env,
		Verbose: options.Verbose,
	}
	
	result, err := ge.commandExecutor.ExecuteCommand(ctx, rendered, cmdOptions)
	if err != nil || result.ExitCode != 0 {
		return fmt.Errorf("rollback command failed: %w", err)
	}
	
	return nil
}