package rules

import (
	"os"
	"path/filepath"
	"testing"
)

// goMissingModOutput is the Go error when no module file exists.
const goMissingModOutput = "go: go.mod file not found in current directory or any parent directory; " +
	"see 'go help modules'\n"

// stubGoModuleName replaces the module name of the rule.
func stubGoModuleName(t *testing.T, name string) {
	t.Helper()
	original := goModuleName
	goModuleName = func() string { return name }
	t.Cleanup(func() { goModuleName = original })
}

func TestRuleGoModInit(t *testing.T) {
	resetTestState(t)
	rule := Get("go_mod_init")
	if rule == nil {
		t.Fatal("go_mod_init rule is not registered")
	}

	stubGoModuleName(t, "app")
	assertMatch(t, rule, "go build .", goMissingModOutput)
	assertMatch(t, rule, "go run main.go", goMissingModOutput)
	assertNotMatch(t, rule, "go build .", "")
	assertNotMatch(t, rule, "cargo build", goMissingModOutput)

	assertNewCommands(t, rule, "go build .", goMissingModOutput,
		[]string{"go mod init app && go build ."})
}

func TestGoModuleName(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "my app v2")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
	if got := goModuleName(); got != "my-app-v2" {
		t.Fatalf("goModuleName() = %q, want my-app-v2", got)
	}
}
