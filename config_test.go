package diffy

import (
	"testing"
)

func TestWithTerraformRoot(t *testing.T) {
	opts := &SchemaValidatorOptions{}
	path := "/test/path"

	WithTerraformRoot(path)(opts)

	if opts.TerraformRoot != path {
		t.Errorf("WithTerraformRoot() = %s, want %s", opts.TerraformRoot, path)
	}
}

func TestWithGitHubIssueCreation(t *testing.T) {
	opts := &SchemaValidatorOptions{}

	WithGitHubIssueCreation()(opts)

	if !opts.CreateGitHubIssue {
		t.Error("WithGitHubIssueCreation() should set CreateGitHubIssue to true")
	}
}

func TestWithExcludedResources(t *testing.T) {
	tests := []struct {
		name      string
		resources []string
		wantLen   int
	}{
		{
			name:      "single resource",
			resources: []string{"azurerm_resource_group"},
			wantLen:   1,
		},
		{
			name:      "multiple resources",
			resources: []string{"azurerm_resource_group", "azurerm_virtual_network"},
			wantLen:   2,
		},
		{
			name:      "empty",
			resources: []string{},
			wantLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &SchemaValidatorOptions{}

			WithExcludedResources(tt.resources...)(opts)

			if len(opts.ExcludedResources) != tt.wantLen {
				t.Errorf("ExcludedResources length = %d, want %d", len(opts.ExcludedResources), tt.wantLen)
			}
		})
	}
}

func TestWithExcludedDataSources(t *testing.T) {
	tests := []struct {
		name        string
		dataSources []string
		wantLen     int
	}{
		{
			name:        "single data source",
			dataSources: []string{"azurerm_resource_group"},
			wantLen:     1,
		},
		{
			name:        "multiple data sources",
			dataSources: []string{"azurerm_resource_group", "azurerm_virtual_network"},
			wantLen:     2,
		},
		{
			name:        "empty",
			dataSources: []string{},
			wantLen:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &SchemaValidatorOptions{}

			WithExcludedDataSources(tt.dataSources...)(opts)

			if len(opts.ExcludedDataSources) != tt.wantLen {
				t.Errorf("ExcludedDataSources length = %d, want %d", len(opts.ExcludedDataSources), tt.wantLen)
			}
		})
	}
}

func TestOptionChaining(t *testing.T) {
	opts := &SchemaValidatorOptions{}

	// Apply multiple options
	WithTerraformRoot("/test")(opts)
	WithExcludedResources("resource1", "resource2")(opts)
	WithExcludedDataSources("data1")(opts)
	WithGitHubIssueCreation()(opts)
	customParser := &DefaultHCLParser{}
	customRunner := &DefaultTerraformRunner{}
	WithParser(customParser)(opts)
	WithTerraformRunner(customRunner)(opts)

	if opts.TerraformRoot != "/test" {
		t.Error("TerraformRoot not set correctly")
	}

	if len(opts.ExcludedResources) != 2 {
		t.Errorf("ExcludedResources length = %d, want 2", len(opts.ExcludedResources))
	}

	if len(opts.ExcludedDataSources) != 1 {
		t.Errorf("ExcludedDataSources length = %d, want 1", len(opts.ExcludedDataSources))
	}

	if !opts.CreateGitHubIssue {
		t.Error("CreateGitHubIssue not set correctly")
	}

	if opts.Parser != customParser {
		t.Error("Parser not set correctly")
	}

	if opts.TerraformRunner != customRunner {
		t.Error("TerraformRunner not set correctly")
	}
}
