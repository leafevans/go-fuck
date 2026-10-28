package rules

import "testing"

// gitStashCherryPickError is the output of a cherry-pick with local changes.
var gitStashCherryPickError = "error: Your local changes would be overwritten by cherry-pick.\n" +
	"hint: Commit your changes or stash them to proceed.\n" +
	"fatal: cherry-pick failed"

// gitStashRebaseError is the output of a rebase with uncommitted changes.
var gitStashRebaseError = "Cannot rebase: Your index contains uncommitted changes.\n" +
	"Please commit or stash them."

func TestRuleGitStash(t *testing.T) {
	resetTestState(t)
	rule := Get("git_stash")
	if rule == nil {
		t.Fatal("git_stash rule is not registered")
	}

	t.Run("matches a commit or stash request", func(t *testing.T) {
		assertMatch(t, rule, "git cherry-pick a1b2c3d", gitStashCherryPickError)
		assertMatch(t, rule, "git rebase -i HEAD~7", gitStashRebaseError)
	})

	t.Run("does not match without output", func(t *testing.T) {
		assertNotMatch(t, rule, "git cherry-pick a1b2c3d", "")
		assertNotMatch(t, rule, "git rebase -i HEAD~7", "")
	})

	t.Run("new command stashes the changes", func(t *testing.T) {
		assertNewCommands(t, rule, "git cherry-pick a1b2c3d", gitStashCherryPickError,
			[]string{"git stash && git cherry-pick a1b2c3d"})
		assertNewCommands(t, rule, "git rebase -i HEAD~7", gitStashRebaseError,
			[]string{"git stash && git rebase -i HEAD~7"})
	})
}
