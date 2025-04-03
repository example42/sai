package service

import (
	"fmt"
	"os/exec"
)

// SystemdProvider handles Linux systemd service operations
type SystemdProvider struct {
	BaseProvider
}

// Execute runs systemd commands
func (p *SystemdProvider) Execute(action, service string) error {
	// Validate action
	if !IsValidAction(action) {
		return fmt.Errorf("unsupported action '%s' for Systemd provider", action)
	}

	// Check if in dry run mode
	if p.IsDryRun() {
		fmt.Printf("[DRY RUN] Would execute %s service %s using systemd\n", action, service)
		return nil
	}

	fmt.Printf("Executing %s service %s using systemd\n", action, service)

	var cmd *exec.Cmd
	switch action {
	case ActionStart:
		cmd = exec.Command("systemctl", "start", service)
	case ActionStop:
		cmd = exec.Command("systemctl", "stop", service)
	case ActionRestart:
		cmd = exec.Command("systemctl", "restart", service)
	case ActionEnable:
		cmd = exec.Command("systemctl", "enable", service)
	case ActionDisable:
		cmd = exec.Command("systemctl", "disable", service)
	default:
		fmt.Printf("Action %s not implemented for systemd\n", action)
		return nil
	}

	// Would actually run the command in a real implementation
	fmt.Printf("Would run: %s\n", cmd.String())
	return nil
}

// NewSystemdProvider creates a new Systemd provider
func NewSystemdProvider() *SystemdProvider {
	return &SystemdProvider{
		BaseProvider: BaseProvider{Name: "systemd"},
	}
}
