package providers

import (
	"fmt"
	"os/exec"
	containerprovider "sai/cmd/providers/container"
)

// ContainerProvider interface for all container providers
type ContainerProvider interface {
	Provider
	GetContainerTool() string
}

// BaseContainerProvider common functionality for container providers
type BaseContainerProvider struct {
	Name string
}

// GetContainerTool returns the container orchestration tool name
func (p *BaseContainerProvider) GetContainerTool() string {
	return p.Name
}

// HelmProvider handles Helm chart operations
type HelmProvider struct {
	BaseContainerProvider
}

// Execute runs Helm commands
func (p *HelmProvider) Execute(action, resource string) error {
	fmt.Printf("Executing %s %s with Helm provider\n", action, resource)

	var cmd *exec.Cmd
	switch action {
	case "install":
		cmd = exec.Command("helm", "install", resource)
	case "uninstall":
		cmd = exec.Command("helm", "uninstall", resource)
	case "status":
		cmd = exec.Command("helm", "status", resource)
	case "upgrade":
		cmd = exec.Command("helm", "upgrade", resource)
	case "list":
		cmd = exec.Command("helm", "list")
	case "search":
		cmd = exec.Command("helm", "search", "repo", resource)
	default:
		fmt.Printf("Action %s not implemented for Helm\n", action)
		return nil
	}

	// Would actually run the command in a real implementation
	fmt.Printf("Would run: %s\n", cmd.String())
	return nil
}

// KubectlProvider handles Kubernetes operations
type KubectlProvider struct {
	BaseContainerProvider
	Namespace string
}

// Execute runs Kubectl commands
func (p *KubectlProvider) Execute(action, resource string) error {
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

// containerProviderAdapter adapts a container.Provider to providers.Provider
type containerProviderAdapter struct {
	provider containerprovider.Provider
}

// Execute implements the Provider interface
func (a *containerProviderAdapter) Execute(action, resource string) error {
	return a.provider.Execute(action, resource)
}

// NewContainerProvider creates the appropriate container provider based on the name
func NewContainerProvider(name string) Provider {
	provider := containerprovider.NewProvider(name)
	return &containerProviderAdapter{provider: provider}
}
