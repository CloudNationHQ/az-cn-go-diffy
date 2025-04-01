package diffy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// SimpleLogger is a basic implementation of the Logger interface
type SimpleLogger struct{}

// Logf implements the Logger interface
func (l *SimpleLogger) Logf(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

// SchemaValidatorOptions contains options for schema validation
type SchemaValidatorOptions struct {
	TerraformRoot     string
	CreateGitHubIssue bool
	Logger            Logger
	GitHubToken       string
	GitHubOwner       string
	GitHubRepo        string
	Silent            bool
}

// SchemaValidatorOption is a function that configures SchemaValidatorOptions
type SchemaValidatorOption func(*SchemaValidatorOptions)

// WithTerraformRoot sets the root directory for Terraform files
func WithTerraformRoot(path string) SchemaValidatorOption {
	return func(opts *SchemaValidatorOptions) {
		opts.TerraformRoot = path
	}
}

// WithGitHubIssueCreation enables GitHub issue creation with token
func WithGitHubIssueCreation() SchemaValidatorOption {
	return func(opts *SchemaValidatorOptions) {
		opts.CreateGitHubIssue = true
		opts.GitHubToken = os.Getenv("GITHUB_TOKEN")
		// Let the GetRepoInfo method handle owner/repo if not specified
	}
}


// ValidateSchema validates Terraform schema with the specified options
func ValidateSchema(options ...SchemaValidatorOption) ([]ValidationFinding, error) {
	// Default options
	opts := &SchemaValidatorOptions{
		TerraformRoot:     "../../",
		Logger:            &SimpleLogger{},
		CreateGitHubIssue: false,
		Silent:            false,
	}

	// Apply options
	for _, option := range options {
		option(opts)
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

	// Create GitHub issue if enabled
	if opts.CreateGitHubIssue && len(findings) > 0 {
		ctx := context.Background()
		if err := createGitHubIssue(ctx, opts, findings); err != nil {
			opts.Logger.Logf("Failed to create GitHub issue: %v", err)
		}
	}

	return findings, nil
}

// validateProject is the internal implementation of project validation
func validateProject(opts *SchemaValidatorOptions) ([]ValidationFinding, error) {
	// Resolve absolute path
	absRoot, err := filepath.Abs(opts.TerraformRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for %s: %w", opts.TerraformRoot, err)
	}

	// Run validation on root directory
	rootFindings, err := ValidateTerraformSchemaInDirectory(opts.Logger, absRoot, "")
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
			findings, err := ValidateTerraformSchemaInDirectory(opts.Logger, sm.Path, sm.Name)
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

// outputFindings prints validation findings to stdout
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

// createGitHubIssue creates a GitHub issue with validation findings
func createGitHubIssue(ctx context.Context, opts *SchemaValidatorOptions, findings []ValidationFinding) error {
	// Skip if no findings
	if len(findings) == 0 {
		return nil
	}

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

	// Create or update issue
	return issueManager.CreateOrUpdateIssue(ctx, findings)
}
