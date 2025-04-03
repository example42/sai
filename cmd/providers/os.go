package providers

import (
	ospkg "sai/cmd/providers/os"
)

// This file is kept as a placeholder
// Actual OS providers are in the os/ subdirectory

// osProviderAdapter adapts an os.Provider to providers.Provider
type osProviderAdapter struct {
	provider ospkg.Provider
}

// Execute implements the Provider interface
func (a *osProviderAdapter) Execute(action, software string) error {
	return a.provider.Execute(action, software)
}

// NewOSProvider creates the appropriate OS provider based on the name
func NewOSProvider(name string) Provider {
	provider := ospkg.NewProvider(name)
	return &osProviderAdapter{provider: provider}
}
