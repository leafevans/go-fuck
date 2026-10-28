package rules

import "testing"

// gitRebaseNoChangesOutput is the output of git rebase --continue when there
// are no changes left to apply.
func gitRebaseNoChangesOutput() string {
	return "Applying: Test commit\n" +
		"No changes - did you forget to use 'git add'?\n" +
		"If there is nothing left to stage, chances are that something else\n" +
		"already introduced the same changes; you might want to skip this patch.\n\n" +
		"When you have resolved this problem, run \"git rebase --continue\".\n" +
		"If you prefer to skip this patch, run \"git rebase --skip\" instead.\n" +
		"To check out the original branch and stop rebasing, run \"git rebase --abort\".\n\n"
}

func TestRuleGitRebaseNoChanges(t *testing.T) {
	resetTestState(t)
	rule := Get("git_rebase_no_changes")
	if rule == nil {
		t.Fatal("git_rebase_no_changes rule is not registered")
	}

	t.Run("matches an empty patch", func(t *testing.T) {
		assertMatch(t, rule, "git rebase --continue", gitRebaseNoChangesOutput())
	})

	t.Run("does not match without output", func(t *testing.T) {
		assertNotMatch(t, rule, "git rebase --continue", "")
		assertNotMatch(t, rule, "git rebase --skip", "")
		assertNotMatch(t, rule, "rebase --continue", gitRebaseNoChangesOutput())
	})

	t.Run("new command skips the patch", func(t *testing.T) {
		assertNewCommands(t, rule, "git rebase --continue", gitRebaseNoChangesOutput(),
			[]string{"git rebase --skip"})
	})
}
