package rules

import "testing"

// gitBranchDeleteOutput is the output git prints for an unmerged branch.
func gitBranchDeleteOutput() string {
	return "error: The branch 'branch' is not fully merged.\n" +
		"If you are sure you want to delete it, run 'git branch -D branch'.\n\n"
}

func TestRuleGitBranchDelete(t *testing.T) {
	resetTestState(t)
	rule := Get("git_branch_delete")
	if rule == nil {
		t.Fatal("git_branch_delete rule is not registered")
	}

	assertMatch(t, rule, "git branch -d branch", gitBranchDeleteOutput())
	assertNotMatch(t, rule, "git branch -d branch", "")
	assertNotMatch(t, rule, "ls", gitBranchDeleteOutput())

	assertNewCommands(t, rule, "git branch -d branch", gitBranchDeleteOutput(),
		[]string{"git branch -D branch"})
}
