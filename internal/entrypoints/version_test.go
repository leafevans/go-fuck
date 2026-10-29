package entrypoints

import (
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/rules"
	"github.com/leafevans/go-fuck/internal/shells"
)

func TestVersionInfo(t *testing.T) {
	resetEntrypointsState(t)
	shells.Set(shells.NewBash())

	info := versionInfo(true)
	if info.Version != consts.Version {
		t.Fatalf("version = %q, want %q", info.Version, consts.Version)
	}
	if info.GoVersion != runtime.Version() {
		t.Fatalf("Go version = %q, want %q", info.GoVersion, runtime.Version())
	}
	if want := runtime.GOOS + "/" + runtime.GOARCH; info.Platform != want {
		t.Fatalf("platform = %q, want %q", info.Platform, want)
	}
	if info.Rules != rules.Count() || info.Rules == 0 {
		t.Fatalf("rules = %d, want %d", info.Rules, rules.Count())
	}
	if !strings.Contains(info.Shell, "Bash") {
		t.Fatalf("shell = %q, want Bash", info.Shell)
	}
	if info.ConfigPath == "" {
		t.Fatal("config path is empty")
	}
	if info.InstantMode == "" {
		t.Fatal("instant mode state is empty")
	}
	if info.Alias != "fuck" {
		t.Fatalf("alias = %q, want fuck", info.Alias)
	}
}

// TestVersionInfoCompactShellName checks that the plain one line printed to a
// pipe does not need to detect the shell version.
func TestVersionInfoCompactShellName(t *testing.T) {
	resetEntrypointsState(t)
	shell := shells.NewBash()
	shells.Set(shell)

	info := versionInfo(false)
	if want := shell.FriendlyName(); info.Shell != want {
		t.Fatalf("shell = %q, want %q", info.Shell, want)
	}
}
