package handlers

import (
	"fmt"
	"runtime"
)

// DebugHandler handles the debug command
type DebugHandler struct {
	BaseHandler
}

// NewDebugHandler creates a new debug handler
func NewDebugHandler() *DebugHandler {
	return &DebugHandler{
		BaseHandler: BaseHandler{
			Action: "debug",
		},
	}
}

// Handle executes the debug command
func (h *DebugHandler) Handle(software string, provider string) {
	fmt.Printf("Debug information for %s:\n", software)
	fmt.Println("-----------------------------------")

	// Display system information
	fmt.Println("System information:")
	fmt.Printf("  OS: %s\n", runtime.GOOS)
	fmt.Printf("  Architecture: %s\n", runtime.GOARCH)
	fmt.Printf("  Go version: %s\n", runtime.Version())

	// Display software specific debug info
	fmt.Println("\nSoftware debug info:")
	if provider != "" {
		fmt.Printf("  Using provider: %s\n", provider)
	} else {
		fmt.Printf("  Using default provider\n")
	}

	fmt.Println("\nDiagnostic commands that would be run:")
	switch software {
	case "nginx":
		fmt.Println("  - nginx -t (Test configuration)")
		fmt.Println("  - nginx -V (Version and build information)")
	case "redis":
		fmt.Println("  - redis-cli info (Server information)")
		fmt.Println("  - redis-cli ping (Connection test)")
	default:
		fmt.Println("  - which " + software + " (Binary location)")
		fmt.Println("  - " + software + " --version (Version information)")
	}
}
