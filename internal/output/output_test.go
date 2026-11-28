package output

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/leafevans/go-fuck/v4/internal/conf"
)

// resetOutputState resets the settings used by the output tests.
func resetOutputState(t *testing.T) {
	t.Helper()
	conf.Reset()
	t.Cleanup(conf.Reset)
}

// slowCommand returns a portable command that runs longer than one second.
func slowCommand() string {
	if runtime.GOOS == "windows" {
		return "ping -n 6 127.0.0.1"
	}
	return "sleep 5"
}

func TestRerun(t *testing.T) {
	resetOutputState(t)
	// The whole test suite runs packages in parallel; a generous wait keeps a
	// loaded machine from turning the healthy echo into a timeout.
	conf.Current().WaitCommand = 15
	command := "echo hello"
	output, ok := Rerun(command)
	if !ok {
		t.Fatal("Rerun() did not report an output")
	}
	if !strings.Contains(output, "hello") {
		t.Fatalf("Rerun() = %q, want it to contain hello", output)
	}
}

func TestRerunMergesConfiguredEnvironment(t *testing.T) {
	resetOutputState(t)
	conf.Current().WaitCommand = 15
	conf.Current().Env = map[string]string{"GOFUCK_TEST_ENV": "value"}

	var command string
	if runtime.GOOS == "windows" {
		// The command is run through cmd.exe explicitly so that the test does
		// not depend on the shell that re-runs it.
		command = "cmd /c echo %GOFUCK_TEST_ENV%"
	} else {
		command = "echo $GOFUCK_TEST_ENV"
	}
	output, ok := Rerun(command)
	if !ok {
		t.Fatal("Rerun() did not report an output")
	}
	if !strings.Contains(output, "value") {
		t.Fatalf("Rerun() = %q, the configured environment was not merged", output)
	}
}

func TestRerunTimeout(t *testing.T) {
	resetOutputState(t)
	conf.Current().WaitCommand = 1
	command := slowCommand()

	started := time.Now()
	output, ok := Rerun(command)
	elapsed := time.Since(started)

	if ok {
		t.Fatalf("Rerun() reported an output for a timed out command: %q", output)
	}
	// The bound is generous on purpose: the whole test suite runs packages in
	// parallel, so a loaded machine can delay the kill by a few seconds.
	if elapsed > 20*time.Second {
		t.Fatalf("Rerun() took %s, the command was not killed on timeout", elapsed)
	}
}
