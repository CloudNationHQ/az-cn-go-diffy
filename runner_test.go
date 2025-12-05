package diffy

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultTerraformRunnerInitCachesByDir(t *testing.T) {
	helperDir := t.TempDir()
	logFile := filepath.Join(helperDir, "log.txt")
	script := filepath.Join(helperDir, "terraform")

	writeExecutable(t, script, `#!/bin/sh
echo "init:$PWD" >> "`+logFile+`"
`)

	t.Setenv("PATH", helperDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	dir := t.TempDir()

	runner := NewTerraformRunner()

	if err := runner.Init(context.Background(), dir); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if err := runner.Init(context.Background(), dir); err != nil {
		t.Fatalf("second Init returned error: %v", err)
	}

	logContent := readFile(t, logFile)
	lines := strings.Split(strings.TrimSpace(logContent), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected terraform init to run once, got %d entries: %q", len(lines), logContent)
	}
	if !strings.Contains(lines[0], dir) {
		t.Fatalf("log entry should include working dir %s, got %q", dir, lines[0])
	}
}

func TestDefaultTerraformRunnerGetSchemaCachesResult(t *testing.T) {
	helperDir := t.TempDir()
	logFile := filepath.Join(helperDir, "log.txt")
	script := filepath.Join(helperDir, "terraform")

	writeExecutable(t, script, `#!/bin/sh
if [ "$1" = "providers" ]; then
  echo '{"provider_schemas":{"registry.terraform.io/hashicorp/azurerm":{"resource_schemas":{},"data_source_schemas":{}}}}'
  echo "schema:$PWD" >> "`+logFile+`"
  exit 0
fi
echo "unexpected args: $@" >&2
exit 1
`)

	t.Setenv("PATH", helperDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	dir := t.TempDir()
	runner := NewTerraformRunner()

	s1, err := runner.GetSchema(context.Background(), dir)
	if err != nil {
		t.Fatalf("GetSchema returned error: %v", err)
	}
	if s1 == nil || len(s1.ProviderSchemas) == 0 {
		t.Fatalf("schema should be populated")
	}

	s2, err := runner.GetSchema(context.Background(), dir)
	if err != nil {
		t.Fatalf("second GetSchema returned error: %v", err)
	}

	if s1 != s2 {
		t.Fatalf("expected cached schema pointer, got different instances")
	}

	logContent := strings.TrimSpace(readFile(t, logFile))
	if logContent == "" {
		t.Fatalf("expected terraform providers schema command to run once")
	}
	if strings.Count(logContent, "\n")+1 != 1 {
		t.Fatalf("expected single schema invocation, got log: %q", logContent)
	}
}

func TestDefaultTerraformRunnerInitError(t *testing.T) {
	helperDir := t.TempDir()
	script := filepath.Join(helperDir, "terraform")
	writeExecutable(t, script, `#!/bin/sh
echo "boom" >&2
exit 2
`)
	t.Setenv("PATH", helperDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	runner := NewTerraformRunner()
	if err := runner.Init(context.Background(), t.TempDir()); err == nil {
		t.Fatalf("expected init error")
	}
}

func TestDefaultTerraformRunnerGetSchemaError(t *testing.T) {
	helperDir := t.TempDir()
	script := filepath.Join(helperDir, "terraform")
	writeExecutable(t, script, `#!/bin/sh
exit 3
`)
	t.Setenv("PATH", helperDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	runner := NewTerraformRunner()
	if _, err := runner.GetSchema(context.Background(), t.TempDir()); err == nil {
		t.Fatalf("expected schema error")
	}
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("failed to write helper script: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(data)
}
