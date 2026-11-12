package rules

import "testing"

// stubGitRemotes replaces the configured remote list of the rule.
func stubGitRemotes(t *testing.T, names []string) {
	t.Helper()
	original := gitRemoteNamesCommand
	gitRemoteNamesCommand = func() []string { return names }
	t.Cleanup(func() { gitRemoteNamesCommand = original })
}

func TestRuleGitUnknownRemote(t *testing.T) {
	resetTestState(t)
	rule := Get("git_unknown_remote")
	if rule == nil {
		t.Fatal("git_unknown_remote rule is not registered")
	}

	stubGitRemotes(t, []string{"origin", "upstream"})
	assertMatch(t, rule, "git remote remove orign", "error: No such remote: 'orign'\n")
	assertMatch(t, rule, "git remote rm orign", "error: No such remote: 'orign'\n")
	assertMatch(t, rule, "git remote rename orign upstream", "error: No such remote: 'orign'\n")
	assertMatch(t, rule, "git remote get-url orign", "error: No such remote 'orign'\n")
	assertMatch(t, rule, "git remote set-url orign https://example.com/x",
		"error: No such remote 'orign'\n")
	assertMatch(t, rule, "git remote show orign",
		"fatal: 'orign' does not appear to be a git repository\n"+
			"fatal: Could not read from remote repository.\n")

	assertNotMatch(t, rule, "git remote add orign url", "error: No such remote: 'orign'\n")
	assertNotMatch(t, rule, "git push origin main", "error: No such remote: 'orign'\n")
	assertNotMatch(t, rule, "git remote remove orign", "")
	assertNotMatch(t, rule, "git remote remove zzz", "error: No such remote: 'zzz'\n")

	assertNewCommands(t, rule, "git remote remove orign", "error: No such remote: 'orign'\n",
		[]string{"git remote remove origin"})
	assertNewCommands(t, rule, "git remote rename orign upstream", "error: No such remote: 'orign'\n",
		[]string{"git remote rename origin upstream"})
	assertNewCommands(t, rule, "git remote get-url orign", "error: No such remote 'orign'\n",
		[]string{"git remote get-url origin"})
	assertNewCommands(t, rule, "git remote set-url orign https://example.com/x",
		"error: No such remote 'orign'\n",
		[]string{"git remote set-url origin https://example.com/x"})
}
