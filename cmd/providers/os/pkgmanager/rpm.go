package pkgmanager

import (
	"fmt"
	"os/exec"
)

// RPMProvider handles RPM-based package operations
type RPMProvider struct {
	BaseProvider
}

// Execute runs RPM commands
func (p *RPMProvider) Execute(action, software string) error {
	// Validate action
	if !IsValidAction(action) {
		return fmt.Errorf("unsupported action '%s' for RPM provider", action)
	}

	fmt.Printf("Executing %s %s with RPM provider\n", action, software)

	var cmd *exec.Cmd
	switch action {
	case ActionInstall:
		cmd = exec.Command("rpm", "-i", software)
	case ActionUninstall:
		cmd = exec.Command("rpm", "-e", software)
	case ActionStatus:
		cmd = exec.Command("rpm", "-q", software)
	case ActionList:
		cmd = exec.Command("rpm", "-qa")
	case ActionSearch:
		cmd = exec.Command("rpm", "-qa", fmt.Sprintf("*%s*", software))
	case ActionInfo:
		cmd = exec.Command("rpm", "-qi", software)
	default:
		fmt.Printf("Action %s not implemented for RPM\n", action)
		return nil
	}

	// Would actually run the command in a real implementation
	fmt.Printf("Would run: %s\n", cmd.String())
	return nil
}

// NewRPMProvider creates a new RPM provider
func NewRPMProvider() *RPMProvider {
	return &RPMProvider{
		BaseProvider: BaseProvider{Name: "rpm"},
	}
}
