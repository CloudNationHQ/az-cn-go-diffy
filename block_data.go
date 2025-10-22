package diffy

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func NewBlockData() BlockData {
	return BlockData{
		Properties:    make(map[string]bool),
		StaticBlocks:  make(map[string][]*ParsedBlock),
		DynamicBlocks: make(map[string]*ParsedBlock),
		IgnoreChanges: []string{},
	}
}

func (blockData *BlockData) ParseAttributes(body *hclsyntax.Body) {
	for name := range body.Attributes {
		blockData.Properties[name] = true
	}
}

func (blockData *BlockData) ParseBlocks(body *hclsyntax.Body) {
	for _, block := range body.Blocks {
		switch block.Type {
		case "lifecycle":
			blockData.parseLifecycle(block.Body)
		case "dynamic":
			if len(block.Labels) == 1 {
				blockData.parseDynamicBlock(block.Body, block.Labels[0])
			}
		default:
			parsed := ParseSyntaxBody(block.Body)
			blockData.StaticBlocks[block.Type] = append(blockData.StaticBlocks[block.Type], parsed)
		}
	}
}

func (blockData *BlockData) parseLifecycle(body *hclsyntax.Body) {
	for name, attribute := range body.Attributes {
		if name == "ignore_changes" {
			extracted := extractIgnoreChanges(attribute)
			blockData.IgnoreChanges = append(blockData.IgnoreChanges, extracted...)
		}
	}
}

func extractIgnoreChanges(attribute *hclsyntax.Attribute) []string {
	value, diags := attribute.Expr.Value(nil)
	if diags == nil || !diags.HasErrors() {
		extracted := extractIgnoreChangesFromValue(value)
		if len(extracted) > 0 {
			return extracted
		}
	}
	return extractIgnoreChangesFromExpr(attribute.Expr)
}

func (blockData *BlockData) parseDynamicBlock(body *hclsyntax.Body, name string) {
	blockData.Properties[name] = true
	contentBlock := findContentBlockInBody(body)
	parsed := ParseSyntaxBody(contentBlock)
	if existing := blockData.DynamicBlocks[name]; existing != nil {
		mergeBlocks(existing, parsed)
	} else {
		blockData.DynamicBlocks[name] = parsed
	}
}

func findContentBlockInBody(body *hclsyntax.Body) *hclsyntax.Body {
	for _, block := range body.Blocks {
		if block.Type == "content" {
			return block.Body
		}
	}
	return body
}

func extractIgnoreChangesFromExpr(expr hclsyntax.Expression) []string {
	switch e := expr.(type) {
	case *hclsyntax.TupleConsExpr:
		var results []string
		for _, item := range e.Exprs {
			results = append(results, extractIgnoreChangesFromExpr(item)...)
		}
		return results
	case *hclsyntax.ScopeTraversalExpr:
		if len(e.Traversal) > 0 {
			return []string{e.Traversal.RootName()}
		}
	case *hclsyntax.TemplateExpr:
		if len(e.Parts) == 1 {
			if lit, ok := e.Parts[0].(*hclsyntax.LiteralValueExpr); ok {
				return extractIgnoreChangesFromExpr(lit)
			}
		}
	case *hclsyntax.LiteralValueExpr:
		return extractIgnoreChangesFromValue(e.Val)
	}
	return nil
}

func (blockData *BlockData) Validate(
	resourceType, path string,
	schema *SchemaBlock,
	parentIgnore []string,
	findings *[]ValidationFinding,
) {
	if schema == nil {
		return
	}

	ignore := make([]string, len(parentIgnore), len(parentIgnore)+len(blockData.IgnoreChanges))
	copy(ignore, parentIgnore)
	ignore = append(ignore, blockData.IgnoreChanges...)

	blockData.validateAttributes(resourceType, path, schema, ignore, findings)
	blockData.validateBlocks(resourceType, path, schema, ignore, findings)
}

func (blockData *BlockData) validateAttributes(
	resourceType, path string,
	schema *SchemaBlock,
	ignore []string,
	findings *[]ValidationFinding,
) {
	for name, attribute := range schema.Attributes {
		if name == "id" {
			continue
		}

		if attribute.Computed && !attribute.Optional && !attribute.Required {
			continue
		}

		if attribute.Deprecated {
			continue
		}

		if isIgnored(ignore, name) {
			continue
		}

		if !blockData.Properties[name] {
			*findings = append(*findings, ValidationFinding{
				ResourceType: resourceType,
				Path:         path,
				Name:         name,
				Required:     attribute.Required,
				IsBlock:      false,
			})
		}
	}
}

func (blockData *BlockData) validateBlocks(
	resourceType, path string,
	schema *SchemaBlock,
	ignore []string,
	findings *[]ValidationFinding,
) {
	for name, blockType := range schema.BlockTypes {
		if name == "timeouts" || isIgnored(ignore, name) {
			continue
		}

		if blockType.Deprecated {
			continue
		}

		staticBlocks := blockData.StaticBlocks[name]
		dynamic := blockData.DynamicBlocks[name]

		if len(staticBlocks) == 0 && dynamic == nil {
			*findings = append(*findings, ValidationFinding{
				ResourceType: resourceType,
				Path:         path,
				Name:         name,
				Required:     blockType.MinItems > 0,
				IsBlock:      true,
			})
			continue
		}

		for i, blk := range staticBlocks {
			blockPath := fmt.Sprintf("%s.%s", path, name)
			if len(staticBlocks) > 1 {
				blockPath = fmt.Sprintf("%s.%s[%d]", path, name, i)
			}
			blk.Data.Validate(resourceType, blockPath, blockType.Block, ignore, findings)
		}

		if dynamic != nil {
			blockPath := fmt.Sprintf("%s.%s", path, name)
			dynamic.Data.Validate(resourceType, blockPath, blockType.Block, ignore, findings)
		}
	}
}

func isIgnored(ignore []string, name string) bool {
	for _, item := range ignore {
		if item == "*all*" {
			return true
		}
		if strings.EqualFold(item, name) {
			return true
		}
	}
	return false
}

func mergeBlocks(dest, src *ParsedBlock) {
	for key := range src.Data.Properties {
		dest.Data.Properties[key] = true
	}

	for key, blocks := range src.Data.StaticBlocks {
		dest.Data.StaticBlocks[key] = append(dest.Data.StaticBlocks[key], blocks...)
	}

	for key, value := range src.Data.DynamicBlocks {
		if existing, ok := dest.Data.DynamicBlocks[key]; ok {
			mergeBlocks(existing, value)
		} else {
			dest.Data.DynamicBlocks[key] = value
		}
	}

	dest.Data.IgnoreChanges = append(dest.Data.IgnoreChanges, src.Data.IgnoreChanges...)
}
