package handlers

import (
	"fmt"
	"strings"
)

// InstallHandler handles the install command
type InstallHandler struct {
	BaseHandler
}

// NewInstallHandler creates a new install handler
func NewInstallHandler() *InstallHandler {
	return &InstallHandler{
		BaseHandler: BaseHandler{
			Action: "install",
		},
	}
}

// Handle executes the install command
func (h *InstallHandler) Handle(software string, provider string) {
	// If no software is specified, show available installation options
	if software == "" {
		h.showInstallationOptions(provider)
		return
	}

	// Otherwise, proceed with normal installation
	h.BaseHandler.Handle(software, provider)
}

// showInstallationOptions displays available installation options
func (h *InstallHandler) showInstallationOptions(provider string) {
	fmt.Println("Available installation options:")

	// If a specific provider is requested, show options for that provider
	if provider != "" {
		fmt.Printf("Installation options for provider: %s\n", provider)
		// Here you would list packages available from this provider
		return
	}

	// Otherwise, show options for all providers
	fmt.Println("Installation options by provider:")

	// List providers by type
	for providerType, providers := range ProvidersByType {
		fmt.Printf("\n%s providers:\n", strings.Title(string(providerType)))
		for _, p := range providers {
			fmt.Printf("  - %s\n", p)
		}
	}

	fmt.Println("\nTo install a specific package, use: sai install <package_name>")
	fmt.Println("To specify a provider, use: sai install <package_name> --provider <provider_name>")
}
