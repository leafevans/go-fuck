package entrypoints

import (
	"strings"
	"testing"
)

func TestMainAliasUsesExecutableName(t *testing.T) {
	resetEntrypointsState(t)

	code, stdout, _ := runMain(t, "gof", "--alias")
	if code != 0 {
		t.Fatalf("Main(--alias) = %d, want 0", code)
	}
	if !strings.Contains(stdout, "gof --shell-logger") {
		t.Fatalf("alias output does not use the executable name: %q", stdout)
	}
}

func TestMainUsageUsesExecutableName(t *testing.T) {
	resetEntrypointsState(t)

	code, stdout, stderr := runMain(t, "gof", "--help")
	if code != 0 {
		t.Fatalf("Main(--help) = %d, want 0", code)
	}
	if stderr != "" {
		t.Fatalf("help stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "gof help") || !strings.Contains(stdout, "  gof [") {
		t.Fatalf("help output does not use the executable name: %q", stdout)
	}
}
