package cli

import (
	"github.com/spf13/cobra"
)

// disableCmd represents the disable command
var disableCmd = &cobra.Command{
	Use:   "disable [software]",
	Short: "Disable software service at boot",
	Long: `Disable the service for the specified software from starting automatically at boot.
This command will disable the service using the appropriate service manager (systemd, launchd, etc.).

The system will validate that the service exists before attempting to disable it.
Use --dry-run to see what commands would be executed without disabling the service.

Examples:
  sai disable nginx                    # Disable nginx service at boot
  sai disable nginx --dry-run          # Show what would be executed without disabling
  sai disable nginx --yes              # Disable nginx without confirmation prompt
  sai disable nginx --provider systemd # Use specific provider for service management`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeServiceCommand("disable", args[0])
	},
}

func init() {
	rootCmd.AddCommand(disableCmd)
}