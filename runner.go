package diffy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type DefaultTerraformRunner struct{}

func NewTerraformRunner() *DefaultTerraformRunner {
	return &DefaultTerraformRunner{}
}

func (r *DefaultTerraformRunner) Init(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "terraform", "init")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform init failed in %s: %w\nOutput: %s", dir, err, string(output))
	}
	return nil
}

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

func ValidateTerraformSchemaInDirectory(logger Logger, dir, submoduleName string) ([]ValidationFinding, error) {
	return ValidateTerraformSchemaInDirectoryWithOptions(logger, dir, submoduleName, nil, nil)
}

func ValidateTerraformSchemaInDirectoryWithOptions(logger Logger, dir, submoduleName string, excludedResources, excludedDataSources []string) ([]ValidationFinding, error) {
	mainTf := filepath.Join(dir, "main.tf")
	if _, err := os.Stat(mainTf); os.IsNotExist(err) {
		return []ValidationFinding{}, nil
	}

	parser := NewHCLParser()
	runner := NewTerraformRunner()

	defer func() {
		os.RemoveAll(filepath.Join(dir, ".terraform"))
		os.Remove(filepath.Join(dir, "terraform.tfstate"))
		os.Remove(filepath.Join(dir, ".terraform.lock.hcl"))
	}()

	return ValidateTerraformSchemaWithOptions(logger, dir, submoduleName, parser, runner, excludedResources, excludedDataSources)
}
