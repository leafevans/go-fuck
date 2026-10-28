package rules

import "testing"

const condaMistypeResponse = "\n" +
	"\n" +
	"CommandNotFoundError: No command 'conda lst'.\n" +
	"Did you mean 'conda list'?\n" +
	"\n" +
	"    "

func TestRuleCondaMistype(t *testing.T) {
	resetTestState(t)
	rule := Get("conda_mistype")
	if rule == nil {
		t.Fatal("conda_mistype rule is not registered")
	}

	assertMatch(t, rule, "conda lst", condaMistypeResponse)
	assertNotMatch(t, rule, "codna list", "bash: codna: command not found")

	assertNewCommands(t, rule, "conda lst", condaMistypeResponse,
		[]string{"conda list"})
}
