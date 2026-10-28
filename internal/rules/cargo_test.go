package rules

import "testing"

func TestRuleCargo(t *testing.T) {
	resetTestState(t)
	rule := Get("cargo")
	if rule == nil {
		t.Fatal("cargo rule is not registered")
	}

	assertMatch(t, rule, "cargo", "")
	assertNotMatch(t, rule, "cargo build", "")
	assertNotMatch(t, rule, "ls", "")

	assertNewCommands(t, rule, "cargo", "", []string{"cargo build"})
}
