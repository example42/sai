package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	provider   string
	verbose    bool
	dryRun     bool
	yes        bool
	quiet      bool
	jsonOutput bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sai",
	Short: "SAI - Software Action Interface",
	Long: `SAI is a lightweight CLI tool for executing software management actions 
using provider-based configurations. The core philosophy is "Do everything on 
every software on every OS" through a unified interface.`,
	Version: "0.1.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.sai.yaml)")
	rootCmd.PersistentFlags().StringVarP(&provider, "provider", "p", "", "force specific provider")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be executed without running commands")
	rootCmd.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "automatically confirm all prompts")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output results in JSON format")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Import config package to ensure dependencies are included
	_ = "sai/internal/config"
	
	if verbose {
		fmt.Println("Verbose mode enabled")
	}
}