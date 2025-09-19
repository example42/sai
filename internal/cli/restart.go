package cli

import (
	"github.com/spf13/cobra"
)

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
	Use:   "restart [software]",
	Short: "Restart software service",
	Long: `Restart the service for the specified software.
This command will restart the service using the appropriate service manager (systemd, launchd, etc.).

The system will validate that the service exists before attempting to restart it.
Use --dry-run to see what commands would be executed without restarting the service.

Examples:
  sai restart nginx                    # Restart nginx service
  sai restart nginx --dry-run          # Show what would be executed without restarting
  sai restart nginx --yes              # Restart nginx without confirmation prompt
  sai restart nginx --provider systemd # Use specific provider for service management`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeServiceCommand("restart", args[0])
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}