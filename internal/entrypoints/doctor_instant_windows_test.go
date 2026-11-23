//go:build windows

package entrypoints

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/consts"
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

const instantProfileLine = "iex \"$(go-fuck --alias --enable-experimental-instant-mode)\""

func TestDoctorInstantProfileOrder(t *testing.T) {
	t.Setenv(consts.OutputLogEnv, "")
	t.Setenv(consts.LegacyOutputLogEnv, "")

	broken := instantProfile(t, instantProfileLine, "function prompt { \"PS> \" }")
	profiles := []string{broken}
	if path, problem := doctorInstantProfileProblem(profiles); !problem || path != broken {
		t.Fatalf("doctorInstantProfileProblem() = %q, %t, want the broken profile", path, problem)
	}
	if got := doctorInstantCheck(profiles); got.Status != doctorWarn {
		t.Fatalf("doctorInstantCheck() = %+v, want warn", got)
	}
}

func TestDoctorInstantProfileOrderIsFine(t *testing.T) {
	t.Setenv(consts.OutputLogEnv, "")
	t.Setenv(consts.LegacyOutputLogEnv, "")

	good := instantProfile(t, "function prompt { \"PS> \" }", instantProfileLine)
	profiles := []string{good}
	if _, problem := doctorInstantProfileProblem(profiles); problem {
		t.Fatal("doctorInstantProfileProblem() reported a correct profile")
	}
	if got := doctorInstantCheck(profiles); got.Status != doctorInfo {
		t.Fatalf("doctorInstantCheck() = %+v, want info", got)
	}
}

func TestDoctorInstantProfileWithoutInstantLine(t *testing.T) {
	t.Setenv(consts.OutputLogEnv, "")
	t.Setenv(consts.LegacyOutputLogEnv, "")

	plain := instantProfile(t, "function prompt { \"PS> \" }")
	if _, problem := doctorInstantProfileProblem([]string{plain}); problem {
		t.Fatal("doctorInstantProfileProblem() reported a profile without the instant mode line")
	}
}

func TestDoctorUsesSamePowerShellProfilesForAliasAndInstantChecks(t *testing.T) {
	active := instantProfile(t, instantProfileLine, "function prompt { 'PS> ' }")
	hostProfile := filepath.Join(filepath.Dir(active), "Microsoft.PowerShell_profile.ps1")
	if err := os.WriteFile(hostProfile, []byte("# No alias in the current-host profile.\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fallback := instantProfile(t, instantProfileLine, "function global:prompt { 'PS> ' }")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(consts.OutputLogEnv, "")
	t.Setenv(consts.LegacyOutputLogEnv, "")
	original := doctorPowerShellProfiles
	t.Cleanup(func() { doctorPowerShellProfiles = original })

	for _, tt := range []struct {
		name, sessionPath, wantPath string
		wantLookups                 int
	}{
		{"current session, alias in all-hosts profile", hostProfile, active, 0},
		{"older alias", "", fallback, 1},
		{"relative path", "profile.ps1", fallback, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(consts.PowerShellProfileEnv, tt.sessionPath)
			lookups := 0
			doctorPowerShellProfiles = func() []string {
				lookups++
				return []string{fallback}
			}
			checks := doctorChecks(doctorReport{Binary: executable}, doctorShellState{
				Name: "powershell", Info: "PowerShell", ConfigPath: "$profile",
			})
			if lookups != tt.wantLookups {
				t.Errorf("profile lookups = %d, want %d", lookups, tt.wantLookups)
			}
			alias := findDoctorCheck(t, checks, "alias")
			if alias.Status != doctorOK || !strings.Contains(alias.Message, tt.wantPath) {
				t.Errorf("alias check = %+v, want ok for %q", alias, tt.wantPath)
			}
			instant := findDoctorCheck(t, checks, "instant")
			if instant.Status != doctorWarn || !strings.Contains(instant.Hint, tt.wantPath) {
				t.Errorf("instant check = %+v, want warning for %q", instant, tt.wantPath)
			}
		})
	}
}
