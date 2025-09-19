package cli

import (
	"github.com/spf13/cobra"
)

// cpuCmd represents the cpu command
var cpuCmd = &cobra.Command{
	Use:   "cpu [software]",
	Short: "Display CPU usage statistics",
	Long: `Display CPU usage statistics for the specified software or general system CPU usage if no software is specified.
This command will show CPU usage using appropriate monitoring tools (ps, top, htop, etc.).

This is an information-only command that executes without confirmation prompts.
The output includes CPU percentage, process information, and other relevant metrics.

Examples:
  sai cpu nginx                        # Show CPU usage for nginx processes
  sai cpu nginx --verbose              # Show detailed CPU information
  sai cpu nginx --json                 # Output CPU stats in JSON format
  sai cpu                             # Show general system CPU usage
  sai cpu --provider htop             # Use specific provider for CPU monitoring`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return executeGeneralSystemCommand("cpu")
		}
		return executeServiceCommand("cpu", args[0])
	},
}

func init() {
	rootCmd.AddCommand(cpuCmd)
}