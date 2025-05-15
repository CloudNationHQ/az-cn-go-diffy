package diffy

import (
	"context"
)

// BlockProcessor interface defines methods for processing HCL blocks
type BlockProcessor interface {
	ParseAttributes(body *Body)
	ParseBlocks(body *Body)
	Validate(resourceType, path string, schema *SchemaBlock, parentIgnore []string, findings *[]ValidationFinding)
}

// SchemaValidator validates resources against their schema
type SchemaValidator interface {
	ValidateResources(resources []ParsedResource, schema TerraformSchema, providers map[string]ProviderConfig, dir, submoduleName string) []ValidationFinding
	ValidateDataSources(dataSources []ParsedDataSource, schema TerraformSchema, providers map[string]ProviderConfig, dir, submoduleName string) []ValidationFinding
}

// IssueManager creates or updates issues based on validation findings
type IssueManager interface {
	CreateOrUpdateIssue(ctx context.Context, findings []ValidationFinding) error
}

// Logger provides logging capabilities
type Logger interface {
	Logf(format string, args ...any)
}

// TerraformSchema represents the schema for Terraform providers
type TerraformSchema struct {
	ProviderSchemas map[string]*ProviderSchema `json:"provider_schemas"`
}

// ProviderSchema contains schemas for resources and data sources
type ProviderSchema struct {
	ResourceSchemas   map[string]*ResourceSchema `json:"resource_schemas"`
	DataSourceSchemas map[string]*ResourceSchema `json:"data_source_schemas"`
}

// ResourceSchema defines the schema for a resource or data source
type ResourceSchema struct {
	Block *SchemaBlock `json:"block"`
}

// SchemaBlock defines the structure of a block in a schema
type SchemaBlock struct {
	Attributes map[string]*SchemaAttribute `json:"attributes"`
	BlockTypes map[string]*SchemaBlockType `json:"block_types"`
}

// SchemaAttribute defines an attribute in a schema
type SchemaAttribute struct {
	Required   bool `json:"required"`
	Optional   bool `json:"optional"`
	Computed   bool `json:"computed"`
	Deprecated bool `json:"deprecated"`
}

// SchemaBlockType defines a nested block type
type SchemaBlockType struct {
	Nesting    string       `json:"nesting"`
	MinItems   int          `json:"min_items"`
	MaxItems   int          `json:"max_items"`
	Block      *SchemaBlock `json:"block"`
	Deprecated bool         `json:"deprecated"`
}

// ValidationFinding represents a finding during validation
type ValidationFinding struct {
	ResourceType  string
	Path          string // e.g., "root" or "root.some_nested_block"
	Name          string
	Required      bool
	IsBlock       bool
	IsDataSource  bool   // If true, this is a data source, not a resource
	SubmoduleName string // empty => root, else submodule name
}

// ProviderConfig defines configuration for a provider
type ProviderConfig struct {
	Source  string
	Version string
}

// ParsedResource represents a parsed Terraform resource
type ParsedResource struct {
	Type string
	Name string
	Data BlockData
}

// ParsedDataSource represents a parsed Terraform data source
type ParsedDataSource struct {
	Type string
	Name string
	Data BlockData
}

// BlockData contains the parsed data from a block
type BlockData struct {
	Properties    map[string]bool
	StaticBlocks  map[string]*ParsedBlock
	DynamicBlocks map[string]*ParsedBlock
	IgnoreChanges []string
}

// ParsedBlock represents a parsed block
type ParsedBlock struct {
	Data BlockData
}

// Body represents a generic HCL body interface
// This is a simplified interface for the example, in real use you'd
// use the actual HCL types from hashicorp/hcl
type Body struct {
	Attributes map[string]any
	Blocks     []*Block
}

// Block represents a generic HCL block
type Block struct {
	Type   string
	Labels []string
	Body   *Body
}

// SubModule represents a Terraform submodule
type SubModule struct {
	Name string
	Path string
}
