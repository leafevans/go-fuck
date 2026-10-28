package rules

import "testing"

func TestRuleGitRemoteDelete(t *testing.T) {
	resetTestState(t)
	rule := Get("git_remote_delete")
	if rule == nil {
		t.Fatal("git_remote_delete rule is not registered")
	}

	t.Run("matches remote delete", func(t *testing.T) {
		assertMatch(t, rule, "git remote delete foo", "")
	})

	t.Run("does not match other commands", func(t *testing.T) {
		for _, script := range []string{
			"git remote remove foo",
			"git remote add foo",
			"git commit",
		} {
			assertNotMatch(t, rule, script, "")
		}
	})

	t.Run("new command replaces delete with remove", func(t *testing.T) {
		assertNewCommands(t, rule, "git remote delete foo", "",
			[]string{"git remote remove foo"})
		assertNewCommands(t, rule, "git remote delete delete", "",
			[]string{"git remote remove delete"})
	})
}
