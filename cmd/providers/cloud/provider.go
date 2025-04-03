package cloud

// Provider interface defines methods for cloud provider implementations
type Provider interface {
	Execute(action, resource string) error
	GetCloudPlatform() string
	SetRegion(region string)
	GetRegion() string
	IsDryRun() bool
}

// Supported cloud actions
const (
	ActionStart    = "start"
	ActionStop     = "stop"
	ActionStatus   = "status"
	ActionCreate   = "create"
	ActionDelete   = "delete"
	ActionList     = "list"
	ActionDescribe = "describe"
)

// AllCloudActions contains all supported cloud provider actions
var AllCloudActions = []string{
	ActionStart,
	ActionStop,
	ActionStatus,
	ActionCreate,
	ActionDelete,
	ActionList,
	ActionDescribe,
}

// IsValidCloudAction checks if the given action is supported by cloud providers
func IsValidCloudAction(action string) bool {
	for _, a := range AllCloudActions {
		if a == action {
			return true
		}
	}
	return false
}

// Global variable to track dry run mode
var isDryRunMode = false

// SetDryRun sets the dry run mode for cloud providers
func SetDryRun(enabled bool) {
	isDryRunMode = enabled
}

// BaseCloudProvider common functionality for cloud providers
type BaseCloudProvider struct {
	Name   string
	Region string
}

// GetCloudPlatform returns the cloud platform name
func (p *BaseCloudProvider) GetCloudPlatform() string {
	return p.Name
}

// SetRegion sets the region for the cloud provider
func (p *BaseCloudProvider) SetRegion(region string) {
	p.Region = region
}

// GetRegion gets the current region
func (p *BaseCloudProvider) GetRegion() string {
	return p.Region
}

// IsDryRun checks if dry run mode is enabled
func (p *BaseCloudProvider) IsDryRun() bool {
	return isDryRunMode
}
