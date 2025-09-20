package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceCommands(t *testing.T) {
	// Test that all service management commands are properly registered
	commands := rootCmd.Commands()
	
	// Create a map for easier lookup
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Use] = true
	}
	
	// Test service lifecycle commands
	assert.True(t, commandMap["start [software]"], "start command should be registered")
	assert.True(t, commandMap["stop [software]"], "stop command should be registered")
	assert.True(t, commandMap["restart [software]"], "restart command should be registered")
	assert.True(t, commandMap["enable [software]"], "enable command should be registered")
	assert.True(t, commandMap["disable [software]"], "disable command should be registered")
	assert.True(t, commandMap["status [software]"], "status command should be registered")
	
	// Test monitoring commands
	assert.True(t, commandMap["logs [software]"], "logs command should be registered")
	assert.True(t, commandMap["cpu [software]"], "cpu command should be registered")
	assert.True(t, commandMap["memory [software]"], "memory command should be registered")
	assert.True(t, commandMap["io [software]"], "io command should be registered")
	
	// Test additional service commands
	assert.True(t, commandMap["check [software]"], "check command should be registered")
	assert.True(t, commandMap["config [software]"], "config command should be registered")
}

func TestServiceCommandProperties(t *testing.T) {
	// Test start command properties
	assert.Equal(t, "start [software]", startCmd.Use)
	assert.Equal(t, "Start software service", startCmd.Short)
	assert.Nil(t, startCmd.Args(nil, []string{"nginx"})) // ExactArgs(1) - one arg OK
	
	// Test stop command properties
	assert.Equal(t, "stop [software]", stopCmd.Use)
	assert.Equal(t, "Stop software service", stopCmd.Short)
	assert.Nil(t, stopCmd.Args(nil, []string{"nginx"})) // ExactArgs(1) - one arg OK
	
	// Test status command properties (information-only)
	assert.Equal(t, "status [software]", statusCmd.Use)
	assert.Equal(t, "Check software service status", statusCmd.Short)
	assert.Nil(t, statusCmd.Args(nil, []string{"nginx"})) // ExactArgs(1) - one arg OK
	
	// Test logs command properties (can work with or without software parameter)
	assert.Equal(t, "logs [software]", logsCmd.Use)
	assert.Equal(t, "Display software service logs", logsCmd.Short)
	assert.Nil(t, logsCmd.Args(nil, []string{}))        // MaximumNArgs(1) - no args OK
	assert.Nil(t, logsCmd.Args(nil, []string{"nginx"})) // MaximumNArgs(1) - one arg OK
	
	// Test cpu command properties (can work with or without software parameter)
	assert.Equal(t, "cpu [software]", cpuCmd.Use)
	assert.Equal(t, "Display CPU usage statistics", cpuCmd.Short)
	assert.Nil(t, cpuCmd.Args(nil, []string{}))        // MaximumNArgs(1) - no args OK
	assert.Nil(t, cpuCmd.Args(nil, []string{"nginx"})) // MaximumNArgs(1) - one arg OK
}

func TestServiceCommandHelp(t *testing.T) {
	// Test that service commands have proper help text
	assert.Contains(t, startCmd.Long, "Start the service for the specified software")
	assert.Contains(t, startCmd.Long, "systemd, launchd")
	assert.Contains(t, startCmd.Long, "--dry-run")
	
	assert.Contains(t, statusCmd.Long, "information-only command")
	assert.Contains(t, statusCmd.Long, "executes without confirmation prompts")
	
	assert.Contains(t, logsCmd.Long, "general system logs if no software is specified")
	assert.Contains(t, logsCmd.Long, "information-only command")
	
	assert.Contains(t, cpuCmd.Long, "general system CPU usage if no software is specified")
	assert.Contains(t, cpuCmd.Long, "information-only command")
}

func TestServiceCommandExamples(t *testing.T) {
	// Test that service commands have proper examples
	assert.Contains(t, startCmd.Long, "sai start nginx")
	assert.Contains(t, startCmd.Long, "sai start nginx --dry-run")
	assert.Contains(t, startCmd.Long, "sai start nginx --yes")
	assert.Contains(t, startCmd.Long, "sai start nginx --provider systemd")
	
	assert.Contains(t, statusCmd.Long, "sai status nginx")
	assert.Contains(t, statusCmd.Long, "sai status nginx --json")
	assert.Contains(t, statusCmd.Long, "sai status nginx --verbose")
	
	assert.Contains(t, logsCmd.Long, "sai logs nginx")
	assert.Contains(t, logsCmd.Long, "sai logs")
	
	assert.Contains(t, cpuCmd.Long, "sai cpu nginx")
	assert.Contains(t, cpuCmd.Long, "sai cpu")
}