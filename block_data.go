package diffy

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func NewBlockData() BlockData {
	return BlockData{
		Properties:    make(map[string]bool),
		StaticBlocks:  make(map[string]*ParsedBlock),
		DynamicBlocks: make(map[string]*ParsedBlock),
		IgnoreChanges: []string{},
	}
}

func (bd *BlockData) ParseAttributes(body *hclsyntax.Body) {
	for name := range body.Attributes {
		bd.Properties[name] = true
	}
}

func (bd *BlockData) ParseBlocks(body *hclsyntax.Body) {
	directIgnoreChanges := extractLifecycleIgnoreChangesFromAST(body)
	if len(directIgnoreChanges) > 0 {
		bd.IgnoreChanges = append(bd.IgnoreChanges, directIgnoreChanges...)
	}

	for _, block := range body.Blocks {
		switch block.Type {
		case "lifecycle":
			bd.parseLifecycle(block.Body)
		case "dynamic":
			if len(block.Labels) == 1 {
				bd.parseDynamicBlock(block.Body, block.Labels[0])
			}
		default:
			parsed := ParseSyntaxBody(block.Body)
			bd.StaticBlocks[block.Type] = parsed
		}
	}
}

func (bd *BlockData) parseLifecycle(body *hclsyntax.Body) {
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

func (bd *BlockData) parseDynamicBlock(body *hclsyntax.Body, name string) {
	contentBlock := findContentBlockInBody(body)
	parsed := ParseSyntaxBody(contentBlock)
	if existing := bd.DynamicBlocks[name]; existing != nil {
		mergeBlocks(existing, parsed)
	} else {
		bd.DynamicBlocks[name] = parsed
	}
}

func findContentBlockInBody(body *hclsyntax.Body) *hclsyntax.Body {
	for _, b := range body.Blocks {
		if b.Type == "content" {
			return b.Body
		}
	}
	return body
}

func (bd *BlockData) Validate(
	resourceType, path string,
	schema *SchemaBlock,
	parentIgnore []string,
	findings *[]ValidationFinding,
) {
	if schema == nil {
		return
	}

	ignore := make([]string, len(parentIgnore), len(parentIgnore)+len(bd.IgnoreChanges))
	copy(ignore, parentIgnore)
	ignore = append(ignore, bd.IgnoreChanges...)

	bd.validateAttributes(resourceType, path, schema, ignore, findings)
	bd.validateBlocks(resourceType, path, schema, ignore, findings)
}

func (bd *BlockData) validateAttributes(
	resType, path string,
	schema *SchemaBlock,
	ignore []string,
	findings *[]ValidationFinding,
) {
	for name, attr := range schema.Attributes {
		if name == "id" {
			continue
		}

		if attr.Computed && !attr.Optional && !attr.Required {
			continue
		}

		if attr.Deprecated {
			continue
		}

		if isIgnored(ignore, name) {
			continue
		}

		if !bd.Properties[name] {
			*findings = append(*findings, ValidationFinding{
				ResourceType: resType,
				Path:         path,
				Name:         name,
				Required:     attr.Required,
				IsBlock:      false,
			})
		}
	}
}

func (bd *BlockData) validateBlocks(
	resType, path string,
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

		static := bd.StaticBlocks[name]
		dynamic := bd.DynamicBlocks[name]

		if static == nil && dynamic == nil {
			*findings = append(*findings, ValidationFinding{
				ResourceType: resType,
				Path:         path,
				Name:         name,
				Required:     blockType.MinItems > 0,
				IsBlock:      true,
			})
			continue
		}

		var target *ParsedBlock
		if static != nil {
			target = static
		} else {
			target = dynamic
		}

		newPath := fmt.Sprintf("%s.%s", path, name)
		target.Data.Validate(resType, newPath, blockType.Block, ignore, findings)
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
	for k := range src.Data.Properties {
		dest.Data.Properties[k] = true
	}

	for k, v := range src.Data.StaticBlocks {
		if existing, ok := dest.Data.StaticBlocks[k]; ok {
			mergeBlocks(existing, v)
		} else {
			dest.Data.StaticBlocks[k] = v
		}
	}

	for k, v := range src.Data.DynamicBlocks {
		if existing, ok := dest.Data.DynamicBlocks[k]; ok {
			mergeBlocks(existing, v)
		} else {
			dest.Data.DynamicBlocks[k] = v
		}
	}

	dest.Data.IgnoreChanges = append(dest.Data.IgnoreChanges, src.Data.IgnoreChanges...)
}
