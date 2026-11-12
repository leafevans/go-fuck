package rules

import "testing"

// stubGitSubmodulePaths replaces the configured submodule paths of the rule.
func stubGitSubmodulePaths(t *testing.T, paths []string) {
	t.Helper()
	original := gitSubmodulePathsCommand
	gitSubmodulePathsCommand = func() []string { return paths }
	t.Cleanup(func() { gitSubmodulePathsCommand = original })
}

func TestRuleGitSubmoduleMissing(t *testing.T) {
	resetTestState(t)
	rule := Get("git_submodule_missing")
	if rule == nil {
		t.Fatal("git_submodule_missing rule is not registered")
	}

	stubGitSubmodulePaths(t, []string{"vendor/foo", "vendor/bar"})
	assertMatch(t, rule, "git submodule update vendor/fooo",
		"error: pathspec 'vendor/fooo' did not match any file(s) known to git\n")
	assertMatch(t, rule, "git submodule status vendor/fooo",
		"error: pathspec 'vendor/fooo' did not match any file(s) known to git\n")

	assertNotMatch(t, rule, "git submodule update zzz",
		"error: pathspec 'zzz' did not match any file(s) known to git\n")
	assertNotMatch(t, rule, "git checkout vendor/fooo",
		"error: pathspec 'vendor/fooo' did not match any file(s) known to git\n")
	assertNotMatch(t, rule, "git submodule update vendor/fooo", "")

	assertNewCommands(t, rule, "git submodule update vendor/fooo",
		"error: pathspec 'vendor/fooo' did not match any file(s) known to git\n",
		[]string{"git submodule update vendor/foo"})
}
