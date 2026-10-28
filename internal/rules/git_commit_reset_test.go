package rules

import "testing"

func TestRuleGitCommitReset(t *testing.T) {
	resetTestState(t)
	rule := Get("git_commit_reset")
	if rule == nil {
		t.Fatal("git_commit_reset rule is not registered")
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

	for _, script := range []string{"git commit -m \"test commit\"", "git commit"} {
		t.Run("new_command/"+script, func(t *testing.T) {
			assertNewCommands(t, rule, script, "", []string{"git reset HEAD~"})
		})
	}
}
