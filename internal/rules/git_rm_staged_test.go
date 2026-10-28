package rules

import "testing"

// gitRmStagedOutput returns the output of git rm on a file with staged changes.
func gitRmStagedOutput(target string) string {
	return "error: the following file has changes staged in the index:\n    " + target +
		"\n(use --cached to keep the file, or -f to force removal)"
}

func TestRuleGitRmStaged(t *testing.T) {
	resetTestState(t)
	rule := Get("git_rm_staged")
	if rule == nil {
		t.Fatal("git_rm_staged rule is not registered")
	}

	t.Run("matches staged changes", func(t *testing.T) {
		assertMatch(t, rule, "git rm foo", gitRmStagedOutput("foo"))
		assertMatch(t, rule, "git rm foo bar", gitRmStagedOutput("bar"))
	})

	t.Run("does not match without output", func(t *testing.T) {
		for _, script := range []string{"git rm foo", "git rm foo bar", "git rm"} {
			assertNotMatch(t, rule, script, "")
		}
	})

	t.Run("new command suggests --cached and -f", func(t *testing.T) {
		assertNewCommands(t, rule, "git rm foo", gitRmStagedOutput("foo"),
			[]string{"git rm --cached foo", "git rm -f foo"})
		assertNewCommands(t, rule, "git rm foo bar", gitRmStagedOutput("bar"),
			[]string{"git rm --cached foo bar", "git rm -f foo bar"})
	})
}
