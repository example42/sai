package handlers

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureOutput captures stdout during test execution
func captureOutput(f func()) string {
	// Save original stdout
	oldStdout := os.Stdout

	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the function
	f()

	// Restore original stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}

// Define a common interface that all handlers implement
type HandlerInterface interface {
	Handle(string, string)
}

// TestBaseHandlerPackageCommands tests the BaseHandler with package management commands
func TestBaseHandlerPackageCommands(t *testing.T) {
	testCases := []struct {
		handlerFactory func() HandlerInterface
		action         string
		software       string
		provider       string
		expectedOutput string
	}{
		{func() HandlerInterface { return NewInstallHandler() }, "install", "nginx", "", "install nginx using os provider"},
		{func() HandlerInterface { return NewInstallHandler() }, "install", "nginx", "apt", "install nginx using os provider apt"},
		{func() HandlerInterface { return NewStatusHandler() }, "status", "nginx", "", "status nginx using os provider"},
		{func() HandlerInterface { return NewListHandler() }, "list", "nginx", "", "list nginx using os provider"},
		{func() HandlerInterface { return NewSearchHandler() }, "search", "nginx", "", "search nginx using os provider"},
		{func() HandlerInterface { return NewInfoHandler() }, "info", "nginx", "", "info nginx using os provider"},
		{func() HandlerInterface { return NewUninstallHandler() }, "uninstall", "nginx", "", "uninstall nginx using os provider"},
	}

	for _, tc := range testCases {
		t.Run(tc.action, func(t *testing.T) {
			// Reset dry run mode
			SetDryRun(false)

			// Capture output when executing the handler
			output := captureOutput(func() {
				handler := tc.handlerFactory()
				handler.Handle(tc.software, tc.provider)
			})

			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got: %s", tc.expectedOutput, output)
			}
		})
	}
}

// TestBaseHandlerServiceCommands tests the BaseHandler with service management commands
func TestBaseHandlerServiceCommands(t *testing.T) {
	testCases := []struct {
		handlerFactory func() HandlerInterface
		action         string
		service        string
		expectedOutput string
	}{
		{func() HandlerInterface { return NewStartHandler() }, "start", "redis", "start service redis using"},
		{func() HandlerInterface { return NewStopHandler() }, "stop", "redis", "stop service redis using"},
		{func() HandlerInterface { return NewRestartHandler() }, "restart", "redis", "restart service redis using"},
		{func() HandlerInterface { return NewEnableHandler() }, "enable", "redis", "enable service redis using"},
		{func() HandlerInterface { return NewDisableHandler() }, "disable", "redis", "disable service redis using"},
	}

	for _, tc := range testCases {
		t.Run(tc.action, func(t *testing.T) {
			// Reset dry run mode
			SetDryRun(false)

			// Capture output when executing the handler
			output := captureOutput(func() {
				handler := tc.handlerFactory()
				handler.Handle(tc.service, "")
			})

			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got: %s", tc.expectedOutput, output)
			}
		})
	}
}

// TestDryRunMode tests that dry run mode prevents actual command execution
func TestDryRunMode(t *testing.T) {
	testCases := []struct {
		name           string
		handlerFactory func() HandlerInterface
		software       string
		expectedOutput string
	}{
		{"Package Command", func() HandlerInterface { return NewInstallHandler() }, "nginx", "[DRY RUN] Command would be executed"},
		{"Service Command", func() HandlerInterface { return NewStartHandler() }, "redis", "[DRY RUN] Service command would be executed"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Enable dry run mode
			SetDryRun(true)

			// Capture output when executing the handler
			output := captureOutput(func() {
				handler := tc.handlerFactory()
				handler.Handle(tc.software, "")
			})

			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got: %s", tc.expectedOutput, output)
			}

			// Reset dry run mode
			SetDryRun(false)
		})
	}
}

// TestUtilityHandlers tests the utility command handlers
func TestUtilityHandlers(t *testing.T) {
	testCases := []struct {
		handlerFactory func() HandlerInterface
		name           string
		software       string
		expectedOutput string
	}{
		{func() HandlerInterface { return NewHelpHandler() }, "help", "nginx", "SAI - Smart Software Management CLI"},
		{func() HandlerInterface { return NewDebugHandler() }, "debug", "nginx", "Debug information for nginx"},
		{func() HandlerInterface { return NewTroubleshootHandler() }, "troubleshoot", "nginx", "Troubleshooting nginx"},
		{func() HandlerInterface { return NewConfigHandler() }, "config", "nginx", "Configuring nginx settings"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture output when executing the handler
			output := captureOutput(func() {
				handler := tc.handlerFactory()
				handler.Handle(tc.software, "")
			})

			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got: %s", tc.expectedOutput, output)
			}
		})
	}
}

// TestHandlerCreation ensures all handlers can be created without errors
func TestHandlerCreation(t *testing.T) {
	handlers := []struct {
		name           string
		handlerFactory func() HandlerInterface
	}{
		{"InstallHandler", func() HandlerInterface { return NewInstallHandler() }},
		{"StatusHandler", func() HandlerInterface { return NewStatusHandler() }},
		{"ListHandler", func() HandlerInterface { return NewListHandler() }},
		{"SearchHandler", func() HandlerInterface { return NewSearchHandler() }},
		{"InfoHandler", func() HandlerInterface { return NewInfoHandler() }},
		{"UninstallHandler", func() HandlerInterface { return NewUninstallHandler() }},
		{"UpgradeHandler", func() HandlerInterface { return NewUpgradeHandler() }},
		{"StartHandler", func() HandlerInterface { return NewStartHandler() }},
		{"StopHandler", func() HandlerInterface { return NewStopHandler() }},
		{"RestartHandler", func() HandlerInterface { return NewRestartHandler() }},
		{"EnableHandler", func() HandlerInterface { return NewEnableHandler() }},
		{"DisableHandler", func() HandlerInterface { return NewDisableHandler() }},
		{"HelpHandler", func() HandlerInterface { return NewHelpHandler() }},
		{"DebugHandler", func() HandlerInterface { return NewDebugHandler() }},
		{"TroubleshootHandler", func() HandlerInterface { return NewTroubleshootHandler() }},
		{"ConfigHandler", func() HandlerInterface { return NewConfigHandler() }},
	}

	for _, h := range handlers {
		t.Run(h.name, func(t *testing.T) {
			handler := h.handlerFactory()
			if handler == nil {
				t.Errorf("Failed to create %s", h.name)
			}
		})
	}
}
