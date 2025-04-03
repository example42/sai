package os

import (
	"runtime"
	"sai/cmd/providers/os/pkgmanager"
	"sai/cmd/providers/os/service"
)

// Provider interface defines methods for OS provider implementations
type Provider interface {
	Execute(action, software string) error
	GetPackageManager() string
}

// pkgManagerAdapter adapts a pkgmanager.Provider to os.Provider
type pkgManagerAdapter struct {
	provider pkgmanager.Provider
}

// Execute implements the Provider interface
func (a *pkgManagerAdapter) Execute(action, software string) error {
	return a.provider.Execute(action, software)
}

// GetPackageManager returns the package manager name
func (a *pkgManagerAdapter) GetPackageManager() string {
	return a.provider.GetPackageManager()
}

// NewProvider creates the appropriate OS provider based on the name
func NewProvider(name string) Provider {
	pkgProvider := pkgmanager.GetProvider(name)
	return &pkgManagerAdapter{provider: pkgProvider}
}

// NewServiceProvider creates the appropriate service provider for the current OS
func NewServiceProvider() service.Provider {
	osType := runtime.GOOS
	return service.GetProvider(osType)
}
