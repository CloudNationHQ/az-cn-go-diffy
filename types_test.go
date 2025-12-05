package diffy

import (
	"errors"
	"testing"
)

func TestParseError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ParseError
		wantText []string
	}{
		{
			name: "error with nested error",
			err: &ParseError{
				File:    "main.tf",
				Message: "invalid syntax",
				Err:     errors.New("unexpected token"),
			},
			wantText: []string{"main.tf", "invalid syntax", "unexpected token"},
		},
		{
			name: "error without nested error",
			err: &ParseError{
				File:    "variables.tf",
				Message: "missing attribute",
			},
			wantText: []string{"variables.tf", "missing attribute"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()

			for _, want := range tt.wantText {
				if !contains(got, want) {
					t.Errorf("Error() = %q, should contain %q", got, want)
				}
			}
		})
	}
}

func TestParseError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	parseErr := &ParseError{
		File:    "test.tf",
		Message: "test error",
		Err:     innerErr,
	}

	unwrapped := parseErr.Unwrap()
	if unwrapped != innerErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, innerErr)
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		wantText []string
	}{
		{
			name: "error with nested error",
			err: &ValidationError{
				ResourceType: "azurerm_virtual_network",
				Message:      "missing required attribute",
				Err:          errors.New("location is required"),
			},
			wantText: []string{"azurerm_virtual_network", "missing required attribute", "location is required"},
		},
		{
			name: "error without nested error",
			err: &ValidationError{
				ResourceType: "azurerm_subnet",
				Message:      "invalid configuration",
			},
			wantText: []string{"azurerm_subnet", "invalid configuration"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()

			for _, want := range tt.wantText {
				if !contains(got, want) {
					t.Errorf("Error() = %q, should contain %q", got, want)
				}
			}
		})
	}
}

func TestValidationError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	validationErr := &ValidationError{
		ResourceType: "test_resource",
		Message:      "test error",
		Err:          innerErr,
	}

	unwrapped := validationErr.Unwrap()
	if unwrapped != innerErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, innerErr)
	}
}

func TestGitHubError(t *testing.T) {
	tests := []struct {
		name     string
		err      *GitHubError
		wantText []string
	}{
		{
			name: "error with nested error",
			err: &GitHubError{
				Operation: "create issue",
				Message:   "API request failed",
				Err:       errors.New("401 Unauthorized"),
			},
			wantText: []string{"create issue", "API request failed", "401 Unauthorized"},
		},
		{
			name: "error without nested error",
			err: &GitHubError{
				Operation: "update issue",
				Message:   "issue not found",
			},
			wantText: []string{"update issue", "issue not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()

			for _, want := range tt.wantText {
				if !contains(got, want) {
					t.Errorf("Error() = %q, should contain %q", got, want)
				}
			}
		})
	}
}

func TestGitHubError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	ghErr := &GitHubError{
		Operation: "test operation",
		Message:   "test error",
		Err:       innerErr,
	}

	unwrapped := ghErr.Unwrap()
	if unwrapped != innerErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, innerErr)
	}
}
