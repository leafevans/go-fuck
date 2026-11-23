package types

import "testing"

func TestCorrectedCommandExplanationAndBeforeRun(t *testing.T) {
	called := false
	action := func(*Command, string) error { called = true; return nil }
	for _, tc := range []struct {
		name    string
		command CorrectedCommand
		reason  string
		before  string
	}{
		{name: "no metadata"},
		{name: "rule fallback", command: CorrectedCommand{RuleName: "git_not_command"}, reason: "Matched rule: git_not_command"},
		{name: "custom reason", command: CorrectedCommand{RuleName: "git_not_command", Reason: "Correct a typo"}, reason: "Correct a typo"},
		{name: "action fallback", command: CorrectedCommand{SideEffect: action}, before: "Run an additional rule action"},
		{name: "described action", command: CorrectedCommand{SideEffect: action, SideEffectDescription: "Refresh cache"}, before: "Refresh cache"},
		{name: "description without action", command: CorrectedCommand{SideEffectDescription: "Refresh cache"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.command.Explanation(); got != tc.reason {
				t.Errorf("Explanation() = %q, want %q", got, tc.reason)
			}
			if got := tc.command.BeforeRun(); got != tc.before {
				t.Errorf("BeforeRun() = %q, want %q", got, tc.before)
			}
		})
	}
	if called {
		t.Fatal("describing a suggestion executed its side effect")
	}
}
