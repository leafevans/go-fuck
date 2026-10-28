package rules

import "testing"

func TestRuleGoRun(t *testing.T) {
	resetTestState(t)
	rule := Get("go_run")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "go run foo", "")
		assertMatch(t, rule, "go run bar", "")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "go run foo.go", "")
		assertNotMatch(t, rule, "go build foo", "")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "go run foo", "", []string{"go run foo.go"})
		assertNewCommands(t, rule, "go run bar", "", []string{"go run bar.go"})
	})
}
