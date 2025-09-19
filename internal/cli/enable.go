package cli

import (
	"github.com/spf13/cobra"
)

// enableCmd represents the enable command
var enableCmd = &cobra.Command{
	Use:   "enable [software]",
	Short: "Enable software service at boot",
	Long: `Enable the service for the specified software to start automatically at boot.
This command will enable the service using the appropriate service manager (systemd, launchd, etc.).

The system will validate that the service exists before attempting to enable it.
Use --dry-run to see what commands would be executed without enabling the service.

Examples:
  sai enable nginx                     # Enable nginx service at boot
  sai enable nginx --dry-run           # Show what would be executed without enabling
  sai enable nginx --yes               # Enable nginx without confirmation prompt
  sai enable nginx --provider systemd  # Use specific provider for service management`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeServiceCommand("enable", args[0])
	},
}

func init() {
	rootCmd.AddCommand(enableCmd)
}