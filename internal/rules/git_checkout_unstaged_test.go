package rules

import "testing"

// gitCheckoutUnstagedOutput is the output of a checkout or switch that would
// overwrite local changes.
func gitCheckoutUnstagedOutput() string {
	return "error: Your local changes to the following files would be overwritten by checkout:\n" +
		"\tf.txt\n" +
		"Please commit your changes or stash them before you switch branches.\n" +
		"Aborting\n"
}

func TestRuleGitCheckoutUnstaged(t *testing.T) {
	resetTestState(t)
	rule := Get("git_checkout_unstaged")
	if rule == nil {
		t.Fatal("git_checkout_unstaged rule is not registered")
	}
	if rule.Priority != 900 {
		t.Errorf("git_checkout_unstaged priority = %d, want 900", rule.Priority)
	}

	assertMatch(t, rule, "git checkout other", gitCheckoutUnstagedOutput())
	assertMatch(t, rule, "git switch other", gitCheckoutUnstagedOutput())
	assertNotMatch(t, rule, "git merge other", gitCheckoutUnstagedOutput())
	assertNotMatch(t, rule, "git checkout other", "")
	assertNotMatch(t, rule, "git checkout -- f.txt", "error: pathspec 'f.txt' did not match any file(s) known to git")

	assertNewCommands(t, rule, "git checkout other", gitCheckoutUnstagedOutput(),
		[]string{"git stash && git checkout other && git stash pop"})
	assertNewCommands(t, rule, "git switch other", gitCheckoutUnstagedOutput(),
		[]string{"git stash && git switch other && git stash pop"})
}
