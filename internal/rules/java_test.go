package rules

import "testing"

func TestRuleJava(t *testing.T) {
	resetTestState(t)
	rule := Get("java")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "java foo.java", "")
		assertMatch(t, rule, "java bar.java", "")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "java foo", "")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "java foo.java", "", []string{"java foo"})
		assertNewCommands(t, rule, "java bar.java", "", []string{"java bar"})
	})
}
