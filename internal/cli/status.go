package cli

import (
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status [software]",
	Short: "Check software service status",
	Long: `Check the current status of the service for the specified software.
This command will display service status using the appropriate service manager (systemd, launchd, etc.).

This is an information-only command that executes without confirmation prompts.
The status includes whether the service is running, enabled, and other relevant information.

Examples:
  sai status nginx                     # Check nginx service status
  sai status nginx --json              # Output status in JSON format
  sai status nginx --provider systemd  # Use specific provider for service management
  sai status nginx --verbose           # Show detailed status information`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeServiceCommand("status", args[0])
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}