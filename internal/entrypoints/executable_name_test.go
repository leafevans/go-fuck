package entrypoints

import (
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/consts"
)

func TestMainAliasUsesExecutableName(t *testing.T) {
	resetEntrypointsState(t)

	code, stdout, _ := runMain(t, "gof", "--alias")
	if code != 0 {
		t.Fatalf("Main(--alias) = %d, want 0", code)
	}
	if !strings.Contains(stdout, "gof") || strings.Contains(stdout, consts.AppName) {
		t.Fatalf("alias output does not use the executable name: %q", stdout)
	}
}

func TestMainUsageUsesExecutableName(t *testing.T) {
	resetEntrypointsState(t)

	code, _, stderr := runMain(t, "gof", "--help")
	if code != 0 {
		t.Fatalf("Main(--help) = %d, want 0", code)
	}
	if !strings.Contains(stderr, "usage: gof ") {
		t.Fatalf("help output does not use the executable name: %q", stderr)
	}
}
