package rules

import "testing"

func TestRuleLsAll(t *testing.T) {
	resetTestState(t)
	rule := Get("ls_all")
	if rule == nil {
		t.Fatal("ls_all rule is not registered")
	}

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "ls", "")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "ls", "file.py\n")
		assertNotMatch(t, rule, "cat", "")
	})

	t.Run("new command", func(t *testing.T) {
		assertNewCommands(t, rule, "ls empty_dir", "", []string{"ls -A empty_dir"})
		assertNewCommands(t, rule, "ls", "", []string{"ls -A"})
	})
}
