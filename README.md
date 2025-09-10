# diffy [![Go Reference](https://pkg.go.dev/badge/github.com/cloudnationhq/az-cn-go-diffy.svg)](https://pkg.go.dev/github.com/cloudnationhq/az-cn-go-diffy)

Terraform schema validation tool that identifies missing required and optional properties in your configurations.

## Installation

```bash
go get github.com/cloudnationhq/az-cn-go-diffy
```

## Usage

### Basic validation

```go
findings, err := diffy.ValidateSchema(
	diffy.WithTerraformRoot("../module"),
)
```

### With GitHub issue creation

```go
findings, err := diffy.ValidateSchema(
	diffy.WithTerraformRoot("../module"),
	diffy.WithGitHubIssueCreation(),
)
```

### With exclusions

```go
findings, err := diffy.ValidateSchema(
	diffy.WithTerraformRoot("../module"),
	diffy.WithExcludedResources("azurerm_resource_group", "azurerm_virtual_network"),
	diffy.WithExcludedDataSources("azurerm_client_config"),
)
```

### Environment variables

Set exclusions via environment variables (useful in CI/CD):

```bash
export TERRAFORM_ROOT="/path/to/terraform"
export EXCLUDED_RESOURCES="azurerm_resource_group,azurerm_virtual_network"
export EXCLUDED_DATA_SOURCES="azurerm_client_config,azurerm_subscription"
```

## Features

Validates resources and data sources against provider schemas

Recursively validates submodules

Creates GitHub issues with validation findings

Supports exclusions for resources and data sources

Respects terraform lifecycle blocks and ignore_changes

Handles nested dynamic blocks

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
