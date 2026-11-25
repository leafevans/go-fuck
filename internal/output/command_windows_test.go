//go:build windows

package output

import (
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/leafevans/go-fuck/v4/internal/consts"
)

// stubPowerShellAncestor replaces the process tree lookup of the package.
func stubPowerShellAncestor(t *testing.T, executable string) {
	t.Helper()
	original := powerShellAncestor
	powerShellAncestor = func() string { return executable }
	t.Cleanup(func() { powerShellAncestor = original })
}

// stubPowerShellPath replaces the executable lookup of the package.
func stubPowerShellPath(t *testing.T, path string) {
	t.Helper()
	original := powerShellPath
	powerShellPath = func(string) string { return path }
	t.Cleanup(func() { powerShellPath = original })
}

// clearShellPathEnv removes the alias environment variable so the tests that
// expect the process tree fallback are not influenced by a caller that runs the
// tests from a configured shell.
func clearShellPathEnv(t *testing.T) {
	t.Helper()
	t.Setenv(consts.ShellPathEnv, "")
}

// decodePowerShellCommand decodes a -EncodedCommand argument.
func decodePowerShellCommand(t *testing.T, encoded string) string {
	t.Helper()
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decode %q: %v", encoded, err)
	}
	units := make([]uint16, len(decoded)/2)
	for index := range units {
		units[index] = uint16(decoded[index*2]) | uint16(decoded[index*2+1])<<8
	}
	return string(utf16.Decode(units))
}

func TestPowerShellExecutable(t *testing.T) {
	testCases := []struct {
		name      string
		ancestors []string
		want      string
	}{
		{"none", nil, ""},
		{"no power shell", []string{"go.exe", "git.exe"}, ""},
		{"windows powershell", []string{"powershell.exe", "explorer.exe"}, "powershell.exe"},
		{"powershell 7", []string{"go.exe", "pwsh.exe", "WindowTerminal.exe"}, "pwsh.exe"},
		{"case insensitive", []string{"PowerShell.exe"}, "powershell.exe"},
		{"nearest wins", []string{"powershell.exe", "pwsh.exe"}, "powershell.exe"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := powerShellExecutable(testCase.ancestors); got != testCase.want {
				t.Fatalf("powerShellExecutable(%q) = %q, want %q",
					testCase.ancestors, got, testCase.want)
			}
		})
	}
}

func TestEncodePowerShellCommand(t *testing.T) {
	commands := []string{
		"",
		"echo hello",
		"cd C:\\missing",
		"git commit -m \"hello world\"",
		"echo \u2713",
		"echo \U0001f600",
	}
	for _, command := range commands {
		encoded := encodePowerShellCommand(command)
		if got := decodePowerShellCommand(t, encoded); got != command {
			t.Fatalf("encodePowerShellCommand(%q) decodes to %q", command, got)
		}
	}
}

func TestShellCommandUsesPowerShell(t *testing.T) {
	clearShellPathEnv(t)
	stubPowerShellAncestor(t, "powershell.exe")
	stubPowerShellPath(t, "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe")
	command := "cd C:\\missing"
	cmd := shellCommand(command)
	want := []string{"C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
		"-NoProfile", "-NonInteractive", "-EncodedCommand",
		encodePowerShellCommand(powerShellReRunScript(command))}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("shellCommand() args = %q, want %q", cmd.Args, want)
	}
}

func TestShellCommandFallsBackWhenPowerShellIsMissing(t *testing.T) {
	clearShellPathEnv(t)
	stubPowerShellAncestor(t, "pwsh.exe")
	stubPowerShellPath(t, "")
	t.Setenv("COMSPEC", "")
	cmd := shellCommand("echo hello")
	want := []string{"cmd.exe", "/c", "echo hello"}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("shellCommand() args = %q, want %q", cmd.Args, want)
	}
}

func TestShellCommandFallsBackToCmd(t *testing.T) {
	clearShellPathEnv(t)
	stubPowerShellAncestor(t, "")
	t.Setenv("COMSPEC", "")
	cmd := shellCommand("echo hello")
	want := []string{"cmd.exe", "/c", "echo hello"}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("shellCommand() args = %q, want %q", cmd.Args, want)
	}
}

func TestShellCommandUsesComspec(t *testing.T) {
	clearShellPathEnv(t)
	stubPowerShellAncestor(t, "")
	t.Setenv("COMSPEC", "C:\\Windows\\System32\\cmd.exe")
	cmd := shellCommand("echo hello")
	if got := cmd.Args[0]; got != "C:\\Windows\\System32\\cmd.exe" {
		t.Fatalf("shellCommand() executable = %q, want COMSPEC", got)
	}
}

// writePowerShellStub writes an empty file that looks like a PowerShell
// executable.
func writePowerShellStub(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

func TestShellCommandUsesPowerShellFromEnv(t *testing.T) {
	stubPowerShellAncestor(t, "")
	path := writePowerShellStub(t, "pwsh.exe")
	t.Setenv(consts.ShellPathEnv, path)
	command := "cd C:\\missing"
	cmd := shellCommand(command)
	want := []string{path, "-NoProfile", "-NonInteractive", "-EncodedCommand",
		encodePowerShellCommand(powerShellReRunScript(command))}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("shellCommand() args = %q, want %q", cmd.Args, want)
	}
}

func TestShellCommandIgnoresNonPowerShellEnv(t *testing.T) {
	stubPowerShellAncestor(t, "")
	t.Setenv(consts.ShellPathEnv, writePowerShellStub(t, "cmd.exe"))
	t.Setenv("COMSPEC", "")
	cmd := shellCommand("echo hello")
	want := []string{"cmd.exe", "/c", "echo hello"}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("shellCommand() args = %q, want %q", cmd.Args, want)
	}
}

func TestShellCommandIgnoresMissingPowerShellEnv(t *testing.T) {
	stubPowerShellAncestor(t, "")
	t.Setenv(consts.ShellPathEnv, filepath.Join(t.TempDir(), "pwsh.exe"))
	t.Setenv("COMSPEC", "")
	cmd := shellCommand("echo hello")
	want := []string{"cmd.exe", "/c", "echo hello"}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("shellCommand() args = %q, want %q", cmd.Args, want)
	}
}
func TestPowerShellReRunScript(t *testing.T) {
	got := powerShellReRunScript("cd C:\\missing")
	want := "$ProgressPreference = 'SilentlyContinue'; " +
		"[System.Threading.Thread]::CurrentThread.CurrentUICulture = [System.Globalization.CultureInfo]::GetCultureInfo('en-US'); " +
		"try { & { cd C:\\missing }" +
		" *>&1 | ForEach-Object { if ($_ -is [System.Management.Automation.ErrorRecord])" +
		" { $_.Exception.Message } else { $_ } } } catch { $_.Exception.Message }"
	if got != want {
		t.Fatalf("powerShellReRunScript() = %q, want %q", got, want)
	}
}

func TestPowerShellReRunScriptNormalizesDiagnosticLanguage(t *testing.T) {
	for _, executable := range []string{"powershell.exe", "pwsh.exe"} {
		t.Run(executable, func(t *testing.T) {
			path := powerShellPath(executable)
			if path == "" {
				t.Skipf("%s is not installed", executable)
			}
			directory := strings.ReplaceAll(t.TempDir(), "'", "''")
			command := "[System.Globalization.CultureInfo]::CurrentUICulture.Name; " +
				"New-Item -ItemType Directory -Path '" + directory + "' -ErrorAction Stop"
			script := "[System.Threading.Thread]::CurrentThread.CurrentUICulture = " +
				"[System.Globalization.CultureInfo]::GetCultureInfo('zh-CN'); " + powerShellReRunScript(command)
			output, err := exec.Command(path, "-NoProfile", "-NonInteractive", "-EncodedCommand", encodePowerShellCommand(script)).CombinedOutput()
			if err != nil {
				t.Fatalf("PowerShell rerun failed: %v\n%s", err, output)
			}
			if text := string(output); !strings.Contains(text, "en-US") ||
				!strings.Contains(text, "An item with the specified name") || !strings.Contains(text, "already exists.") {
				t.Fatalf("PowerShell diagnostics were not normalized: %q", output)
			}
		})
	}
}
