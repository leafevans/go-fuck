package rules

import "testing"

// stubGitLocalBranches replaces the local branch list of the rule.
func stubGitLocalBranches(t *testing.T, names []string) {
	t.Helper()
	original := gitLocalBranchesCommand
	gitLocalBranchesCommand = func() []string { return names }
	t.Cleanup(func() { gitLocalBranchesCommand = original })
}

func TestRuleGitBranchMissing(t *testing.T) {
	resetTestState(t)
	rule := Get("git_branch_missing")
	if rule == nil {
		t.Fatal("git_branch_missing rule is not registered")
	}

	stubGitLocalBranches(t, []string{"feature", "main"})
	assertMatch(t, rule, "git branch -d featre", "error: branch 'featre' not found\n")
	assertMatch(t, rule, "git branch -D featre", "error: branch 'featre' not found\n")
	assertMatch(t, rule, "git branch --delete featre", "error: branch 'featre' not found\n")
	assertMatch(t, rule, "git branch -m featre renamed", "fatal: no branch named 'featre'\n")
	assertMatch(t, rule, "git branch -M featre renamed", "fatal: no branch named 'featre'\n")
	assertMatch(t, rule, "git -C repo branch -d featre", "error: branch 'featre' not found\n")

	assertNotMatch(t, rule, "git branch -d zzz", "error: branch 'zzz' not found\n")
	assertNotMatch(t, rule, "git checkout featre", "error: branch 'featre' not found\n")
	assertNotMatch(t, rule, "git push origin featre", "error: branch 'featre' not found\n")
	assertNotMatch(t, rule, "git branch -d featre", "")

	assertNewCommands(t, rule, "git branch -d featre", "error: branch 'featre' not found\n",
		[]string{"git branch -d feature"})
	assertNewCommands(t, rule, "git branch -m featre renamed", "fatal: no branch named 'featre'\n",
		[]string{"git branch -m feature renamed"})
	assertNewCommands(t, rule, "git -C repo branch -d featre", "error: branch 'featre' not found\n",
		[]string{"git -C repo branch -d feature"})
}
