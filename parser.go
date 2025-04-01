package diffy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// HCLParser parses Terraform HCL files
type HCLParser interface {
	ParseProviderRequirements(ctx context.Context, filename string) (map[string]ProviderConfig, error)
	ParseMainFile(ctx context.Context, filename string) ([]ParsedResource, []ParsedDataSource, error)
}

// TerraformRunner runs Terraform commands
type TerraformRunner interface {
	Init(ctx context.Context, dir string) error
	GetSchema(ctx context.Context, dir string) (*TerraformSchema, error)
}

// DefaultHCLParser implements HCLParser
type DefaultHCLParser struct{}

// NewHCLParser creates a new HCL parser
func NewHCLParser() *DefaultHCLParser {
	return &DefaultHCLParser{}
}

// ParseProviderRequirements parses provider requirements from a terraform.tf file
func (p *DefaultHCLParser) ParseProviderRequirements(ctx context.Context, filename string) (map[string]ProviderConfig, error) {
	parser := hclparse.NewParser()
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return map[string]ProviderConfig{}, nil
	}
	f, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse error in file %s: %v", filename, diags)
	}
	body, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return nil, fmt.Errorf("invalid body in file %s", filename)
	}
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

// ParseMainFile parses a main.tf file to extract resources and data sources
func (p *DefaultHCLParser) ParseMainFile(ctx context.Context, filename string) ([]ParsedResource, []ParsedDataSource, error) {
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, nil, fmt.Errorf("parse error in file %s: %v", filename, diags)
	}
	body, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return nil, nil, fmt.Errorf("invalid body in file %s", filename)
	}
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

// ParseSyntaxBody parses a hclsyntax.Body into a ParsedBlock
func ParseSyntaxBody(body *hclsyntax.Body) *ParsedBlock {
	bd := NewBlockData()
	blk := &ParsedBlock{Data: bd}
	bd.ParseAttributes(body)
	bd.ParseBlocks(body)
	return blk
}

// ParseSyntaxAttributes extracts attributes from a hclsyntax.Body
func (bd *BlockData) ParseSyntaxAttributes(body *hclsyntax.Body) {
	for name := range body.Attributes {
		bd.Properties[name] = true
	}
}

// ParseSyntaxBlocks processes all blocks in a hclsyntax.Body
func (bd *BlockData) ParseSyntaxBlocks(body *hclsyntax.Body) {
	directIgnoreChanges := extractLifecycleIgnoreChangesFromAST(body)
	if len(directIgnoreChanges) > 0 {
		bd.IgnoreChanges = append(bd.IgnoreChanges, directIgnoreChanges...)
	}

	for _, block := range body.Blocks {
		switch block.Type {
		case "lifecycle":
			bd.parseLifecycleFromAST(block.Body)
		case "dynamic":
			if len(block.Labels) == 1 {
				bd.parseDynamicBlockFromAST(block.Body, block.Labels[0])
			}
		default:
			parsed := ParseSyntaxBody(block.Body)
			bd.StaticBlocks[block.Type] = parsed
		}
	}
}

// parseLifecycleFromAST extracts ignore_changes from a lifecycle block
func (bd *BlockData) parseLifecycleFromAST(body *hclsyntax.Body) {
	for name, attr := range body.Attributes {
		if name == "ignore_changes" {
			val, diags := attr.Expr.Value(nil)
			if diags == nil || !diags.HasErrors() {
				extracted := extractIgnoreChanges(val)
				bd.IgnoreChanges = append(bd.IgnoreChanges, extracted...)
			}
		}
	}
}

// parseDynamicBlockFromAST processes a dynamic block
func (bd *BlockData) parseDynamicBlockFromAST(body *hclsyntax.Body, name string) {
	contentBlock := findContentBlockFromAST(body)
	parsed := ParseSyntaxBody(contentBlock)
	if existing := bd.DynamicBlocks[name]; existing != nil {
		mergeBlocks(existing, parsed)
	} else {
		bd.DynamicBlocks[name] = parsed
	}
}

// findContentBlockFromAST finds the content block within a dynamic block
func findContentBlockFromAST(body *hclsyntax.Body) *hclsyntax.Body {
	for _, b := range body.Blocks {
		if b.Type == "content" {
			return b.Body
		}
	}
	return body
}

// extractIgnoreChanges extracts ignore_changes values from a cty.Value
func extractIgnoreChanges(val cty.Value) []string {
	var changes []string
	if val.Type().IsCollectionType() {
		for it := val.ElementIterator(); it.Next(); {
			_, v := it.Element()
			if v.Type() == cty.String {
				str := v.AsString()
				if str == "all" {
					return []string{"*all*"}
				}
				changes = append(changes, str)
			}
		}
	}
	return changes
}

// extractLifecycleIgnoreChangesFromAST extracts ignore_changes from AST
func extractLifecycleIgnoreChangesFromAST(body *hclsyntax.Body) []string {
	var ignoreChanges []string

	for _, block := range body.Blocks {
		if block.Type == "lifecycle" {
			for name, attr := range block.Body.Attributes {
				if name == "ignore_changes" {
					if listExpr, ok := attr.Expr.(*hclsyntax.TupleConsExpr); ok {
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

// NormalizeSource normalizes a provider source
func NormalizeSource(source string) string {
	if strings.Contains(source, "/") && !strings.Contains(source, "registry.terraform.io/") {
		return "registry.terraform.io/" + source
	}
	return source
}

// FindSubmodules finds submodules in a directory
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
