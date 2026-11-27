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

	"github.com/leafevans/go-fuck/v4/internal/consts"
)

// TestRunLoggedFishShell starts an interactive fish on a pseudo terminal with
// the instant mode prompt marker, feeds a command through it and reads the
// output back from the terminal log.
func TestRunLoggedFishShell(t *testing.T) {
	fish, err := exec.LookPath("fish")
	if err != nil {
		t.Skipf("fish is not available: %v", err)
	}

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "fish"), 0o755); err != nil {
		t.Fatalf("create the fish config directory: %v", err)
	}
	config := "function __gofuck_mark_prompt --on-event fish_prompt\n" +
		"    printf '%s\\r' '" + consts.UserCommandMark + "'\n" +
		"end\n"
	if err := os.WriteFile(filepath.Join(dir, "fish", "config.fish"), []byte(config), 0o644); err != nil {
		t.Fatalf("write the fish config: %v", err)
	}

	wrapper := filepath.Join(dir, "shell")
	script := "#!/bin/sh\nexec " + fish + " -i\n"
	if err := os.WriteFile(wrapper, []byte(script), 0o755); err != nil {
		t.Fatalf("write shell wrapper: %v", err)
	}

	logPath := filepath.Join(dir, "instant.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	defer logFile.Close()

	// fish reads its configuration from XDG_CONFIG_HOME/fish/config.fish.
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)
	t.Setenv("TERM", "dumb")

	input := strings.NewReader("echo fish-instant-output\nexit\n")
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
	if !strings.Contains(string(data), "fish-instant-output") {
		t.Fatalf("terminal log does not contain the command output: %q", string(data))
	}

	output, ok := LastOutput(logPath, "echo fish-instant-output")
	if !ok || !strings.Contains(output, "fish-instant-output") {
		t.Fatalf("LastOutput() = %q, %t, want fish-instant-output, true", output, ok)
	}
}
