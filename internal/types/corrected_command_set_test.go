package types

import "testing"

func setEffectOne(*Command, string) error { return nil }
func setEffectTwo(*Command, string) error { return nil }

func TestCorrectedCommandSetAgreesWithEqual(t *testing.T) {
	commands := []*CorrectedCommand{
		{Script: "git status"},
		{Script: "git status", Priority: 10, RuleName: "other_rule"},
		{Script: "git status", SideEffect: setEffectOne},
		{Script: "git status", SideEffect: setEffectTwo},
		{Script: "git status", SideEffect: setEffectOne, Priority: 1},
		{Script: "git diff", SideEffect: setEffectOne},
		{Script: "git diff"},
	}
	for _, first := range commands {
		for _, second := range commands {
			var seen CorrectedCommandSet
			if !seen.Add(first) {
				t.Fatal("the first command was treated as a duplicate")
			}
			if got, want := seen.Add(second), !first.Equal(second); got != want {
				t.Fatalf("Add(%v) after %v = %t, want %t", second, first, got, want)
			}
		}
	}
	seen := NewCorrectedCommandSet(len(commands))
	for i, want := range []bool{true, false, true, true, false, true, true} {
		if got := seen.Add(commands[i]); got != want {
			t.Fatalf("Add(command %d) = %t, want %t", i, got, want)
		}
	}
}
