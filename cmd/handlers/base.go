package handlers

import (
	"fmt"
	"runtime"
	"strings"

	"sai/cmd/providers"
	"sai/cmd/providers/os/service"
)

// CommandHandler function type
type CommandHandler func(string, string)

// ProviderType represents a category of providers
type ProviderType string

// Provider interface defines methods for provider implementations
type Provider interface {
	Execute(action, software string) error
}

// Provider types
const (
	ProviderTypeOS        ProviderType = "os"        // OS package managers
	ProviderTypeContainer ProviderType = "container" // Container orchestration tools
	ProviderTypeCloud     ProviderType = "cloud"     // Cloud service providers
)

// Supported OS providers
const (
	ProviderRPM    = "rpm"
	ProviderAPT    = "apt"
	ProviderBrew   = "brew"
	ProviderWinget = "winget"
	ProviderPacman = "pacman"
	ProviderZypper = "zypper"
)

// Supported container providers
const (
	ProviderHelm    = "helm"
	ProviderKubectl = "kubectl"
)

// Supported cloud providers
const (
	ProviderAWS   = "aws"
	ProviderAzure = "azure"
	ProviderGCP   = "gcp"
)

// OS constants
const (
	OSLinux   = "linux"
	OSWindows = "windows"
	OSMacOS   = "darwin"
)

// Linux distro constants
const (
	LinuxRedHat = "redhat"
	LinuxDebian = "debian"
	LinuxUbuntu = "ubuntu"
	LinuxSuse   = "suse"
	LinuxArch   = "arch"
	LinuxOther  = "other"
)

// OSDistroKey represents a combination of OS and distro
type OSDistroKey struct {
	OS     string
	Distro string
}

// ProvidersByType maps provider types to lists of providers
var ProvidersByType = map[ProviderType][]string{
	ProviderTypeOS: {
		ProviderRPM,
		ProviderAPT,
		ProviderBrew,
		ProviderWinget,
		ProviderPacman,
		ProviderZypper,
	},
	ProviderTypeContainer: {
		ProviderHelm,
		ProviderKubectl,
	},
	ProviderTypeCloud: {
		ProviderAWS,
		ProviderAzure,
		ProviderGCP,
	},
}

// AllProviders contains all supported providers, created by merging all provider types
var AllProviders = func() []string {
	var all []string
	for _, providers := range ProvidersByType {
		all = append(all, providers...)
	}
	return all
}()

// DefaultProvidersByOSDistro maps OS+distro combinations directly to providers
var DefaultProvidersByOSDistro = map[OSDistroKey]string{
	// Windows mapping
	{OSWindows, ""}: ProviderWinget,

	// macOS mapping
	{OSMacOS, ""}: ProviderBrew,

	// Linux distro mappings
	{OSLinux, LinuxRedHat}: ProviderRPM,
	{OSLinux, LinuxDebian}: ProviderAPT,
	{OSLinux, LinuxUbuntu}: ProviderAPT,
	{OSLinux, LinuxSuse}:   ProviderZypper,
	{OSLinux, LinuxArch}:   ProviderPacman,
	{OSLinux, LinuxOther}:  ProviderAPT,
}

// BaseHandler provides common functionality for all handlers
type BaseHandler struct {
	Action       string
	Provider     string
	ProviderType ProviderType
}

// OSProvider handles OS package manager operations
type OSProvider struct {
	Name string
}

// Execute runs the command for an OS provider
func (p *OSProvider) Execute(action, software string) error {
	fmt.Printf("Executing %s %s with OS provider %s\n", action, software, p.Name)
	// Actual implementation would run the appropriate package manager command
	return nil
}

// ContainerProvider handles container operations
type ContainerProvider struct {
	Name string
}

// Execute runs the command for a container provider
func (p *ContainerProvider) Execute(action, software string) error {
	fmt.Printf("Executing %s %s with container provider %s\n", action, software, p.Name)
	// Actual implementation would run the appropriate container command
	return nil
}

// CloudProvider handles cloud operations
type CloudProvider struct {
	Name string
}

// Execute runs the command for a cloud provider
func (p *CloudProvider) Execute(action, software string) error {
	fmt.Printf("Executing %s %s with cloud provider %s\n", action, software, p.Name)
	// Actual implementation would run the appropriate cloud command
	return nil
}

// newProvider creates a new provider instance based on provider name and type
func newProvider(name string, providerType ProviderType) providers.Provider {
	switch providerType {
	case ProviderTypeOS:
		return providers.NewOSProvider(name)
	case ProviderTypeContainer:
		return providers.NewContainerProvider(name)
	case ProviderTypeCloud:
		return providers.NewCloudProvider(name)
	default:
		return providers.NewOSProvider("apt") // Default fallback
	}
}

// validateProvider validates the provider string and returns its type
func validateProvider(provider string) (string, ProviderType) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	if provider == "" {
		return "", ""
	}

	for providerType, providers := range ProvidersByType {
		for _, p := range providers {
			if p == provider {
				return provider, providerType
			}
		}
	}

	return "", ""
}

// detectLinuxDistro attempts to determine the Linux distribution
func detectLinuxDistro() string {
	// Implementation omitted for brevity
	return LinuxDebian
}

// detectCurrentOS gets the current OS and distro
func detectCurrentOS() (string, string) {
	os := runtime.GOOS
	var distro string
	if os == OSLinux {
		distro = detectLinuxDistro()
	}
	return os, distro
}

// isServiceAction checks if the action is a service operation
func isServiceAction(action string) bool {
	serviceActions := []string{"start", "stop", "restart", "enable", "disable"}
	for _, a := range serviceActions {
		if action == a {
			return true
		}
	}
	return false
}

// DetectDefaultProvider detects the default provider for the current OS
func (h *BaseHandler) DetectDefaultProvider() string {
	os, distro := detectCurrentOS()
	key := OSDistroKey{OS: os, Distro: distro}
	if provider, ok := DefaultProvidersByOSDistro[key]; ok {
		return provider
	}
	// Default fallback
	return ProviderAPT
}

// SetProvider sets the provider and type based on the specified provider name
func (h *BaseHandler) SetProvider(provider string) {
	if provider == "" {
		h.Provider = h.DetectDefaultProvider()
		h.ProviderType = ProviderTypeOS
		return
	}

	validProvider, providerType := validateProvider(provider)
	if validProvider != "" {
		h.Provider = validProvider
		h.ProviderType = providerType
	} else {
		h.Provider = h.DetectDefaultProvider()
		h.ProviderType = ProviderTypeOS
	}
}

// GetProvider returns the provider name and type
func (h *BaseHandler) GetProvider() (string, ProviderType) {
	return h.Provider, h.ProviderType
}

// formatMessage formats the message for the given action
func formatMessage(action, software, provider string, providerType ProviderType) string {
	if isServiceAction(action) {
		return fmt.Sprintf("%s service %s using %s provider", action, software, provider)
	}
	return fmt.Sprintf("%s %s using %s provider %s", action, software, providerType, provider)
}

// handleServiceAction handles service actions
func (h *BaseHandler) handleServiceAction(software string) {
	// Check if in dry run mode
	if IsDryRun() {
		fmt.Printf("[DRY RUN] Service command would be executed: %s service %s\n", h.Action, software)
		return
	}

	fmt.Printf("%s service %s\n", h.Action, software)
	// Create a service provider for the current OS
	serviceProvider := service.GetProvider(runtime.GOOS)
	err := serviceProvider.Execute(h.Action, software)
	if err != nil {
		fmt.Printf("Error executing service command: %v\n", err)
	}
}

// handlePackageAction handles package actions
func (h *BaseHandler) handlePackageAction(software string) {
	// Get provider details
	provider, providerType := h.GetProvider()

	// Check if in dry run mode
	if IsDryRun() {
		fmt.Printf("[DRY RUN] Command would be executed: %s %s using %s provider %s\n",
			h.Action, software, providerType, provider)
		return
	}

	fmt.Println(formatMessage(h.Action, software, provider, providerType))

	providerImpl := newProvider(provider, providerType)
	err := providerImpl.Execute(h.Action, software)
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
	}
}

// Handle executes the handler with specified software and provider
func (h *BaseHandler) Handle(software, provider string) {
	h.SetProvider(provider)

	if isServiceAction(h.Action) {
		h.handleServiceAction(software)
	} else {
		h.handlePackageAction(software)
	}
}
