package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"sai/internal/config"
	"sai/internal/output"
)

// UserInterface handles user interactions and confirmations
type UserInterface struct {
	config    *config.Config
	formatter *output.OutputFormatter
	reader    *bufio.Reader
}

// ProviderOption represents a provider option for user selection
type ProviderOption struct {
	Name        string
	PackageName string
	Version     string
	IsInstalled bool
	Description string
	Command     string // New field for displaying the actual command (Requirement 15.3)
}

// NewUserInterface creates a new user interface
func NewUserInterface(cfg *config.Config, formatter *output.OutputFormatter) *UserInterface {
	return &UserInterface{
		config:    cfg,
		formatter: formatter,
		reader:    bufio.NewReader(os.Stdin),
	}
}

// ShowProviderSelection displays provider options and prompts for selection (Requirement 1.3)
func (ui *UserInterface) ShowProviderSelection(software string, options []*ProviderOption) (*ProviderOption, error) {
	if ui.formatter.IsJSONMode() {
		return ui.handleJSONProviderSelection(software, options)
	}

	if len(options) == 0 {
		return nil, fmt.Errorf("no providers available for %s", software)
	}

	if len(options) == 1 {
		return options[0], nil
	}

	ui.formatter.ShowInfo(fmt.Sprintf("Multiple providers available for %s:", software))
	fmt.Println()

	for i, option := range options {
		status := "Available"
		if option.IsInstalled {
			status = ui.formatter.FormatJSON(map[string]string{"status": "Installed"})
			if !ui.formatter.IsJSONMode() {
				status = "Installed"
			}
		}

		fmt.Printf("%d. %s\n", i+1, ui.formatter.FormatProviderName(option.Name))
		
		// Show command instead of package details (Requirements 15.1, 15.3)
		if option.Command != "" {
			fmt.Printf("   Command: %s\n", option.Command)
		} else {
			// Fallback to package info if no command available
			fmt.Printf("   Package: %s\n", option.PackageName)
			if option.Version != "" {
				fmt.Printf("   Version: %s\n", option.Version)
			}

		}
		
		fmt.Printf("   Status: %s\n\n", status)
	}

	for {
		fmt.Printf("Select provider (1-%d): ", len(options))
		input, err := ui.reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read user input: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(options) {
			ui.formatter.ShowError(fmt.Errorf("invalid selection. Please enter a number between 1 and %d", len(options)))
			continue
		}

		return options[choice-1], nil
	}
}

// ConfirmAction prompts for confirmation of system-changing actions (Requirements 9.1, 9.2)
func (ui *UserInterface) ConfirmAction(action, software, provider string, commands []string) (bool, error) {
	if ui.formatter.IsJSONMode() {
		return ui.handleJSONConfirmation(action, software, provider, commands)
	}

	// Check if confirmation is required for this action
	if !ui.config.RequiresConfirmation(action) {
		return true, nil
	}

	// Show command preview
	ui.formatter.ShowCommandPreview(commands, provider)

	// Prompt for confirmation
	prompt := fmt.Sprintf("Execute %s for %s using %s? (y/N): ", action, software, provider)
	fmt.Print(prompt)

	input, err := ui.reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes", nil
}

// PromptForInput prompts the user for input with a message
func (ui *UserInterface) PromptForInput(message string) (string, error) {
	if ui.formatter.IsJSONMode() {
		return "", fmt.Errorf("interactive input not supported in JSON mode")
	}

	fmt.Print(message)
	input, err := ui.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read user input: %w", err)
	}

	return strings.TrimSpace(input), nil
}

// PromptForConfirmation prompts for a yes/no confirmation
func (ui *UserInterface) PromptForConfirmation(message string) (bool, error) {
	if ui.formatter.IsJSONMode() {
		return false, fmt.Errorf("interactive confirmation not supported in JSON mode")
	}

	fmt.Printf("%s (y/N): ", message)
	input, err := ui.reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes", nil
}

// ShowCommandPreview displays commands that will be executed
func (ui *UserInterface) ShowCommandPreview(commands []string, provider string) {
	ui.formatter.ShowCommandPreview(commands, provider)
}

// IsQuietMode returns whether quiet mode is enabled
func (ui *UserInterface) IsQuietMode() bool {
	return ui.formatter.IsQuietMode()
}

// IsVerboseMode returns whether verbose mode is enabled
func (ui *UserInterface) IsVerboseMode() bool {
	return ui.formatter.IsVerboseMode()
}

// handleJSONProviderSelection handles provider selection in JSON mode
func (ui *UserInterface) handleJSONProviderSelection(software string, options []*ProviderOption) (*ProviderOption, error) {
	selectionData := map[string]interface{}{
		"type":     "provider_selection_required",
		"software": software,
		"options":  options,
		"message":  "Multiple providers available. Use --provider flag to specify one.",
	}

	fmt.Println(ui.formatter.FormatJSON(selectionData))
	return nil, fmt.Errorf("provider selection required in non-interactive mode")
}

// handleJSONConfirmation handles confirmation in JSON mode
func (ui *UserInterface) handleJSONConfirmation(action, software, provider string, commands []string) (bool, error) {
	confirmationData := map[string]interface{}{
		"type":     "confirmation_required",
		"action":   action,
		"software": software,
		"provider": provider,
		"commands": commands,
		"message":  "Use --yes flag to skip confirmation prompts.",
	}

	fmt.Println(ui.formatter.FormatJSON(confirmationData))
	return false, fmt.Errorf("confirmation required in non-interactive mode")
}

// ShowTable displays data in a table format
func (ui *UserInterface) ShowTable(headers []string, rows [][]string) {
	if ui.formatter.IsJSONMode() {
		tableData := map[string]interface{}{
			"type":    "table",
			"headers": headers,
			"rows":    rows,
		}
		fmt.Println(ui.formatter.FormatJSON(tableData))
		return
	}

	if len(rows) == 0 {
		return
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Print header
	ui.printTableRow(headers, colWidths)
	ui.printTableSeparator(colWidths)

	// Print rows
	for _, row := range rows {
		ui.printTableRow(row, colWidths)
	}
}

// printTableRow prints a single table row
func (ui *UserInterface) printTableRow(cells []string, colWidths []int) {
	for i, cell := range cells {
		if i < len(colWidths) {
			fmt.Printf("%-*s", colWidths[i]+2, cell)
		}
	}
	fmt.Println()
}

// printTableSeparator prints a table separator line
func (ui *UserInterface) printTableSeparator(colWidths []int) {
	for _, width := range colWidths {
		fmt.Print(strings.Repeat("-", width+2))
	}
	fmt.Println()
}