package cli

import (
	"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs [software]",
	Short: "Display software service logs",
	Long: `Display logs for the specified software service or general system logs if no software is specified.
This command will show logs using the appropriate log management system (journalctl, log files, etc.).

This is an information-only command that executes without confirmation prompts.
Use flags to control log output format and filtering.

Examples:
  sai logs nginx                       # Show nginx service logs
  sai logs nginx --verbose             # Show detailed log information
  sai logs nginx --json                # Output logs in JSON format
  sai logs                            # Show general system service logs
  sai logs --provider systemd         # Use specific provider for log management`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return executeGeneralSystemCommand("logs")
		}
		return executeServiceCommand("logs", args[0])
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
}