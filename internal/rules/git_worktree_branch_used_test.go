package rules

import "testing"

func TestRuleGitWorktreeBranchUsed(t *testing.T) {
	resetTestState(t)
	rule := Get("git_worktree_branch_used")
	if rule == nil {
		t.Fatal("git_worktree_branch_used rule is not registered")
	}

	assertMatch(t, rule, "git worktree add ../wt feature",
		"fatal: 'feature' is already used by worktree at '/repo/feature'\n")

	assertNotMatch(t, rule, "git worktree remove ../wt",
		"fatal: 'feature' is already used by worktree at '/repo/feature'\n")
	assertNotMatch(t, rule, "git worktree add ../wt feature", "")

	assertNewCommands(t, rule, "git worktree add ../wt feature",
		"fatal: 'feature' is already used by worktree at '/repo/feature'\n",
		[]string{"git worktree add --force ../wt feature"})
}
