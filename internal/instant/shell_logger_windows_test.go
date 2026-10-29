//go:build windows

package instant

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/leafevans/go-fuck/internal/consts"
)

// findPowerShell returns the PowerShell executable used by the tests.
func findPowerShell(t *testing.T) string {
	t.Helper()

	if path, err := exec.LookPath("powershell.exe"); err == nil {
		return path
	}
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		t.Skip("SystemRoot is not set")
	}
	path := filepath.Join(systemRoot, "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	if _, err := os.Stat(path); err != nil {
		t.Skip("PowerShell is not available")
	}
	return path
}

func TestConPTYIsAvailable(t *testing.T) {
	if !Supported() {
		t.Fatal("Supported() = false, want true on Windows 10 1809 or newer")
	}
}

func TestIsPowerShell(t *testing.T) {
	cases := map[string]bool{
		"C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe": true,
		"C:/Program Files/PowerShell/7/pwsh.exe":                    true,
		"C:/Windows/System32/cmd.exe":                               false,
		"powershell":                                                false,
		"":                                                          false,
	}
	for path, want := range cases {
		if got := isPowerShell(path); got != want {
			t.Fatalf("isPowerShell(%q) = %t, want %t", path, got, want)
		}
	}
}

func TestQuoteWindowsCommandLineArg(t *testing.T) {
	const simple = "C:/Windows/powershell.exe"
	if got := quoteWindowsCommandLineArg(simple); got != simple {
		t.Fatalf("quoteWindowsCommandLineArg() = %q, want %q", got, simple)
	}

	const spaced = "C:/Program Files/PowerShell/7/pwsh.exe"
	want := string('"') + spaced + string('"')
	if got := quoteWindowsCommandLineArg(spaced); got != want {
		t.Fatalf("quoteWindowsCommandLineArg() = %q, want %q", got, want)
	}
}

func TestShellCommandLineUsesTheRequestedShell(t *testing.T) {
	powershell := findPowerShell(t)
	t.Setenv(consts.ShellPathEnv, powershell)

	commandLine, err := shellCommandLine()
	if err != nil {
		t.Fatalf("shellCommandLine() error = %v", err)
	}
	if !strings.HasPrefix(commandLine, quoteWindowsCommandLineArg(powershell)) || !strings.HasSuffix(commandLine, " -NoLogo") {
		t.Fatalf("shellCommandLine() = %q", commandLine)
	}
}

// TestRunLoggedShell starts an interactive PowerShell on a pseudo console,
// feeds commands through it and reads their output back from the terminal log.
func TestRunLoggedShell(t *testing.T) {
	powershell := findPowerShell(t)

	dir := t.TempDir()
	profile := filepath.Join(dir, "instant-profile.ps1")
	content := "function prompt { $mark = ([string]([char]0x200B)) * " +
		strconv.Itoa(len([]rune(consts.UserCommandMark))) +
		"; return $mark + 'PS> ' }" + string(rune(13)) + string(rune(10))
	if err := os.WriteFile(profile, []byte(content), 0o600); err != nil {
		t.Fatalf("write the profile: %v", err)
	}

	logPath := filepath.Join(dir, "instant.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open the log: %v", err)
	}
	defer logFile.Close()

	commandLine := quoteWindowsCommandLineArg(powershell) + " -NoLogo -NoProfile"
	enter := string(rune(13))
	input := strings.NewReader(
		". '" + profile + "'" + enter +
			"echo instant-output" + enter +
			"exit" + enter,
	)

	type result struct {
		code int
		err  error
	}
	finished := make(chan result, 1)
	go func() {
		code, err := runLoggedShell(commandLine, input, io.Discard, NewRingWriter(logFile))
		finished <- result{code: code, err: err}
	}()

	select {
	case outcome := <-finished:
		if outcome.err != nil {
			t.Fatalf("runLoggedShell() error = %v", outcome.err)
		}
		if outcome.code != 0 {
			t.Fatalf("runLoggedShell() exit code = %d, want 0", outcome.code)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("runLoggedShell() did not finish in time")
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read the log: %v", err)
	}
	if !strings.HasPrefix(string(data), conPTYSizeRecord) {
		t.Fatalf("terminal log does not start with the size record: %q", string(data))
	}
	if !strings.Contains(string(data), "instant-output") {
		t.Fatalf("terminal log does not contain the command output: %q", string(data))
	}

	output, ok := LastOutput(logPath, "echo instant-output")
	if !ok || !strings.Contains(output, "instant-output") {
		t.Logf("log: %q", string(data))
		t.Fatalf("LastOutput() = %q, %t, want instant-output, true", output, ok)
	}
}
