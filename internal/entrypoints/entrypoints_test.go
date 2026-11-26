package entrypoints

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v4/internal/conf"
	"github.com/leafevans/go-fuck/v4/internal/consts"
	"github.com/leafevans/go-fuck/v4/internal/instant"
	"github.com/leafevans/go-fuck/v4/internal/logs"
	"github.com/leafevans/go-fuck/v4/internal/types"
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
	consts.SetExecutableName(consts.AppName)
	logs.SetOutput(io.Discard)
	t.Cleanup(func() {
		conf.Reset()
		consts.SetExecutableName(consts.AppName)
		logs.SetOutput(io.Discard)
	})
	return configDir
}

func TestMainVersion(t *testing.T) {
	resetEntrypointsState(t)

	code, stdout, stderr := runMain(t, "go-fuck", "--version")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	if stderr != "" {
		t.Fatalf("version stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "go-fuck") || !strings.Contains(stdout, "using Go") {
		t.Fatalf("version output = %q", stdout)
	}
	if !strings.Contains(stdout, "Bash") {
		t.Fatalf("version output does not contain the detected shell: %q", stdout)
	}
}

func TestMainHelp(t *testing.T) {
	resetEntrypointsState(t)
	code, stdout, stderr := runMain(t, "go-fuck", "--help")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	if stderr != "" {
		t.Fatalf("help stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "go-fuck help") || !strings.Contains(stdout, "Usage") {
		t.Fatalf("help output = %q", stdout)
	}
	if !strings.Contains(stdout, "--alias") {
		t.Fatalf("help output = %q", stdout)
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
	if !strings.Contains(stdout, "GOFUCK_INSTANT_MODE=True") {
		t.Fatalf("default auto alias did not initialize capture: %q", stdout)
	}
	if strings.Contains(stdout, "\x1b[") || strings.Contains(stdout, "go-fuck alias") {
		t.Fatalf("alias output is not raw shell code: %q", stdout)
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
	if !strings.Contains(stdout, "TF_ALIAS=please") {
		t.Fatalf("custom alias name was not passed into the logged shell: %q", stdout)
	}
}

func TestMainAliasCaptureModeRerun(t *testing.T) {
	resetEntrypointsState(t)
	code, stdout, stderr := runMain(t, "go-fuck", "--alias", "--capture-mode", "rerun")
	if code != 0 || stderr != "" {
		t.Fatalf("Main(--alias --capture-mode rerun) = %d, %q", code, stderr)
	}
	if !strings.Contains(stdout, "function fuck") ||
		!strings.Contains(stdout, "TF_ALIAS_VERSION=4") ||
		strings.Contains(stdout, "GOFUCK_INSTANT_MODE") {
		t.Fatalf("rerun alias output = %q", stdout)
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

func TestMainShellLoggerUnsupported(t *testing.T) {
	resetEntrypointsState(t)
	if instant.Supported() {
		t.Skip("the shell logger starts an interactive shell on this platform")
	}
	buffer := &bytes.Buffer{}
	logs.SetOutput(buffer)

	code, _, stderr := runMain(t, "go-fuck", "-l", "log")
	if code != 1 {
		t.Fatalf("Main() = %d, want 1", code)
	}
	if !strings.Contains(buffer.String()+stderr, "Shell logger doesn't support your platform") {
		t.Fatalf("unexpected output: %q %q", buffer.String(), stderr)
	}
}

func TestMainInstantModeAlias(t *testing.T) {
	resetEntrypointsState(t)

	code, stdout, _ := runMain(t, "go-fuck", "--alias", "--enable-experimental-instant-mode")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	if !strings.Contains(stdout, "--shell-logger") {
		t.Fatalf("instant mode alias does not start the shell logger: %q", stdout)
	}
	if !strings.Contains(stdout, "GOFUCK_INSTANT_MODE=True") {
		t.Fatalf("instant mode alias does not export the mode: %q", stdout)
	}
	if strings.Contains(stdout, "function fuck") {
		t.Fatalf("instant mode initialization should not define the alias: %q", stdout)
	}

	t.Setenv(consts.InstantModeEnv, "true")
	code, stdout, _ = runMain(t, "go-fuck", "--alias", "--enable-experimental-instant-mode")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	if !strings.Contains(stdout, "__gofuck_mark=$'") || !strings.Contains(stdout, "function fuck") {
		t.Fatalf("logged shell alias is missing the prompt marker or the alias: %q", stdout)
	}
}

func TestMainPowershellInstantModeAlias(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell instant mode is only available on Windows")
	}
	resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "powershell")

	code, stdout, _ := runMain(t, "go-fuck", "--alias", "--enable-experimental-instant-mode")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	for _, want := range []string{"$env:GOFUCK_INSTANT_MODE", "$env:GOFUCK_OUTPUT_LOG", "$env:GOFUCK_SHELL", "--shell-logger"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("instant mode initializer does not contain %q: %q", want, stdout)
		}
	}

	t.Setenv(consts.InstantModeEnv, "true")
	code, stdout, _ = runMain(t, "go-fuck", "--alias", "--enable-experimental-instant-mode")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	for _, want := range []string{"global:prompt", "function fuck", "[char]0x200B"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("logged shell alias does not contain %q: %q", want, stdout)
		}
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
	if !strings.Contains(stderr, "FAIL") {
		t.Fatalf("stderr = %q, want the shared failure status", stderr)
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

// helperEnv selects the behavior of the test binary when it is executed as a
// fake external program.
const helperEnv = "GOFUCK_ENTRYPOINT_HELPER"

// TestMain lets the test binary act as a fake external program.
func TestMain(m *testing.M) {
	if mode := os.Getenv(helperEnv); mode != "" {
		os.Exit(runHelper(mode))
	}
	os.Exit(m.Run())
}

// runHelper implements the fake program modes.
func runHelper(mode string) int {
	switch mode {
	case "profile":
		fmt.Print("/tmp/go-fuck-fake-profile.ps1")
		return 0
	case "profile-empty":
		fmt.Print("   ")
		return 0
	case "fail":
		return 3
	}
	return 2
}

// installFakeExecutable copies the test binary into a temporary directory under
// the given name and points PATH at it.
func installFakeExecutable(t *testing.T, name, mode string) {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() error = %v", err)
	}
	data, err := os.ReadFile(executable)
	if err != nil {
		t.Fatalf("read the test binary: %v", err)
	}
	fileName := name
	if runtime.GOOS == "windows" {
		fileName += ".exe"
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, fileName), data, 0o755); err != nil {
		t.Fatalf("write the fake executable: %v", err)
	}
	t.Setenv("PATH", dir)
	t.Setenv(helperEnv, mode)
}

// makeReadOnly removes the write permission of a file. Tests that depend on the
// permission being enforced skip themselves when running as root.
func makeReadOnly(t *testing.T, path string) {
	t.Helper()
	if err := os.Chmod(path, 0o444); err != nil {
		t.Fatalf("make %s read only: %v", path, err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })
}

func TestMainFishInstantModeAlias(t *testing.T) {
	resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "fish")

	code, stdout, _ := runMain(t, "go-fuck", "--alias", "--enable-experimental-instant-mode")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	for _, want := range []string{"status --is-interactive", "--shell-logger", "GOFUCK_OUTPUT_LOG", "status fish-path", "< /dev/tty"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("fish instant mode initializer does not contain %q: %q", want, stdout)
		}
	}
	if strings.Contains(stdout, "function fuck") {
		t.Fatalf("instant mode initialization should not define the alias: %q", stdout)
	}

	t.Setenv(consts.InstantModeEnv, "true")
	code, stdout, _ = runMain(t, "go-fuck", "--alias", "--enable-experimental-instant-mode")
	if code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	for _, want := range []string{"function __gofuck_mark_prompt --on-event fish_prompt", "function fuck"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("fish marked alias does not contain %q: %q", want, stdout)
		}
	}
}
