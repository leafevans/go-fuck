package rules

import "testing"

// gitSwitchExistingBranchOutput is the output of git switch -c with a branch
// that already exists.
func gitSwitchExistingBranchOutput(branch string) string {
	return "fatal: a branch named '" + branch + "' already exists\n"
}

func TestRuleGitSwitchExistingBranch(t *testing.T) {
	resetTestState(t)
	rule := Get("git_switch_existing_branch")
	if rule == nil {
		t.Fatal("git_switch_existing_branch rule is not registered")
	}

	assertMatch(t, rule, "git switch -c feature", gitSwitchExistingBranchOutput("feature"))
	assertMatch(t, rule, "git switch --create feature", gitSwitchExistingBranchOutput("feature"))
	assertNotMatch(t, rule, "git branch feature", gitSwitchExistingBranchOutput("feature"))
	assertNotMatch(t, rule, "git switch feature", gitSwitchExistingBranchOutput("feature"))
	assertNotMatch(t, rule, "git switch -c feature", "")

	assertNewCommands(t, rule, "git switch -c feature", gitSwitchExistingBranchOutput("feature"),
		[]string{"git switch feature"})
	assertNewCommands(t, rule, "git switch --create feature", gitSwitchExistingBranchOutput("feature"),
		[]string{"git switch feature"})
}
