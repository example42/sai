package cloud

import (
	"fmt"
	"os/exec"
	"strings"
)

// GCPProvider handles Google Cloud Platform operations
type GCPProvider struct {
	BaseCloudProvider
	Project string
}

// Execute runs GCP CLI commands
func (p *GCPProvider) Execute(action, resource string) error {
	// Validate action
	if !IsValidCloudAction(action) {
		return fmt.Errorf("unsupported action '%s' for GCP provider", action)
	}

	// Check if in dry run mode
	if p.IsDryRun() {
		fmt.Printf("[DRY RUN] Would execute %s %s with GCP provider\n", action, resource)
		return nil
	}

	fmt.Printf("Executing %s %s with GCP provider\n", action, resource)

	// Set default region if not specified
	region := p.Region
	if region == "" {
		region = "us-central1" // Default GCP region
	}

	// Parse resource type and name
	resourceType, resourceName := parseCloudResource(resource)

	var cmd *exec.Cmd
	switch action {
	case "start":
		switch resourceType {
		case "compute":
			cmd = exec.Command("gcloud", "compute", "instances", "start", resourceName, "--zone", region)
		case "sql":
			cmd = exec.Command("gcloud", "sql", "instances", "start", resourceName)
		default:
			fmt.Printf("Resource type %s not supported for GCP start action\n", resourceType)
			return nil
		}
	case "stop":
		switch resourceType {
		case "compute":
			cmd = exec.Command("gcloud", "compute", "instances", "stop", resourceName, "--zone", region)
		case "sql":
			cmd = exec.Command("gcloud", "sql", "instances", "stop", resourceName)
		default:
			fmt.Printf("Resource type %s not supported for GCP stop action\n", resourceType)
			return nil
		}
	case "status":
		switch resourceType {
		case "compute":
			cmd = exec.Command("gcloud", "compute", "instances", "describe", resourceName, "--zone", region)
		case "sql":
			cmd = exec.Command("gcloud", "sql", "instances", "describe", resourceName)
		case "storage":
			cmd = exec.Command("gsutil", "ls", fmt.Sprintf("gs://%s", resourceName))
		default:
			fmt.Printf("Resource type %s not supported for GCP status action\n", resourceType)
			return nil
		}
	case "create":
		switch resourceType {
		case "compute":
			cmd = exec.Command("gcloud", "compute", "instances", "create", resourceName, "--zone", region, "--machine-type", "e2-micro")
		case "storage":
			cmd = exec.Command("gsutil", "mb", fmt.Sprintf("gs://%s", resourceName))
		default:
			fmt.Printf("Resource type %s not supported for GCP create action\n", resourceType)
			return nil
		}
	case "delete":
		switch resourceType {
		case "compute":
			cmd = exec.Command("gcloud", "compute", "instances", "delete", resourceName, "--zone", region, "--quiet")
		case "storage":
			cmd = exec.Command("gsutil", "rm", "-r", fmt.Sprintf("gs://%s", resourceName))
		default:
			fmt.Printf("Resource type %s not supported for GCP delete action\n", resourceType)
			return nil
		}
	case "list":
		switch resourceType {
		case "compute":
			cmd = exec.Command("gcloud", "compute", "instances", "list")
		case "storage":
			cmd = exec.Command("gsutil", "ls")
		case "sql":
			cmd = exec.Command("gcloud", "sql", "instances", "list")
		default:
			cmd = exec.Command("gcloud", resourceType, "--help")
		}
	default:
		fmt.Printf("Action %s not implemented for GCP\n", action)
		return nil
	}

	// Add project if specified
	if p.Project != "" && !strings.Contains(cmd.String(), "gsutil") {
		args := append([]string{"--project", p.Project}, cmd.Args[1:]...)
		cmd = exec.Command(cmd.Args[0], args...)
	}

	// Would actually run the command in a real implementation
	fmt.Printf("Would run: %s\n", cmd.String())
	return nil
}

// NewGCPProvider creates a new GCP provider
func NewGCPProvider() *GCPProvider {
	return &GCPProvider{
		BaseCloudProvider: BaseCloudProvider{
			Name:   "gcp",
			Region: "us-central1-a",
		},
	}
}
