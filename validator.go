package diffy

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// DefaultSchemaValidator implements SchemaValidator
type DefaultSchemaValidator struct {
	logger Logger
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator(logger Logger) *DefaultSchemaValidator {
	return &DefaultSchemaValidator{
		logger: logger,
	}
}

// ValidateResources validates resources against a schema
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

// ValidateDataSources validates data sources against a schema
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

// ValidateTerraformSchema validates a directory against Terraform schema
func ValidateTerraformSchema(logger Logger, dir, submoduleName string, parser HCLParser, runner TerraformRunner) ([]ValidationFinding, error) {
	ctx := context.Background()

	// Parse provider requirements
	tfFile := filepath.Join(dir, "terraform.tf")
	providers, err := parser.ParseProviderRequirements(ctx, tfFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provider config in %s: %w", dir, err)
	}

	// Initialize terraform
	if err := runner.Init(ctx, dir); err != nil {
		return nil, err
	}

	// Get schema
	tfSchema, err := runner.GetSchema(ctx, dir)
	if err != nil {
		return nil, err
	}

	// Parse main file
	mainTf := filepath.Join(dir, "main.tf")
	resources, dataSources, err := parser.ParseMainFile(ctx, mainTf)
	if err != nil {
		return nil, fmt.Errorf("parseMainFile in %s: %w", dir, err)
	}

	// Validate resources and data sources
	validator := NewSchemaValidator(logger)
	var findings []ValidationFinding
	findings = append(findings, validator.ValidateResources(resources, *tfSchema, providers, dir, submoduleName)...)
	findings = append(findings, validator.ValidateDataSources(dataSources, *tfSchema, providers, dir, submoduleName)...)

	return findings, nil
}

// DeduplicateFindings removes duplicate findings
func DeduplicateFindings(findings []ValidationFinding) []ValidationFinding {
	seen := make(map[string]bool)
	var result []ValidationFinding

	for _, f := range findings {
		key := fmt.Sprintf("%s|%s|%s|%v|%v|%s",
			f.ResourceType,
			f.Path,
			f.Name,
			f.IsBlock,
			f.IsDataSource,
			f.SubmoduleName,
		)

		if !seen[key] {
			seen[key] = true
			result = append(result, f)
		}
	}

	return result
}

// FormatFinding formats a validation finding as a string
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
