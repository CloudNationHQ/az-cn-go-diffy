package diffy

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

type DefaultSchemaValidator struct {
	logger Logger
}

func NewSchemaValidator(logger Logger) *DefaultSchemaValidator {
	return &DefaultSchemaValidator{
		logger: logger,
	}
}

func (validator *DefaultSchemaValidator) ValidateResources(
	resources []ParsedResource,
	schema TerraformSchema,
	providers map[string]ProviderConfig,
	dir, submoduleName string,
) []ValidationFinding {
	return validator.validateEntities(resources, schema, providers, dir, submoduleName, false)
}

func (validator *DefaultSchemaValidator) ValidateDataSources(
	dataSources []ParsedDataSource,
	schema TerraformSchema,
	providers map[string]ProviderConfig,
	dir, submoduleName string,
) []ValidationFinding {
	return validator.validateEntities(dataSources, schema, providers, dir, submoduleName, true)
}

// validateEntities is a generic function that handles validation for both resources and data sources
func (validator *DefaultSchemaValidator) validateEntities(
	entities any,
	schema TerraformSchema,
	providers map[string]ProviderConfig,
	dir, submoduleName string,
	isDataSource bool,
) []ValidationFinding {
	var findings []ValidationFinding

	// Type assertion to handle both []ParsedResource and []ParsedDataSource
	var entityList []struct {
		Type string
		Name string
		Data BlockData
	}

	switch e := entities.(type) {
	case []ParsedResource:
		for _, r := range e {
			entityList = append(entityList, struct {
				Type string
				Name string
				Data BlockData
			}{r.Type, r.Name, r.Data})
		}
	case []ParsedDataSource:
		for _, ds := range e {
			entityList = append(entityList, struct {
				Type string
				Name string
				Data BlockData
			}{ds.Type, ds.Name, ds.Data})
		}
	default:
		return findings
	}

	for _, entity := range entityList {
		provName := strings.SplitN(entity.Type, "_", 2)[0]
		cfg, ok := providers[provName]
		if !ok {
			validator.logger.Logf("No provider config for %s type %s in %s",
				map[bool]string{true: "data source", false: "resource"}[isDataSource],
				entity.Type, dir)
			continue
		}

		pSchema, ok := schema.ProviderSchemas[cfg.Source]
		if !ok {
			validator.logger.Logf("No provider schema found for source %s in %s", cfg.Source, dir)
			continue
		}

		// Get the appropriate schema based on entity type
		var resSchema *ResourceSchema
		var schemaExists bool
		if isDataSource {
			resSchema, schemaExists = pSchema.DataSourceSchemas[entity.Type]
		} else {
			resSchema, schemaExists = pSchema.ResourceSchemas[entity.Type]
		}

		if !schemaExists {
			entityType := map[bool]string{true: "data source", false: "resource"}[isDataSource]
			validator.logger.Logf("No %s schema found for %s in provider %s (dir=%s)",
				entityType, entity.Type, cfg.Source, dir)
			continue
		}

		var localFindings []ValidationFinding
		entity.Data.Validate(entity.Type, "root", resSchema.Block, entity.Data.IgnoreChanges, &localFindings)

		for i := range localFindings {
			shouldExclude := false
			for _, ignored := range entity.Data.IgnoreChanges {
				if strings.EqualFold(ignored, localFindings[i].Name) {
					shouldExclude = true
					break
				}
			}

			if !shouldExclude {
				localFindings[i].SubmoduleName = submoduleName
				localFindings[i].IsDataSource = isDataSource
				findings = append(findings, localFindings[i])
			}
		}
	}

	return findings
}

func ValidateTerraformSchema(logger Logger, dir, submoduleName string, parser HCLParser, runner TerraformRunner) ([]ValidationFinding, error) {
	return ValidateTerraformSchemaWithOptions(logger, dir, submoduleName, parser, runner, nil, nil)
}

func ValidateTerraformSchemaWithOptions(logger Logger, dir, submoduleName string, parser HCLParser, runner TerraformRunner, excludedResources, excludedDataSources []string) ([]ValidationFinding, error) {
	ctx := context.Background()

	tfFile := filepath.Join(dir, "terraform.tf")
	providers, err := parser.ParseProviderRequirements(ctx, tfFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provider config in %s: %w", dir, err)
	}

	if err := runner.Init(ctx, dir); err != nil {
		return nil, err
	}

	tfSchema, err := runner.GetSchema(ctx, dir)
	if err != nil {
		return nil, err
	}

	mainTf := filepath.Join(dir, "main.tf")
	resources, dataSources, err := parser.ParseMainFile(ctx, mainTf)
	if err != nil {
		return nil, fmt.Errorf("parseMainFile in %s: %w", dir, err)
	}

	resources = filterResources(resources, excludedResources)
	dataSources = filterDataSources(dataSources, excludedDataSources)

	validator := NewSchemaValidator(logger)
	var findings []ValidationFinding
	findings = append(findings, validator.ValidateResources(resources, *tfSchema, providers, dir, submoduleName)...)
	findings = append(findings, validator.ValidateDataSources(dataSources, *tfSchema, providers, dir, submoduleName)...)

	return findings, nil
}

func filterResources(resources []ParsedResource, excluded []string) []ParsedResource {
	if len(excluded) == 0 {
		return resources
	}

	var filtered []ParsedResource
	for _, resource := range resources {
		if !slices.Contains(excluded, resource.Type) {
			filtered = append(filtered, resource)
		}
	}
	return filtered
}

func filterDataSources(dataSources []ParsedDataSource, excluded []string) []ParsedDataSource {
	if len(excluded) == 0 {
		return dataSources
	}

	var filtered []ParsedDataSource
	for _, dataSource := range dataSources {
		if !slices.Contains(excluded, dataSource.Type) {
			filtered = append(filtered, dataSource)
		}
	}
	return filtered
}

func DeduplicateFindings(findings []ValidationFinding) []ValidationFinding {
	seen := make(map[string]struct{})
	result := make([]ValidationFinding, 0, len(findings))

	for _, finding := range findings {
		key := fmt.Sprintf("%s|%s|%s|%v|%v|%s",
			finding.ResourceType,
			finding.Path,
			finding.Name,
			finding.IsBlock,
			finding.IsDataSource,
			finding.SubmoduleName,
		)

		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, finding)
		}
	}

	return result
}

func FormatFinding(finding ValidationFinding) string {
	cleanPath := strings.ReplaceAll(finding.Path, "root.", "")

	if cleanPath == "root" {
		cleanPath = "root"
	}

	requiredOptional := "optional"
	if finding.Required {
		requiredOptional = "required"
	}

	blockOrProp := "property"
	if finding.IsBlock {
		blockOrProp = "block"
	}

	entityType := "resource"
	if finding.IsDataSource {
		entityType = "data source"
	}

	place := cleanPath
	if finding.SubmoduleName != "" {
		place = place + " in submodule " + finding.SubmoduleName
	}

	return fmt.Sprintf("%s: missing %s %s %s in %s (%s)",
		finding.ResourceType, requiredOptional, blockOrProp, finding.Name, place, entityType)
}
