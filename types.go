// Package diffy provides core types and data structures
package diffy

import (
	"fmt"
)

type ParseError struct {
	File    string
	Message string
	Err     error
}

func (e *ParseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("parse error in %s: %s: %v", e.File, e.Message, e.Err)
	}
	return fmt.Sprintf("parse error in %s: %s", e.File, e.Message)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

type ValidationError struct {
	ResourceType string
	Message      string
	Err          error
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("validation error for %s: %s: %v", e.ResourceType, e.Message, e.Err)
	}
	return fmt.Sprintf("validation error for %s: %s", e.ResourceType, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

type GitHubError struct {
	Operation string
	Message   string
	Err       error
}

func (e *GitHubError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("GitHub %s error: %s: %v", e.Operation, e.Message, e.Err)
	}
	return fmt.Sprintf("GitHub %s error: %s", e.Operation, e.Message)
}

func (e *GitHubError) Unwrap() error {
	return e.Err
}

type TerraformSchema struct {
	ProviderSchemas map[string]*ProviderSchema `json:"provider_schemas"`
}

type ProviderSchema struct {
	ResourceSchemas   map[string]*ResourceSchema `json:"resource_schemas"`
	DataSourceSchemas map[string]*ResourceSchema `json:"data_source_schemas"`
}

type ResourceSchema struct {
	Block *SchemaBlock `json:"block"`
}

type SchemaBlock struct {
	Attributes map[string]*SchemaAttribute `json:"attributes"`
	BlockTypes map[string]*SchemaBlockType `json:"block_types"`
}

type SchemaAttribute struct {
	Required   bool `json:"required"`
	Optional   bool `json:"optional"`
	Computed   bool `json:"computed"`
	Deprecated bool `json:"deprecated"`
}

type SchemaBlockType struct {
	Nesting    string       `json:"nesting"`
	MinItems   int          `json:"min_items"`
	MaxItems   int          `json:"max_items"`
	Block      *SchemaBlock `json:"block"`
	Deprecated bool         `json:"deprecated"`
}

type ValidationFinding struct {
	ResourceType  string
	Path          string
	Name          string
	Required      bool
	IsBlock       bool
	IsDataSource  bool
	SubmoduleName string
}

type ProviderConfig struct {
	Source  string
	Version string
}

type ParsedResource struct {
	Type string
	Name string
	Data BlockData
}

type ParsedDataSource struct {
	Type string
	Name string
	Data BlockData
}

type BlockData struct {
	Properties    map[string]bool
	StaticBlocks  map[string][]*ParsedBlock
	DynamicBlocks map[string]*ParsedBlock
	IgnoreChanges []string
}

type ParsedBlock struct {
	Data BlockData
}

type Body struct {
	Attributes map[string]any
	Blocks     []*Block
}

type Block struct {
	Type   string
	Labels []string
	Body   *Body
}

type SubModule struct {
	Name string
	Path string
}
