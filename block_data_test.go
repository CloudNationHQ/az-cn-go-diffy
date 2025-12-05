package diffy

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func TestNewBlockDataInitializesCollections(t *testing.T) {
	bd := NewBlockData()

	if bd.Properties == nil {
		t.Fatalf("Properties map should be initialized")
	}
	if bd.StaticBlocks == nil {
		t.Fatalf("StaticBlocks map should be initialized")
	}
	if bd.DynamicBlocks == nil {
		t.Fatalf("DynamicBlocks map should be initialized")
	}
	if bd.IgnoreChanges == nil {
		t.Fatalf("IgnoreChanges slice should be initialized")
	}
}

func TestParseAttributesAndBlocks(t *testing.T) {
	body := parseHCLBody(t, `
name          = "vnet"
address_space = ["10.0.0.0/16"]

lifecycle {
  ignore_changes = [tags, "location"]
}

dynamic "subnet" {
  for_each = var.subnets
  content {
    name           = each.key
    address_prefix = each.value
  }
}

subnet {
  name           = "subnet1"
  address_prefix = "10.0.1.0/24"
}
`)

	bd := NewBlockData()
	bd.ParseAttributes(body)
	bd.ParseBlocks(body)

	if diff := cmp.Diff(map[string]bool{
		"name":          true,
		"address_space": true,
		"subnet":        true,
	}, bd.Properties); diff != "" {
		t.Fatalf("Properties mismatch (-want +got):\n%s", diff)
	}

	if len(bd.StaticBlocks["subnet"]) != 1 {
		t.Fatalf("Expected 1 static subnet block, got %d", len(bd.StaticBlocks["subnet"]))
	}

	staticSubnet := bd.StaticBlocks["subnet"][0]
	if diff := cmp.Diff(map[string]bool{
		"name":           true,
		"address_prefix": true,
	}, staticSubnet.Data.Properties); diff != "" {
		t.Fatalf("Static subnet properties mismatch (-want +got):\n%s", diff)
	}

	dynamicSubnet := bd.DynamicBlocks["subnet"]
	if dynamicSubnet == nil {
		t.Fatalf("Dynamic subnet block should be parsed")
	}

	if diff := cmp.Diff(map[string]bool{
		"name":           true,
		"address_prefix": true,
	}, dynamicSubnet.Data.Properties); diff != "" {
		t.Fatalf("Dynamic subnet properties mismatch (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff([]string{"tags"}, bd.IgnoreChanges, cmpopts.SortSlices(func(a, b string) bool {
		return a < b
	})); diff != "" {
		t.Fatalf("IgnoreChanges mismatch (-want +got):\n%s", diff)
	}
}

func TestBlockDataValidateReportsMissingParts(t *testing.T) {
	schema := &SchemaBlock{
		Attributes: map[string]*SchemaAttribute{
			"name":     {Required: true},
			"location": {Required: true},
			"tags":     {Optional: true},
		},
		BlockTypes: map[string]*SchemaBlockType{
			"subnet": {
				MinItems: 1,
				Block: &SchemaBlock{
					Attributes: map[string]*SchemaAttribute{
						"name": {Required: true},
					},
					BlockTypes: map[string]*SchemaBlockType{},
				},
			},
		},
	}

	bd := BlockData{
		Properties:    map[string]bool{"location": true}, // missing "name"
		StaticBlocks:  map[string][]*ParsedBlock{},
		DynamicBlocks: map[string]*ParsedBlock{},
		IgnoreChanges: []string{"tags"},
	}

	var findings []ValidationFinding
	bd.Validate("azurerm_virtual_network", "azurerm_virtual_network.test", schema, nil, &findings)

	want := []ValidationFinding{
		{
			ResourceType: "azurerm_virtual_network",
			Path:         "azurerm_virtual_network.test",
			Name:         "name",
			Required:     true,
			IsBlock:      false,
		},
		{
			ResourceType: "azurerm_virtual_network",
			Path:         "azurerm_virtual_network.test",
			Name:         "subnet",
			Required:     true,
			IsBlock:      true,
		},
	}

	lessFinding := func(a, b ValidationFinding) bool {
		if a.Path == b.Path {
			return a.Name < b.Name
		}
		return a.Path < b.Path
	}

	if diff := cmp.Diff(want, findings, cmpopts.SortSlices(lessFinding)); diff != "" {
		t.Fatalf("Unexpected validation findings (-want +got):\n%s", diff)
	}
}

func TestMergeBlocksCombinesData(t *testing.T) {
	dest := &ParsedBlock{
		Data: BlockData{
			Properties: map[string]bool{"name": true},
			StaticBlocks: map[string][]*ParsedBlock{
				"tags": {{
					Data: BlockData{
						Properties: map[string]bool{"environment": true},
						StaticBlocks: map[string][]*ParsedBlock{
							"metadata": {},
						},
						DynamicBlocks: map[string]*ParsedBlock{},
						IgnoreChanges: []string{},
					},
				}},
			},
			DynamicBlocks: map[string]*ParsedBlock{
				"rule": {
					Data: BlockData{
						Properties:    map[string]bool{"priority": true},
						StaticBlocks:  map[string][]*ParsedBlock{},
						DynamicBlocks: map[string]*ParsedBlock{},
						IgnoreChanges: []string{},
					},
				},
			},
			IgnoreChanges: []string{"name"},
		},
	}

	src := &ParsedBlock{
		Data: BlockData{
			Properties: map[string]bool{"location": true},
			StaticBlocks: map[string][]*ParsedBlock{
				"tags": {{
					Data: BlockData{
						Properties:    map[string]bool{"costcenter": true},
						StaticBlocks:  map[string][]*ParsedBlock{},
						DynamicBlocks: map[string]*ParsedBlock{},
						IgnoreChanges: []string{},
					},
				}},
			},
			DynamicBlocks: map[string]*ParsedBlock{
				"rule": {
					Data: BlockData{
						Properties:    map[string]bool{"action": true},
						StaticBlocks:  map[string][]*ParsedBlock{},
						DynamicBlocks: map[string]*ParsedBlock{},
						IgnoreChanges: []string{},
					},
				},
			},
			IgnoreChanges: []string{"tags"},
		},
	}

	mergeBlocks(dest, src)

	if diff := cmp.Diff(map[string]bool{"name": true, "location": true}, dest.Data.Properties); diff != "" {
		t.Fatalf("Merged properties mismatch (-want +got):\n%s", diff)
	}

	if len(dest.Data.StaticBlocks["tags"]) != 2 {
		t.Fatalf("Expected 2 tag blocks after merge, got %d", len(dest.Data.StaticBlocks["tags"]))
	}

	if diff := cmp.Diff(map[string]bool{"priority": true, "action": true}, dest.Data.DynamicBlocks["rule"].Data.Properties); diff != "" {
		t.Fatalf("Merged dynamic block properties mismatch (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff([]string{"name", "tags"}, dest.Data.IgnoreChanges, cmpopts.SortSlices(func(a, b string) bool {
		return a < b
	})); diff != "" {
		t.Fatalf("IgnoreChanges not merged as expected (-want +got):\n%s", diff)
	}
}

func parseHCLBody(t *testing.T, src string) *hclsyntax.Body {
	t.Helper()

	file, diags := hclsyntax.ParseConfig([]byte(src), "test.hcl", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("failed to parse HCL: %v", diags)
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		t.Fatalf("unexpected body type %T", file.Body)
	}
	return body
}
