package entrypoints

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSetupPowerShellMigratesCapturePolicy(t *testing.T) {
	for _, legacy := range []string{
		`iex "$(go-fuck --alias)"`,
		`iex "$(go-fuck --alias --enable-experimental-instant-mode)"`,
	} {
		t.Run(legacy, func(t *testing.T) {
			if runtime.GOOS != "windows" && strings.Contains(legacy, "--enable-experimental-instant-mode") {
				t.Skip("PowerShell instant initialization is only supported on Windows")
			}
			home := resetEntrypointsState(t)
			t.Setenv("TF_SHELL", "powershell")
			profile := filepath.Join(home, "profile.ps1")
			stubPowerShellProfiles(t, []string{profile})
			original := "# keep\r\n  " + legacy + " # alias\r\n"
			if err := os.WriteFile(profile, []byte(original), 0o600); err != nil {
				t.Fatal(err)
			}
			for range 2 {
				code, _, stderr := runMain(t, "go-fuck", "--setup", "--capture-mode=rerun")
				if code != 0 || stderr != "" {
					t.Fatalf("setup = %d, stderr=%q", code, stderr)
				}
			}
			want := "# keep\r\n  iex \"$(go-fuck --alias --capture-mode=rerun)\" # alias\r\n"
			if got := readFile(t, profile); got != want {
				t.Fatalf("profile = %q, want %q", got, want)
			}
		})
	}
}

func TestSetupPowerShellRejectsAmbiguousCaptureMigration(t *testing.T) {
	home := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "powershell")
	profile := filepath.Join(home, "profile.ps1")
	stubPowerShellProfiles(t, []string{profile})
	original := strings.Repeat("iex \"$(go-fuck --alias)\"\n", 2)
	if err := os.WriteFile(profile, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runMain(t, "go-fuck", "--setup", "--capture-mode=rerun")
	if code != 1 || !strings.Contains(stderr, "multiple go-fuck initializations") {
		t.Fatalf("setup = %d, stderr=%q", code, stderr)
	}
	if got := readFile(t, profile); got != original {
		t.Fatalf("ambiguous profile changed: %q", got)
	}
}
