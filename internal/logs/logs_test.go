package logs

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/consts"
)

// resetLogsState redirects the log output to a buffer and disables colors.
func resetLogsState(t *testing.T) *bytes.Buffer {
	t.Helper()
	conf.Reset()
	conf.Current().NoColors = true
	buffer := &bytes.Buffer{}
	SetOutput(buffer)
	t.Cleanup(func() {
		conf.Reset()
		SetOutput(io.Discard)
	})
	return buffer
}

func TestWarn(t *testing.T) {
	buffer := resetLogsState(t)
	Warn("title")
	if got, want := buffer.String(), "[WARN] title"+newline; got != want {
		t.Fatalf("Warn() wrote %q, want %q", got, want)
	}
}

func TestWarnWithColors(t *testing.T) {
	buffer := resetLogsState(t)
	conf.Current().NoColors = false
	Warn("title")
	got := buffer.String()
	if !strings.Contains(got, "\x1b[41m") || !strings.Contains(got, "[WARN] title") {
		t.Fatalf("Warn() wrote %q, want colored output", got)
	}
}

func TestException(t *testing.T) {
	buffer := resetLogsState(t)
	Exception("Can't do it", errors.New("boom"))
	got := buffer.String()
	if !strings.Contains(got, "[WARN] Can't do it:") {
		t.Fatalf("Exception() wrote %q", got)
	}
	if !strings.Contains(got, "boom") {
		t.Fatalf("Exception() wrote %q, want the error message", got)
	}
	if !strings.Contains(got, "----------------------------") {
		t.Fatalf("Exception() wrote %q, want the separator", got)
	}
}

func TestRuleFailed(t *testing.T) {
	buffer := resetLogsState(t)
	RuleFailed("test", errors.New("denied"))
	if !strings.Contains(buffer.String(), "[WARN] Rule test:") {
		t.Fatalf("RuleFailed() wrote %q", buffer.String())
	}
}

func TestFailed(t *testing.T) {
	buffer := resetLogsState(t)
	Failed("No fucks given")
	if got, want := buffer.String(), "No fucks given"+newline; got != want {
		t.Fatalf("Failed() wrote %q, want %q", got, want)
	}
}

func TestShowCorrectedCommand(t *testing.T) {
	buffer := resetLogsState(t)
	ShowCorrectedCommand("ls", false)
	want := consts.UserCommandMark + "ls" + newline
	if got := buffer.String(); got != want {
		t.Fatalf("ShowCorrectedCommand() wrote %q, want %q", got, want)
	}

	buffer.Reset()
	ShowCorrectedCommand("ls", true)
	want = consts.UserCommandMark + "ls (+side effect)" + newline
	if got := buffer.String(); got != want {
		t.Fatalf("ShowCorrectedCommand() wrote %q, want %q", got, want)
	}
}

func TestConfirmText(t *testing.T) {
	buffer := resetLogsState(t)
	ConfirmText("ls", false)
	want := consts.UserCommandMark + "\x1b[1K\rls [enter/↑/↓/ctrl+c]"
	if got := buffer.String(); got != want {
		t.Fatalf("ConfirmText() wrote %q, want %q", got, want)
	}
}

func TestDebug(t *testing.T) {
	buffer := resetLogsState(t)
	Debug("hello %s", "world")
	if got := buffer.String(); got != "" {
		t.Fatalf("Debug() wrote %q while debug is disabled", got)
	}

	conf.Current().Debug = true
	Debug("hello %s", "world")
	if got, want := buffer.String(), "DEBUG: hello world"+newline; got != want {
		t.Fatalf("Debug() wrote %q, want %q", got, want)
	}
}

func TestDebugTime(t *testing.T) {
	buffer := resetLogsState(t)
	conf.Current().Debug = true
	stop := DebugTime("Total")
	stop()
	if got := buffer.String(); !strings.Contains(got, "DEBUG: Total took:") {
		t.Fatalf("DebugTime() wrote %q", got)
	}
}

func TestNewLine(t *testing.T) {
	buffer := resetLogsState(t)
	NewLine()
	if got := buffer.String(); got != newline {
		t.Fatalf("NewLine() wrote %q, want a newline", got)
	}
}

func TestVersion(t *testing.T) {
	buffer := resetLogsState(t)
	Version("1.2.3", "go1.22", "Bash 5.2")
	got := buffer.String()
	for _, part := range []string{consts.AppName, "1.2.3", "go1.22", "Bash 5.2"} {
		if !strings.Contains(got, part) {
			t.Fatalf("Version() wrote %q, want it to contain %q", got, part)
		}
	}
}
