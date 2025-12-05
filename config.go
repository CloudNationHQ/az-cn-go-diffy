// Package diffy provides configuration options for schema validation
package diffy

import (
	"os"
)

type SchemaValidatorOptions struct {
	TerraformRoot       string
	CreateGitHubIssue   bool
	Logger              Logger
	GitHubToken         string
	GitHubOwner         string
	GitHubRepo          string
	Silent              bool
	ExcludedResources   []string
	ExcludedDataSources []string
	Parser              HCLParser
	TerraformRunner     TerraformRunner
}

type SchemaValidatorOption func(*SchemaValidatorOptions)

func WithTerraformRoot(path string) SchemaValidatorOption {
	return func(opts *SchemaValidatorOptions) {
		opts.TerraformRoot = path
	}
}

func WithGitHubIssueCreation() SchemaValidatorOption {
	return func(opts *SchemaValidatorOptions) {
		opts.CreateGitHubIssue = true
		opts.GitHubToken = os.Getenv("GITHUB_TOKEN")
	}
}

func WithExcludedResources(resources ...string) SchemaValidatorOption {
	return func(opts *SchemaValidatorOptions) {
		opts.ExcludedResources = append(opts.ExcludedResources, resources...)
	}
}

func WithExcludedDataSources(dataSources ...string) SchemaValidatorOption {
	return func(opts *SchemaValidatorOptions) {
		opts.ExcludedDataSources = append(opts.ExcludedDataSources, dataSources...)
	}
}

func WithParser(parser HCLParser) SchemaValidatorOption {
	return func(opts *SchemaValidatorOptions) {
		opts.Parser = parser
	}
}

func WithTerraformRunner(runner TerraformRunner) SchemaValidatorOption {
	return func(opts *SchemaValidatorOptions) {
		opts.TerraformRunner = runner
	}
}
