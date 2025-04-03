package service

// Provider interface defines methods for service management
type Provider interface {
	Execute(action, service string) error
	GetServiceManager() string
	IsDryRun() bool
}

// Supported service actions
const (
	ActionStart   = "start"
	ActionStop    = "stop"
	ActionRestart = "restart"
	ActionEnable  = "enable"
	ActionDisable = "disable"
)

// AllServiceActions contains all supported service management actions
var AllServiceActions = []string{
	ActionStart,
	ActionStop,
	ActionRestart,
	ActionEnable,
	ActionDisable,
}

// IsValidAction checks if the given action is supported by service providers
func IsValidAction(action string) bool {
	for _, a := range AllServiceActions {
		if a == action {
			return true
		}
	}
	return false
}

// Global variable to track dry run mode
var isDryRunMode = false

// SetDryRun sets the dry run mode for service providers
func SetDryRun(enabled bool) {
	isDryRunMode = enabled
}

// BaseProvider common functionality for service providers
type BaseProvider struct {
	Name string
}

// GetServiceManager returns the service manager name
func (p *BaseProvider) GetServiceManager() string {
	return p.Name
}

// IsDryRun checks if dry run mode is enabled
func (p *BaseProvider) IsDryRun() bool {
	return isDryRunMode
}

// GetProvider returns the appropriate service provider for the given OS
func GetProvider(osType string) Provider {
	switch osType {
	case "darwin":
		return NewBrewProvider()
	default:
		return NewSystemdProvider()
	}
}
