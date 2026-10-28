//go:build linux || darwin

package instant

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/leafevans/go-fuck/internal/consts"
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
