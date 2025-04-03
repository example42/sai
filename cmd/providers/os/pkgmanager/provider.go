package pkgmanager

// Provider interface defines methods for package manager implementations
type Provider interface {
	Execute(action, software string) error
	GetPackageManager() string
	IsDryRun() bool // Add method to check if dry run mode is active
}

// Supported package manager actions
const (
	ActionInstall   = "install"
	ActionUninstall = "uninstall"
	ActionStatus    = "status"
	ActionList      = "list"
	ActionSearch    = "search"
	ActionUpgrade   = "upgrade"
	ActionInfo      = "info"
)

// AllActions contains all supported package manager actions
var AllActions = []string{
	ActionInstall,
	ActionUninstall,
	ActionStatus,
	ActionList,
	ActionSearch,
	ActionUpgrade,
	ActionInfo,
}

// IsValidAction checks if the given action is supported by package managers
func IsValidAction(action string) bool {
	for _, a := range AllActions {
		if a == action {
			return true
		}
	}
	return false
}

// Global variable to track dry run mode
var isDryRunMode = false

// SetDryRun sets the dry run mode for package managers
func SetDryRun(enabled bool) {
	isDryRunMode = enabled
}

// BaseProvider common functionality for package manager providers
type BaseProvider struct {
	Name string
}

// GetPackageManager returns the package manager name
func (p *BaseProvider) GetPackageManager() string {
	return p.Name
}

// IsDryRun checks if dry run mode is enabled
func (p *BaseProvider) IsDryRun() bool {
	return isDryRunMode
}
