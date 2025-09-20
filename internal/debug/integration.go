package debug

import (
	"time"
)

// Global debug manager instance (will be set by CLI)
var globalDebugManager *DebugManager

// SetGlobalDebugManager sets the global debug manager instance
func SetGlobalDebugManager(dm *DebugManager) {
	globalDebugManager = dm
}

// GetGlobalDebugManager returns the global debug manager instance
func GetGlobalDebugManager() *DebugManager {
	return globalDebugManager
}

// IsDebugEnabled returns whether debug mode is globally enabled
func IsDebugEnabled() bool {
	if globalDebugManager == nil {
		return false
	}
	return globalDebugManager.IsEnabled()
}

// Convenience functions for common debug operations

// LogProviderDetectionGlobal logs provider detection using the global debug manager
func LogProviderDetectionGlobal(allProviders []string, availableProviders []string, detectionResults map[string]bool, detectionTime time.Duration) {
	if globalDebugManager != nil {
		globalDebugManager.LogProviderDetection(allProviders, availableProviders, detectionResults, detectionTime)
	}
}

// LogTemplateResolutionGlobal logs template resolution using the global debug manager
func LogTemplateResolutionGlobal(template string, variables map[string]interface{}, result string, success bool, resolutionTime time.Duration, err error) {
	if globalDebugManager != nil {
		globalDebugManager.LogTemplateResolution(template, variables, result, success, resolutionTime, err)
	}
}

// LogCommandExecutionGlobal logs command execution using the global debug manager
func LogCommandExecutionGlobal(command string, provider string, args []string, env []string, workingDir string, exitCode int, output string, stderr string, duration time.Duration) {
	if globalDebugManager != nil {
		globalDebugManager.LogCommandExecution(command, provider, args, env, workingDir, exitCode, output, stderr, duration)
	}
}

// LogConfigurationLoadingGlobal logs configuration loading using the global debug manager
func LogConfigurationLoadingGlobal(configPath string, configFound bool, configData map[string]interface{}, envOverrides map[string]string, loadTime time.Duration, err error) {
	if globalDebugManager != nil {
		globalDebugManager.LogConfigurationLoading(configPath, configFound, configData, envOverrides, loadTime, err)
	}
}

// LogSaidataLoadingGlobal logs saidata loading using the global debug manager
func LogSaidataLoadingGlobal(software string, saidataPath string, osOverride string, mergeResults map[string]interface{}, loadTime time.Duration, success bool, err error) {
	if globalDebugManager != nil {
		globalDebugManager.LogSaidataLoading(software, saidataPath, osOverride, mergeResults, loadTime, success, err)
	}
}

// LogDecisionMakingGlobal logs decision-making processes using the global debug manager
func LogDecisionMakingGlobal(decisionType string, context map[string]interface{}, options []string, selectedOption string, reasoning string, decisionTime time.Duration) {
	if globalDebugManager != nil {
		globalDebugManager.LogDecisionMaking(decisionType, context, options, selectedOption, reasoning, decisionTime)
	}
}

// LogInternalStateGlobal logs internal state using the global debug manager
func LogInternalStateGlobal(component string, state map[string]interface{}) {
	if globalDebugManager != nil {
		globalDebugManager.LogInternalState(component, state)
	}
}