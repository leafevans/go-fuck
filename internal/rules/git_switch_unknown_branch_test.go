package rules

import "testing"

// gitSwitchUnknownBranchOutput is the output of git switch with a branch that
// does not exist.
func gitSwitchUnknownBranchOutput(branch string) string {
	return "fatal: invalid reference: " + branch + "\n"
}

// gitSwitchStubBranches replaces the branch lookup of the rules with the given
// names.
func gitSwitchStubBranches(t *testing.T, branches ...string) {
	t.Helper()
	original := gitCheckoutBranches
	gitCheckoutBranches = func() []string { return branches }
	t.Cleanup(func() { gitCheckoutBranches = original })
}

func TestRuleGitSwitchUnknownBranch(t *testing.T) {
	resetTestState(t)
	rule := Get("git_switch_unknown_branch")
	if rule == nil {
		t.Fatal("git_switch_unknown_branch rule is not registered")
	}

	gitSwitchStubBranches(t, "main", "feature", "release/1.6")

	assertMatch(t, rule, "git switch feture", gitSwitchUnknownBranchOutput("feture"))
	assertMatch(t, rule, "git switch --discard-changes feture", gitSwitchUnknownBranchOutput("feture"))
	assertNotMatch(t, rule, "git checkout feture", gitSwitchUnknownBranchOutput("feture"))
	assertNotMatch(t, rule, "git switch main", "")
	assertNotMatch(t, rule, "git switch feture", "Already on 'feature'")

	assertNewCommands(t, rule, "git switch feture", gitSwitchUnknownBranchOutput("feture"),
		[]string{"git switch feature", "git switch -c feture"})

	gitCheckoutBranches = func() []string { return nil }
	assertNewCommands(t, rule, "git switch feture", gitSwitchUnknownBranchOutput("feture"),
		[]string{"git switch -c feture"})
}
