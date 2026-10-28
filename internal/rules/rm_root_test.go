package rules

import "testing"

func TestRuleRmRoot(t *testing.T) {
	resetTestState(t)
	rule := Get("rm_root")
	if rule == nil {
		t.Fatal("rm_root rule is not registered")
	}
	if rule.EnabledByDefault {
		t.Error("rm_root must be disabled by default")
	}

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "rm -rf /", "add --no-preserve-root")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "ls", "add --no-preserve-root")
		assertNotMatch(t, rule, "rm --no-preserve-root /", "add --no-preserve-root")
		assertNotMatch(t, rule, "rm -rf /", "")
	})

	t.Run("new command", func(t *testing.T) {
		assertNewCommands(t, rule, "rm -rf /", "",
			[]string{"rm -rf / --no-preserve-root"})
	})
}
