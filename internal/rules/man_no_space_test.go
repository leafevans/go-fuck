package rules

import "testing"

func TestRuleManNoSpace(t *testing.T) {
	resetTestState(t)
	rule := Get("man_no_space")
	if rule == nil {
		t.Fatal("man_no_space rule is not registered")
	}

	assertMatch(t, rule, "mandiff", "mandiff: command not found")
	assertNotMatch(t, rule, "", "")

	assertNewCommands(t, rule, "mandiff", "", []string{"man diff"})
}
