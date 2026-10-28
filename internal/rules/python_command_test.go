package rules

import "testing"

func TestRulePythonCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("python_command")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "temp.py", "Permission denied")
		assertMatch(t, rule, "temp.py", "bash: temp.py: command not found")
		assertMatch(t, rule, "sudo temp.py", "Permission denied")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "", "")
		assertNotMatch(t, rule, "temp.py", "")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "./test_sudo.py", "", []string{"python ./test_sudo.py"})
		assertNewCommands(t, rule, "sudo ./test_sudo.py", "", []string{"sudo python ./test_sudo.py"})
	})
}
