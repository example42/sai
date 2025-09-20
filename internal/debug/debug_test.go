package debug

import (
	"os"
	"testing"
	"time"
)

func TestDebugManager(t *testing.T) {
	// Test creating debug manager
	dm := NewDebugManager(true)
	if dm == nil {
		t.Fatal("Failed to create debug manager")
	}

	if !dm.IsEnabled() {
		t.Error("Debug manager should be enabled")
	}

	// Test logging provider detection
	allProviders := []string{"apt", "brew", "docker"}
	availableProviders := []string{"apt", "docker"}
	detectionResults := map[string]bool{
		"apt":    true,
		"brew":   false,
		"docker": true,
	}
	detectionTime := 100 * time.Millisecond

	dm.LogProviderDetection(allProviders, availableProviders, detectionResults, detectionTime)

	// Test logging template resolution
	template := "{{sai_package 'apt'}}"
	variables := map[string]interface{}{
		"software": "nginx",
		"provider": "apt",
	}
	result := "nginx"
	resolutionTime := 50 * time.Millisecond

	dm.LogTemplateResolution(template, variables, result, true, resolutionTime, nil)

	// Test logging command execution
	command := "apt install nginx"
	provider := "apt"
	args := []string{"install", "nginx"}
	env := []string{"PATH=/usr/bin"}
	workingDir := "/tmp"
	exitCode := 0
	output := "Reading package lists...\nDone"
	stderr := ""
	duration := 2 * time.Second

	dm.LogCommandExecution(command, provider, args, env, workingDir, exitCode, output, stderr, duration)

	// Test logging configuration loading
	configPath := "/etc/sai/config.yaml"
	configFound := true
	configData := map[string]interface{}{
		"log_level": "debug",
		"timeout":   "30s",
	}
	envOverrides := map[string]string{
		"SAI_LOG_LEVEL": "debug",
	}
	loadTime := 10 * time.Millisecond

	dm.LogConfigurationLoading(configPath, configFound, configData, envOverrides, loadTime, nil)

	// Test logging saidata loading
	software := "nginx"
	saidataPath := "/etc/sai/saidata/ng/nginx/default.yaml"
	osOverride := "ubuntu/22.04"
	mergeResults := map[string]interface{}{
		"packages": 1,
		"services": 1,
	}
	saidataLoadTime := 20 * time.Millisecond

	dm.LogSaidataLoading(software, saidataPath, osOverride, mergeResults, saidataLoadTime, true, nil)

	// Test logging decision making
	decisionType := "provider_selection"
	context := map[string]interface{}{
		"software": "nginx",
		"action":   "install",
	}
	options := []string{"apt", "snap"}
	selectedOption := "apt"
	reasoning := "apt has higher priority"
	decisionTime := 5 * time.Millisecond

	dm.LogDecisionMaking(decisionType, context, options, selectedOption, reasoning, decisionTime)

	// Test logging internal state
	component := "provider_manager"
	state := map[string]interface{}{
		"loaded_providers": 10,
		"available_count":  7,
	}

	dm.LogInternalState(component, state)

	// Test performance metrics
	dm.ShowPerformanceMetrics()

	// Clean up
	if err := dm.Close(); err != nil {
		t.Errorf("Failed to close debug manager: %v", err)
	}
}

func TestDebugManagerDisabled(t *testing.T) {
	// Test creating disabled debug manager
	dm := NewDebugManager(false)
	if dm == nil {
		t.Fatal("Failed to create debug manager")
	}

	if dm.IsEnabled() {
		t.Error("Debug manager should be disabled")
	}

	// Test that logging doesn't panic when disabled
	dm.LogProviderDetection([]string{"apt"}, []string{"apt"}, map[string]bool{"apt": true}, time.Millisecond)
	dm.LogTemplateResolution("test", nil, "result", true, time.Millisecond, nil)
	dm.LogCommandExecution("test", "apt", []string{}, []string{}, "", 0, "", "", time.Millisecond)
	dm.LogConfigurationLoading("", false, nil, nil, time.Millisecond, nil)
	dm.LogSaidataLoading("test", "", "", nil, time.Millisecond, true, nil)
	dm.LogDecisionMaking("test", nil, []string{}, "", "", time.Millisecond)
	dm.LogInternalState("test", nil)

	// Should not show metrics when disabled
	dm.ShowPerformanceMetrics()

	// Clean up
	if err := dm.Close(); err != nil {
		t.Errorf("Failed to close debug manager: %v", err)
	}
}

func TestGlobalDebugFunctions(t *testing.T) {
	// Test global debug functions
	dm := NewDebugManager(true)
	SetGlobalDebugManager(dm)

	if !IsDebugEnabled() {
		t.Error("Global debug should be enabled")
	}

	// Test global logging functions
	LogProviderDetectionGlobal([]string{"apt"}, []string{"apt"}, map[string]bool{"apt": true}, time.Millisecond)
	LogTemplateResolutionGlobal("test", nil, "result", true, time.Millisecond, nil)
	LogCommandExecutionGlobal("test", "apt", []string{}, []string{}, "", 0, "", "", time.Millisecond)
	LogConfigurationLoadingGlobal("", false, nil, nil, time.Millisecond, nil)
	LogSaidataLoadingGlobal("test", "", "", nil, time.Millisecond, true, nil)
	LogDecisionMakingGlobal("test", nil, []string{}, "", "", time.Millisecond)
	LogInternalStateGlobal("test", nil)

	// Clean up
	if err := dm.Close(); err != nil {
		t.Errorf("Failed to close debug manager: %v", err)
	}

	// Test with nil global manager
	SetGlobalDebugManager(nil)
	if IsDebugEnabled() {
		t.Error("Global debug should be disabled when manager is nil")
	}

	// Should not panic with nil manager
	LogProviderDetectionGlobal([]string{"apt"}, []string{"apt"}, map[string]bool{"apt": true}, time.Millisecond)
}

func TestDebugOutputFile(t *testing.T) {
	// Test that debug output file is created
	dm := NewDebugManager(true)
	defer dm.Close()

	// Log something to ensure file is created
	dm.LogSystemInfo()

	// Check that output file exists (it should be in temp directory)
	if dm.outputFile == nil {
		t.Error("Debug output file should be created when debug is enabled")
	}

	// Verify file exists
	if dm.outputFile != nil {
		if _, err := os.Stat(dm.outputFile.Name()); os.IsNotExist(err) {
			t.Error("Debug output file should exist on filesystem")
		}
	}
}