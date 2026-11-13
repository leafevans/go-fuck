package types

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/conf"
	"github.com/leafevans/go-fuck/v2/internal/logs"
	"github.com/leafevans/go-fuck/v2/internal/shells"
)

// TestRuleIsMatchDebugTimer covers the debug timer branch of Rule.IsMatch.
func TestRuleIsMatchDebugTimer(t *testing.T) {
	resetTypesState(t)
	conf.Current().Debug = true
	buffer := &bytes.Buffer{}
	logs.SetOutput(buffer)

	rule := NewRule("debug_rule", func(*Command) bool { return true }, nil)
	if !rule.IsMatch(NewCommand("ls", "")) {
		t.Fatal("IsMatch() = false, want true")
	}
	if !strings.Contains(buffer.String(), "Trying rule: debug_rule;") {
		t.Fatalf("debug output = %q, want the rule timer", buffer.String())
	}
}

// TestGetCorrectedCommandsRecoversFromPanic covers the panic recovery of
// Rule.GetCorrectedCommands.
func TestGetCorrectedCommandsRecoversFromPanic(t *testing.T) {
	resetTypesState(t)
	buffer := &bytes.Buffer{}
	logs.SetOutput(buffer)

	rule := NewRule("panic_rule", nil, func(*Command) ([]string, error) {
		panic("denied")
	})
	if corrected := rule.GetCorrectedCommands(NewCommand("ls", "")); corrected != nil {
		t.Fatalf("GetCorrectedCommands() = %v, want nil", corrected)
	}
	if !strings.Contains(buffer.String(), "Rule panic_rule:") {
		t.Fatalf("log output = %q, want the rule failure", buffer.String())
	}
}

// TestGetCorrectedCommandsReportsError covers the error branch of
// Rule.GetCorrectedCommands.
func TestGetCorrectedCommandsReportsError(t *testing.T) {
	resetTypesState(t)
	buffer := &bytes.Buffer{}
	logs.SetOutput(buffer)

	failure := errors.New("cannot correct")
	rule := NewRule("error_rule", nil, func(*Command) ([]string, error) {
		return []string{"ignored"}, failure
	})
	if corrected := rule.GetCorrectedCommands(NewCommand("ls", "")); corrected != nil {
		t.Fatalf("GetCorrectedCommands() = %v, want nil", corrected)
	}
	if !strings.Contains(buffer.String(), "cannot correct") {
		t.Fatalf("log output = %q, want the rule error", buffer.String())
	}
}

// TestCorrectedCommandEqualNil covers the nil comparisons of
// CorrectedCommand.Equal.
func TestCorrectedCommandEqualNil(t *testing.T) {
	var missing *CorrectedCommand
	if !missing.Equal(nil) {
		t.Fatal("Equal(nil) = false for two nil commands")
	}
	if missing.Equal(&CorrectedCommand{Script: "ls"}) {
		t.Fatal("a nil command is equal to a non nil command")
	}
	if (&CorrectedCommand{Script: "ls"}).Equal(nil) {
		t.Fatal("a non nil command is equal to a nil command")
	}
}

// TestCorrectedCommandRunReportsSideEffectFailure covers the side effect error
// branch of CorrectedCommand.Run.
func TestCorrectedCommandRunReportsSideEffectFailure(t *testing.T) {
	resetTypesState(t)
	buffer := &bytes.Buffer{}
	logs.SetOutput(buffer)
	output := newStdoutBuffer(t)

	sideEffect := func(*Command, string) error { return errors.New("side effect failed") }
	corrected := &CorrectedCommand{Script: "ls -la", SideEffect: sideEffect}
	corrected.Run(NewCommand("ls", ""))

	if got := output.String(); got != "ls -la" {
		t.Fatalf("Run() wrote %q, want %q", got, "ls -la")
	}
	if !strings.Contains(buffer.String(), "side effect failed") {
		t.Fatalf("log output = %q, want the side effect error", buffer.String())
	}
}

// failingHistoryShell fails to record history entries.
type failingHistoryShell struct {
	*shells.Generic
}

// PutToHistory always returns an error.
func (f *failingHistoryShell) PutToHistory(string) error {
	return errors.New("history is not writable")
}

// TestCorrectedCommandRunReportsHistoryFailure covers the history error branch
// of CorrectedCommand.Run.
func TestCorrectedCommandRunReportsHistoryFailure(t *testing.T) {
	resetTypesState(t)
	buffer := &bytes.Buffer{}
	logs.SetOutput(buffer)
	newStdoutBuffer(t)

	shells.Set(&failingHistoryShell{Generic: shells.NewGeneric()})
	conf.Current().AlterHistory = true

	(&CorrectedCommand{Script: "ls -la"}).Run(NewCommand("ls", ""))
	if !strings.Contains(buffer.String(), "history is not writable") {
		t.Fatalf("log output = %q, want the history error", buffer.String())
	}
}
