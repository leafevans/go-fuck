package rules

import "testing"

func TestRuleGitCommitAmend(t *testing.T) {
	resetTestState(t)
	rule := Get("git_commit_amend")
	if rule == nil {
		t.Fatal("git_commit_amend rule is not registered")
	}
	if rule.Priority != 1500 {
		t.Errorf("git_commit_amend priority = %d, want 1500", rule.Priority)
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
			assertNewCommands(t, rule, script, "", []string{"git commit --amend"})
		})
	}
}
