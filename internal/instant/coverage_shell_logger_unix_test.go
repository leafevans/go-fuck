//go:build linux || darwin

package instant

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// failingWriter is an io.Writer that fails every write.
type failingWriter struct{}

// Write always returns an error.
func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

// lookupTool returns the absolute path of a standard tool or skips the test.
func lookupTool(t *testing.T, name string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Skipf("%s is not available: %v", name, err)
	}
	return path
}

// TestSupportedOnUnix covers the supported platform answer.
func TestSupportedOnUnix(t *testing.T) {
	if !Supported() {
		t.Fatal("Supported() = false on a platform with pseudo terminals")
	}
}

// TestRunShellLoggerWithoutShell covers the missing SHELL error of
// RunShellLogger.
func TestRunShellLoggerWithoutShell(t *testing.T) {
	t.Setenv("SHELL", "")
	code, err := RunShellLogger(filepath.Join(t.TempDir(), "instant.log"))
	if err == nil {
		t.Fatal("RunShellLogger() error = nil, want the missing SHELL error")
	}
	if code != 1 {
		t.Fatalf("RunShellLogger() = %d, want 1", code)
	}
}

// TestRunShellLoggerLogOpenFailure covers the log file error of
// RunShellLogger.
func TestRunShellLoggerLogOpenFailure(t *testing.T) {
	t.Setenv("SHELL", lookupTool(t, "true"))
	code, err := RunShellLogger(t.TempDir())
	if err == nil {
		t.Fatal("RunShellLogger() error = nil, want the log open error")
	}
	if code != 1 {
		t.Fatalf("RunShellLogger() = %d, want 1", code)
	}
}

// TestRunShellLoggerRunsShell covers the successful RunShellLogger path with a
// shell that exits immediately.
func TestRunShellLoggerRunsShell(t *testing.T) {
	t.Setenv("SHELL", lookupTool(t, "true"))
	logPath := filepath.Join(t.TempDir(), "instant.log")
	code, err := RunShellLogger(logPath)
	if err != nil {
		t.Fatalf("RunShellLogger() error = %v", err)
	}
	if code != 0 {
		t.Fatalf("RunShellLogger() = %d, want 0", code)
	}
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("the terminal log was not created: %v", err)
	}
}

// TestRunLoggedShellStartFailure covers the command start error of
// runLoggedShell.
func TestRunLoggedShellStartFailure(t *testing.T) {
	code, err := runLoggedShell(filepath.Join(t.TempDir(), "missing-shell"),
		strings.NewReader(""), io.Discard, io.Discard)
	if err == nil {
		t.Fatal("runLoggedShell() error = nil, want the start error")
	}
	if code != 1 {
		t.Fatalf("runLoggedShell() = %d, want 1", code)
	}
}

// TestRunLoggedShellWriteFailures covers the output and log write failures of
// runLoggedShell.
func TestRunLoggedShellWriteFailures(t *testing.T) {
	echo := lookupTool(t, "echo")

	if _, err := runLoggedShell(echo, strings.NewReader(""), failingWriter{}, io.Discard); err != nil {
		t.Fatalf("runLoggedShell(stdout failure) error = %v", err)
	}
	if _, err := runLoggedShell(echo, strings.NewReader(""), io.Discard, failingWriter{}); err != nil {
		t.Fatalf("runLoggedShell(log failure) error = %v", err)
	}
}

// TestRunLoggedShellSignalExit covers the negative exit code handling for a
// shell that is killed by a signal.
func TestRunLoggedShellSignalExit(t *testing.T) {
	script := filepath.Join(t.TempDir(), "shell")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nkill -9 $$\n"), 0o755); err != nil {
		t.Fatalf("write the shell script: %v", err)
	}
	code, err := runLoggedShell(script, strings.NewReader(""), io.Discard, io.Discard)
	if err != nil {
		t.Fatalf("runLoggedShell() error = %v", err)
	}
	if code != 1 {
		t.Fatalf("runLoggedShell() = %d, want 1 for a signalled shell", code)
	}
}

// TestMakeRawClearsTerminalAttributes covers makeRaw.
func TestMakeRawClearsTerminalAttributes(t *testing.T) {
	const (
		inputFlags = syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP |
			syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
		outputFlags = syscall.OPOST
		localFlags  = syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
		controlFlag = syscall.CSIZE | syscall.PARENB
	)
	state := syscall.Termios{
		Iflag: inputFlags,
		Oflag: outputFlags,
		Lflag: localFlags,
		Cflag: controlFlag,
	}
	raw := makeRaw(&state)
	if raw.Iflag&inputFlags != 0 {
		t.Fatalf("makeRaw() Iflag = %#x, want the input flags cleared", raw.Iflag)
	}
	if raw.Oflag&outputFlags != 0 {
		t.Fatalf("makeRaw() Oflag = %#x, want the output flags cleared", raw.Oflag)
	}
	if raw.Lflag&localFlags != 0 {
		t.Fatalf("makeRaw() Lflag = %#x, want the local flags cleared", raw.Lflag)
	}
	if raw.Cflag&syscall.PARENB != 0 {
		t.Fatalf("makeRaw() Cflag = %#x, want the parity flag cleared", raw.Cflag)
	}
	if raw.Cflag&syscall.CSIZE != syscall.CS8 {
		t.Fatalf("makeRaw() Cflag = %#x, want the 8 bit character size", raw.Cflag)
	}
	if raw.Cc[syscall.VMIN] != 1 || raw.Cc[syscall.VTIME] != 0 {
		t.Fatalf("makeRaw() Cc = %v, want VMIN 1 and VTIME 0", raw.Cc)
	}
	if state.Iflag&inputFlags != inputFlags {
		t.Fatal("makeRaw() modified the input state")
	}
}

// TestMakeRawTerminalNonTerminal covers the non terminal branches of
// makeRawTerminal.
func TestMakeRawTerminalNonTerminal(t *testing.T) {
	file, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	defer file.Close()

	if restore := makeRawTerminal(file); restore != nil {
		t.Fatal("makeRawTerminal() = non-nil for a non terminal file")
	}
	if restore := makeRawTerminal(strings.NewReader("")); restore != nil {
		t.Fatal("makeRawTerminal() = non-nil for a reader that is not a file")
	}
}
