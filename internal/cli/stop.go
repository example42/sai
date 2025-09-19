package cli

import (
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop [software]",
	Short: "Stop software service",
	Long: `Stop the service for the specified software.
This command will stop the service using the appropriate service manager (systemd, launchd, etc.).

The system will validate that the service exists and is running before attempting to stop it.
Use --dry-run to see what commands would be executed without stopping the service.

Examples:
  sai stop nginx                       # Stop nginx service
  sai stop nginx --dry-run             # Show what would be executed without stopping
  sai stop nginx --yes                 # Stop nginx without confirmation prompt
  sai stop nginx --provider systemd    # Use specific provider for service management`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeServiceCommand("stop", args[0])
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}