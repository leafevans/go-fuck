package rules

import "testing"

func TestRuleGitHookBypass(t *testing.T) {
	resetTestState(t)
	rule := Get("git_hook_bypass")
	if rule == nil {
		t.Fatal("git_hook_bypass rule is not registered")
	}
	if rule.Priority != 1100 {
		t.Errorf("git_hook_bypass priority = %d, want 1100", rule.Priority)
	}
	if rule.RequiresOutput {
		t.Error("git_hook_bypass must not require the command output")
	}

	for _, script := range []string{
		"git am",
		"git commit",
		"git commit -m 'foo bar'",
		"git push",
		"git push -u foo bar",
	} {
		t.Run("match/"+script, func(t *testing.T) {
			assertMatch(t, rule, script, "")
		})
	}

	for _, script := range []string{"git add foo", "git status", "git diff foo bar"} {
		t.Run("not_match/"+script, func(t *testing.T) {
			assertNotMatch(t, rule, script, "")
		})
	}

	commandCases := []struct {
		script string
		want   string
	}{
		{"git am", "git am --no-verify"},
		{"git commit", "git commit --no-verify"},
		{"git commit -m 'foo bar'", "git commit --no-verify -m 'foo bar'"},
		{"git push", "git push --no-verify"},
		{"git push -p", "git push --no-verify -p"},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, "", []string{testCase.want})
		})
	}
}
