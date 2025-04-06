package cmd

import (
	"fmt"
	"os"
	"strings"

	"sai/cmd/handlers"

	"github.com/spf13/cobra"
)

var (
	providerFlag string
	dryRunFlag   bool
	yesFlag      bool
	forceFlag    bool
)

// SupportedCommands map of all supported commands
var SupportedCommands = map[string]handlers.CommandHandler{
	"apply":        func(software string, provider string) { handlers.NewApplyHandler().Handle(software, provider) },
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
	"reload":       func(software string, provider string) { handlers.NewReloadHandler().Handle(software, provider) },
	"enable":       func(software string, provider string) { handlers.NewEnableHandler().Handle(software, provider) },
	"disable":      func(software string, provider string) { handlers.NewDisableHandler().Handle(software, provider) },
	"list":         func(software string, provider string) { handlers.NewListHandler().Handle(software, provider) },
	"search":       func(software string, provider string) { handlers.NewSearchHandler().Handle(software, provider) },
	"update":       func(software string, provider string) { handlers.NewUpdateHandler().Handle(software, provider) },
	"ask":          func(software string, provider string) { handlers.NewAskHandler().Handle(software, provider) },
	"help":         func(software string, provider string) { handlers.NewHelpHandler().Handle(software, provider) },
	"inspect":      func(software string, provider string) { handlers.NewInspectHandler().Handle(software, provider) },
}

// Commands that can be used without arguments
var NoArgCommands = map[string]bool{
	"status":  true,
	"help":    true,
	"apply":   true,
	"install": true,
	"ask":     true,
	"search":  true,
}

var rootCmd = &cobra.Command{
	Use:   "sai",
	Short: "SAI whatever on every software everywhere",
	Long: `SAI is a tool that lets you manage software components via a consistent command interface.
Usage:
  sai <action> <software> [flags]
  sai <action> [flags]  # For commands that don't require software argument
Example:
  sai install nginx
  sai status redis
  sai reload nginx
  sai status   # Shows status of all services
  sai help     # Shows help information
  sai apply    # Applies configuration from sai.yaml
  sai install  # Lists available installation options and providers
  sai inspect nginx  # Inspects the nginx package or service
  sai ask      # Start a chat with an LLM to get help
  sai search   # Search for software and get recommendations`,
}

// handleCommand processes commands in the format: sai <action> <software>
func handleCommand(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("requires at least 1 arg")
	}

	action := strings.ToLower(args[0])

	// Set the dry run mode, yes mode, and force mode in the handlers package
	handlers.SetDryRun(dryRunFlag)
	handlers.SetYes(yesFlag)
	handlers.SetForce(forceFlag)

	if handler, ok := SupportedCommands[action]; ok {
		// Debug output
		fmt.Printf("Command: %s, Is NoArgCommand: %v\n", action, NoArgCommands[action])

		// Check if this is a no-arg command
		if NoArgCommands[action] {
			// For no-arg commands, pass empty string as software
			handler("", providerFlag)
			return nil
		}

		// For commands that require software argument
		if len(args) < 2 {
			return fmt.Errorf("command '%s' requires a software argument", action)
		}
		software := args[1]
		handler(software, providerFlag)
		return nil
	}

	return fmt.Errorf("unsupported action: %s", action)
}

func Execute() {
	// Enable positional arguments with flags
	cobra.EnableCommandSorting = false

	// Add a run handler for the root command
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Please specify an action")
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
	rootCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically answer yes to all prompts")
	rootCmd.PersistentFlags().BoolVarP(&forceFlag, "force", "f", false, "Force the operation, bypassing safety checks")

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
