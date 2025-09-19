package cli

import (
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [software]",
	Short: "Start software service",
	Long: `Start the service for the specified software.
This command will start the service using the appropriate service manager (systemd, launchd, etc.).

The system will validate that the service exists before attempting to start it.
Use --dry-run to see what commands would be executed without starting the service.

Examples:
  sai start nginx                      # Start nginx service
  sai start nginx --dry-run            # Show what would be executed without starting
  sai start nginx --yes                # Start nginx without confirmation prompt
  sai start nginx --provider systemd   # Use specific provider for service management`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeServiceCommand("start", args[0])
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}