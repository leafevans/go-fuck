package rules

import "testing"

// gitPullUncommittedChangesOutput is the output of git pull with local changes.
func gitPullUncommittedChangesOutput() string {
	return "error: Cannot pull with rebase: You have unstaged changes."
}

func TestRuleGitPullUncommittedChanges(t *testing.T) {
	resetTestState(t)
	rule := Get("git_pull_uncommitted_changes")
	if rule == nil {
		t.Fatal("git_pull_uncommitted_changes rule is not registered")
	}

	t.Run("matches unstaged changes", func(t *testing.T) {
		assertMatch(t, rule, "git pull", gitPullUncommittedChangesOutput())
	})

	t.Run("does not match without output", func(t *testing.T) {
		assertNotMatch(t, rule, "git pull", "")
	})

	t.Run("does not match other applications", func(t *testing.T) {
		assertNotMatch(t, rule, "ls", gitPullUncommittedChangesOutput())
	})

	t.Run("new command stashes the changes", func(t *testing.T) {
		assertNewCommands(t, rule, "git pull", gitPullUncommittedChangesOutput(),
			[]string{"git stash && git pull && git stash pop"})
	})
}
