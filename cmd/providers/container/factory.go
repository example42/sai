package container

// NewProvider creates the appropriate container provider based on the name
func NewProvider(name string) Provider {
	switch name {
	case "helm":
		return NewHelmProvider()
	case "kubectl":
		return NewKubectlProvider()
	default:
		// Return Kubectl provider as default
		return NewKubectlProvider()
	}
}
