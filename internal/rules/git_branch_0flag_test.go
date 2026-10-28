package rules

import "testing"

// gitBranch0flagExistsOutput is the output git prints for an existing branch.
func gitBranch0flagExistsOutput() string {
	return "fatal: A branch named 'bar' already exists."
}

func TestRuleGitBranch0flag(t *testing.T) {
	resetTestState(t)
	rule := Get("git_branch_0flag")
	if rule == nil {
		t.Fatal("git_branch_0flag rule is not registered")
	}

	for _, script := range []string{
		"git branch 0a",
		"git branch 0d",
		"git branch 0f",
		"git branch 0r",
		"git branch 0v",
		"git branch 0d foo",
		"git branch 0D foo",
	} {
		t.Run("match/"+script, func(t *testing.T) {
			assertMatch(t, rule, script, gitBranch0flagExistsOutput())
		})
	}

	for _, script := range []string{
		"git branch -a",
		"git branch -r",
		"git branch -v",
		"git branch -d foo",
		"git branch -D foo",
	} {
		t.Run("not_match/"+script, func(t *testing.T) {
			assertNotMatch(t, rule, script, "")
		})
	}

	existingBranchCases := []struct {
		script string
		want   string
	}{
		{"git branch 0a", "git branch -D 0a && git branch -a"},
		{"git branch 0v", "git branch -D 0v && git branch -v"},
		{"git branch 0d foo", "git branch -D 0d && git branch -d foo"},
		{"git branch 0D foo", "git branch -D 0D && git branch -D foo"},
		{"git branch 0l 'maint-*'", "git branch -D 0l && git branch -l 'maint-*'"},
		{"git branch 0u upstream", "git branch -D 0u && git branch -u upstream"},
	}
	for _, testCase := range existingBranchCases {
		t.Run("new_command_branch_exists/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, gitBranch0flagExistsOutput(),
				[]string{testCase.want})
		})
	}

	notValidObjectCases := []struct {
		script string
		want   string
	}{
		{"git branch 0l 'maint-*'", "git branch -l 'maint-*'"},
		{"git branch 0u upstream", "git branch -u upstream"},
	}
	for _, testCase := range notValidObjectCases {
		t.Run("new_command_not_valid_object/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script,
				"fatal: Not a valid object name: 'bar'.", []string{testCase.want})
		})
	}
}
