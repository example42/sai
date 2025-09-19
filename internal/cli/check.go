package cli

import (
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check [software]",
	Short: "Verify if software is working correctly",
	Long: `Verify if the specified software is working correctly by running health checks.
This command will perform various checks to ensure the software is functioning properly.

This is an information-only command that executes without confirmation prompts.
The checks may include service status, configuration validation, connectivity tests, etc.

Examples:
  sai check nginx                      # Check if nginx is working correctly
  sai check nginx --verbose            # Show detailed check information
  sai check nginx --json               # Output check results in JSON format
  sai check nginx --provider systemd   # Use specific provider for health checks`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeServiceCommand("check", args[0])
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}