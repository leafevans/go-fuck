package rules

import "testing"

// kustomizeUnknownCommandOutput is the output of kustomize for a misspelled
// command; the suggestions follow the "Did you mean this?" line.
func kustomizeUnknownCommandOutput() string {
	return "Error: unknown command \"bild\" for \"kustomize\"\n" +
		"\n" +
		"Did you mean this?\n" +
		"\tbuild\n" +
		"\n" +
		"Run 'kustomize --help' for usage.\n"
}

func TestRuleKustomizeUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("kustomize_unknown_command")
	if rule == nil {
		t.Fatal("kustomize_unknown_command rule is not registered")
	}

	assertMatch(t, rule, "kustomize bild", kustomizeUnknownCommandOutput())
	assertNotMatch(t, rule, "kubectl bild", kustomizeUnknownCommandOutput())
	assertNotMatch(t, rule, "kustomize build .", "Error: accumulating resources\n")
	assertNotMatch(t, rule, "kustomize bild", "Error: unknown command \"bild\" for \"kustomize\"\n")

	got, err := rule.GetNewCommand(testCommand("kustomize bild", kustomizeUnknownCommandOutput()))
	if err != nil {
		t.Fatalf("rule returned an error: %v", err)
	}
	if len(got) != 1 || got[0] != "kustomize build" {
		t.Fatalf("rule returned %q, want [kustomize build]", got)
	}
}
