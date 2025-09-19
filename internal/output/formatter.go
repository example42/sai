package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"sai/internal/config"
)

// OutputFormatter handles all output formatting for the SAI CLI
type OutputFormatter struct {
	config      *config.Config
	verboseMode bool
	quietMode   bool
	jsonMode    bool
	colorEnabled bool
}

// NewOutputFormatter creates a new output formatter with the given configuration
func NewOutputFormatter(cfg *config.Config, verbose, quiet, jsonOutput bool) *OutputFormatter {
	return &OutputFormatter{
		config:       cfg,
		verboseMode:  verbose,
		quietMode:    quiet,
		jsonMode:     jsonOutput,
		colorEnabled: !jsonOutput && isColorSupported(),
	}
}

// FormatCommand formats a command for display before execution (Requirement 10.1)
func (f *OutputFormatter) FormatCommand(command string, provider string) string {
	if f.jsonMode {
		return command // JSON mode doesn't format commands
	}

	providerTag := f.FormatProviderName(provider)
	boldCommand := f.bold(command)
	
	return fmt.Sprintf("%s %s", providerTag, boldCommand)
}

// FormatProviderName formats provider name with configurable background color (Requirement 10.2)
func (f *OutputFormatter) FormatProviderName(provider string) string {
	if f.jsonMode || !f.colorEnabled {
		return fmt.Sprintf("[%s]", provider)
	}

	colorFunc := f.getColorFunc(f.config.Output.ProviderColor)
	return colorFunc(fmt.Sprintf(" %s ", provider))
}

// FormatCommandOutput formats command output with exit status (Requirement 10.3)
func (f *OutputFormatter) FormatCommandOutput(output string, exitCode int) string {
	if f.jsonMode {
		return output // JSON mode handles output differently
	}

	if f.quietMode && exitCode == 0 {
		return "" // Suppress successful output in quiet mode
	}

	result := output
	if f.config.Output.ShowExitCodes {
		status := f.formatExitStatus(exitCode)
		if output != "" {
			result = fmt.Sprintf("%s\n%s", output, status)
		} else {
			result = status
		}
	}

	return result
}

// FormatJSON formats data as JSON output
func (f *OutputFormatter) FormatJSON(data interface{}) string {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal JSON: %s"}`, err.Error())
	}
	return string(jsonData)
}

// ShowProgress displays a progress message
func (f *OutputFormatter) ShowProgress(message string) {
	if f.quietMode || f.jsonMode {
		return
	}

	if f.colorEnabled {
		color.New(color.FgCyan).Println(message)
	} else {
		fmt.Println(message)
	}
}

// ShowError displays an error message
func (f *OutputFormatter) ShowError(err error) {
	if f.jsonMode {
		errorData := map[string]interface{}{
			"error": err.Error(),
			"type":  "error",
		}
		fmt.Println(f.FormatJSON(errorData))
		return
	}

	if f.colorEnabled {
		errorColor := f.getColorFunc(f.config.Output.ErrorColor)
		fmt.Fprintf(os.Stderr, "%s %s\n", errorColor("Error:"), err.Error())
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	}
}

// ShowSuccess displays a success message
func (f *OutputFormatter) ShowSuccess(message string) {
	if f.quietMode {
		return
	}

	if f.jsonMode {
		successData := map[string]interface{}{
			"message": message,
			"type":    "success",
		}
		fmt.Println(f.FormatJSON(successData))
		return
	}

	if f.colorEnabled {
		successColor := f.getColorFunc(f.config.Output.SuccessColor)
		fmt.Println(successColor(message))
	} else {
		fmt.Println(message)
	}
}

// ShowWarning displays a warning message
func (f *OutputFormatter) ShowWarning(message string) {
	if f.quietMode {
		return
	}

	if f.jsonMode {
		warningData := map[string]interface{}{
			"message": message,
			"type":    "warning",
		}
		fmt.Println(f.FormatJSON(warningData))
		return
	}

	if f.colorEnabled {
		color.New(color.FgYellow).Printf("Warning: %s\n", message)
	} else {
		fmt.Printf("Warning: %s\n", message)
	}
}

// ShowInfo displays an informational message
func (f *OutputFormatter) ShowInfo(message string) {
	if f.quietMode {
		return
	}

	if f.jsonMode {
		infoData := map[string]interface{}{
			"message": message,
			"type":    "info",
		}
		fmt.Println(f.FormatJSON(infoData))
		return
	}

	fmt.Println(message)
}

// ShowDebug displays a debug message (only in verbose mode)
func (f *OutputFormatter) ShowDebug(message string) {
	if !f.verboseMode || f.quietMode {
		return
	}

	if f.jsonMode {
		debugData := map[string]interface{}{
			"message": message,
			"type":    "debug",
		}
		fmt.Println(f.FormatJSON(debugData))
		return
	}

	if f.colorEnabled {
		color.New(color.FgMagenta).Printf("Debug: %s\n", message)
	} else {
		fmt.Printf("Debug: %s\n", message)
	}
}

// ShowCommandPreview shows commands that will be executed
func (f *OutputFormatter) ShowCommandPreview(commands []string, provider string) {
	if f.quietMode {
		return
	}

	if f.jsonMode {
		previewData := map[string]interface{}{
			"commands": commands,
			"provider": provider,
			"type":     "command_preview",
		}
		fmt.Println(f.FormatJSON(previewData))
		return
	}

	if len(commands) == 0 {
		return
	}

	fmt.Printf("Commands to execute:\n")
	for _, cmd := range commands {
		fmt.Printf("  %s\n", f.FormatCommand(cmd, provider))
	}
	fmt.Println()
}

// IsQuietMode returns whether quiet mode is enabled
func (f *OutputFormatter) IsQuietMode() bool {
	return f.quietMode
}

// IsVerboseMode returns whether verbose mode is enabled
func (f *OutputFormatter) IsVerboseMode() bool {
	return f.verboseMode
}

// IsJSONMode returns whether JSON output mode is enabled
func (f *OutputFormatter) IsJSONMode() bool {
	return f.jsonMode
}

// formatExitStatus formats the exit status with appropriate colors
func (f *OutputFormatter) formatExitStatus(exitCode int) string {
	if !f.colorEnabled {
		if exitCode == 0 {
			return "✓ Success"
		}
		return fmt.Sprintf("✗ Failed (exit code: %d)", exitCode)
	}

	if exitCode == 0 {
		successColor := f.getColorFunc(f.config.Output.SuccessColor)
		return successColor("✓ Success")
	}

	errorColor := f.getColorFunc(f.config.Output.ErrorColor)
	return errorColor(fmt.Sprintf("✗ Failed (exit code: %d)", exitCode))
}

// bold formats text in bold if colors are enabled
func (f *OutputFormatter) bold(text string) string {
	if !f.colorEnabled {
		return text
	}
	return color.New(color.Bold).Sprint(text)
}

// getColorFunc returns a color function based on color name
func (f *OutputFormatter) getColorFunc(colorName string) func(a ...interface{}) string {
	if !f.colorEnabled {
		return func(a ...interface{}) string {
			return fmt.Sprint(a...)
		}
	}

	switch strings.ToLower(colorName) {
	case "black":
		return color.New(color.FgBlack).Sprint
	case "red":
		return color.New(color.FgRed).Sprint
	case "green":
		return color.New(color.FgGreen).Sprint
	case "yellow":
		return color.New(color.FgYellow).Sprint
	case "blue":
		return color.New(color.FgBlue, color.BgWhite).Sprint
	case "magenta":
		return color.New(color.FgMagenta).Sprint
	case "cyan":
		return color.New(color.FgCyan).Sprint
	case "white":
		return color.New(color.FgWhite, color.BgBlack).Sprint
	default:
		return color.New(color.FgBlue, color.BgWhite).Sprint
	}
}

// isColorSupported checks if the terminal supports colors
func isColorSupported() bool {
	// Check if NO_COLOR environment variable is set
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if TERM environment variable indicates color support
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return false
	}

	// Check if we're in a CI environment that might not support colors
	if os.Getenv("CI") != "" && os.Getenv("COLORTERM") == "" && os.Getenv("FORCE_COLOR") == "" {
		return false
	}

	return true
}