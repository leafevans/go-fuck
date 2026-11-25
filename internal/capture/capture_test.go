package capture

import (
	"os"
	"path/filepath"
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

func TestProvidedOutputOverridesAllSources(t *testing.T) {
	resetCaptureState(t)
	t.Setenv(consts.OutputLogEnv, captureLog(t, "echo log", "from-log"))
	empty := ""
	result := Capture(Request{
		Script:     "echo log",
		Expanded:   "echo from-rerun",
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
		Expanded:   "echo from-rerun",
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
		Expanded:   "echo from-rerun",
		AllowRerun: true,
		Mode:       ModeAuto,
	})
	if !result.OK || result.Source != SourceRerun || !strings.Contains(result.Output, "from-rerun") {
		t.Fatalf("Capture() = %+v, want rerun fallback", result)
	}
	if result.Duration <= 0 {
		t.Fatalf("Capture() duration = %v, want elapsed rerun time", result.Duration)
	}
}

func TestInstantModeDoesNotRerun(t *testing.T) {
	resetCaptureState(t)
	result := Capture(Request{
		Script:     "echo missing",
		Expanded:   "echo must-not-run",
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
		Expanded:   "echo from-rerun",
		AllowRerun: true,
		Mode:       ModeRerun,
	})
	if !result.OK || result.Source != SourceRerun || !strings.Contains(result.Output, "from-rerun") {
		t.Fatalf("Capture() = %+v, want forced rerun output", result)
	}
}
