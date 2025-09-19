package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed software packages",
	Long: `List all installed software packages across all available providers.
This provides a comprehensive view of software installed on the system.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing installed software...")
		// TODO: Implement list logic
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}