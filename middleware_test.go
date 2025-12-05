package diffy

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestLoggingMiddleware_LogsErrors(t *testing.T) {
	logger := &stubLogger{}
	mw := LoggingMiddleware(logger)

	sentinel := errors.New("boom")
	findings, err := mw(nil, sentinel)

	if !errors.Is(err, sentinel) {
		t.Fatalf("expected error to be returned unchanged")
	}

	if len(findings) != 0 {
		t.Fatalf("expected findings to pass through unchanged, got %d items", len(findings))
	}

	if !logger.contains("Validation failed") {
		t.Fatalf("expected failure message to be logged, got %v", logger.messages)
	}
}

func TestLoggingMiddleware_LogsCounts(t *testing.T) {
	tests := []struct {
		name          string
		findings      []ValidationFinding
		wantSubstring string
	}{
		{
			name:          "no findings",
			findings:      nil,
			wantSubstring: "no issues found",
		},
		{
			name: "with findings",
			findings: []ValidationFinding{
				{ResourceType: "azurerm_virtual_network", Path: "test", Name: "name"},
				{ResourceType: "azurerm_virtual_network", Path: "test", Name: "subnet", IsBlock: true},
			},
			wantSubstring: "found 2 issues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &stubLogger{}
			mw := LoggingMiddleware(logger)

			findings, err := mw(tt.findings, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(findings) != len(tt.findings) {
				t.Fatalf("findings should pass through unchanged, got %d items", len(findings))
			}

			if !logger.contains(tt.wantSubstring) {
				t.Fatalf("expected log to contain %q, got %v", tt.wantSubstring, logger.messages)
			}
		})
	}
}

func TestApplyMiddlewareInvokesWrapper(t *testing.T) {
	logger := &stubLogger{}
	wrapped := LoggingMiddleware(logger)

	findings, err := ApplyMiddleware([]ValidationFinding{{Name: "name"}}, nil, wrapped)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) != 1 || findings[0].Name != "name" {
		t.Fatalf("findings should be returned unchanged, got %#v", findings)
	}

	if len(logger.messages) == 0 {
		t.Fatalf("middleware should log through provided logger")
	}
}

type stubLogger struct {
	messages []string
}

func (l *stubLogger) Logf(format string, args ...any) {
	l.messages = append(l.messages, fmt.Sprintf(format, args...))
}

func (l *stubLogger) contains(substr string) bool {
	for _, m := range l.messages {
		if strings.Contains(m, substr) {
			return true
		}
	}
	return false
}

func TestParsedDataSource_Structure(t *testing.T) {
	pds := ParsedDataSource{
		Type: "azurerm_resource_group",
		Name: "existing",
		Data: BlockData{
			Properties:    map[string]bool{"name": true},
			StaticBlocks:  make(map[string][]*ParsedBlock),
			DynamicBlocks: make(map[string]*ParsedBlock),
			IgnoreChanges: []string{},
		},
	}

	if pds.Type != "azurerm_resource_group" {
		t.Errorf("Type = %s, want azurerm_resource_group", pds.Type)
	}

	if pds.Name != "existing" {
		t.Errorf("Name = %s, want existing", pds.Name)
	}

	if !pds.Data.Properties["name"] {
		t.Error("Data should have 'name' property")
	}
}

func TestSubModule_Structure(t *testing.T) {
	sm := SubModule{
		Name: "network",
		Path: "/path/to/modules/network",
	}

	if sm.Name != "network" {
		t.Errorf("Name = %s, want network", sm.Name)
	}

	if sm.Path != "/path/to/modules/network" {
		t.Errorf("Path = %s, want /path/to/modules/network", sm.Path)
	}
}
