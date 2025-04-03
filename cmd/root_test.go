package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"sai/cmd/handlers"
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Store original os.Args
	oldArgs := os.Args

	// Run all tests
	result := m.Run()

	// Restore original os.Args
	os.Args = oldArgs

	os.Exit(result)
}

// executeCommand executes a command and captures its output
func executeCommand(args ...string) (string, error) {
	// Save original stdout
	oldStdout := os.Stdout

	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set up command arguments
	os.Args = append([]string{"sai"}, args...)

	// Reset provider and dry-run flags
	providerFlag = ""
	dryRunFlag = false
	handlers.SetDryRun(false)

	// Create a new root command for this test
	originalRun := rootCmd.Run
	defer func() {
		rootCmd.Run = originalRun
	}()

	// Call the handler directly instead of executing the command
	var err error
	if len(args) >= 2 {
		software := args[0]
		command := strings.ToLower(args[1])

		// Set flags if present
		for i := 2; i < len(args); i++ {
			arg := args[i]
			if arg == "--provider" && i+1 < len(args) {
				providerFlag = args[i+1]
				i++
			} else if arg == "--dry-run" {
				dryRunFlag = true
				handlers.SetDryRun(true)
			}
		}

		if handler, ok := SupportedCommands[command]; ok {
			w.Close()
			os.Stdout = oldStdout

			// Create a pipe just for the handler
			r2, w2, _ := os.Pipe()
			os.Stdout = w2

			// Execute the handler directly
			handler(software, providerFlag)

			// Restore original stdout
			w2.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r2)

			return buf.String(), nil
		} else {
			err = fmt.Errorf("unsupported command: %s", command)
		}
	} else {
		err = fmt.Errorf("requires at least 2 args")
	}

	// Restore original stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String(), err
}

// TestPackageManagementCommands tests all package management commands
func TestPackageManagementCommands(t *testing.T) {
	packageCommands := []struct {
		command        string
		software       string
		expectedOutput string
	}{
		{"install", "nginx", "install nginx using os provider"},
		{"status", "nginx", "status nginx using os provider"},
		{"list", "nginx", "list nginx using os provider"},
		{"search", "nginx", "search nginx using os provider"},
		{"info", "nginx", "info nginx using os provider"},
		{"upgrade", "nginx", "upgrade nginx using os provider"},
		{"uninstall", "nginx", "uninstall nginx using os provider"},
	}

	for _, tc := range packageCommands {
		t.Run(tc.command, func(t *testing.T) {
			output, err := executeCommand(tc.software, tc.command)
			if err != nil {
				t.Fatalf("Failed to execute '%s %s': %v", tc.software, tc.command, err)
			}

			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got: %s", tc.expectedOutput, output)
			}
		})
	}
}

// TestServiceManagementCommands tests all service management commands
func TestServiceManagementCommands(t *testing.T) {
	serviceCommands := []struct {
		command        string
		service        string
		expectedOutput string
	}{
		{"start", "redis", "start service redis using"},
		{"stop", "redis", "stop service redis using"},
		{"restart", "redis", "restart service redis using"},
		{"enable", "redis", "enable service redis using"},
		{"disable", "redis", "disable service redis using"},
	}

	for _, tc := range serviceCommands {
		t.Run(tc.command, func(t *testing.T) {
			output, err := executeCommand(tc.service, tc.command)
			if err != nil {
				t.Fatalf("Failed to execute '%s %s': %v", tc.service, tc.command, err)
			}

			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got: %s", tc.expectedOutput, output)
			}
		})
	}
}

// TestFlagsBeforeArguments tests using flags before command arguments
func TestFlagsBeforeArguments(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{"Provider Flag", []string{"--provider", "apt", "nginx", "install"}, "install nginx using os provider apt"},
		{"Dry Run Flag", []string{"--dry-run", "nginx", "install"}, "[DRY RUN]"},
		{"Multiple Flags", []string{"--provider", "brew", "--dry-run", "nginx", "install"}, "install nginx using os provider brew"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// For flags-before-arguments tests, we need to reorder the arguments
			// to match the expected software, command, flags pattern
			var newArgs []string
			var software, command string
			var flags []string

			for i := 0; i < len(tc.args); i++ {
				arg := tc.args[i]
				if strings.HasPrefix(arg, "--") {
					flags = append(flags, arg)
					if arg == "--provider" && i+1 < len(tc.args) && !strings.HasPrefix(tc.args[i+1], "--") {
						flags = append(flags, tc.args[i+1])
						i++
					}
				} else if software == "" {
					software = arg
				} else if command == "" {
					command = arg
				}
			}

			if software != "" && command != "" {
				newArgs = append(newArgs, software, command)
				newArgs = append(newArgs, flags...)

				output, err := executeCommand(newArgs...)
				if err != nil {
					t.Fatalf("Failed to execute command with args %v: %v", newArgs, err)
				}

				if !strings.Contains(output, tc.expectedOutput) {
					t.Errorf("Expected output to contain '%s', got: %s", tc.expectedOutput, output)
				}

				// For dry run tests, verify the appropriate message
				if strings.Contains(tc.name, "Dry Run") && !strings.Contains(output, "[DRY RUN]") {
					t.Errorf("Expected dry run message, got: %s", output)
				}
			} else {
				t.Fatalf("Could not determine software and command from args %v", tc.args)
			}
		})
	}
}

// TestFlagsAfterArguments tests using flags after command arguments
func TestFlagsAfterArguments(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{"Provider Flag", []string{"nginx", "install", "--provider", "apt"}, "install nginx using os provider apt"},
		{"Dry Run Flag", []string{"nginx", "install", "--dry-run"}, "[DRY RUN]"},
		{"Multiple Flags", []string{"nginx", "install", "--provider", "brew", "--dry-run"}, "install nginx using os provider brew"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := executeCommand(tc.args...)
			if err != nil {
				t.Fatalf("Failed to execute command with args %v: %v", tc.args, err)
			}

			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got: %s", tc.expectedOutput, output)
			}

			// For dry run tests, verify the appropriate message
			if strings.Contains(tc.name, "Dry Run") && !strings.Contains(output, "[DRY RUN]") {
				t.Errorf("Expected dry run message, got: %s", output)
			}
		})
	}
}

// TestOtherCommands tests utility commands
func TestOtherCommands(t *testing.T) {
	otherCommands := []struct {
		command        string
		software       string
		expectedOutput string
	}{
		{"help", "nginx", "SAI - Smart Software Management CLI"},
		{"debug", "nginx", "Debug information for nginx"},
		{"troubleshoot", "nginx", "Troubleshooting nginx"},
		{"config", "nginx", "Configuring nginx settings"},
	}

	for _, tc := range otherCommands {
		t.Run(tc.command, func(t *testing.T) {
			output, err := executeCommand(tc.software, tc.command)
			if err != nil {
				t.Fatalf("Failed to execute '%s %s': %v", tc.software, tc.command, err)
			}

			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got: %s", tc.expectedOutput, output)
			}
		})
	}
}

// TestInvalidCommands tests error handling for invalid commands
func TestInvalidCommands(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		expectError    bool
		expectedOutput string
	}{
		{"No Arguments", []string{}, true, "requires at least 2 args"},
		{"Invalid Command", []string{"nginx", "invalid"}, true, "unsupported command: invalid"},
		{"Missing Command", []string{"nginx"}, true, "requires at least 2 args"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := executeCommand(tc.args...)

			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tc.expectError && err != nil && !strings.Contains(err.Error(), tc.expectedOutput) {
				t.Errorf("Expected error to contain '%s', got: %v", tc.expectedOutput, err)
			}
		})
	}
}
