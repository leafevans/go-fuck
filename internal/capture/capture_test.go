package capture

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v4/internal/conf"
	"github.com/leafevans/go-fuck/v4/internal/consts"
)

func resetCaptureState(t *testing.T) {
	t.Helper()
	conf.Reset()
	conf.Current().WaitCommand = 15
	t.Setenv(consts.OutputLogEnv, "")
	t.Setenv(consts.LegacyOutputLogEnv, "")
	t.Cleanup(conf.Reset)
}

func captureLog(t *testing.T, command, commandOutput string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "instant.log")
	content := consts.UserCommandMark + "$ " + command + "\n" + commandOutput + "\n" +
		consts.UserCommandMark + "$ "
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func portableEcho(text string) string {
	if runtime.GOOS == "windows" {
		return "echo " + text
	}
	return "echo " + text
}

func TestProvidedOutputOverridesAllSources(t *testing.T) {
	resetCaptureState(t)
	t.Setenv(consts.OutputLogEnv, captureLog(t, "echo log", "from-log"))
	empty := ""
	result := Capture(Request{
		Script:     "echo log",
		Expanded:   portableEcho("from-rerun"),
		Provided:   &empty,
		AllowRerun: true,
		Mode:       ModeAuto,
	})
	if !result.OK || result.Source != SourceProvided || result.Output != "" {
		t.Fatalf("Capture() = %+v, want an empty provided output", result)
	}
}

func TestAutoPrefersInstantLog(t *testing.T) {
	resetCaptureState(t)
	t.Setenv(consts.OutputLogEnv, captureLog(t, "echo log", "from-log"))
	result := Capture(Request{
		Script:     "echo log",
		Expanded:   portableEcho("from-rerun"),
		AllowRerun: true,
		Mode:       ModeAuto,
	})
	if !result.OK || result.Source != SourceInstantLog || result.Output != "from-log" {
		t.Fatalf("Capture() = %+v, want the instant log output", result)
	}
}

func TestAutoFallsBackToRerun(t *testing.T) {
	resetCaptureState(t)
	result := Capture(Request{
		Script:     "echo fallback",
		Expanded:   portableEcho("from-rerun"),
		AllowRerun: true,
		Mode:       ModeAuto,
	})
	if !result.OK || result.Source != SourceRerun || !strings.Contains(result.Output, "from-rerun") {
		t.Fatalf("Capture() = %+v, want rerun fallback", result)
	}
}

func TestInstantModeDoesNotRerun(t *testing.T) {
	resetCaptureState(t)
	result := Capture(Request{
		Script:     "echo missing",
		Expanded:   portableEcho("must-not-run"),
		AllowRerun: true,
		Mode:       ModeInstant,
	})
	if result.OK || result.Source != SourceNone || result.Output != "" {
		t.Fatalf("Capture() = %+v, want unavailable instant output", result)
	}
}

func TestRerunModeIgnoresInstantLog(t *testing.T) {
	resetCaptureState(t)
	t.Setenv(consts.OutputLogEnv, captureLog(t, "echo rerun", "from-log"))
	result := Capture(Request{
		Script:     "echo rerun",
		Expanded:   portableEcho("from-rerun"),
		AllowRerun: true,
		Mode:       ModeRerun,
	})
	if !result.OK || result.Source != SourceRerun || !strings.Contains(result.Output, "from-rerun") {
		t.Fatalf("Capture() = %+v, want forced rerun output", result)
	}
}
