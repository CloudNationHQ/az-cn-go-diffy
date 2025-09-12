# diffy [![Go Reference](https://pkg.go.dev/badge/github.com/cloudnationhq/az-cn-go-diffy.svg)](https://pkg.go.dev/github.com/cloudnationhq/az-cn-go-diffy)

A terraform schema validation tool that ensures your infrastructure configurations are complete and compliant with provider specifications.

Automatically detects missing required properties, validates optional attributes, and helps maintain configuration quality across teams and projects.

## Why diffy?

Terraform configurations can become complex and inconsistent over time. Missing required properties, outdated attribute names, and incomplete resource definitions can lead to deployment failures and configuration drift.

Diffy helps you:

Catch configuration errors before deployment

Ensure compliance with provider schema requirements

Maintain consistency across large infrastructure codebases

Reduce debugging time during infrastructure changes

Support automated validation in CI/CD pipelines

## Installation

`go get github.com/cloudnationhq/az-cn-go-diffy`

## Usage

See the [examples/usage](examples/usage/) directory for examples and test cases.

## Features

`Schema Validation`

Validates all Terraform resources and data sources against their provider schemas

Identifies missing required properties that would cause deployment failures

Detects deprecated or invalid attribute configurations

Supports recursive validation of nested modules and submodules

`GitHub Integration`

Automatically creates GitHub issues for validation findings

Provides detailed, actionable feedback on configuration problems

Enables team collaboration on infrastructure quality improvements

`Flexible Configuration`

Supports resource and data source exclusions for custom validation rules

Environment variable configuration for CI/CD integration

Configurable logging levels and output formats

Middleware pattern for custom validation extensions

`Advanced Terraform Support`

Respects Terraform lifecycle blocks and ignore_changes directives

Handles complex dynamic blocks and nested configurations

Works with all major Terraform providers and custom providers

## Configuration

`Environment Variables`

Configure diffy through environment variables for CI/CD pipelines:

`TERRAFORM_ROOT`: Path to your Terraform configuration root directory

`EXCLUDED_RESOURCES`: Comma-separated list of resource types to exclude from validation

`EXCLUDED_DATA_SOURCES`: Comma-separated list of data source types to exclude

`GITHUB_TOKEN`: Personal access token for GitHub issue creation (optional)

## Notes

The `TERRAFORM_ROOT` environment variable takes highest priority when set

A Terraform root path must be specified either via environment variable or configuration option

GitHub integration requires appropriate repository permissions and a valid token

Validation respects Terraform's lifecycle ignore_changes directives

## Contributors

We welcome contributions from the community! Whether it's reporting a bug, suggesting a new feature, or submitting a pull request, your input is highly valued. <br><br>

<a href="https://github.com/cloudnationhq/az-cn-go-diffy/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=cloudnationhq/az-cn-go-diffy" />
</a>
