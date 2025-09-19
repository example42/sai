package validation

import (
	"fmt"
	"os"

	"github.com/xeipuuv/gojsonschema"
	"sai/internal/types"
)

// ProviderValidator validates provider data against the JSON schema
type ProviderValidator struct {
	schemaLoader gojsonschema.JSONLoader
}

// NewProviderValidator creates a new provider validator
func NewProviderValidator(schemaPath string) (*ProviderValidator, error) {
	// Load schema from file
	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	schemaLoader := gojsonschema.NewBytesLoader(schemaData)
	
	return &ProviderValidator{
		schemaLoader: schemaLoader,
	}, nil
}

// ValidateProvider validates a provider data structure against the schema
func (v *ProviderValidator) ValidateProvider(provider *types.ProviderData) error {
	// Convert provider to JSON for validation
	jsonData, err := provider.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert provider to JSON: %w", err)
	}

	// Create document loader
	documentLoader := gojsonschema.NewBytesLoader(jsonData)

	// Validate
	result, err := gojsonschema.Validate(v.schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if !result.Valid() {
		var errorMessages []string
		for _, desc := range result.Errors() {
			errorMessages = append(errorMessages, desc.String())
		}
		return fmt.Errorf("provider validation failed: %v", errorMessages)
	}

	return nil
}

// ValidateProviderYAML validates provider YAML data against the schema
func (v *ProviderValidator) ValidateProviderYAML(yamlData []byte) error {
	// First parse the YAML
	provider, err := types.LoadProviderFromYAML(yamlData)
	if err != nil {
		return fmt.Errorf("failed to parse provider YAML: %w", err)
	}

	// Then validate against schema
	return v.ValidateProvider(provider)
}

// ValidateProviderFile validates a provider file against the schema
func (v *ProviderValidator) ValidateProviderFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read provider file: %w", err)
	}

	return v.ValidateProviderYAML(data)
}

// ValidationResult contains validation results and details
type ValidationResult struct {
	Valid  bool
	Errors []string
	File   string
}

// ValidateAllProviders validates all provider files in a directory
func (v *ProviderValidator) ValidateAllProviders(providerDir string) ([]ValidationResult, error) {
	entries, err := os.ReadDir(providerDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read provider directory: %w", err)
	}

	var results []ValidationResult

	for _, entry := range entries {
		if entry.IsDir() || !isYAMLFile(entry.Name()) {
			continue
		}

		filePath := fmt.Sprintf("%s/%s", providerDir, entry.Name())
		result := ValidationResult{
			File: entry.Name(),
		}

		err := v.ValidateProviderFile(filePath)
		if err != nil {
			result.Valid = false
			result.Errors = []string{err.Error()}
		} else {
			result.Valid = true
		}

		results = append(results, result)
	}

	return results, nil
}

// isYAMLFile checks if a file has a YAML extension
func isYAMLFile(filename string) bool {
	return len(filename) > 5 && (filename[len(filename)-5:] == ".yaml" || filename[len(filename)-4:] == ".yml")
}

// GetValidationSummary returns a summary of validation results
func GetValidationSummary(results []ValidationResult) (int, int, []string) {
	var valid, invalid int
	var errors []string

	for _, result := range results {
		if result.Valid {
			valid++
		} else {
			invalid++
			for _, err := range result.Errors {
				errors = append(errors, fmt.Sprintf("%s: %s", result.File, err))
			}
		}
	}

	return valid, invalid, errors
}