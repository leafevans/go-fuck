package entrypoints

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
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
		checks := doctorChecks(report, doctorShellState{
			Name: "powershell", Info: "PowerShell 5.1", ConfigPath: "$profile",
		})
		alias := findDoctorCheck(t, checks, "alias")
		if alias.Status != doctorInfo || !strings.Contains(alias.Message, "PowerShell profile") {
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

func TestDoctorAICheck(t *testing.T) {
	t.Cleanup(conf.Reset)
	settings := conf.Default()
	conf.Set(settings)
	if got := doctorAICheck(); got.Status != doctorInfo {
		t.Fatalf("disabled AI check = %+v, want info", got)
	}

	settings.AIEnabled = true
	conf.Set(settings)
	if got := doctorAICheck(); got.Status != doctorWarn {
		t.Fatalf("AI check without endpoint = %+v, want warn", got)
	}

	settings.AIEndpoint = "https://api.example.com/v1"
	settings.AIModel = "example"
	conf.Set(settings)
	t.Setenv("GOFUCK_AI_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	if got := doctorAICheck(); got.Status != doctorWarn {
		t.Fatalf("AI check without key = %+v, want warn", got)
	}

	t.Setenv("GOFUCK_AI_API_KEY", "secret")
	if got := doctorAICheck(); got.Status != doctorOK {
		t.Fatalf("configured AI check = %+v, want ok", got)
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
