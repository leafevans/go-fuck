package entrypoints

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/types"
)

// captureOutput replaces standard output and standard error while fn runs.
func captureOutput(t *testing.T, fn func() int) (int, string, string) {
	t.Helper()
	oldStdout, oldStderr := os.Stdout, os.Stderr

	outReader, outWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	errReader, errWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout, os.Stderr = outWriter, errWriter

	// The corrected command is printed through an injectable writer that is
	// initialized from os.Stdout at package load time.
	previousStdout := types.Stdout()
	types.SetStdout(outWriter)

	outChannel := make(chan string, 1)
	errChannel := make(chan string, 1)
	go func() {
		data, _ := io.ReadAll(outReader)
		outChannel <- string(data)
	}()
	go func() {
		data, _ := io.ReadAll(errReader)
		errChannel <- string(data)
	}()

	code := fn()

	_ = outWriter.Close()
	_ = errWriter.Close()
	os.Stdout, os.Stderr = oldStdout, oldStderr
	types.SetStdout(previousStdout)
	return code, <-outChannel, <-errChannel
}

// runMain runs Main with the given arguments, restoring them afterwards.
func runMain(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	previousArgs := os.Args
	os.Args = args
	t.Cleanup(func() {
		os.Args = previousArgs
	})
	return captureOutput(t, Main)
}

// resetEntrypointsState isolates the settings and the shell of the tests.
func resetEntrypointsState(t *testing.T) string {
	t.Helper()
	configDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("HOME", configDir)
	t.Setenv("USERPROFILE", configDir)
	t.Setenv("TF_HISTORY", "")
	t.Setenv("TF_SHELL", "bash")
	conf.Reset()
	logs.SetOutput(io.Discard)
	t.Cleanup(func() {
		conf.Reset()
		logs.SetOutput(io.Discard)
	})
	return configDir
}

func TestMainVersion(t *testing.T) {
	resetEntrypointsState(t)
	buffer := &bytes.Buffer{}
	logs.SetOutput(buffer)

	code, _, _ := runMain(t, "go-fuck", "--version")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	output := buffer.String()
	if !strings.Contains(output, "go-fuck") || !strings.Contains(output, "using Go") {
		t.Fatalf("version output = %q", output)
	}
	if !strings.Contains(output, "Bash") {
		t.Fatalf("version output does not contain the detected shell: %q", output)
	}
}

func TestMainHelp(t *testing.T) {
	resetEntrypointsState(t)
	code, _, stderr := runMain(t, "go-fuck", "--help")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	if !strings.Contains(stderr, "usage: go-fuck") {
		t.Fatalf("help output = %q", stderr)
	}
	if !strings.Contains(stderr, "--alias") {
		t.Fatalf("help output = %q", stderr)
	}
}

func TestMainAlias(t *testing.T) {
	resetEntrypointsState(t)
	code, stdout, _ := runMain(t, "go-fuck", "--alias")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	if !strings.Contains(stdout, "TF_ALIAS") || !strings.Contains(stdout, "go-fuck") {
		t.Fatalf("alias output = %q", stdout)
	}
}

func TestMainAliasCustomName(t *testing.T) {
	resetEntrypointsState(t)
	code, stdout, _ := runMain(t, "go-fuck", "--alias", "please")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	if !strings.Contains(stdout, "please") {
		t.Fatalf("alias output = %q", stdout)
	}
}

func TestMainUsage(t *testing.T) {
	resetEntrypointsState(t)
	code, _, stderr := runMain(t, "go-fuck")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	if !strings.Contains(stderr, "usage: go-fuck") {
		t.Fatalf("usage output = %q", stderr)
	}
}

func TestMainShellLoggerNotSupported(t *testing.T) {
	resetEntrypointsState(t)
	buffer := &bytes.Buffer{}
	logs.SetOutput(buffer)

	code, _, stderr := runMain(t, "go-fuck", "-l", "log")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	if !strings.Contains(buffer.String()+stderr, "Shell logger is not supported") {
		t.Fatalf("unexpected output: %q %q", buffer.String(), stderr)
	}
}

func TestMainUnknownFlag(t *testing.T) {
	resetEntrypointsState(t)
	code, _, stderr := runMain(t, "go-fuck", "--unknown")
	if code != 2 {
		t.Fatalf("Main() = %d, want 2", code)
	}
	if !strings.Contains(stderr, "usage: go-fuck") {
		t.Fatalf("usage output = %q", stderr)
	}
}
func TestMainForceCommand(t *testing.T) {
	resetEntrypointsState(t)
	// Main re-reads the settings, so a generous wait time has to be provided
	// through the environment instead of the current settings.
	t.Setenv("GOFUCK_WAIT_COMMAND", "60")
	code, stdout, _ := runMain(t, "go-fuck", "-y", "--force-command", "echo 'hello"+string(rune(34))+"")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	if !strings.Contains(stdout, "echo "+string(rune(34))+"hello"+string(rune(34))) {
		t.Fatalf("stdout = %q, want the corrected command", stdout)
	}
}
