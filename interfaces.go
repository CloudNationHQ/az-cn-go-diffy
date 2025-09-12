// Package diffy provides core interfaces for the validation system
package diffy

import (
	"context"
)

// BlockProcessor defines the interface for processing HCL blocks
type BlockProcessor interface {
	ParseAttributes(body *Body)
	ParseBlocks(body *Body)
	Validate(resourceType, path string, schema *SchemaBlock, parentIgnore []string, findings *[]ValidationFinding)
}

// SchemaValidator defines the interface for schema validation
type SchemaValidator interface {
	ValidateResources(resources []ParsedResource, schema TerraformSchema, providers map[string]ProviderConfig, dir, submoduleName string) []ValidationFinding
	ValidateDataSources(dataSources []ParsedDataSource, schema TerraformSchema, providers map[string]ProviderConfig, dir, submoduleName string) []ValidationFinding
}

// IssueManager defines the interface for GitHub issue management
type IssueManager interface {
	CreateOrUpdateIssue(ctx context.Context, findings []ValidationFinding) error
}

// Logger defines the interface for logging operations
type Logger interface {
	Logf(format string, args ...any)
}

// RepositoryInfoProvider defines the interface for repository information
type RepositoryInfoProvider interface {
	GetRepoInfo() (owner, name string)
}
