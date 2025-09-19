package cli

import (
	"github.com/spf13/cobra"
)

// memoryCmd represents the memory command
var memoryCmd = &cobra.Command{
	Use:   "memory [software]",
	Short: "Display memory usage statistics",
	Long: `Display memory usage statistics for the specified software or general system memory usage if no software is specified.
This command will show memory usage using appropriate monitoring tools (ps, free, htop, etc.).

This is an information-only command that executes without confirmation prompts.
The output includes memory usage, RSS, VSZ, and other relevant memory metrics.

Examples:
  sai memory nginx                     # Show memory usage for nginx processes
  sai memory nginx --verbose           # Show detailed memory information
  sai memory nginx --json              # Output memory stats in JSON format
  sai memory                          # Show general system memory usage
  sai memory --provider htop          # Use specific provider for memory monitoring`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return executeGeneralSystemCommand("memory")
		}
		return executeServiceCommand("memory", args[0])
	},
}

func init() {
	rootCmd.AddCommand(memoryCmd)
}