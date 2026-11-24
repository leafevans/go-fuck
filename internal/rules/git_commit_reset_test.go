package rules

import "testing"

func TestRuleGitCommitReset(t *testing.T) {
	resetTestState(t)
	rule := Get("git_commit_reset")
	if rule == nil {
		t.Fatal("git_commit_reset rule is not registered")
	}
	if rule.Priority != 1500 {
		t.Errorf("git_commit_reset priority = %d, want 1500", rule.Priority)
	}

	matchCases := []struct {
		script string
		output string
	}{
		{"git commit -m \"test\"", "test output"},
		{"git commit", ""},
	}
	for _, testCase := range matchCases {
		t.Run("match/"+testCase.script, func(t *testing.T) {
			assertMatch(t, rule, testCase.script, testCase.output)
		})
	}

	for _, script := range []string{"git branch foo", "git checkout feature/test_commit", "git push"} {
		t.Run("not_match/"+script, func(t *testing.T) {
			assertNotMatch(t, rule, script, "")
		})
	}
	// A failure with a concrete cause is not fixed by rewriting history.
	for _, testCase := range []struct {
		script string
		output string
	}{
		{"git commit -m x", "husky - pre-commit hook exited with code 1 (error)"},
		{"git commit -m x", "Author identity unknown"},
		{"git commit", "nothing to commit, working tree clean"},
		{"git commit", "error: Committing is not possible because you have unmerged files."},
	} {
		t.Run("not_match_unrelated/"+testCase.script, func(t *testing.T) {
			assertNotMatch(t, rule, testCase.script, testCase.output)
		})
	}

	for _, script := range []string{"git commit -m \"test commit\"", "git commit"} {
		t.Run("new_command/"+script, func(t *testing.T) {
			assertNewCommands(t, rule, script, "", []string{"git reset HEAD~"})
		})
	}
}
