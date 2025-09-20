package cli

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"sai/internal/config"
	"sai/internal/interfaces"
	"sai/internal/output"
	"sai/internal/ui"
)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display system and provider statistics",
	Long: `Display comprehensive statistics about available providers, actions, and system capabilities.
This provides detailed information about the SAI environment and available functionality.

This is an information-only command that executes without confirmation prompts.
The output includes provider availability, supported actions, system information, and capability matrix.

Examples:
  sai stats                            # Show all statistics
  sai stats --verbose                  # Show detailed statistics with additional information
  sai stats --json                     # Output statistics in JSON format`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeStatsCommand()
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}

// SystemStats represents comprehensive system and provider statistics
type SystemStats struct {
	System    SystemInfo      `json:"system"`
	Providers []ProviderStats `json:"providers"`
	Actions   ActionStats     `json:"actions"`
	Summary   StatsSummary    `json:"summary"`
}

// SystemInfo represents system information
type SystemInfo struct {
	Platform     string `json:"platform"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Version      string `json:"version"`
	SAIVersion   string `json:"sai_version"`
}

// ProviderStats represents statistics for a provider
type ProviderStats struct {
	Name         string   `json:"name"`
	DisplayName  string   `json:"display_name"`
	Type         string   `json:"type"`
	Available    bool     `json:"available"`
	Priority     int      `json:"priority"`
	Platforms    []string `json:"platforms"`
	Capabilities []string `json:"capabilities"`
	Actions      []string `json:"actions"`
	Executable   string   `json:"executable,omitempty"`
	Status       string   `json:"status"`
	Error        string   `json:"error,omitempty"`
}

// ActionStats represents statistics about available actions
type ActionStats struct {
	TotalActions      int                    `json:"total_actions"`
	ActionsByCategory map[string][]string    `json:"actions_by_category"`
	ActionProviders   map[string][]string    `json:"action_providers"`
	Coverage          map[string]int         `json:"coverage"` // How many providers support each action
}

// StatsSummary represents summary statistics
type StatsSummary struct {
	TotalProviders     int `json:"total_providers"`
	AvailableProviders int `json:"available_providers"`
	TotalActions       int `json:"total_actions"`
	PlatformSupport    int `json:"platform_support"` // Percentage of providers supporting current platform
}

// executeStatsCommand implements the stats command functionality (Requirement 6.2)
func executeStatsCommand() error {
	// Get global configuration and flags
	config := GetGlobalConfig()
	flags := GetGlobalFlags()

	// Check for debug flag (from --debug global flag)
	debug := flags.Debug

	// Create output formatter
	formatter := output.NewOutputFormatter(config, flags.Verbose, flags.Quiet, flags.JSONOutput)

	// Create managers and dependencies
	actionManager, userInterface, err := createManagers(config, formatter)
	if err != nil {
		formatter.ShowError(fmt.Errorf("failed to initialize managers: %w", err))
		return err
	}

	// Show debug information if requested (Requirement 13.4)
	if debug {
		fmt.Println("[DEBUG] === SAI Stats Command Debug Information ===")
		if pm := actionManager.GetProviderManager(); pm != nil {
			// Try to cast to concrete type to access debug methods
			if providerManager, ok := pm.(interface {
				LogProviderDetection(bool)
				GetDetectionStats() interface{}
				GetCacheStats() interface{}
			}); ok {
				providerManager.LogProviderDetection(debug)
			}
		}
	}

	// Show progress
	if !flags.Quiet {
		formatter.ShowProgress("Gathering system and provider statistics...")
	}

	// Collect statistics
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	stats, err := collectSystemStats(ctx, actionManager, config)
	if err != nil {
		formatter.ShowError(fmt.Errorf("failed to collect statistics: %w", err))
		return err
	}

	// Display results
	if flags.JSONOutput {
		fmt.Println(formatter.FormatJSON(stats))
	} else {
		displayStats(stats, formatter, userInterface, flags.Verbose)
	}

	// Show additional debug information after stats
	if debug {
		fmt.Println("\n[DEBUG] === Additional Debug Information ===")
		if pm := actionManager.GetProviderManager(); pm != nil {
			// Try to cast to concrete type to access debug methods
			if providerManager, ok := pm.(interface {
				GetDetectionStats() interface{}
				GetCacheStats() interface{}
			}); ok {
				// Use reflection-like approach to get stats
				detectionStatsInterface := providerManager.GetDetectionStats()
				cacheStatsInterface := providerManager.GetCacheStats()
				
				// Basic debug output without type assertions
				fmt.Printf("[DEBUG] Detection and cache statistics available\n")
				fmt.Printf("[DEBUG] Detection stats type: %T\n", detectionStatsInterface)
				fmt.Printf("[DEBUG] Cache stats type: %T\n", cacheStatsInterface)
			}
		}
	}

	return nil
}

// collectSystemStats collects comprehensive system and provider statistics
func collectSystemStats(ctx context.Context, actionManager interfaces.ActionManager, config *config.Config) (*SystemStats, error) {
	stats := &SystemStats{
		System: SystemInfo{
			Platform:     runtime.GOOS,
			OS:           getOSInfo(),
			Architecture: runtime.GOARCH,
			Version:      getOSVersion(),
			SAIVersion:   "0.1.0",
		},
		Providers: []ProviderStats{},
		Actions: ActionStats{
			ActionsByCategory: make(map[string][]string),
			ActionProviders:   make(map[string][]string),
			Coverage:          make(map[string]int),
		},
		Summary: StatsSummary{},
	}

	// Collect provider statistics from actual provider manager
	providerStats := getProviderStats(actionManager)
	stats.Providers = providerStats

	// Calculate provider summary
	stats.Summary.TotalProviders = len(providerStats)
	availableCount := 0
	platformSupportCount := 0
	
	for _, provider := range providerStats {
		if provider.Available {
			availableCount++
		}
		if supportsPlatform(provider.Platforms, stats.System.Platform) {
			platformSupportCount++
		}
	}
	
	stats.Summary.AvailableProviders = availableCount
	if stats.Summary.TotalProviders > 0 {
		stats.Summary.PlatformSupport = (platformSupportCount * 100) / stats.Summary.TotalProviders
	}

	// Collect action statistics
	actionStats := collectActionStats(providerStats)
	stats.Actions = actionStats
	stats.Summary.TotalActions = actionStats.TotalActions

	return stats, nil
}

// getProviderStats returns statistics for providers (only available ones by default, all in verbose mode)
func getProviderStats(actionManager interfaces.ActionManager) []ProviderStats {
	providerManager := actionManager.GetProviderManager()
	
	// Get all providers and available providers
	allProviders := providerManager.GetAllProviders()
	availableProviders := providerManager.GetAvailableProviders()
	
	// Create a map of available providers for quick lookup
	availableMap := make(map[string]bool)
	for _, provider := range availableProviders {
		availableMap[provider.Provider.Name] = true
	}
	
	var stats []ProviderStats
	
	// Process all providers to get complete information
	for _, provider := range allProviders {
		available := availableMap[provider.Provider.Name]
		
		// Get action names
		var actionNames []string
		for actionName := range provider.Actions {
			actionNames = append(actionNames, actionName)
		}
		sort.Strings(actionNames) // Sort for consistent output
		
		// Determine status and error message
		status := "not available"
		errorMsg := ""
		if available {
			status = "available"
		} else {
			// Fallback error messages based on provider configuration
			if provider.Provider.Executable != "" {
				errorMsg = fmt.Sprintf("executable '%s' not found", provider.Provider.Executable)
			} else if len(provider.Provider.Platforms) > 0 {
				errorMsg = fmt.Sprintf("not available on this platform (supports: %v)", provider.Provider.Platforms)
			} else {
				errorMsg = "not available on this platform"
			}
		}
		
		// Get display name
		displayName := provider.Provider.DisplayName
		if displayName == "" {
			displayName = strings.Title(strings.ReplaceAll(provider.Provider.Name, "_", " "))
		}
		
		stat := ProviderStats{
			Name:         provider.Provider.Name,
			DisplayName:  displayName,
			Type:         provider.Provider.Type,
			Available:    available,
			Priority:     provider.Provider.Priority,
			Platforms:    provider.Provider.Platforms,
			Capabilities: provider.Provider.Capabilities,
			Actions:      actionNames,
			Executable:   provider.Provider.Executable,
			Status:       status,
			Error:        errorMsg,
		}
		
		stats = append(stats, stat)
	}
	
	// Sort stats by availability first, then by name
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Available != stats[j].Available {
			return stats[i].Available // Available providers first
		}
		return stats[i].Name < stats[j].Name
	})
	
	return stats
}

// collectActionStats collects statistics about available actions
func collectActionStats(providers []ProviderStats) ActionStats {
	actionSet := make(map[string]bool)
	actionProviders := make(map[string][]string)
	coverage := make(map[string]int)
	
	// Categories for organizing actions
	categories := map[string][]string{
		"Software Management": {"install", "uninstall", "upgrade", "search", "info", "version"},
		"Service Management":  {"start", "stop", "restart", "enable", "disable", "status"},
		"Monitoring":          {"logs", "cpu", "memory", "io", "check"},
		"System":              {"list", "stats", "apply"},
	}

	// Collect all actions and their providers
	for _, provider := range providers {
		if !provider.Available {
			continue
		}
		
		for _, action := range provider.Actions {
			actionSet[action] = true
			actionProviders[action] = append(actionProviders[action], provider.Name)
			coverage[action]++
		}
	}

	// Organize actions by category
	actionsByCategory := make(map[string][]string)
	for category, categoryActions := range categories {
		var availableActions []string
		for _, action := range categoryActions {
			if actionSet[action] {
				availableActions = append(availableActions, action)
			}
		}
		if len(availableActions) > 0 {
			sort.Strings(availableActions)
			actionsByCategory[category] = availableActions
		}
	}

	// Add uncategorized actions
	var uncategorized []string
	for action := range actionSet {
		categorized := false
		for _, categoryActions := range categories {
			for _, categoryAction := range categoryActions {
				if action == categoryAction {
					categorized = true
					break
				}
			}
			if categorized {
				break
			}
		}
		if !categorized {
			uncategorized = append(uncategorized, action)
		}
	}
	if len(uncategorized) > 0 {
		sort.Strings(uncategorized)
		actionsByCategory["Other"] = uncategorized
	}

	return ActionStats{
		TotalActions:      len(actionSet),
		ActionsByCategory: actionsByCategory,
		ActionProviders:   actionProviders,
		Coverage:          coverage,
	}
}

// displayStats displays the statistics in a formatted way
func displayStats(stats *SystemStats, formatter *output.OutputFormatter, userInterface *ui.UserInterface, verbose bool) {
	// System Information
	fmt.Println("System Information:")
	fmt.Printf("  Platform: %s\n", stats.System.Platform)
	fmt.Printf("  OS: %s\n", stats.System.OS)
	fmt.Printf("  Architecture: %s\n", stats.System.Architecture)
	if stats.System.Version != "" {
		fmt.Printf("  Version: %s\n", stats.System.Version)
	}
	fmt.Printf("  SAI Version: %s\n", stats.System.SAIVersion)
	fmt.Println()

	// Provider Statistics
	fmt.Println("Provider Statistics:")
	fmt.Printf("  Total Providers: %d\n", stats.Summary.TotalProviders)
	fmt.Printf("  Available Providers: %d\n", stats.Summary.AvailableProviders)
	fmt.Printf("  Platform Support: %d%%\n", stats.Summary.PlatformSupport)
	fmt.Println()

	// Provider Details
	if verbose {
		fmt.Println("Provider Details:")
		
		// Group providers by type
		providersByType := make(map[string][]ProviderStats)
		for _, provider := range stats.Providers {
			providersByType[provider.Type] = append(providersByType[provider.Type], provider)
		}

		for providerType, providers := range providersByType {
			fmt.Printf("\n  %s:\n", strings.Title(strings.ReplaceAll(providerType, "_", " ")))
			
			headers := []string{"Name", "Status", "Priority", "Actions", "Platforms", "Executable"}
			var rows [][]string
			
			for _, provider := range providers {
				status := provider.Status
				if !provider.Available && provider.Error != "" {
					status = fmt.Sprintf("%s (%s)", status, provider.Error)
				}
				
				actionsStr := fmt.Sprintf("%d actions", len(provider.Actions))
				if len(provider.Actions) <= 3 && len(provider.Actions) > 0 {
					actionsStr = strings.Join(provider.Actions, ", ")
				}
				
				platformsStr := strings.Join(provider.Platforms, ", ")
				if len(platformsStr) > 20 {
					platformsStr = platformsStr[:17] + "..."
				}
				
				executable := provider.Executable
				if executable == "" {
					executable = "(none)"
				}
				
				rows = append(rows, []string{
					provider.DisplayName,
					status,
					fmt.Sprintf("%d", provider.Priority),
					actionsStr,
					platformsStr,
					executable,
				})
			}
			
			userInterface.ShowTable(headers, rows)
		}
	} else {
		// Simple provider list - ONLY show available providers (Requirement 13.1)
		fmt.Println("Available Providers:")
		availableCount := 0
		for _, provider := range stats.Providers {
			if provider.Available {
				fmt.Printf("  %s (%s) - %d actions\n", 
					formatter.FormatProviderName(provider.Name), 
					provider.Type, 
					len(provider.Actions))
				availableCount++
			}
		}
		
		if availableCount == 0 {
			fmt.Println("  No providers are currently available on this system.")
			fmt.Println("  Use --verbose to see unavailable providers and reasons.")
		}
		
		// Only show unavailable providers in verbose mode
		if verbose {
			fmt.Println("\nUnavailable Providers:")
			unavailableCount := 0
			for _, provider := range stats.Providers {
				if !provider.Available {
					reason := "not available"
					if provider.Error != "" {
						reason = provider.Error
					}
					fmt.Printf("  %s (%s) - %s\n", provider.Name, provider.Type, reason)
					unavailableCount++
				}
			}
			
			if unavailableCount == 0 {
				fmt.Println("  All providers are available on this system.")
			}
		}
	}

	fmt.Println()

	// Action Statistics
	fmt.Println("Action Statistics:")
	fmt.Printf("  Total Actions: %d\n", stats.Actions.TotalActions)
	fmt.Println()

	if verbose {
		fmt.Println("Actions by Category:")
		for category, actions := range stats.Actions.ActionsByCategory {
			fmt.Printf("  %s: %s\n", category, strings.Join(actions, ", "))
		}
		fmt.Println()

		fmt.Println("Action Coverage (providers supporting each action):")
		var sortedActions []string
		for action := range stats.Actions.Coverage {
			sortedActions = append(sortedActions, action)
		}
		sort.Strings(sortedActions)

		headers := []string{"Action", "Providers", "Coverage"}
		var rows [][]string
		
		for _, action := range sortedActions {
			providers := stats.Actions.ActionProviders[action]
			coverage := stats.Actions.Coverage[action]
			
			providersStr := strings.Join(providers, ", ")
			if len(providersStr) > 30 {
				providersStr = providersStr[:27] + "..."
			}
			
			rows = append(rows, []string{
				action,
				providersStr,
				fmt.Sprintf("%d", coverage),
			})
		}
		
		userInterface.ShowTable(headers, rows)
	} else {
		fmt.Println("Available Actions:")
		for category, actions := range stats.Actions.ActionsByCategory {
			fmt.Printf("  %s: %s\n", category, strings.Join(actions, ", "))
		}
	}
}

// Helper functions

func getOSInfo() string {
	switch runtime.GOOS {
	case "linux":
		return "Linux"
	case "darwin":
		return "macOS"
	case "windows":
		return "Windows"
	default:
		return runtime.GOOS
	}
}

func getOSVersion() string {
	// This is a placeholder - in a real implementation, this would detect the actual OS version
	switch runtime.GOOS {
	case "linux":
		return "Ubuntu 22.04" // Placeholder
	case "darwin":
		return "macOS 13.0" // Placeholder
	case "windows":
		return "Windows 11" // Placeholder
	default:
		return "Unknown"
	}
}

func supportsPlatform(platforms []string, currentPlatform string) bool {
	// Map Go runtime platform names to our provider platform names
	platformMap := map[string]string{
		"darwin":  "macos",
		"linux":   "linux", 
		"windows": "windows",
	}
	
	// Get the provider platform name for the current system platform
	providerPlatform := platformMap[currentPlatform]
	if providerPlatform == "" {
		providerPlatform = currentPlatform // fallback to original name
	}
	
	for _, platform := range platforms {
		if platform == currentPlatform || platform == providerPlatform {
			return true
		}
	}
	return false
}

func getProviderStatus(executable string, available bool) string {
	if available {
		return "available"
	}
	return fmt.Sprintf("not available (%s not found)", executable)
}