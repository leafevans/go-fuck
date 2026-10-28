package rules

import "testing"

func TestRuleGitBranchList(t *testing.T) {
	resetTestState(t)
	rule := Get("git_branch_list")
	if rule == nil {
		t.Fatal("git_branch_list rule is not registered")
	}

	assertMatch(t, rule, "git branch list", "")

	for _, script := range []string{"", "git commit", "git branch", "git stash list"} {
		t.Run("not_match/"+script, func(t *testing.T) {
			assertNotMatch(t, rule, script, "")
		})
	}

	assertNewCommands(t, rule, "git branch list", "",
		[]string{"git branch --delete list && git branch"})
}
