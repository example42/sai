package container

// Provider interface defines methods for container provider implementations
type Provider interface {
	Execute(action, resource string) error
	GetContainerTool() string
	IsDryRun() bool
}

// Supported container actions
const (
	ActionInstall   = "install"
	ActionUninstall = "uninstall"
	ActionStatus    = "status"
	ActionStart     = "start"
	ActionStop      = "stop"
	ActionRestart   = "restart"
	ActionCreate    = "create"
	ActionDelete    = "delete"
	ActionList      = "list"
	ActionSearch    = "search"
	ActionUpgrade   = "upgrade"
	ActionLogs      = "logs"
	ActionDescribe  = "describe"
)

// AllContainerActions contains all supported container provider actions
var AllContainerActions = []string{
	ActionInstall,
	ActionUninstall,
	ActionStatus,
	ActionStart,
	ActionStop,
	ActionRestart,
	ActionCreate,
	ActionDelete,
	ActionList,
	ActionSearch,
	ActionUpgrade,
	ActionLogs,
	ActionDescribe,
}

// IsValidContainerAction checks if the given action is supported by container providers
func IsValidContainerAction(action string) bool {
	for _, a := range AllContainerActions {
		if a == action {
			return true
		}
	}
	return false
}

// Global variable to track dry run mode
var isDryRunMode = false

// SetDryRun sets the dry run mode for container providers
func SetDryRun(enabled bool) {
	isDryRunMode = enabled
}

// BaseContainerProvider common functionality for container providers
type BaseContainerProvider struct {
	Name string
}

// GetContainerTool returns the container orchestration tool name
func (p *BaseContainerProvider) GetContainerTool() string {
	return p.Name
}

// IsDryRun checks if dry run mode is enabled
func (p *BaseContainerProvider) IsDryRun() bool {
	return isDryRunMode
}
