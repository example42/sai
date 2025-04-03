package container

import (
	"fmt"
	"os/exec"
)

// KubectlProvider handles Kubernetes operations
type KubectlProvider struct {
	BaseContainerProvider
	Namespace string
}

// Execute runs Kubectl commands
func (p *KubectlProvider) Execute(action, resource string) error {
	// Validate action
	if !IsValidContainerAction(action) {
		return fmt.Errorf("unsupported action '%s' for Kubectl provider", action)
	}

	// Check if in dry run mode
	if p.IsDryRun() {
		fmt.Printf("[DRY RUN] Would execute %s %s with Kubectl provider\n", action, resource)
		return nil
	}

	fmt.Printf("Executing %s %s with Kubectl provider\n", action, resource)

	var cmd *exec.Cmd
	namespace := p.Namespace
	if namespace == "" {
		namespace = "default"
	}

	switch action {
	case "install", "create":
		cmd = exec.Command("kubectl", "apply", "-f", resource, "-n", namespace)
	case "uninstall", "delete":
		cmd = exec.Command("kubectl", "delete", "-f", resource, "-n", namespace)
	case "status", "describe":
		parts := splitResourceType(resource)
		if len(parts) == 2 {
			cmd = exec.Command("kubectl", "describe", parts[0], parts[1], "-n", namespace)
		} else {
			cmd = exec.Command("kubectl", "describe", resource, "-n", namespace)
		}
	case "start", "stop", "restart":
		parts := splitResourceType(resource)
		if len(parts) == 2 {
			cmd = exec.Command("kubectl", "rollout", "restart", parts[0], parts[1], "-n", namespace)
		} else {
			cmd = exec.Command("kubectl", "rollout", "restart", resource, "-n", namespace)
		}
	case "logs":
		cmd = exec.Command("kubectl", "logs", resource, "-n", namespace)
	case "list":
		cmd = exec.Command("kubectl", "get", resource, "-n", namespace)
	default:
		fmt.Printf("Action %s not implemented for Kubectl\n", action)
		return nil
	}

	// Would actually run the command in a real implementation
	fmt.Printf("Would run: %s\n", cmd.String())
	return nil
}

// splitResourceType splits "type/name" into [type, name]
func splitResourceType(resource string) []string {
	var parts []string
	for i := 0; i < len(resource); i++ {
		if resource[i] == '/' {
			return []string{resource[:i], resource[i+1:]}
		}
	}
	return parts
}

// NewKubectlProvider creates a new Kubectl provider
func NewKubectlProvider() *KubectlProvider {
	return &KubectlProvider{
		BaseContainerProvider: BaseContainerProvider{Name: "kubectl"},
		Namespace:             "default",
	}
}
