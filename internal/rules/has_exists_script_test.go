package rules

import "testing"

func TestRuleHasExistsScript(t *testing.T) {
	resetTestState(t)
	rule := Get("has_exists_script")
	if rule == nil {
		t.Fatal("has_exists_script rule is not registered")
	}

	original := hasExistsScriptPathExists
	t.Cleanup(func() { hasExistsScriptPathExists = original })

	hasExistsScriptPathExists = func(string) bool { return true }
	assertMatch(t, rule, "main", "main: command not found")
	assertMatch(t, rule, "main --help", "main: command not found")
	assertNotMatch(t, rule, "main", "")

	hasExistsScriptPathExists = func(string) bool { return false }
	assertNotMatch(t, rule, "main", "main: command not found")

	assertNewCommands(t, rule, "main --help", "", []string{"./main --help"})
}
