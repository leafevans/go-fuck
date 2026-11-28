//go:build windows

package instant

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	defer input.Close()
	finished := make(chan int, 1)
	go func() {
		code, _ := runLoggedShellFromTerminal(quoteWindowsCommandLineArg(shell)+" -NoLogo -NoProfile", input, io.Discard, NewRingWriter(logFile), true)
		finished <- code
		close(finished)
	}()

	// The shell reports the console close as a control event, so only the
	// prompt return matters, not the exit code.
	// Wait until the shell printed its banner, then close the terminal input.
	deadline := time.Now().Add(15 * time.Second)
	for {
		if info, err := os.Stat(logPath); err == nil && info.Size() > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("logged shell did not start")
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err := closeInput.Close(); err != nil {
		t.Fatal(err)
	}

	select {
	case <-finished:
	case <-time.After(15 * time.Second):
		t.Fatal("runLoggedShellFromTerminal() did not stop after the terminal closed")
	}
}
