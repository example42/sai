# Design Document

## Overview

This design implements three alternative installation providers (source, binary, script) for SAI, extending the existing provider architecture to support manual installation patterns. The implementation focuses on type safety, template function consistency, and seamless integration with existing SAI features.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    SAI CLI Interface                        │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                Provider Engine                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐│
│  │   Package   │ │  Container  │ │   Alternative Providers ││
│  │  Managers   │ │  Platforms  │ │                         ││
│  │             │ │             │ │  ┌─────────┐            ││
│  │ • apt       │ │ • docker    │ │  │ source  │            ││
│  │ • brew      │ │ • helm      │ │  │ binary  │            ││
│  │ • dnf       │ │ • k8s       │ │  │ script  │            ││
│  └─────────────┘ └─────────────┘ │  └─────────┘            ││
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                Template Engine                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  Template Functions                                     ││
│  │  • sai_source(idx, field)                             ││
│  │  • sai_binary(idx, field)                             ││
│  │  • sai_script(idx, field)                             ││
│  │  • Existing functions (sai_package, sai_service, etc.) ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                 SaiData System                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  Data Types                                             ││
│  │  • Source (build configurations)                       ││
│  │  • Binary (download configurations)                    ││
│  │  • Script (execution configurations)                   ││
│  │  • Existing types (Package, Service, etc.)             ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

### Provider Integration Flow

```
User Command → Provider Selection → Template Resolution → Action Execution
     │                │                    │                    │
     │                │                    │                    ▼
     │                │                    │            ┌─────────────┐
     │                │                    │            │   Source    │
     │                │                    │            │   • Download│
     │                │                    │            │   • Extract │
     │                │                    │            │   • Build   │
     │                │                    │            │   • Install │
     │                │                    │            └─────────────┘
     │                │                    │                    │
     │                │                    │            ┌─────────────┐
     │                │                    │            │   Binary    │
     │                │                    │            │   • Download│
     │                │                    │            │   • Verify  │
     │                │                    │            │   • Extract │
     │                │                    │            │   • Install │
     │                │                    │            └─────────────┘
     │                │                    │                    │
     │                │                    │            ┌─────────────┐
     │                │                    │            │   Script    │
     │                │                    │            │   • Download│
     │                │                    │            │   • Verify  │
     │                │                    │            │   • Execute │
     │                │                    │            └─────────────┘
     │                │                    │
     │                ▼                    ▼
     │        ┌─────────────────┐ ┌─────────────────┐
     │        │ Provider Config │ │ Template Engine │
     │        │ Resolution      │ │ Field Resolution│
     │        └─────────────────┘ └─────────────────┘
     │
     ▼
┌─────────────────┐
│ Action Routing  │
│ • install       │
│ • uninstall     │
│ • upgrade       │
│ • version       │
│ • info          │
└─────────────────┘
```

## Components and Interfaces

### 1. Type System Extensions

#### Source Type Definition
```go
type Source struct {
    Name            string                 `yaml:"name" json:"name"`
    URL             string                 `yaml:"url" json:"url"`
    Version         string                 `yaml:"version,omitempty" json:"version,omitempty"`
    BuildSystem     string                 `yaml:"build_system" json:"build_system"`
    BuildDir        string                 `yaml:"build_dir,omitempty" json:"build_dir,omitempty"`
    SourceDir       string                 `yaml:"source_dir,omitempty" json:"source_dir,omitempty"`
    InstallPrefix   string                 `yaml:"install_prefix,omitempty" json:"install_prefix,omitempty"`
    ConfigureArgs   []string               `yaml:"configure_args,omitempty" json:"configure_args,omitempty"`
    BuildArgs       []string               `yaml:"build_args,omitempty" json:"build_args,omitempty"`
    InstallArgs     []string               `yaml:"install_args,omitempty" json:"install_args,omitempty"`
    Prerequisites   []string               `yaml:"prerequisites,omitempty" json:"prerequisites,omitempty"`
    Environment     map[string]string      `yaml:"environment,omitempty" json:"environment,omitempty"`
    Checksum        string                 `yaml:"checksum,omitempty" json:"checksum,omitempty"`
    CustomCommands  *SourceCustomCommands  `yaml:"custom_commands,omitempty" json:"custom_commands,omitempty"`
}

type SourceCustomCommands struct {
    Download   string `yaml:"download,omitempty" json:"download,omitempty"`
    Extract    string `yaml:"extract,omitempty" json:"extract,omitempty"`
    Configure  string `yaml:"configure,omitempty" json:"configure,omitempty"`
    Build      string `yaml:"build,omitempty" json:"build,omitempty"`
    Install    string `yaml:"install,omitempty" json:"install,omitempty"`
    Uninstall  string `yaml:"uninstall,omitempty" json:"uninstall,omitempty"`
    Validation string `yaml:"validation,omitempty" json:"validation,omitempty"`
    Version    string `yaml:"version,omitempty" json:"version,omitempty"`
}
```

#### Binary Type Definition
```go
type Binary struct {
    Name         string                 `yaml:"name" json:"name"`
    URL          string                 `yaml:"url" json:"url"`
    Version      string                 `yaml:"version,omitempty" json:"version,omitempty"`
    Architecture string                 `yaml:"architecture,omitempty" json:"architecture,omitempty"`
    Platform     string                 `yaml:"platform,omitempty" json:"platform,omitempty"`
    Checksum     string                 `yaml:"checksum,omitempty" json:"checksum,omitempty"`
    InstallPath  string                 `yaml:"install_path,omitempty" json:"install_path,omitempty"`
    Executable   string                 `yaml:"executable,omitempty" json:"executable,omitempty"`
    Archive      *ArchiveConfig         `yaml:"archive,omitempty" json:"archive,omitempty"`
    Permissions  string                 `yaml:"permissions,omitempty" json:"permissions,omitempty"`
    CustomCommands *BinaryCustomCommands `yaml:"custom_commands,omitempty" json:"custom_commands,omitempty"`
}

type ArchiveConfig struct {
    Format      string `yaml:"format,omitempty" json:"format,omitempty"`
    StripPrefix string `yaml:"strip_prefix,omitempty" json:"strip_prefix,omitempty"`
    ExtractPath string `yaml:"extract_path,omitempty" json:"extract_path,omitempty"`
}

type BinaryCustomCommands struct {
    Download   string `yaml:"download,omitempty" json:"download,omitempty"`
    Extract    string `yaml:"extract,omitempty" json:"extract,omitempty"`
    Install    string `yaml:"install,omitempty" json:"install,omitempty"`
    Uninstall  string `yaml:"uninstall,omitempty" json:"uninstall,omitempty"`
    Validation string `yaml:"validation,omitempty" json:"validation,omitempty"`
    Version    string `yaml:"version,omitempty" json:"version,omitempty"`
}
```

#### Script Type Definition
```go
type Script struct {
    Name         string                 `yaml:"name" json:"name"`
    URL          string                 `yaml:"url" json:"url"`
    Version      string                 `yaml:"version,omitempty" json:"version,omitempty"`
    Interpreter  string                 `yaml:"interpreter,omitempty" json:"interpreter,omitempty"`
    Checksum     string                 `yaml:"checksum,omitempty" json:"checksum,omitempty"`
    Arguments    []string               `yaml:"arguments,omitempty" json:"arguments,omitempty"`
    Environment  map[string]string      `yaml:"environment,omitempty" json:"environment,omitempty"`
    WorkingDir   string                 `yaml:"working_dir,omitempty" json:"working_dir,omitempty"`
    Timeout      int                    `yaml:"timeout,omitempty" json:"timeout,omitempty"`
    CustomCommands *ScriptCustomCommands `yaml:"custom_commands,omitempty" json:"custom_commands,omitempty"`
}

type ScriptCustomCommands struct {
    Download   string `yaml:"download,omitempty" json:"download,omitempty"`
    Install    string `yaml:"install,omitempty" json:"install,omitempty"`
    Uninstall  string `yaml:"uninstall,omitempty" json:"uninstall,omitempty"`
    Validation string `yaml:"validation,omitempty" json:"validation,omitempty"`
    Version    string `yaml:"version,omitempty" json:"version,omitempty"`
}
```

### 2. Template Function Architecture

#### Template Function Interface
```go
type TemplateFunction interface {
    Name() string
    Execute(args ...interface{}) (string, error)
    Validate(args ...interface{}) error
}

// Source template function
func (e *TemplateEngine) saiSource(args ...interface{}) string {
    return e.executeTemplateFunction("sai_source", args...)
}

// Binary template function  
func (e *TemplateEngine) saiBinary(args ...interface{}) string {
    return e.executeTemplateFunction("sai_binary", args...)
}

// Script template function
func (e *TemplateEngine) saiScript(args ...interface{}) string {
    return e.executeTemplateFunction("sai_script", args...)
}
```

#### Field Resolution Strategy
```go
type FieldResolver struct {
    saidata      *types.SoftwareData
    providerName string
}

func (r *FieldResolver) ResolveSourceField(idx int, field string) (string, error) {
    // 1. Check provider-specific sources
    if providerConfig := r.saidata.GetProviderConfig(r.providerName); providerConfig != nil {
        if len(providerConfig.Sources) > idx {
            if value := r.extractSourceField(&providerConfig.Sources[idx], field); value != "" {
                return value, nil
            }
        }
    }
    
    // 2. Check default sources
    if len(r.saidata.Sources) > idx {
        if value := r.extractSourceField(&r.saidata.Sources[idx], field); value != "" {
            return value, nil
        }
    }
    
    // 3. Generate defaults
    return r.generateSourceDefault(field)
}
```

### 3. Provider Implementation Architecture

#### Provider File Structure
```yaml
# providers/source.yaml
version: "1.0"
provider:
  name: "source"
  type: "source"
  capabilities: ["install", "uninstall", "upgrade", "version", "info"]

actions:
  install:
    steps:
      - name: "install-prerequisites"
        command: "{{sai_source(0, 'prerequisites_install_cmd')}}"
      - name: "download-source"
        command: "{{sai_source(0, 'download_cmd')}}"
      # ... additional steps
```

#### Provider Action Execution Flow
```go
type ActionExecutor struct {
    provider     *Provider
    templateEngine *TemplateEngine
    executor     CommandExecutor
}

func (ae *ActionExecutor) ExecuteAction(action string, context *ActionContext) error {
    // 1. Resolve templates
    resolvedAction, err := ae.templateEngine.ResolveAction(ae.provider.Actions[action], context)
    if err != nil {
        return fmt.Errorf("template resolution failed: %w", err)
    }
    
    // 2. Execute steps
    for _, step := range resolvedAction.Steps {
        if err := ae.executor.Execute(step); err != nil {
            return ae.handleStepFailure(step, err)
        }
    }
    
    // 3. Validate result
    return ae.validateActionResult(resolvedAction)
}
```

## Data Models

### SoftwareData Extensions
```go
type SoftwareData struct {
    // Existing fields...
    Sources  []Source  `yaml:"sources,omitempty" json:"sources,omitempty"`
    Binaries []Binary  `yaml:"binaries,omitempty" json:"binaries,omitempty"`
    Scripts  []Script  `yaml:"scripts,omitempty" json:"scripts,omitempty"`
    // ... rest of fields
}

type ProviderConfig struct {
    // Existing fields...
    Sources  []Source  `yaml:"sources,omitempty" json:"sources,omitempty"`
    Binaries []Binary  `yaml:"binaries,omitempty" json:"binaries,omitempty"`
    Scripts  []Script  `yaml:"scripts,omitempty" json:"scripts,omitempty"`
    // ... rest of fields
}
```

### Default Value Generation
```go
type DefaultGenerator struct {
    metadata *Metadata
    platform PlatformInfo
}

func (dg *DefaultGenerator) GenerateSourceDefaults(source *Source) {
    if source.BuildDir == "" {
        source.BuildDir = fmt.Sprintf("/tmp/sai-build-%s", dg.metadata.Name)
    }
    if source.SourceDir == "" {
        source.SourceDir = fmt.Sprintf("%s/%s-%s", source.BuildDir, dg.metadata.Name, source.Version)
    }
    if source.InstallPrefix == "" {
        source.InstallPrefix = "/usr/local"
    }
}

func (dg *DefaultGenerator) GenerateBinaryDefaults(binary *Binary) {
    if binary.InstallPath == "" {
        binary.InstallPath = "/usr/local/bin"
    }
    if binary.Permissions == "" {
        binary.Permissions = "0755"
    }
}
```

## Error Handling

### Error Types
```go
type ProviderError struct {
    Provider string
    Action   string
    Step     string
    Cause    error
}

type TemplateResolutionError struct {
    Function string
    Field    string
    Cause    error
}

type ValidationError struct {
    Type    string
    Field   string
    Value   string
    Message string
}
```

### Rollback Strategy
```go
type RollbackManager struct {
    actions []RollbackAction
}

type RollbackAction struct {
    Description string
    Command     string
    IgnoreError bool
}

func (rm *RollbackManager) ExecuteRollback() error {
    for i := len(rm.actions) - 1; i >= 0; i-- {
        action := rm.actions[i]
        if err := rm.executeRollbackAction(action); err != nil && !action.IgnoreError {
            return fmt.Errorf("rollback failed at step %s: %w", action.Description, err)
        }
    }
    return nil
}
```

## Testing Strategy

### Unit Testing
- Template function resolution with various input combinations
- Type marshaling/unmarshaling for all new data structures
- Default value generation logic
- Error handling and rollback scenarios

### Integration Testing
- End-to-end provider execution with real software packages
- Cross-platform compatibility testing
- Provider interaction with existing SAI features
- Template resolution with complex saidata configurations

### Test Data Structure
```
tests/
├── unit/
│   ├── template/
│   │   ├── source_function_test.go
│   │   ├── binary_function_test.go
│   │   └── script_function_test.go
│   ├── types/
│   │   └── alternative_providers_test.go
│   └── providers/
│       ├── source_provider_test.go
│       ├── binary_provider_test.go
│       └── script_provider_test.go
├── integration/
│   ├── source_builds_test.go
│   ├── binary_downloads_test.go
│   └── script_execution_test.go
└── fixtures/
    ├── saidata/
    │   ├── nginx_source.yaml
    │   ├── terraform_binary.yaml
    │   └── docker_script.yaml
    └── providers/
        ├── source.yaml
        ├── binary.yaml
        └── script.yaml
```