package container

import (
	"fmt"
	"os/exec"
)

// HelmProvider handles Helm chart operations
type HelmProvider struct {
	BaseContainerProvider
}

// Execute runs Helm commands
func (p *HelmProvider) Execute(action, resource string) error {
	// Validate action
	if !IsValidContainerAction(action) {
		return fmt.Errorf("unsupported action '%s' for Helm provider", action)
	}

	// Check if in dry run mode
	if p.IsDryRun() {
		fmt.Printf("[DRY RUN] Would execute %s %s with Helm provider\n", action, resource)
		return nil
	}

	fmt.Printf("Executing %s %s with Helm provider\n", action, resource)

	var cmd *exec.Cmd
	switch action {
	case ActionInstall:
		cmd = exec.Command("helm", "install", resource)
	case ActionUninstall:
		cmd = exec.Command("helm", "uninstall", resource)
	case ActionStatus:
		cmd = exec.Command("helm", "status", resource)
	case ActionUpgrade:
		cmd = exec.Command("helm", "upgrade", resource)
	case ActionList:
		cmd = exec.Command("helm", "list")
	case ActionSearch:
		cmd = exec.Command("helm", "search", "repo", resource)
	default:
		fmt.Printf("Action %s not implemented for Helm\n", action)
		return nil
	}

	// Would actually run the command in a real implementation
	fmt.Printf("Would run: %s\n", cmd.String())
	return nil
}

// NewHelmProvider creates a new Helm provider
func NewHelmProvider() *HelmProvider {
	return &HelmProvider{
		BaseContainerProvider: BaseContainerProvider{Name: "helm"},
	}
}
