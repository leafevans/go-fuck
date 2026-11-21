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
	assertMatch(t, rule, "sudo main --help", "main: command not found")

	hasExistsScriptPathExists = func(string) bool { return false }
	assertNotMatch(t, rule, "main", "main: command not found")

	assertNewCommands(t, rule, "main --help", "", []string{"./main --help"})
}

func TestRuleHasExistsScriptSkipsUnneededPathCheck(t *testing.T) {
	resetTestState(t)
	rule := Get("has_exists_script")
	original := hasExistsScriptPathExists
	t.Cleanup(func() { hasExistsScriptPathExists = original })
	checks := 0
	hasExistsScriptPathExists = func(string) bool {
		checks++
		return true
	}
	assertNotMatch(t, rule, "main", "")
	assertNotMatch(t, rule, "git push", "fatal: no upstream branch")
	assertNotMatch(t, rule, "", "command not found")
	if checks != 0 {
		t.Fatalf("checked the filesystem %d times for unrelated errors or empty commands", checks)
	}
}
