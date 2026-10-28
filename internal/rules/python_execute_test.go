package rules

import "testing"

func TestRulePythonExecute(t *testing.T) {
	resetTestState(t)
	rule := Get("python_execute")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "python foo", "")
		assertMatch(t, rule, "python bar", "")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "python foo.py", "")
		assertNotMatch(t, rule, "vim python foo", "")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "python foo", "", []string{"python foo.py"})
		assertNewCommands(t, rule, "python bar", "", []string{"python bar.py"})
	})
}
