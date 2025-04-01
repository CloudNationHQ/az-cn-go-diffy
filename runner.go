package diffy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// DefaultTerraformRunner implements TerraformRunner
type DefaultTerraformRunner struct{}

// NewTerraformRunner creates a new Terraform runner
func NewTerraformRunner() *DefaultTerraformRunner {
	return &DefaultTerraformRunner{}
}

// Init runs terraform init in the specified directory
func (r *DefaultTerraformRunner) Init(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "terraform", "init")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform init failed in %s: %w\nOutput: %s", dir, err, string(output))
	}
	return nil
}

// GetSchema gets the provider schema using terraform providers schema
func (r *DefaultTerraformRunner) GetSchema(ctx context.Context, dir string) (*TerraformSchema, error) {
	cmd := exec.CommandContext(ctx, "terraform", "providers", "schema", "-json")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get schema in %s: %w", dir, err)
	}

	var tfSchema TerraformSchema
	if err := json.Unmarshal(output, &tfSchema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return &tfSchema, nil
}

// ValidateTerraformSchemaInDirectory validates the Terraform schema in a directory
func ValidateTerraformSchemaInDirectory(logger Logger, dir, submoduleName string) ([]ValidationFinding, error) {
	mainTf := filepath.Join(dir, "main.tf")
	if _, err := os.Stat(mainTf); os.IsNotExist(err) {
		return nil, nil
	}

	parser := NewHCLParser()
	runner := NewTerraformRunner()

	// Create cleanup function
	defer func() {
		os.RemoveAll(filepath.Join(dir, ".terraform"))
		os.Remove(filepath.Join(dir, "terraform.tfstate"))
		os.Remove(filepath.Join(dir, ".terraform.lock.hcl"))
	}()

	return ValidateTerraformSchema(logger, dir, submoduleName, parser, runner)
}

// ValidateTerraformProject validates an entire Terraform project including submodules
func ValidateTerraformProject(logger Logger, terraformRoot string) ([]ValidationFinding, error) {
	// Validate root directory
	rootFindings, err := ValidateTerraformSchemaInDirectory(logger, terraformRoot, "")
	if err != nil {
		return nil, fmt.Errorf("failed to validate root at %s: %v", terraformRoot, err)
	}

	var allFindings []ValidationFinding
	allFindings = append(allFindings, rootFindings...)

	// Validate submodules
	modulesDir := filepath.Join(terraformRoot, "modules")
	submodules, err := FindSubmodules(modulesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find submodules in %s: %v", modulesDir, err)
	}

	for _, sm := range submodules {
		findings, err := ValidateTerraformSchemaInDirectory(logger, sm.Path, sm.Name)
		if err != nil {
			logger.Logf("Failed to validate submodule %s: %v", sm.Name, err)
			continue
		}
		allFindings = append(allFindings, findings...)
	}

	// Deduplicate findings
	deduplicatedFindings := DeduplicateFindings(allFindings)

	return deduplicatedFindings, nil
}
