package ui

import (
	"bytes"
	"io"
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/types"
)

// testLogBuffer receives the log output of the UI tests.
var testLogBuffer = &bytes.Buffer{}

// resetUIState resets the global state used by the UI tests.
func resetUIState(t *testing.T) {
	t.Helper()
	conf.Reset()
	conf.Current().NoColors = true
	testLogBuffer = &bytes.Buffer{}
	logs.SetOutput(testLogBuffer)
	previousReadKey := ReadKey
	t.Cleanup(func() {
		conf.Reset()
		logs.SetOutput(io.Discard)
		ReadKey = previousReadKey
	})
}

// setKeys makes ReadKey return the given keys in order.
func setKeys(t *testing.T, keys ...consts.Key) {
	t.Helper()
	index := 0
	ReadKey = func() (consts.Key, error) {
		if index >= len(keys) {
			return "", io.EOF
		}
		key := keys[index]
		index++
		return key, nil
	}
}

// logOutput returns the log output captured by resetUIState.
func logOutput() string {
	return testLogBuffer.String()
}

func TestReadActions(t *testing.T) {
	resetUIState(t)
	setKeys(t, "\n", "\r", "x", "y", consts.KeyUp, "k", consts.KeyDown, "j", consts.KeyCtrlC, "q")

	want := []consts.Action{
		consts.ActionSelect, consts.ActionSelect,
		consts.ActionPrevious, consts.ActionPrevious,
		consts.ActionNext, consts.ActionNext,
		consts.ActionAbort, consts.ActionAbort,
	}
	next := ReadActions()
	for i, expected := range want {
		if got := next(); got != expected {
			t.Fatalf("action %d = %q, want %q", i, got, expected)
		}
	}
}

func TestCommandSelector(t *testing.T) {
	commands := []*types.CorrectedCommand{
		{Script: "ls", Priority: 100},
		{Script: "cd", Priority: 100},
	}
	selector, ok := NewCommandSelector(commands)
	if !ok {
		t.Fatal("NewCommandSelector() failed")
	}
	if selector.Value().Script != "ls" {
		t.Fatalf("unexpected initial value: %q", selector.Value().Script)
	}
	selector.Next()
	if selector.Value().Script != "cd" {
		t.Fatalf("unexpected value after Next: %q", selector.Value().Script)
	}
	selector.Next()
	if selector.Value().Script != "ls" {
		t.Fatalf("unexpected value after Next wrap: %q", selector.Value().Script)
	}
	selector.Previous()
	if selector.Value().Script != "cd" {
		t.Fatalf("unexpected value after Previous: %q", selector.Value().Script)
	}
}

func TestNewCommandSelectorEmpty(t *testing.T) {
	if _, ok := NewCommandSelector(nil); ok {
		t.Fatal("NewCommandSelector(nil) should fail")
	}
}

func TestSelectCommandEmpty(t *testing.T) {
	resetUIState(t)
	if command := SelectCommand(nil); command != nil {
		t.Fatalf("SelectCommand(nil) = %v, want nil", command)
	}
	if got := logOutput(); got != "No fucks given"+newline {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestSelectCommandEmptyCustomAlias(t *testing.T) {
	resetUIState(t)
	t.Setenv("TF_ALIAS", "please")
	if command := SelectCommand(nil); command != nil {
		t.Fatalf("SelectCommand(nil) = %v, want nil", command)
	}
	if got := logOutput(); got != "Nothing found"+newline {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestSelectCommandWithoutConfirmation(t *testing.T) {
	resetUIState(t)
	conf.Current().RequireConfirmation = false
	commands := []*types.CorrectedCommand{{Script: "ls", Priority: 100}}

	selected := SelectCommand(commands)
	if selected == nil || selected.Script != "ls" {
		t.Fatalf("SelectCommand() = %v, want ls", selected)
	}
	want := consts.UserCommandMark + "ls" + newline
	if got := logOutput(); got != want {
		t.Fatalf("unexpected output: %q, want %q", got, want)
	}
}

func TestSelectCommandWithConfirmation(t *testing.T) {
	resetUIState(t)
	setKeys(t, "\n")
	commands := []*types.CorrectedCommand{{Script: "ls", Priority: 100}}

	if selected := SelectCommand(commands); selected == nil || selected.Script != "ls" {
		t.Fatalf("SelectCommand() = %v, want ls", selected)
	}
	want := consts.UserCommandMark + "\x1b[1K\rls [enter/↑/↓/ctrl+c]" + newline
	if got := logOutput(); got != want {
		t.Fatalf("unexpected output: %q, want %q", got, want)
	}
}

func TestSelectCommandWithConfirmationAbort(t *testing.T) {
	resetUIState(t)
	setKeys(t, consts.KeyCtrlC)
	commands := []*types.CorrectedCommand{{Script: "ls", Priority: 100}}

	if selected := SelectCommand(commands); selected != nil {
		t.Fatalf("SelectCommand() = %v, want nil", selected)
	}
	want := consts.UserCommandMark + "\x1b[1K\rls [enter/↑/↓/ctrl+c]" + newline + "Aborted" + newline
	if got := logOutput(); got != want {
		t.Fatalf("unexpected output: %q, want %q", got, want)
	}
}

func TestSelectCommandWithSideEffect(t *testing.T) {
	resetUIState(t)
	setKeys(t, "\n")
	sideEffect := func(*types.Command, string) error { return nil }
	commands := []*types.CorrectedCommand{{Script: "ls", SideEffect: sideEffect, Priority: 100}}

	if selected := SelectCommand(commands); selected == nil {
		t.Fatal("SelectCommand() returned nil")
	}
	want := consts.UserCommandMark + "\x1b[1K\rls (+side effect) [enter/↑/↓/ctrl+c]" + newline
	if got := logOutput(); got != want {
		t.Fatalf("unexpected output: %q, want %q", got, want)
	}
}

func TestSelectCommandSelectSecond(t *testing.T) {
	resetUIState(t)
	setKeys(t, consts.KeyDown, "\n")
	commands := []*types.CorrectedCommand{
		{Script: "ls", Priority: 100},
		{Script: "cd", Priority: 100},
	}

	selected := SelectCommand(commands)
	if selected == nil || selected.Script != "cd" {
		t.Fatalf("SelectCommand() = %v, want cd", selected)
	}
	want := consts.UserCommandMark + "\x1b[1K\rls [enter/↑/↓/ctrl+c]" +
		consts.UserCommandMark + "\x1b[1K\rcd [enter/↑/↓/ctrl+c]" + newline
	if got := logOutput(); got != want {
		t.Fatalf("unexpected output: %q, want %q", got, want)
	}
}

func TestReadActionsKeyError(t *testing.T) {
	resetUIState(t)
	ReadKey = func() (consts.Key, error) { return "", io.EOF }
	next := ReadActions()
	if action := next(); action != consts.ActionAbort {
		t.Fatalf("action = %q, want abort", action)
	}
}
