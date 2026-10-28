package output

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/leafevans/go-fuck/internal/conf"
)

// resetOutputState resets the settings used by the output tests.
func resetOutputState(t *testing.T) {
	t.Helper()
	conf.Reset()
	t.Cleanup(conf.Reset)
}

// echoCommand returns a portable command that prints the given text.
func echoCommand(text string) string {
	if runtime.GOOS == "windows" {
		return "echo " + text
	}
	return "echo " + text
}

// slowCommand returns a portable command that runs longer than one second.
func slowCommand() string {
	if runtime.GOOS == "windows" {
		return "ping -n 6 127.0.0.1"
	}
	return "sleep 5"
}

func TestGetOutput(t *testing.T) {
	resetOutputState(t)
	command := echoCommand("hello")
	output, ok := GetOutput(command, command)
	if !ok {
		t.Fatal("GetOutput() did not report an output")
	}
	if !strings.Contains(output, "hello") {
		t.Fatalf("GetOutput() = %q, want it to contain hello", output)
	}
}

func TestGetOutputMergesConfiguredEnvironment(t *testing.T) {
	resetOutputState(t)
	conf.Current().Env = map[string]string{"GOFUCK_TEST_ENV": "value"}

	var command string
	if runtime.GOOS == "windows" {
		command = "echo %GOFUCK_TEST_ENV%"
	} else {
		command = "echo $GOFUCK_TEST_ENV"
	}
	output, ok := GetOutput(command, command)
	if !ok {
		t.Fatal("GetOutput() did not report an output")
	}
	if !strings.Contains(output, "value") {
		t.Fatalf("GetOutput() = %q, the configured environment was not merged", output)
	}
}

func TestGetOutputTimeout(t *testing.T) {
	resetOutputState(t)
	conf.Current().WaitCommand = 1
	command := slowCommand()

	started := time.Now()
	output, ok := GetOutput(command, command)
	elapsed := time.Since(started)

	if ok {
		t.Fatalf("GetOutput() reported an output for a timed out command: %q", output)
	}
	if elapsed > 4*time.Second {
		t.Fatalf("GetOutput() took %s, the command was not killed on timeout", elapsed)
	}
}
