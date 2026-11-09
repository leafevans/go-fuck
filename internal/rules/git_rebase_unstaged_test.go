package rules

import "testing"

// gitRebaseUnstagedOutput is the output of a rebase with unstaged changes.
func gitRebaseUnstagedOutput() string {
	return "error: cannot rebase: You have unstaged changes.\n" +
		"error: Please commit or stash them.\n"
}

func TestRuleGitRebaseUnstaged(t *testing.T) {
	resetTestState(t)
	rule := Get("git_rebase_unstaged")
	if rule == nil {
		t.Fatal("git_rebase_unstaged rule is not registered")
	}
	if rule.Priority != 900 {
		t.Errorf("git_rebase_unstaged priority = %d, want 900", rule.Priority)
	}

	assertMatch(t, rule, "git rebase master", gitRebaseUnstagedOutput())
	assertMatch(t, rule, "git rebase --continue", "error: cannot rebase: Your index contains uncommitted changes.\n")
	assertNotMatch(t, rule, "git pull --rebase", "error: Cannot pull with rebase: You have unstaged changes.\n")
	assertNotMatch(t, rule, "git merge master", gitRebaseUnstagedOutput())
	assertNotMatch(t, rule, "git rebase master", "")

	assertNewCommands(t, rule, "git rebase master", gitRebaseUnstagedOutput(),
		[]string{"git stash && git rebase master && git stash pop"})
}
