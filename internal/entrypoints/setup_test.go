package entrypoints

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readFile reads a file in tests.
func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func TestMainSetupBash(t *testing.T) {
	configDir := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "bash")
	configPath := filepath.Join(configDir, ".bashrc")
	if err := os.WriteFile(configPath, []byte("export EXISTING=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, _, stderr := runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("Main(--setup) = %d, want 0; stderr = %s", code, stderr)
	}
	if !strings.Contains(stderr, "now configured") {
		t.Fatalf("stderr = %q, want a success message", stderr)
	}
	content := readFile(t, configPath)
	if !strings.Contains(content, "export EXISTING=1") {
		t.Fatalf("existing configuration was lost: %q", content)
	}
	if got := strings.Count(content, "go-fuck --alias"); got != 1 {
		t.Fatalf("configuration contains %d alias lines, want 1: %q", got, content)
	}

	code, _, stderr = runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("second Main(--setup) = %d, want 0", code)
	}
	if !strings.Contains(stderr, "already configured") {
		t.Fatalf("stderr = %q, want an already configured message", stderr)
	}
	if got := strings.Count(readFile(t, configPath), "go-fuck --alias"); got != 1 {
		t.Fatalf("configuration contains %d alias lines after the second run, want 1", got)
	}
}

func TestMainSetupAppendsNewline(t *testing.T) {
	configDir := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "bash")
	configPath := filepath.Join(configDir, ".bashrc")
	if err := os.WriteFile(configPath, []byte("export EXISTING=1"), 0o644); err != nil {
		t.Fatal(err)
	}

	if code, _, stderr := runMain(t, "go-fuck", "--setup"); code != 0 {
		t.Fatalf("Main(--setup) = %d, want 0; stderr = %s", code, stderr)
	}
	content := readFile(t, configPath)
	if !strings.HasPrefix(content, "export EXISTING=1\n") {
		t.Fatalf("configuration was not separated by a newline: %q", content)
	}
}

func TestMainSetupManualPowerShell(t *testing.T) {
	resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "pwsh")

	code, _, stderr := runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("Main(--setup) = %d, want 0", code)
	}
	if !strings.Contains(stderr, "Add the following line") || !strings.Contains(stderr, "go-fuck --alias") {
		t.Fatalf("stderr = %q, want manual configuration instructions", stderr)
	}
}

func TestMainSetupMissingConfigFile(t *testing.T) {
	resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "zsh")

	code, _, stderr := runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("Main(--setup) = %d, want 0", code)
	}
	if !strings.Contains(stderr, "~/.zshrc") {
		t.Fatalf("stderr = %q, want the zsh configuration path", stderr)
	}
}
