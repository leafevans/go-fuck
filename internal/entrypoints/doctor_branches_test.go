package entrypoints

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/conf"
)

// findDoctorCheck returns the named check of the report.
func findDoctorCheck(t *testing.T, checks []doctorCheck, name string) doctorCheck {
	t.Helper()
	for _, check := range checks {
		if check.Name == name {
			return check
		}
	}
	t.Fatalf("no %q check in %+v", name, checks)
	return doctorCheck{}
}

func TestDoctorChecksForShellStates(t *testing.T) {
	t.Cleanup(conf.Reset)
	conf.Reset()
	report := doctorReport{
		Version:   "1.2.3",
		GoVersion: "go1.27.3",
		Platform:  "linux/amd64",
		Binary:    "/usr/bin/go-fuck",
	}

	t.Run("bash with alias", func(t *testing.T) {
		checks := doctorChecks(report, doctorShellState{
			Name: "bash", Info: "Bash 5.2", ConfigPath: "~/.bashrc",
			AliasFound: true, History: 3,
		})
		if got := findDoctorCheck(t, checks, "shell").Status; got != doctorOK {
			t.Fatalf("shell status = %q, want ok", got)
		}
		alias := findDoctorCheck(t, checks, "alias")
		if alias.Status != doctorOK || !strings.Contains(alias.Message, "found in ~/.bashrc") {
			t.Fatalf("alias check = %+v", alias)
		}
		if got := findDoctorCheck(t, checks, "history").Status; got != doctorOK {
			t.Fatalf("history status = %q, want ok", got)
		}
	})

	t.Run("generic shell", func(t *testing.T) {
		checks := doctorChecks(report, doctorShellState{Name: "generic", Info: "Generic"})
		if got := findDoctorCheck(t, checks, "shell").Status; got != doctorWarn {
			t.Fatalf("shell status = %q, want warn", got)
		}
		alias := findDoctorCheck(t, checks, "alias")
		if alias.Status != doctorWarn || !strings.Contains(alias.Hint, "--alias") {
			t.Fatalf("alias check = %+v", alias)
		}
		if got := findDoctorCheck(t, checks, "history").Status; got != doctorInfo {
			t.Fatalf("history status = %q, want info", got)
		}
	})

	t.Run("powershell profile", func(t *testing.T) {
		original := doctorPowerShellProfiles
		t.Cleanup(func() { doctorPowerShellProfiles = original })
		profile := filepath.Join(t.TempDir(), "profile.ps1")
		if err := os.WriteFile(profile, []byte("iex \"$(go-fuck --alias)\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		doctorPowerShellProfiles = func() []string { return []string{profile} }

		checks := doctorChecks(report, doctorShellState{
			Name: "powershell", Info: "PowerShell 5.1", ConfigPath: "$profile",
		})
		alias := findDoctorCheck(t, checks, "alias")
		if alias.Status != doctorOK || !strings.Contains(alias.Message, profile) {
			t.Fatalf("alias check = %+v", alias)
		}
		if got := findDoctorCheck(t, checks, "history").Status; got != doctorInfo {
			t.Fatalf("history status = %q, want info", got)
		}
	})

	t.Run("unreadable configuration", func(t *testing.T) {
		checks := doctorChecks(report, doctorShellState{
			Name: "zsh", Info: "Zsh 5.9", ConfigPath: "~/.zshrc",
			ReadErr: errors.New("permission denied"),
		})
		alias := findDoctorCheck(t, checks, "alias")
		if alias.Status != doctorWarn || !strings.Contains(alias.Message, "no alias found") {
			t.Fatalf("alias check = %+v", alias)
		}
		if got := findDoctorCheck(t, checks, "history").Status; got != doctorWarn {
			t.Fatalf("history status = %q, want warn", got)
		}
	})

	t.Run("unknown profile path", func(t *testing.T) {
		checks := doctorChecks(report, doctorShellState{
			Name: "tcsh", Info: "Tcsh 6.24", ConfigPath: "$custom",
		})
		alias := findDoctorCheck(t, checks, "alias")
		if alias.Status != doctorInfo || !strings.Contains(alias.Message, "$custom") {
			t.Fatalf("alias check = %+v", alias)
		}
	})
}

func TestDoctorFzfCheck(t *testing.T) {
	t.Cleanup(conf.Reset)
	original := doctorIsAvailable
	t.Cleanup(func() { doctorIsAvailable = original })

	doctorIsAvailable = func(string) bool { return false }
	if got := doctorFzfCheck(); got.Status != doctorInfo {
		t.Fatalf("fzf check = %+v, want info", got)
	}

	doctorIsAvailable = func(string) bool { return true }
	settings := conf.Default()
	conf.Set(settings)
	if got := doctorFzfCheck(); got.Status != doctorInfo {
		t.Fatalf("disabled fzf check = %+v, want info", got)
	}

	settings.Fzf = true
	conf.Set(settings)
	if got := doctorFzfCheck(); got.Status != doctorOK {
		t.Fatalf("enabled fzf check = %+v, want ok", got)
	}
}

func TestDoctorReportFailsWithoutExecutable(t *testing.T) {
	resetEntrypointsState(t)
	original := doctorExecutable
	t.Cleanup(func() { doctorExecutable = original })
	doctorExecutable = func() (string, error) { return "", errors.New("no executable") }
	originalShell := doctorCollectShell
	t.Cleanup(func() { doctorCollectShell = originalShell })
	doctorCollectShell = func() doctorShellState {
		return doctorShellState{Name: "bash", Info: "Bash 5.2", History: 1}
	}

	report := collectDoctorReport()
	binary := findDoctorCheck(t, report.Checks, "binary")
	if binary.Status != doctorFail {
		t.Fatalf("binary check = %+v, want fail", binary)
	}
}

func TestExpandHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if got := expandHome("relative/path"); got != "relative/path" {
		t.Fatalf("expandHome(relative/path) = %q", got)
	}
	if got := expandHome("~/file"); got != filepath.Join(home, "file") {
		t.Fatalf("expandHome(~/file) = %q", got)
	}
	if _, err := os.Stat(home); err != nil {
		t.Fatalf("test home does not exist: %v", err)
	}
}

// TestDoctorPowerShellAliasCheck covers the profile lookup used for PowerShell,
// where $PROFILE is a variable and the alias can only be found by asking the
// installed flavors for their profile paths.
func TestDoctorPowerShellAliasCheck(t *testing.T) {
	original := doctorPowerShellProfiles
	t.Cleanup(func() { doctorPowerShellProfiles = original })

	t.Run("alias found in one flavor", func(t *testing.T) {
		dir := t.TempDir()
		configured := filepath.Join(dir, "Microsoft.PowerShell_profile.ps1")
		plain := filepath.Join(dir, "pwsh_profile.ps1")
		if err := os.WriteFile(configured, []byte("iex \"$(go-fuck --alias)\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(plain, []byte("Set-Alias ll Get-ChildItem\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		doctorPowerShellProfiles = func() []string { return []string{configured, plain} }

		check := doctorPowerShellAliasCheck()
		if check.Status != doctorOK {
			t.Fatalf("status = %q, want ok: %+v", check.Status, check)
		}
		if !strings.Contains(check.Message, configured) || strings.Contains(check.Message, plain) {
			t.Fatalf("message = %q, want the configured profile only", check.Message)
		}
	})

	t.Run("no alias in the profiles", func(t *testing.T) {
		profile := filepath.Join(t.TempDir(), "profile.ps1")
		if err := os.WriteFile(profile, []byte("Set-Alias ll Get-ChildItem\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		doctorPowerShellProfiles = func() []string { return []string{profile} }

		check := doctorPowerShellAliasCheck()
		if check.Status != doctorWarn || !strings.Contains(check.Message, profile) {
			t.Fatalf("status = %q, message = %q, want a warning for the profile", check.Status, check.Message)
		}
		if !strings.Contains(check.Hint, "--setup") {
			t.Fatalf("hint = %q, want the setup command", check.Hint)
		}
	})

	t.Run("profiles cannot be read", func(t *testing.T) {
		profile := filepath.Join(t.TempDir(), "missing", "profile.ps1")
		doctorPowerShellProfiles = func() []string { return []string{profile} }

		check := doctorPowerShellAliasCheck()
		if check.Status != doctorInfo || !strings.Contains(check.Message, profile) {
			t.Fatalf("status = %q, message = %q, want info for the unreadable profile", check.Status, check.Message)
		}
	})

	t.Run("no profiles", func(t *testing.T) {
		doctorPowerShellProfiles = func() []string { return nil }

		check := doctorPowerShellAliasCheck()
		if check.Status != doctorInfo || !strings.Contains(check.Message, "no PowerShell profile found") {
			t.Fatalf("status = %q, message = %q, want info without profiles", check.Status, check.Message)
		}
		if !strings.Contains(check.Hint, "--setup") {
			t.Fatalf("hint = %q, want the setup command", check.Hint)
		}
	})
}
