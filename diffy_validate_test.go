package diffy

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSchemaUsesInjectedParserAndRunner(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "main.tf"), "# test")

	parser := &validateStubParser{
		providerSource: "registry.terraform.io/hashicorp/azurerm",
		resources:      nil,
	}
	runner := &validateStubRunner{
		schema: &TerraformSchema{
			ProviderSchemas: map[string]*ProviderSchema{
				parser.providerSource: {
					ResourceSchemas:   map[string]*ResourceSchema{},
					DataSourceSchemas: map[string]*ResourceSchema{},
				},
			},
		},
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	orig := os.Stdout
	defer func() { os.Stdout = orig }()
	os.Stdout = w

	findings, err := ValidateSchema(
		WithTerraformRoot(root),
		WithParser(parser),
		WithTerraformRunner(runner),
		func(opts *SchemaValidatorOptions) {
			opts.Logger = &SimpleLogger{}
			opts.Silent = false
		},
	)
	if err != nil {
		t.Fatalf("ValidateSchema returned error: %v", err)
	}
	w.Close()
	var out bytes.Buffer
	if _, err := out.ReadFrom(r); err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %d", len(findings))
	}

	if got := out.String(); got == "" || !contains(got, "No validation findings.") {
		t.Fatalf("outputFindings should write success message, got %q", got)
	}
}

func TestCreateGitHubIssueRequiresToken(t *testing.T) {
	opts := &SchemaValidatorOptions{
		GitHubOwner: "owner",
		GitHubRepo:  "repo",
	}
	if err := createGitHubIssue(context.Background(), opts, nil); err == nil {
		t.Fatalf("expected token error")
	}
}

func TestValidateSchemaMissingRootReturnsError(t *testing.T) {
	if _, err := ValidateSchema(); err == nil {
		t.Fatalf("expected error when Terraform root not provided")
	}
}

func TestValidateSchemaUsesEnvRoot(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "main.tf"), "# test")
	t.Setenv("TERRAFORM_ROOT", root)

	parser := &validateStubParser{providerSource: "registry.terraform.io/hashicorp/azurerm"}
	runner := &validateStubRunner{
		schema: &TerraformSchema{
			ProviderSchemas: map[string]*ProviderSchema{
				parser.providerSource: {
					ResourceSchemas:   map[string]*ResourceSchema{},
					DataSourceSchemas: map[string]*ResourceSchema{},
				},
			},
		},
	}

	if _, err := ValidateSchema(
		WithParser(parser),
		WithTerraformRunner(runner),
	); err != nil {
		t.Fatalf("ValidateSchema returned error: %v", err)
	}
}

func TestValidateTerraformSchemaWithOptionsPropagatesInitError(t *testing.T) {
	parser := &validateStubParser{providerSource: "registry.terraform.io/hashicorp/azurerm"}
	runner := &failingRunner{}

	_, err := ValidateTerraformSchemaWithOptions(
		&SimpleLogger{},
		t.TempDir(),
		"",
		parser,
		runner,
		nil,
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "init failed") {
		t.Fatalf("expected init failure, got %v", err)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

// stubParser and stubRunner mirror the lightweight fakes in validator tests.
type validateStubParser struct {
	providerSource string
	resources      []ParsedResource
}

func (p *validateStubParser) ParseProviderRequirements(_ context.Context, _ string) (map[string]ProviderConfig, error) {
	return map[string]ProviderConfig{
		"azurerm": {Source: p.providerSource},
	}, nil
}

func (p *validateStubParser) ParseMainFile(ctx context.Context, filename string) ([]ParsedResource, []ParsedDataSource, error) {
	return p.ParseTerraformFiles(ctx, []string{filename})
}

func (p *validateStubParser) ParseTerraformFiles(_ context.Context, _ []string) ([]ParsedResource, []ParsedDataSource, error) {
	return p.resources, nil, nil
}

type validateStubRunner struct {
	schema *TerraformSchema
}

func (r *validateStubRunner) Init(_ context.Context, _ string) error {
	return nil
}

func (r *validateStubRunner) GetSchema(_ context.Context, _ string) (*TerraformSchema, error) {
	if r.schema == nil {
		return nil, fmt.Errorf("no schema")
	}
	return r.schema, nil
}

type failingRunner struct{}

func (f *failingRunner) Init(_ context.Context, _ string) error {
	return fmt.Errorf("init failed")
}

func (f *failingRunner) GetSchema(_ context.Context, _ string) (*TerraformSchema, error) {
	return nil, fmt.Errorf("should not be called")
}

func TestValidateSchemaAppliesEnvExcludedProperties(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "main.tf"), "# test")
	t.Setenv("EXCLUDED_PROPERTIES", "azurerm_key_vault.access_policy, azurerm_key_vault.tags")

	parser := &validateStubParser{
		providerSource: "registry.terraform.io/hashicorp/azurerm",
		resources: []ParsedResource{
			{Type: "azurerm_key_vault", Name: "this", Data: NewBlockData()},
		},
	}
	runner := &validateStubRunner{
		schema: &TerraformSchema{
			ProviderSchemas: map[string]*ProviderSchema{
				parser.providerSource: {
					ResourceSchemas: map[string]*ResourceSchema{
						"azurerm_key_vault": {
							Block: &SchemaBlock{
								Attributes: map[string]*SchemaAttribute{
									"tags":                          {Optional: true},
									"public_network_access_enabled": {Optional: true},
								},
								BlockTypes: map[string]*SchemaBlockType{
									"access_policy": {Block: &SchemaBlock{}},
								},
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
		func(opts *SchemaValidatorOptions) { opts.Silent = true },
	)
	if err != nil {
		t.Fatalf("ValidateSchema returned error: %v", err)
	}

	if len(findings) != 1 || findings[0].Name != "public_network_access_enabled" {
		t.Fatalf("expected only public_network_access_enabled finding, got %+v", findings)
	}
}
