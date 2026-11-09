package rules

import "testing"

// gitMergeUnstagedOutput is the output of a merge that would overwrite local
// changes.
func gitMergeUnstagedOutput() string {
	return "error: Your local changes to the following files would be overwritten by merge:\n" +
		"\tf.txt\n" +
		"Please commit your changes or stash them before you merge.\n" +
		"Aborting\n"
}

func TestRuleGitMergeUnstaged(t *testing.T) {
	resetTestState(t)
	rule := Get("git_merge_unstaged")
	if rule == nil {
		t.Fatal("git_merge_unstaged rule is not registered")
	}
	if rule.Priority != 900 {
		t.Errorf("git_merge_unstaged priority = %d, want 900", rule.Priority)
	}

	assertMatch(t, rule, "git merge other", gitMergeUnstagedOutput())
	assertNotMatch(t, rule, "git checkout other", "error: Your local changes to the following files would be overwritten by checkout:\n\tf.txt\n")
	assertNotMatch(t, rule, "git rebase other", gitMergeUnstagedOutput())
	assertNotMatch(t, rule, "git merge other", "")

	assertNewCommands(t, rule, "git merge other", gitMergeUnstagedOutput(),
		[]string{"git stash && git merge other && git stash pop"})
}
