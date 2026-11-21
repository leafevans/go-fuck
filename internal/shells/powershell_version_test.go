package shells

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/leafevans/go-fuck/v2/internal/consts"
)

func TestPowershellVersion(t *testing.T) {
	dir := t.TempDir()
	core := filepath.Join(dir, "pwsh.exe")
	legacy := filepath.Join(dir, "powershell.exe")
	for _, tc := range []struct {
		name, path, reported, probed, want string
		probe                              bool
	}{
		{"current alias", legacy, "7.7.0-preview.1", "", "7.7.0-preview.1", false},
		{"older alias core", core, "", "7.7.0", "7.7.0", true},
		{"older alias Windows", legacy, "", "5.1.2.3", "5.1.2.3", true},
		{"failed probe", core, "", "", "", true},
		{"no caller", "", "", "", "", false},
		{"relative caller", "pwsh.exe", "", "", "", false},
		{"other executable", filepath.Join(dir, "other.exe"), "", "", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(consts.ShellPathEnv, tc.path)
			t.Setenv(consts.PowerShellVersionEnv, tc.reported)
			original := runPowershellVersionProbe
			t.Cleanup(func() { runPowershellVersionProbe = original })
			var calls []string
			runPowershellVersionProbe = func(name string, args ...string) string {
				calls = append(calls, name)
				wantArgs := []string{"-NoLogo", "-NoProfile", "-NonInteractive", "-Command", "$PSVersionTable.PSVersion.ToString()"}
				if !slices.Equal(args, wantArgs) {
					t.Errorf("%s arguments = %q, want %q", name, args, wantArgs)
				}
				return tc.probed
			}
			want := "PowerShell"
			if tc.want != "" {
				want += " " + tc.want
			}
			if got := NewPowershell().Info(); got != want {
				t.Errorf("Info() = %q, want %q", got, want)
			}
			var wantCalls []string
			if tc.probe {
				wantCalls = []string{tc.path}
			}
			if !slices.Equal(calls, wantCalls) {
				t.Errorf("probed %q, want %q", calls, wantCalls)
			}
		})
	}
}

func TestPowershellProbeTimeout(t *testing.T) {
	const helperEnv = "GOFUCK_TEST_STALLED_POWERSHELL"
	if os.Getenv(helperEnv) == "1" {
		time.Sleep(10 * time.Second)
		return
	}
	t.Setenv(helperEnv, "1")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	if got := powershellProbeOutput(100*time.Millisecond, exe, "-test.run=^TestPowershellProbeTimeout$"); got != "" {
		t.Fatalf("stalled probe output = %q, want empty", got)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("stalled probe took %v to stop", elapsed)
	}
}
