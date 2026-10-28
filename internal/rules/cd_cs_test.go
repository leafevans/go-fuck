package rules

import "testing"

func TestRuleCdCs(t *testing.T) {
	resetTestState(t)
	rule := Get("cd_cs")
	if rule == nil {
		t.Fatal("cd_cs rule is not registered")
	}
	if rule.Priority != 900 {
		t.Errorf("cd_cs priority is %d, want 900", rule.Priority)
	}

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "cs", "cs: command not found")
		assertMatch(t, rule, "cs /etc/", "cs: command not found")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "", "")
		assertNotMatch(t, rule, "cd /etc/", "")
		assertNotMatch(t, rule, "csfoo", "")
	})

	t.Run("new command", func(t *testing.T) {
		assertNewCommands(t, rule, "cs /etc/", "cs: command not found",
			[]string{"cd /etc/"})
		assertNewCommands(t, rule, "cs", "cs: command not found",
			[]string{"cd"})
	})
}
