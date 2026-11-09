package rules

import "testing"

// gitAmendInitialCommitOutput is the output of git commit --amend in a
// repository without commits.
func gitAmendInitialCommitOutput() string {
	return "fatal: You have nothing to amend.\n"
}

func TestRuleGitAmendInitialCommit(t *testing.T) {
	resetTestState(t)
	rule := Get("git_amend_initial_commit")
	if rule == nil {
		t.Fatal("git_amend_initial_commit rule is not registered")
	}

	assertMatch(t, rule, "git commit --amend", gitAmendInitialCommitOutput())
	assertMatch(t, rule, "git commit --amend -m \"first\"", gitAmendInitialCommitOutput())
	assertNotMatch(t, rule, "git commit -m \"first\"", gitAmendInitialCommitOutput())
	assertNotMatch(t, rule, "git commit --amend", "")
	assertNotMatch(t, rule, "svn commit --amend", gitAmendInitialCommitOutput())

	assertNewCommands(t, rule, "git commit --amend", gitAmendInitialCommitOutput(),
		[]string{"git commit"})
	assertNewCommands(t, rule, "git commit --amend -m \"first\"", gitAmendInitialCommitOutput(),
		[]string{"git commit -m \"first\""})
}
