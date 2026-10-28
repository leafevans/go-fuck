package rules

import "testing"

func TestRuleSlLs(t *testing.T) {
	resetTestState(t)
	rule := Get("sl_ls")
	if rule == nil {
		t.Fatal("sl_ls rule is not registered")
	}

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "sl", "")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "ls", "")
		assertNotMatch(t, rule, "sl -a", "")
	})

	t.Run("new command", func(t *testing.T) {
		assertNewCommands(t, rule, "sl", "", []string{"ls"})
	})
}
