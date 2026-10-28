package rules

import "testing"

func TestRuleGrepRecursive(t *testing.T) {
	resetTestState(t)
	rule := Get("grep_recursive")
	if rule == nil {
		t.Fatal("grep_recursive rule is not registered")
	}

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "grep blah .", "grep: .: Is a directory")
		assertMatch(t, rule, "grep café .", "grep: .: Is a directory")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "", "")
		assertNotMatch(t, rule, "grep blah .", "")
	})

	t.Run("new command", func(t *testing.T) {
		assertNewCommands(t, rule, "grep blah .", "",
			[]string{"grep -r blah ."})
		assertNewCommands(t, rule, "grep café .", "",
			[]string{"grep -r café ."})
	})
}
