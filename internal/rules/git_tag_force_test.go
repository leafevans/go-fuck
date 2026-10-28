package rules

import "testing"

func TestRuleGitTagForce(t *testing.T) {
	resetTestState(t)
	rule := Get("git_tag_force")
	if rule == nil {
		t.Fatal("git_tag_force rule is not registered")
	}

	const output = "fatal: tag 'alert' already exists"

	t.Run("matches an existing tag", func(t *testing.T) {
		assertMatch(t, rule, "git tag alert", output)
	})

	t.Run("does not match without output", func(t *testing.T) {
		assertNotMatch(t, rule, "git tag alert", "")
	})

	t.Run("new command forces the tag", func(t *testing.T) {
		assertNewCommands(t, rule, "git tag alert", output,
			[]string{"git tag --force alert"})
	})
}
