package provider

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/xeipuuv/gojsonschema"

	"sai/internal/types"
)

// ProviderLoader implements the provider loading functionality
type ProviderLoader struct {
	schemaPath   string
	schema       *gojsonschema.Schema
	mu           sync.RWMutex
	watchers     map[string]*fsnotify.Watcher
	callbacks    map[string][]func(*types.ProviderData)
}

// NewProviderLoader creates a new provider loader with schema validation
func NewProviderLoader(schemaPath string) (*ProviderLoader, error) {
	loader := &ProviderLoader{
		schemaPath: schemaPath,
		watchers:   make(map[string]*fsnotify.Watcher),
		callbacks:  make(map[string][]func(*types.ProviderData)),
	}

	// Load and compile the JSON schema
	if err := loader.loadSchema(); err != nil {
		return nil, fmt.Errorf("failed to load provider schema: %w", err)
	}

	return loader, nil
}

// loadSchema loads and compiles the JSON schema for validation
func (pl *ProviderLoader) loadSchema() error {
	if pl.schemaPath == "" {
		return fmt.Errorf("schema path is required")
	}

	schemaData, err := os.ReadFile(pl.schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file %s: %w", pl.schemaPath, err)
	}

	schemaLoader := gojsonschema.NewBytesLoader(schemaData)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return fmt.Errorf("failed to compile schema: %w", err)
	}

	pl.schema = schema
	return nil
}

// LoadFromFile loads a single provider from a YAML file
func (pl *ProviderLoader) LoadFromFile(filepath string) (*types.ProviderData, error) {
	// Read the YAML file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read provider file %s: %w", filepath, err)
	}

	// Parse YAML
	provider, err := types.LoadProviderFromYAML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provider YAML from %s: %w", filepath, err)
	}

	// Validate against schema
	if err := pl.ValidateProvider(provider); err != nil {
		return nil, fmt.Errorf("provider validation failed for %s: %w", filepath, err)
	}

	return provider, nil
}

// LoadFromDirectory loads all provider YAML files from a directory
func (pl *ProviderLoader) LoadFromDirectory(dirpath string) ([]*types.ProviderData, error) {
	var providers []*types.ProviderData
	var loadErrors []string

	// Walk through the directory and find YAML files
	err := filepath.WalkDir(dirpath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-YAML files
		if d.IsDir() || (!strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml")) {
			return nil
		}

		// Load the provider
		provider, loadErr := pl.LoadFromFile(path)
		if loadErr != nil {
			loadErrors = append(loadErrors, fmt.Sprintf("%s: %v", path, loadErr))
			return nil // Continue loading other files
		}

		providers = append(providers, provider)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk provider directory %s: %w", dirpath, err)
	}

	// Report any load errors as warnings but don't fail completely
	if len(loadErrors) > 0 {
		return providers, fmt.Errorf("some providers failed to load: %s", strings.Join(loadErrors, "; "))
	}

	return providers, nil
}

// ValidateProvider validates a provider configuration against the JSON schema
func (pl *ProviderLoader) ValidateProvider(provider *types.ProviderData) error {
	if pl.schema == nil {
		return fmt.Errorf("schema not loaded")
	}

	// Convert provider to JSON for validation
	jsonData, err := provider.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert provider to JSON: %w", err)
	}

	// Validate against schema
	documentLoader := gojsonschema.NewBytesLoader(jsonData)
	result, err := pl.schema.Validate(documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}

	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, desc.String())
		}
		return fmt.Errorf("provider validation failed: %s", strings.Join(errors, "; "))
	}

	// Additional business logic validation
	if err := pl.validateProviderLogic(provider); err != nil {
		return fmt.Errorf("provider logic validation failed: %w", err)
	}

	return nil
}

// validateProviderLogic performs additional validation beyond schema
func (pl *ProviderLoader) validateProviderLogic(provider *types.ProviderData) error {
	// Validate provider name is not empty
	if provider.Provider.Name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	// Validate that at least one action is defined
	if len(provider.Actions) == 0 {
		return fmt.Errorf("provider must define at least one action")
	}

	// Validate each action
	for actionName, action := range provider.Actions {
		if !action.IsValid() {
			return fmt.Errorf("action %s is invalid: must have template, command, script, or steps", actionName)
		}

		// Validate steps if present
		if action.HasSteps() {
			for i, step := range action.Steps {
				if step.Command == "" {
					return fmt.Errorf("action %s step %d: command cannot be empty", actionName, i)
				}
			}
		}

		// Validate retry configuration
		if action.Retry != nil {
			if action.Retry.Attempts < 1 {
				return fmt.Errorf("action %s: retry attempts must be at least 1", actionName)
			}
			if action.Retry.Delay < 0 {
				return fmt.Errorf("action %s: retry delay cannot be negative", actionName)
			}
			if action.Retry.Backoff != "" && action.Retry.Backoff != "linear" && action.Retry.Backoff != "exponential" {
				return fmt.Errorf("action %s: retry backoff must be 'linear' or 'exponential'", actionName)
			}
		}

		// Validate validation configuration
		if action.Validation != nil {
			if action.Validation.Command == "" {
				return fmt.Errorf("action %s: validation command cannot be empty", actionName)
			}
			if action.Validation.Timeout < 0 {
				return fmt.Errorf("action %s: validation timeout cannot be negative", actionName)
			}
		}
	}

	// Validate provider type
	validTypes := []string{
		"package_manager", "container", "binary", "source", "cloud", "custom",
		"debug", "trace", "profile", "security", "sbom", "troubleshoot",
		"network", "audit", "backup", "filesystem", "system", "monitoring",
		"io", "memory", "monitor", "process", "file", "directory", "command",
		"service", "port", "log", "config", "data", "temp", "cache",
	}
	
	typeValid := false
	for _, validType := range validTypes {
		if provider.Provider.Type == validType {
			typeValid = true
			break
		}
	}
	
	if !typeValid {
		return fmt.Errorf("invalid provider type: %s", provider.Provider.Type)
	}

	return nil
}

// WatchDirectory watches a directory for changes and calls callbacks when providers are updated
func (pl *ProviderLoader) WatchDirectory(dirpath string, callback func(*types.ProviderData)) error {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	// Create a new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Add the directory to the watcher
	err = watcher.Add(dirpath)
	if err != nil {
		watcher.Close()
		return fmt.Errorf("failed to watch directory %s: %w", dirpath, err)
	}

	// Store the watcher and callback
	pl.watchers[dirpath] = watcher
	if pl.callbacks[dirpath] == nil {
		pl.callbacks[dirpath] = make([]func(*types.ProviderData), 0)
	}
	pl.callbacks[dirpath] = append(pl.callbacks[dirpath], callback)

	// Start watching in a goroutine
	go pl.watchLoop(dirpath, watcher)

	return nil
}

// watchLoop handles file system events for a directory
func (pl *ProviderLoader) watchLoop(dirpath string, watcher *fsnotify.Watcher) {
	defer watcher.Close()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Only process YAML files
			if !strings.HasSuffix(event.Name, ".yaml") && !strings.HasSuffix(event.Name, ".yml") {
				continue
			}

			// Handle write and create events
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				pl.handleFileChange(dirpath, event.Name)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			// Log error but continue watching
			fmt.Printf("File watcher error for %s: %v\n", dirpath, err)
		}
	}
}

// handleFileChange processes a file change event
func (pl *ProviderLoader) handleFileChange(dirpath, filepath string) {
	// Load the updated provider
	provider, err := pl.LoadFromFile(filepath)
	if err != nil {
		fmt.Printf("Failed to reload provider from %s: %v\n", filepath, err)
		return
	}

	// Call all registered callbacks
	pl.mu.RLock()
	callbacks := pl.callbacks[dirpath]
	pl.mu.RUnlock()

	for _, callback := range callbacks {
		callback(provider)
	}
}

// StopWatching stops watching a directory
func (pl *ProviderLoader) StopWatching(dirpath string) error {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	watcher, exists := pl.watchers[dirpath]
	if !exists {
		return fmt.Errorf("no watcher found for directory: %s", dirpath)
	}

	err := watcher.Close()
	delete(pl.watchers, dirpath)
	delete(pl.callbacks, dirpath)

	return err
}

// StopAllWatching stops all directory watchers
func (pl *ProviderLoader) StopAllWatching() error {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	var errors []string
	for dirpath, watcher := range pl.watchers {
		if err := watcher.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", dirpath, err))
		}
	}

	// Clear all watchers and callbacks
	pl.watchers = make(map[string]*fsnotify.Watcher)
	pl.callbacks = make(map[string][]func(*types.ProviderData))

	if len(errors) > 0 {
		return fmt.Errorf("errors stopping watchers: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetSupportedProviderTypes returns the list of supported provider types
func (pl *ProviderLoader) GetSupportedProviderTypes() []string {
	return []string{
		"package_manager", "container", "binary", "source", "cloud", "custom",
		"debug", "trace", "profile", "security", "sbom", "troubleshoot",
		"network", "audit", "backup", "filesystem", "system", "monitoring",
		"io", "memory", "monitor", "process",
	}
}

// Note: ProviderLoader will be used by ProviderManager which implements the full interface