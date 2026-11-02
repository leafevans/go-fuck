package rules

import "testing"

// glabUnknownCommandOutput is the output of glab for a misspelled top level
// command.
func glabUnknownCommandOutput() string {
	return `unknown command "buidl" for "glab"

Did you mean this?
	ci
	label
Usage:  glab <command> <subcommand> [flags]
Available commands:
  ci
  label
  mr
`
}

func TestRuleGlabUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("glab_unknown_command")
	if rule == nil {
		t.Fatal("glab_unknown_command rule is not registered")
	}

	assertMatch(t, rule, "glab buidl", glabUnknownCommandOutput())
	assertNotMatch(t, rule, "gh buidl", glabUnknownCommandOutput())
	assertNotMatch(t, rule, "glab mr", "! [remote rejected]")
	assertNotMatch(t, rule, "glab buidl", `unknown command "buidl" for "glab"`)

	got, err := rule.GetNewCommand(testCommand("glab buidl", glabUnknownCommandOutput()))
	if err != nil {
		t.Fatalf("rule %q returned an error: %v", rule.Name, err)
	}
	want := []string{"glab ci", "glab label"}
	if len(got) != len(want) {
		t.Fatalf("rule %q returned %q, want %q", rule.Name, got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("rule %q returned %q, want %q", rule.Name, got, want)
		}
	}
}
