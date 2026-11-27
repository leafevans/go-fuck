//go:build !windows

package capture

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/leafevans/go-fuck/v4/internal/consts"
)

func TestInstantCaptureDoesNotRepeatSideEffects(t *testing.T) {
	resetCaptureState(t)
	counter := filepath.Join(t.TempDir(), "runs")
	t.Setenv("GOFUCK_TEST_COUNTER", counter)
	const command = "echo run >> \"$GOFUCK_TEST_COUNTER\"; echo diagnostic"
	if out, err := exec.Command("sh", "-c", command).CombinedOutput(); err != nil || string(out) != "diagnostic\n" {
		t.Fatalf("original command: %q, %v", out, err)
	}
	t.Setenv(consts.OutputLogEnv, captureLog(t, command, "diagnostic"))
	request := Request{Script: command, Expanded: command, AllowRerun: true, Mode: ModeInstant}
	if got := Capture(request); !got.OK || got.Source != SourceInstantLog || got.Output != "diagnostic" {
		t.Fatalf("capture with log: %+v", got)
	}
	t.Setenv(consts.OutputLogEnv, captureLog(t, "echo unrelated", "other output"))
	if got := Capture(request); got.OK || got.Source != SourceNone || got.Output != "" {
		t.Fatalf("capture without a matching entry: %+v", got)
	}
	if data, err := os.ReadFile(counter); err != nil || string(data) != "run\n" {
		t.Fatalf("original command should run once; counter = %q, %v", data, err)
	}
}
