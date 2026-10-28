package rules

import "testing"

// gitMainMasterOutput builds the output git prints for an unknown branch.
func gitMainMasterOutput(branchName string) string {
	if branchName == "" {
		return ""
	}
	return "error: pathspec '" + branchName + "' did not match any file(s) known to git"
}

func TestRuleGitMainMaster(t *testing.T) {
	resetTestState(t)
	rule := Get("git_main_master")
	if rule == nil {
		t.Fatal("git_main_master rule is not registered")
	}
	if rule.Priority != 1200 {
		t.Errorf("git_main_master priority = %d, want 1200", rule.Priority)
	}

	matchCases := []struct {
		script     string
		branchName string
	}{
		{"git checkout main", "main"},
		{"git checkout master", "master"},
		{"git show main", "main"},
	}
	for _, testCase := range matchCases {
		t.Run("match/"+testCase.script, func(t *testing.T) {
			assertMatch(t, rule, testCase.script, gitMainMasterOutput(testCase.branchName))
		})
	}

	notMatchCases := []struct {
		script     string
		branchName string
	}{
		{"git checkout master", ""},
		{"git checkout main", ""},
		{"git checkout wibble", "wibble"},
	}
	for _, testCase := range notMatchCases {
		t.Run("not_match/"+testCase.script, func(t *testing.T) {
			assertNotMatch(t, rule, testCase.script, gitMainMasterOutput(testCase.branchName))
		})
	}

	commandCases := []struct {
		script     string
		branchName string
		want       string
	}{
		{"git checkout main", "main", "git checkout master"},
		{"git checkout master", "master", "git checkout main"},
		{"git checkout wibble", "wibble", "git checkout wibble"},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, gitMainMasterOutput(testCase.branchName),
				[]string{testCase.want})
		})
	}
}
