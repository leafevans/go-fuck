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
	"time"

	"github.com/leafevans/go-fuck/v2/internal/consts"
)

// TestRunLoggedShell starts an interactive bash on a pseudo terminal, feeds a
// command through it and reads the output back from the terminal log.
func TestRunLoggedShell(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skipf("bash is not available: %v", err)
	}

	dir := t.TempDir()
	wrapper := filepath.Join(dir, "shell")
	script := "#!/bin/sh\nexec " + bash + " --noprofile --norc -i\n"
	if err := os.WriteFile(wrapper, []byte(script), 0o755); err != nil {
		t.Fatalf("write shell wrapper: %v", err)
	}

	logPath := filepath.Join(dir, "instant.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	defer logFile.Close()

	t.Setenv("PS1", consts.UserCommandMark+"$ ")
	t.Setenv("TERM", "dumb")
	t.Setenv("HISTFILE", filepath.Join(dir, "history"))

	input := strings.NewReader("echo instant-output\nexit\n")
	type result struct {
		code int
		err  error
	}
	finished := make(chan result, 1)
	go func() {
		code, err := runLoggedShell(wrapper, input, io.Discard, NewRingWriter(logFile))
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
	case <-time.After(20 * time.Second):
		t.Fatal("runLoggedShell() did not finish in time")
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(data), "instant-output") {
		t.Fatalf("terminal log does not contain the command output: %q", string(data))
	}

	output, ok := LastOutput(logPath, "echo instant-output")
	if !ok || !strings.Contains(output, "instant-output") {
		t.Fatalf("LastOutput() = %q, %t, want instant-output, true", output, ok)
	}
}

// TestRunLoggedShellMultilineCommand checks that a command typed over several
// lines is still found in the terminal log.
func TestRunLoggedShellMultilineCommand(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skipf("bash is not available: %v", err)
	}

	dir := t.TempDir()
	wrapper := filepath.Join(dir, "shell")
	script := "#!/bin/sh\nexec " + bash + " --noprofile --norc -i\n"
	if err := os.WriteFile(wrapper, []byte(script), 0o755); err != nil {
		t.Fatalf("write shell wrapper: %v", err)
	}

	logPath := filepath.Join(dir, "instant.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	defer logFile.Close()

	t.Setenv("PS1", consts.UserCommandMark+"$ ")
	t.Setenv("TERM", "dumb")
	t.Setenv("HISTFILE", filepath.Join(dir, "history"))

	input := strings.NewReader("if true; then\necho instant-multiline\nfi\nexit\n")
	finished := make(chan error, 1)
	go func() {
		_, err := runLoggedShell(wrapper, input, io.Discard, NewRingWriter(logFile))
		finished <- err
	}()

	select {
	case err := <-finished:
		if err != nil {
			t.Fatalf("runLoggedShell() error = %v", err)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("runLoggedShell() did not finish in time")
	}

	output, ok := LastOutput(logPath, "if true; then\necho instant-multiline\nfi")
	if !ok || !strings.Contains(output, "instant-multiline") {
		t.Fatalf("LastOutput() = %q, %t, want the multi-line command output", output, ok)
	}
}

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
