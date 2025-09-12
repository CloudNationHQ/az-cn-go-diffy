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

func (blockData *BlockData) ParseAttributes(body *hclsyntax.Body) {
	for name := range body.Attributes {
		blockData.Properties[name] = true
	}
}

func (blockData *BlockData) ParseBlocks(body *hclsyntax.Body) {
	directIgnoreChanges := extractLifecycleIgnoreChangesFromAST(body)
	if len(directIgnoreChanges) > 0 {
		blockData.IgnoreChanges = append(blockData.IgnoreChanges, directIgnoreChanges...)
	}

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
			blockData.StaticBlocks[block.Type] = parsed
		}
	}
}

func (blockData *BlockData) parseLifecycle(body *hclsyntax.Body) {
	for name, attribute := range body.Attributes {
		if name == "ignore_changes" {
			value, diags := attribute.Expr.Value(nil)
			if diags == nil || !diags.HasErrors() {
				extracted := extractIgnoreChangesFromValue(value)
				blockData.IgnoreChanges = append(blockData.IgnoreChanges, extracted...)
			}
		}
	}
}

func (blockData *BlockData) parseDynamicBlock(body *hclsyntax.Body, name string) {
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

		static := blockData.StaticBlocks[name]
		dynamic := blockData.DynamicBlocks[name]

		if static == nil && dynamic == nil {
			*findings = append(*findings, ValidationFinding{
				ResourceType: resourceType,
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
		target.Data.Validate(resourceType, newPath, blockType.Block, ignore, findings)
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

	for key, value := range src.Data.StaticBlocks {
		if existing, ok := dest.Data.StaticBlocks[key]; ok {
			mergeBlocks(existing, value)
		} else {
			dest.Data.StaticBlocks[key] = value
		}
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
