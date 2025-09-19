package cli

import (
	"github.com/spf13/cobra"
)

// ioCmd represents the io command
var ioCmd = &cobra.Command{
	Use:   "io [software]",
	Short: "Display I/O statistics",
	Long: `Display I/O statistics for the specified software or general system I/O usage if no software is specified.
This command will show I/O usage using appropriate monitoring tools (iotop, iostat, etc.).

This is an information-only command that executes without confirmation prompts.
The output includes read/write operations, disk usage, and other relevant I/O metrics.

Examples:
  sai io nginx                         # Show I/O usage for nginx processes
  sai io nginx --verbose               # Show detailed I/O information
  sai io nginx --json                  # Output I/O stats in JSON format
  sai io                              # Show general system I/O usage
  sai io --provider iotop             # Use specific provider for I/O monitoring`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return executeGeneralSystemCommand("io")
		}
		return executeServiceCommand("io", args[0])
	},
}

func init() {
	rootCmd.AddCommand(ioCmd)
}