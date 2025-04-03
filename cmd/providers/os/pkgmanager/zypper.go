package pkgmanager

import (
	"fmt"
	"os/exec"
)

// ZypperProvider handles SUSE package operations
type ZypperProvider struct {
	BaseProvider
}

// Execute runs Zypper commands
func (p *ZypperProvider) Execute(action, software string) error {
	// Validate action
	if !IsValidAction(action) {
		return fmt.Errorf("unsupported action '%s' for Zypper provider", action)
	}

	fmt.Printf("Executing %s %s with Zypper provider\n", action, software)

	var cmd *exec.Cmd
	switch action {
	case ActionInstall:
		cmd = exec.Command("zypper", "install", "-y", software)
	case ActionUninstall:
		cmd = exec.Command("zypper", "remove", "-y", software)
	case ActionStatus:
		cmd = exec.Command("zypper", "info", software)
	case ActionList:
		cmd = exec.Command("zypper", "packages", "--installed-only")
	case ActionSearch:
		cmd = exec.Command("zypper", "search", software)
	case ActionUpgrade:
		cmd = exec.Command("zypper", "update", "-y", software)
	case ActionInfo:
		cmd = exec.Command("zypper", "info", software)
	default:
		fmt.Printf("Action %s not implemented for Zypper\n", action)
		return nil
	}

	// Would actually run the command in a real implementation
	fmt.Printf("Would run: %s\n", cmd.String())
	return nil
}

// NewZypperProvider creates a new Zypper provider
func NewZypperProvider() *ZypperProvider {
	return &ZypperProvider{
		BaseProvider: BaseProvider{Name: "zypper"},
	}
}
