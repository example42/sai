package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sai/internal/config"
)

var (
	cfgFile    string
	provider   string
	verbose    bool
	dryRun     bool
	yes        bool
	quiet      bool
	jsonOutput bool
	
	// Global configuration instance
	globalConfig *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sai",
	Short: "SAI - Software Action Interface",
	Long: `SAI is a lightweight CLI tool for executing software management actions 
using provider-based configurations. The core philosophy is "Do everything on 
every software on every OS" through a unified interface.

SAI supports:
  • Software management: install, uninstall, upgrade, search, info, version
  • Service management: start, stop, restart, enable, disable, status
  • System monitoring: logs, cpu, memory, io, check
  • Batch operations: apply, stats, list

Examples:
  sai install nginx                    # Install nginx using best available provider
  sai install nginx --provider apt     # Install nginx using apt provider
  sai start nginx                      # Start nginx service
  sai status nginx                     # Check nginx service status
  sai list                            # List all installed software
  sai apply actions.yaml              # Execute batch operations`,
	Version: "0.1.0",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate flags first
		if err := ValidateFlags(); err != nil {
			return err
		}
		// Then initialize configuration
		return initializeConfig()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags with detailed help text
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", 
		"config file path (searches: ./sai.yaml, ~/.sai/config.yaml, /etc/sai/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&provider, "provider", "p", "", 
		"force specific provider (apt, brew, docker, etc.)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, 
		"enable detailed output and logging information")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, 
		"show what would be executed without running commands")
	rootCmd.PersistentFlags().BoolVarP(&yes, "yes", "y", false, 
		"automatically confirm all prompts (unattended mode)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, 
		"suppress non-essential output (minimal output mode)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, 
		"output results in JSON format for programmatic consumption")

	// Flag validation and mutual exclusivity
	rootCmd.MarkFlagsMutuallyExclusive("verbose", "quiet")
	rootCmd.MarkFlagsMutuallyExclusive("json", "quiet")

	// Set up command completion
	rootCmd.CompletionOptions.DisableDefaultCmd = false
	rootCmd.CompletionOptions.HiddenDefaultCmd = false
	SetupCompletion()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in standard locations
		viper.SetConfigName("sai")
		viper.SetConfigType("yaml")
		
		// Add search paths
		viper.AddConfigPath(".")
		viper.AddConfigPath("./.sai")
		
		if home, err := os.UserHomeDir(); err == nil {
			viper.AddConfigPath(home + "/.sai")
			viper.AddConfigPath(home + "/.config/sai")
		}
		
		viper.AddConfigPath("/etc/sai")
		viper.AddConfigPath("/usr/local/etc/sai")
	}

	// Environment variable support
	viper.SetEnvPrefix("SAI")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Read config file if available
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
}

// initializeConfig loads and validates the configuration
func initializeConfig() error {
	var err error
	globalConfig, err = config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Apply flag overrides to configuration
	applyFlagOverrides()

	// Set up logging based on configuration and flags
	logLevel := globalConfig.LogLevel
	if verbose {
		logLevel = "debug"
	} else if quiet {
		logLevel = "error"
	}
	config.SetupLogging(logLevel)

	return nil
}

// applyFlagOverrides applies command-line flag values to the global configuration
func applyFlagOverrides() {
	if provider != "" {
		globalConfig.DefaultProvider = provider
	}
	
	// Override confirmation settings based on --yes flag
	if yes {
		globalConfig.Confirmations.Install = false
		globalConfig.Confirmations.Uninstall = false
		globalConfig.Confirmations.Upgrade = false
		globalConfig.Confirmations.SystemChanges = false
		globalConfig.Confirmations.ServiceOps = false
	}
	
	// Override output settings based on flags
	if quiet {
		globalConfig.Output.ShowCommands = false
		globalConfig.Output.ShowExitCodes = false
	} else if verbose {
		globalConfig.Output.ShowCommands = true
		globalConfig.Output.ShowExitCodes = true
	}
}

// GetGlobalConfig returns the global configuration instance
func GetGlobalConfig() *config.Config {
	return globalConfig
}

// GetGlobalFlags returns the current global flag values
func GetGlobalFlags() GlobalFlags {
	return GlobalFlags{
		Config:     cfgFile,
		Provider:   provider,
		Verbose:    verbose,
		DryRun:     dryRun,
		Yes:        yes,
		Quiet:      quiet,
		JSONOutput: jsonOutput,
	}
}

// GlobalFlags represents the global command-line flags
type GlobalFlags struct {
	Config     string
	Provider   string
	Verbose    bool
	DryRun     bool
	Yes        bool
	Quiet      bool
	JSONOutput bool
}

// ValidateFlags performs validation on flag combinations and values
func ValidateFlags() error {
	// Validate provider name if specified
	if provider != "" {
		validProviders := []string{
			"apt", "brew", "dnf", "yum", "pacman", "zypper", "apk",
			"docker", "helm", "npm", "pip", "cargo", "go", "gem",
			"choco", "winget", "scoop", "flatpak", "snap",
		}
		
		isValid := false
		for _, validProvider := range validProviders {
			if provider == validProvider {
				isValid = true
				break
			}
		}
		
		if !isValid {
			return fmt.Errorf("invalid provider '%s'. Valid providers: %s", 
				provider, strings.Join(validProviders, ", "))
		}
	}

	// Validate config file exists if specified
	if cfgFile != "" {
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			return fmt.Errorf("config file '%s' does not exist", cfgFile)
		}
	}

	return nil
}

// SetupCompletion configures command completion for the CLI
func SetupCompletion() {
	// Provider name completion
	rootCmd.RegisterFlagCompletionFunc("provider", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		providers := []string{
			"apt\tDebian/Ubuntu package manager",
			"brew\tmacOS package manager", 
			"dnf\tFedora package manager",
			"yum\tRHEL/CentOS package manager",
			"pacman\tArch Linux package manager",
			"zypper\topenSUSE package manager",
			"apk\tAlpine Linux package manager",
			"docker\tDocker container manager",
			"helm\tKubernetes package manager",
			"npm\tNode.js package manager",
			"pip\tPython package manager",
			"cargo\tRust package manager",
			"go\tGo module manager",
			"gem\tRuby package manager",
			"choco\tWindows Chocolatey",
			"winget\tWindows Package Manager",
			"scoop\tWindows Scoop",
			"flatpak\tLinux application distribution",
			"snap\tUniversal Linux packages",
		}
		return providers, cobra.ShellCompDirectiveNoFileComp
	})

	// Config file completion
	rootCmd.RegisterFlagCompletionFunc("config", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"yaml", "yml"}, cobra.ShellCompDirectiveFilterFileExt
	})
}