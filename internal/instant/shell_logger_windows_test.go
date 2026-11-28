//go:build windows

package instant

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestTerminalStdinDetection checks that only consoles are treated as an
// interactive terminal input.
func TestTerminalStdinDetection(t *testing.T) {
	if terminalStdin(strings.NewReader("")) {
		t.Fatal("strings.Reader was detected as a terminal")
	}
	input, output := io.Pipe()
	defer output.Close()
	defer input.Close()
	if terminalStdin(input) {
		t.Fatal("io.PipeReader was detected as a terminal")
	}
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer devNull.Close()
	if terminalStdin(devNull) {
		t.Fatal("os.DevNull was detected as a terminal")
	}
}

// TestRunShellLoggerRejectsPipedInput checks that a non-interactive host does
// not start the logged shell, and that the failure happens before the log file
// exists so the profile falls back to the regular alias.
func TestRunShellLoggerRejectsPipedInput(t *testing.T) {
	if !Supported() {
		t.Skip("ConPTY is not available")
	}
	logPath := filepath.Join(t.TempDir(), "instant.log")
	if _, err := runShellLogger(logPath, strings.NewReader(""), io.Discard); err == nil {
		t.Fatal("runShellLogger() error = nil, want the non-terminal error")
	}
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Fatalf("log file exists after a refused start: %v", err)
	}
}

// TestLoggedShellPreservesExitAndOutput checks ordinary exit with stdin still
// waiting for input. The final output must reach both writers before return.
func TestLoggedShellPreservesExitAndOutput(t *testing.T) {
	if !Supported() {
		t.Skip("ConPTY unavailable")
	}
	shell, err := exec.LookPath("powershell.exe")
	if err != nil {
		t.Skip(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	input, keys := io.Pipe()
	defer input.Close()
	defer keys.Close()
	logPath := filepath.Join(t.TempDir(), "instant.log")
	log, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	var output bytes.Buffer
	command := quoteWindowsCommandLineArg(shell) + " -NoLogo -NoProfile -Command " + syscall.EscapeArg("Write-Output 'final-frame'; exit 42")
	if code, err := runLoggedShellContext(ctx, command, input, &output, NewRingWriter(log), true); err != nil || code != 42 {
		t.Fatalf("logged shell exit = %d, %v, want 42 and no error", code, err)
	}
	if err := log.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "final-frame") || !bytes.Contains(data, []byte("final-frame")) {
		t.Fatalf("output missing at exit: stdout %q, log %q", output.String(), data)
	}
	if err := os.Remove(logPath); err != nil {
		t.Fatalf("remove log after session exit: %v", err)
	}
}

// TestRunLoggedShellStopsWhenTerminalCloses checks that a closing terminal ends
// the logged shell instead of leaving the logger waiting forever.
func TestRunLoggedShellStopsWhenTerminalCloses(t *testing.T) {
	if !Supported() {
		t.Skip("ConPTY is not available")
	}
	shell, err := exec.LookPath("pwsh.exe")
	if err != nil {
		shell, err = exec.LookPath("powershell.exe")
	}
	if err != nil {
		t.Skip("PowerShell is not available")
	}

	logPath := filepath.Join(t.TempDir(), "instant.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	defer logFile.Close()

	input, closeInput := io.Pipe()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	finished := make(chan shellExit, 1)
	go func() {
		command := quoteWindowsCommandLineArg(shell) + " -NoLogo -NoProfile -Command " + syscall.EscapeArg("Write-Output 'session-ready'; Start-Sleep -Seconds 45")
		code, err := runLoggedShellContext(ctx, command, input, io.Discard, NewRingWriter(logFile), true)
		finished <- shellExit{code, err}
		close(finished)
	}()

	defer func() {
		cancel()
		_ = closeInput.Close()
		_ = input.Close()
		select {
		case <-finished:
		case <-time.After(10 * time.Second):
			t.Error("terminal-close test cleanup did not complete")
		}
	}()
	// The size header alone does not prove PowerShell has started. Wait for
	// actual child output before modelling the terminal input closing.
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		if data, err := os.ReadFile(logPath); err == nil && strings.Contains(string(data), "session-ready") {
			break
		}
		select {
		case result := <-finished:
			t.Fatalf("logged shell exited before readiness: %d, %v", result.code, result.err)
		case <-ctx.Done():
			t.Fatal("logged shell did not start")
		case <-ticker.C:
		}
	}
	if err := closeInput.Close(); err != nil {
		t.Fatal(err)
	}

	select {
	case result := <-finished:
		if result.err != nil || result.code == 0 {
			t.Fatalf("terminal close exit = %d, %v, want interrupted exit and no error", result.code, result.err)
		}
	case <-ctx.Done():
		t.Fatal("logged shell did not stop after the terminal input closed")
	}
}

var errSessionSink = errors.New("test sink failed")

type failingSessionWriter struct{ calls int }

func (w *failingSessionWriter) Write(p []byte) (int, error) {
	w.calls++
	// Let the log size record through, then fail on actual child output.
	if w.calls > 1 {
		return 0, errSessionSink
	}
	return len(p), nil
}

func TestLoggedShellWriterFailureDrainsOutput(t *testing.T) {
	if !Supported() {
		t.Skip("ConPTY unavailable")
	}
	shell, err := exec.LookPath("powershell.exe")
	if err != nil {
		t.Skip(err)
	}
	for _, which := range []string{"stdout", "log"} {
		t.Run(which, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			sink := &failingSessionWriter{}
			var out, log io.Writer = io.Discard, io.Discard
			if which == "stdout" {
				out = sink
			} else {
				log = sink
			}
			command := quoteWindowsCommandLineArg(shell) + " -NoLogo -NoProfile -Command " + syscall.EscapeArg("for ($i = 0; $i -lt 128; $i++) { Write-Output ('x' * 4096) }; Start-Sleep -Seconds 25")
			_, err := runLoggedShellContext(ctx, command, strings.NewReader(""), out, log, false)
			if ctx.Err() != nil {
				t.Fatalf("session did not stop on writer failure: %v", ctx.Err())
			}
			if !errors.Is(err, errSessionSink) {
				t.Fatalf("error = %v, want sink failure", err)
			}
		})
	}
}

type initialRecordFailure struct{}

func (initialRecordFailure) Write([]byte) (int, error) { return 0, errSessionSink }
func TestLoggedShellInitialRecordFailure(t *testing.T) {
	// Failure must precede CreateProcess: even an invalid command must report
	// the sink error rather than a child-start error.
	_, err := runLoggedShellContext(t.Context(), "missing-test-executable", strings.NewReader(""), io.Discard, initialRecordFailure{}, false)
	if !errors.Is(err, errSessionSink) {
		t.Fatalf("error = %v", err)
	}
}
