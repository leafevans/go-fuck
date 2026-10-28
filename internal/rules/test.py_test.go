package rules

import "testing"

func TestRuleTestDotPy(t *testing.T) {
	resetTestState(t)
	rule := Get("test.py")

	t.Run("priority", func(t *testing.T) {
		// Python: priority = 900, so that it comes before python_command.
		if rule.Priority != 900 {
			t.Errorf("rule %q has priority %d, want %d", rule.Name, rule.Priority, 900)
		}
	})

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "test.py", "bash: test.py: command not found")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "test.py", "")
		assertNotMatch(t, rule, "./test.py", "bash: ./test.py: command not found")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "test.py", "bash: test.py: command not found",
			[]string{"pytest"})
	})
}
