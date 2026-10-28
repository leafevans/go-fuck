package rules

import "testing"

func TestRuleJavac(t *testing.T) {
	resetTestState(t)
	rule := Get("javac")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "javac foo", "")
		assertMatch(t, rule, "javac bar", "")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "javac foo.java", "")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "javac foo", "", []string{"javac foo.java"})
		assertNewCommands(t, rule, "javac bar", "", []string{"javac bar.java"})
	})
}
