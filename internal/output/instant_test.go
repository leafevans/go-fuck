package output

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/consts"
)

// writeInstantLog writes a terminal log with one finished command call.
func writeInstantLog(t *testing.T, command, commandOutput string) string {
	t.Helper()
	log := consts.UserCommandMark + "$ " + command + "\n" + commandOutput + "\n" + consts.UserCommandMark + "$ "
	path := filepath.Join(t.TempDir(), "instant.log")
	if err := os.WriteFile(path, []byte(log), 0o600); err != nil {
		t.Fatalf("write instant log: %v", err)
	}
	return path
}

func TestGetOutputReadsInstantModeLog(t *testing.T) {
	resetOutputState(t)
	conf.Current().InstantMode = true
	logPath := writeInstantLog(t, "git brnch", "fatal: not a git repository")
	t.Setenv(consts.OutputLogEnv, logPath)
	t.Setenv(consts.LegacyOutputLogEnv, "")

	output, ok := GetOutput("git brnch", "git brnch")
	if !ok {
		t.Fatal("GetOutput() did not report the instant mode output")
	}
	if want := "fatal: not a git repository"; output != want {
		t.Fatalf("GetOutput() = %q, want %q", output, want)
	}
}

func TestGetOutputFallsBackWhenInstantLogIsMissing(t *testing.T) {
	resetOutputState(t)
	// The command is executed for real; a generous timeout keeps the test
	// reliable when the whole suite runs in parallel on a loaded machine.
	conf.Current().WaitCommand = 60
	conf.Current().InstantMode = true
	t.Setenv(consts.OutputLogEnv, filepath.Join(t.TempDir(), "missing.log"))
	t.Setenv(consts.LegacyOutputLogEnv, "")

	command := echoCommand("fallback-output")
	output, ok := GetOutput(command, command)
	if !ok || !strings.Contains(output, "fallback-output") {
		t.Fatalf("GetOutput() = %q, %t, want the re-run output", output, ok)
	}
}
