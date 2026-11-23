package entrypoints

import (
	"bytes"
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v3/internal/consts"
	"github.com/leafevans/go-fuck/v3/internal/logs"
	"github.com/leafevans/go-fuck/v3/internal/rules"
	"github.com/leafevans/go-fuck/v3/internal/shells"
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
	if info.Executable == "" {
		t.Fatal("executable path is empty")
	}
	if !strings.Contains(info.Module, "go-fuck") {
		t.Fatalf("module = %q, want the go-fuck module", info.Module)
	}
}

type versionTestShell struct{ shells.Shell }

func (versionTestShell) Info() string { return "Bash 5.2" }

func TestPrintVersionRedirected(t *testing.T) {
	resetEntrypointsState(t)
	shells.Set(versionTestShell{shells.NewBash()})
	var output bytes.Buffer
	logs.SetOutput(&output)
	for _, tc := range []struct {
		verbose bool
		shell   string
	}{{false, "Bash"}, {true, "Bash 5.2"}} {
		output.Reset()
		printVersionTo(&output, tc.verbose)
		want := consts.AppName + " " + consts.Version + " using Go " + runtime.Version() + " and " + tc.shell + "\n"
		if got := output.String(); got != want {
			t.Errorf("printVersionTo(%t) = %q, want %q", tc.verbose, got, want)
		}
	}
}

func TestVersionInfoCompactBanner(t *testing.T) {
	resetEntrypointsState(t)
	shells.Set(versionTestShell{shells.NewBash()})
	info := versionInfo(false)
	if info.Shell != "Bash 5.2" || info.Rules == 0 {
		t.Fatalf("compact banner missing shell version or rule count: %+v", info)
	}
	if info.ConfigPath != "" || info.InstantMode != "" || info.Fzf || info.Alias != "" || info.Executable != "" || info.Module != "" {
		t.Fatalf("compact banner collected verbose fields: %+v", info)
	}
}
