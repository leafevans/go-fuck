package rules

import "testing"

// gitStashPopOutput is the output of git stash pop with conflicting local
// changes.
func gitStashPopOutput() string {
	return "error: Your local changes to the following files would be overwritten by merge:"
}

func TestRuleGitStashPop(t *testing.T) {
	resetTestState(t)
	rule := Get("git_stash_pop")
	if rule == nil {
		t.Fatal("git_stash_pop rule is not registered")
	}

	t.Run("matches a failed stash pop", func(t *testing.T) {
		assertMatch(t, rule, "git stash pop", gitStashPopOutput())
	})

	t.Run("does not match without output", func(t *testing.T) {
		assertNotMatch(t, rule, "git stash", "")
	})

	t.Run("new command stages the changes", func(t *testing.T) {
		assertNewCommands(t, rule, "git stash pop", gitStashPopOutput(),
			[]string{"git add --update && git stash pop && git reset ."})
	})
}
