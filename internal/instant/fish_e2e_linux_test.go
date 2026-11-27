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
	logPath := loggedFish(t, "echo fish-instant-output\nexit\n")
	assertFishOutput(t, logPath, "echo fish-instant-output", "fish-instant-output")
}

func TestRunLoggedFishLatestOutput(t *testing.T) {
	logPath := loggedFish(t, "set sample first-value\necho $sample\nset sample second-value\necho $sample\nexit\n")
	assertFishOutput(t, logPath, "echo $sample", "second-value")
}

func TestRunLoggedFishCancelledInput(t *testing.T) {
	input, send := io.Pipe()
	defer input.Close()
	defer send.Close()
	// Wait for the editor to start, then send Ctrl-C without submitting the
	// first line. Do not wait for a particular repaint or a fresh prompt:
	// newer fish versions can clear the input in place.
	var terminal fishPromptInput
	terminal.send = send
	logPath := runFish(t, input, &terminal)
	assertFishOutput(t, logPath, "echo after-interrupt", "after-interrupt")
	if got, ok := LastOutput(logPath, "echo cancelled-input"); ok {
		t.Fatalf("cancelled input matched with output %q", got)
	}
}

type fishPromptInput struct {
	send    *io.PipeWriter
	seen    string
	started bool
}

func (p *fishPromptInput) Write(data []byte) (int, error) {
	if !p.started {
		p.seen += string(data)
		if strings.Contains(p.seen, "fish> ") {
			p.started = true
			if _, err := io.WriteString(p.send, "echo cancelled-input\x03echo after-interrupt\nexit\n"); err != nil {
				return 0, err
			}
		}
	}
	return len(data), nil
}

func assertFishOutput(t *testing.T, logPath, command, want string) {
	t.Helper()
	got, ok := LastOutput(logPath, command)
	if !ok || got != want {
		t.Fatalf("LastOutput(%q) = %q, %t; want %q, true", command, got, ok, want)
	}
}

func loggedFish(t *testing.T, commands string) string {
	t.Helper()
	return runFish(t, strings.NewReader(commands), io.Discard)
}

func runFish(t *testing.T, input io.Reader, output io.Writer) string {
	t.Helper()
	fish, err := exec.LookPath("fish")
	if err != nil {
		t.Skipf("fish is not available: %v", err)
	}

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "fish"), 0o755); err != nil {
		t.Fatalf("create the fish config directory: %v", err)
	}
	config := "function fish_prompt; printf 'fish> '; end\n" +
		"function __gofuck_mark_prompt --on-event fish_prompt\n" +
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
	t.Setenv("XDG_DATA_HOME", dir)
	t.Setenv("HOME", dir)
	t.Setenv("TERM", "xterm")

	type result struct {
		code int
		err  error
	}
	finished := make(chan result, 1)
	go func() {
		code, err := runLoggedShell(wrapper, input, output, NewRingWriter(logFile))
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
		data, _ := os.ReadFile(logPath)
		t.Fatalf("runLoggedShell() did not finish in time; log: %q", data)
	}

	return logPath
}
