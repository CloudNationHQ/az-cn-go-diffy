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

func (v *DefaultSchemaValidator) ValidateResources(
	resources []ParsedResource,
	schema TerraformSchema,
	providers map[string]ProviderConfig,
	dir, submoduleName string,
) []ValidationFinding {
	var findings []ValidationFinding

	for _, r := range resources {
		provName := strings.SplitN(r.Type, "_", 2)[0]
		cfg, ok := providers[provName]
		if !ok {
			v.logger.Logf("No provider config for resource type %s in %s", r.Type, dir)
			continue
		}

		pSchema, ok := schema.ProviderSchemas[cfg.Source]
		if !ok {
			v.logger.Logf("No provider schema found for source %s in %s", cfg.Source, dir)
			continue
		}

		resSchema, ok := pSchema.ResourceSchemas[r.Type]
		if !ok {
			v.logger.Logf("No resource schema found for %s in provider %s (dir=%s)", r.Type, cfg.Source, dir)
			continue
		}

		var local []ValidationFinding
		r.Data.Validate(r.Type, "root", resSchema.Block, r.Data.IgnoreChanges, &local)

		for i := range local {
			shouldExclude := false
			for _, ignored := range r.Data.IgnoreChanges {
				if strings.EqualFold(ignored, local[i].Name) {
					shouldExclude = true
					break
				}
			}

			if !shouldExclude {
				local[i].SubmoduleName = submoduleName
				findings = append(findings, local[i])
			}
		}
	}

	return findings
}

func (v *DefaultSchemaValidator) ValidateDataSources(
	dataSources []ParsedDataSource,
	schema TerraformSchema,
	providers map[string]ProviderConfig,
	dir, submoduleName string,
) []ValidationFinding {
	var findings []ValidationFinding

	for _, ds := range dataSources {
		provName := strings.SplitN(ds.Type, "_", 2)[0]
		cfg, ok := providers[provName]
		if !ok {
			v.logger.Logf("No provider config for data source type %s in %s", ds.Type, dir)
			continue
		}

		pSchema, ok := schema.ProviderSchemas[cfg.Source]
		if !ok {
			v.logger.Logf("No provider schema found for source %s in %s", cfg.Source, dir)
			continue
		}

		dsSchema, ok := pSchema.DataSourceSchemas[ds.Type]
		if !ok {
			v.logger.Logf("No data source schema found for %s in provider %s (dir=%s)", ds.Type, cfg.Source, dir)
			continue
		}

		var local []ValidationFinding
		ds.Data.Validate(ds.Type, "root", dsSchema.Block, ds.Data.IgnoreChanges, &local)

		for i := range local {
			shouldExclude := false
			for _, ignored := range ds.Data.IgnoreChanges {
				if strings.EqualFold(ignored, local[i].Name) {
					shouldExclude = true
					break
				}
			}

			if !shouldExclude {
				local[i].SubmoduleName = submoduleName
				local[i].IsDataSource = true
				findings = append(findings, local[i])
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
	for _, r := range resources {
		if !slices.Contains(excluded, r.Type) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func filterDataSources(dataSources []ParsedDataSource, excluded []string) []ParsedDataSource {
	if len(excluded) == 0 {
		return dataSources
	}
	
	var filtered []ParsedDataSource
	for _, ds := range dataSources {
		if !slices.Contains(excluded, ds.Type) {
			filtered = append(filtered, ds)
		}
	}
	return filtered
}

func DeduplicateFindings(findings []ValidationFinding) []ValidationFinding {
	seen := make(map[string]struct{})
	result := make([]ValidationFinding, 0, len(findings))

	for _, f := range findings {
		key := fmt.Sprintf("%s|%s|%s|%v|%v|%s",
			f.ResourceType,
			f.Path,
			f.Name,
			f.IsBlock,
			f.IsDataSource,
			f.SubmoduleName,
		)

		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, f)
		}
	}

	return result
}

func FormatFinding(f ValidationFinding) string {
	cleanPath := strings.ReplaceAll(f.Path, "root.", "")

	if cleanPath == "root" {
		cleanPath = "root"
	}

	requiredOptional := "optional"
	if f.Required {
		requiredOptional = "required"
	}

	blockOrProp := "property"
	if f.IsBlock {
		blockOrProp = "block"
	}

	entityType := "resource"
	if f.IsDataSource {
		entityType = "data source"
	}

	place := cleanPath
	if f.SubmoduleName != "" {
		place = place + " in submodule " + f.SubmoduleName
	}

	return fmt.Sprintf("%s: missing %s %s %s in %s (%s)",
		f.ResourceType, requiredOptional, blockOrProp, f.Name, place, entityType)
}
