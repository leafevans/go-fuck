//go:build windows

package entrypoints

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/shells"
)

// instantProfile returns a temporary PowerShell profile with the given lines.
func instantProfile(t *testing.T, lines ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "profile.ps1")
	content := ""
	for _, line := range lines {
		content += line + "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	return path
}

// stubInstantProfiles points the doctor at the given profile files.
func stubInstantProfiles(t *testing.T, paths ...string) {
	t.Helper()
	originalShell := shells.Current()
	originalProfiles := powerShellProfiles
	shells.Set(shells.NewPowershell())
	powerShellProfiles = func() []string { return paths }
	t.Cleanup(func() {
		powerShellProfiles = originalProfiles
		shells.Set(originalShell)
	})
}

const instantProfileLine = "iex \"$(go-fuck --alias --enable-experimental-instant-mode)\""

func TestDoctorInstantProfileOrder(t *testing.T) {
	t.Setenv(consts.OutputLogEnv, "")
	t.Setenv(consts.LegacyOutputLogEnv, "")

	broken := instantProfile(t, instantProfileLine, "function prompt { \"PS> \" }")
	stubInstantProfiles(t, broken)
	if path, problem := doctorInstantProfileProblem(); !problem || path != broken {
		t.Fatalf("doctorInstantProfileProblem() = %q, %t, want the broken profile", path, problem)
	}
	if got := doctorInstantCheck(); got.Status != doctorWarn {
		t.Fatalf("doctorInstantCheck() = %+v, want warn", got)
	}
}

func TestDoctorInstantProfileOrderIsFine(t *testing.T) {
	t.Setenv(consts.OutputLogEnv, "")
	t.Setenv(consts.LegacyOutputLogEnv, "")

	good := instantProfile(t, "function prompt { \"PS> \" }", instantProfileLine)
	stubInstantProfiles(t, good)
	if _, problem := doctorInstantProfileProblem(); problem {
		t.Fatal("doctorInstantProfileProblem() reported a correct profile")
	}
	if got := doctorInstantCheck(); got.Status != doctorInfo {
		t.Fatalf("doctorInstantCheck() = %+v, want info", got)
	}
}

func TestDoctorInstantProfileWithoutInstantLine(t *testing.T) {
	t.Setenv(consts.OutputLogEnv, "")
	t.Setenv(consts.LegacyOutputLogEnv, "")

	plain := instantProfile(t, "function prompt { \"PS> \" }")
	stubInstantProfiles(t, plain)
	if _, problem := doctorInstantProfileProblem(); problem {
		t.Fatal("doctorInstantProfileProblem() reported a profile without the instant mode line")
	}
}
