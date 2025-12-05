package diffy

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDeduplicateFindings(t *testing.T) {
	tests := []struct {
		name      string
		findings  []ValidationFinding
		wantCount int
	}{
		{
			name: "no duplicates",
			findings: []ValidationFinding{
				{
					ResourceType: "azurerm_virtual_network",
					Name:         "test1",
					Path:         "location",
					Required:     true,
				},
				{
					ResourceType: "azurerm_virtual_network",
					Name:         "test2",
					Path:         "name",
					Required:     true,
				},
			},
			wantCount: 2,
		},
		{
			name: "exact duplicates",
			findings: []ValidationFinding{
				{
					ResourceType: "azurerm_virtual_network",
					Name:         "test",
					Path:         "location",
					Required:     true,
				},
				{
					ResourceType: "azurerm_virtual_network",
					Name:         "test",
					Path:         "location",
					Required:     true,
				},
			},
			wantCount: 1,
		},
		{
			name: "mixed duplicates and unique",
			findings: []ValidationFinding{
				{
					ResourceType: "azurerm_virtual_network",
					Name:         "test1",
					Path:         "location",
					Required:     true,
				},
				{
					ResourceType: "azurerm_virtual_network",
					Name:         "test1",
					Path:         "location",
					Required:     true,
				},
				{
					ResourceType: "azurerm_virtual_network",
					Name:         "test2",
					Path:         "name",
					Required:     true,
				},
			},
			wantCount: 2,
		},
		{
			name:      "empty findings",
			findings:  []ValidationFinding{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeduplicateFindings(tt.findings)
			if len(got) != tt.wantCount {
				t.Errorf("DeduplicateFindings() returned %d findings, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestFormatFinding(t *testing.T) {
	tests := []struct {
		name         string
		finding      ValidationFinding
		wantContains []string
	}{
		{
			name: "resource attribute",
			finding: ValidationFinding{
				ResourceType: "azurerm_virtual_network",
				Name:         "test",
				Path:         "location",
				Required:     true,
				IsBlock:      false,
			},
			wantContains: []string{"azurerm_virtual_network", "test", "location"},
		},
		{
			name: "resource block",
			finding: ValidationFinding{
				ResourceType: "azurerm_virtual_network",
				Name:         "test",
				Path:         "subnet",
				Required:     true,
				IsBlock:      true,
			},
			wantContains: []string{"azurerm_virtual_network", "test", "subnet"},
		},
		{
			name: "data source attribute",
			finding: ValidationFinding{
				ResourceType: "azurerm_virtual_network",
				Name:         "existing",
				Path:         "name",
				Required:     true,
				IsDataSource: true,
			},
			wantContains: []string{"azurerm_virtual_network", "existing", "name"},
		},
		{
			name: "submodule data source",
			finding: ValidationFinding{
				ResourceType:  "azurerm_storage_account",
				Name:          "account",
				Path:          "root.block",
				Required:      false,
				IsBlock:       true,
				IsDataSource:  true,
				SubmoduleName: "network",
			},
			wantContains: []string{"azurerm_storage_account", "block", "submodule network", "data source"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatFinding(tt.finding)

			for _, want := range tt.wantContains {
				if !contains(got, want) {
					t.Errorf("FormatFinding() = %q, should contain %q", got, want)
				}
			}
		})
	}
}

func TestFilterResources(t *testing.T) {
	tests := []struct {
		name      string
		resources []ParsedResource
		excluded  []string
		wantCount int
	}{
		{
			name: "no exclusions",
			resources: []ParsedResource{
				{Type: "azurerm_virtual_network", Name: "test1"},
				{Type: "azurerm_subnet", Name: "test1"},
			},
			excluded:  []string{},
			wantCount: 2,
		},
		{
			name: "exclude one type",
			resources: []ParsedResource{
				{Type: "azurerm_virtual_network", Name: "test1"},
				{Type: "azurerm_subnet", Name: "test1"},
			},
			excluded:  []string{"azurerm_subnet"},
			wantCount: 1,
		},
		{
			name: "exclude all",
			resources: []ParsedResource{
				{Type: "azurerm_virtual_network", Name: "test1"},
				{Type: "azurerm_subnet", Name: "test1"},
			},
			excluded:  []string{"azurerm_virtual_network", "azurerm_subnet"},
			wantCount: 0,
		},
		{
			name: "exact match exclusion",
			resources: []ParsedResource{
				{Type: "azurerm_virtual_network", Name: "test1"},
				{Type: "azurerm_subnet", Name: "test1"},
				{Type: "azurerm_network_security_group", Name: "test1"},
			},
			excluded:  []string{"azurerm_virtual_network"},
			wantCount: 2, // subnet and nsg remain
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterResources(tt.resources, tt.excluded)

			if len(filtered) != tt.wantCount {
				t.Errorf("filterResources() returned %d resources, want %d", len(filtered), tt.wantCount)
			}
		})
	}
}

func TestFilterDataSources(t *testing.T) {
	tests := []struct {
		name        string
		dataSources []ParsedDataSource
		excluded    []string
		wantCount   int
	}{
		{
			name: "no exclusions",
			dataSources: []ParsedDataSource{
				{Type: "azurerm_virtual_network", Name: "existing1"},
				{Type: "azurerm_subnet", Name: "existing1"},
			},
			excluded:  []string{},
			wantCount: 2,
		},
		{
			name: "exclude one type",
			dataSources: []ParsedDataSource{
				{Type: "azurerm_virtual_network", Name: "existing1"},
				{Type: "azurerm_subnet", Name: "existing1"},
			},
			excluded:  []string{"azurerm_subnet"},
			wantCount: 1,
		},
		{
			name: "exclude all",
			dataSources: []ParsedDataSource{
				{Type: "azurerm_virtual_network", Name: "existing1"},
				{Type: "azurerm_subnet", Name: "existing1"},
			},
			excluded:  []string{"azurerm_virtual_network", "azurerm_subnet"},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterDataSources(tt.dataSources, tt.excluded)

			if len(filtered) != tt.wantCount {
				t.Errorf("filterDataSources() returned %d data sources, want %d", len(filtered), tt.wantCount)
			}
		})
	}
}

func TestValidateEntitiesMissingProviderOrSchema(t *testing.T) {
	validator := NewSchemaValidator(&SimpleLogger{})

	findings := validator.validateEntities(
		[]ParsedResource{{Type: "azurerm_virtual_network", Name: "test", Data: BlockData{}}},
		TerraformSchema{},
		map[string]ProviderConfig{}, // missing provider config, should log and skip
		".",
		"",
		false,
	)

	if len(findings) != 0 {
		t.Fatalf("expected no findings when provider config missing, got %d", len(findings))
	}

	// Provider exists but schema missing
	findings = validator.validateEntities(
		[]ParsedResource{{Type: "azurerm_virtual_network", Name: "test", Data: BlockData{}}},
		TerraformSchema{
			ProviderSchemas: map[string]*ProviderSchema{
				"registry.terraform.io/hashicorp/azurerm": {
					ResourceSchemas:   map[string]*ResourceSchema{},
					DataSourceSchemas: map[string]*ResourceSchema{},
				},
			},
		},
		map[string]ProviderConfig{"azurerm": {Source: "registry.terraform.io/hashicorp/azurerm"}},
		".",
		"",
		false,
	)

	if len(findings) != 0 {
		t.Fatalf("expected no findings when schema missing, got %d", len(findings))
	}
}

func TestValidateBlocksMultipleStaticAndDynamic(t *testing.T) {
	schema := &SchemaBlock{
		Attributes: map[string]*SchemaAttribute{},
		BlockTypes: map[string]*SchemaBlockType{
			"child": {
				MinItems: 1,
				Block: &SchemaBlock{
					Attributes: map[string]*SchemaAttribute{
						"required_attr": {Required: true},
					},
					BlockTypes: map[string]*SchemaBlockType{},
				},
			},
		},
	}

	bd := BlockData{
		Properties:   map[string]bool{},
		StaticBlocks: map[string][]*ParsedBlock{},
		DynamicBlocks: map[string]*ParsedBlock{
			"child": {
				Data: BlockData{
					Properties:    map[string]bool{}, // missing required_attr
					StaticBlocks:  map[string][]*ParsedBlock{},
					DynamicBlocks: map[string]*ParsedBlock{},
					IgnoreChanges: nil,
				},
			},
		},
	}

	// Add multiple static children to exercise indexed paths
	bd.StaticBlocks["child"] = []*ParsedBlock{
		{Data: BlockData{Properties: map[string]bool{}}},
		{Data: BlockData{Properties: map[string]bool{}}},
	}

	var findings []ValidationFinding
	bd.validateBlocks("azurerm_virtual_network", "root", schema, nil, &findings)

	if len(findings) != 3 {
		t.Fatalf("expected findings for each static and dynamic child, got %d", len(findings))
	}

	paths := map[string]int{}
	for _, f := range findings {
		paths[f.Path]++
		if f.Name != "required_attr" {
			t.Fatalf("unexpected finding: %+v", f)
		}
	}

	if paths["root.child"] != 1 || paths["root.child[0]"] != 1 || paths["root.child[1]"] != 1 {
		t.Fatalf("expected indexed paths for static blocks and plain path for dynamic, got %v", paths)
	}
}

func TestValidateTerraformSchemaSuccess(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "main.tf"), "# stub")

	parser := &stubParser{
		providerSource: "registry.terraform.io/hashicorp/azurerm",
		resources: []ParsedResource{
			{
				Type: "azurerm_virtual_network",
				Name: "test",
				Data: BlockData{
					Properties: map[string]bool{"location": true},
					StaticBlocks: map[string][]*ParsedBlock{
						"subnet": {{
							Data: BlockData{
								Properties:    map[string]bool{"name": true},
								StaticBlocks:  map[string][]*ParsedBlock{},
								DynamicBlocks: map[string]*ParsedBlock{},
								IgnoreChanges: nil,
							},
						}},
					},
					DynamicBlocks: map[string]*ParsedBlock{},
					IgnoreChanges: nil,
				},
			},
		},
	}

	runner := &stubRunner{
		schema: &TerraformSchema{
			ProviderSchemas: map[string]*ProviderSchema{
				"registry.terraform.io/hashicorp/azurerm": {
					ResourceSchemas: map[string]*ResourceSchema{
						"azurerm_virtual_network": {
							Block: &SchemaBlock{
								Attributes: map[string]*SchemaAttribute{
									"location": {Required: true},
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
							},
						},
					},
					DataSourceSchemas: map[string]*ResourceSchema{},
				},
			},
		},
	}

	findings, err := ValidateTerraformSchema(&SimpleLogger{}, dir, "", parser, runner)
	if err != nil {
		t.Fatalf("ValidateTerraformSchema returned error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %d", len(findings))
	}
}

func TestValidateTerraformSchemaInitError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "main.tf"), "# stub")

	parser := &stubParser{
		providerSource: "registry.terraform.io/hashicorp/azurerm",
		resources:      nil,
	}

	runner := &stubRunner{schema: &TerraformSchema{}}
	runnerFail := &failingRunner{}

	if _, err := ValidateTerraformSchema(&SimpleLogger{}, dir, "", parser, runnerFail); err == nil {
		t.Fatalf("expected init error")
	}

	// Ensure success path still works with stub runner
	if _, err := ValidateTerraformSchema(&SimpleLogger{}, dir, "", parser, runner); err != nil {
		t.Fatalf("unexpected error with stub runner: %v", err)
	}
}

func TestWalkTerraformFiles(t *testing.T) {
	dir := t.TempDir()

	tfA := filepath.Join(dir, "a.tf")
	tfB := filepath.Join(dir, "b.tf")
	other := filepath.Join(dir, "notes.txt")

	writeFile(t, tfA, "resource \"x\" \"a\" {}")
	writeFile(t, tfB, "resource \"x\" \"b\" {}")
	writeFile(t, other, "# not terraform")

	files, err := walkTerraformFiles(dir)
	if err != nil {
		t.Fatalf("walkTerraformFiles returned error: %v", err)
	}

	want := []string{tfA, tfB}
	if len(files) != len(want) {
		t.Fatalf("walkTerraformFiles returned %d files, want %d", len(files), len(want))
	}

	for i := range want {
		if files[i] != want[i] {
			t.Fatalf("walkTerraformFiles order mismatch at %d: got %s want %s", i, files[i], want[i])
		}
	}
}

func TestWalkTerraformFilesMissingDir(t *testing.T) {
	_, err := walkTerraformFiles(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatalf("expected error for missing directory")
	}
}

func TestValidateProjectWithSubmodules(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main.tf"), "# root")

	modulesDir := filepath.Join(root, "modules", "network")
	if err := os.MkdirAll(modulesDir, 0o755); err != nil {
		t.Fatalf("failed to create module dir: %v", err)
	}
	writeFile(t, filepath.Join(modulesDir, "main.tf"), "# module")

	parser := &stubParser{
		providerSource: "registry.terraform.io/hashicorp/azurerm",
		resources: []ParsedResource{
			{
				Type: "azurerm_virtual_network",
				Name: "test",
				Data: BlockData{
					Properties:    map[string]bool{}, // missing required attribute to trigger finding
					StaticBlocks:  make(map[string][]*ParsedBlock),
					DynamicBlocks: make(map[string]*ParsedBlock),
					IgnoreChanges: nil,
				},
			},
		},
	}

	runner := &stubRunner{
		schema: &TerraformSchema{
			ProviderSchemas: map[string]*ProviderSchema{
				"registry.terraform.io/hashicorp/azurerm": {
					ResourceSchemas: map[string]*ResourceSchema{
						"azurerm_virtual_network": {
							Block: &SchemaBlock{
								Attributes: map[string]*SchemaAttribute{
									"location": {Required: true},
								},
								BlockTypes: map[string]*SchemaBlockType{},
							},
						},
					},
					DataSourceSchemas: map[string]*ResourceSchema{},
				},
			},
		},
	}

	opts := &SchemaValidatorOptions{
		TerraformRoot:   root,
		Logger:          &SimpleLogger{},
		Silent:          true,
		Parser:          parser,
		TerraformRunner: runner,
	}

	findings, err := validateProject(opts)
	if err != nil {
		t.Fatalf("validateProject returned error: %v", err)
	}

	if len(findings) != 2 {
		t.Fatalf("expected findings for root and submodule, got %d", len(findings))
	}

	submodules := map[string]int{}
	for _, f := range findings {
		submodules[f.SubmoduleName]++
		if f.Name != "location" || !f.Required {
			t.Fatalf("unexpected finding: %+v", f)
		}
	}

	if submodules[""] != 1 || submodules["network"] != 1 {
		t.Fatalf("findings should include one root and one network submodule entry, got %v", submodules)
	}

	if !runner.initCalled(root) || !runner.initCalled(modulesDir) {
		t.Fatalf("runner.Init should be invoked for root and submodule, got %v", runner.inited)
	}
}

type stubParser struct {
	providerSource string
	resources      []ParsedResource
}

func (p *stubParser) ParseProviderRequirements(_ context.Context, _ string) (map[string]ProviderConfig, error) {
	return map[string]ProviderConfig{
		"azurerm": {Source: p.providerSource},
	}, nil
}

func (p *stubParser) ParseMainFile(ctx context.Context, filename string) ([]ParsedResource, []ParsedDataSource, error) {
	return p.ParseTerraformFiles(ctx, []string{filename})
}

func (p *stubParser) ParseTerraformFiles(_ context.Context, _ []string) ([]ParsedResource, []ParsedDataSource, error) {
	return p.resources, nil, nil
}

type stubRunner struct {
	schema *TerraformSchema
	inited map[string]bool
}

func (r *stubRunner) Init(_ context.Context, dir string) error {
	if r.inited == nil {
		r.inited = make(map[string]bool)
	}
	r.inited[dir] = true
	return nil
}

func (r *stubRunner) GetSchema(_ context.Context, _ string) (*TerraformSchema, error) {
	return r.schema, nil
}

func (r *stubRunner) initCalled(dir string) bool {
	return r.inited != nil && r.inited[dir]
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

func TestNewSchemaValidator(t *testing.T) {
	logger := &SimpleLogger{}
	validator := NewSchemaValidator(logger)

	if validator == nil {
		t.Fatal("NewSchemaValidator() should not return nil")
	}

	if validator.logger != logger {
		t.Error("NewSchemaValidator() should set the logger")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
