package diffy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type HCLParser interface {
	ParseProviderRequirements(ctx context.Context, filename string) (map[string]ProviderConfig, error)
	ParseMainFile(ctx context.Context, filename string) ([]ParsedResource, []ParsedDataSource, error)
	ParseTerraformFiles(ctx context.Context, filenames []string) ([]ParsedResource, []ParsedDataSource, error)
}

type TerraformRunner interface {
	Init(ctx context.Context, dir string) error
	GetSchema(ctx context.Context, dir string) (*TerraformSchema, error)
}

type DefaultHCLParser struct{}

func NewHCLParser() *DefaultHCLParser {
	return &DefaultHCLParser{}
}

func (parser *DefaultHCLParser) ParseProviderRequirements(ctx context.Context, filename string) (map[string]ProviderConfig, error) {
	f, err := parser.parseHCLFile(filename)
	if err != nil {
		return nil, err
	}

	body, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return nil, &ParseError{
			File:    filename,
			Message: "invalid HCL body type",
		}
	}

	return parser.parseProviderRequirementsFromBody(body)
}

func (parser *DefaultHCLParser) ParseMainFile(ctx context.Context, filename string) ([]ParsedResource, []ParsedDataSource, error) {
	return parser.ParseTerraformFiles(ctx, []string{filename})
}

func (parser *DefaultHCLParser) ParseTerraformFiles(_ context.Context, files []string) ([]ParsedResource, []ParsedDataSource, error) {
	var allResources []ParsedResource
	var allDataSources []ParsedDataSource

	for _, filename := range files {
		f, err := parser.parseHCLFile(filename)
		if err != nil {
			return nil, nil, err
		}

		body, ok := f.Body.(*hclsyntax.Body)
		if !ok {
			return nil, nil, &ParseError{
				File:    filename,
				Message: "invalid HCL body type",
			}
		}

		resources, dataSources, err := parser.parseMainFileFromBody(body)
		if err != nil {
			return nil, nil, err
		}

		allResources = append(allResources, resources...)
		allDataSources = append(allDataSources, dataSources...)
	}

	return allResources, allDataSources, nil
}

func (parser *DefaultHCLParser) parseHCLFile(filename string) (*hcl.File, error) {
	hclParser := hclparse.NewParser()
	f, diags := hclParser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, &ParseError{
			File:    filename,
			Message: "failed to parse HCL file",
			Err:     fmt.Errorf("%v", diags),
		}
	}
	return f, nil
}

func (parser *DefaultHCLParser) parseProviderRequirementsFromBody(body *hclsyntax.Body) (map[string]ProviderConfig, error) {
	providers := make(map[string]ProviderConfig)
	for _, blk := range body.Blocks {
		if blk.Type == "terraform" {
			for _, innerBlk := range blk.Body.Blocks {
				if innerBlk.Type == "required_providers" {
					attrs, _ := innerBlk.Body.JustAttributes()
					for name, attr := range attrs {
						val, _ := attr.Expr.Value(nil)
						if val.Type().IsObjectType() {
							pc := ProviderConfig{}
							if val.Type().HasAttribute("source") {
								if sourceVal := val.GetAttr("source"); !sourceVal.IsNull() {
									pc.Source = NormalizeSource(sourceVal.AsString())
								}
							}
							if val.Type().HasAttribute("version") {
								if versionVal := val.GetAttr("version"); !versionVal.IsNull() {
									pc.Version = versionVal.AsString()
								}
							}
							if pc.Source == "" {
								pc.Source = NormalizeSource("hashicorp/" + name)
							}
							providers[name] = pc
						}
					}
				}
			}
		}
	}
	return providers, nil
}

func (parser *DefaultHCLParser) parseMainFileFromBody(body *hclsyntax.Body) ([]ParsedResource, []ParsedDataSource, error) {
	var resources []ParsedResource
	var dataSources []ParsedDataSource

	for _, blk := range body.Blocks {
		if blk.Type == "resource" && len(blk.Labels) >= 2 {
			parsed := ParseSyntaxBody(blk.Body)

			res := ParsedResource{
				Type: blk.Labels[0],
				Name: blk.Labels[1],
				Data: parsed.Data,
			}
			resources = append(resources, res)
		}

		if blk.Type == "data" && len(blk.Labels) >= 2 {
			parsed := ParseSyntaxBody(blk.Body)

			ds := ParsedDataSource{
				Type: blk.Labels[0],
				Name: blk.Labels[1],
				Data: parsed.Data,
			}
			dataSources = append(dataSources, ds)
		}
	}
	return resources, dataSources, nil
}

func ParseSyntaxBody(body *hclsyntax.Body) *ParsedBlock {
	bd := NewBlockData()
	bd.ParseAttributes(body)
	bd.ParseBlocks(body)
	return &ParsedBlock{Data: bd}
}

func extractIgnoreChangesFromValue(val cty.Value) []string {
	var changes []string
	if val.Type().IsCollectionType() {
		for it := val.ElementIterator(); it.Next(); {
			_, element := it.Element()
			if element.Type() == cty.String {
				change := element.AsString()
				if change == "all" {
					return []string{"*all*"}
				}
				changes = append(changes, change)
			}
		}
	}
	return changes
}

func NormalizeSource(source string) string {
	if strings.Contains(source, "/") && !strings.Contains(source, "registry.terraform.io/") {
		return "registry.terraform.io/" + source
	}
	return source
}

func FindSubmodules(modulesDir string) ([]SubModule, error) {
	var result []SubModule

	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		return result, nil
	}

	for _, e := range entries {
		if e.IsDir() {
			subName := e.Name()
			subPath := filepath.Join(modulesDir, subName)
			mainTf := filepath.Join(subPath, "main.tf")

			if _, err := os.Stat(mainTf); err == nil {
				result = append(result, SubModule{Name: subName, Path: subPath})
			}
		}
	}

	return result, nil
}
