package handlers

import (
	"fmt"
	"os/exec"
)

const (
	ProviderDocker = "docker"
)

// InspectHandler handles the inspect action
type InspectHandler struct {
	BaseHandler
}

// NewInspectHandler creates a new InspectHandler
func NewInspectHandler() *InspectHandler {
	return &InspectHandler{
		BaseHandler: BaseHandler{
			Action: "inspect",
		},
	}
}

// Handle executes the inspect action
func (h *InspectHandler) Handle(software string, provider string) {
	if software == "" {
		fmt.Println("Error: No software specified for inspection")
		return
	}

	// Check if in dry run mode
	if IsDryRun() {
		fmt.Printf("[DRY RUN] Would inspect %s\n", software)
		return
	}

	fmt.Printf("Inspecting %s...\n", software)

	// Get provider details
	provider, providerType := h.GetProvider()

	// Execute the appropriate inspection command based on the provider
	switch providerType {
	case ProviderTypeOS:
		h.inspectOSPackage(software, provider)
	case ProviderTypeContainer:
		h.inspectContainer(software, provider)
	case ProviderTypeCloud:
		h.inspectCloudResource(software, provider)
	default:
		fmt.Printf("Unsupported provider type for inspection: %s\n", providerType)
	}
}

// inspectOSPackage inspects an OS package
func (h *InspectHandler) inspectOSPackage(software, provider string) {
	var cmd *exec.Cmd

	switch provider {
	case ProviderAPT:
		cmd = exec.Command("dpkg", "-L", software)
	case ProviderRPM:
		cmd = exec.Command("rpm", "-ql", software)
	case ProviderBrew:
		cmd = exec.Command("brew", "list", software)
	case ProviderPacman:
		cmd = exec.Command("pacman", "-Ql", software)
	default:
		fmt.Printf("Unsupported OS provider for inspection: %s\n", provider)
		return
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error inspecting package: %v\n", err)
		return
	}

	fmt.Println(string(output))
}

// inspectContainer inspects a container
func (h *InspectHandler) inspectContainer(software, provider string) {
	var cmd *exec.Cmd

	switch provider {
	case ProviderDocker:
		cmd = exec.Command("docker", "inspect", software)
	case ProviderKubectl:
		cmd = exec.Command("kubectl", "describe", "pod", software)
	default:
		fmt.Printf("Unsupported container provider for inspection: %s\n", provider)
		return
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error inspecting container: %v\n", err)
		return
	}

	fmt.Println(string(output))
}

// inspectCloudResource inspects a cloud resource
func (h *InspectHandler) inspectCloudResource(software, provider string) {
	var cmd *exec.Cmd

	switch provider {
	case ProviderAWS:
		cmd = exec.Command("aws", "ec2", "describe-instances", "--instance-ids", software)
	case ProviderAzure:
		cmd = exec.Command("az", "vm", "show", "--name", software, "--resource-group", "default")
	case ProviderGCP:
		cmd = exec.Command("gcloud", "compute", "instances", "describe", software)
	default:
		fmt.Printf("Unsupported cloud provider for inspection: %s\n", provider)
		return
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error inspecting cloud resource: %v\n", err)
		return
	}

	fmt.Println(string(output))
}
