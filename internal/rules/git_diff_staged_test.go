package rules

import "testing"

func TestRuleGitDiffStaged(t *testing.T) {
	resetTestState(t)
	rule := Get("git_diff_staged")
	if rule == nil {
		t.Fatal("git_diff_staged rule is not registered")
	}

	for _, script := range []string{"git diff foo", "git diff"} {
		t.Run("match/"+script, func(t *testing.T) {
			assertMatch(t, rule, script, "")
		})
	}

	for _, script := range []string{"git diff --staged", "git tag", "git branch", "git log"} {
		t.Run("not_match/"+script, func(t *testing.T) {
			assertNotMatch(t, rule, script, "")
		})
	}

	commandCases := []struct {
		script string
		want   string
	}{
		{"git diff", "git diff --staged"},
		{"git diff foo", "git diff --staged foo"},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, "", []string{testCase.want})
		})
	}
}
