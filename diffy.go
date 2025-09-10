// Package diffy validates Terraform configurations against provider schemas.
package diffy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SimpleLogger implements Logger interface.
type SimpleLogger struct{}

func (l *SimpleLogger) Logf(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

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

func ValidateSchema(options ...SchemaValidatorOption) ([]ValidationFinding, error) {
	// Initialize with minimal defaults
	opts := &SchemaValidatorOptions{
		Logger:            &SimpleLogger{},
		CreateGitHubIssue: false,
		Silent:            false,
	}

	// Apply options
	for _, option := range options {
		option(opts)
	}

	// Check for TERRAFORM_ROOT environment variable (highest priority)
	if envRoot := os.Getenv("TERRAFORM_ROOT"); envRoot != "" {
		opts.TerraformRoot = envRoot
	}

	// Check for exclusion environment variables
	if envExcludedResources := os.Getenv("EXCLUDED_RESOURCES"); envExcludedResources != "" {
		resources := strings.Split(envExcludedResources, ",")
		for i, r := range resources {
			resources[i] = strings.TrimSpace(r)
		}
		opts.ExcludedResources = append(opts.ExcludedResources, resources...)
	}

	if envExcludedDataSources := os.Getenv("EXCLUDED_DATA_SOURCES"); envExcludedDataSources != "" {
		dataSources := strings.Split(envExcludedDataSources, ",")
		for i, ds := range dataSources {
			dataSources[i] = strings.TrimSpace(ds)
		}
		opts.ExcludedDataSources = append(opts.ExcludedDataSources, dataSources...)
	}

	// Validate TerraformRoot is set
	if opts.TerraformRoot == "" {
		return nil, fmt.Errorf("terraform root path not specified - set TERRAFORM_ROOT environment variable or use WithTerraformRoot option")
	}

	// Validate Terraform project
	findings, err := validateProject(opts)
	if err != nil {
		return nil, err
	}

	// Output findings to console if not silent
	if !opts.Silent {
		outputFindings(findings)
	}

	// Always check GitHub issue creation when enabled, even with no findings
	if opts.CreateGitHubIssue {
		ctx := context.Background()
		if err := createGitHubIssue(ctx, opts, findings); err != nil {
			opts.Logger.Logf("Failed to create/update GitHub issue: %v", err)
		}
	}

	return findings, nil
}

func validateProject(opts *SchemaValidatorOptions) ([]ValidationFinding, error) {
	// Resolve absolute path
	absRoot, err := filepath.Abs(opts.TerraformRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for %s: %w", opts.TerraformRoot, err)
	}

	// Run validation on root directory
	rootFindings, err := ValidateTerraformSchemaInDirectoryWithOptions(opts.Logger, absRoot, "", opts.ExcludedResources, opts.ExcludedDataSources)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var allFindings []ValidationFinding
	allFindings = append(allFindings, rootFindings...)

	// Always validate submodules - this is now the default behavior
	modulesDir := filepath.Join(absRoot, "modules")
	submodules, err := FindSubmodules(modulesDir)
	if err != nil {
		// Just log and continue if no submodules are found - no need to error out
		if !opts.Silent {
			fmt.Printf("Note: No submodules found in %s\n", modulesDir)
		}
	} else {
		for _, sm := range submodules {
			findings, err := ValidateTerraformSchemaInDirectoryWithOptions(opts.Logger, sm.Path, sm.Name, opts.ExcludedResources, opts.ExcludedDataSources)
			if err != nil {
				opts.Logger.Logf("Failed to validate submodule %s: %v", sm.Name, err)
				continue
			}
			allFindings = append(allFindings, findings...)
		}
	}

	// Deduplicate findings
	deduplicatedFindings := DeduplicateFindings(allFindings)

	return deduplicatedFindings, nil
}

func outputFindings(findings []ValidationFinding) {
	if len(findings) == 0 {
		fmt.Println("No validation findings.")
		return
	}

	fmt.Printf("Found %d issues:\n", len(findings))

	for _, f := range findings {
		fmt.Println(FormatFinding(f))
	}
}

func createGitHubIssue(ctx context.Context, opts *SchemaValidatorOptions, findings []ValidationFinding) error {
	// Get GitHub token
	if opts.GitHubToken == "" {
		return fmt.Errorf("GitHub token not provided")
	}

	owner := opts.GitHubOwner
	repo := opts.GitHubRepo

	// If owner/repo not specified, try to determine from git
	if owner == "" || repo == "" {
		gi := NewGitRepoInfo(opts.TerraformRoot)
		owner, repo = gi.GetRepoInfo()
		if owner == "" || repo == "" {
			return fmt.Errorf("could not determine repository info for GitHub issue creation")
		}
	}

	// Create issue manager
	issueManager := NewGitHubIssueManager(owner, repo, opts.GitHubToken)

	// If no findings, check if there's an existing issue to close
	if len(findings) == 0 {
		return issueManager.CloseExistingIssuesIfEmpty(ctx)
	}

	// Create or update issue
	return issueManager.CreateOrUpdateIssue(ctx, findings)
}
