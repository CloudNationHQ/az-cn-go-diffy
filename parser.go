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

	return parser.parseMainFileFromBody(body)
}

// parseHCLFile is a helper function that handles common HCL file parsing with error handling
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
							if sourceVal := val.GetAttr("source"); !sourceVal.IsNull() {
								pc.Source = NormalizeSource(sourceVal.AsString())
							}
							if versionVal := val.GetAttr("version"); !versionVal.IsNull() {
								pc.Version = versionVal.AsString()
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

			ignoreChanges := extractLifecycleIgnoreChangesFromAST(blk.Body)
			if len(ignoreChanges) > 0 {
				parsed.Data.IgnoreChanges = append(parsed.Data.IgnoreChanges, ignoreChanges...)
			}

			res := ParsedResource{
				Type: blk.Labels[0],
				Name: blk.Labels[1],
				Data: parsed.Data,
			}
			resources = append(resources, res)
		}

		if blk.Type == "data" && len(blk.Labels) >= 2 {
			parsed := ParseSyntaxBody(blk.Body)

			ignoreChanges := extractLifecycleIgnoreChangesFromAST(blk.Body)
			if len(ignoreChanges) > 0 {
				parsed.Data.IgnoreChanges = append(parsed.Data.IgnoreChanges, ignoreChanges...)
			}

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
	blk := &ParsedBlock{Data: bd}
	bd.ParseAttributes(body)
	bd.ParseBlocks(body)
	return blk
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

func extractLifecycleIgnoreChangesFromAST(body *hclsyntax.Body) []string {
	var ignoreChanges []string

	for _, block := range body.Blocks {
		if block.Type == "lifecycle" {
			for name, attribute := range block.Body.Attributes {
				if name == "ignore_changes" {
					if listExpr, ok := attribute.Expr.(*hclsyntax.TupleConsExpr); ok {
						for _, expr := range listExpr.Exprs {
							switch exprType := expr.(type) {
							case *hclsyntax.ScopeTraversalExpr:
								if len(exprType.Traversal) > 0 {
									ignoreChanges = append(ignoreChanges, exprType.Traversal.RootName())
								}
							case *hclsyntax.TemplateExpr:
								if len(exprType.Parts) == 1 {
									if literalPart, ok := exprType.Parts[0].(*hclsyntax.LiteralValueExpr); ok && literalPart.Val.Type() == cty.String {
										ignoreChanges = append(ignoreChanges, literalPart.Val.AsString())
									}
								}
							case *hclsyntax.LiteralValueExpr:
								if exprType.Val.Type() == cty.String {
									ignoreChanges = append(ignoreChanges, exprType.Val.AsString())
								}
							}
						}
					}
				}
			}
		}
	}

	return ignoreChanges
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
