package types

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// ProviderData represents the complete provider configuration loaded from YAML
type ProviderData struct {
	Version  string                 `yaml:"version" json:"version"`
	Provider ProviderInfo          `yaml:"provider" json:"provider"`
	Actions  map[string]Action     `yaml:"actions" json:"actions"`
	Mappings *Mappings             `yaml:"mappings,omitempty" json:"mappings,omitempty"`
}

// ProviderInfo contains metadata about the provider
type ProviderInfo struct {
	Name         string   `yaml:"name" json:"name"`
	DisplayName  string   `yaml:"display_name,omitempty" json:"display_name,omitempty"`
	Description  string   `yaml:"description,omitempty" json:"description,omitempty"`
	Type         string   `yaml:"type" json:"type"`
	Platforms    []string `yaml:"platforms,omitempty" json:"platforms,omitempty"`
	Capabilities []string `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	Priority     int      `yaml:"priority,omitempty" json:"priority,omitempty"`
	Executable   string   `yaml:"executable,omitempty" json:"executable,omitempty"`
}

// Action represents a single action that can be performed by the provider
type Action struct {
	Description   string            `yaml:"description,omitempty" json:"description,omitempty"`
	Template      string            `yaml:"template,omitempty" json:"template,omitempty"`
	Command       string            `yaml:"command,omitempty" json:"command,omitempty"`
	Script        string            `yaml:"script,omitempty" json:"script,omitempty"`
	Steps         []Step            `yaml:"steps,omitempty" json:"steps,omitempty"`
	RequiresRoot  bool              `yaml:"requires_root,omitempty" json:"requires_root,omitempty"`
	Timeout       int               `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Retry         *RetryConfig      `yaml:"retry,omitempty" json:"retry,omitempty"`
	Validation    *Validation       `yaml:"validation,omitempty" json:"validation,omitempty"`
	Rollback      string            `yaml:"rollback,omitempty" json:"rollback,omitempty"`
	Variables     map[string]string `yaml:"variables,omitempty" json:"variables,omitempty"`
	Detection     string            `yaml:"detection,omitempty" json:"detection,omitempty"`
}

// Step represents a single step in a multi-step action
type Step struct {
	Name          string `yaml:"name,omitempty" json:"name,omitempty"`
	Command       string `yaml:"command" json:"command"`
	Condition     string `yaml:"condition,omitempty" json:"condition,omitempty"`
	IgnoreFailure bool   `yaml:"ignore_failure,omitempty" json:"ignore_failure,omitempty"`
	Timeout       int    `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// RetryConfig defines retry behavior for actions
type RetryConfig struct {
	Attempts int    `yaml:"attempts,omitempty" json:"attempts,omitempty"`
	Delay    int    `yaml:"delay,omitempty" json:"delay,omitempty"`
	Backoff  string `yaml:"backoff,omitempty" json:"backoff,omitempty"`
}

// Validation defines validation criteria for action success
type Validation struct {
	Command          string `yaml:"command" json:"command"`
	ExpectedExitCode int    `yaml:"expected_exit_code,omitempty" json:"expected_exit_code,omitempty"`
	ExpectedOutput   string `yaml:"expected_output,omitempty" json:"expected_output,omitempty"`
	Timeout          int    `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// Mappings define how to map saidata logical components to provider-specific implementations
type Mappings struct {
	Packages    map[string]PackageMapping   `yaml:"packages,omitempty" json:"packages,omitempty"`
	Services    map[string]ServiceMapping   `yaml:"services,omitempty" json:"services,omitempty"`
	Files       map[string]FileMapping      `yaml:"files,omitempty" json:"files,omitempty"`
	Directories map[string]DirectoryMapping `yaml:"directories,omitempty" json:"directories,omitempty"`
	Commands    map[string]CommandMapping   `yaml:"commands,omitempty" json:"commands,omitempty"`
	Ports       map[string]PortMapping      `yaml:"ports,omitempty" json:"ports,omitempty"`
	Variables   map[string]VariableMapping  `yaml:"variables,omitempty" json:"variables,omitempty"`
}

// PackageMapping maps logical packages to provider packages
type PackageMapping struct {
	Name           string   `yaml:"name" json:"name"`
	Version        string   `yaml:"version,omitempty" json:"version,omitempty"`
	Repository     string   `yaml:"repository,omitempty" json:"repository,omitempty"`
	Alternatives   []string `yaml:"alternatives,omitempty" json:"alternatives,omitempty"`
	InstallOptions string   `yaml:"install_options,omitempty" json:"install_options,omitempty"`
}

// ServiceMapping maps logical services to provider services
type ServiceMapping struct {
	Name        string            `yaml:"name" json:"name"`
	Type        string            `yaml:"type,omitempty" json:"type,omitempty"`
	ConfigFiles []string          `yaml:"config_files,omitempty" json:"config_files,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
}

// FileMapping maps logical files to provider file paths
type FileMapping struct {
	Path     string `yaml:"path" json:"path"`
	Owner    string `yaml:"owner,omitempty" json:"owner,omitempty"`
	Group    string `yaml:"group,omitempty" json:"group,omitempty"`
	Mode     string `yaml:"mode,omitempty" json:"mode,omitempty"`
	Template string `yaml:"template,omitempty" json:"template,omitempty"`
}

// DirectoryMapping maps logical directories to provider directory paths
type DirectoryMapping struct {
	Path   string `yaml:"path" json:"path"`
	Owner  string `yaml:"owner,omitempty" json:"owner,omitempty"`
	Group  string `yaml:"group,omitempty" json:"group,omitempty"`
	Mode   string `yaml:"mode,omitempty" json:"mode,omitempty"`
	Create bool   `yaml:"create,omitempty" json:"create,omitempty"`
}

// CommandMapping maps logical commands to provider command paths
type CommandMapping struct {
	Path         string   `yaml:"path" json:"path"`
	Alternatives []string `yaml:"alternatives,omitempty" json:"alternatives,omitempty"`
	Wrapper      string   `yaml:"wrapper,omitempty" json:"wrapper,omitempty"`
}

// PortMapping maps logical ports to provider port configurations
type PortMapping struct {
	Port         interface{} `yaml:"port,omitempty" json:"port,omitempty"` // Can be int or string
	Configurable bool        `yaml:"configurable,omitempty" json:"configurable,omitempty"`
	ConfigKey    string      `yaml:"config_key,omitempty" json:"config_key,omitempty"`
}

// VariableMapping maps saidata variables to provider-specific values
type VariableMapping struct {
	Value       interface{} `yaml:"value,omitempty" json:"value,omitempty"` // Can be string, int, or bool
	ConfigKey   string      `yaml:"config_key,omitempty" json:"config_key,omitempty"`
	Environment string      `yaml:"environment,omitempty" json:"environment,omitempty"`
}

// LoadProviderFromYAML loads provider data from YAML bytes
func LoadProviderFromYAML(data []byte) (*ProviderData, error) {
	var provider ProviderData
	if err := yaml.Unmarshal(data, &provider); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provider YAML: %w", err)
	}
	
	// Set default values
	if provider.Provider.Priority == 0 {
		provider.Provider.Priority = 50 // Default priority
	}
	
	// Set default timeouts for actions
	for name, action := range provider.Actions {
		if action.Timeout == 0 {
			action.Timeout = 300 // Default 5 minutes
		}
		if action.Validation != nil && action.Validation.Timeout == 0 {
			action.Validation.Timeout = 30 // Default 30 seconds for validation
		}
		if action.Retry != nil {
			if action.Retry.Attempts == 0 {
				action.Retry.Attempts = 3
			}
			if action.Retry.Delay == 0 {
				action.Retry.Delay = 5
			}
			if action.Retry.Backoff == "" {
				action.Retry.Backoff = "linear"
			}
		}
		provider.Actions[name] = action
	}
	
	return &provider, nil
}

// ToJSON converts the provider data to JSON for validation
func (p *ProviderData) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// GetTimeout returns the timeout for an action with fallback to default
func (a *Action) GetTimeout() time.Duration {
	if a.Timeout > 0 {
		return time.Duration(a.Timeout) * time.Second
	}
	return 300 * time.Second // Default 5 minutes
}

// HasSteps returns true if the action has multiple steps
func (a *Action) HasSteps() bool {
	return len(a.Steps) > 0
}

// GetCommand returns the appropriate command for the action
func (a *Action) GetCommand() string {
	if a.Template != "" {
		return a.Template
	}
	if a.Command != "" {
		return a.Command
	}
	if a.Script != "" {
		return a.Script
	}
	return ""
}

// IsValid checks if the action has at least one execution method
func (a *Action) IsValid() bool {
	return a.Template != "" || a.Command != "" || a.Script != "" || len(a.Steps) > 0
}

// GetPortAsInt returns the port as an integer, handling both int and string types
func (p *PortMapping) GetPortAsInt() (int, error) {
	switch v := p.Port.(type) {
	case int:
		return v, nil
	case string:
		// Try to parse as integer
		var port int
		if err := json.Unmarshal([]byte(v), &port); err != nil {
			return 0, fmt.Errorf("invalid port format: %s", v)
		}
		return port, nil
	default:
		return 0, fmt.Errorf("port must be int or string, got %T", v)
	}
}

// GetValueAsString returns the variable value as a string
func (v *VariableMapping) GetValueAsString() string {
	switch val := v.Value.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%d", val)
	case bool:
		return fmt.Sprintf("%t", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}