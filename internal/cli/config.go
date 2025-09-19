package cli

import (
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config [software]",
	Short: "Show software configuration files",
	Long: `Show configuration files for the specified software.
This command will display the location and contents of configuration files.

This is an information-only command that executes without confirmation prompts.
The output includes configuration file paths, contents, and validation status.

Examples:
  sai config nginx                     # Show nginx configuration files
  sai config nginx --verbose           # Show detailed configuration information
  sai config nginx --json              # Output configuration in JSON format
  sai config nginx --provider systemd  # Use specific provider for configuration management`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeServiceCommand("config", args[0])
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}