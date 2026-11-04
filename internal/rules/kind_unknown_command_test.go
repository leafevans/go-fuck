package rules

import "testing"

// kindUnknownCommandOutput is the output of kind for a misspelled top level
// command; the suggestions follow the "Did you mean this?" line.
func kindUnknownCommandOutput() string {
	return "ERROR: unknown command \"creat\" for \"kind\"\n" +
		"\n" +
		"Did you mean this?\n" +
		"\tcreate\n" +
		"\n" +
		"Run 'kind --help' for usage.\n"
}

func TestRuleKindUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("kind_unknown_command")
	if rule == nil {
		t.Fatal("kind_unknown_command rule is not registered")
	}

	assertMatch(t, rule, "kind creat", kindUnknownCommandOutput())
	assertNotMatch(t, rule, "kubectl creat", kindUnknownCommandOutput())
	assertNotMatch(t, rule, "kind create", "Creating cluster \"kind\" ...\n")
	assertNotMatch(t, rule, "kind creat", "ERROR: unknown command \"creat\" for \"kind\"\n")

	got, err := rule.GetNewCommand(testCommand("kind creat", kindUnknownCommandOutput()))
	if err != nil {
		t.Fatalf("rule returned an error: %v", err)
	}
	if len(got) != 1 || got[0] != "kind create" {
		t.Fatalf("rule returned %q, want [kind create]", got)
	}
}
