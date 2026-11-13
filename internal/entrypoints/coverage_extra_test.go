package entrypoints

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/argumentparser"
	"github.com/leafevans/go-fuck/v2/internal/logs"
	"github.com/leafevans/go-fuck/v2/internal/shells"
)

// helperEnv selects the behavior of the test binary when it is executed as a
// fake external program.
const helperEnv = "GOFUCK_ENTRYPOINT_HELPER"

// TestMain lets the test binary act as a fake external program.
func TestMain(m *testing.M) {
	if mode := os.Getenv(helperEnv); mode != "" {
		os.Exit(runHelper(mode))
	}
	os.Exit(m.Run())
}

// runHelper implements the fake program modes.
func runHelper(mode string) int {
	switch mode {
	case "profile":
		fmt.Print("/tmp/go-fuck-fake-profile.ps1")
		return 0
	case "profile-empty":
		fmt.Print("   ")
		return 0
	case "fail":
		return 3
	}
	return 2
}

// installFakeExecutable copies the test binary into a temporary directory under
// the given name and points PATH at it.
func installFakeExecutable(t *testing.T, name, mode string) {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() error = %v", err)
	}
	data, err := os.ReadFile(executable)
	if err != nil {
		t.Fatalf("read the test binary: %v", err)
	}
	fileName := name
	if runtime.GOOS == "windows" {
		fileName += ".exe"
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, fileName), data, 0o755); err != nil {
		t.Fatalf("write the fake executable: %v", err)
	}
	t.Setenv("PATH", dir)
	t.Setenv(helperEnv, mode)
}

// makeReadOnly removes the write permission of a file. Tests that depend on the
// permission being enforced skip themselves when running as root.
func makeReadOnly(t *testing.T, path string) {
	t.Helper()
	if err := os.Chmod(path, 0o444); err != nil {
		t.Fatalf("make %s read only: %v", path, err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })
}

func TestSetupReportsSettingsLoadFailure(t *testing.T) {
	configDir := resetEntrypointsState(t)
	settingsPath := filepath.Join(configDir, "go-fuck", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, []byte("{invalid"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, _, stderr := runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("Main(--setup) = %d, want 0; stderr = %s", code, stderr)
	}
	if !strings.Contains(stderr, "can't load settings") {
		t.Fatalf("stderr = %q, want a settings load failure", stderr)
	}
}

func TestSetupAppendFailure(t *testing.T) {
	configDir := resetEntrypointsState(t)
	configPath := filepath.Join(configDir, ".bashrc")
	if err := os.WriteFile(configPath, []byte("export EXISTING=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	makeReadOnly(t, configPath)

	code, _, stderr := runMain(t, "go-fuck", "--setup")
	if code == 0 {
		t.Skip("the read only configuration file could be written (running as root?)")
	}
	if code != 1 || !strings.Contains(stderr, "can't update") {
		t.Fatalf("Main(--setup) = %d, stderr = %q, want a failed update", code, stderr)
	}
}

func TestSetupReplaceInstantModeLineFailure(t *testing.T) {
	configDir := resetEntrypointsState(t)
	configPath := filepath.Join(configDir, ".bashrc")
	legacy := "eval $(go-fuck --alias --enable-experimental-instant-mode)\n"
	if err := os.WriteFile(configPath, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	makeReadOnly(t, configPath)

	code, _, stderr := runMain(t, "go-fuck", "--setup", "--enable-experimental-instant-mode")
	if code == 0 {
		t.Skip("the read only configuration file could be written (running as root?)")
	}
	if code != 1 || !strings.Contains(stderr, "can't update") {
		t.Fatalf("Main(--setup) = %d, stderr = %q, want a failed update", code, stderr)
	}
}

func TestSetupPowerShellFailures(t *testing.T) {
	t.Run("profile is a directory", func(t *testing.T) {
		resetEntrypointsState(t)
		t.Setenv("TF_SHELL", "powershell")
		stubPowerShellProfiles(t, []string{t.TempDir()})

		code, _, stderr := runMain(t, "go-fuck", "--setup")
		if code != 1 || !strings.Contains(stderr, "can't read") {
			t.Fatalf("Main(--setup) = %d, stderr = %q, want a failed read", code, stderr)
		}
	})

	t.Run("profile directory cannot be created", func(t *testing.T) {
		resetEntrypointsState(t)
		t.Setenv("TF_SHELL", "powershell")
		var profile string
		if runtime.GOOS == "windows" {
			// A path below a regular file is reported as a missing file on
			// Windows, so the profile lookup passes and creating its directory
			// fails.
			blocker := filepath.Join(t.TempDir(), "blocker")
			if err := os.WriteFile(blocker, nil, 0o644); err != nil {
				t.Fatal(err)
			}
			profile = filepath.Join(blocker, "Documents", "profile.ps1")
		} else {
			// A read only parent directory makes the profile lookup report a
			// missing file and creating its directory fail.
			parent := t.TempDir()
			if err := os.Chmod(parent, 0o500); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(parent, 0o755) })
			profile = filepath.Join(parent, "Documents", "profile.ps1")
		}
		stubPowerShellProfiles(t, []string{profile})

		code, _, stderr := runMain(t, "go-fuck", "--setup")
		if code == 0 {
			t.Skip("the missing profile directory could be created (running as root?)")
		}
		if code != 1 || !strings.Contains(stderr, "can't create") {
			t.Fatalf("Main(--setup) = %d, stderr = %q, want a failed directory creation", code, stderr)
		}
	})

	t.Run("profile cannot be appended", func(t *testing.T) {
		resetEntrypointsState(t)
		t.Setenv("TF_SHELL", "powershell")
		profile := filepath.Join(t.TempDir(), "profile.ps1")
		if err := os.WriteFile(profile, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		makeReadOnly(t, profile)
		stubPowerShellProfiles(t, []string{profile})

		code, _, stderr := runMain(t, "go-fuck", "--setup")
		if code == 0 {
			t.Skip("the read only profile could be written (running as root?)")
		}
		if code != 1 || !strings.Contains(stderr, "can't update") {
			t.Fatalf("Main(--setup) = %d, stderr = %q, want a failed update", code, stderr)
		}
	})
}

func TestPowerShellProfileFakeExecutable(t *testing.T) {
	cases := []struct {
		name     string
		mode     string
		wantPath string
		wantOK   bool
	}{
		{name: "reports a profile", mode: "profile", wantPath: "/tmp/go-fuck-fake-profile.ps1", wantOK: true},
		{name: "fails", mode: "fail"},
		{name: "reports an empty profile", mode: "profile-empty"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			resetEntrypointsState(t)
			installFakeExecutable(t, "pwsh", testCase.mode)

			path, ok := powerShellProfile("pwsh")
			if ok != testCase.wantOK {
				t.Fatalf("powerShellProfile() = %q, %t, want ok = %t", path, ok, testCase.wantOK)
			}
			if path != testCase.wantPath {
				t.Fatalf("powerShellProfile() = %q, want %q", path, testCase.wantPath)
			}
		})
	}
}

func TestFixCommandReportsSettingsLoadFailure(t *testing.T) {
	configDir := resetEntrypointsState(t)
	settingsPath := filepath.Join(configDir, "go-fuck", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, []byte("{invalid"), 0o644); err != nil {
		t.Fatal(err)
	}
	buffer := &bytes.Buffer{}
	logs.SetOutput(buffer)

	code, _, stderr := runMain(t, "go-fuck", "--force-command", "   ")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0; stderr = %s", code, stderr)
	}
	if !strings.Contains(buffer.String(), "Can't load settings") {
		t.Fatalf("log output = %q, want a settings load failure", buffer.String())
	}
}

func TestMainFixCommandEmptyRawCommand(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{name: "text output", args: []string{"go-fuck"}},
		{name: "json output", args: []string{"go-fuck", "--json"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			resetEntrypointsState(t)
			t.Setenv("TF_ALIAS", "go-fuck-coverage-alias")
			t.Setenv("TF_HISTORY", "go-fuck-coverage-alias")

			if code, _, stderr := runMain(t, testCase.args...); code != 0 {
				t.Fatalf("Main(%v) = %d, want 0; stderr = %s", testCase.args, code, stderr)
			}
		})
	}
}

func TestMainFixCommandBlankCommand(t *testing.T) {
	resetEntrypointsState(t)
	if code, _, stderr := runMain(t, "go-fuck", "--force-command", "   "); code != 0 {
		t.Fatalf("Main() = %d, want 0; stderr = %s", code, stderr)
	}
}

func TestGetRawCommandMarkedGenericShell(t *testing.T) {
	resetRawCommandState(t, t.TempDir())
	shells.Set(shells.NewGeneric())
	t.Setenv("TF_HISTORY", rawCommandMarker+"git status")

	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), []string{"git status"})
}

func TestGetRawCommandBashMarkedEdgeCases(t *testing.T) {
	resetRawCommandState(t, t.TempDir())
	shells.Set(shells.NewBash())

	// The entry before the alias call is empty.
	t.Setenv("TF_HISTORY", rawCommandMarker+"\t \n\t fuck")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), nil)

	// The last entry is empty and does not look like the alias.
	t.Setenv("TF_HISTORY", rawCommandMarker+"\t ")
	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), nil)
}

func TestGetRawCommandPowerShellEmptyHistory(t *testing.T) {
	resetRawCommandState(t, t.TempDir())
	shells.Set(shells.NewPowershell())
	t.Setenv("TF_HISTORY", "\n\n")

	assertRawCommand(t, getRawCommand(&argumentparser.Args{}), nil)
}
