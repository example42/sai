package handlers

import "fmt"

// HelpHandler handles the help command
type HelpHandler struct {
	BaseHandler
}

// NewHelpHandler creates a new help handler
func NewHelpHandler() *HelpHandler {
	return &HelpHandler{
		BaseHandler: BaseHandler{
			Action: "help",
		},
	}
}

// Handle executes the help command
func (h *HelpHandler) Handle(software string, provider string) {
	fmt.Println("SAI - Smart Software Management CLI")
	fmt.Println("-----------------------------------")
	fmt.Println("Available commands:")
	fmt.Println("  Package Management:")
	fmt.Println("    install    - Install software")
	fmt.Println("    uninstall  - Remove software")
	fmt.Println("    upgrade    - Upgrade software")
	fmt.Println("    status     - Check software status")
	fmt.Println("    list       - List installed software")
	fmt.Println("    search     - Search for software")
	fmt.Println("    info       - Show software information")
	fmt.Println("")
	fmt.Println("  Service Management:")
	fmt.Println("    start      - Start a service")
	fmt.Println("    stop       - Stop a service")
	fmt.Println("    restart    - Restart a service")
	fmt.Println("    enable     - Enable a service to start at boot")
	fmt.Println("    disable    - Disable a service at boot")
	fmt.Println("")
	fmt.Println("  Other Commands:")
	fmt.Println("    help       - Show this help message")
	fmt.Println("    config     - Configure settings")
	fmt.Println("    debug      - Debugging information")
	fmt.Println("")
	fmt.Println("Global Flags:")
	fmt.Println("    --provider - Specify a provider to use for the command")
	fmt.Println("    --dry-run  - Show what commands would be executed without running them")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  sai <software> <command>")
	fmt.Println("  sai <software> <command> [flags]  (recommended)")
	fmt.Println("  sai [flags] <software> <command>")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  sai nginx install")
	fmt.Println("  sai redis start")
	fmt.Println("  sai nginx install --provider apt")
	fmt.Println("  sai nginx install --dry-run")
	fmt.Println("  sai --provider apt nginx install")
}
