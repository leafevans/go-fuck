package rules

import "testing"

func TestRuleCdParent(t *testing.T) {
	resetTestState(t)
	rule := Get("cd_parent")
	if rule == nil {
		t.Fatal("cd_parent rule is not registered")
	}

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "cd..", "cd..: command not found")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "", "")
		assertNotMatch(t, rule, "cd ..", "")
		assertNotMatch(t, rule, "cd...", "")
	})

	t.Run("new command", func(t *testing.T) {
		assertNewCommands(t, rule, "cd..", "", []string{"cd .."})
	})
}
