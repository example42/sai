package cli

import (
	"sai/internal/interfaces"
	"sai/internal/types"
)

// MockTemplateEngine is a placeholder template engine implementation
type MockTemplateEngine struct{}

func (m *MockTemplateEngine) Render(templateStr string, context *interfaces.TemplateContext) (string, error) {
	// Simple placeholder - just return the template string as-is
	return templateStr, nil
}

func (m *MockTemplateEngine) ValidateTemplate(templateStr string) error {
	return nil
}

func (m *MockTemplateEngine) SetSafetyMode(enabled bool) {
	// No-op for mock
}

func (m *MockTemplateEngine) SetSaidata(saidata *types.SoftwareData) {
	// No-op for mock
}

// MockLogger is a placeholder logger implementation
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields ...interfaces.LogField) {
	// No-op for mock
}

func (m *MockLogger) Info(msg string, fields ...interfaces.LogField) {
	// No-op for mock
}

func (m *MockLogger) Warn(msg string, fields ...interfaces.LogField) {
	// No-op for mock
}

func (m *MockLogger) Error(msg string, err error, fields ...interfaces.LogField) {
	// No-op for mock
}

func (m *MockLogger) Fatal(msg string, err error, fields ...interfaces.LogField) {
	// No-op for mock
}

func (m *MockLogger) WithFields(fields ...interfaces.LogField) interfaces.Logger {
	return m
}

func (m *MockLogger) SetLevel(level interfaces.LogLevel) {
	// No-op for mock
}

func (m *MockLogger) GetLevel() interfaces.LogLevel {
	return interfaces.LogLevelInfo
}