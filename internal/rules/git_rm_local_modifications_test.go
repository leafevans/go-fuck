package rules

import "testing"

// gitRmLocalModificationsOutput returns the output of git rm on a file with
// local modifications.
func gitRmLocalModificationsOutput(target string) string {
	return "error: the following file has local modifications:\n    " + target +
		"\n(use --cached to keep the file, or -f to force removal)"
}

func TestRuleGitRmLocalModifications(t *testing.T) {
	resetTestState(t)
	rule := Get("git_rm_local_modifications")
	if rule == nil {
		t.Fatal("git_rm_local_modifications rule is not registered")
	}

	t.Run("matches local modifications", func(t *testing.T) {
		assertMatch(t, rule, "git rm foo", gitRmLocalModificationsOutput("foo"))
		assertMatch(t, rule, "git rm foo bar", gitRmLocalModificationsOutput("bar"))
	})

	t.Run("does not match without output", func(t *testing.T) {
		for _, script := range []string{"git rm foo", "git rm foo bar", "git rm"} {
			assertNotMatch(t, rule, script, "")
		}
	})

	t.Run("new command suggests --cached and -f", func(t *testing.T) {
		assertNewCommands(t, rule, "git rm foo", gitRmLocalModificationsOutput("foo"),
			[]string{"git rm --cached foo", "git rm -f foo"})
		assertNewCommands(t, rule, "git rm foo bar", gitRmLocalModificationsOutput("bar"),
			[]string{"git rm --cached foo bar", "git rm -f foo bar"})
	})
}
