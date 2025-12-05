package diffy

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestNormalizeSource(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "full source path",
			source: "registry.terraform.io/hashicorp/azurerm",
			want:   "registry.terraform.io/hashicorp/azurerm",
		},
		{
			name:   "short source path",
			source: "hashicorp/azurerm",
			want:   "registry.terraform.io/hashicorp/azurerm",
		},
		{
			name:   "single name without slash",
			source: "azurerm",
			want:   "azurerm", // No slash, returns as-is
		},
		{
			name:   "custom registry with slash",
			source: "custom.registry.io/myorg/myprovider",
			want:   "registry.terraform.io/custom.registry.io/myorg/myprovider", // Has slash, gets prefixed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeSource(tt.source)
			if got != tt.want {
				t.Errorf("NormalizeSource(%q) = %q, want %q", tt.source, got, tt.want)
			}
		})
	}
}

func TestFindSubmodules(t *testing.T) {
	tests := []struct {
		name      string
		structure map[string]bool // path -> isDir
		wantCount int
		wantErr   bool
	}{
		{
			name: "single submodule",
			structure: map[string]bool{
				"network":         true,
				"network/main.tf": false,
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "multiple submodules",
			structure: map[string]bool{
				"network":         true,
				"network/main.tf": false,
				"storage":         true,
				"storage/main.tf": false,
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "no submodules",
			structure: map[string]bool{},
			wantCount: 0,
			wantErr:   false, // FindSubmodules returns nil error even if directory doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			for path, isDir := range tt.structure {
				fullPath := filepath.Join(tmpDir, path)
				if isDir {
					if err := os.MkdirAll(fullPath, 0755); err != nil {
						t.Fatalf("Failed to create directory %s: %v", path, err)
					}
				} else {
					dir := filepath.Dir(fullPath)
					if err := os.MkdirAll(dir, 0755); err != nil {
						t.Fatalf("Failed to create parent directory for %s: %v", path, err)
					}
					if err := os.WriteFile(fullPath, []byte("# test file"), 0644); err != nil {
						t.Fatalf("Failed to create file %s: %v", path, err)
					}
				}
			}

			got, err := FindSubmodules(tmpDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindSubmodules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(got) != tt.wantCount {
				t.Errorf("FindSubmodules() returned %d submodules, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestExtractIgnoreChangesFromValue(t *testing.T) {
	val := cty.ListVal([]cty.Value{
		cty.StringVal("name"),
		cty.StringVal("all"),
		cty.StringVal("tags"),
	})

	got := extractIgnoreChangesFromValue(val)

	if len(got) != 1 || got[0] != "*all*" {
		t.Fatalf("extractIgnoreChangesFromValue returned %v, want [*all*]", got)
	}
}

func TestParseProviderRequirements(t *testing.T) {
	tests := []struct {
		name        string
		tfContent   string
		wantErr     bool
		wantCount   int
		checkResult func(*testing.T, map[string]ProviderConfig)
	}{
		{
			name: "single provider",
			tfContent: `
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 3.0.0"
    }
  }
}
`,
			wantErr:   false,
			wantCount: 1,
			checkResult: func(t *testing.T, providers map[string]ProviderConfig) {
				if provider, ok := providers["azurerm"]; ok {
					if provider.Source != "registry.terraform.io/hashicorp/azurerm" {
						t.Errorf("Source = %s, want registry.terraform.io/hashicorp/azurerm", provider.Source)
					}
					if provider.Version != ">= 3.0.0" {
						t.Errorf("Version = %s, want >= 3.0.0", provider.Version)
					}
				} else {
					t.Error("azurerm provider not found")
				}
			},
		},
		{
			name: "multiple providers",
			tfContent: `
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.1"
    }
  }
}
`,
			wantErr:   false,
			wantCount: 2,
			checkResult: func(t *testing.T, providers map[string]ProviderConfig) {
				if len(providers) != 2 {
					t.Errorf("Got %d providers, want 2", len(providers))
				}
			},
		},
		{
			name: "no terraform block",
			tfContent: `
resource "azurerm_virtual_network" "test" {
  name = "test"
}
`,
			wantErr:   false,
			wantCount: 0,
			checkResult: func(t *testing.T, providers map[string]ProviderConfig) {
				if len(providers) != 0 {
					t.Errorf("Got %d providers, want 0", len(providers))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tfFile := filepath.Join(tmpDir, "main.tf")
			if err := os.WriteFile(tfFile, []byte(tt.tfContent), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			parser := NewHCLParser()
			ctx := context.Background()
			got, err := parser.ParseProviderRequirements(ctx, tfFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProviderRequirements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != tt.wantCount {
					t.Errorf("Got %d providers, want %d", len(got), tt.wantCount)
				}
				if tt.checkResult != nil {
					tt.checkResult(t, got)
				}
			}
		})
	}
}

func TestParseTerraformFiles(t *testing.T) {
	tests := []struct {
		name            string
		files           map[string]string
		wantResources   int
		wantDataSources int
		wantErr         bool
	}{
		{
			name: "single resource",
			files: map[string]string{
				"main.tf": `
resource "azurerm_virtual_network" "test" {
  name                = "test-vnet"
  resource_group_name = "test-rg"
  location            = "westeurope"
  address_space       = ["10.0.0.0/16"]
}
`,
			},
			wantResources:   1,
			wantDataSources: 0,
			wantErr:         false,
		},
		{
			name: "resources and data sources",
			files: map[string]string{
				"main.tf": `
resource "azurerm_virtual_network" "test" {
  name = "test"
}

data "azurerm_resource_group" "existing" {
  name = "existing-rg"
}
`,
			},
			wantResources:   1,
			wantDataSources: 1,
			wantErr:         false,
		},
		{
			name: "multiple files",
			files: map[string]string{
				"main.tf": `
resource "azurerm_virtual_network" "test1" {
  name = "test1"
}
`,
				"data.tf": `
data "azurerm_resource_group" "existing" {
  name = "existing"
}
`,
			},
			wantResources:   1,
			wantDataSources: 1,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			var fileList []string

			for filename, content := range tt.files {
				tfFile := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(tfFile, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to create test file %s: %v", filename, err)
				}
				fileList = append(fileList, tfFile)
			}

			parser := NewHCLParser()
			ctx := context.Background()
			resources, dataSources, err := parser.ParseTerraformFiles(ctx, fileList)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTerraformFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(resources) != tt.wantResources {
					t.Errorf("Got %d resources, want %d", len(resources), tt.wantResources)
				}
				if len(dataSources) != tt.wantDataSources {
					t.Errorf("Got %d data sources, want %d", len(dataSources), tt.wantDataSources)
				}
			}
		})
	}
}

func TestParseMainFile(t *testing.T) {
	tmpDir := t.TempDir()
	tfFile := filepath.Join(tmpDir, "main.tf")

	content := `
resource "azurerm_virtual_network" "test" {
  name = "test"
}
`
	if err := os.WriteFile(tfFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewHCLParser()
	ctx := context.Background()
	resources, dataSources, err := parser.ParseMainFile(ctx, tfFile)

	if err != nil {
		t.Errorf("ParseMainFile() error = %v", err)
	}

	if len(resources) != 1 {
		t.Errorf("Got %d resources, want 1", len(resources))
	}

	if len(dataSources) != 0 {
		t.Errorf("Got %d data sources, want 0", len(dataSources))
	}
}

func TestParseHCLFile_InvalidFile(t *testing.T) {
	parser := NewHCLParser()
	ctx := context.Background()

	_, _, err := parser.ParseMainFile(ctx, "/non/existent/file.tf")
	if err == nil {
		t.Error("ParseMainFile() should return error for non-existent file")
	}
}

func TestParseHCLFile_InvalidSyntax(t *testing.T) {
	tmpDir := t.TempDir()
	tfFile := filepath.Join(tmpDir, "invalid.tf")

	content := `resource "test" {`
	if err := os.WriteFile(tfFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewHCLParser()
	ctx := context.Background()

	_, _, err := parser.ParseMainFile(ctx, tfFile)
	if err == nil {
		t.Error("ParseMainFile() should return error for invalid HCL syntax")
	}
}
