package cloud

// NewProvider creates the appropriate cloud provider based on the name
func NewProvider(name string) Provider {
	switch name {
	case "aws":
		return NewAWSProvider()
	case "azure":
		return NewAzureProvider()
	case "gcp":
		return NewGCPProvider()
	default:
		// Return AWS provider as default
		return NewAWSProvider()
	}
}
