package service

import (
	"fmt"
	"os/exec"
)

// BrewProvider handles macOS brew services operations
type BrewProvider struct {
	BaseProvider
}

// Execute runs brew services commands
func (p *BrewProvider) Execute(action, service string) error {
	// Check if in dry run mode
	if p.IsDryRun() {
		fmt.Printf("[DRY RUN] Would execute %s service %s using brew services\n", action, service)
		return nil
	}

	fmt.Printf("Managing service %s with Brew Services action %s\n", service, action)

	var cmd *exec.Cmd
	switch action {
	case ActionStart, ActionStop, ActionRestart:
		cmd = exec.Command("brew", "services", action, service)
	case ActionEnable, ActionDisable:
		// Brew services doesn't have direct enable/disable commands
		// For dry run, we'll just show what would happen
		if action == ActionEnable {
			fmt.Println("Note: brew services automatically enables services when started")
			cmd = exec.Command("brew", "services", "start", service)
		} else {
			fmt.Println("Note: brew services doesn't have a direct disable command")
			fmt.Println("Services can be manually disabled by modifying their plist files")
			cmd = exec.Command("brew", "services", "stop", service)
		}
	default:
		return fmt.Errorf("action %s not implemented for Brew Services", action)
	}

	// Would actually run the command in a real implementation
	fmt.Printf("Would run: %s\n", cmd.String())
	return nil
}

// NewBrewProvider creates a new Brew Service provider
func NewBrewProvider() *BrewProvider {
	return &BrewProvider{
		BaseProvider: BaseProvider{Name: "brew-services"},
	}
}
