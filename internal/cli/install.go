package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [software]",
	Short: "Install software packages",
	Long: `Install software packages using the appropriate provider.
The system will detect the best provider automatically or use the one specified with --provider.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		software := args[0]
		fmt.Printf("Installing %s...\n", software)
		// TODO: Implement installation logic
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}