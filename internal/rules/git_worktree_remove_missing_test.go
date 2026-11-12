package rules

import "testing"

// stubGitWorktreePaths replaces the registered worktree paths of the rule.
func stubGitWorktreePaths(t *testing.T, paths []string) {
	t.Helper()
	original := gitWorktreePathsCommand
	gitWorktreePathsCommand = func() []string { return paths }
	t.Cleanup(func() { gitWorktreePathsCommand = original })
}

func TestRuleGitWorktreeRemoveMissing(t *testing.T) {
	resetTestState(t)
	rule := Get("git_worktree_remove_missing")
	if rule == nil {
		t.Fatal("git_worktree_remove_missing rule is not registered")
	}

	stubGitWorktreePaths(t, []string{"/repo/feature", "/repo/with space"})
	assertMatch(t, rule, "git worktree remove /repo/featre",
		"fatal: '/repo/featre' is not a working tree\n")
	assertMatch(t, rule, "git worktree remove '/repo/wit space'",
		"fatal: '/repo/wit space' is not a working tree\n")

	assertNotMatch(t, rule, "git worktree remove /tmp/zzz",
		"fatal: '/tmp/zzz' is not a working tree\n")
	// The main working tree cannot be removed, so it is not suggested.
	stubGitWorktreePaths(t, []string{"/repo/feature"})
	assertNotMatch(t, rule, "git worktree remove /repo/main",
		"fatal: '/repo/main' is not a working tree\n")
	assertNotMatch(t, rule, "git worktree add /repo/new feature",
		"fatal: '/repo/featre' is not a working tree\n")
	assertNotMatch(t, rule, "git worktree remove /repo/featre", "")

	stubGitWorktreePaths(t, []string{"/repo/feature", "/repo/with space"})
	assertNewCommands(t, rule, "git worktree remove /repo/featre",
		"fatal: '/repo/featre' is not a working tree\n",
		[]string{"git worktree remove /repo/feature"})
	assertNewCommands(t, rule, "git worktree remove '/repo/wit space'",
		"fatal: '/repo/wit space' is not a working tree\n",
		[]string{"git worktree remove '/repo/with space'"})
}
