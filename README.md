# diffy [![Go Reference](https://pkg.go.dev/badge/github.com/cloudnationhq/az-cn-go-diffy.svg)](https://pkg.go.dev/github.com/cloudnation/az-cn-go-diffy)

Diffy is a terraform schema validation tool that identifies missing required and optional properties in your configurations.

## Installation

```zsh
go get github.com/cloudnationhq/az-cn-go-diffy
```

## Usage

as a local test with a relative path:

```go
func TestTerraformSchemaValidation(t *testing.T) {
	findings, err := diffy.ValidateSchema(
		diffy.WithTerraformRoot("../module"),
		func(opts *diffy.SchemaValidatorOptions) {
			opts.Silent = true
		},
	)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}
	for _, finding := range findings {
		t.Logf("%s", diffy.FormatFinding(finding))
	}
	if len(findings) > 0 {
		t.Errorf("Found %d missing properties/blocks in root or submodules. See logs above.", len(findings))
	}
}
```

within github actions:

```go
func TestTerraformSchemaValidation(t *testing.T) {
	findings, err := diffy.ValidateSchema(
		diffy.WithGitHubIssueCreation(),
		func(opts *diffy.SchemaValidatorOptions) {
			opts.Silent = true
		},
	)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}
	for _, finding := range findings {
		t.Logf("%s", diffy.FormatFinding(finding))
	}
	if len(findings) > 0 {
		t.Errorf("Found %d missing properties/blocks in root or submodules. See logs above.", len(findings))
	}
}
```

```yaml
- name: Run schema tests
  working-directory: called/tests
  run: |
    go test -v -run TestTerraformSchemaValidation diffy_test.go
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    TERRAFORM_ROOT: "${{ github.workspace }}/caller"
```

## Features

Automatically validates terraform resources against their provider schemas.

Recursively validates all submodules.

Optionally creates or updates gitHub issues with validation findings.

Designed to work seamlessly in CI/CD workflows with environment variable support.

Respects terraform lifecycle blocks and ignore_changes settings.

Properly handles nested dynamic blocks in your terraform configurations.

Identifies both missing required and optional properties.

## Options

Diffy supports a functional options pattern for configuration:

`WithTerraformRoot(path):` Sets the root directory for Terraform files (defaults to "../../" if not specified)

`WithGitHubIssueCreation():` Enables GitHub issue creation based on validation findings

## Contributors

We welcome contributions from the community! Whether it's reporting a bug, suggesting a new feature, or submitting a pull request, your input is highly valued. <br><br>

<a href="https://github.com/cloudnationhq/az-cn-go-diffy/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=cloudnationhq/az-cn-go-diffy" />
</a>

## Notes

The `TERRAFORM_ROOT` environment variable takes highest priority if set.

A path must be specified either through the `TERRAFORM_ROOT` environment variable or via `WithTerraformRoot()` option.

GitHub issue creation requires a `GITHUB_TOKEN` environment variable.

This approach supports both local testing and CI/CD environments with the same code.
