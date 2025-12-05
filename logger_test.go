package diffy

import (
	"strings"
	"testing"
)

func TestSimpleLogger_Logf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []any
		want   string
	}{
		{
			name:   "simple message",
			format: "Test message",
			args:   []any{},
			want:   "Test message",
		},
		{
			name:   "formatted message",
			format: "Resource %s has %d issues",
			args:   []any{"azurerm_virtual_network", 3},
			want:   "Resource azurerm_virtual_network has 3 issues",
		},
		{
			name:   "multiple arguments",
			format: "%s.%s missing attribute %s",
			args:   []any{"azurerm_virtual_network", "test", "location"},
			want:   "azurerm_virtual_network.test missing attribute location",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &SimpleLogger{}

			// Redirect output (we can't easily capture fmt.Printf, but we can test the interface)
			logger.Logf(tt.format, tt.args...)

			// Since SimpleLogger uses fmt.Printf, we just verify it doesn't panic
			// In a real scenario, you'd use a mock or capture stdout
		})
	}
}

func TestSimpleLogger_Interface(t *testing.T) {
	var _ Logger = (*SimpleLogger)(nil)

	logger := &SimpleLogger{}

	// Test that it implements the Logger interface
	logger.Logf("Test message")
	logger.Logf("Formatted: %s", "value")
	logger.Logf("Multiple: %s %d %v", "test", 42, true)
}

func TestLoggerUsage(t *testing.T) {
	logger := &SimpleLogger{}

	// Test various logging scenarios
	testCases := []struct {
		name  string
		logFn func()
	}{
		{
			name: "validation start",
			logFn: func() {
				logger.Logf("Starting validation for directory: %s", "/test/path")
			},
		},
		{
			name: "finding logged",
			logFn: func() {
				logger.Logf("[%s] %s.%s is missing %s", "Resource", "azurerm_vnet", "test", "location")
			},
		},
		{
			name: "completion message",
			logFn: func() {
				logger.Logf("Validation complete. Found %d issues", 5)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Just ensure logging doesn't panic
			tc.logFn()
		})
	}
}

// MockLogger for testing
type MockLogger struct {
	messages []string
}

func (m *MockLogger) Logf(format string, args ...any) {
	msg := format
	if len(args) > 0 {
		// Simple formatting for testing
		msg = strings.TrimSpace(format)
	}
	m.messages = append(m.messages, msg)
}

func TestMockLogger(t *testing.T) {
	mock := &MockLogger{
		messages: make([]string, 0),
	}

	mock.Logf("First message")
	mock.Logf("Second message: %s", "test")
	mock.Logf("Third message")

	if len(mock.messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(mock.messages))
	}

	if mock.messages[0] != "First message" {
		t.Errorf("First message = %q, want %q", mock.messages[0], "First message")
	}
}

func TestLoggerWithValidator(t *testing.T) {
	mock := &MockLogger{
		messages: make([]string, 0),
	}

	validator := NewSchemaValidator(mock)

	if validator.logger != mock {
		t.Error("Validator should use the provided logger")
	}

	// Verify logger is set
	mock.Logf("Validator created with logger")

	if len(mock.messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(mock.messages))
	}
}
