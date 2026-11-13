package entrypoints

import (
	json "encoding/json/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/leafevans/go-fuck/v2/internal/conf"
	"github.com/leafevans/go-fuck/v2/internal/consts"
)

// writeTestFile creates a file below dir, creating the parent directories.
func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
}

// shellConfigDir returns the directory that stores the shell configuration
// files in the test environment.
func shellConfigDir(t *testing.T) string {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() error = %v", err)
	}
	return home
}

func TestDoctorReportsShellIntegration(t *testing.T) {
	resetEntrypointsState(t)
	home := shellConfigDir(t)
	writeTestFile(t, filepath.Join(home, ".bashrc"), "eval \"$(go-fuck --alias)\"\n")
	writeTestFile(t, filepath.Join(home, ".bash_history"), "git status\n")

	code, stdout, stderr := runMain(t, "go-fuck", "--doctor")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0; stderr = %q", code, stderr)
	}
	for _, want := range []string{
		"go-fuck doctor",
		consts.Version,
		"rules",
		"shell",
		"Bash",
		"alias",
		"found in ~/.bashrc",
		"config",
		"history",
		"instant",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("doctor output %q does not contain %q", stdout, want)
		}
	}
}

func TestDoctorJSON(t *testing.T) {
	resetEntrypointsState(t)

	code, stdout, stderr := runMain(t, "go-fuck", "--doctor", "--json")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0; stderr = %q", code, stderr)
	}
	if strings.Contains(stdout, "go-fuck doctor") {
		t.Fatalf("the JSON report contains the human header: %q", stdout)
	}
	var report doctorReport
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatalf("json.Unmarshal(%q) error = %v", stdout, err)
	}
	if report.Version != consts.Version {
		t.Fatalf("report.Version = %q, want %q", report.Version, consts.Version)
	}
	names := map[string]bool{}
	for _, check := range report.Checks {
		names[check.Name] = true
	}
	for _, want := range []string{
		"version", "binary", "shell", "alias", "config", "history", "fzf", "instant",
	} {
		if !names[want] {
			t.Fatalf("the report has no %q check: %+v", want, report.Checks)
		}
	}
}

func TestDoctorWarnsAboutMissingAlias(t *testing.T) {
	resetEntrypointsState(t)

	code, stdout, _ := runMain(t, "go-fuck", "--doctor")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	if !strings.Contains(stdout, "no alias found") {
		t.Fatalf("doctor output %q does not warn about the missing alias", stdout)
	}
	if !strings.Contains(stdout, "go-fuck --setup") {
		t.Fatalf("doctor output %q does not suggest --setup", stdout)
	}
}

func TestDoctorWarnsAboutUnknownSettings(t *testing.T) {
	resetEntrypointsState(t)
	settingsDir := conf.UserDir()
	writeTestFile(t, filepath.Join(settingsDir, "settings.json"),
		`{"rules": ["git_push"], "rule": ["typo"]}`)

	code, stdout, _ := runMain(t, "go-fuck", "--doctor")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	for _, want := range []string{"ignored unknown keys", "rule"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("doctor output %q does not contain %q", stdout, want)
		}
	}
}

func TestDoctorFailsOnInvalidSettings(t *testing.T) {
	resetEntrypointsState(t)
	settingsDir := conf.UserDir()
	writeTestFile(t, filepath.Join(settingsDir, "settings.json"), "{invalid")

	code, stdout, _ := runMain(t, "go-fuck", "--doctor")
	if code != 1 {
		t.Fatalf("Main() = %d, want 1", code)
	}
	if !strings.Contains(stdout, "invalid settings file") {
		t.Fatalf("doctor output %q does not report the invalid settings file", stdout)
	}
}

func TestSanitizeDoctorReportReplacesInvalidUTF8(t *testing.T) {
	invalid := string([]byte{0xff, 0xfe})
	report := doctorReport{
		Version:   invalid,
		GoVersion: invalid,
		Platform:  invalid,
		Binary:    invalid,
		Shell:     invalid,
		Checks: []doctorCheck{{
			Name:    invalid,
			Status:  doctorWarn,
			Message: invalid,
			Hint:    invalid,
		}},
	}

	sanitizeDoctorReport(&report)
	values := []string{report.Version, report.GoVersion, report.Platform, report.Binary, report.Shell,
		report.Checks[0].Name, report.Checks[0].Message, report.Checks[0].Hint}
	for index, value := range values {
		if !utf8.ValidString(value) {
			t.Fatalf("sanitized value %d is not valid UTF-8: %q", index, value)
		}
		if !strings.Contains(value, string(utf8.RuneError)) {
			t.Fatalf("sanitized value %d = %q, want replacement characters", index, value)
		}
	}
}

func TestRecoverEntrypoints(t *testing.T) {
	code, _, stderr := captureOutput(t, func() int {
		return recoverEntrypoints(func() int { panic("boom") })
	})
	if code != 1 {
		t.Fatalf("recoverEntrypoints() = %d, want 1", code)
	}
	for _, want := range []string{"internal error", "boom", "github.com/leafevans/go-fuck/issues"} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("stderr = %q, does not contain %q", stderr, want)
		}
	}
}

func TestDoctorLogHasMarker(t *testing.T) {
	dir := t.TempDir()
	if doctorLogHasMarker(filepath.Join(dir, "missing")) {
		t.Fatal("doctorLogHasMarker(missing) = true")
	}
	plain := filepath.Join(dir, "plain.log")
	if err := os.WriteFile(plain, []byte("no marker here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if doctorLogHasMarker(plain) {
		t.Fatal("doctorLogHasMarker(plain) = true")
	}
	marked := filepath.Join(dir, "marked.log")
	if err := os.WriteFile(marked, []byte("prefix "+consts.UserCommandMark+"$ "), 0o644); err != nil {
		t.Fatal(err)
	}
	if !doctorLogHasMarker(marked) {
		t.Fatal("doctorLogHasMarker(marked) = false")
	}
}
