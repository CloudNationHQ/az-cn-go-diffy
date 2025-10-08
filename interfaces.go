// Package diffy provides core interfaces for the validation system
package diffy

import (
	"context"
)

type BlockProcessor interface {
	ParseAttributes(body *Body)
	ParseBlocks(body *Body)
	Validate(resourceType, path string, schema *SchemaBlock, parentIgnore []string, findings *[]ValidationFinding)
}

type SchemaValidator interface {
	ValidateResources(resources []ParsedResource, schema TerraformSchema, providers map[string]ProviderConfig, dir, submoduleName string) []ValidationFinding
	ValidateDataSources(dataSources []ParsedDataSource, schema TerraformSchema, providers map[string]ProviderConfig, dir, submoduleName string) []ValidationFinding
}

type IssueManager interface {
	CreateOrUpdateIssue(ctx context.Context, findings []ValidationFinding) error
}

type Logger interface {
	Logf(format string, args ...any)
}

type RepositoryInfoProvider interface {
	GetRepoInfo() (owner, name string)
}
