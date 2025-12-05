package diffy

import (
	"context"
	"path/filepath"
	"testing"
)

// Ensures the HCL in examples/module still parses and validates against a minimal schema.
func TestExampleModuleValidatesWithStubRunner(t *testing.T) {
	root := filepath.Join("examples", "module")

	parser := &exampleStubParser{providerSource: "registry.terraform.io/hashicorp/azurerm"}
	runner := &exampleStubRunner{
		schema: &TerraformSchema{
			ProviderSchemas: map[string]*ProviderSchema{
				parser.providerSource: {
					ResourceSchemas: map[string]*ResourceSchema{
						"azurerm_linux_function_app": {
							Block: &SchemaBlock{
								Attributes: map[string]*SchemaAttribute{},
								BlockTypes: map[string]*SchemaBlockType{},
							},
						},
					},
					DataSourceSchemas: map[string]*ResourceSchema{},
				},
			},
		},
	}

	findings, err := ValidateSchema(
		WithTerraformRoot(root),
		WithParser(parser),
		WithTerraformRunner(runner),
		func(opts *SchemaValidatorOptions) {
			opts.Silent = true
		},
	)
	if err != nil {
		t.Fatalf("ValidateSchema returned error: %v", err)
	}
	if len(findings) > 0 {
		t.Fatalf("example validation produced %d findings: %+v", len(findings), findings)
	}
}

type exampleStubParser struct {
	providerSource string
	resources      []ParsedResource
}

func (p *exampleStubParser) ParseProviderRequirements(_ context.Context, _ string) (map[string]ProviderConfig, error) {
	return map[string]ProviderConfig{
		"azurerm": {Source: p.providerSource},
	}, nil
}

func (p *exampleStubParser) ParseMainFile(ctx context.Context, filename string) ([]ParsedResource, []ParsedDataSource, error) {
	return p.ParseTerraformFiles(ctx, []string{filename})
}

func (p *exampleStubParser) ParseTerraformFiles(_ context.Context, _ []string) ([]ParsedResource, []ParsedDataSource, error) {
	return p.resources, nil, nil
}

type exampleStubRunner struct {
	schema *TerraformSchema
}

func (r *exampleStubRunner) Init(_ context.Context, _ string) error {
	return nil
}

func (r *exampleStubRunner) GetSchema(_ context.Context, _ string) (*TerraformSchema, error) {
	return r.schema, nil
}
