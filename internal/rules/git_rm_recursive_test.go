package rules

import "testing"

// gitRmRecursiveOutput returns the output of git rm on a directory without -r.
func gitRmRecursiveOutput(target string) string {
	return "fatal: not removing '" + target + "' recursively without -r"
}

func TestRuleGitRmRecursive(t *testing.T) {
	resetTestState(t)
	rule := Get("git_rm_recursive")
	if rule == nil {
		t.Fatal("git_rm_recursive rule is not registered")
	}

	t.Run("matches a directory without -r", func(t *testing.T) {
		assertMatch(t, rule, "git rm foo", gitRmRecursiveOutput("foo"))
		assertMatch(t, rule, "git rm foo bar", gitRmRecursiveOutput("foo bar"))
	})

	t.Run("does not match without output", func(t *testing.T) {
		assertNotMatch(t, rule, "git rm foo", "")
		assertNotMatch(t, rule, "git rm foo bar", "")
	})

	t.Run("new command adds -r", func(t *testing.T) {
		assertNewCommands(t, rule, "git rm foo", gitRmRecursiveOutput("foo"),
			[]string{"git rm -r foo"})
		assertNewCommands(t, rule, "git rm foo bar", gitRmRecursiveOutput("foo bar"),
			[]string{"git rm -r foo bar"})
	})
}
