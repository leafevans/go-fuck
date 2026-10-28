package rules

import "testing"

// gitBranchDeleteCheckedOutOutput is the output git prints for a checked out
// branch.
func gitBranchDeleteCheckedOutOutput() string {
	return "error: Cannot delete branch 'foo' checked out at '/bar/foo'"
}

func TestRuleGitBranchDeleteCheckedOut(t *testing.T) {
	resetTestState(t)
	rule := Get("git_branch_delete_checked_out")
	if rule == nil {
		t.Fatal("git_branch_delete_checked_out rule is not registered")
	}

	for _, script := range []string{"git branch -d foo", "git branch -D foo"} {
		t.Run("match/"+script, func(t *testing.T) {
			assertMatch(t, rule, script, gitBranchDeleteCheckedOutOutput())
		})
		t.Run("not_match/"+script, func(t *testing.T) {
			assertNotMatch(t, rule, script, "Deleted branch foo (was a1b2c3d).")
		})
	}

	cases := []struct {
		script string
		want   string
	}{
		{"git branch -d foo", "git checkout master && git branch -D foo"},
		{"git branch -D foo", "git checkout master && git branch -D foo"},
	}
	for _, testCase := range cases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, gitBranchDeleteCheckedOutOutput(),
				[]string{testCase.want})
		})
	}
}
