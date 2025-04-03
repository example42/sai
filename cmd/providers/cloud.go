package providers

import (
	"fmt"
	"os/exec"
	cloudprovider "sai/cmd/providers/cloud"
	"strings"
)

// CloudProvider interface for all cloud providers
type CloudProvider interface {
	Provider
	GetCloudPlatform() string
	SetRegion(region string)
	GetRegion() string
}

// BaseCloudProvider common functionality for cloud providers
type BaseCloudProvider struct {
	Name   string
	Region string
}

// GetCloudPlatform returns the cloud platform name
func (p *BaseCloudProvider) GetCloudPlatform() string {
	return p.Name
}

// SetRegion sets the region for the cloud provider
func (p *BaseCloudProvider) SetRegion(region string) {
	p.Region = region
}

// GetRegion gets the current region
func (p *BaseCloudProvider) GetRegion() string {
	return p.Region
}

// AWSProvider handles AWS cloud operations
type AWSProvider struct {
	BaseCloudProvider
	Profile string
}

// Execute runs AWS CLI commands
func (p *AWSProvider) Execute(action, resource string) error {
	fmt.Printf("Executing %s %s with AWS provider\n", action, resource)

	// Set default region if not specified
	region := p.Region
	if region == "" {
		region = "us-east-1" // Default AWS region
	}

	// Parse resource type and name
	resourceType, resourceName := parseCloudResource(resource)

	var cmd *exec.Cmd
	switch action {
	case "start":
		switch resourceType {
		case "ec2":
			cmd = exec.Command("aws", "ec2", "start-instances", "--instance-ids", resourceName, "--region", region)
		case "rds":
			cmd = exec.Command("aws", "rds", "start-db-instance", "--db-instance-identifier", resourceName, "--region", region)
		default:
			fmt.Printf("Resource type %s not supported for AWS start action\n", resourceType)
			return nil
		}
	case "stop":
		switch resourceType {
		case "ec2":
			cmd = exec.Command("aws", "ec2", "stop-instances", "--instance-ids", resourceName, "--region", region)
		case "rds":
			cmd = exec.Command("aws", "rds", "stop-db-instance", "--db-instance-identifier", resourceName, "--region", region)
		default:
			fmt.Printf("Resource type %s not supported for AWS stop action\n", resourceType)
			return nil
		}
	case "status":
		switch resourceType {
		case "ec2":
			cmd = exec.Command("aws", "ec2", "describe-instances", "--instance-ids", resourceName, "--region", region)
		case "rds":
			cmd = exec.Command("aws", "rds", "describe-db-instances", "--db-instance-identifier", resourceName, "--region", region)
		case "s3":
			cmd = exec.Command("aws", "s3", "ls", resourceName, "--region", region)
		default:
			fmt.Printf("Resource type %s not supported for AWS status action\n", resourceType)
			return nil
		}
	case "create":
		switch resourceType {
		case "ec2":
			cmd = exec.Command("aws", "ec2", "run-instances", "--image-id", "ami-12345678", "--count", "1", "--instance-type", "t2.micro", "--region", region)
		case "s3":
			cmd = exec.Command("aws", "s3", "mb", fmt.Sprintf("s3://%s", resourceName), "--region", region)
		default:
			fmt.Printf("Resource type %s not supported for AWS create action\n", resourceType)
			return nil
		}
	case "delete":
		switch resourceType {
		case "ec2":
			cmd = exec.Command("aws", "ec2", "terminate-instances", "--instance-ids", resourceName, "--region", region)
		case "s3":
			cmd = exec.Command("aws", "s3", "rb", fmt.Sprintf("s3://%s", resourceName), "--force", "--region", region)
		default:
			fmt.Printf("Resource type %s not supported for AWS delete action\n", resourceType)
			return nil
		}
	case "list":
		switch resourceType {
		case "ec2":
			cmd = exec.Command("aws", "ec2", "describe-instances", "--region", region)
		case "s3":
			cmd = exec.Command("aws", "s3", "ls", "--region", region)
		case "rds":
			cmd = exec.Command("aws", "rds", "describe-db-instances", "--region", region)
		default:
			cmd = exec.Command("aws", resourceType, "help")
		}
	default:
		fmt.Printf("Action %s not implemented for AWS\n", action)
		return nil
	}

	// Add profile if specified
	if p.Profile != "" {
		args := append([]string{"--profile", p.Profile}, cmd.Args[1:]...)
		cmd = exec.Command(cmd.Args[0], args...)
	}

	// Would actually run the command in a real implementation
	fmt.Printf("Would run: %s\n", cmd.String())
	return nil
}

// AzureProvider handles Azure cloud operations
type AzureProvider struct {
	BaseCloudProvider
	Subscription  string
	ResourceGroup string
}

// Execute runs Azure CLI commands
func (p *AzureProvider) Execute(action, resource string) error {
	fmt.Printf("Executing %s %s with Azure provider\n", action, resource)

	// Set default location if not specified
	location := p.Region
	if location == "" {
		location = "eastus" // Default Azure location
	}

	// Parse resource type and name
	resourceType, resourceName := parseCloudResource(resource)

	var cmd *exec.Cmd
	switch action {
	case "start":
		switch resourceType {
		case "vm":
			cmd = exec.Command("az", "vm", "start", "--name", resourceName, "--resource-group", p.ResourceGroup)
		case "webapp":
			cmd = exec.Command("az", "webapp", "start", "--name", resourceName, "--resource-group", p.ResourceGroup)
		default:
			fmt.Printf("Resource type %s not supported for Azure start action\n", resourceType)
			return nil
		}
	case "stop":
		switch resourceType {
		case "vm":
			cmd = exec.Command("az", "vm", "stop", "--name", resourceName, "--resource-group", p.ResourceGroup)
		case "webapp":
			cmd = exec.Command("az", "webapp", "stop", "--name", resourceName, "--resource-group", p.ResourceGroup)
		default:
			fmt.Printf("Resource type %s not supported for Azure stop action\n", resourceType)
			return nil
		}
	case "status":
		switch resourceType {
		case "vm":
			cmd = exec.Command("az", "vm", "show", "--name", resourceName, "--resource-group", p.ResourceGroup)
		case "webapp":
			cmd = exec.Command("az", "webapp", "show", "--name", resourceName, "--resource-group", p.ResourceGroup)
		default:
			fmt.Printf("Resource type %s not supported for Azure status action\n", resourceType)
			return nil
		}
	case "create":
		switch resourceType {
		case "vm":
			cmd = exec.Command("az", "vm", "create", "--name", resourceName, "--resource-group", p.ResourceGroup, "--image", "UbuntuLTS", "--location", location)
		case "webapp":
			cmd = exec.Command("az", "webapp", "create", "--name", resourceName, "--resource-group", p.ResourceGroup, "--plan", "myAppServicePlan", "--location", location)
		default:
			fmt.Printf("Resource type %s not supported for Azure create action\n", resourceType)
			return nil
		}
	case "delete":
		switch resourceType {
		case "vm":
			cmd = exec.Command("az", "vm", "delete", "--name", resourceName, "--resource-group", p.ResourceGroup, "--yes")
		case "webapp":
			cmd = exec.Command("az", "webapp", "delete", "--name", resourceName, "--resource-group", p.ResourceGroup)
		default:
			fmt.Printf("Resource type %s not supported for Azure delete action\n", resourceType)
			return nil
		}
	case "list":
		switch resourceType {
		case "vm":
			cmd = exec.Command("az", "vm", "list", "--resource-group", p.ResourceGroup)
		case "webapp":
			cmd = exec.Command("az", "webapp", "list", "--resource-group", p.ResourceGroup)
		default:
			cmd = exec.Command("az", resourceType, "--help")
		}
	default:
		fmt.Printf("Action %s not implemented for Azure\n", action)
		return nil
	}

	// Add subscription if specified
	if p.Subscription != "" {
		args := append([]string{"--subscription", p.Subscription}, cmd.Args[1:]...)
		cmd = exec.Command(cmd.Args[0], args...)
	}

	// Would actually run the command in a real implementation
	fmt.Printf("Would run: %s\n", cmd.String())
	return nil
}

// GCPProvider handles Google Cloud Platform operations
type GCPProvider struct {
	BaseCloudProvider
	Project string
}

// Execute runs GCP gcloud commands
func (p *GCPProvider) Execute(action, resource string) error {
	fmt.Printf("Executing %s %s with GCP provider\n", action, resource)

	// Set default zone if not specified
	zone := p.Region
	if zone == "" {
		zone = "us-central1-a" // Default GCP zone
	}

	// Parse resource type and name
	resourceType, resourceName := parseCloudResource(resource)

	var cmd *exec.Cmd
	switch action {
	case "start":
		switch resourceType {
		case "compute":
			cmd = exec.Command("gcloud", "compute", "instances", "start", resourceName, "--zone", zone)
		case "sql":
			cmd = exec.Command("gcloud", "sql", "instances", "start", resourceName)
		default:
			fmt.Printf("Resource type %s not supported for GCP start action\n", resourceType)
			return nil
		}
	case "stop":
		switch resourceType {
		case "compute":
			cmd = exec.Command("gcloud", "compute", "instances", "stop", resourceName, "--zone", zone)
		case "sql":
			cmd = exec.Command("gcloud", "sql", "instances", "stop", resourceName)
		default:
			fmt.Printf("Resource type %s not supported for GCP stop action\n", resourceType)
			return nil
		}
	case "status":
		switch resourceType {
		case "compute":
			cmd = exec.Command("gcloud", "compute", "instances", "describe", resourceName, "--zone", zone)
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
			cmd = exec.Command("gcloud", "compute", "instances", "create", resourceName, "--zone", zone, "--machine-type", "e2-micro")
		case "storage":
			cmd = exec.Command("gsutil", "mb", fmt.Sprintf("gs://%s", resourceName))
		default:
			fmt.Printf("Resource type %s not supported for GCP create action\n", resourceType)
			return nil
		}
	case "delete":
		switch resourceType {
		case "compute":
			cmd = exec.Command("gcloud", "compute", "instances", "delete", resourceName, "--zone", zone, "--quiet")
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

// parseCloudResource splits "type/name" into (type, name)
func parseCloudResource(resource string) (string, string) {
	parts := strings.Split(resource, "/")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return resource, "" // If no slash, assume it's just the resource type
}

// cloudProviderAdapter adapts a cloud.Provider to providers.Provider
type cloudProviderAdapter struct {
	provider cloudprovider.Provider
}

// Execute implements the Provider interface
func (a *cloudProviderAdapter) Execute(action, resource string) error {
	return a.provider.Execute(action, resource)
}

// NewCloudProvider creates the appropriate cloud provider based on the name
func NewCloudProvider(name string) Provider {
	provider := cloudprovider.NewProvider(name)
	return &cloudProviderAdapter{provider: provider}
}
