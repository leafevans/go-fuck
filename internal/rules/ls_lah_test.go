package rules

import "testing"

func TestRuleLsLah(t *testing.T) {
	resetTestState(t)
	rule := Get("ls_lah")
	if rule == nil {
		t.Fatal("ls_lah rule is not registered")
	}

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "ls", "")
		assertMatch(t, rule, "ls file.py", "")
		assertMatch(t, rule, "ls /opt", "")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "ls -lah /opt", "")
		assertNotMatch(t, rule, "pacman -S binutils", "")
		assertNotMatch(t, rule, "lsof", "")
	})

	t.Run("new command", func(t *testing.T) {
		assertNewCommands(t, rule, "ls file.py", "", []string{"ls -lah file.py"})
		assertNewCommands(t, rule, "ls", "", []string{"ls -lah"})
	})
}
