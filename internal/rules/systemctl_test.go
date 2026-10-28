package rules

import "testing"

// systemctlUnknownOperation is the error printed for a misordered command.
const systemctlUnknownOperation = "Unknown operation 'nginx'."

func TestRuleSystemctl(t *testing.T) {
	resetTestState(t)
	rule := Get("systemctl")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "systemctl nginx start", systemctlUnknownOperation)
		assertMatch(t, rule, "sudo systemctl nginx start", systemctlUnknownOperation)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "systemctl start nginx", "")
		assertNotMatch(t, rule, "sudo systemctl nginx", systemctlUnknownOperation)
		assertNotMatch(t, rule, "systemctl nginx", systemctlUnknownOperation)
		assertNotMatch(t, rule, "systemctl start wtf", "Failed to start wtf.service: Unit wtf.service failed to load: No such file or directory.")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "systemctl nginx start", "", []string{"systemctl start nginx"})
		assertNewCommands(t, rule, "sudo systemctl nginx start", "", []string{"sudo systemctl start nginx"})
	})
}
