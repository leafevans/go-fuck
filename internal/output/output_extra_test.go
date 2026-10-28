package output

import (
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/leafevans/go-fuck/internal/conf"
)

// envValue returns the value of a key in an environment slice and the number
// of times it appears.
func envValue(environment []string, key string) (string, int) {
	value := ""
	count := 0
	for _, entry := range environment {
		entryKey, entryValue, found := strings.Cut(entry, "=")
		if !found || entryKey != key {
			continue
		}
		value = entryValue
		count++
	}
	return value, count
}

func TestGetOutputCapturesStderr(t *testing.T) {
	resetOutputState(t)
	// A generous timeout keeps the test reliable while other packages run in
	// parallel and the machine is under load.
	conf.Current().WaitCommand = 60
	command := "echo stderr-marker 1>&2"

	result, ok := GetOutput(command, command)
	if !ok {
		t.Fatal("GetOutput() did not report an output")
	}
	if !strings.Contains(result, "stderr-marker") {
		t.Fatalf("GetOutput() = %q, want it to contain the stderr output", result)
	}
}

func TestGetOutputUsesWaitSlowCommand(t *testing.T) {
	resetOutputState(t)
	command := slowCommand()

	settings := conf.Current()
	// WaitCommand is much larger than the runtime of the command: when the
	// slow command timeout is not applied the command finishes successfully
	// and the test fails.
	settings.WaitCommand = 60
	settings.WaitSlowCommand = 1
	settings.SlowCommands = []string{strings.Fields(command)[0]}

	started := time.Now()
	result, ok := GetOutput(command, command)
	elapsed := time.Since(started)

	if ok {
		t.Fatalf("GetOutput() = %q for a timed out slow command", result)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("GetOutput() took %s, the slow command timeout was not applied", elapsed)
	}
}

func TestGetOutputIgnoresWaitSlowCommandForFastCommands(t *testing.T) {
	resetOutputState(t)
	command := echoCommand("fast")

	settings := conf.Current()
	settings.WaitCommand = 60
	settings.WaitSlowCommand = 1
	settings.SlowCommands = []string{"definitely-not-this-command"}

	result, ok := GetOutput(command, command)
	if !ok {
		t.Fatal("GetOutput() did not report an output for a fast command")
	}
	if !strings.Contains(result, "fast") {
		t.Fatalf("GetOutput() = %q, want it to contain fast", result)
	}
}

func TestShellCommandPlatform(t *testing.T) {
	command := "echo hello"
	want := []string{"/bin/sh", "-c", command}
	if runtime.GOOS == "windows" {
		t.Setenv("COMSPEC", "")
		want = []string{"cmd.exe", "/c", command}
	}

	cmd := shellCommand(command)
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("shellCommand() args = %q, want %q", cmd.Args, want)
	}
}

func TestKillTreeWithoutStartedProcess(t *testing.T) {
	// killTree must be a no-op when the command was never started.
	killTree(shellCommand("echo hello"))
}

func TestMergedEnvOverridesExistingVariables(t *testing.T) {
	t.Setenv("GOFUCK_MERGE_TEST", "old")
	t.Setenv("GOFUCK_MERGE_KEPT", "kept")

	environment := mergedEnv(map[string]string{
		"GOFUCK_MERGE_TEST": "new",
		"GOFUCK_MERGE_NEW":  "added",
	})

	value, count := envValue(environment, "GOFUCK_MERGE_TEST")
	if value != "new" || count != 1 {
		t.Fatalf("mergedEnv() GOFUCK_MERGE_TEST = %q (count %d), want new (count 1)", value, count)
	}
	value, count = envValue(environment, "GOFUCK_MERGE_NEW")
	if value != "added" || count != 1 {
		t.Fatalf("mergedEnv() GOFUCK_MERGE_NEW = %q (count %d), want added (count 1)", value, count)
	}
	value, _ = envValue(environment, "GOFUCK_MERGE_KEPT")
	if value != "kept" {
		t.Fatalf("mergedEnv() GOFUCK_MERGE_KEPT = %q, want kept", value)
	}

	environment = mergedEnv(nil)
	value, _ = envValue(environment, "GOFUCK_MERGE_TEST")
	if value != "old" {
		t.Fatalf("mergedEnv() GOFUCK_MERGE_TEST = %q without overrides, want old", value)
	}
}
