package types

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"errors"

	"github.com/leafevans/go-fuck/v2/internal/conf"
	"github.com/leafevans/go-fuck/v2/internal/consts"
	"github.com/leafevans/go-fuck/v2/internal/logs"
	"github.com/leafevans/go-fuck/v2/internal/shells"
)

// resetTypesState resets the state shared by the tests in this package.
func resetTypesState(t *testing.T) {
	t.Helper()
	conf.Reset()
	// Command fixtures are executed by the output reader, so give it a generous
	// timeout to keep the test stable on a heavily loaded machine.
	conf.Current().WaitCommand = 30
	shells.Set(shells.NewGeneric())
	logs.SetOutput(io.Discard)
	t.Cleanup(func() {
		conf.Reset()
		shells.Set(shells.NewGeneric())
		logs.SetOutput(io.Discard)
		stdout = nil
	})
}

// newStdoutBuffer redirects the corrected command output to a buffer.
func newStdoutBuffer(t *testing.T) *bytes.Buffer {
	t.Helper()
	buffer := &bytes.Buffer{}
	previous := stdout
	stdout = buffer
	t.Cleanup(func() {
		stdout = previous
	})
	return buffer
}

func TestCommandFields(t *testing.T) {
	resetTypesState(t)

	command := NewCommand("ls -la", "output")
	if command.Script != "ls -la" || command.Output != "output" || !command.HasOutput() {
		t.Fatalf("unexpected command: %+v", command)
	}

	withoutOutput := NewCommandWithoutOutput("ls")
	if withoutOutput.HasOutput() {
		t.Fatal("command without output reports an output")
	}
}

func TestCommandScriptParts(t *testing.T) {
	resetTypesState(t)

	command := NewCommand(`git commit -m "hello world"`, "")
	parts := command.ScriptParts()
	want := []string{"git", "commit", "-m", "hello world"}
	if strings.Join(parts, "|") != strings.Join(want, "|") {
		t.Fatalf("ScriptParts() = %q, want %q", parts, want)
	}

	// The parts are cached between calls; callers that need to mutate the
	// result have to copy it first.
	if second := command.ScriptParts(); len(second) != len(want) || second[0] != "git" {
		t.Fatalf("ScriptParts() is not cached correctly: %q", second)
	}
}

func TestCommandUpdate(t *testing.T) {
	resetTypesState(t)

	command := NewCommand("git brnch", "output")
	updated := command.WithScript("git branch")
	if command.Script != "git brnch" {
		t.Fatal("WithScript mutated the original command")
	}
	if updated.Script != "git branch" || updated.Output != "output" || !updated.HasOutput() {
		t.Fatalf("unexpected updated command: %+v", updated)
	}

	withOutput := command.WithOutput("new output")
	if withOutput.Script != "git brnch" || withOutput.Output != "new output" {
		t.Fatalf("unexpected command: %+v", withOutput)
	}

	both := command.Update("next", "next output")
	if both.Script != "next" || both.Output != "next output" {
		t.Fatalf("unexpected command: %+v", both)
	}
}

func TestFromRawScriptEmpty(t *testing.T) {
	resetTypesState(t)

	for _, raw := range [][]string{{}, {""}, {"", ""}, {"   "}} {
		if _, err := FromRawScript(raw, true); err != ErrEmptyCommand {
			t.Errorf("FromRawScript(%q) error = %v, want ErrEmptyCommand", raw, err)
		}
	}
}

func TestFromRawScript(t *testing.T) {
	resetTypesState(t)

	command, err := FromRawScript([]string{"echo", "hello"}, true)
	if err != nil {
		t.Fatalf("FromRawScript() error = %v", err)
	}
	if command.Script != "echo hello" {
		t.Fatalf("Script = %q, want %q", command.Script, "echo hello")
	}
	if !command.HasOutput() {
		t.Fatal("echo output was not captured")
	}
	if !strings.Contains(command.Output, "hello") {
		t.Fatalf("Output = %q, want it to contain %q", command.Output, "hello")
	}
}

func TestFormatRawScript(t *testing.T) {
	if got := FormatRawScript([]string{" ls", "-la"}); got != "ls -la" {
		t.Fatalf("FormatRawScript() = %q, want %q", got, "ls -la")
	}
}

func TestRuleDefaults(t *testing.T) {
	resetTypesState(t)

	rule := NewRule("test", func(*Command) bool { return true }, func(*Command) ([]string, error) { return nil, nil })
	if !rule.EnabledByDefault || !rule.RequiresOutput || rule.Priority != 1000 {
		t.Fatalf("unexpected defaults: %+v", rule)
	}
}

func TestRuleIsEnabled(t *testing.T) {
	resetTypesState(t)

	enabled := NewRule("git", nil, nil)
	disabled := NewRule("git", nil, nil)
	disabled.EnabledByDefault = false

	cases := []struct {
		rules    []string
		rule     *Rule
		expected bool
	}{
		{[]string{"All rules enabled"}, enabled, true},
		{[]string{"All rules enabled"}, disabled, false},
		{[]string{}, enabled, false},
		{[]string{}, disabled, false},
		{[]string{"All rules enabled", "git"}, disabled, true},
		{[]string{"git"}, disabled, true},
		{[]string{"other"}, enabled, false},
	}
	for _, testCase := range cases {
		conf.Current().Rules = testCase.rules
		if got := testCase.rule.IsEnabled(); got != testCase.expected {
			t.Errorf("IsEnabled() with rules %v = %v, want %v", testCase.rules, got, testCase.expected)
		}
	}
}

func TestRuleIsEnabledProbesOnlyDefaults(t *testing.T) {
	resetTypesState(t)
	rule := NewRule("tool", nil, nil)
	calls := 0
	rule.DefaultProbe = func() bool {
		calls++
		return false
	}
	conf.Current().Rules = []string{consts.AllEnabled, rule.Name}
	if !rule.IsEnabled() || calls != 0 {
		t.Fatal("explicitly enabled rule should bypass the default probe")
	}
	conf.Current().Rules = nil
	if rule.IsEnabled() || calls != 0 {
		t.Fatal("unlisted rule should bypass the default probe in an explicit rule set")
	}
	conf.Current().Rules = []string{consts.AllEnabled}
	if rule.IsEnabled() || calls != 1 {
		t.Fatal("default rule set should consult the default probe")
	}
}

func TestRuleIsMatch(t *testing.T) {
	resetTypesState(t)

	rule := NewRule("test", func(command *Command) bool {
		return command.Script == "cd .."
	}, nil)
	if !rule.IsMatch(NewCommand("cd ..", "")) {
		t.Fatal("rule should match")
	}
	if rule.IsMatch(NewCommand("cd /", "")) {
		t.Fatal("rule should not match")
	}
}

func TestRuleRequiresOutput(t *testing.T) {
	resetTypesState(t)

	rule := NewRule("test", func(*Command) bool { return true }, nil)
	if rule.IsMatch(NewCommandWithoutOutput("ls")) {
		t.Fatal("rule requiring output matched a command without output")
	}
	rule.RequiresOutput = false
	if !rule.IsMatch(NewCommandWithoutOutput("ls")) {
		t.Fatal("rule not requiring output did not match")
	}
}

func TestRuleFailureIsReported(t *testing.T) {
	resetTypesState(t)

	buffer := &bytes.Buffer{}
	logs.SetOutput(buffer)
	rule := NewRule("test", func(*Command) bool { panic("denied") }, nil)
	rule.RequiresOutput = false
	if rule.IsMatch(NewCommand("ls", "")) {
		t.Fatal("panicking rule reported a match")
	}
	firstLine, _, _ := strings.Cut(buffer.String(), string(rune(10)))
	// The warning prefix is colored, so the label and the title are checked
	// separately instead of relying on one contiguous string.
	if !strings.Contains(firstLine, "WARN") || !strings.Contains(firstLine, "Rule test:") {
		t.Fatalf("unexpected log output: %q", buffer.String())
	}
}

func TestGetCorrectedCommands(t *testing.T) {
	resetTypesState(t)

	rule := NewRule("test", nil, func(command *Command) ([]string, error) {
		return []string{command.Script + "!", command.Script + "@"}, nil
	})
	rule.Priority = 100
	corrected := rule.GetCorrectedCommands(NewCommand("ls", ""))
	if len(corrected) != 2 {
		t.Fatalf("got %d corrected commands, want 2", len(corrected))
	}
	if corrected[0].Script != "ls!" || corrected[0].Priority != 100 {
		t.Fatalf("unexpected first command: %+v", corrected[0])
	}
	if corrected[1].Script != "ls@" || corrected[1].Priority != 200 {
		t.Fatalf("unexpected second command: %+v", corrected[1])
	}
}

func TestCorrectedCommandEqual(t *testing.T) {
	if !(&CorrectedCommand{Script: "ls", Priority: 100}).Equal(&CorrectedCommand{Script: "ls", Priority: 200}) {
		t.Fatal("priority should be ignored")
	}
	if (&CorrectedCommand{Script: "ls"}).Equal(&CorrectedCommand{Script: "ls -la"}) {
		t.Fatal("different scripts are equal")
	}
	sideEffect := func(*Command, string) error { return nil }
	if (&CorrectedCommand{Script: "ls"}).Equal(&CorrectedCommand{Script: "ls", SideEffect: sideEffect}) {
		t.Fatal("different side effects are equal")
	}
	if !(&CorrectedCommand{Script: "ls", SideEffect: sideEffect}).Equal(
		&CorrectedCommand{Script: "ls", SideEffect: sideEffect}) {
		t.Fatal("same side effects are not equal")
	}
}

func TestCorrectedCommandRun(t *testing.T) {
	resetTypesState(t)

	cases := []struct {
		name     string
		repeat   bool
		debug    bool
		expected string
	}{
		{"plain", false, false, "git brunch"},
		{"repeat", true, false, "git brunch || fuck --repeat --force-command 'git brunch'"},
		{"repeat debug", true, true, "git brunch || fuck --repeat --debug --force-command 'git brunch'"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			resetTypesState(t)
			conf.Current().Repeat = testCase.repeat
			conf.Current().Debug = testCase.debug
			conf.Current().AlterHistory = false
			buffer := newStdoutBuffer(t)

			(&CorrectedCommand{Script: "git brunch", Priority: 1000}).Run(NewCommand("git brunch", ""))
			if buffer.String() != testCase.expected {
				t.Fatalf("Run() wrote %q, want %q", buffer.String(), testCase.expected)
			}
		})
	}
}

// recordingShell records history entries.
type recordingShell struct {
	*shells.Generic
	recorded []string
}

func (r *recordingShell) PutToHistory(command string) error {
	r.recorded = append(r.recorded, command)
	return nil
}

func TestCorrectedCommandRunSideEffectAndHistory(t *testing.T) {
	resetTypesState(t)
	shell := &recordingShell{Generic: shells.NewGeneric()}
	shells.Set(shell)
	conf.Current().AlterHistory = true
	newStdoutBuffer(t)

	var sideEffectArguments []string
	command := NewCommand("ls", "")
	corrected := &CorrectedCommand{
		Script: "ls -la",
		SideEffect: func(oldCommand *Command, script string) error {
			sideEffectArguments = []string{oldCommand.Script, script}
			return nil
		},
	}
	corrected.Run(command)

	if len(sideEffectArguments) != 2 || sideEffectArguments[0] != "ls" || sideEffectArguments[1] != "ls -la" {
		t.Fatalf("unexpected side effect arguments: %q", sideEffectArguments)
	}
	if len(shell.recorded) != 1 || shell.recorded[0] != "ls -la" {
		t.Fatalf("unexpected history: %q", shell.recorded)
	}
}

// TestFromScriptAndOutput checks that a caller supplied output is used as is
// instead of re-running the command.
func TestFromScriptAndOutput(t *testing.T) {
	resetTypesState(t)

	command, err := FromScriptAndOutput([]string{"git", "push"}, "fatal: no upstream branch\n")
	if err != nil {
		t.Fatalf("FromScriptAndOutput() error = %v", err)
	}
	if command.Script != "git push" || command.Output != "fatal: no upstream branch\n" || !command.HasOutput() {
		t.Fatalf("unexpected command: %+v", command)
	}

	for _, raw := range [][]string{{}, {""}, {"", ""}, {"   "}} {
		if _, err := FromScriptAndOutput(raw, "output"); err != ErrEmptyCommand {
			t.Errorf("FromScriptAndOutput(%q) error = %v, want ErrEmptyCommand", raw, err)
		}
	}
}

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
