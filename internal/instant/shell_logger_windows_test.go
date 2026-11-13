//go:build windows

package instant

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/leafevans/go-fuck/v2/internal/consts"
)

// testPowerShell returns the path of Windows PowerShell, or skips the test.
func testPowerShell(t *testing.T) string {
	t.Helper()
	if path, err := exec.LookPath("powershell.exe"); err == nil {
		return path
	}
	if root := os.Getenv("SystemRoot"); root != "" {
		candidate := filepath.Join(root, "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	t.Skip("powershell.exe is not available")
	return ""
}

// TestRunLoggedShell starts an interactive Windows PowerShell under a pseudo
// console, feeds a command through it and reads the output back from the
// terminal log.
func TestRunLoggedShell(t *testing.T) {
	powershell := testPowerShell(t)

	dir := t.TempDir()
	logPath := filepath.Join(dir, "instant.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	defer logFile.Close()

	prompt := "function prompt { ([string]([char]0x200B)) * 10 + 'PS> ' }"
	commandLine := quoteWindowsCommandLineArg(powershell) + " -NoLogo -NoProfile -NoExit -Command " +
		quoteWindowsCommandLineArg(prompt)
	input := strings.NewReader("Write-Output instant-output\r\nexit\r\n")

	type result struct {
		code int
		err  error
	}
	finished := make(chan result, 1)
	go func() {
		code, err := runLoggedShell(commandLine, input, &bytes.Buffer{}, NewRingWriter(logFile))
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
	case <-time.After(60 * time.Second):
		t.Fatal("runLoggedShell() did not finish in time")
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(data), "instant-output") {
		t.Fatalf("terminal log does not contain the command output: %q", string(data))
	}
	if !strings.Contains(string(data), consts.UserCommandMark) {
		t.Fatalf("terminal log does not contain the prompt marker")
	}

	output, ok := LastOutput(logPath, "Write-Output instant-output")
	if !ok || !strings.Contains(output, "instant-output") {
		t.Fatalf("LastOutput() = %q, %t, want instant-output, true", output, ok)
	}
}
