package cmd

import (
	"fmt"
	"os"
	"strings"

	"sai/cmd/handlers"

	"github.com/spf13/cobra"
)

var providerFlag string
var dryRunFlag bool

// SupportedCommands map of all supported commands
var SupportedCommands = map[string]handlers.CommandHandler{
	"install":      func(software string, provider string) { handlers.NewInstallHandler().Handle(software, provider) },
	"test":         func(software string, provider string) { handlers.NewTestHandler().Handle(software, provider) },
	"build":        func(software string, provider string) { handlers.NewBuildHandler().Handle(software, provider) },
	"log":          func(software string, provider string) { handlers.NewLogHandler().Handle(software, provider) },
	"check":        func(software string, provider string) { handlers.NewCheckHandler().Handle(software, provider) },
	"observe":      func(software string, provider string) { handlers.NewObserveHandler().Handle(software, provider) },
	"trace":        func(software string, provider string) { handlers.NewTraceHandler().Handle(software, provider) },
	"config":       func(software string, provider string) { handlers.NewConfigHandler().Handle(software, provider) },
	"info":         func(software string, provider string) { handlers.NewInfoHandler().Handle(software, provider) },
	"debug":        func(software string, provider string) { handlers.NewDebugHandler().Handle(software, provider) },
	"troubleshoot": func(software string, provider string) { handlers.NewTroubleshootHandler().Handle(software, provider) },
	"monitor":      func(software string, provider string) { handlers.NewMonitorHandler().Handle(software, provider) },
	"upgrade":      func(software string, provider string) { handlers.NewUpgradeHandler().Handle(software, provider) },
	"uninstall":    func(software string, provider string) { handlers.NewUninstallHandler().Handle(software, provider) },
	"status":       func(software string, provider string) { handlers.NewStatusHandler().Handle(software, provider) },
	"start":        func(software string, provider string) { handlers.NewStartHandler().Handle(software, provider) },
	"stop":         func(software string, provider string) { handlers.NewStopHandler().Handle(software, provider) },
	"restart":      func(software string, provider string) { handlers.NewRestartHandler().Handle(software, provider) },
	"enable":       func(software string, provider string) { handlers.NewEnableHandler().Handle(software, provider) },
	"disable":      func(software string, provider string) { handlers.NewDisableHandler().Handle(software, provider) },
	"list":         func(software string, provider string) { handlers.NewListHandler().Handle(software, provider) },
	"search":       func(software string, provider string) { handlers.NewSearchHandler().Handle(software, provider) },
	"update":       func(software string, provider string) { handlers.NewUpdateHandler().Handle(software, provider) },
	"ask":          func(software string, provider string) { handlers.NewAskHandler().Handle(software, provider) },
	"help":         func(software string, provider string) { handlers.NewHelpHandler().Handle(software, provider) },
}

var rootCmd = &cobra.Command{
	Use:   "sai",
	Short: "SAI is a smart software command runner",
	Long: `SAI is a CLI tool that lets you manage software components via a consistent command interface.
Usage:
  sai <software> <command> [flags]
Example:
  sai nginx install
  sai redis status
  sai ec2 start --provider aws`,
}

// softwareCmd is dynamically created for each software
func createSoftwareCmd(software string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   software,
		Short: fmt.Sprintf("Commands for %s", software),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Printf("Please specify a command for %s\n", software)
				_ = cmd.Usage()
				os.Exit(1)
			}
		},
	}

	// Add command subcommands to the software command
	for cmdName, handler := range SupportedCommands {
		actionCmd := &cobra.Command{
			Use:   cmdName,
			Short: fmt.Sprintf("%s %s", cmdName, software),
			Run: func(cmdName string, handler handlers.CommandHandler) func(*cobra.Command, []string) {
				return func(cmd *cobra.Command, args []string) {
					// Set the dry run mode in the handlers package
					handlers.SetDryRun(dryRunFlag)
					handler(software, providerFlag)
				}
			}(cmdName, handler),
		}

		// Add the provider and dry-run flags to each command
		actionCmd.Flags().StringVar(&providerFlag, "provider", "", "Specify a provider to use for this command")
		actionCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Show what commands would be executed without running them")

		cmd.AddCommand(actionCmd)
	}

	return cmd
}

// handleCommand processes commands in the format: sai <software> <command>
func handleCommand(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("requires at least 2 args")
	}

	software := args[0]
	command := args[1]

	// Set the dry run mode in the handlers package
	handlers.SetDryRun(dryRunFlag)

	if handler, ok := SupportedCommands[strings.ToLower(command)]; ok {
		handler(software, providerFlag)
		return nil
	}

	return fmt.Errorf("unsupported command: %s", command)
}

func Execute() {
	// Enable positional arguments with flags
	cobra.EnableCommandSorting = false

	// Add a run handler for the root command that supports the old format
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Please specify a software and command")
			_ = cmd.Usage()
			os.Exit(1)
		}

		if err := handleCommand(cmd, args); err != nil {
			fmt.Println(err)
			_ = cmd.Usage()
			os.Exit(1)
		}
	}

	// Add global flags to the root command
	rootCmd.PersistentFlags().StringVar(&providerFlag, "provider", "", "Specify a provider to use for this command")
	rootCmd.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "Show what commands would be executed without running them")

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
