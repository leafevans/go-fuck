package rules

import "testing"

func TestRuleGitHelpAliased(t *testing.T) {
	resetTestState(t)
	rule := Get("git_help_aliased")
	if rule == nil {
		t.Fatal("git_help_aliased rule is not registered")
	}

	matchCases := []struct {
		script string
		output string
	}{
		{"git help st", "`git st' is aliased to `status'"},
		{"git help ds", "`git ds' is aliased to `diff --staged'"},
	}
	for _, testCase := range matchCases {
		t.Run("match/"+testCase.script, func(t *testing.T) {
			assertMatch(t, rule, testCase.script, testCase.output)
		})
	}

	notMatchCases := []struct {
		script string
		output string
	}{
		{"git help status", "GIT-STATUS(1)...Git Manual...GIT-STATUS(1)"},
		{"git help diff", "GIT-DIFF(1)...Git Manual...GIT-DIFF(1)"},
	}
	for _, testCase := range notMatchCases {
		t.Run("not_match/"+testCase.script, func(t *testing.T) {
			assertNotMatch(t, rule, testCase.script, testCase.output)
		})
	}

	commandCases := []struct {
		script string
		output string
		want   string
	}{
		{"git help st", "`git st' is aliased to `status'", "git help status"},
		{"git help ds", "`git ds' is aliased to `diff --staged'", "git help diff"},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, testCase.output,
				[]string{testCase.want})
		})
	}
}
