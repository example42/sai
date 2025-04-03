package providers

// Provider interface defines methods for provider implementations
type Provider interface {
	Execute(action, software string) error
}

// Supported actions
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
	ActionInfo      = "info"
	ActionLogs      = "logs"
	ActionDescribe  = "describe"
)

// AllActions contains all supported actions
var AllActions = []string{
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
	ActionInfo,
	ActionLogs,
	ActionDescribe,
}

// IsValidAction checks if the given action is supported
func IsValidAction(action string) bool {
	for _, a := range AllActions {
		if a == action {
			return true
		}
	}
	return false
}
